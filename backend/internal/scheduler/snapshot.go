package scheduler

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/maker-arena/backend/internal/engine"
	"github.com/maker-arena/backend/internal/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// SnapshotScheduler periodically records account and position snapshots
type SnapshotScheduler struct {
	db              *gorm.DB
	interval        time.Duration
	positionManager *engine.PositionManager
	matchingEngine *engine.MatchingEngine
	stopChan        chan struct{}
	wg              sync.WaitGroup
}

// NewSnapshotScheduler creates a new snapshot scheduler
func NewSnapshotScheduler(db *gorm.DB, interval time.Duration, pm *engine.PositionManager, me *engine.MatchingEngine) *SnapshotScheduler {
	return &SnapshotScheduler{
		db:              db,
		interval:        interval,
		positionManager: pm,
		matchingEngine:  me,
		stopChan:        make(chan struct{}),
	}
}

// Start starts the snapshot scheduler
func (s *SnapshotScheduler) Start() {
	s.wg.Add(1)
	go s.run()

	log.Printf("Snapshot scheduler started with interval: %v", s.interval)
}

// Stop stops the snapshot scheduler
func (s *SnapshotScheduler) Stop() {
	close(s.stopChan)
	s.wg.Wait()

	log.Println("Snapshot scheduler stopped")
}

// run runs the snapshot scheduler
func (s *SnapshotScheduler) run() {
	defer s.wg.Done()

	// Run immediately on start
	s.RecordSnapshots()

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.RecordSnapshots()
		}
	}
}

// RecordSnapshots records snapshots for all strategies
func (s *SnapshotScheduler) RecordSnapshots() {
	// Get all strategies
	var strategies []models.Strategy
	if err := s.db.Find(&strategies).Error; err != nil {
		log.Printf("Failed to get strategies: %v", err)
		return
	}

	for _, strategy := range strategies {
		s.recordStrategySnapshot(strategy)
	}
}

// recordStrategySnapshot records snapshots for a single strategy
func (s *SnapshotScheduler) recordStrategySnapshot(strategy models.Strategy) {
	// Get unrealized PnL
	unrealizedPnl := decimal.Zero
	if s.positionManager != nil {
		unrealizedPnl = s.positionManager.CalculateUnrealizedPnl(strategy.ID)
	}

	// Calculate total equity
	totalEquity := strategy.Balance.Add(unrealizedPnl)

	// Create account snapshot
	accountSnapshot := models.AccountSnapshot{
		ID:            uuid.New().String(),
		StrategyID:    strategy.ID,
		Balance:       strategy.Balance,
		UnrealizedPnl: unrealizedPnl,
		TotalEquity:   totalEquity,
		CreatedAt:     time.Now(),
	}

	if err := s.db.Create(&accountSnapshot).Error; err != nil {
		log.Printf("Failed to create account snapshot for strategy %s: %v", strategy.ID, err)
	}

	// Get positions
	if s.positionManager != nil {
		positions := s.positionManager.GetPositions(strategy.ID, "")

		for _, pos := range positions {
			// Get current price from ticker
			currentPrice := pos.CurrentPrice
			if currentPrice.IsZero() && s.matchingEngine != nil {
				ticker := s.matchingEngine.GetTicker(pos.Symbol)
				if ticker != nil {
					currentPrice = ticker.Price
				}
			}

			// Calculate position value
			positionValue := currentPrice.Mul(pos.Quantity)

			// Calculate unrealized PnL for this position
			var unrealizedPnl decimal.Decimal
			if pos.Side == "long" {
				unrealizedPnl = currentPrice.Sub(pos.EntryPrice).Mul(pos.Quantity)
			} else {
				unrealizedPnl = pos.EntryPrice.Sub(currentPrice).Mul(pos.Quantity)
			}

			// Create position snapshot
			positionSnapshot := models.PositionSnapshot{
				ID:             uuid.New().String(),
				StrategyID:     strategy.ID,
				Symbol:         pos.Symbol,
				UnrealizedPnl:  unrealizedPnl,
				PositionValue:  positionValue,
				AvgPrice:       pos.EntryPrice,
				CreatedAt:      time.Now(),
			}

			if err := s.db.Create(&positionSnapshot).Error; err != nil {
				log.Printf("Failed to create position snapshot for strategy %s, symbol %s: %v",
					strategy.ID, pos.Symbol, err)
			}
		}
	}

	fmt.Printf("Recorded snapshots for strategy %s\n", strategy.ID)
}
