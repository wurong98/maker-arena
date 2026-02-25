package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Ticker represents market ticker data
type Ticker struct {
	Symbol                string          `gorm:"primaryKey;type:varchar(32)" json:"symbol"`
	Price                 decimal.Decimal `gorm:"type:numeric(20,8)" json:"price"`
	PriceChange24h        decimal.Decimal `gorm:"type:numeric(20,8)" json:"priceChange24h"`
	PriceChangePercent24h decimal.Decimal `gorm:"type:numeric(20,8)" json:"priceChangePercent24h"`
	UpdatedAt             time.Time       `json:"updatedAt"`
}

// TableName returns the table name for Ticker
func (Ticker) TableName() string {
	return "tickers"
}
