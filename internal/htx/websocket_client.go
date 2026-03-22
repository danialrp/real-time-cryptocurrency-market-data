package htx

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/fasthttp/websocket"
	"golang.org/x/time/rate"
	"irbtc_streamer/internal/markup"
	"irbtc_streamer/internal/model"
	"irbtc_streamer/internal/redis"
	"irbtc_streamer/internal/stream"
	"irbtc_streamer/internal/utils"
)

type WebSocketClient struct {
	conn           *websocket.Conn
	dispatcher     *stream.Dispatcher
	subscribed     map[string]bool
	invalidTopics  map[string]bool
	reconnectFunc  func() error
	subscribeDelay time.Duration
	ID             int // connection ID for logging
}

func NewWebSocketClient(ctx context.Context, dispatcher *stream.Dispatcher, id int) (*WebSocketClient, error) {
	conn, _, err := websocket.DefaultDialer.Dial("wss://api-aws.huobi.pro/ws", nil)
	if err != nil {
		return nil, err
	}

	client := &WebSocketClient{
		conn:           conn,
		dispatcher:     dispatcher,
		subscribed:     make(map[string]bool),
		invalidTopics:  make(map[string]bool),
		subscribeDelay: 30 * time.Microsecond,
		ID:             id,
	}

	client.reconnectFunc = func() error {
		newConn, _, err := websocket.DefaultDialer.Dial("wss://api-aws.huobi.pro/ws", nil)
		if err != nil {
			return err
		}
		client.conn = newConn
		client.subscribed = make(map[string]bool)
		return nil
	}

	return client, nil
}

func (c *WebSocketClient) Listen() error {
	log.Printf("📡 [HTX #%d] Listening started...", c.ID)

	defer func() {
		if r := recover(); r != nil {
			log.Printf("❌ [HTX #%d] GLOBAL PANIC: %v", c.ID, r)
		}
	}()

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("[HTX #%d] read error: %w", c.ID, err)
		}

		decompressedMsg, err := utils.DecompressGZIP(msg)
		if err != nil {
			log.Printf("❌ Decompress error: %v", err)
			continue
		}

		if strings.Contains(string(decompressedMsg), `"ping":`) {
			var ping map[string]int64
			if err := json.Unmarshal(decompressedMsg, &ping); err == nil {
				pong := fmt.Sprintf(`{"pong": %d}`, ping["ping"])
				_ = c.conn.WriteMessage(websocket.TextMessage, []byte(pong))
			}
			continue
		}

		if strings.Contains(string(decompressedMsg), `"err-msg":"invalid topic`) {
			var parsed map[string]interface{}
			if err := json.Unmarshal(decompressedMsg, &parsed); err == nil {
				if topic, ok := parsed["err-msg"].(string); ok {
					t := extractTopicFromError(topic)
					c.invalidTopics[t] = true
					log.Printf("⚠️ Invalid topic: %s", t)
				}
			}
			continue
		}

		if strings.Contains(string(decompressedMsg), `"status":"error"`) {
			log.Printf("❌ Server error: %s", decompressedMsg)
			continue
		}

		var base struct {
			Ch   string          `json:"ch"`
			Tick json.RawMessage `json:"tick"`
			Ts   int64           `json:"ts"`
		}
		if err := json.Unmarshal(decompressedMsg, &base); err != nil {
			log.Printf("❌ Parse error: %v", err)
			continue
		}

		parts := strings.Split(base.Ch, ".")
		if len(parts) >= 3 {
			symbol := parts[1]
			topic := strings.Join(parts[2:], ".")

			// Clean base.Tick (remove id, seqId from all levels)
			tickCleaned, err := utils.RemoveKeysRecursive(base.Tick, []string{"id", "seqId"})
			if err != nil {
				continue
			}

			ttl := getRawTTLByTopic(topic)
			_ = redis.PublishRawStreamWithTTL(symbol, topic, tickCleaned, ttl)

			if c.dispatcher != nil {
				c.dispatcher.Broadcast(topic, symbol, tickCleaned)
			}

			go markup.Convert(symbol, topic, tickCleaned)
		}
	}
}

func (c *WebSocketClient) AutoResubscribe(symbols []model.Symbol, channels []string) {
	ctx := context.Background()
	limiter := rate.NewLimiter(rate.Every(25*time.Millisecond), 1)

	total := 0
	success := 0
	skipped := 0
	failed := 0

	for _, s := range symbols {
		for _, ch := range channels {
			topic := fmt.Sprintf("market.%s.%s", s.Symbol, ch)
			total++

			if c.invalidTopics[topic] || c.subscribed[topic] {
				skipped++
				continue
			}

			if err := limiter.Wait(ctx); err != nil {
				log.Printf("⚠️ Rate limiter wait error: %v", err)
			}

			err := c.Subscribe(s.Symbol, ch)
			if err != nil {
				log.Printf("❌ Subscribe failed [%s.%s]: %v", s.Symbol, ch, err)
				failed++
				continue
			}

			c.subscribed[topic] = true
			success++
			// log.Printf("✅ Subscribed: %s.%s", s.Symbol, ch)
		}
	}

	log.Printf("🔁 [HTX #%d] Subscribe complete | total: %d | success: %d | skipped: %d | failed: %d",
		c.ID, total, success, skipped, failed)
}

func (c *WebSocketClient) Reconnect(ctx context.Context) error {
	log.Println("🔌 Reconnecting...")
	return c.reconnectFunc()
}

func (c *WebSocketClient) Subscribe(symbol, topic string) error {
	channel := fmt.Sprintf("market.%s.%s", symbol, topic)
	req := map[string]string{
		"sub": channel,
		"id":  channel,
	}
	msg, err := json.Marshal(req)
	if err != nil {
		return err
	}
	return c.conn.WriteMessage(websocket.TextMessage, msg)
}

func extractTopicFromError(msg string) string {
	start := strings.Index(msg, "market.")
	if start == -1 {
		return msg
	}
	return msg[start:]
}

func getRawTTLByTopic(topic string) time.Duration {
	switch {
	case strings.HasPrefix(topic, "kline.1year"):
		return 24 * time.Hour
	case strings.HasPrefix(topic, "kline.1mon"):
		return 6 * time.Hour
	case strings.HasPrefix(topic, "kline.1week"):
		return 3 * time.Hour
	case strings.HasPrefix(topic, "kline.1day"):
		return 1 * time.Hour
	case strings.HasPrefix(topic, "kline.4hour"):
		return 30 * time.Minute
	case strings.HasPrefix(topic, "kline.60min"):
		return 20 * time.Minute
	case strings.HasPrefix(topic, "kline.30min"):
		return 15 * time.Minute
	case strings.HasPrefix(topic, "kline.15min"):
		return 10 * time.Minute
	case strings.HasPrefix(topic, "kline.5min"):
		return 5 * time.Minute
	case strings.HasPrefix(topic, "kline.1min"):
		return 1 * time.Minute
	default:
		return 60 * time.Second
	}
}
