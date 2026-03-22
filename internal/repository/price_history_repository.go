package repository

import (
	"gorm.io/gorm"
	"irbtc_streamer/database/table"
)

func InsertPriceHistory(db *gorm.DB, symbol string, price float64) {
	priceHistory := table.PriceHistoryTable{
		Symbol: symbol,
		Price:  price,
	}
	db.Create(&priceHistory)
}
