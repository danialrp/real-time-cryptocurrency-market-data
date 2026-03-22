package service

import (
	"context"
	"log"
	"time"

	"irbtc_streamer/internal/htx"
	"irbtc_streamer/internal/model"
	"irbtc_streamer/internal/stream"
)

func StartHTXWebSocketService(ctx context.Context, symbols []model.Symbol, dispatcher *stream.Dispatcher, id int, topics []string) (*htx.WebSocketClient, error) {
	client, err := htx.NewWebSocketClient(ctx, dispatcher, id)
	if err != nil {
		log.Println("❌ WebSocket client error:", err)
		return nil, err
	}

	go func() {
		for {
			// log.Printf("▶️ [HTX #%d] Starting WebSocket client...", id)

			client.AutoResubscribe(symbols, topics)

			err := client.Listen()
			if err != nil {
				log.Printf("❌ [HTX #%d] Connection dropped: %v", id, err)
			}

			log.Printf("🔌 [HTX #%d] Waiting 3s before reconnect...", id)
			time.Sleep(3 * time.Second)

			if err := client.Reconnect(ctx); err != nil {
				log.Printf("❌ [HTX #%d] Reconnect failed: %v", id, err)
				time.Sleep(5 * time.Second)
				continue
			}

			log.Printf("✅ [HTX #%d] Reconnected, resubscribing...", id)
		}
	}()

	return client, nil
}
