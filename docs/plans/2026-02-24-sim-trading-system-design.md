# 模拟交易训练系统设计文档

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
  - 买单：trade.price > order.price 且当前价格 <= order.price 时不成交，trade.price 继续上涨超过 order.price 时成交
  - 卖单：trade.price < order.price 且当前价格 >= order.price 时不成交，trade.price 继续下跌低于 order.price 时成交
- 限价单挂单等待成交

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
| price | NUMERIC(20,8) | 挂单价格 |
| quantity | NUMERIC(20,8) | 挂单数量 |
| filled_quantity | NUMERIC(20,8) | 已成交数量 |
| status | VARCHAR(16) | open/filled/canceled/liquidated |
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
| GET | /api/v1/strategies | 列表策略 |
| GET | /api/v1/strategies/:id | 策略详情 |
| PUT | /api/v1/strategies/:id | 更新策略 |
| DELETE | /api/v1/strategies/:id | 删除策略 |

### 5.3 交易接口（ccxt 兼容）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/exchange/createOrder | 创建订单 |
| POST | /api/v1/exchange/cancelOrder | 取消订单 |
| GET | /api/v1/exchange/getOrders | 查询订单列表 |
| GET | /api/v1/exchange/getOrder/:id | 订单详情 |
| GET | /api/v1/exchange/getPosition | 查询持仓 |
| GET | /api/v1/exchange/getBalance | 查询余额 |

### 5.4 数据接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/fills | 成交记录 |
| GET | /api/v1/snapshots | 收益曲线快照 |

### 5.5 响应格式

**createOrder**
```json
{
  "id": "uuid",
  "symbol": "btcusdc",
  "side": "buy",
  "price": "50000.00",
  "quantity": "0.01",
  "filledQuantity": "0",
  "status": "open",
  "createdAt": "2024-01-01T00:00:00Z"
}
```

**getPosition**
```json
{
  "symbol": "btcusdc",
  "side": "long",
  "quantity": "0.01",
  "entryPrice": "50000.00",
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

## 6. 系统架构

### 6.1 组件设计

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
```

## 8. 待确认事项

- [ ] 强平记录是否需要持久化（已包含）
- [ ] 订单状态枚举确认：open/filled/canceled/liquidated
