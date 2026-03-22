package ws

import (
	"log"

	"github.com/gofiber/websocket/v2"
	"irbtc_streamer/internal/stream"
)

type WebSocketHandler struct {
	Dispatcher *stream.Dispatcher
}

func NewWebSocketHandler(dispatcher *stream.Dispatcher) *WebSocketHandler {
	return &WebSocketHandler{
		Dispatcher: dispatcher,
	}
}

// WebSocket endpoint handler
func (h *WebSocketHandler) HandleConnection(c *websocket.Conn) {
	defer c.Close()

	// Read query parameters
	symbol := c.Query("symbol")
	topic := c.Query("topic")

	if symbol == "" || topic == "" {
		errMsg := "Missing 'symbol' or 'topic' parameter"
		log.Println("❌", errMsg)
		if err := c.WriteMessage(websocket.TextMessage, []byte(errMsg)); err != nil {
			log.Printf("❌ Error writing error message to WebSocket: %v", err)
		}

		return
	}

	// Subscribe to the topic and symbol
	clientCh := h.Dispatcher.Subscribe(topic, symbol)
	defer h.Dispatcher.Unsubscribe(topic, symbol, clientCh)

	log.Printf("✅ New subscriber: %s:%s", topic, symbol)

	for msg := range clientCh {
		if err := c.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Printf("❌ Error writing to WebSocket: %v", err)
			return
		}
	}
}

func HandleConnection(c *websocket.Conn) {
	conn := NewConnection(c)
	conn.Listen()
}
