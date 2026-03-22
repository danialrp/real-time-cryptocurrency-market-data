package repository

import (
	"gorm.io/gorm"
	"irbtc_streamer/database/table"
)

func InsertPriceChanges(db *gorm.DB, symbol string, changePercentage float64, startPrice, endPrice float64, changeType string) {
	priceChanges := table.PriceChangesTable{
		Symbol:           symbol,
		ChangePercentage: changePercentage,
		StartPrice:       startPrice,
		EndPrice:         endPrice,
		ChangeType:       changeType,
	}
	db.Create(&priceChanges)
}
