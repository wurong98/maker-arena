# 模拟交易系统后端实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 构建 Go + PostgreSQL 后端服务，提供模拟交易 API 和 WebSocket 数据流

**Architecture:**
- REST API 服务使用 Go 标准库 + gorilla/mux 路由
- 数据库使用 PostgreSQL，ORM 使用 GORM
- 币安 WebSocket 使用官方 SDK
- 撮合引擎使用内存撮合

**Tech Stack:** Go 1.21+, PostgreSQL, GORM, gorilla/mux, binance-go

---

## 阶段一：项目初始化

### Task 1: 初始化 Go 模块

**Files:**
- Create: `backend/go.mod`
- Create: `backend/cmd/server/main.go`
- Create: `backend/internal/config/config.go`
- Create: `backend/config.yaml`

**Step 1: 创建 go.mod**

```bash
mkdir -p backend/cmd/server backend/internal/config backend/internal/models backend/internal/handlers backend/internal/engine backend/internal/websocket
cd backend
go mod init github.com/maker-arena/backend
```

**Step 2: 创建配置文件**

```yaml
database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "password"
  dbname: "maker_arena"

app:
  host: "0.0.0.0"
  port: 8080

admin:
  password: "admin123"

binance:
  ws_url: "wss://stream.binance.com:9443/ws"

symbols:
  - "btcusdc"
  - "ethusdc"

snapshot:
  interval: "1h"

trading:
  maker_fee_rate: "0.0004"
  funding_rate: "0"
  funding_interval: "8h"
```

**Step 3: 创建 config.go**

```go
package config

import (
    "os"
    "github.com/joho/godotenv"
)

type Config struct {
    Database DatabaseConfig `yaml:"database"`
    App      AppConfig      `yaml:"app"`
    Admin    AdminConfig    `yaml:"admin"`
    Binance  BinanceConfig  `yaml:"binance"`
    Symbols  []string       `yaml:"symbols"`
    Snapshot SnapshotConfig `yaml:"snapshot"`
    Trading  TradingConfig  `yaml:"trading"`
}

type DatabaseConfig struct {
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    User     string `yaml:"user"`
    Password string `yaml:"password"`
    DBName   string `yaml:"dbname"`
}

type AppConfig struct {
    Host string `yaml:"host"`
    Port int    `yaml:"port"`
}

type AdminConfig struct {
    Password string `yaml:"password"`
}

type BinanceConfig struct {
    WSURL string `yaml:"ws_url"`
}

type SnapshotConfig struct {
    Interval string `yaml:"interval"`
}

type TradingConfig struct {
    MakerFeeRate  string `yaml:"maker_fee_rate"`
    FundingRate   string `yaml:"funding_rate"`
    FundingInterval string `yaml:"funding_interval"`
}

func Load(path string) (*Config, error) {
    err := godotenv.Load()
    if err != nil {
        // ignore if .env not found
    }

    // YAML loading implementation
    // ...
}
```

**Step 4: 提交**

```bash
git add backend/
git commit -m "feat: 初始化 Go 项目结构"
```

---

### Task 2: 数据库模型定义

**Files:**
- Create: `backend/internal/models/strategy.go`
- Create: `backend/internal/models/order.go`
- Create: `backend/internal/models/fill.go`
- Create: `backend/internal/models/position.go`
- Create: `backend/internal/models/liquidation.go`
- Create: `backend/internal/models/snapshot.go`
- Create: `backend/internal/models/ticker.go`

**Step 1: 创建 strategy.go**

```go
package models

import (
    "time"
    "github.com/shopspring/decimal"
)

type Strategy struct {
    ID          string          `gorm:"primaryKey;type:uuid" json:"id"`
    APIKey      string          `gorm:"uniqueIndex;type:varchar(64)" json:"apiKey"`
    Name        string          `gorm:"type:varchar(128)" json:"name"`
    Description string          `gorm:"type:text" json:"description"`
    Enabled     bool            `gorm:"default:true" json:"enabled"`
    Balance     decimal.Decimal `gorm:"type:numeric(20,8);default:5000" json:"balance"`
    CreatedAt   time.Time       `json:"createdAt"`
    UpdatedAt   time.Time       `json:"updatedAt"`
}
```

**Step 2: 创建 order.go**

```go
package models

type Order struct {
    ID            string          `gorm:"primaryKey;type:uuid" json:"id"`
    StrategyID    string          `gorm:"type:uuid;index" json:"strategyId"`
    Symbol        string          `gorm:"type:varchar(32)" json:"symbol"`
    Side          string          `gorm:"type:varchar(8)" json:"side"` // buy/sell
    Type          string          `gorm:"type:varchar(8)" json:"type"` // limit/market
    Price         decimal.Decimal `gorm:"type:numeric(20,8)" json:"price"`
    Quantity      decimal.Decimal `gorm:"type:numeric(20,8)" json:"quantity"`
    FilledQuantity decimal.Decimal `gorm:"type:numeric(20,8);default:0" json:"filledQuantity"`
    Status        string          `gorm:"type:varchar(16)" json:"status"` // open/filled/canceled/liquidated
    TimeInForce   string          `gorm:"type:varchar(8)" json:"timeInForce"`
    TTL           int             `gorm:"default:0" json:"ttl"`
    CreatedAt     time.Time       `json:"createdAt"`
    UpdatedAt     time.Time       `json:"updatedAt"`
}
```

**Step 3: 创建 fill.go**

```go
package models

type Fill struct {
    ID         string          `gorm:"primaryKey;type:uuid" json:"id"`
    OrderID    string          `gorm:"type:uuid;index" json:"orderId"`
    StrategyID string          `gorm:"type:uuid;index" json:"strategyId"`
    Symbol     string          `gorm:"type:varchar(32)" json:"symbol"`
    Side       string          `gorm:"type:varchar(8)" json:"side"`
    Price      decimal.Decimal `gorm:"type:numeric(20,8)" json:"price"`
    Quantity   decimal.Decimal `gorm:"type:numeric(20,8)" json:"quantity"`
    Fee        decimal.Decimal `gorm:"type:numeric(20,8)" json:"fee"`
    CreatedAt  time.Time       `json:"createdAt"`
}
```

**Step 4: 创建 position.go**

```go
package models

type Position struct {
    ID         string          `gorm:"primaryKey;type:uuid" json:"id"`
    StrategyID string          `gorm:"type:uuid;uniqueIndex:idx_strategy_symbol" json:"strategyId"`
    Symbol     string          `gorm:"type:varchar(32);uniqueIndex:idx_strategy_symbol" json:"symbol"`
    Side       string          `gorm:"type:varchar(8)" json:"side"` // long/short
    Quantity   decimal.Decimal `gorm:"type:numeric(20,8)" json:"quantity"`
    EntryPrice decimal.Decimal `gorm:"type:numeric(20,8)" json:"entryPrice"`
    Leverage   int             `gorm:"default:100" json:"leverage"`
    CreatedAt  time.Time       `json:"createdAt"`
    UpdatedAt  time.Time       `json:"updatedAt"`
}
```

**Step 5: 创建 liquidation.go**

```go
package models

type Liquidation struct {
    ID               string          `gorm:"primaryKey;type:uuid" json:"id"`
    StrategyID       string          `gorm:"type:uuid;index" json:"strategyId"`
    StrategyName     string          `gorm:"type:varchar(128)" json:"strategyName"`
    Symbol           string          `gorm:"type:varchar(32)" json:"symbol"`
    Side             string          `gorm:"type:varchar(8)" json:"side"`
    LiquidationPrice decimal.Decimal `gorm:"type:numeric(20,8)" json:"liquidationPrice"`
    Quantity         decimal.Decimal `gorm:"type:numeric(20,8)" json:"quantity"`
    CreatedAt        time.Time       `json:"createdAt"`
}
```

**Step 6: 创建 snapshot.go**

```go
package models

type AccountSnapshot struct {
    ID             string          `gorm:"primaryKey;type:uuid" json:"id"`
    StrategyID     string          `gorm:"type:uuid;index" json:"strategyId"`
    Balance        decimal.Decimal `gorm:"type:numeric(20,8)" json:"balance"`
    UnrealizedPnl  decimal.Decimal `gorm:"type:numeric(20,8)" json:"unrealizedPnl"`
    TotalEquity    decimal.Decimal `gorm:"type:numeric(20,8)" json:"totalEquity"`
    CreatedAt      time.Time       `json:"createdAt"`
}

type PositionSnapshot struct {
    ID             string          `gorm:"primaryKey;type:uuid" json:"id"`
    StrategyID     string          `gorm:"type:uuid;index" json:"strategyId"`
    Symbol         string          `gorm:"type:varchar(32)" json:"symbol"`
    UnrealizedPnl  decimal.Decimal `gorm:"type:numeric(20,8)" json:"unrealizedPnl"`
    PositionValue  decimal.Decimal `gorm:"type:numeric(20,8)" json:"positionValue"`
    AvgPrice       decimal.Decimal `gorm:"type:numeric(20,8)" json:"avgPrice"`
    CreatedAt      time.Time       `json:"createdAt"`
}
```

**Step 7: 创建 ticker.go**

```go
package models

type Ticker struct {
    Symbol              string          `gorm:"primaryKey;type:varchar(32)" json:"symbol"`
    Price               decimal.Decimal `gorm:"type:numeric(20,8)" json:"price"`
    PriceChange24h      decimal.Decimal `gorm:"type:numeric(20,8)" json:"priceChange24h"`
    PriceChangePercent24h decimal.Decimal `gorm:"type:numeric(20,8)" json:"priceChangePercent24h"`
    UpdatedAt           time.Time       `json:"updatedAt"`
}
```

**Step 8: 提交**

```bash
git add backend/internal/models/
git commit -m "feat: 定义数据库模型"
```

---

### Task 3: 数据库迁移

**Files:**
- Create: `backend/internal/database/migration.go`

**Step 1: 创建迁移文件**

```go
package database

import (
    "log"
    "github.com/maker-arena/backend/internal/models"
    "gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
    log.Println("Running database migrations...")

    err := db.AutoMigrate(
        &models.Strategy{},
        &models.Order{},
        &models.Fill{},
        &models.Position{},
        &models.Liquidation{},
        &models.AccountSnapshot{},
        &models.PositionSnapshot{},
        &models.Ticker{},
    )

    if err != nil {
        return err
    }

    log.Println("Database migrations completed")
    return nil
}
```

**Step 2: 提交**

```bash
git add backend/internal/database/
git commit -m "feat: 添加数据库迁移"
```

---

## 阶段二：核心 API 实现

### Task 4: 策略管理 API

**Files:**
- Modify: `backend/internal/handlers/strategy.go` (new)
- Modify: `backend/internal/models/strategy.go` (add TableName)

**Step 1: 添加 TableName 方法**

```go
func (Strategy) TableName() string {
    return "strategies"
}
```

**Step 2: 创建 handlers/strategy.go**

```go
package handlers

import (
    "net/http"
    "github.com/google/uuid"
    "github.com/maker-arena/backend/internal/models"
    "github.com/shopspring/decimal"
    "gorm.io/gorm"
)

type StrategyHandler struct {
    DB *gorm.DB
}

func NewStrategyHandler(db *gorm.DB) *StrategyHandler {
    return &StrategyHandler{DB: db}
}

func (h *StrategyHandler) List(w http.ResponseWriter, r *http.Request) {
    // 实现分页逻辑
    page := r.URL.Query().Get("page")
    limit := r.URL.Query().Get("limit")
    // ...
}

func (h *StrategyHandler) Create(w http.ResponseWriter, r *http.Request) {
    // 创建策略
}

func (h *StrategyHandler) Get(w http.ResponseWriter, r *http.Request) {
    // 获取单个策略
}

func (h *StrategyHandler) Update(w http.ResponseWriter, r *http.Request) {
    // 更新策略
}

func (h *StrategyHandler) Delete(w http.ResponseWriter, r *http.Request) {
    // 删除策略（级联删除）
}
```

**Step 3: 提交**

```bash
git add backend/internal/handlers/
git commit -m "feat: 实现策略管理 API"
```

---

### Task 5: 交易接口 (createOrder, cancelOrder, getOrders, getPosition, getBalance)

**Files:**
- Create: `backend/internal/handlers/exchange.go`

**Step 1: 创建 exchange.go**

```go
package handlers

type ExchangeHandler struct {
    DB *gorm.DB
    // 订单引擎引用
}

func NewExchangeHandler(db *gorm.DB) *ExchangeHandler {
    return &ExchangeHandler{DB: db}
}

func (h *ExchangeHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
    // 解析请求，创建订单
    // 验证余额
    // 添加到撮合引擎
}

func (h *ExchangeHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
    // 取消订单
}

func (h *ExchangeHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
    // 查询订单列表
}

func (h *ExchangeHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
    // 查询单个订单
}

func (h *ExchangeHandler) GetPosition(w http.ResponseWriter, r *http.Request) {
    // 查询持仓
}

func (h *ExchangeHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
    // 查询余额
}
```

**Step 2: 提交**

```bash
git add backend/internal/handlers/exchange.go
git commit -m "feat: 实现交易接口"
```

---

### Task 6: 数据接口 (fills, snapshots, liquidations, ticker, statistics)

**Files:**
- Create: `backend/internal/handlers/data.go`

**Step 1: 创建 data.go**

```go
package handlers

func (h *Handler) GetFills(w http.ResponseWriter, r *http.Request) {
    // 成交记录
}

func (h *Handler) GetAccountSnapshots(w http.ResponseWriter, r *http.Request) {
    // 账户快照
}

func (h *Handler) GetPositionSnapshots(w http.ResponseWriter, r *http.Request) {
    // 持仓快照
}

func (h *Handler) GetLiquidations(w http.ResponseWriter, r *http.Request) {
    // 强平记录
}

func (h *Handler) GetTicker(w http.ResponseWriter, r *http.Request) {
    // 行情
}

func (h *Handler) GetStatistics(w http.ResponseWriter, r *http.Request) {
    // 系统统计
}
```

**Step 2: 提交**

```bash
git add backend/internal/handlers/data.go
git commit -m "feat: 实现数据接口"
```

---

## 阶段三：撮合引擎与 WebSocket

### Task 7: 撮合引擎

**Files:**
- Create: `backend/internal/engine/matching.go`
- Create: `backend/internal/engine/position.go`

**Step 1: 创建 matching.go**

```go
package engine

type OrderBook struct {
    // 买单队列（按价格降序）
    Bids []Order
    // 卖单队列（按价格升序）
    Asks []Order
}

type MatchingEngine struct {
    orderBooks map[string]*OrderBook // symbol -> orderbook
    orders     map[string]*Order      // orderID -> order
    positions  map[string]*Position  // strategyID+symbol -> position
}

func NewMatchingEngine() *MatchingEngine {
    return &MatchingEngine{
        orderBooks: make(map[string]*OrderBook),
        orders:     make(map[string]*Order),
        positions:  make(map[string]*Position),
    }
}

func (e *MatchingEngine) AddOrder(order *Order) error {
    // 添加订单到订单簿
}

func (e *MatchingEngine) CancelOrder(orderID string) error {
    // 取消订单
}

func (e *MatchingEngine) Match(symbol string, price, quantity decimal.Decimal) {
    // 撮合逻辑
    // 价格穿过挂单价时成交
}

func (e *MatchingEngine) GetOrders(strategyID string) []*Order {
    // 获取策略的订单
}

func (e *MatchingEngine) GetPosition(strategyID, symbol string) *Position {
    // 获取持仓
}
```

**Step 2: 创建 position.go**

```go
package engine

func (e *MatchingEngine) UpdatePosition(strategyID, symbol, side string, quantity, price decimal.Decimal, isClose bool) error {
    // 更新持仓
    // 计算未实现盈亏
}

func (e *MatchingEngine) CheckLiquidation(strategyID string) (bool, error) {
    // 检查是否需要强平
}

func (e *MatchingEngine) Liquidate(strategyID, symbol string) error {
    // 执行强平
}
```

**Step 3: 提交**

```bash
git add backend/internal/engine/
git commit -m "feat: 实现撮合引擎"
```

---

### Task 8: 币安 WebSocket 客户端

**Files:**
- Create: `backend/internal/websocket/client.go`

**Step 1: 创建 client.go**

```go
package websocket

import (
    "log"
    "github.com/adshao/go-binance/v2"
    "github.com/maker-arena/backend/internal/engine"
)

type BinanceClient struct {
    client     *binance.WSClient
    engine     *engine.MatchingEngine
    symbols    []string
}

func NewBinanceClient(engine *engine.MatchingEngine, symbols []string) *BinanceClient {
    return &BinanceClient{
        engine:  engine,
        symbols: symbols,
    }
}

func (c *BinanceClient) Start() error {
    // 连接币安 WebSocket
    // 订阅 @trade 流
    // 处理消息，调用撮合引擎
}

func (c *BinanceClient) Stop() {
    // 关闭连接
}

func (c *BinanceClient) handleTrade(symbol string, price, quantity decimal.Decimal) {
    // 处理成交数据
    c.engine.Match(symbol, price, quantity)
}
```

**Step 2: 提交**

```bash
git add backend/internal/websocket/
git commit -m "feat: 实现币安 WebSocket 客户端"
```

---

### Task 9: 快照调度器

**Files:**
- Create: `backend/internal/scheduler/snapshot.go`

**Step 1: 创建 snapshot.go**

```go
package scheduler

type SnapshotScheduler struct {
    DB      *gorm.DB
    ticker  *time.Ticker
}

func NewSnapshotScheduler(db *gorm.DB, interval time.Duration) *SnapshotScheduler {
    return &SnapshotScheduler{
        DB:      db,
        ticker:  time.NewTicker(interval),
    }
}

func (s *SnapshotScheduler) Start() {
    // 定时记录快照
}

func (s *SnapshotScheduler) Stop() {
    s.ticker.Stop()
}

func (s *SnapshotScheduler) RecordSnapshots() {
    // 记录账户快照和持仓快照
}
```

**Step 2: 提交**

```bash
git add backend/internal/scheduler/
git commit -m "feat: 实现快照调度器"
```

---

## 阶段四：路由与启动

### Task 10: 路由配置

**Files:**
- Create: `backend/internal/router/router.go`

**Step 1: 创建 router.go**

```go
package router

import (
    "net/http"
    "github.com/gorilla/mux"
    "github.com/maker-arena/backend/internal/handlers"
)

func NewRouter(h *handlers.Handler) *mux.Router {
    r := mux.NewRouter()

    // 策略管理
    r.HandleFunc("/api/v1/strategies", h.ListStrategies).Methods(http.MethodGet)
    r.HandleFunc("/api/v1/strategies", h.CreateStrategy).Methods(http.MethodPost)
    r.HandleFunc("/api/v1/strategies/{id}", h.GetStrategy).Methods(http.MethodGet)
    r.HandleFunc("/api/v1/strategies/{id}", h.UpdateStrategy).Methods(http.MethodPut)
    r.HandleFunc("/api/v1/strategies/{id}", h.DeleteStrategy).Methods(http.MethodDelete)

    // 交易接口
    r.HandleFunc("/api/v1/exchange/createOrder", h.CreateOrder).Methods(http.MethodPost)
    r.HandleFunc("/api/v1/exchange/cancelOrder", h.CancelOrder).Methods(http.MethodPost)
    r.HandleFunc("/api/v1/exchange/getOrders", h.GetOrders).Methods(http.MethodGet)
    r.HandleFunc("/api/v1/exchange/getOrder/{id}", h.GetOrder).Methods(http.MethodGet)
    r.HandleFunc("/api/v1/exchange/getPosition", h.GetPosition).Methods(http.MethodGet)
    r.HandleFunc("/api/v1/exchange/getBalance", h.GetBalance).Methods(http.MethodGet)

    // 数据接口
    r.HandleFunc("/api/v1/fills", h.GetFills).Methods(http.MethodGet)
    r.HandleFunc("/api/v1/snapshots/account", h.GetAccountSnapshots).Methods(http.MethodGet)
    r.HandleFunc("/api/v1/snapshots/position", h.GetPositionSnapshots).Methods(http.MethodGet)
    r.HandleFunc("/api/v1/liquidations", h.GetLiquidations).Methods(http.MethodGet)
    r.HandleFunc("/api/v1/market/ticker", h.GetTicker).Methods(http.MethodGet)
    r.HandleFunc("/api/v1/statistics", h.GetStatistics).Methods(http.MethodGet)

    return r
}
```

**Step 2: 提交**

```bash
git add backend/internal/router/
git commit -m "feat: 配置路由"
```

---

### Task 11: 主程序与启动

**Files:**
- Modify: `backend/cmd/server/main.go`

**Step 1: 创建 main.go**

```go
package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"
    "github.com/maker-arena/backend/internal/config"
    "github.com/maker-arena/backend/internal/database"
    "github.com/maker-arena/backend/internal/engine"
    "github.com/maker-arena/backend/internal/handlers"
    "github.com/maker-arena/backend/internal/router"
    "github.com/maker-arena/backend/internal/scheduler"
    "github.com/maker-arena/backend/internal/websocket"
)

func main() {
    // 加载配置
    cfg, err := config.Load("config.yaml")
    if err != nil {
        log.Fatal(err)
    }

    // 连接数据库
    db, err := database.Connect(cfg.Database)
    if err != nil {
        log.Fatal(err)
    }

    // 运行迁移
    if err := database.Migrate(db); err != nil {
        log.Fatal(err)
    }

    // 初始化撮合引擎
    matchingEngine := engine.NewMatchingEngine()

    // 初始化处理器
    h := handlers.NewHandler(db, matchingEngine, cfg)

    // 创建路由
    r := router.New(h)

    // 启动 WebSocket 客户端
    wsClient := websocket.NewBinanceClient(matchingEngine, cfg.Symbols)
    if err := wsClient.Start(); err != nil {
        log.Printf("WebSocket error: %v", err)
    }

    // 启动快照调度器
    snapshotScheduler := scheduler.NewSnapshotScheduler(db, time.Hour)
    go snapshotScheduler.Start()

    // 优雅关闭
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down...")
    wsClient.Stop()
    snapshotScheduler.Stop()
}
```

**Step 2: 提交**

```bash
git add backend/cmd/server/
git commit -m "feat: 实现主程序"
```

---

## 阶段五：测试与验证

### Task 12: API 测试

**Files:**
- Create: `backend/test/api_test.go`

**Step 1: 创建测试**

```go
package test

import (
    "testing"
    "net/http"
    "net/http/httptest"
)

func TestHealth(t *testing.T) {
    // 测试服务器健康
}

func TestCreateOrder(t *testing.T) {
    // 测试创建订单
}

func TestGetStatistics(t *testing.T) {
    // 测试统计接口
}
```

**Step 2: 运行测试**

```bash
cd backend
go test ./...
```

**Step 3: 提交**

```bash
git add backend/test/
git commit -m "test: 添加 API 测试"
```

---

## 计划完成

后端实现计划已完成，包含 12 个主要任务：
1. 项目初始化
2. 数据库模型
3. 数据库迁移
4. 策略管理 API
5. 交易接口
6. 数据接口
7. 撮合引擎
8. WebSocket 客户端
9. 快照调度器
10. 路由配置
11. 主程序
12. 测试

**Plan complete and saved to `docs/plans/2026-02-25-backend-implementation-plan.md`.**
