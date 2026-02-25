package engine

import (
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

// MarketData manages market ticker data independently
// This eliminates the circular dependency between MatchingEngine and PositionManager
type MarketData struct {
	mu       sync.RWMutex
	tickers  map[string]*Ticker
}

// NewMarketData creates a new MarketData instance
func NewMarketData() *MarketData {
	return &MarketData{
		tickers: make(map[string]*Ticker),
	}
}

// UpdateTicker updates the ticker for a symbol
func (md *MarketData) UpdateTicker(symbol string, price, previousPrice decimal.Decimal) {
	md.mu.Lock()
	defer md.mu.Unlock()

	ticker, exists := md.tickers[symbol]
	if !exists {
		ticker = &Ticker{Symbol: symbol}
		md.tickers[symbol] = ticker
	}

	ticker.PreviousPrice = ticker.Price
	ticker.Price = price
	if !ticker.PreviousPrice.IsZero() {
		ticker.PriceChange24h = price.Sub(ticker.PreviousPrice)
		if ticker.PreviousPrice.GreaterThan(decimal.Zero) {
			ticker.PriceChangePercent24h = ticker.PriceChange24h.Div(ticker.PreviousPrice).Mul(decimal.NewFromInt(100))
		}
	}
	ticker.UpdatedAt = time.Now()
}

// GetTicker returns a copy of the ticker for a symbol
func (md *MarketData) GetTicker(symbol string) *Ticker {
	md.mu.RLock()
	defer md.mu.RUnlock()

	if ticker, ok := md.tickers[symbol]; ok {
		t := *ticker
		return &t
	}
	return nil
}

// GetAllTickers returns all tickers (copies)
func (md *MarketData) GetAllTickers() map[string]*Ticker {
	md.mu.RLock()
	defer md.mu.RUnlock()

	result := make(map[string]*Ticker)
	for k, v := range md.tickers {
		t := *v
		result[k] = &t
	}
	return result
}
