package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"irbtc_streamer/internal/ws"
)

// var wsHandler *ws.WebSocketHandler

func RegisterWebSocketHandler(handler *ws.WebSocketHandler) {
	// wsHandler = handler
}

func SetupRoutes(app *fiber.App) {
	// Health Check Route
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// API Routes
	//

	// Traditional WebSocket Routes
	// 

	
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	
	// WebSocket multi-subscribe endpoint
	app.Get("/ws", websocket.New(ws.HandleConnection))
}
