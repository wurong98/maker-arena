package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Strategy represents a trading strategy
type Strategy struct {
	ID          string          `gorm:"primaryKey;type:uuid" json:"id"`
	APIKey      string          `gorm:"uniqueIndex;type:varchar(64)" json:"apiKey"`
	Name        string          `gorm:"type:varchar(128)" json:"name"`
	Description string          `gorm:"type:text" json:"description"`
	Enabled     bool            `gorm:"default:true" json:"enabled"`
	Balance     decimal.Decimal `gorm:"type:numeric(20,8);default:5000" json:"balance"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

// TableName returns the table name for Strategy
func (Strategy) TableName() string {
	return "strategies"
}
