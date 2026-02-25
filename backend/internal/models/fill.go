package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Fill represents a trade fill/execution
type Fill struct {
	ID         string          `gorm:"primaryKey;type:uuid;not null" json:"id"`
	OrderID    string          `gorm:"type:uuid;index;not null" json:"orderId"`
	StrategyID string          `gorm:"type:uuid;index;not null" json:"strategyId"`
	Symbol     string          `gorm:"type:varchar(32);not null" json:"symbol"`
	Side       string          `gorm:"type:varchar(8);not null" json:"side"`
	Price      decimal.Decimal `gorm:"type:numeric(20,8)" json:"price"`
	Quantity   decimal.Decimal `gorm:"type:numeric(20,8)" json:"quantity"`
	Fee        decimal.Decimal `gorm:"type:numeric(20,8)" json:"fee"`
	CreatedAt  time.Time       `json:"createdAt"`
	UpdatedAt  time.Time       `json:"updatedAt"`
}

// TableName returns the table name for Fill
func (Fill) TableName() string {
	return "fills"
}
