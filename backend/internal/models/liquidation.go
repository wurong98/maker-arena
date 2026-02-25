package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Liquidation represents a liquidation event
type Liquidation struct {
	ID               string          `gorm:"primaryKey;type:uuid;not null" json:"id"`
	StrategyID       string          `gorm:"type:uuid;index;not null" json:"strategyId"`
	StrategyName     string          `gorm:"type:varchar(128);not null" json:"strategyName"`
	Symbol           string          `gorm:"type:varchar(32);not null" json:"symbol"`
	Side             string          `gorm:"type:varchar(8);not null" json:"side"`
	LiquidationPrice decimal.Decimal `gorm:"type:numeric(20,8)" json:"liquidationPrice"`
	Quantity         decimal.Decimal `gorm:"type:numeric(20,8)" json:"quantity"`
	CreatedAt        time.Time       `json:"createdAt"`
}

// TableName returns the table name for Liquidation
func (Liquidation) TableName() string {
	return "liquidations"
}
