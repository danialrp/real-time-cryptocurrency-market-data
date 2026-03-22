package table

import "time"

type PriceHistoryTable struct {
	ID        int64     `gorm:"primaryKey"`
	Symbol    string    `gorm:"not null"`
	Price     float64   `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
