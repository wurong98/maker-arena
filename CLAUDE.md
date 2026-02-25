# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Maker Arena 是一个基于币安 WebSocket @trade 数据流的模拟仅挂单交易训练系统，用于策略的准实盘测试。

## Common Commands

```bash
# Run development server
cd backend && go run cmd/server/main.go

# Run tests
cd backend && go test ./...

# Build binary
cd backend && go build -o bin/server cmd/server/main.go
```

## Configuration

- 配置文件: `backend/config.yaml`（复制 `config.example.yaml` 创建）
- 数据库: PostgreSQL 14+，需手动创建数据库 `maker_arena`
- Admin 密码在配置文件中设置

## Architecture

```
┌────────────────┬────────────────────────┐
│   Frontend     │   HTML/JS              │
│   (Port 8080)  │   index.html, strategy.html
└───────┬────────┴────────────────────────┘
        │ HTTP/WebSocket
┌───────▼─────────────────────────────────┐
│   Backend (Go)                         │
│   ├── router/      - HTTP 路由         │
│   ├── handlers/    - API 处理器         │
│   ├── engine/      - 撮合引擎 & 持仓管理│
│   ├── websocket/   - 币安 WebSocket    │
│   ├── scheduler/   - 快照调度器         │
│   ├── database/    - GORM 数据库        │
│   └── models/     - 数据模型           │
└─────────────────────────────────────────┘
```

## Key Components

- **MatchingEngine** (`backend/internal/engine/matching.go`): 订单撮合引擎，价格穿过挂单价时成交
- **PositionManager** (`backend/internal/engine/position.go`): 100倍杠杆持仓管理，含强平机制
- **BinanceClient** (`backend/internal/websocket/client.go`): 订阅币安 WebSocket 行情流

## Trading Rules

- 成交规则：买单价格上穿挂单价成交，卖单价格下穿挂单价成交
- 杠杆：100倍
- 强平价：多单 = 开仓价 × 0.99，空单 = 开仓价 × 1.01
- Maker 费率：0.04%

## API Endpoints

| Category | Path | Description |
|----------|------|-------------|
| Strategy | /api/v1/strategies | 策略 CRUD |
| Trading | /api/v1/exchange/createOrder | 创建订单 |
| Trading | /api/v1/exchange/cancelOrder | 取消订单 |
| Trading | /api/v1/exchange/getOrders | 订单列表 |
| Trading | /api/v1/exchange/getPosition | 持仓 |
| Trading | /api/v1/exchange/getBalance | 余额 |
| Data | /api/v1/fills | 成交记录 |
| Data | /api/v1/snapshots/account | 账户快照 |
| Data | /api/v1/market/ticker | 实时行情 |

## Database

- ORM: GORM with PostgreSQL driver
- Migrations: `backend/internal/database/migration.go`
- Models: `backend/internal/models/`
