package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Order represents a trading order
type Order struct {
	ID              string          `gorm:"primaryKey;type:uuid;not null" json:"id"`
	StrategyID      string          `gorm:"type:uuid;index;not null" json:"strategyId"`
	Symbol          string          `gorm:"type:varchar(32);not null" json:"symbol"`
	Side            string          `gorm:"type:varchar(8);not null" json:"side"`         // buy/sell
	Type            string          `gorm:"type:varchar(8);not null" json:"type"`         // limit/market
	Price           decimal.Decimal `gorm:"type:numeric(20,8)" json:"price"`
	Quantity        decimal.Decimal `gorm:"type:numeric(20,8)" json:"quantity"`
	FilledQuantity  decimal.Decimal `gorm:"type:numeric(20,8);default:0" json:"filledQuantity"`
	Status          string          `gorm:"type:varchar(16);not null" json:"status"` // open/filled/canceled/liquidated
	TimeInForce     string          `gorm:"type:varchar(8);not null" json:"timeInForce"` // GTC/IOC/FOK
	TTL             int             `gorm:"default:0" json:"ttl"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
}

// TableName returns the table name for Order
func (Order) TableName() string {
	return "orders"
}
