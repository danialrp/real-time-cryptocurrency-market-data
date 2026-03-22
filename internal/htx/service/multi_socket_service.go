package service

import (
	"context"
	"log"
	"strings"
	"time"

	"irbtc_streamer/internal/htx"
	"irbtc_streamer/internal/model"
	"irbtc_streamer/internal/stream"
	"irbtc_streamer/internal/utils"
)

func StartHTXMultiConnectionService(ctx context.Context, allSymbols []model.Symbol, dispatcher *stream.Dispatcher) ([]*htx.WebSocketClient, error) {
	chunkSize := utils.EnvInt("SOCKET_CHUNK_SIZE", 40)
	rawTopics := utils.EnvString("SOCKET_TOPICS_HTX", "depth.step0,trade.detail,bbo,ticker")
	topics := parseTopics(rawTopics)

	chunks := chunkSymbols(allSymbols, chunkSize)

	log.Printf("📦 HTX Multi-Connection Setup | Total Symbols: %d | Chunk Size: %d | Total Connections: %d",
		len(allSymbols), chunkSize, len(chunks))

	var clients []*htx.WebSocketClient

	for i, symbolsChunk := range chunks {
		// log.Printf("🧵 Spawning connection HTX #%d with %d symbols", i, len(symbolsChunk))

		time.Sleep(200 * time.Millisecond) // delay to avoid concurrent connect burst

		symbolsCopy := symbolsChunk // capture for closure
		go func(idx int) {
			connCtx, cancel := context.WithCancel(ctx)
			defer cancel()

			// log.Printf("⚙️  [HTX #%d] Initializing goroutine with %d symbols and %d topics", idx, len(symbolsCopy), len(topics))

			var client *htx.WebSocketClient
			var err error

			for retry := 0; retry < 5; retry++ {
				client, err = StartHTXWebSocketService(connCtx, symbolsCopy, dispatcher, idx, topics)
				if err == nil {
					break
				}

				log.Printf("⏳ [HTX #%d] Retry %d after error: %v", idx, retry+1, err)
				time.Sleep(5 * time.Second)
			}

			if err != nil {
				log.Printf("❌ [HTX #%d] Failed to connect after retries: %v", idx, err)
				return
			}

			// log.Printf("✅ [HTX #%d] Connected with %d symbols", idx, len(symbolsCopy))
			clients = append(clients, client)
		}(i)
	}

	log.Println(utils.GenerateSubscribeLog(len(allSymbols), len(topics), chunkSize))

	return clients, nil
}

func chunkSymbols(symbols []model.Symbol, size int) [][]model.Symbol {
	var chunks [][]model.Symbol
	for size < len(symbols) {
		symbols, chunks = symbols[size:], append(chunks, symbols[0:size:size])
	}
	return append(chunks, symbols)
}

func parseTopics(s string) []string {
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
