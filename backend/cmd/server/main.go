package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/maker-arena/backend/internal/config"
	"github.com/maker-arena/backend/internal/database"
	"github.com/maker-arena/backend/internal/engine"
	"github.com/maker-arena/backend/internal/router"
	"github.com/maker-arena/backend/internal/scheduler"
	"github.com/maker-arena/backend/internal/websocket"
	"github.com/shopspring/decimal"
)

func main() {
	// Load configuration
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run database migrations
	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Parse maker fee rate
	makerFeeRate, err := decimal.NewFromString(cfg.Trading.MakerFeeRate)
	if err != nil {
		log.Fatalf("Invalid maker fee rate: %v", err)
	}

	// Create position manager
	positionManager := engine.NewPositionManager(db, 100, func(symbol string) *engine.Ticker {
		return nil // Will be set later
	})
	positionManager.Start()

	// Create matching engine
	matchingEngine := engine.NewMatchingEngine(db, makerFeeRate, positionManager)
	matchingEngine.Start()

	// Set ticker getter for position manager
	positionManager.SetTickerGetter(func(symbol string) *engine.Ticker {
		return matchingEngine.GetTicker(symbol)
	})

	// Create snapshot scheduler
	snapshotInterval, err := cfg.Snapshot.IntervalDuration()
	if err != nil {
		log.Fatalf("Invalid snapshot interval: %v", err)
	}
	snapshotScheduler := scheduler.NewSnapshotScheduler(db, snapshotInterval, positionManager, matchingEngine)
	snapshotScheduler.Start()

	// Create Binance WebSocket client
	binanceClient := websocket.NewBinanceClient(cfg.Binance.WSURL, cfg.Symbols, matchingEngine)
	binanceClient.Start()

	// Create router
	r := router.Setup(db, cfg, matchingEngine, positionManager)

	// Create HTTP server
	srv := &http.Server{
		Addr:    cfg.App.Addr(),
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting HTTP server on %s", cfg.App.Addr())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown steps according to design doc section 6.6:
	// 1. Stop accepting new orders (set a flag or close listener)
	// 2. Wait for in-flight orders to complete
	// 3. Save all in-memory state to database
	// 4. Close WebSocket connections
	// 5. Shutdown HTTP server

	// Stop Binance WebSocket client
	binanceClient.Stop()

	// Stop snapshot scheduler
	snapshotScheduler.Stop()

	// Stop matching engine
	matchingEngine.Stop()

	// Stop position manager
	positionManager.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Step 5: Shutdown HTTP server (this will stop accepting new connections)
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
