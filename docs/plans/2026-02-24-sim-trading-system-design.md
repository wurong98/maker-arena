# 模拟交易训练系统设计文档

## 目录

- [1. 项目概述](#1-项目概述)
- [2. 技术栈](#2-技术栈)
- [3. 核心机制](#3-核心机制)
  - [3.1 合约机制](#31-合约机制)
  - [3.2 成交逻辑](#32-成交逻辑)
  - [3.3 强平价格计算](#33-强平价格计算)
  - [3.4 持仓规则](#34-持仓规则)
- [4. 数据库设计](#4-数据库设计)
  - [4.1 表结构](#41-表结构)
  - [4.2 精度处理](#42-精度处理)
  - [4.3 索引设计](#43-索引设计)
  - [4.4 外键约束](#44-外键约束)
- [5. API 设计](#5-api-设计)
  - [5.1 认证](#51-认证)
  - [5.2 策略管理 API](#52-策略管理-api)
  - [5.3 交易接口（ccxt 兼容）](#53-交易接口ccxt-兼容)
  - [5.4 数据接口](#54-数据接口)
  - [5.5 响应格式](#55-响应格式)
- [6. 系统架构](#6-系统架构)
  - [6.1 组件设计](#61-组件设计)
  - [6.2 启动恢复](#62-启动恢复)
  - [6.3 数据订阅](#63-数据订阅)
  - [6.4 并发控制](#64-并发控制)
  - [6.5 WebSocket 重连机制](#65-websocket-重连机制)
  - [6.6 优雅关闭](#66-优雅关闭)
  - [6.7 数据一致性保证](#67-数据一致性保证)
- [7. 配置文件](#7-配置文件)
  - [7.1 配置说明](#71-配置说明)
  - [7.2 安全建议](#72-安全建议)
- [8. 待确认事项](#8-待确认事项)

## 1. 项目概述

基于币安 WebSocket @trade 数据流的模拟仅挂单交易训练系统，用于策略的准实盘测试。

## 2. 技术栈

- **语言**: Go
- **数据库**: PostgreSQL
- **数据源**: 币安 WebSocket `@trade` 流
- **API 兼容**: ccxt 风格 REST API

## 3. 核心机制

### 3.1 合约机制

- 简化永续合约（单向持仓）
- 100 倍杠杆
- 100% 强平（亏损等于保证金时强平）
- 强平后余额为负数，无法继续开单
- 标记价格 = 最新成交价

### 3.2 成交逻辑

- 订阅币安 WebSocket `btcusdc@trade`（小写）
- 价格严格穿过挂单价才成交：
  - **买单（long）**：价格上穿挂单价时成交（last_price 从低于或等于 order.price 上涨超过 order.price）
  - **卖单（short）**：价格下穿挂单价时成交（last_price 从高于或等于 order.price 下跌低于 order.price）
- 限价单挂单等待成交
- 支持部分成交（订单数量可分多次成交）
- **手续费**：Maker 费率 0.04%（可配置），成交时从余额扣除
- **资金费率**：模拟设为 0（可配置），每 8 小时结算一次

#### 限价单参数

| 参数 | 说明 |
|------|------|
| type | 订单类型：`limit`（限价单，默认）/ `market`（市价单） |
| timeInForce | 有效期：`GTC`（成交为止，默认）/ `IOC`（立即成交或取消）/ `FOK`（全部成交或取消） |
| ttl | 订单过期时间（秒），可选，默认永不过期 |

### 3.3 强平价格计算

**单交易对**：
- 多单：强平价格 = 开仓价 × (1 - 1/杠杆)
- 空单：强平价格 = 开仓价 × (1 + 1/杠杆)

**多交易对全仓**：
- 账户维度检查：总未实现亏损 >= 总保证金时，整个账户强平

### 3.4 持仓规则

- 不同交易对持仓相互独立
- 同一交易对只能持有一个方向（多或空）
- 每个策略初始资金 5000 USDC
- 强平后余额为负数（记录最大亏损），该策略被禁用（enabled=false），无法继续开单

## 4. 数据库设计

### 4.1 表结构

**strategies** - 策略表
| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| api_key | VARCHAR(64) | 唯一索引，API 识别 |
| name | VARCHAR(128) | 策略名称 |
| description | TEXT | 描述 |
| enabled | BOOLEAN | 是否启用 |
| balance | NUMERIC(20,8) | 余额（初始 5000） |
| created_at | TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | 更新时间 |

**orders** - 订单表
| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| strategy_id | UUID | 外键关联 strategies |
| symbol | VARCHAR(32) | 交易对（小写，如 btcusdc） |
| side | VARCHAR(8) | buy/sell |
| type | VARCHAR(8) | limit/market |
| price | NUMERIC(20,8) | 挂单价格 |
| quantity | NUMERIC(20,8) | 挂单数量 |
| filled_quantity | NUMERIC(20,8) | 已成交数量 |
| status | VARCHAR(16) | open/filled/canceled/liquidated |
| time_in_force | VARCHAR(8) | GTC/IOC/FOK |
| ttl | INT | 订单过期秒数，0 表示永不过期 |
| created_at | TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | 更新时间 |

**fills** - 成交记录表
| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| order_id | UUID | 外键关联 orders |
| strategy_id | UUID | 外键关联 strategies |
| symbol | VARCHAR(32) | 交易对 |
| side | VARCHAR(8) | buy/sell |
| price | NUMERIC(20,8) | 成交价格 |
| quantity | NUMERIC(20,8) | 成交数量 |
| fee | NUMERIC(20,8) | 手续费 |
| created_at | TIMESTAMP | 成交时间 |

**positions** - 持仓表
| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| strategy_id | UUID | 外键关联 strategies |
| symbol | VARCHAR(32) | 交易对 |
| side | VARCHAR(8) | long/short |
| quantity | NUMERIC(20,8) | 持仓数量 |
| entry_price | NUMERIC(20,8) | 开仓价格 |
| leverage | INT | 杠杆倍数 |
| created_at | TIMESTAMP | 开仓时间 |
| updated_at | TIMESTAMP | 更新时间 |

**liquidations** - 强平记录表
| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| strategy_id | UUID | 外键关联 strategies |
| symbol | VARCHAR(32) | 交易对 |
| side | VARCHAR(8) | long/short |
| liquidation_price | NUMERIC(20,8) | 强平价格 |
| quantity | NUMERIC(20,8) | 强平数量 |
| created_at | TIMESTAMP | 强平时间 |

**position_snapshots** - 持仓快照表
| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| strategy_id | UUID | 外键关联 strategies |
| symbol | VARCHAR(32) | 交易对 |
| unrealized_pnl | NUMERIC(20,8) | 未实现盈亏 |
| position_value | NUMERIC(20,8) | 持仓价值 |
| avg_price | NUMERIC(20,8) | 平均价格 |
| created_at | TIMESTAMP | 快照时间 |

**account_snapshots** - 账户快照表
| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| strategy_id | UUID | 外键关联 strategies |
| balance | NUMERIC(20,8) | 余额 |
| unrealized_pnl | NUMERIC(20,8) | 未实现盈亏 |
| total_equity | NUMERIC(20,8) | 总资产 |
| created_at | TIMESTAMP | 快照时间 |

### 4.2 精度处理

- 所有金额字段使用 PostgreSQL `NUMERIC(20,8)` 类型
- Go 中使用 `github.com/shopspring/decimal` 包处理
- 价格精度：2-8 位小数（根据交易对）
- 数量精度：0-8 位小数
- 最小下单量：0.001

### 4.3 索引设计

```sql
-- orders 表索引
CREATE INDEX idx_orders_strategy_symbol ON orders(strategy_id, symbol);
CREATE INDEX idx_orders_status ON orders(status);

-- fills 表索引
CREATE INDEX idx_fills_strategy_symbol ON fills(strategy_id, symbol);
CREATE INDEX idx_fills_order_id ON fills(order_id);

-- positions 表索引（每个策略每个交易对只能有一个持仓）
CREATE UNIQUE INDEX idx_positions_strategy_symbol ON positions(strategy_id, symbol);

-- liquidations 表索引
CREATE INDEX idx_liquidations_strategy ON liquidations(strategy_id);

-- snapshots 表索引
CREATE INDEX idx_position_snapshots_strategy_time ON position_snapshots(strategy_id, created_at);
CREATE INDEX idx_account_snapshots_strategy_time ON account_snapshots(strategy_id, created_at);
```

### 4.4 外键约束

为保证数据一致性，建议添加以下外键约束：

```sql
-- orders 表
ALTER TABLE orders ADD CONSTRAINT fk_orders_strategy FOREIGN KEY (strategy_id) REFERENCES strategies(id);

-- fills 表
ALTER TABLE fills ADD CONSTRAINT fk_fills_order FOREIGN KEY (order_id) REFERENCES orders(id);
ALTER TABLE fills ADD CONSTRAINT fk_fills_strategy FOREIGN KEY (strategy_id) REFERENCES strategies(id);

-- positions 表
ALTER TABLE positions ADD CONSTRAINT fk_positions_strategy FOREIGN KEY (strategy_id) REFERENCES strategies(id);

-- liquidations 表
ALTER TABLE liquidations ADD CONSTRAINT fk_liquidations_strategy FOREIGN KEY (strategy_id) REFERENCES strategies(id);

-- snapshots 表
ALTER TABLE position_snapshots ADD CONSTRAINT fk_position_snapshots_strategy FOREIGN KEY (strategy_id) REFERENCES strategies(id);
ALTER TABLE account_snapshots ADD CONSTRAINT fk_account_snapshots_strategy FOREIGN KEY (strategy_id) REFERENCES strategies(id);
```

> **注意**：如对性能有较高要求，可考虑移除外键约束，改用应用层保证数据一致性。

## 5. API 设计

### 5.1 认证

**交易接口**（需要 API Key）：
- 通过 HTTP Header 传递：`X-API-Key: <strategy_api_key>`
- 以下接口需要认证：
  - POST /api/v1/exchange/createOrder
  - POST /api/v1/exchange/cancelOrder

**策略管理接口**（需要密码）：
- 通过 HTTP Header 传递：`X-Admin-Password: <password>`
- 密码明文存储在 config.yaml 中
- 以下接口需要认证：
  - POST /api/v1/strategies
  - PUT /api/v1/strategies/:id
  - DELETE /api/v1/strategies/:id

**其他接口**：公开访问，无需认证

### 5.2 策略管理 API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/strategies | 创建策略 |
| GET | /api/v1/strategies | 列表策略（支持分页） |
| GET | /api/v1/strategies/:id | 策略详情 |
| PUT | /api/v1/strategies/:id | 更新策略 |
| DELETE | /api/v1/strategies/:id | 删除策略（级联删除关联数据） |

#### 5.2.1 策略列表（支持分页）

**GET /api/v1/strategies**

**参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| limit | int | 否 | 每页条数，默认 20，最大 100 |

**响应**:
```json
{
  "strategies": [
    {
      "id": "uuid",
      "apiKey": "abc123...",
      "name": "策略A",
      "description": "测试策略",
      "enabled": true,
      "balance": "5100.00",
      "createdAt": "2024-01-01T00:00:00Z",
      "updatedAt": "2024-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "totalPages": 5
  }
}
```

#### 5.2.2 创建策略

**POST /api/v1/strategies**

**参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 策略名称，最大 128 字符 |
| description | string | 否 | 策略描述，最大 1000 字符 |
| balance | string | 否 | 初始资金，默认 5000 |
| api_key | string | 否 | 自定义 API Key，默认自动生成 UUID |

### 5.3 交易接口（ccxt 兼容）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/exchange/createOrder | 创建订单 |
| POST | /api/v1/exchange/cancelOrder | 取消订单 |
| GET | /api/v1/exchange/getOrders | 查询订单列表 |
| GET | /api/v1/exchange/getOrder/:id | 订单详情 |
| GET | /api/v1/exchange/getPosition | 查询持仓 |
| GET | /api/v1/exchange/getBalance | 查询余额 |

#### 5.3.1 订单操作

**创建订单 POST /api/v1/exchange/createOrder**

**认证**: `X-API-Key: <strategy_api_key>`

**参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| symbol | string | 是 | 交易对 (btcusdc) |
| side | string | 是 | buy/sell |
| type | string | 否 | limit/market，默认 limit |
| quantity | string | 是 | 数量 |
| price | string | 否 | 价格，type=limit 时必填 |
| timeInForce | string | 否 | GTC/IOC/FOK，默认 GTC |
| ttl | int | 否 | 订单过期秒数，默认永不过期 |

**响应**:
```json
{
  "id": "uuid",
  "symbol": "btcusdc",
  "side": "buy",
  "type": "limit",
  "price": "50000.00",
  "quantity": "0.01",
  "filledQuantity": "0",
  "status": "open",
  "createdAt": "2024-01-01T00:00:00Z"
}
```

**取消订单 POST /api/v1/exchange/cancelOrder**

**认证**: `X-API-Key: <strategy_api_key>`

**参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| order_id | string | 是 | 订单 ID |

**响应**:
```json
{
  "id": "uuid",
  "symbol": "btcusdc",
  "side": "buy",
  "price": "50000.00",
  "quantity": "0.01",
  "filledQuantity": "0",
  "status": "canceled",
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:01:00Z"
}
```

#### 5.3.2 查询接口

**GET /api/v1/exchange/getOrders**

**参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| strategy_id | string | 是 | 策略 ID |
| page | int | 否 | 页码，默认 1 |
| limit | int | 否 | 每页条数，默认 20，最大 100 |

**GET /api/v1/exchange/getPosition**

**参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| strategy_id | string | 是 | 策略 ID |
| symbol | string | 否 | 交易对，不传则返回所有持仓 |

**GET /api/v1/exchange/getBalance**

**参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| strategy_id | string | 是 | 策略 ID |

**GET /api/v1/exchange/getOrder/:id**

**响应**: 返回单个订单详情

### 5.4 数据接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/fills | 成交记录 |
| GET | /api/v1/snapshots/account | 账户收益快照 |
| GET | /api/v1/snapshots/position | 持仓快照 |
| GET | /api/v1/liquidations | 强平记录 |
| GET | /api/v1/market/ticker | 最新行情 |
| GET | /api/v1/statistics | 系统统计 |

#### 5.4.1 成交记录

**GET /api/v1/fills**

**参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| strategy_id | string | 否 | 策略 ID，不传则返回所有 |
| page | int | 否 | 页码，默认 1 |
| limit | int | 否 | 每页条数，默认 20，最大 100 |

**响应**:
```json
{
  "fills": [
    {
      "id": "uuid",
      "orderId": "uuid",
      "strategyId": "uuid",
      "symbol": "btcusdc",
      "side": "buy",
      "price": "50000.00",
      "quantity": "0.01",
      "fee": "0.20",
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "totalPages": 5
  }
}
```

#### 5.4.2 账户收益快照

**GET /api/v1/snapshots/account**

**参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| strategy_id | string | 是 | 策略 ID |
| start_time | string | 否 | 开始时间 (RFC3339) |
| end_time | string | 否 | 结束时间 (RFC3339) |
| limit | int | 否 | 返回条数，默认 100 |

**响应**:
```json
{
  "snapshots": [
    {
      "id": "uuid",
      "strategyId": "uuid",
      "balance": "5010.00",
      "unrealizedPnl": "10.00",
      "totalEquity": "5020.00",
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

#### 5.4.3 持仓快照

**GET /api/v1/snapshots/position**

**参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| strategy_id | string | 是 | 策略 ID |
| symbol | string | 否 | 交易对，不传则返回所有持仓 |
| start_time | string | 否 | 开始时间 |
| end_time | string | 否 | 结束时间 |
| limit | int | 否 | 返回条数，默认 100 |

**响应**:
```json
{
  "snapshots": [
    {
      "id": "uuid",
      "strategyId": "uuid",
      "symbol": "btcusdc",
      "unrealizedPnl": "10.00",
      "positionValue": "500.00",
      "avgPrice": "50000.00",
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

#### 5.4.4 强平记录

**GET /api/v1/liquidations**

**参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| strategy_id | string | 否 | 策略 ID，不传则返回所有 |
| page | int | 否 | 页码，默认 1 |
| limit | int | 否 | 每页条数，默认 20，最大 100 |

**响应**:
```json
{
  "liquidations": [
    {
      "id": "uuid",
      "strategyId": "uuid",
      "strategyName": "策略名称",
      "symbol": "btcusdc",
      "side": "long",
      "liquidationPrice": "49500.00",
      "quantity": "0.01",
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "totalPages": 5
  }
}
```

#### 5.4.5 行情接口

**GET /api/v1/market/ticker**

**响应**:
```json
{
  "tickers": [
    {
      "symbol": "btcusdc",
      "price": "67234.50",
      "priceChange24h": "1234.50",
      "priceChangePercent24h": "1.87",
      "updatedAt": "2026-02-25T14:30:25Z"
    }
  ]
}
```

#### 5.4.6 系统统计

**GET /api/v1/statistics**

**响应**:
```json
{
  "totalStrategies": 10,
  "totalFills": 1500,
  "openOrders": 25
}
```

### 5.5 响应格式

**getPosition**
```json
{
  "symbol": "btcusdc",
  "side": "long",
  "quantity": "0.01",
  "entryPrice": "50000.00",
  "currentPrice": "50100.00",
  "leverage": 100,
  "positionValue": "500.00",
  "liquidationPrice": "49500.00",
  "unrealizedPnl": "10.00"
}
```

**getBalance**
```json
{
  "balance": "5010.00",
  "unrealizedPnl": "10.00",
  "totalEquity": "5020.00",
  "usedMargin": "5.00",
  "availableMargin": "5015.00"
}
```

**错误响应**
```json
{
  "error": {
    "code": "INSUFFICIENT_BALANCE",
    "message": "余额不足"
  }
}
```

**通用错误码**

| 错误码 | 说明 |
|--------|------|
| INSUFFICIENT_BALANCE | 余额不足 |
| INVALID_ORDER | 无效订单参数 |
| ORDER_NOT_FOUND | 订单不存在 |
| POSITION_NOT_FOUND | 持仓不存在 |
| STRATEGY_DISABLED | 策略已禁用 |
| UNAUTHORIZED | 未授权 |
| INVALID_API_KEY | API Key 无效 |
| SYMBOL_NOT_FOUND | 交易对不存在 |

## 6. 系统架构

### 6.1 组件设计

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           模拟交易训练系统                                │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌──────────────┐     ┌──────────────┐     ┌──────────────┐          │
│  │ HTTP Server  │────▶│  Order Engine │◀────│   Position    │          │
│  │  (REST API)  │     │  (撮合引擎)    │     │   Manager    │          │
│  └──────┬───────┘     └──────┬───────┘     └──────┬───────┘          │
│         │                    │                    │                    │
│         │                    ▼                    │                    │
│         │            ┌──────────────┐            │                    │
│         │            │  PostgreSQL   │            │                    │
│         │            │    (持久化)    │            │                    │
│         │            └──────────────┘            │                    │
│         │                                       │                    │
│         │                    ┌──────────────────┘                    │
│         │                    ▼                                         │
│  ┌──────┴───────┐     ┌──────────────┐                              │
│  │   Strategy   │◀────│   Snapshot   │                              │
│  │   Manager    │     │  Scheduler    │                              │
│  └──────────────┘     └──────────────┘                              │
│                              │                                         │
├──────────────────────────────┼────────────────────────────────────────┤
│                              ▼                                         │
│                    ┌──────────────────┐                               │
│                    │ WebSocket Manager │◀── btcusdc@trade            │
│                    │   (币安数据流)     │◀── ethusdc@trade            │
│                    └──────────────────┘                               │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

**组件说明**：

| 组件 | 职责 |
|------|------|
| HTTP Server | 提供 REST API 接口，处理订单创建/取消等请求 |
| Order Engine | 内存撮合引擎，维护活跃订单，处理成交逻辑 |
| Position Manager | 持仓管理，计算未实现盈亏，强平检查 |
| WebSocket Manager | 管理币安 WebSocket 连接，接收实时成交数据 |
| Snapshot Scheduler | 定时记录账户/持仓快照 |
| Strategy Manager | 策略管理（创建、更新、禁用） |
| PostgreSQL | 数据持久化存储 |

1. **WebSocket Manager**: 管理币安 WebSocket 连接，订阅所有配置交易对的 @trade 流
2. **Order Engine**: 内存撮合引擎，维护所有活跃订单，按价格排序
3. **Position Manager**: 持仓管理，计算未实现盈亏，强平检查
4. **Snapshot Scheduler**: 定时任务，定时记录快照（默认 1h，可配置）
5. **HTTP Server**: REST API 服务

### 6.2 启动恢复

- 服务启动时从数据库加载所有 status='open' 的订单
- 加载到内存中继续处理

### 6.3 数据订阅

- 一个 WebSocket 连接订阅所有交易对的 @trade 流
- 通过消息中的 symbol 字段区分

### 6.4 并发控制

- 使用 Go channel 实现订单处理队列，保证订单处理顺序
- 订单成交使用数据库事务保证原子性：
  ```go
  tx := db.Begin()
  // 1. 更新订单状态为 filled
  // 2. 插入成交记录
  // 3. 更新或创建持仓
  // 4. 扣除手续费
  tx.Commit()
  ```
- 使用 Redis 分布式锁或数据库 `SELECT FOR UPDATE` 防止并发冲突（可选）

### 6.5 WebSocket 重连机制

- 监听 WebSocket 连接状态，断开后自动重连
- 重连策略：指数退避（1s, 2s, 4s, 8s...，最大 30s）
- 重连成功后重新订阅所有交易对
- 记录断连日志，便于排查问题

### 6.6 优雅关闭

- 监听系统信号（SIGINT, SIGTERM）
- 收到信号后：
  1. 停止接收新订单
  2. 等待当前处理中的订单完成
  3. 保存所有内存状态到数据库
  4. 关闭 WebSocket 连接
  5. 关闭 HTTP 服务器

### 6.7 数据一致性保证

- 定期（每 5 分钟）从数据库加载状态做校验
- 内存状态与数据库状态不一致时，以数据库为准
- 关键操作（订单成交、强平）使用数据库事务

## 7. 配置文件

config.yaml：
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
  password: "your-admin-password-here"

binance:
  ws_url: "wss://stream.binance.com:9443/ws"

symbols:
  - "btcusdc"
  - "ethusdc"

snapshot:
  interval: "1h"

trading:
  maker_fee_rate: "0.0004"  # Maker 手续费率
  funding_rate: "0"         # 资金费率
  funding_interval: "8h"   # 资金结算间隔
```

### 7.1 配置说明

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| database.password | 数据库密码 | - |
| admin.password | 管理密码 | - |
| trading.maker_fee_rate | Maker 手续费率 | 0.0004 (0.04%) |
| trading.funding_rate | 资金费率 | 0 |
| trading.funding_interval | 资金结算间隔 | 8h |

### 7.2 安全建议

- **生产环境**：密码应使用环境变量或 secrets 管理工具（如 Vault）存储
- **开发环境**：可使用 config.yaml 明文配置
- 示例：
  ```yaml
  database:
    password: "${DB_PASSWORD}"  # 从环境变量读取
  admin:
    password: "${ADMIN_PASSWORD}"
  ```

## 8. 待确认事项

- [x] 强平记录持久化（已包含）
- [x] 订单状态枚举：open/filled/canceled/liquidated
- [x] 手续费机制（已添加）
- [x] 资金费率机制（已添加，暂设为 0）
- [x] 并发控制设计（已添加）
- [x] WebSocket 重连机制（已添加）
- [x] 优雅关闭机制（已添加）
- [x] 成交记录手续费字段（已添加 fee 字段）
- [x] 快照接口拆分（已拆分为 account 和 position）
- [x] 系统统计 API（已添加 /api/v1/statistics）
- [x] 策略列表分页（已支持）
- [x] 策略级联删除（已确认）
