package redis

import (
	"context"
	"fmt"
	"time"
)

// incrementWithTTL tries to increase a counter and sets an expiration
func incrementWithTTL(key string, limit int, ttl time.Duration) (bool, int, error) {
	ctx := context.Background()

	pipe := ClientLimiter.TxPipeline()

	count := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, ttl)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, err
	}

	val := int(count.Val())
	return val <= limit, val, nil
}

// DecrementKey reduces the count manually, e.g. when connection closes
func DecrementKey(key string) error {
	ctx := context.Background()

	script := `
local current = redis.call("GET", KEYS[1])
if current and tonumber(current) > 0 then
	return redis.call("DECR", KEYS[1])
else
	return 0
end
`
	return ClientLimiter.Eval(ctx, script, []string{key}).Err()
}

// AllowConnection checks if IP is allowed to establish new WebSocket connection
func AllowConnection(ip string, limit int) (bool, int, error) {
	key := fmt.Sprintf("limit:conn:%s", ip)
	return incrementWithTTL(key, limit, 10*time.Minute)
}

// AllowSubscription checks if IP is allowed to subscribe to more topics
func AllowSubscription(ip string, limit int) (bool, int, error) {
	key := fmt.Sprintf("limit:sub:%s", ip)
	return incrementWithTTL(key, limit, 10*time.Minute)
}

func AllowSubscriptionBulk(ip string, limit int, count int) (bool, int, error) {
	key := fmt.Sprintf("limit:sub:%s", ip)

	ctx := context.Background()
	pipe := ClientLimiter.TxPipeline()

	// current := pipe.Get(ctx, key)
	incr := pipe.IncrBy(ctx, key, int64(count))
	pipe.Expire(ctx, key, 10*time.Minute)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, err
	}

	// currentVal, _ := current.Int()
	newVal := int(incr.Val())
	return newVal <= limit, newVal, nil
}

func DeleteKey(key string) error {
	return ClientLimiter.Del(context.Background(), key).Err()
}
