package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Fill represents a trade fill/execution
type Fill struct {
	ID         string          `gorm:"primaryKey;type:uuid" json:"id"`
	OrderID    string          `gorm:"type:uuid;index" json:"orderId"`
	StrategyID string          `gorm:"type:uuid;index" json:"strategyId"`
	Symbol     string          `gorm:"type:varchar(32)" json:"symbol"`
	Side       string          `gorm:"type:varchar(8)" json:"side"`
	Price      decimal.Decimal `gorm:"type:numeric(20,8)" json:"price"`
	Quantity   decimal.Decimal `gorm:"type:numeric(20,8)" json:"quantity"`
	Fee        decimal.Decimal `gorm:"type:numeric(20,8)" json:"fee"`
	CreatedAt  time.Time       `json:"createdAt"`
}

// TableName returns the table name for Fill
func (Fill) TableName() string {
	return "fills"
}
