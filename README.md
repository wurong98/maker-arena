# Maker Arena - 模拟交易训练系统

基于币安 WebSocket @trade 数据流的模拟仅挂单交易训练系统，用于策略的准实盘测试。

## 快速开始

### 前置要求

- Go 1.21+
- PostgreSQL 14+
- 现代浏览器

### 1. 克隆项目

```bash
git clone <repository-url>
cd maker-arena
```

### 2. 创建数据库

```bash
# 假设 PostgreSQL 已安装并运行
# 创建数据库
createdb -U postgres maker_arena

# 或使用 psql
psql -U postgres -c "CREATE DATABASE maker_arena;"
```

### 3. 配置数据库

编辑 `backend/config.yaml`：

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
```

### 4. 启动后端

```bash
cd backend
go run cmd/server/main.go
```

后端将在 http://localhost:8080 启动

### 5. 启动前端

```bash
# 方式1: 直接打开
cd frontend
# 用浏览器打开 index.html

# 方式2: 使用 Python
cd frontend
python3 -m http.server 8081
# 访问 http://localhost:8081
```

## 系统架构

```
┌─────────────────────────────────────────┐
│              前端 (HTML/JS)              │
│   index.html  |  strategy.html           │
└────────────────┬────────────────────────┘
                 │ HTTP API
┌────────────────▼────────────────────────┐
│              后端 (Go)                  │
│  ┌──────────┐  ┌──────────┐            │
│  │  REST    │  │ Matching  │            │
│  │  API     │  │  Engine   │            │
│  └────┬─────┘  └────┬─────┘            │
│       │              │                   │
│  ┌────▼──────────────▼────┐             │
│  │   PostgreSQL           │             │
│  └────────────────────────┘             │
└────────────────┬────────────────────────┘
                  │ WebSocket
┌────────────────▼────────────────────────┐
│           币安 (Binance)                 │
│      btcusdc@trade                      │
└─────────────────────────────────────────┘
```

## API 接口

### 策略管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/strategies | 策略列表 |
| POST | /api/v1/strategies | 创建策略 |
| GET | /api/v1/strategies/:id | 策略详情 |
| PUT | /api/v1/strategies/:id | 更新策略 |
| DELETE | /api/v1/strategies/:id | 删除策略 |

### 交易接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/exchange/createOrder | 创建订单 |
| POST | /api/v1/exchange/cancelOrder | 取消订单 |
| GET | /api/v1/exchange/getOrders | 订单列表 |
| GET | /api/v1/exchange/getPosition | 持仓 |
| GET | /api/v1/exchange/getBalance | 余额 |

### 数据接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/fills | 成交记录 |
| GET | /api/v1/snapshots/account | 账户快照 |
| GET | /api/v1/snapshots/position | 持仓快照 |
| GET | /api/v1/liquidations | 强平记录 |
| GET | /api/v1/market/ticker | 行情 |
| GET | /api/v1/statistics | 系统统计 |

## 核心机制

### 成交逻辑
- 订阅币安 WebSocket `@trade` 流
- 价格严格穿过挂单价才成交：
  - 买单：价格上穿挂单价时成交
  - 卖单：价格下穿挂单价时成交

### 强平机制
- 100 倍杠杆
- 多单强平价 = 开仓价 × (1 - 1/100)
- 空单强平价 = 开仓价 × (1 + 1/100)

### 手续费
- Maker 费率：0.04%

## 目录结构

```
maker-arena/
├── backend/                    # Go 后端
│   ├── cmd/server/main.go      # 入口
│   ├── config.yaml             # 配置
│   └── internal/
│       ├── config/             # 配置加载
│       ├── database/           # 数据库
│       ├── engine/             # 撮合引擎
│       ├── handlers/           # API 处理器
│       ├── router/             # 路由
│       ├── scheduler/         # 调度器
│       ├── websocket/          # WebSocket
│       └── models/             # 数据模型
├── frontend/                  # 前端
│   ├── index.html              # 交易界面
│   ├── strategy.html          # 策略管理
│   ├── css/style.css          # 样式
│   └── js/                    # 脚本
└── docs/plans/                # 设计文档
```

## 开发

### 运行测试

```bash
cd backend
go test ./...
```

### 构建

```bash
cd backend
go build -o bin/server cmd/server/main.go
```

## 许可证

MIT
