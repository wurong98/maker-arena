package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

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
	_ = router.Setup(db, cfg)

	// Start server in goroutine
	// TODO: Implement server startup

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	// TODO: Implement graceful shutdown
}
