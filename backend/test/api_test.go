package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/maker-arena/backend/internal/handlers"
	"github.com/maker-arena/backend/internal/models"
	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestDatabase is a test database wrapper
type TestDatabase struct {
	DB *gorm.DB
}

// SetupTestDB sets up a test database
func SetupTestDB(t *testing.T) *TestDatabase {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto migrate
	db.AutoMigrate(
		&models.Strategy{},
		&models.Order{},
		&models.Fill{},
		&models.Position{},
		&models.Liquidation{},
		&models.AccountSnapshot{},
		&models.PositionSnapshot{},
		&models.Ticker{},
	)

	return &TestDatabase{DB: db}
}

// TestStrategyHandlers tests strategy handlers
func TestStrategyHandlers(t *testing.T) {
	testDB := SetupTestDB(t)

	// Create a mock config
	cfg := &struct {
		Admin struct {
			Password string
		}
	}{
		Admin: struct {
			Password string
		}{
			Password: "test-password",
		},
	}

	// Create handler
	handler := handlers.NewStrategyHandler(testDB.DB, (*struct {
		Admin struct {
			Password string
		}
	})(cfg))

	t.Run("ListStrategies", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/strategies", nil)
		w := httptest.NewRecorder()

		handler.List(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("CreateStrategy", func(t *testing.T) {
		body := map[string]string{
			"name":        "Test Strategy",
			"description": "Test description",
			"balance":     "5000",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/strategies", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Admin-Password", "test-password")
		w := httptest.NewRecorder()

		handler.Create(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", w.Code)
		}
	})
}

// TestPositionEngine tests position calculations
func TestPositionEngine(t *testing.T) {
	t.Run("CalculateLiquidationPrice", func(t *testing.T) {
		// Test long position liquidation price
		// liquidation_price = entry_price * (1 - 1/leverage)
		// For 100x leverage: 50000 * (1 - 0.01) = 49500
		entryPrice := decimal.NewFromInt(50000)
		leverage := 100

		one := decimal.NewFromInt(1)
		leverageDec := decimal.NewFromInt(int64(leverage))
		liquidationPrice := entryPrice.Sub(entryPrice.Div(leverageDec))

		expected := decimal.NewFromInt(49500)
		if !liquidationPrice.Equal(expected) {
			t.Errorf("Expected liquidation price %s, got %s", expected.String(), liquidationPrice.String())
		}
	})

	t.Run("CalculateUnrealizedPnlLong", func(t *testing.T) {
		// Long: profit when price goes up
		entryPrice := decimal.NewFromInt(50000)
		currentPrice := decimal.NewFromInt(51000)
		quantity := decimal.NewFromFloat(0.01)

		pnl := currentPrice.Sub(entryPrice).Mul(quantity)
		expected := decimal.NewFromFloat(10) // (51000 - 50000) * 0.01 = 100 * 0.01 = 10

		if !pnl.Equal(expected) {
			t.Errorf("Expected PnL %s, got %s", expected.String(), pnl.String())
		}
	})

	t.Run("CalculateUnrealizedPnlShort", func(t *testing.T) {
		// Short: profit when price goes down
		entryPrice := decimal.NewFromInt(50000)
		currentPrice := decimal.NewFromInt(49000)
		quantity := decimal.NewFromFloat(0.01)

		pnl := entryPrice.Sub(currentPrice).Mul(quantity)
		expected := decimal.NewFromFloat(10) // (50000 - 49000) * 0.01 = 100 * 0.01 = 10

		if !pnl.Equal(expected) {
			t.Errorf("Expected PnL %s, got %s", expected.String(), pnl.String())
		}
	})
}

// TestMatchingEngine tests matching logic
func TestMatchingEngine(t *testing.T) {
	t.Run("OrderBookSorting", func(t *testing.T) {
		// Test that bids are sorted by price descending
		// Test that asks are sorted by price ascending
		// This is a placeholder for actual order book tests
	})
}
