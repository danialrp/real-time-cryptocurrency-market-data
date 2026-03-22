package markup

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"irbtc_streamer/internal/redis"
	"irbtc_streamer/internal/schema"
	"irbtc_streamer/internal/stream"
)

const (
	usdtIrtRedisKey = "irbtc:price.rater:usdtirt" // Select source (rater or core)
)

type USDTIRT struct {
	SellRate string `json:"sell_rate"`
	BuyRate  string `json:"buy_rate"`
}

func Convert(symbol string, topic string, rawTick []byte) {
	if !strings.HasSuffix(symbol, "usdt") {
		return
	}

	if !isSupportedIRTTopic(topic) && !isKlineTopic(topic) && topic != "trade.detail" {
		return
	}

	usdtRaw, err := redis.ClientGeneral.Get(context.Background(), usdtIrtRedisKey).Result()
	if err != nil {
		log.Printf("❌ usdtirt not available: %v", err)
		return
	}

	var usdt USDTIRT
	if err := json.Unmarshal([]byte(usdtRaw), &usdt); err != nil {
		log.Printf("❌ failed to unmarshal usdtirt: %v", err)
		return
	}

	// askRate, _ := strconv.ParseFloat(usdt.SellRate, 64)
	// USIN SAME USDT RATE FOR PREVENTING MODIFY SPEREADS
	askRate, _ := strconv.ParseFloat(usdt.BuyRate, 64)
	bidRate, _ := strconv.ParseFloat(usdt.BuyRate, 64)

	base := strings.TrimSuffix(symbol, "usdt")
	key := fmt.Sprintf("irbtc:live:%s:%sirt", topic, base)
	ctx := context.Background()
	ts := time.Now().UnixMilli()
	ttl := getTTLByTopic(topic)

	var tick stream.StandardTick

	switch {
	case topic == "bbo":
		var bbo schema.BBO
		if err := json.Unmarshal(rawTick, &bbo); err != nil {
			return
		}
		tick = stream.StandardTick{
			Ask:       bbo.Ask * askRate,
			AskSize:   bbo.AskSize,
			Bid:       bbo.Bid * bidRate,
			BidSize:   bbo.BidSize,
			LastPrice: bbo.Ask * askRate,
			Ts:        ts,
		}

	case topic == "ticker":
		var t schema.Ticker
		if err := json.Unmarshal(rawTick, &t); err != nil {
			return
		}
		tick = stream.StandardTick{
			Ask:       t.Ask * askRate,
			AskSize:   t.AskSize,
			Bid:       t.Bid * bidRate,
			BidSize:   t.BidSize,
			LastPrice: t.LastPrice * askRate,
			LastSize:  t.LastSize,
			Amount:    t.Amount,
			Close:     t.Close * bidRate,
			Open:      t.Open * askRate,
			High:      t.High * askRate,
			Low:       t.Low * bidRate,
			Count:     t.Count,
			Vol:       t.Vol * askRate,
			Ts:        ts,
		}

	case topic == "depth.step0":
		var d schema.Depth
		if err := json.Unmarshal(rawTick, &d); err != nil {
			return
		}
		tick = stream.StandardTick{
			Bids: convertLevels(d.Bids, bidRate),
			Asks: convertLevels(d.Asks, askRate),
			Ts:   ts,
		}

	case isKlineTopic(topic):
		var k schema.Kline
		if err := json.Unmarshal(rawTick, &k); err != nil {
			return
		}
		tick = stream.StandardTick{
			Amount: k.Amount,
			Close:  k.Close * bidRate,
			Count:  k.Count,
			High:   k.High * askRate,
			Low:    k.Low * bidRate,
			Open:   k.Open * askRate,
			Vol:    k.Vol * askRate,
			Ts:     ts,
		}

	case topic == "trade.detail":
		var raw schema.TradeList
		if err := json.Unmarshal(rawTick, &raw); err != nil {
			return
		}

		trades := make([]stream.TradeItem, len(raw.Data))
		for i, t := range raw.Data {
			trades[i] = stream.TradeItem{
				Price:     t.Price * askRate,
				Amount:    t.Amount,
				Direction: t.Direction,
			}
		}

		tick = stream.StandardTick{
			Trades: trades,
			Ts:     time.Now().UnixMilli(),
		}
	}

	data, _ := json.Marshal(tick)
	_ = redis.ClientStreamIRT.Set(ctx, key, data, ttl).Err()
	_ = redis.ClientStreamIRT.Publish(ctx, key, data).Err()
}

func convertLevels(levels [][]float64, rate float64) [][]float64 {
	var converted [][]float64
	for _, level := range levels {
		if len(level) >= 2 {
			converted = append(converted, []float64{
				level[0] * rate,
				level[1],
			})
		}
	}
	return converted
}

func isSupportedIRTTopic(topic string) bool {
	supported := []string{
		"bbo", "ticker", "depth.step0", "trade.detail",
	}
	for _, ch := range supported {
		if ch == topic {
			return true
		}
	}
	return false
}

func isKlineTopic(topic string) bool {
	return strings.HasPrefix(topic, "kline.")
}

func getTTLByTopic(topic string) time.Duration {
	switch topic {
	case "bbo", "ticker", "depth.step0", "trade.detail":
		return 60 * time.Second
	case "kline.1min":
		return 2 * time.Minute
	case "kline.5min":
		return 10 * time.Minute
	case "kline.15min":
		return 30 * time.Minute
	case "kline.30min":
		return 1 * time.Hour
	case "kline.60min", "kline.1hour":
		return 2 * time.Hour
	case "kline.4hour":
		return 4 * time.Hour
	case "kline.1day":
		return 24 * time.Hour
	case "kline.1week":
		return 7 * 24 * time.Hour
	case "kline.1mon":
		return 30 * 24 * time.Hour
	case "kline.1year":
		return 365 * 24 * time.Hour
	default:
		return 1 * time.Minute
	}
}
