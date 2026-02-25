package router

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/maker-arena/backend/internal/config"
	"github.com/maker-arena/backend/internal/engine"
	"github.com/maker-arena/backend/internal/handlers"
	"gorm.io/gorm"
)

// Setup sets up the router with all routes
func Setup(db *gorm.DB, cfg *config.Config, matchingEngine *engine.MatchingEngine, positionManager *engine.PositionManager) *mux.Router {
	r := mux.NewRouter()

	// Create handlers
	strategyHandler := handlers.NewStrategyHandler(db, cfg)
	exchangeHandler := handlers.NewExchangeHandler(db, cfg, matchingEngine, positionManager)
	dataHandler := handlers.NewDataHandler(db, cfg, matchingEngine)

	// Serve static frontend files
	frontendDir := "./frontend"
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, frontendDir+"/index.html")
	})
	r.HandleFunc("/strategy", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, frontendDir+"/strategy.html")
	})
	r.PathPrefix("/css/").Handler(http.StripPrefix("/css/", http.FileServer(http.Dir(frontendDir+"/css"))))
	r.PathPrefix("/js/").Handler(http.StripPrefix("/js/", http.FileServer(http.Dir(frontendDir+"/js"))))

	// API v1 routes
	api := r.PathPrefix("/api/v1").Subrouter()

	// Strategy management routes (admin)
	api.HandleFunc("/strategies", strategyHandler.List).Methods("GET")
	api.HandleFunc("/strategies", strategyHandler.Create).Methods("POST")
	api.HandleFunc("/strategies/{id}", strategyHandler.Get).Methods("GET")
	api.HandleFunc("/strategies/{id}", strategyHandler.Update).Methods("PUT")
	api.HandleFunc("/strategies/{id}", strategyHandler.Delete).Methods("DELETE")

	// Exchange routes (authenticated)
	api.HandleFunc("/exchange/createOrder", exchangeHandler.CreateOrder).Methods("POST")
	api.HandleFunc("/exchange/cancelOrder", exchangeHandler.CancelOrder).Methods("POST")
	api.HandleFunc("/exchange/getOrders", exchangeHandler.GetOrders).Methods("GET")
	api.HandleFunc("/exchange/getOrder/{id}", exchangeHandler.GetOrder).Methods("GET")
	api.HandleFunc("/exchange/getPosition", exchangeHandler.GetPosition).Methods("GET")
	api.HandleFunc("/exchange/getBalance", exchangeHandler.GetBalance).Methods("GET")

	// Data routes (public)
	api.HandleFunc("/fills", dataHandler.GetFills).Methods("GET")
	api.HandleFunc("/snapshots/account", dataHandler.GetAccountSnapshots).Methods("GET")
	api.HandleFunc("/snapshots/position", dataHandler.GetPositionSnapshots).Methods("GET")
	api.HandleFunc("/liquidations", dataHandler.GetLiquidations).Methods("GET")
	api.HandleFunc("/market/ticker", dataHandler.GetTicker).Methods("GET")
	api.HandleFunc("/statistics", dataHandler.GetStatistics).Methods("GET")

	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	return r
}
