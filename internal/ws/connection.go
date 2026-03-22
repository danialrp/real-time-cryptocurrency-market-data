package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	goredis "github.com/go-redis/redis/v8"
	"github.com/gofiber/websocket/v2"
	"irbtc_streamer/internal/redis"
	"irbtc_streamer/internal/stream"
)

type Subscription struct {
	Symbol string `json:"symbol"`
	Topic  string `json:"topic"`
}

type SubscribeMessage struct {
	Op   string         `json:"op"`   // "subscribe", "unsubscribe"
	Args []Subscription `json:"args"` // list of channels
}

type Connection struct {
	Conn          *websocket.Conn
	Subscriptions map[string]context.CancelFunc
	WriteChan     chan []byte
}

func NewConnection(conn *websocket.Conn) *Connection {
	return &Connection{
		Conn:          conn,
		Subscriptions: make(map[string]context.CancelFunc),
		WriteChan:     make(chan []byte, 2000),
	}
}

func (c *Connection) Listen() {
	// Log pong handler
	c.Conn.SetPongHandler(func(appData string) error {
		// log.Printf("📶 pong received from client: %s", c.Conn.RemoteAddr())
		return nil
	})

	remoteIP := strings.Split(c.Conn.RemoteAddr().String(), ":")[0]

	// Limit max 3 connections per IP
	allowed, count, err := redis.AllowConnection(remoteIP, 3)
	if err != nil || !allowed {
		errMsg := fmt.Sprintf(`{"error":"connection limit exceeded (count: %d)"}`, count)
		_ = c.Conn.WriteMessage(websocket.TextMessage, []byte(errMsg))
		// log.Printf("❌ too many connections from IP %s", remoteIP)
		_ = c.Conn.Close()
		return
	}

	// Writer Goroutine (safe against nil Conn)
	go func() {
		for msg := range c.WriteChan {
			if c.Conn == nil {
				log.Println("⚠️ Conn is nil, skipping write")
				break
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Printf("❌ write error: %v", err)
				break
			}
		}
	}()

	// Ping check
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if c.Conn == nil {
				return
			}
			// Send actual WebSocket ping
			err := c.Conn.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(5*time.Second))
			if err != nil {
				// log.Printf("❌ Ping failed, closing connection: %v", err)
				_ = c.Conn.Close()
				return
			}
		}
	}()

	// Cleanup Conn & Subscriptions after disconnect
	defer func() {
		for _, cancel := range c.Subscriptions {
			cancel()
		}
		c.Conn = nil
		_ = redis.DecrementKey(fmt.Sprintf("limit:conn:%s", remoteIP))
		_ = redis.DeleteKey(fmt.Sprintf("limit:sub:%s", remoteIP))
		close(c.WriteChan)
		// log.Println("🧹 connection cleanup done")
	}()

	// Listener for incoming subscription/unsubscription messages
	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			if !strings.Contains(err.Error(), "close 1000") {
				log.Printf("❌ abnormal WebSocket close: %v", err)
			}
			break
		}

		var sm SubscribeMessage
		if err := json.Unmarshal(msg, &sm); err != nil {
			log.Printf("❌ invalid JSON: %v", err)
			continue
		}

		switch sm.Op {
		case "subscribe":
			// Limit max 20 subscriptions per IP
			requested := len(sm.Args)
			allowedSub, subCount, err := redis.AllowSubscriptionBulk(remoteIP, 20, requested)
			if err != nil || !allowedSub {
				errMsg := fmt.Sprintf(`{"error":"subscription limit exceeded (count: %d)"}`, subCount)
				_ = c.Conn.WriteMessage(websocket.TextMessage, []byte(errMsg))
				// log.Printf("❌ too many subscriptions from IP %s", remoteIP)
				break
			}

			for _, sub := range sm.Args {
				key := fmt.Sprintf("irbtc:live:%s:%s", sub.Topic, sub.Symbol)

				// Validation for allowed topic
				isValid, errMsg := isValidTopic(sub.Topic, sub.Symbol)
				if !isValid {
					_ = c.Conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"error":"%s"}`, errMsg)))
					// log.Printf("❌ rejected topic: %s", sub.Topic)
					continue
				}

				if _, exists := c.Subscriptions[key]; exists {
					continue
				}

				ctx, cancel := context.WithCancel(context.Background())
				c.Subscriptions[key] = cancel
				go c.listenToChannel(ctx, key)
			}

		case "unsubscribe":
			for _, sub := range sm.Args {
				key := fmt.Sprintf("irbtc:live:%s:%s", sub.Topic, sub.Symbol)
				if cancel, exists := c.Subscriptions[key]; exists {
					cancel()
					delete(c.Subscriptions, key)
					_ = redis.DecrementKey(fmt.Sprintf("limit:sub:%s", remoteIP))
				}
			}
		default:
			log.Printf("⚠️ unknown op: %s", sm.Op)
		}
	}
}

func (c *Connection) listenToChannel(ctx context.Context, redisKey string) {
	parts := strings.Split(redisKey, ":")
	if len(parts) < 4 {
		log.Printf("⚠️ invalid redis key: %s", redisKey)
		return
	}
	symbol := parts[3]

	client := redis.ClientStreamRaw
	if strings.HasSuffix(symbol, "irt") {
		client = redis.ClientStreamIRT
	}

	pubsub := client.Subscribe(ctx, redisKey)
	ch := pubsub.ChannelWithSubscriptions(ctx, 2000)

	// log.Printf("📡 subscribed to [%s]", redisKey)

	for {
		select {
		case <-ctx.Done():
			_ = pubsub.Close()
			// log.Printf("🛑 unsubscribed from [%s]", redisKey)
			return

		case m := <-ch:
			if msg, ok := m.(*goredis.Message); ok {
				select {
				case c.WriteChan <- []byte(msg.Payload):
				default:
					log.Printf("⚠️ WriteChan full for client [%s]", redisKey)
				}
			}
		}
	}
}

func isValidTopic(topic, symbol string) (bool, string) {
	if strings.HasSuffix(symbol, "irt") {
		if !stream.IsSupportedIRTTopic(topic) {
			return false, fmt.Sprintf("unsupported IRT topic: %s", topic)
		}
	} else {
		if !stream.IsSupportedHTXChannels(topic) {
			return false, fmt.Sprintf("unsupported RAW topic: %s", topic)
		}
	}
	return true, ""
}
