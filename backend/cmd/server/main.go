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
	"github.com/maker-arena/backend/internal/router"
)

func main() {
	// Load configuration
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := database.Initialize(cfg.Database.DSN())
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create router
	r := router.Setup(db, cfg)

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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Step 5: Shutdown HTTP server (this will stop accepting new connections)
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Additional cleanup steps (to be implemented as services are added):
	// - Close WebSocket connections
	// - Save in-memory state to database
	// - Close database connections

	log.Println("Server exited properly")
}
