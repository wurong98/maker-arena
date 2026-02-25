package database

import (
	"log"

	"github.com/maker-arena/backend/internal/models"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	log.Println("Running database migrations...")

	err := db.AutoMigrate(
		&models.Strategy{},
		&models.Order{},
		&models.Fill{},
		&models.Position{},
		&models.Liquidation{},
		&models.AccountSnapshot{},
		&models.PositionSnapshot{},
		&models.Ticker{},
	)

	if err != nil {
		return err
	}

	log.Println("Database migrations completed")
	return nil
}
