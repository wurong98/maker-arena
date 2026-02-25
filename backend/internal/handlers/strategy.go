package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/maker-arena/backend/internal/config"
	"github.com/maker-arena/backend/internal/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// StrategyHandler handles strategy-related requests
type StrategyHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

// NewStrategyHandler creates a new StrategyHandler
func NewStrategyHandler(db *gorm.DB, cfg *config.Config) *StrategyHandler {
	return &StrategyHandler{db: db, cfg: cfg}
}

// ListStrategyResponse represents the response for listing strategies
type ListStrategyResponse struct {
	Strategies []StrategyResponse `json:"strategies"`
	Pagination Pagination        `json:"pagination"`
}

// StrategyResponse represents a strategy in API responses
type StrategyResponse struct {
	ID          string `json:"id"`
	APIKey      string `json:"apiKey"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	Balance     string `json:"balance"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// Pagination represents pagination info
type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail represents error details
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// List handles GET /strategies
func (h *StrategyHandler) List(w http.ResponseWriter, r *http.Request) {
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
	h.db.Model(&models.Strategy{}).Count(&total)

	// Fetch strategies
	var strategies []models.Strategy
	if err := h.db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&strategies).Error; err != nil {
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Database error")
		return
	}

	// Build response
	strategyResponses := make([]StrategyResponse, len(strategies))
	for i, s := range strategies {
		strategyResponses[i] = StrategyResponse{
			ID:          s.ID,
			APIKey:      s.APIKey,
			Name:        s.Name,
			Description: s.Description,
			Enabled:     s.Enabled,
			Balance:     s.Balance.String(),
			CreatedAt:   s.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   s.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	response := ListStrategyResponse{
		Strategies: strategyResponses,
		Pagination: Pagination{
			Page:       page,
			Limit:      limit,
			Total:      int(total),
			TotalPages: totalPages,
		},
	}

	h.writeJSON(w, http.StatusOK, response)
}

// CreateRequest represents the request for creating a strategy
type CreateStrategyRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Balance     string `json:"balance"`
	APIKey      string `json:"api_key"`
}

// Create handles POST /strategies
func (h *StrategyHandler) Create(w http.ResponseWriter, r *http.Request) {
	// Verify admin password
	password := r.Header.Get("X-Admin-Password")
	if password != h.cfg.Admin.Password {
		h.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid admin password")
		return
	}

	// Parse request
	var req CreateStrategyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Validate required fields
	if req.Name == "" {
		h.writeError(w, http.StatusBadRequest, "INVALID_NAME", "Name is required")
		return
	}
	if len(req.Name) > 128 {
		h.writeError(w, http.StatusBadRequest, "INVALID_NAME", "Name must be less than 128 characters")
		return
	}
	if len(req.Description) > 1000 {
		h.writeError(w, http.StatusBadRequest, "INVALID_DESCRIPTION", "Description must be less than 1000 characters")
		return
	}

	// Parse balance
	balance := decimal.NewFromInt(5000) // default
	if req.Balance != "" {
		b, err := decimal.NewFromString(req.Balance)
		if err != nil || b.LessThan(decimal.Zero) {
			h.writeError(w, http.StatusBadRequest, "INVALID_BALANCE", "Invalid balance")
			return
		}
		balance = b
	}

	// Generate API key if not provided
	apiKey := req.APIKey
	if apiKey == "" {
		apiKey = uuid.New().String()
	} else {
		// Check if API key already exists
		var count int64
		h.db.Model(&models.Strategy{}).Where("api_key = ?", apiKey).Count(&count)
		if count > 0 {
			h.writeError(w, http.StatusBadRequest, "API_KEY_EXISTS", "API key already exists")
			return
		}
	}

	// Create strategy
	strategy := models.Strategy{
		ID:          uuid.New().String(),
		APIKey:      apiKey,
		Name:        req.Name,
		Description: req.Description,
		Enabled:     true,
		Balance:     balance,
	}

	if err := h.db.Create(&strategy).Error; err != nil {
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create strategy")
		return
	}

	response := StrategyResponse{
		ID:          strategy.ID,
		APIKey:      strategy.APIKey,
		Name:        strategy.Name,
		Description: strategy.Description,
		Enabled:     strategy.Enabled,
		Balance:     strategy.Balance.String(),
		CreatedAt:   strategy.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   strategy.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	h.writeJSON(w, http.StatusCreated, response)
}

// UpdateRequest represents the request for updating a strategy
type UpdateStrategyRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     *bool  `json:"enabled"`
	Balance     string `json:"balance"`
}

// Update handles PUT /strategies/:id
func (h *StrategyHandler) Update(w http.ResponseWriter, r *http.Request) {
	// Verify admin password
	password := r.Header.Get("X-Admin-Password")
	if password != h.cfg.Admin.Password {
		h.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid admin password")
		return
	}

	// Get strategy ID from URL
	vars := mux.Vars(r)
	id := vars["id"]

	// Find strategy
	var strategy models.Strategy
	if err := h.db.First(&strategy, "id = ?", id).Error; err != nil {
		h.writeError(w, http.StatusNotFound, "STRATEGY_NOT_FOUND", "Strategy not found")
		return
	}

	// Parse request
	var req UpdateStrategyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Update fields
	if req.Name != "" {
		if len(req.Name) > 128 {
			h.writeError(w, http.StatusBadRequest, "INVALID_NAME", "Name must be less than 128 characters")
			return
		}
		strategy.Name = req.Name
	}
	if req.Description != "" {
		if len(req.Description) > 1000 {
			h.writeError(w, http.StatusBadRequest, "INVALID_DESCRIPTION", "Description must be less than 1000 characters")
			return
		}
		strategy.Description = req.Description
	}
	if req.Enabled != nil {
		strategy.Enabled = *req.Enabled
	}
	if req.Balance != "" {
		b, err := decimal.NewFromString(req.Balance)
		if err != nil || b.LessThan(decimal.Zero) {
			h.writeError(w, http.StatusBadRequest, "INVALID_BALANCE", "Invalid balance")
			return
		}
		strategy.Balance = b
	}

	if err := h.db.Save(&strategy).Error; err != nil {
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update strategy")
		return
	}

	response := StrategyResponse{
		ID:          strategy.ID,
		APIKey:      strategy.APIKey,
		Name:        strategy.Name,
		Description: strategy.Description,
		Enabled:     strategy.Enabled,
		Balance:     strategy.Balance.String(),
		CreatedAt:   strategy.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   strategy.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	h.writeJSON(w, http.StatusOK, response)
}

// Get handles GET /strategies/:id
func (h *StrategyHandler) Get(w http.ResponseWriter, r *http.Request) {
	// Get strategy ID from URL
	vars := mux.Vars(r)
	id := vars["id"]

	// Find strategy
	var strategy models.Strategy
	if err := h.db.First(&strategy, "id = ?", id).Error; err != nil {
		h.writeError(w, http.StatusNotFound, "STRATEGY_NOT_FOUND", "Strategy not found")
		return
	}

	response := StrategyResponse{
		ID:          strategy.ID,
		APIKey:      strategy.APIKey,
		Name:        strategy.Name,
		Description: strategy.Description,
		Enabled:     strategy.Enabled,
		Balance:     strategy.Balance.String(),
		CreatedAt:   strategy.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   strategy.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	h.writeJSON(w, http.StatusOK, response)
}

// Delete handles DELETE /strategies/:id
func (h *StrategyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// Verify admin password
	password := r.Header.Get("X-Admin-Password")
	if password != h.cfg.Admin.Password {
		h.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid admin password")
		return
	}

	// Get strategy ID from URL
	vars := mux.Vars(r)
	id := vars["id"]

	// Find strategy
	var strategy models.Strategy
	if err := h.db.First(&strategy, "id = ?", id).Error; err != nil {
		h.writeError(w, http.StatusNotFound, "STRATEGY_NOT_FOUND", "Strategy not found")
		return
	}

	// Delete strategy (cascade delete handled by foreign key constraints)
	if err := h.db.Delete(&strategy).Error; err != nil {
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete strategy")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetAPIKey handles GET /strategies/:id/api-key
func (h *StrategyHandler) GetAPIKey(w http.ResponseWriter, r *http.Request) {
	// Verify admin password
	password := r.Header.Get("X-Admin-Password")
	if password != h.cfg.Admin.Password {
		h.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid admin password")
		return
	}

	// Get strategy ID from URL
	vars := mux.Vars(r)
	id := vars["id"]

	// Find strategy
	var strategy models.Strategy
	if err := h.db.First(&strategy, "id = ?", id).Error; err != nil {
		h.writeError(w, http.StatusNotFound, "STRATEGY_NOT_FOUND", "Strategy not found")
		return
	}

	// Return API key
	h.writeJSON(w, http.StatusOK, map[string]string{
		"api_key": strategy.APIKey,
	})
}

// Helper function to write JSON response
func (h *StrategyHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Helper function to write error response
func (h *StrategyHandler) writeError(w http.ResponseWriter, status int, code, message string) {
	h.writeJSON(w, status, ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}
