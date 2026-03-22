package service

import (
	"log"
	"context"

	"github.com/gofiber/fiber/v2"
	"irbtc_streamer/internal/router"
	"irbtc_streamer/internal/ws"
	"irbtc_streamer/internal/stream"
)

func StartFiberServer(ctx context.Context, dispatcher *stream.Dispatcher) {
	app := fiber.New()

	wsHandler := ws.NewWebSocketHandler(dispatcher)
	router.RegisterWebSocketHandler(wsHandler)

	router.SetupRoutes(app)

	if err := app.Listen(":9090"); err != nil {
		log.Printf("❌ Failed to start Fiber server: %v", err)
	}

	log.Println("🚀 Server starting on :9090...")
	if err := app.Listen(":9090"); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
