package scheduler

import (
	"context"
	"log"
	"time"

	"irbtc_streamer/internal/contract"
	"irbtc_streamer/internal/htx"
)

type SymbolUpdaterService struct {
	SymbolRepo contract.SymbolRepository
}

func NewSymbolUpdaterService(repo contract.SymbolRepository) *SymbolUpdaterService {
	return &SymbolUpdaterService{SymbolRepo: repo}
}

func (s *SymbolUpdaterService) Start() {
	ticker := time.NewTicker(24 * time.Hour)

	go func() {
		s.RunOnce()
		for range ticker.C {
			s.RunOnce()
		}
	}()
}

func (s *SymbolUpdaterService) RunOnce() {
	ctx := context.Background()

	symbols, err := htx.FetchSymbols(ctx)
	if err != nil {
		log.Println("❌ Error fetching HTX symbols:", err)
		return
	}

	log.Printf("🔍 Total insertable: %d symbols", len(symbols))

	err = s.SymbolRepo.UpdateAll(ctx, symbols)
	if err != nil {
		log.Println("❌ Error updating HTX symbols:", err)
	}
}


