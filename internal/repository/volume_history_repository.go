package repository

import (
	"gorm.io/gorm"
	"irbtc_streamer/database/table"
)

func InsertVolumeHistory(db *gorm.DB, symbol string, volume float64) {
	volumeHistory := table.VolumeHistoryTable{
		Symbol: symbol,
		Volume: volume,
	}
	db.Create(&volumeHistory)
}
