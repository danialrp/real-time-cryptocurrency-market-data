package contract

import ()

// MarketDataProvider is the general interface for any exchange WebSocket client.
type MarketWebSocketClient interface {
	ConnectAndListen() error
	Subscribe(topic string)
}