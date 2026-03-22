package utils

import (
	"fmt"
	"math"
)

func GenerateSubscribeLog(symbolCount, topicCount, chunkSize int) string {
	totalSubs := symbolCount * topicCount
	subsPerConn := chunkSize * topicCount
	numConnections := int(math.Ceil(float64(symbolCount) / float64(chunkSize)))

	return fmt.Sprintf(
		"🔢 %d symbol + %d topic = %d subscribe ✅ chunk size: %d → %d connection × %d sub",
		symbolCount, topicCount, totalSubs, chunkSize, numConnections, subsPerConn,
	)
}
