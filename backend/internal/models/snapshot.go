package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// AccountSnapshot represents an account balance snapshot
type AccountSnapshot struct {
	ID            string          `gorm:"primaryKey;type:uuid" json:"id"`
	StrategyID    string          `gorm:"type:uuid;index" json:"strategyId"`
	Balance       decimal.Decimal `gorm:"type:numeric(20,8)" json:"balance"`
	UnrealizedPnl decimal.Decimal `gorm:"type:numeric(20,8)" json:"unrealizedPnl"`
	TotalEquity   decimal.Decimal `gorm:"type:numeric(20,8)" json:"totalEquity"`
	CreatedAt     time.Time       `json:"createdAt"`
}

// TableName returns the table name for AccountSnapshot
func (AccountSnapshot) TableName() string {
	return "account_snapshots"
}

// PositionSnapshot represents a position PnL snapshot
type PositionSnapshot struct {
	ID            string          `gorm:"primaryKey;type:uuid" json:"id"`
	StrategyID    string          `gorm:"type:uuid;index" json:"strategyId"`
	Symbol        string          `gorm:"type:varchar(32)" json:"symbol"`
	UnrealizedPnl decimal.Decimal `gorm:"type:numeric(20,8)" json:"unrealizedPnl"`
	PositionValue decimal.Decimal `gorm:"type:numeric(20,8)" json:"positionValue"`
	AvgPrice      decimal.Decimal `gorm:"type:numeric(20,8)" json:"avgPrice"`
	CreatedAt     time.Time       `json:"createdAt"`
}

// TableName returns the table name for PositionSnapshot
func (PositionSnapshot) TableName() string {
	return "position_snapshots"
}
