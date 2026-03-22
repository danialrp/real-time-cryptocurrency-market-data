package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

func PublishRawStream(symbol string, topic string, message interface{}) error {
	key := fmt.Sprintf("irbtc:live:%s:%s", topic, symbol)

	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return ClientStreamRaw.Set(context.Background(), key, payload, 60*time.Second).Err()
}
