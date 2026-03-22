package main

import (
	"log"
	"time"

	"irbtc_streamer/database"
	"irbtc_streamer/internal/repository"
	"irbtc_streamer/internal/scheduler"
)

func main() {
	database.CreateOrConnectDatabase()
	db, err := database.NewPostgresConnection()
	if err != nil {
		log.Fatalln("Failed to connect to DB:", err)
	}

	symbolRepo := repository.NewSymbolRepository(db)

	updater := scheduler.NewSymbolUpdaterService(symbolRepo)
	updater.RunOnce()

	ticker := time.NewTicker(24 * time.Hour)
	for range ticker.C {
		updater.RunOnce()
	}

	//!TODO: Remove this
	// for {
	//     time.Sleep(5 * time.Hour)
	// }
}
