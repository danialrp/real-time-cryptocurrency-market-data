package repository

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"irbtc_streamer/database/table"
	"irbtc_streamer/internal/contract"
	"irbtc_streamer/internal/model"
)

type symbolRepository struct {
	db *gorm.DB
}

func NewSymbolRepository(db *gorm.DB) contract.SymbolRepository {
	return &symbolRepository{db: db}
}

func (r *symbolRepository) UpdateAll(ctx context.Context, symbols []model.Symbol) error {
	tx := r.db.WithContext(ctx).Begin()

	for _, s := range symbols {
		entity := table.HTXSymbolTable{
			Symbol:        s.Symbol,
			BaseCurrency:  s.BaseCurrency,
			QuoteCurrency: s.QuoteCurrency,
			State:         s.State,
			TradeEnabled:  s.TradeEnabled,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "symbol"}}, // باید Unique Index روی این ستون باشه
			DoUpdates: clause.AssignmentColumns([]string{"base_currency", "quote_currency", "updated_at"}),
		}).Create(&entity).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (r *symbolRepository) GetAll(ctx context.Context) ([]model.Symbol, error) {
	var entities []table.HTXSymbolTable
	if err := r.db.WithContext(ctx).
		Where("state = ? AND trade_enabled = ?", "online", true).
		Find(&entities).Error; err != nil {
		return nil, err
	}

	var result []model.Symbol
	for _, e := range entities {
		result = append(result, model.Symbol{
			Symbol:        e.Symbol,
			BaseCurrency:  e.BaseCurrency,
			QuoteCurrency: e.QuoteCurrency,
			State:         e.State,
			TradeEnabled:  e.TradeEnabled,
		})
	}

	return result, nil
}

func (r *symbolRepository) GetByQuoteCurrency(ctx context.Context, quote string) ([]model.Symbol, error) {
	var entities []table.HTXSymbolTable

	if err := r.db.WithContext(ctx).
		Where("state = ? AND trade_enabled = ? AND quote_currency = ?", "online", true, strings.ToLower(quote)).
		Find(&entities).Error; err != nil {
		return nil, err
	}

	var result []model.Symbol
	for _, e := range entities {
		result = append(result, model.Symbol{
			Symbol:        e.Symbol,
			BaseCurrency:  e.BaseCurrency,
			QuoteCurrency: e.QuoteCurrency,
			State:         e.State,
			TradeEnabled:  e.TradeEnabled,
		})
	}

	return result, nil
}
