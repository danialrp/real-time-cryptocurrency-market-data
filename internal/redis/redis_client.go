package redis

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

var (
	ClientGeneral   *redis.Client
	ClientStreamRaw *redis.Client
	ClientStreamIRT *redis.Client
	ClientLimiter *redis.Client
)

var Subscriber *redis.Client

func InitRedis() {
	_ = godotenv.Load()

	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	password := os.Getenv("REDIS_PASSWORD")

	dbGeneral := mustAtoi(os.Getenv("REDIS_DB_GENERAL"))
	dbRaw := mustAtoi(os.Getenv("REDIS_DB_STREAM_RAW"))
	dbIRT := mustAtoi(os.Getenv("REDIS_DB_STREAM_IRT"))
	dbLimiter := mustAtoi(os.Getenv("REDIS_DB_LIMITER"))

	ClientGeneral = newRedisClient(host, port, password, dbGeneral)
	ClientStreamRaw = newRedisClient(host, port, password, dbRaw)
	ClientStreamIRT = newRedisClient(host, port, password, dbIRT)
	ClientLimiter = newRedisClient(host, port, password, dbLimiter)

	Subscriber = newRedisClient(host, port, password, dbRaw)

	log.Println("✅ Redis clients initialized.")
}

func newRedisClient(host, port, password string, db int) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       db,
	})

	if _, err := client.Ping(context.Background()).Result(); err != nil {
		log.Fatalf("❌ Redis DB %d connection failed: %v", db, err)
	}

	return client
}

func mustAtoi(val string) int {
	i, err := strconv.Atoi(val)
	if err != nil {
		log.Fatalf("❌ Invalid Redis DB number: %s", val)
	}
	return i
}

func PublishRawStreamWithTTL(symbol string, topic string, payload []byte, ttl time.Duration) error {
	key := fmt.Sprintf("irbtc:live:%s:%s", topic, symbol)

	err := ClientStreamRaw.Set(context.Background(), key, payload, ttl).Err()
	if err != nil {
		return err
	}

	// Publish to Redis pub/sub channel
	_ = ClientStreamRaw.Publish(context.Background(), key, payload).Err()

	return nil
}

func PublishIrtStreamWithTTL(symbol string, topic string, payload []byte, ttl time.Duration) error {
	key := fmt.Sprintf("irbtc:live:%s:%s", topic, symbol)

	err := ClientStreamIRT.Set(context.Background(), key, payload, ttl).Err()
	if err != nil {
		return err
	}

	// Publish to Redis pub/sub channel
	_ = ClientStreamIRT.Publish(context.Background(), key, payload).Err()

	return nil
}

func GetRedisSubscriber() *redis.Client {
	return Subscriber
}

func GetClientBySource(source string) *redis.Client {
	switch strings.ToLower(source) {
	case "irt":
		return ClientStreamIRT
	case "raw":
		return ClientStreamRaw
	default:
		return ClientStreamIRT
	}
}
