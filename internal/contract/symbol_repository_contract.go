package contract

import (
	"context"
	
	"irbtc_streamer/internal/model"
)

type SymbolRepository interface {
	UpdateAll(ctx context.Context, symbols []model.Symbol) error
	GetAll(ctx context.Context) ([]model.Symbol, error)
	GetByQuoteCurrency(ctx context.Context, quote string) ([]model.Symbol, error)
}
