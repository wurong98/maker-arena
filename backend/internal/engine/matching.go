package engine

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/maker-arena/backend/internal/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Order represents an order in the matching engine
type Order struct {
	ID          string
	StrategyID  string
	Symbol      string
	Side        string // buy/sell
	Type        string // limit/market
	Price       decimal.Decimal
	Quantity    decimal.Decimal
	FilledQty   decimal.Decimal
	Status      string // open/filled/canceled/liquidated
	TimeInForce string // GTC/IOC/FOK
	TTL         int    // seconds, 0 = never expire
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Ticker represents market ticker data
type Ticker struct {
	Symbol                string
	Price                 decimal.Decimal
	PriceChange24h        decimal.Decimal
	PriceChangePercent24h decimal.Decimal
	PreviousPrice         decimal.Decimal
	UpdatedAt             time.Time
}

// OrderBook represents an order book for a symbol
type OrderBook struct {
	Symbol string
	Bids   []*Order // buy orders, sorted by price desc
	Asks   []*Order // sell orders, sorted by price asc
	Locker sync.RWMutex
}

// fillInfo holds order and fill quantity
type fillInfo struct {
	order   *Order
	fillQty decimal.Decimal
}

// MatchingEngine handles order matching
type MatchingEngine struct {
	db              *gorm.DB
	makerFeeRate    decimal.Decimal
	positionManager *PositionManager
	orderBooks      map[string]*OrderBook
	tickers         map[string]*Ticker
	mu              sync.RWMutex
	orderChan       chan string // channel for order processing
	stopChan        chan struct{}
	wg              sync.WaitGroup
}

// NewMatchingEngine creates a new matching engine
func NewMatchingEngine(db *gorm.DB, makerFeeRate decimal.Decimal, pm *PositionManager) *MatchingEngine {
	return &MatchingEngine{
		db:              db,
		makerFeeRate:    makerFeeRate,
		positionManager: pm,
		orderBooks:      make(map[string]*OrderBook),
		tickers:         make(map[string]*Ticker),
		orderChan:       make(chan string, 1000),
		stopChan:        make(chan struct{}),
	}
}

// Start starts the matching engine
func (me *MatchingEngine) Start() {
	// Load open orders from database
	var orders []models.Order
	if err := me.db.Where("status = ?", "open").Find(&orders).Error; err != nil {
		fmt.Printf("Failed to load open orders: %v\n", err)
		return
	}

	for _, o := range orders {
		me.AddOrder(&Order{
			ID:          o.ID,
			StrategyID:  o.StrategyID,
			Symbol:      o.Symbol,
			Side:        o.Side,
			Type:        o.Type,
			Price:       o.Price,
			Quantity:    o.Quantity,
			FilledQty:   o.FilledQuantity,
			Status:      o.Status,
			TimeInForce: o.TimeInForce,
			TTL:         o.TTL,
			CreatedAt:   o.CreatedAt,
			UpdatedAt:   o.UpdatedAt,
		})
	}

	// Start order processing goroutine
	me.wg.Add(1)
	go me.processOrders()

	fmt.Printf("Matching engine started with %d open orders\n", len(orders))
}

// Stop stops the matching engine
func (me *MatchingEngine) Stop() {
	close(me.stopChan)
	me.wg.Wait()
	close(me.orderChan)
}

// AddOrder adds an order to the order book
func (me *MatchingEngine) AddOrder(order *Order) {
	me.mu.Lock()
	defer me.mu.Unlock()

	ob, ok := me.orderBooks[order.Symbol]
	if !ok {
		ob = &OrderBook{
			Symbol: order.Symbol,
			Bids:   make([]*Order, 0),
			Asks:   make([]*Order, 0),
		}
		me.orderBooks[order.Symbol] = ob
	}

	ob.Locker.Lock()
	defer ob.Locker.Unlock()

	if order.Side == "buy" {
		ob.Bids = append(ob.Bids, order)
		sort.Slice(ob.Bids, func(i, j int) bool {
			return ob.Bids[i].Price.GreaterThan(ob.Bids[j].Price)
		})
	} else {
		ob.Asks = append(ob.Asks, order)
		sort.Slice(ob.Asks, func(i, j int) bool {
			return ob.Asks[i].Price.LessThan(ob.Asks[j].Price)
		})
	}
}

// CancelOrder removes an order from the order book
func (me *MatchingEngine) CancelOrder(orderID string) {
	me.mu.Lock()
	defer me.mu.Unlock()

	for _, ob := range me.orderBooks {
		ob.Locker.Lock()

		// Find and remove from bids
		for i, o := range ob.Bids {
			if o.ID == orderID {
				ob.Bids = append(ob.Bids[:i], ob.Bids[i+1:]...)
				ob.Locker.Unlock()
				return
			}
		}

		// Find and remove from asks
		for i, o := range ob.Asks {
			if o.ID == orderID {
				ob.Asks = append(ob.Asks[:i], ob.Asks[i+1:]...)
				ob.Locker.Unlock()
				return
			}
		}

		ob.Locker.Unlock()
	}
}

// ProcessOrder processes a single order
func (me *MatchingEngine) ProcessOrder(orderID string) {
	select {
	case me.orderChan <- orderID:
	default:
	}
}

// processOrders continuously processes orders from the channel
func (me *MatchingEngine) processOrders() {
	defer me.wg.Done()

	for {
		select {
		case orderID := <-me.orderChan:
			me.matchOrder(orderID)
		case <-me.stopChan:
			return
		}
	}
}

// HandleTrade processes a trade from Binance WebSocket
func (me *MatchingEngine) HandleTrade(symbol string, price decimal.Decimal, quantity decimal.Decimal, tradeTime int64) {
	me.mu.Lock()
	defer me.mu.Unlock()

	// Update ticker
	ticker, ok := me.tickers[symbol]
	if !ok {
		ticker = &Ticker{
			Symbol:    symbol,
			Price:     price,
			UpdatedAt: time.Now(),
		}
		me.tickers[symbol] = ticker
	}

	// Update price change
	if ticker.PreviousPrice.IsZero() {
		ticker.PriceChange24h = decimal.Zero
		ticker.PriceChangePercent24h = decimal.Zero
	} else {
		ticker.PriceChange24h = price.Sub(ticker.PreviousPrice)
		if ticker.PreviousPrice.GreaterThan(decimal.Zero) {
			ticker.PriceChangePercent24h = ticker.PriceChange24h.Div(ticker.PreviousPrice).Mul(decimal.NewFromInt(100))
		}
	}
	ticker.PreviousPrice = ticker.Price
	ticker.Price = price
	ticker.UpdatedAt = time.UnixMilli(tradeTime)

	// Get order book
	ob, ok := me.orderBooks[symbol]
	if !ok {
		return
	}

	ob.Locker.Lock()
	defer ob.Locker.Unlock()

	// Check for matches according to design doc section 3.2:
	// - Buy order: price crosses from <= order.price to > order.price
	// - Sell order: price crosses from >= order.price to < order.price
	var filledOrders []fillInfo

	// Check buy orders (long) - price goes up past order price
	for _, order := range ob.Bids {
		if order.Status != "open" {
			continue
		}

		// Check if price crossed order price (from below or equal to above)
		if ticker.PreviousPrice.LessThanOrEqual(order.Price) && price.GreaterThan(order.Price) {
			// Calculate fill quantity
			remaining := order.Quantity.Sub(order.FilledQty)
			fillQty := quantity
			if fillQty.GreaterThan(remaining) {
				fillQty = remaining
			}

			if fillQty.GreaterThan(decimal.Zero) {
				order.FilledQty = order.FilledQty.Add(fillQty)
				filledOrders = append(filledOrders, fillInfo{order: order, fillQty: fillQty})
			}
		}
	}

	// Check sell orders (short) - price goes down past order price
	for _, order := range ob.Asks {
		if order.Status != "open" {
			continue
		}

		// Check if price crossed order price (from above or equal to below)
		if ticker.PreviousPrice.GreaterThanOrEqual(order.Price) && price.LessThan(order.Price) {
			// Calculate fill quantity
			remaining := order.Quantity.Sub(order.FilledQty)
			fillQty := quantity
			if fillQty.GreaterThan(remaining) {
				fillQty = remaining
			}

			if fillQty.GreaterThan(decimal.Zero) {
				order.FilledQty = order.FilledQty.Add(fillQty)
				filledOrders = append(filledOrders, fillInfo{order: order, fillQty: fillQty})
			}
		}
	}

	// Process filled orders
	for _, fi := range filledOrders {
		me.executeFill(fi.order, fi.fillQty, price)
	}

	// Check TTL for orders
	now := time.Now()
	for _, order := range ob.Bids {
		if order.TTL > 0 && order.Status == "open" {
			expiry := order.CreatedAt.Add(time.Duration(order.TTL) * time.Second)
			if now.After(expiry) {
				order.Status = "canceled"
				me.db.Model(&models.Order{}).Where("id = ?", order.ID).Updates(map[string]interface{}{
					"status":    "canceled",
					"updated_at": now,
				})
				// Unfreeze margin for TTL expired orders
				marginRequired := order.Price.Mul(order.Quantity).Div(decimal.NewFromInt(100))
				me.db.Model(&models.Strategy{}).Where("id = ?", order.StrategyID).UpdateColumn("frozen_margin", gorm.Expr("frozen_margin - ?", marginRequired))
			}
		}
	}
	for _, order := range ob.Asks {
		if order.TTL > 0 && order.Status == "open" {
			expiry := order.CreatedAt.Add(time.Duration(order.TTL) * time.Second)
			if now.After(expiry) {
				order.Status = "canceled"
				me.db.Model(&models.Order{}).Where("id = ?", order.ID).Updates(map[string]interface{}{
					"status":    "canceled",
					"updated_at": now,
				})
				// Unfreeze margin for TTL expired orders
				marginRequired := order.Price.Mul(order.Quantity).Div(decimal.NewFromInt(100))
				me.db.Model(&models.Strategy{}).Where("id = ?", order.StrategyID).UpdateColumn("frozen_margin", gorm.Expr("frozen_margin - ?", marginRequired))
			}
		}
	}
}

// matchOrder tries to match an order with the order book
func (me *MatchingEngine) matchOrder(orderID string) {
	me.mu.RLock()
	var order *Order
	var ob *OrderBook
	var isBid bool
	for _, book := range me.orderBooks {
		book.Locker.RLock()
		for _, o := range book.Bids {
			if o.ID == orderID {
				order = o
				ob = book
				isBid = true
			}
		}
		for _, o := range book.Asks {
			if o.ID == orderID {
				order = o
				ob = book
				isBid = false
			}
		}
		book.Locker.RUnlock()
	}
	me.mu.RUnlock()

	if order == nil || ob == nil {
		return
	}

	me.mu.RLock()
	var ticker *Ticker
	if t, ok := me.tickers[order.Symbol]; ok {
		ticker = t
	}
	me.mu.RUnlock()

	if ticker == nil {
		return
	}

	// For market orders or IOC/FOK, try to fill immediately
	if order.Type == "market" || order.TimeInForce == "IOC" || order.TimeInForce == "FOK" {
		remaining := order.Quantity.Sub(order.FilledQty)
		if remaining.GreaterThan(decimal.Zero) {
			// Get opposite side orders
			var oppositeOrders []*Order
			if isBid {
				ob.Locker.RLock()
				oppositeOrders = ob.Asks
				ob.Locker.RUnlock()
			} else {
				ob.Locker.RLock()
				oppositeOrders = ob.Bids
				ob.Locker.RUnlock()
			}

			fillQty := decimal.Zero
			avgPrice := decimal.Zero
			for _, o := range oppositeOrders {
				if o.Status != "open" {
					continue
				}

				// Check price condition
				var canFill bool
				if isBid {
					// Buy order: can fill if price <= order price
					canFill = o.Price.LessThanOrEqual(order.Price)
				} else {
					// Sell order: can fill if price >= order price
					canFill = o.Price.GreaterThanOrEqual(order.Price)
				}

				if !canFill {
					continue
				}

				oRemaining := o.Quantity.Sub(o.FilledQty)
				available := oRemaining
				needed := remaining.Sub(fillQty)

				var qty decimal.Decimal
				if available.GreaterThanOrEqual(needed) {
					qty = needed
				} else {
					qty = available
				}

				fillQty = fillQty.Add(qty)
				avgPrice = avgPrice.Add(o.Price.Mul(qty))

				o.FilledQty = o.FilledQty.Add(qty)
				me.executeFill(o, qty, o.Price)

				if fillQty.GreaterThanOrEqual(needed) {
					break
				}
			}

			if fillQty.GreaterThan(decimal.Zero) && avgPrice.GreaterThan(decimal.Zero) {
				avgPrice = avgPrice.Div(fillQty)
				order.FilledQty = order.FilledQty.Add(fillQty)
				me.executeFill(order, fillQty, avgPrice)
			}
		}
	}
}

// executeFill executes a fill for an order
func (me *MatchingEngine) executeFill(order *Order, fillQty decimal.Decimal, price decimal.Decimal) {
	// Update order status
	if order.FilledQty.GreaterThanOrEqual(order.Quantity) {
		order.Status = "filled"
	} else if order.TimeInForce == "FOK" && order.FilledQty.LessThan(order.Quantity) {
		order.Status = "canceled"
	}

	order.UpdatedAt = time.Now()

	// Calculate fee
	fee := price.Mul(fillQty).Mul(me.makerFeeRate)

	// Create fill record
	fill := models.Fill{
		ID:         uuid.New().String(),
		OrderID:    order.ID,
		StrategyID: order.StrategyID,
		Symbol:     order.Symbol,
		Side:       order.Side,
		Price:      price,
		Quantity:   fillQty,
		Fee:        fee,
	}

	// Update position
	if me.positionManager != nil {
		me.positionManager.UpdatePosition(order.StrategyID, order.Symbol, order.Side, fillQty, price, fee)
	}

	// Check liquidation
	if me.positionManager != nil {
		me.positionManager.CheckLiquidation(order.StrategyID, order.Symbol)
	}

	// Save to database in transaction
	tx := me.db.Begin()

	// Update order
	tx.Model(&models.Order{}).Where("id = ?", order.ID).Updates(map[string]interface{}{
		"filled_quantity": order.FilledQty,
		"status":          order.Status,
		"updated_at":      order.UpdatedAt,
	})

	// Create fill record
	tx.Create(&fill)

	// Deduct fee from balance
	tx.Model(&models.Strategy{}).Where("id = ?", order.StrategyID).UpdateColumn("balance", gorm.Expr("balance - ?", fee))

	// Unfreeze margin (full amount since system only supports all-or-nothing fills)
	marginRequired := order.Price.Mul(order.Quantity).Div(decimal.NewFromInt(100))
	tx.Model(&models.Strategy{}).Where("id = ?", order.StrategyID).UpdateColumn("frozen_margin", gorm.Expr("frozen_margin - ?", marginRequired))

	tx.Commit()
}

// GetTicker returns a copy of the ticker for a symbol (to avoid deadlock)
func (me *MatchingEngine) GetTicker(symbol string) *Ticker {
	me.mu.RLock()
	defer me.mu.RUnlock()

	if ticker, ok := me.tickers[symbol]; ok {
		// Return a copy to avoid deadlock
		t := *ticker
		return &t
	}
	return nil
}

// GetAllTickers returns all tickers (copies to avoid deadlock)
func (me *MatchingEngine) GetAllTickers() map[string]*Ticker {
	me.mu.RLock()
	defer me.mu.RUnlock()

	result := make(map[string]*Ticker)
	for k, v := range me.tickers {
		t := *v
		result[k] = &t
	}
	return result
}

// GetOrder returns an order by ID
func (me *MatchingEngine) GetOrder(orderID string) *Order {
	me.mu.RLock()
	defer me.mu.RUnlock()

	for _, ob := range me.orderBooks {
		ob.Locker.RLock()
		for _, o := range ob.Bids {
			if o.ID == orderID {
				ob.Locker.RUnlock()
				return o
			}
		}
		for _, o := range ob.Asks {
			if o.ID == orderID {
				ob.Locker.RUnlock()
				return o
			}
		}
		ob.Locker.RUnlock()
	}
	return nil
}
