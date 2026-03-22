package table

import "time"

type VolumeHistoryTable struct {
	ID        int64     `gorm:"primaryKey"`
	Symbol    string    `gorm:"not null"`
	Volume    float64   `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
