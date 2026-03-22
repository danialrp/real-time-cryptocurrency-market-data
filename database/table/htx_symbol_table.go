package table

import "time"

type HTXSymbolTable struct {
	ID            uint      `gorm:"primaryKey"`
	Symbol        string    `gorm:"uniqueIndex"`
	BaseCurrency  string
	QuoteCurrency string
	State         string    `gorm:"type:varchar(20);index"`
	TradeEnabled  bool      `gorm:"type:boolean;default:false"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (HTXSymbolTable) TableName() string {
	return "htx_symbols"
}
