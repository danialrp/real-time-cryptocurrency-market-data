// مسیر: cmd/runtime/main.go

package main

import (
	"context"
	"log"
	"sync"

	"irbtc_streamer/database"
	htxservice "irbtc_streamer/internal/htx/service"
	"irbtc_streamer/internal/redis"
	"irbtc_streamer/internal/repository"
	"irbtc_streamer/internal/scheduler"
	"irbtc_streamer/internal/service"
	"irbtc_streamer/internal/stream"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("❌ GLOBAL PANIC: %v", r)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	database.CreateOrConnectDatabase()

	db, err := database.NewPostgresConnection()
	if err != nil {
		log.Fatalln("❌ DB connect error:", err)
	}

	redis.InitRedis()

	dispatcher := stream.NewDispatcher()

	symbolRepo := repository.NewSymbolRepository(db)

	updater := scheduler.NewSymbolUpdaterService(symbolRepo)
	updater.RunOnce()

	symbols, err := symbolRepo.GetAll(ctx)
	if err != nil {
		log.Fatal("❌ Error fetching symbols:", err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := htxservice.StartHTXMultiConnectionService(ctx, symbols, dispatcher)
		if err != nil {
			log.Println("❌ HTX Multi-Connection error:", err)
			return
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		service.StartFiberServer(ctx, dispatcher)
	}()

	<-ctx.Done()
	wg.Wait()

	log.Println("🛑 Server shutdown.")
}
