package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/maker-arena/backend/internal/config"
	"github.com/maker-arena/backend/internal/engine"
	"github.com/maker-arena/backend/internal/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ExchangeHandler handles exchange-related requests
type ExchangeHandler struct {
	db               *gorm.DB
	cfg              *config.Config
	matchingEngine   *engine.MatchingEngine
	positionManager  *engine.PositionManager
}

// NewExchangeHandler creates a new ExchangeHandler
func NewExchangeHandler(db *gorm.DB, cfg *config.Config, me *engine.MatchingEngine, pm *engine.PositionManager) *ExchangeHandler {
	return &ExchangeHandler{
		db:              db,
		cfg:             cfg,
		matchingEngine:  me,
		positionManager: pm,
	}
}

// CreateOrderRequest represents the request for creating an order
type CreateOrderRequest struct {
	Symbol     string `json:"symbol"`
	Side       string `json:"side"`
	Type       string `json:"type"`
	Quantity   string `json:"quantity"`
	Price      string `json:"price"`
	TimeInForce string `json:"timeInForce"`
	TTL        int    `json:"ttl"`
}

// CreateOrderResponse represents the response for creating an order
type CreateOrderResponse struct {
	ID             string `json:"id"`
	Symbol         string `json:"symbol"`
	Side           string `json:"side"`
	Type           string `json:"type"`
	Price          string `json:"price"`
	Quantity       string `json:"quantity"`
	FilledQuantity string `json:"filledQuantity"`
	Status         string `json:"status"`
	CreatedAt      string `json:"createdAt"`
}

// CancelOrderRequest represents the request for canceling an order
type CancelOrderRequest struct {
	OrderID string `json:"order_id"`
}

// GetOrdersResponse represents the response for getting orders
type GetOrdersResponse struct {
	Orders     []OrderResponse `json:"orders"`
	Pagination Pagination      `json:"pagination"`
}

// OrderResponse represents an order in API responses
type OrderResponse struct {
	ID             string `json:"id"`
	Symbol         string `json:"symbol"`
	Side           string `json:"side"`
	Type           string `json:"type"`
	Price          string `json:"price"`
	Quantity       string `json:"quantity"`
	FilledQuantity string `json:"filledQuantity"`
	Status         string `json:"status"`
	TimeInForce    string `json:"timeInForce"`
	TTL            int    `json:"ttl"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
}

// GetPositionResponse represents the response for getting a position
type GetPositionResponse struct {
	Symbol          string `json:"symbol"`
	Side            string `json:"side"`
	Quantity        string `json:"quantity"`
	EntryPrice      string `json:"entryPrice"`
	CurrentPrice    string `json:"currentPrice"`
	Leverage        int    `json:"leverage"`
	PositionValue   string `json:"positionValue"`
	LiquidationPrice string `json:"liquidationPrice"`
	UnrealizedPnl   string `json:"unrealizedPnl"`
}

// GetBalanceResponse represents the response for getting balance
type GetBalanceResponse struct {
	Balance         string `json:"balance"`
	FrozenMargin    string `json:"frozenMargin"`
	UnrealizedPnl   string `json:"unrealizedPnl"`
	TotalEquity     string `json:"totalEquity"`
	UsedMargin      string `json:"usedMargin"`
	TotalMargin     string `json:"totalMargin"`
	AvailableMargin string `json:"availableMargin"`
}

// getStrategyByAPIKey retrieves a strategy by API key
func (h *ExchangeHandler) getStrategyByAPIKey(apiKey string) (*models.Strategy, error) {
	var strategy models.Strategy
	if err := h.db.First(&strategy, "api_key = ?", apiKey).Error; err != nil {
		return nil, err
	}
	return &strategy, nil
}

// CreateOrder handles POST /exchange/createOrder
func (h *ExchangeHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	// Get API key from header
	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		h.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing API key")
		return
	}

	// Find strategy
	strategy, err := h.getStrategyByAPIKey(apiKey)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "INVALID_API_KEY", "Invalid API key")
		return
	}

	// Check if strategy is enabled
	if !strategy.Enabled {
		h.writeError(w, http.StatusForbidden, "STRATEGY_DISABLED", "Strategy is disabled")
		return
	}

	// Check if balance is negative (liquidated)
	if strategy.Balance.LessThan(decimal.Zero) {
		h.writeError(w, http.StatusForbidden, "STRATEGY_LIQUIDATED", "Strategy has been liquidated")
		return
	}

	// Parse request
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Validate required fields
	if req.Symbol == "" {
		h.writeError(w, http.StatusBadRequest, "INVALID_SYMBOL", "Symbol is required")
		return
	}
	if req.Side == "" || (req.Side != "buy" && req.Side != "sell") {
		h.writeError(w, http.StatusBadRequest, "INVALID_SIDE", "Side must be buy or sell")
		return
	}
	if req.Quantity == "" {
		h.writeError(w, http.StatusBadRequest, "INVALID_QUANTITY", "Quantity is required")
		return
	}

	// Parse quantity
	quantity, err := decimal.NewFromString(req.Quantity)
	if err != nil || quantity.LessThan(decimal.NewFromFloat(0.001)) { // Min quantity 0.001
		h.writeError(w, http.StatusBadRequest, "INVALID_QUANTITY", "Invalid quantity")
		return
	}

	// Determine order type
	orderType := "limit"
	if req.Type != "" {
		if req.Type != "limit" && req.Type != "market" {
			h.writeError(w, http.StatusBadRequest, "INVALID_TYPE", "Type must be limit or market")
			return
		}
		orderType = req.Type
	}

	// Validate price for limit orders
	var price decimal.Decimal
	if orderType == "limit" {
		if req.Price == "" {
			h.writeError(w, http.StatusBadRequest, "INVALID_PRICE", "Price is required for limit orders")
			return
		}
		price, err = decimal.NewFromString(req.Price)
		if err != nil || price.LessThanOrEqual(decimal.Zero) {
			h.writeError(w, http.StatusBadRequest, "INVALID_PRICE", "Invalid price")
			return
		}

		// 检查挂单是否会导致立即成交
		ticker := h.matchingEngine.GetTicker(req.Symbol)
		if ticker == nil {
			h.writeError(w, http.StatusBadRequest, "NO_MARKET_DATA", "No market data available for this symbol")
			return
		}

		// 检查是否会立即成交
		if req.Side == "buy" {
			// 买单：价格上穿时成交（当前价格 > 挂单价 且 上一笔价格 <= 挂单价）
			if ticker.Price.GreaterThan(price) && ticker.PreviousPrice.LessThanOrEqual(price) {
				h.writeError(w, http.StatusBadRequest, "ORDER_WOULD_IMMEDIATELY_FILL", "Buy order would immediately fill")
				return
			}
		} else if req.Side == "sell" {
			// 卖单：价格下穿时成交（当前价格 < 挂单价 且 上一笔价格 >= 挂单价）
			if ticker.Price.LessThan(price) && ticker.PreviousPrice.GreaterThanOrEqual(price) {
				h.writeError(w, http.StatusBadRequest, "ORDER_WOULD_IMMEDIATELY_FILL", "Sell order would immediately fill")
				return
			}
		}
	} else {
		// Market order - get current price from ticker
		ticker := h.matchingEngine.GetTicker(req.Symbol)
		if ticker == nil {
			h.writeError(w, http.StatusBadRequest, "SYMBOL_NOT_FOUND", "Symbol not found")
			return
		}
		price = ticker.Price
	}

	// Validate timeInForce
	timeInForce := "GTC"
	if req.TimeInForce != "" {
		if req.TimeInForce != "GTC" && req.TimeInForce != "IOC" && req.TimeInForce != "FOK" {
			h.writeError(w, http.StatusBadRequest, "INVALID_TIME_IN_FORCE", "TimeInForce must be GTC, IOC, or FOK")
			return
		}
		timeInForce = req.TimeInForce
	}

	// Calculate required margin
	marginRequired := price.Mul(quantity).Div(decimal.NewFromInt(100)) // 100x leverage = 1% margin
	availableBalance := strategy.Balance.Sub(strategy.FrozenMargin)
	if marginRequired.GreaterThan(availableBalance) {
		h.writeError(w, http.StatusBadRequest, "INSUFFICIENT_BALANCE", "Insufficient balance")
		return
	}

	// Create order in database
	order := models.Order{
		ID:          uuid.New().String(),
		StrategyID:  strategy.ID,
		Symbol:      req.Symbol,
		Side:        req.Side,
		Type:        orderType,
		Price:       price,
		Quantity:    quantity,
		Status:      "open",
		TimeInForce: timeInForce,
		TTL:         req.TTL,
	}

	if err := h.db.Create(&order).Error; err != nil {
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create order")
		return
	}

	// Freeze margin for limit orders (not for IOC/FOK/market which will be processed immediately)
	if orderType == "limit" && timeInForce == "GTC" {
		h.db.Model(&models.Strategy{}).Where("id = ?", strategy.ID).UpdateColumn("frozen_margin", gorm.Expr("frozen_margin + ?", marginRequired))
	}

	// Add to matching engine
	h.matchingEngine.AddOrder(&engine.Order{
		ID:          order.ID,
		StrategyID:  strategy.ID,
		Symbol:      order.Symbol,
		Side:        order.Side,
		Type:        order.Type,
		Price:       order.Price,
		Quantity:    order.Quantity,
		FilledQty:   decimal.Zero,
		Status:      "open",
		TimeInForce: order.TimeInForce,
		TTL:         order.TTL,
		CreatedAt:   order.CreatedAt,
	})

	// Handle IOC/FOK orders immediately
	if orderType == "market" || timeInForce == "IOC" || timeInForce == "FOK" {
		h.matchingEngine.ProcessOrder(order.ID)
	}

	response := CreateOrderResponse{
		ID:             order.ID,
		Symbol:         order.Symbol,
		Side:           order.Side,
		Type:           order.Type,
		Price:          order.Price.String(),
		Quantity:       order.Quantity.String(),
		FilledQuantity: order.FilledQuantity.String(),
		Status:         order.Status,
		CreatedAt:      order.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	h.writeJSON(w, http.StatusCreated, response)
}

// CancelOrder handles POST /exchange/cancelOrder
func (h *ExchangeHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	// Get API key from header
	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		h.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing API key")
		return
	}

	// Find strategy
	strategy, err := h.getStrategyByAPIKey(apiKey)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "INVALID_API_KEY", "Invalid API key")
		return
	}

	// Parse request
	var req CancelOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.OrderID == "" {
		h.writeError(w, http.StatusBadRequest, "INVALID_ORDER_ID", "Order ID is required")
		return
	}

	// Find order
	var order models.Order
	if err := h.db.First(&order, "id = ?", req.OrderID).Error; err != nil {
		h.writeError(w, http.StatusNotFound, "ORDER_NOT_FOUND", "Order not found")
		return
	}

	// Verify order belongs to the strategy
	if order.StrategyID != strategy.ID {
		h.writeError(w, http.StatusForbidden, "UNAUTHORIZED", "Order does not belong to this strategy")
		return
	}

	// Check if order can be canceled
	if order.Status != "open" {
		h.writeError(w, http.StatusBadRequest, "INVALID_ORDER", "Order cannot be canceled")
		return
	}

	// Cancel order in database
	order.Status = "canceled"
	order.UpdatedAt = time.Now()
	if err := h.db.Save(&order).Error; err != nil {
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to cancel order")
		return
	}

	// Unfreeze margin
	marginRequired := order.Price.Mul(order.Quantity).Div(decimal.NewFromInt(100))
	h.db.Model(&models.Strategy{}).Where("id = ?", order.StrategyID).UpdateColumn("frozen_margin", gorm.Expr("frozen_margin - ?", marginRequired))

	// Cancel order in matching engine
	h.matchingEngine.CancelOrder(order.ID)

	response := OrderResponse{
		ID:             order.ID,
		Symbol:         order.Symbol,
		Side:           order.Side,
		Type:           order.Type,
		Price:          order.Price.String(),
		Quantity:       order.Quantity.String(),
		FilledQuantity: order.FilledQuantity.String(),
		Status:         order.Status,
		TimeInForce:    order.TimeInForce,
		TTL:            order.TTL,
		CreatedAt:      order.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      order.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	h.writeJSON(w, http.StatusOK, response)
}

// GetOrders handles GET /exchange/getOrders
func (h *ExchangeHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	strategyID := r.FormValue("strategy_id")
	if strategyID == "" {
		h.writeError(w, http.StatusBadRequest, "INVALID_STRATEGY_ID", "Strategy ID is required")
		return
	}

	// Parse pagination params
	page, _ := strconv.Atoi(r.FormValue("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.FormValue("limit"))
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	// Count total
	var total int64
	query := h.db.Model(&models.Order{})
	if strategyID != "" {
		query = query.Where("strategy_id = ?", strategyID)
	}
	query.Count(&total)

	// Fetch orders
	var orders []models.Order
	query = h.db.Where("strategy_id = ?", strategyID).Order("created_at DESC").Offset(offset).Limit(limit)
	if err := query.Find(&orders).Error; err != nil {
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Database error")
		return
	}

	// Build response
	orderResponses := make([]OrderResponse, len(orders))
	for i, o := range orders {
		orderResponses[i] = OrderResponse{
			ID:             o.ID,
			Symbol:         o.Symbol,
			Side:           o.Side,
			Type:           o.Type,
			Price:          o.Price.String(),
			Quantity:       o.Quantity.String(),
			FilledQuantity: o.FilledQuantity.String(),
			Status:         o.Status,
			TimeInForce:    o.TimeInForce,
			TTL:            o.TTL,
			CreatedAt:      o.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:      o.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	response := GetOrdersResponse{
		Orders: orderResponses,
		Pagination: Pagination{
			Page:       page,
			Limit:      limit,
			Total:      int(total),
			TotalPages: totalPages,
		},
	}

	h.writeJSON(w, http.StatusOK, response)
}

// GetOrder handles GET /exchange/getOrder/:id
func (h *ExchangeHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["id"]

	var order models.Order
	if err := h.db.First(&order, "id = ?", orderID).Error; err != nil {
		h.writeError(w, http.StatusNotFound, "ORDER_NOT_FOUND", "Order not found")
		return
	}

	response := OrderResponse{
		ID:             order.ID,
		Symbol:         order.Symbol,
		Side:           order.Side,
		Type:           order.Type,
		Price:          order.Price.String(),
		Quantity:       order.Quantity.String(),
		FilledQuantity: order.FilledQuantity.String(),
		Status:         order.Status,
		TimeInForce:    order.TimeInForce,
		TTL:            order.TTL,
		CreatedAt:      order.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      order.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	h.writeJSON(w, http.StatusOK, response)
}

// GetPosition handles GET /exchange/getPosition
func (h *ExchangeHandler) GetPosition(w http.ResponseWriter, r *http.Request) {
	strategyID := r.FormValue("strategy_id")
	if strategyID == "" {
		h.writeError(w, http.StatusBadRequest, "INVALID_STRATEGY_ID", "Strategy ID is required")
		return
	}

	symbol := r.FormValue("symbol")

	// Get positions from position manager
	positions := h.positionManager.GetPositions(strategyID, symbol)

	if len(positions) == 0 {
		h.writeJSON(w, http.StatusOK, map[string]interface{}{"positions": []interface{}{}})
		return
	}

	positionResponses := make([]GetPositionResponse, len(positions))
	for i, p := range positions {
		positionResponses[i] = GetPositionResponse{
			Symbol:          p.Symbol,
			Side:            p.Side,
			Quantity:        p.Quantity.String(),
			EntryPrice:      p.EntryPrice.String(),
			CurrentPrice:    p.CurrentPrice.String(),
			Leverage:        p.Leverage,
			PositionValue:   p.PositionValue.String(),
			LiquidationPrice: p.LiquidationPrice.String(),
			UnrealizedPnl:   p.UnrealizedPnl.String(),
		}
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{"positions": positionResponses})
}

// GetBalance handles GET /exchange/getBalance
func (h *ExchangeHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	strategyID := r.FormValue("strategy_id")
	if strategyID == "" {
		h.writeError(w, http.StatusBadRequest, "INVALID_STRATEGY_ID", "Strategy ID is required")
		return
	}

	// Find strategy
	var strategy models.Strategy
	if err := h.db.First(&strategy, "id = ?", strategyID).Error; err != nil {
		h.writeError(w, http.StatusNotFound, "STRATEGY_NOT_FOUND", "Strategy not found")
		return
	}

	// Calculate unrealized PnL
	unrealizedPnl := h.positionManager.CalculateUnrealizedPnl(strategyID)

	// Calculate total equity
	totalEquity := strategy.Balance.Add(unrealizedPnl)

	// Calculate used margin (position margin)
	usedMargin := h.positionManager.CalculateUsedMargin(strategyID)

	// Total margin = position margin + frozen margin (open orders)
	totalMargin := usedMargin.Add(strategy.FrozenMargin)

	// Calculate available margin
	availableMargin := strategy.Balance.Sub(totalMargin)

	response := GetBalanceResponse{
		Balance:         strategy.Balance.String(),
		FrozenMargin:    strategy.FrozenMargin.String(),
		UnrealizedPnl:   unrealizedPnl.String(),
		TotalEquity:     totalEquity.String(),
		UsedMargin:      usedMargin.String(),
		TotalMargin:     totalMargin.String(),
		AvailableMargin: availableMargin.String(),
	}

	h.writeJSON(w, http.StatusOK, response)
}

// writeJSON writes JSON response
func (h *ExchangeHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes error response
func (h *ExchangeHandler) writeError(w http.ResponseWriter, status int, code, message string) {
	h.writeJSON(w, status, ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}
