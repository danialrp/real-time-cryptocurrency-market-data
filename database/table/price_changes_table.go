package table

import "time"

type PriceChangesTable struct {
	ID                int64     `gorm:"primaryKey"`
	Symbol            string    `gorm:"not null"`
	ChangePercentage  float64   `gorm:"not null"`
	StartPrice        float64   `gorm:"not null"`
	EndPrice          float64   `gorm:"not null"`
	ChangeType        string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
