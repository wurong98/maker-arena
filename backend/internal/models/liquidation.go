package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Liquidation represents a liquidation event
type Liquidation struct {
	ID               string          `gorm:"primaryKey;type:uuid" json:"id"`
	StrategyID       string          `gorm:"type:uuid;index" json:"strategyId"`
	StrategyName     string          `gorm:"type:varchar(128)" json:"strategyName"`
	Symbol           string          `gorm:"type:varchar(32)" json:"symbol"`
	Side             string          `gorm:"type:varchar(8)" json:"side"`
	LiquidationPrice decimal.Decimal `gorm:"type:numeric(20,8)" json:"liquidationPrice"`
	Quantity         decimal.Decimal `gorm:"type:numeric(20,8)" json:"quantity"`
	CreatedAt        time.Time       `json:"createdAt"`
}

// TableName returns the table name for Liquidation
func (Liquidation) TableName() string {
	return "liquidations"
}
