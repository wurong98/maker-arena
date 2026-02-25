package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/maker-arena/backend/internal/config"
	"github.com/maker-arena/backend/internal/engine"
	"github.com/maker-arena/backend/internal/models"
	"gorm.io/gorm"
)

// DataHandler handles data-related requests
type DataHandler struct {
	db             *gorm.DB
	cfg            *config.Config
	matchingEngine *engine.MatchingEngine
}

// NewDataHandler creates a new DataHandler
func NewDataHandler(db *gorm.DB, cfg *config.Config, me *engine.MatchingEngine) *DataHandler {
	return &DataHandler{
		db:             db,
		cfg:            cfg,
		matchingEngine: me,
	}
}

// FillResponse represents a fill in API responses
type FillResponse struct {
	ID         string `json:"id"`
	OrderID    string `json:"orderId"`
	StrategyID string `json:"strategyId"`
	Symbol     string `json:"symbol"`
	Side       string `json:"side"`
	Price      string `json:"price"`
	Quantity   string `json:"quantity"`
	Fee        string `json:"fee"`
	CreatedAt  string `json:"createdAt"`
}

// GetFillsResponse represents the response for getting fills
type GetFillsResponse struct {
	Fills      []FillResponse `json:"fills"`
	Pagination Pagination     `json:"pagination"`
}

// LiquidationResponse represents a liquidation in API responses
type LiquidationResponse struct {
	ID               string `json:"id"`
	StrategyID       string `json:"strategyId"`
	StrategyName     string `json:"strategyName"`
	Symbol           string `json:"symbol"`
	Side             string `json:"side"`
	LiquidationPrice string `json:"liquidationPrice"`
	Quantity         string `json:"quantity"`
	CreatedAt        string `json:"createdAt"`
}

// GetLiquidationsResponse represents the response for getting liquidations
type GetLiquidationsResponse struct {
	Liquidations []LiquidationResponse `json:"liquidations"`
	Pagination   Pagination           `json:"pagination"`
}

// AccountSnapshotResponse represents an account snapshot in API responses
type AccountSnapshotResponse struct {
	ID            string `json:"id"`
	StrategyID    string `json:"strategyId"`
	Balance       string `json:"balance"`
	UnrealizedPnl string `json:"unrealizedPnl"`
	TotalEquity   string `json:"totalEquity"`
	CreatedAt     string `json:"createdAt"`
}

// PositionSnapshotResponse represents a position snapshot in API responses
type PositionSnapshotResponse struct {
	ID            string `json:"id"`
	StrategyID    string `json:"strategyId"`
	Symbol        string `json:"symbol"`
	UnrealizedPnl string `json:"unrealizedPnl"`
	PositionValue string `json:"positionValue"`
	AvgPrice      string `json:"avgPrice"`
	CreatedAt     string `json:"createdAt"`
}

// TickerResponse represents a ticker in API responses
type TickerResponse struct {
	Symbol                string `json:"symbol"`
	Price                 string `json:"price"`
	PriceChange24h        string `json:"priceChange24h"`
	PriceChangePercent24h string `json:"priceChangePercent24h"`
	UpdatedAt             string `json:"updatedAt"`
}

// StatisticsResponse represents system statistics
type StatisticsResponse struct {
	TotalStrategies int64 `json:"totalStrategies"`
	TotalFills     int64 `json:"totalFills"`
	OpenOrders     int64 `json:"openOrders"`
}

// GetFills handles GET /fills
func (h *DataHandler) GetFills(w http.ResponseWriter, r *http.Request) {
	strategyID := r.FormValue("strategy_id")

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

	// Build query
	query := h.db.Model(&models.Fill{})
	if strategyID != "" {
		query = query.Where("strategy_id = ?", strategyID)
	}

	// Count total
	var total int64
	query.Count(&total)

	// Fetch fills
	var fills []models.Fill
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&fills).Error; err != nil {
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Database error")
		return
	}

	// Build response
	fillResponses := make([]FillResponse, len(fills))
	for i, f := range fills {
		fillResponses[i] = FillResponse{
			ID:         f.ID,
			OrderID:    f.OrderID,
			StrategyID: f.StrategyID,
			Symbol:     f.Symbol,
			Side:       f.Side,
			Price:      f.Price.String(),
			Quantity:   f.Quantity.String(),
			Fee:        f.Fee.String(),
			CreatedAt:  f.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	response := GetFillsResponse{
		Fills: fillResponses,
		Pagination: Pagination{
			Page:       page,
			Limit:      limit,
			Total:      int(total),
			TotalPages: totalPages,
		},
	}

	h.writeJSON(w, http.StatusOK, response)
}

// GetAccountSnapshots handles GET /snapshots/account
func (h *DataHandler) GetAccountSnapshots(w http.ResponseWriter, r *http.Request) {
	strategyID := r.FormValue("strategy_id")
	if strategyID == "" {
		h.writeError(w, http.StatusBadRequest, "INVALID_STRATEGY_ID", "Strategy ID is required")
		return
	}

	limit, _ := strconv.Atoi(r.FormValue("limit"))
	if limit < 1 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	// Parse time filters
	var startTime, endTime *time.Time
	if st := r.FormValue("start_time"); st != "" {
		if t, err := time.Parse(time.RFC3339, st); err == nil {
			startTime = &t
		}
	}
	if et := r.FormValue("end_time"); et != "" {
		if t, err := time.Parse(time.RFC3339, et); err == nil {
			endTime = &t
		}
	}

	// Build query
	query := h.db.Where("strategy_id = ?", strategyID)
	if startTime != nil {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime != nil {
		query = query.Where("created_at <= ?", endTime)
	}

	// Fetch snapshots
	var snapshots []models.AccountSnapshot
	if err := query.Order("created_at DESC").Limit(limit).Find(&snapshots).Error; err != nil {
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Database error")
		return
	}

	// Build response
	snapshotResponses := make([]AccountSnapshotResponse, len(snapshots))
	for i, s := range snapshots {
		snapshotResponses[i] = AccountSnapshotResponse{
			ID:            s.ID,
			StrategyID:    s.StrategyID,
			Balance:       s.Balance.String(),
			UnrealizedPnl: s.UnrealizedPnl.String(),
			TotalEquity:   s.TotalEquity.String(),
			CreatedAt:     s.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{"snapshots": snapshotResponses})
}

// GetPositionSnapshots handles GET /snapshots/position
func (h *DataHandler) GetPositionSnapshots(w http.ResponseWriter, r *http.Request) {
	strategyID := r.FormValue("strategy_id")
	if strategyID == "" {
		h.writeError(w, http.StatusBadRequest, "INVALID_STRATEGY_ID", "Strategy ID is required")
		return
	}

	symbol := r.FormValue("symbol")

	limit, _ := strconv.Atoi(r.FormValue("limit"))
	if limit < 1 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	// Parse time filters
	var startTime, endTime *time.Time
	if st := r.FormValue("start_time"); st != "" {
		if t, err := time.Parse(time.RFC3339, st); err == nil {
			startTime = &t
		}
	}
	if et := r.FormValue("end_time"); et != "" {
		if t, err := time.Parse(time.RFC3339, et); err == nil {
			endTime = &t
		}
	}

	// Build query
	query := h.db.Where("strategy_id = ?", strategyID)
	if symbol != "" {
		query = query.Where("symbol = ?", symbol)
	}
	if startTime != nil {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime != nil {
		query = query.Where("created_at <= ?", endTime)
	}

	// Fetch snapshots
	var snapshots []models.PositionSnapshot
	if err := query.Order("created_at DESC").Limit(limit).Find(&snapshots).Error; err != nil {
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Database error")
		return
	}

	// Build response
	snapshotResponses := make([]PositionSnapshotResponse, len(snapshots))
	for i, s := range snapshots {
		snapshotResponses[i] = PositionSnapshotResponse{
			ID:            s.ID,
			StrategyID:    s.StrategyID,
			Symbol:        s.Symbol,
			UnrealizedPnl: s.UnrealizedPnl.String(),
			PositionValue: s.PositionValue.String(),
			AvgPrice:      s.AvgPrice.String(),
			CreatedAt:     s.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{"snapshots": snapshotResponses})
}

// GetLiquidations handles GET /liquidations
func (h *DataHandler) GetLiquidations(w http.ResponseWriter, r *http.Request) {
	strategyID := r.FormValue("strategy_id")

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

	// Build query
	query := h.db.Model(&models.Liquidation{})
	if strategyID != "" {
		query = query.Where("strategy_id = ?", strategyID)
	}

	// Count total
	var total int64
	query.Count(&total)

	// Fetch liquidations
	var liquidations []models.Liquidation
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&liquidations).Error; err != nil {
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Database error")
		return
	}

	// Build response
	liquidationResponses := make([]LiquidationResponse, len(liquidations))
	for i, l := range liquidations {
		liquidationResponses[i] = LiquidationResponse{
			ID:               l.ID,
			StrategyID:       l.StrategyID,
			StrategyName:     l.StrategyName,
			Symbol:           l.Symbol,
			Side:             l.Side,
			LiquidationPrice: l.LiquidationPrice.String(),
			Quantity:         l.Quantity.String(),
			CreatedAt:        l.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	response := GetLiquidationsResponse{
		Liquidations: liquidationResponses,
		Pagination: Pagination{
			Page:       page,
			Limit:      limit,
			Total:      int(total),
			TotalPages: totalPages,
		},
	}

	h.writeJSON(w, http.StatusOK, response)
}

// GetTicker handles GET /market/ticker
func (h *DataHandler) GetTicker(w http.ResponseWriter, r *http.Request) {
	// Get all tickers from matching engine
	tickers := h.matchingEngine.GetAllTickers()

	tickerResponses := make([]TickerResponse, 0, len(tickers))
	for symbol, ticker := range tickers {
		tickerResponses = append(tickerResponses, TickerResponse{
			Symbol:                symbol,
			Price:                 ticker.Price.String(),
			PriceChange24h:        ticker.PriceChange24h.String(),
			PriceChangePercent24h: ticker.PriceChangePercent24h.String(),
			UpdatedAt:             ticker.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{"tickers": tickerResponses})
}

// GetStatistics handles GET /statistics
func (h *DataHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	// Count total strategies
	var totalStrategies int64
	h.db.Model(&models.Strategy{}).Count(&totalStrategies)

	// Count total fills
	var totalFills int64
	h.db.Model(&models.Fill{}).Count(&totalFills)

	// Count open orders
	var openOrders int64
	h.db.Model(&models.Order{}).Where("status = ?", "open").Count(&openOrders)

	response := StatisticsResponse{
		TotalStrategies: totalStrategies,
		TotalFills:     totalFills,
		OpenOrders:     openOrders,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// writeJSON writes JSON response
func (h *DataHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes error response
func (h *DataHandler) writeError(w http.ResponseWriter, status int, code, message string) {
	h.writeJSON(w, status, ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}
