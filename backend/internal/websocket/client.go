package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/maker-arena/backend/internal/engine"
	"github.com/shopspring/decimal"
)

// BinanceClient manages connection to Binance WebSocket
type BinanceClient struct {
	wsURL           string
	symbols         []string
	matchingEngine  *engine.MatchingEngine
	conn            *websocket.Conn
	mu              sync.Mutex
	connected       bool
	reconnectDelay  time.Duration
	maxReconnectDelay time.Duration
	stopChan        chan struct{}
	wg              sync.WaitGroup
}

// TradeMessage represents a Binance trade message
type TradeMessage struct {
	EventType     string `json:"e"`
	EventTime     int64  `json:"E"`
	Symbol        string `json:"s"`
	TradeID       int64  `json:"t"`
	Price         string `json:"p"`
	Quantity      string `json:"q"`
	BuyerOrderID  int64  `json:"b"`
	SellerOrderID int64  `json:"a"`
	TradeTime     int64  `json:"T"`
	IsBuyerMaker  bool   `json:"m"`
}

// NewBinanceClient creates a new Binance WebSocket client
func NewBinanceClient(wsURL string, symbols []string, matchingEngine *engine.MatchingEngine) *BinanceClient {
	return &BinanceClient{
		wsURL:             wsURL,
		symbols:           symbols,
		matchingEngine:    matchingEngine,
		reconnectDelay:    1 * time.Second,
		maxReconnectDelay: 30 * time.Second,
		stopChan:          make(chan struct{}),
	}
}

// Start connects to Binance WebSocket and starts handling messages
func (c *BinanceClient) Start() {
	c.wg.Add(1)
	go c.connect()

	log.Printf("Binance WebSocket client started for symbols: %v", c.symbols)
}

// Stop disconnects from Binance WebSocket
func (c *BinanceClient) Stop() {
	// Close WebSocket connection first to unblock ReadMessage
	c.mu.Lock()
	if c.conn != nil {
		c.conn.Close()
	}
	c.mu.Unlock()

	close(c.stopChan)
	c.wg.Wait()
	log.Println("Binance WebSocket client stopped")
}

// connect establishes WebSocket connection
func (c *BinanceClient) connect() {
	defer c.wg.Done()

	// Build stream URL: wss://fstream.binance.com/stream?streams=btcusdc@trade/ethusdc@trade/...
	streams := make([]string, len(c.symbols))
	for i, symbol := range c.symbols {
		streams[i] = fmt.Sprintf("%s@trade", strings.ToLower(symbol))
	}
	streamURL := c.wsURL + "/stream?streams=" + strings.Join(streams, "/")

	for {
		select {
		case <-c.stopChan:
			return
		default:
		}

		log.Printf("Connecting to Binance WebSocket: %s", streamURL)

		// Dial WebSocket
		conn, _, err := websocket.DefaultDialer.Dial(streamURL, http.Header{})
		if err != nil {
			log.Printf("Failed to connect to Binance WebSocket: %v", err)
			c.scheduleReconnect()
			continue
		}

		c.mu.Lock()
		c.conn = conn
		c.connected = true
		c.reconnectDelay = 1 * time.Second // Reset reconnect delay
		c.mu.Unlock()

		log.Println("Connected to Binance WebSocket")

		// Handle messages
		c.handleMessages()

		// Check if stopped
		select {
		case <-c.stopChan:
			return
		default:
		}

		// Schedule reconnect
		c.scheduleReconnect()
	}
}

// handleMessages reads and processes messages from WebSocket
func (c *BinanceClient) handleMessages() {
	c.conn.SetReadLimit(4096)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Start ping goroutine to keep connection alive
	pingStop := make(chan struct{})
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-pingStop:
				return
			case <-ticker.C:
				c.mu.Lock()
				if c.conn != nil {
					if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
						log.Printf("Failed to send ping: %v", err)
						c.mu.Unlock()
						c.conn.Close()
						return
					}
				}
				c.mu.Unlock()
			}
		}
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			close(pingStop)
			return
		}

		c.processMessage(message)
	}
}

// processMessage processes a single WebSocket message
func (c *BinanceClient) processMessage(message []byte) {
	// Try to parse as trade message
	var trade TradeMessage
	if err := json.Unmarshal(message, &trade); err != nil {
		// Not a trade message, ignore
		return
	}

	// Check if it's a trade event
	if trade.EventType != "trade" {
		return
	}

	// Parse price and quantity
	price, err := decimal.NewFromString(trade.Price)
	if err != nil {
		log.Printf("Failed to parse price: %v", err)
		return
	}

	quantity, err := decimal.NewFromString(trade.Quantity)
	if err != nil {
		log.Printf("Failed to parse quantity: %v", err)
		return
	}

	// Convert symbol to lowercase for consistency
	symbol := strings.ToLower(trade.Symbol)

	// Forward to matching engine
	if c.matchingEngine != nil {
		c.matchingEngine.HandleTrade(symbol, price, quantity, trade.TradeTime)
	}
}

// scheduleReconnect schedules a reconnection attempt with exponential backoff
func (c *BinanceClient) scheduleReconnect() {
	c.mu.Lock()
	delay := c.reconnectDelay
	c.reconnectDelay = c.reconnectDelay * 2
	if c.reconnectDelay > c.maxReconnectDelay {
		c.reconnectDelay = c.maxReconnectDelay
	}
	c.connected = false
	c.mu.Unlock()

	log.Printf("Scheduling reconnect in %v", delay)

	select {
	case <-c.stopChan:
		return
	case <-time.After(delay):
	}
}

// IsConnected returns connection status
func (c *BinanceClient) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected
}
