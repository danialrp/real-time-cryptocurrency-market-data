package stream

import (
	"os"
	"strings"
)

var SupportedHTXChannels []string // Channels supported for HTX markets (raw data from HTX WebSocket)
var SupportedIRTChannels []string // Channels supported for IRT-marked up markets

func init() {
	raw := os.Getenv("SOCKET_TOPICS_HTX")
	if raw == "" {
		// fallback default channels
		SupportedHTXChannels = []string{
			"kline.1min",
			"kline.5min",
			"kline.15min",
			"kline.30min",
			"kline.60min",
			"kline.4hour",
			"kline.1day",
			"kline.1week",
			"kline.1mon",
			"kline.1year",
			"depth.step0",
			"trade.detail",
			"bbo",
			"ticker",
		}
	} else {
		SupportedHTXChannels = strings.Split(raw, ",")
	}

	// IRT Channels
	rawIRT := os.Getenv("SOCKET_TOPICS_IRT")
	if rawIRT == "" {
		SupportedIRTChannels = []string{
			"kline.1min",
			"kline.5min",
			"kline.15min",
			"kline.30min",
			"kline.60min",
			"kline.4hour",
			"kline.1day",
			"kline.1week",
			"kline.1mon",
			"kline.1year",
			"depth.step0",
			"trade.detail",
			"bbo",
			"ticker",
		}
	} else {
		SupportedIRTChannels = strings.Split(rawIRT, ",")
	}
}

func IsSupportedHTXChannels(topic string) bool {
	for _, ch := range SupportedHTXChannels {
		if ch == topic {
			return true
		}
	}
	return false
}

func IsSupportedIRTTopic(topic string) bool {
	for _, ch := range SupportedIRTChannels {
		if ch == topic {
			return true
		}
	}
	return false
}
