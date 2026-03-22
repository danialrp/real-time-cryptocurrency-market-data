package main

import (
	"log"
	"time"

	"irbtc_streamer/database"
	"irbtc_streamer/internal/redis"
	"irbtc_streamer/internal/worker"
)

func main() {
	redis.InitRedis()

	database.CreateOrConnectDatabase()

	// db, err := database.NewPostgresConnection()
	// if err != nil {
	// 	log.Fatalf("❌ Failed to connect to database: %v", err)
	// }

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		// Fetch USDT/IRT from irbtc-core
		errc := worker.FetchAndStoreUSDTPriceFromIrbtc()
		if errc != nil {
			log.Printf("❌ USDT price from Irbtc fetch error: %v", errc)
		}

		// Fetch USDT/IRT from irbtc-rater
		errr := worker.FetchAndStoreUSDTPriceFromRater()
		if errr != nil {
			log.Printf("❌ USDT price from Rater fetch error: %v", errr)
		}

		<-ticker.C
	}
}
