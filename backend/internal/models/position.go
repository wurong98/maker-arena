package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Position represents a trading position
type Position struct {
	ID         string          `gorm:"primaryKey;type:uuid;not null" json:"id"`
	StrategyID string          `gorm:"type:uuid;uniqueIndex:idx_strategy_symbol;not null" json:"strategyId"`
	Symbol     string          `gorm:"type:varchar(32);uniqueIndex:idx_strategy_symbol;not null" json:"symbol"`
	Side       string          `gorm:"type:varchar(8);not null" json:"side"` // long/short
	Quantity   decimal.Decimal `gorm:"type:numeric(20,8)" json:"quantity"`
	EntryPrice decimal.Decimal `gorm:"type:numeric(20,8)" json:"entryPrice"`
	Leverage   int             `gorm:"default:100" json:"leverage"`
	CreatedAt  time.Time       `json:"createdAt"`
	UpdatedAt  time.Time       `json:"updatedAt"`
}

// TableName returns the table name for Position
func (Position) TableName() string {
	return "positions"
}
