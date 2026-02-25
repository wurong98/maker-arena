package engine

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/maker-arena/backend/internal/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Position represents a position in the position manager
type Position struct {
	StrategyID      string
	Symbol          string
	Side            string // long/short
	Quantity        decimal.Decimal
	EntryPrice      decimal.Decimal
	CurrentPrice    decimal.Decimal
	Leverage        int
	LiquidationPrice decimal.Decimal
	UnrealizedPnl   decimal.Decimal
	PositionValue   decimal.Decimal
}

// PositionManager manages positions
type PositionManager struct {
	db              *gorm.DB
	leverage        int
	positions       map[string]*Position // key: strategyID_symbol
	mu              sync.RWMutex
	tickerGetter    func(symbol string) *Ticker
}

// NewPositionManager creates a new position manager
func NewPositionManager(db *gorm.DB, leverage int, tickerGetter func(symbol string) *Ticker) *PositionManager {
	return &PositionManager{
		db:        db,
		leverage:  leverage,
		positions: make(map[string]*Position),
	}
}

// Start loads positions from database
func (pm *PositionManager) Start() {
	// Load positions from database
	var positions []models.Position
	if err := pm.db.Find(&positions).Error; err != nil {
		fmt.Printf("Failed to load positions: %v\n", err)
		return
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	for _, p := range positions {
		// Get current price
		currentPrice := decimal.Zero
		if tg := pm.tickerGetter; tg != nil {
			ticker := tg(p.Symbol)
			if ticker != nil {
				currentPrice = ticker.Price
			}
		}

		position := &Position{
			StrategyID: p.StrategyID,
			Symbol:     p.Symbol,
			Side:       p.Side,
			Quantity:   p.Quantity,
			EntryPrice: p.EntryPrice,
			CurrentPrice: currentPrice,
			Leverage:   p.Leverage,
		}

		// Calculate liquidation price
		position.LiquidationPrice = pm.calculateLiquidationPrice(position)

		// Calculate position value and unrealized PnL
		if currentPrice.GreaterThan(decimal.Zero) {
			position.PositionValue = currentPrice.Mul(p.Quantity)
			position.UnrealizedPnl = pm.calculateUnrealizedPnl(position)
		}

		key := pm.positionKey(p.StrategyID, p.Symbol)
		pm.positions[key] = position
	}

	fmt.Printf("Position manager started with %d positions\n", len(positions))
}

// Stop saves positions to database
func (pm *PositionManager) Stop() {
	// Save all positions to database
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for _, p := range pm.positions {
		if p.Quantity.GreaterThan(decimal.Zero) {
			var existing models.Position
			err := pm.db.First(&existing, "strategy_id = ? AND symbol = ?", p.StrategyID, p.Symbol).Error
			if err == gorm.ErrRecordNotFound {
				// Create new position
				pos := models.Position{
					ID:          uuid.New().String(),
					StrategyID:  p.StrategyID,
					Symbol:      p.Symbol,
					Side:        p.Side,
					Quantity:    p.Quantity,
					EntryPrice:  p.EntryPrice,
					Leverage:    p.Leverage,
				}
				pm.db.Create(&pos)
			} else if err == nil {
				// Update existing position
				existing.Quantity = p.Quantity
				existing.EntryPrice = p.EntryPrice
				pm.db.Save(&existing)
			}
		}
	}
}

// positionKey generates a key for position map
func (pm *PositionManager) positionKey(strategyID, symbol string) string {
	return strategyID + "_" + symbol
}

// UpdatePosition updates a position after a fill
func (pm *PositionManager) UpdatePosition(strategyID, symbol, side string, quantity decimal.Decimal, price decimal.Decimal, fee decimal.Decimal) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	key := pm.positionKey(strategyID, symbol)
	position, exists := pm.positions[key]

	if !exists || position.Quantity.IsZero() {
		// No existing position, create new one
		sideName := "long"
		if side == "sell" {
			sideName = "short"
		}

		position = &Position{
			StrategyID:   strategyID,
			Symbol:       symbol,
			Side:         sideName,
			Quantity:     quantity,
			EntryPrice:   price,
			Leverage:     pm.leverage,
		}
		pm.positions[key] = position
	} else {
		// Existing position - check if same direction or opposite
		expectedSide := "long"
		if side == "sell" {
			expectedSide = "short"
		}

		if position.Side == expectedSide {
			// Same direction - add to position
			totalValue := position.EntryPrice.Mul(position.Quantity).Add(price.Mul(quantity))
			position.Quantity = position.Quantity.Add(quantity)
			position.EntryPrice = totalValue.Div(position.Quantity)
		} else {
			// Opposite direction - reduce or flip position
			if quantity.GreaterThanOrEqual(position.Quantity) {
				// Reduce or flip to opposite
				remaining := quantity.Sub(position.Quantity)
				if remaining.IsZero() {
					position.Quantity = decimal.Zero
					position.EntryPrice = decimal.Zero
				} else {
					position.Side = expectedSide
					position.Quantity = remaining
					position.EntryPrice = price
				}
			} else {
				// Reduce position
				position.Quantity = position.Quantity.Sub(quantity)
			}
		}
	}

	// Update current price from ticker
	if tg := pm.tickerGetter; tg != nil {
		ticker := tg(symbol)
		if ticker != nil {
			position.CurrentPrice = ticker.Price
		}
	}

	// Calculate liquidation price
	position.LiquidationPrice = pm.calculateLiquidationPrice(position)

	// Calculate position value and unrealized PnL
	if position.CurrentPrice.GreaterThan(decimal.Zero) {
		position.PositionValue = position.CurrentPrice.Mul(position.Quantity)
		position.UnrealizedPnl = pm.calculateUnrealizedPnl(position)
	}

	// Save to database
	pm.savePosition(position)
}

// calculateLiquidationPrice calculates the liquidation price
// Long: liquidation_price = entry_price * (1 - 1/leverage)
// Short: liquidation_price = entry_price * (1 + 1/leverage)
func (pm *PositionManager) calculateLiquidationPrice(position *Position) decimal.Decimal {
	if position.EntryPrice.IsZero() || position.Quantity.IsZero() {
		return decimal.Zero
	}

	leverage := decimal.NewFromInt(int64(position.Leverage))
	one := decimal.NewFromInt(1)

	if position.Side == "long" {
		// Long: liquidation when price drops
		return position.EntryPrice.Mul(one.Sub(one.Div(leverage)))
	} else {
		// Short: liquidation when price rises
		return position.EntryPrice.Mul(one.Add(one.Div(leverage)))
	}
}

// calculateUnrealizedPnl calculates unrealized profit/loss
func (pm *PositionManager) calculateUnrealizedPnl(position *Position) decimal.Decimal {
	if position.EntryPrice.IsZero() || position.Quantity.IsZero() || position.CurrentPrice.IsZero() {
		return decimal.Zero
	}

	if position.Side == "long" {
		// Long: profit when price goes up
		return position.CurrentPrice.Sub(position.EntryPrice).Mul(position.Quantity)
	} else {
		// Short: profit when price goes down
		return position.EntryPrice.Sub(position.CurrentPrice).Mul(position.Quantity)
	}
}

// CheckLiquidation checks if a position should be liquidated
func (pm *PositionManager) CheckLiquidation(strategyID, symbol string) {
	pm.mu.RLock()
	key := pm.positionKey(strategyID, symbol)
	position, exists := pm.positions[key]
	pm.mu.RUnlock()

	if !exists || position.Quantity.IsZero() {
		return
	}

	// Get current price
	currentPrice := position.CurrentPrice
	if currentPrice.IsZero() {
		return
	}

	// Check if price crossed liquidation price
	var shouldLiquidate bool
	if position.Side == "long" {
		shouldLiquidate = currentPrice.LessThanOrEqual(position.LiquidationPrice)
	} else {
		shouldLiquidate = currentPrice.GreaterThanOrEqual(position.LiquidationPrice)
	}

	if shouldLiquidate {
		pm.Liquidate(strategyID, symbol)
	}
}

// Liquidate liquidates a position
func (pm *PositionManager) Liquidate(strategyID, symbol string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	key := pm.positionKey(strategyID, symbol)
	position, exists := pm.positions[key]
	if !exists || position.Quantity.IsZero() {
		return
	}

	// Get strategy info
	var strategy models.Strategy
	if err := pm.db.First(&strategy, "id = ?", strategyID).Error; err != nil {
		fmt.Printf("Failed to find strategy for liquidation: %v\n", err)
		return
	}

	// Create liquidation record
	liquidation := models.Liquidation{
		ID:               uuid.New().String(),
		StrategyID:       strategyID,
		StrategyName:     strategy.Name,
		Symbol:           symbol,
		Side:             position.Side,
		LiquidationPrice: position.LiquidationPrice,
		Quantity:         position.Quantity,
	}

	pm.db.Create(&liquidation)

	// Calculate realized PnL at liquidation
	var realizedPnl decimal.Decimal
	if position.Side == "long" {
		realizedPnl = position.CurrentPrice.Sub(position.EntryPrice).Mul(position.Quantity)
	} else {
		realizedPnl = position.EntryPrice.Sub(position.CurrentPrice).Mul(position.Quantity)
	}

	// Update balance (will be negative after liquidation)
	newBalance := strategy.Balance.Add(realizedPnl)

	// Disable strategy
	pm.db.Model(&strategy).Updates(map[string]interface{}{
		"enabled": false,
		"balance": newBalance,
	})

	// Delete position
	delete(pm.positions, key)
	pm.db.Where("strategy_id = ? AND symbol = ?", strategyID, symbol).Delete(&models.Position{})

	// Update all orders for this strategy to liquidated
	pm.db.Model(&models.Order{}).Where("strategy_id = ? AND status = ?", strategyID, "open").Update("status", "liquidated")

	fmt.Printf("Liquidated position: strategy=%s, symbol=%s, pnl=%s, new_balance=%s\n",
		strategyID, symbol, realizedPnl.String(), newBalance.String())
}

// GetPositions returns positions for a strategy
func (pm *PositionManager) GetPositions(strategyID, symbol string) []Position {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var result []Position
	for _, p := range pm.positions {
		if p.StrategyID != strategyID {
			continue
		}
		if symbol != "" && p.Symbol != symbol {
			continue
		}

		// Update current price from ticker
		if tg := pm.tickerGetter; tg != nil {
			ticker := tg(p.Symbol)
			if ticker != nil {
				p.CurrentPrice = ticker.Price
				p.UnrealizedPnl = pm.calculateUnrealizedPnl(p)
				p.PositionValue = p.CurrentPrice.Mul(p.Quantity)
			}
		}

		result = append(result, *p)
	}

	return result
}

// CalculateUnrealizedPnl calculates total unrealized PnL for a strategy
func (pm *PositionManager) CalculateUnrealizedPnl(strategyID string) decimal.Decimal {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	total := decimal.Zero
	for _, p := range pm.positions {
		if p.StrategyID != strategyID {
			continue
		}

		// Update current price from ticker
		if tg := pm.tickerGetter; tg != nil {
			ticker := tg(p.Symbol)
			if ticker != nil {
				p.CurrentPrice = ticker.Price
				p.UnrealizedPnl = pm.calculateUnrealizedPnl(p)
			}
		}

		total = total.Add(p.UnrealizedPnl)
	}

	return total
}

// CalculateUsedMargin calculates used margin for a strategy
func (pm *PositionManager) CalculateUsedMargin(strategyID string) decimal.Decimal {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	total := decimal.Zero
	for _, p := range pm.positions {
		if p.StrategyID != strategyID {
			continue
		}

		if p.PositionValue.IsZero() && p.EntryPrice.GreaterThan(decimal.Zero) && p.Quantity.GreaterThan(decimal.Zero) {
			p.PositionValue = p.EntryPrice.Mul(p.Quantity)
		}

		// Margin = position_value / leverage
		margin := p.PositionValue.Div(decimal.NewFromInt(int64(p.Leverage)))
		total = total.Add(margin)
	}

	return total
}

// savePosition saves position to database
func (pm *PositionManager) savePosition(position *Position) {
	if position.Quantity.IsZero() {
		// Delete position if quantity is zero
		pm.db.Where("strategy_id = ? AND symbol = ?", position.StrategyID, position.Symbol).Delete(&models.Position{})
		return
	}

	var existing models.Position
	err := pm.db.First(&existing, "strategy_id = ? AND symbol = ?", position.StrategyID, position.Symbol).Error
	if err == gorm.ErrRecordNotFound {
		// Create new position
		pos := models.Position{
			ID:          uuid.New().String(),
			StrategyID:  position.StrategyID,
			Symbol:      position.Symbol,
			Side:        position.Side,
			Quantity:    position.Quantity,
			EntryPrice:  position.EntryPrice,
			Leverage:    position.Leverage,
		}
		pm.db.Create(&pos)
	} else if err == nil {
		// Update existing position
		existing.Quantity = position.Quantity
		existing.EntryPrice = position.EntryPrice
		existing.Side = position.Side
		existing.Leverage = position.Leverage
		pm.db.Save(&existing)
	}
}

// SetTickerGetter sets the function to get ticker data
func (pm *PositionManager) SetTickerGetter(tickerGetter func(symbol string) *Ticker) {
	pm.tickerGetter = tickerGetter
}
