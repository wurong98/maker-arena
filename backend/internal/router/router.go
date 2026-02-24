package router

import (
	"github.com/gorilla/mux"
	"github.com/maker-arena/backend/internal/config"
	"gorm.io/gorm"
)

// Setup sets up the router
func Setup(db *gorm.DB, cfg *config.Config) *mux.Router {
	// TODO: Implement router setup
	r := mux.NewRouter()
	return r
}
