# 保证金计算修复实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 修复挂单后可用保证金不减少的bug，添加冻结保证金机制

**Architecture:** 在 Strategy 模型添加 FrozenMargin 字段，创建/取消/成交订单时更新冻结保证金，强平检查包含冻结保证金

**Tech Stack:** Go, GORM, PostgreSQL

---

## Task 1: 添加 Strategy.FrozenMargin 字段

**Files:**
- Modify: `backend/internal/models/strategy.go:10-19`

**Step 1: 添加字段**

```go
type Strategy struct {
    ID          string          `gorm:"primaryKey;type:uuid;not null" json:"id"`
    APIKey      string          `gorm:"uniqueIndex;type:varchar(64);not null" json:"apiKey"`
    Name        string          `gorm:"type:varchar(128);not null" json:"name"`
    Description string          `gorm:"type:text" json:"description"`
    Enabled     bool            `gorm:"default:true" json:"enabled"`
    Balance     decimal.Decimal `gorm:"type:numeric(20,8);default:5000" json:"balance"`
    FrozenMargin decimal.Decimal `gorm:"type:numeric(20,8);default:0" json:"frozenMargin"` // 新增
    CreatedAt   time.Time       `json:"createdAt"`
    UpdatedAt   time.Time       `json:"updatedAt"`
}
```

**Step 2: 验证编译**

```bash
cd /home/wurong/workspaces/maker-arena/backend && go build ./...
```

Expected: 编译成功

---

## Task 2: 创建订单时冻结保证金

**Files:**
- Modify: `backend/internal/handlers/exchange.go:216-240`

**Step 1: 添加冻结保证金函数**

在 `ExchangeHandler` 结构体中添加辅助方法：

```go
// freezeMargin 冻结保证金
func (h *ExchangeHandler) freezeMargin(strategyID string, price, quantity decimal.Decimal) error {
    marginRequired := price.Mul(quantity).Div(decimal.NewFromInt(100))
    return h.db.Model(&models.Strategy{}).Where("id = ?", strategyID).
        Update("frozen_margin", gorm.Expr("frozen_margin + ?", marginRequired)).Error
}

// unfreezeMargin 释放冻结保证金
func (h *ExchangeHandler) unfreezeMargin(strategyID string, price, quantity decimal.Decimal) error {
    marginRequired := price.Mul(quantity).Div(decimal.NewFromInt(100))
    return h.db.Model(&models.Strategy{}).Where("id = ?", strategyID).
        Update("frozen_margin", gorm.Expr("frozen_margin - ?", marginRequired)).Error
}
```

**Step 2: 在 CreateOrder 中调用冻结保证金**

在创建订单后（第237行后）添加：

```go
if err := h.db.Create(&order).Error; err != nil {
    h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create order")
    return
}

// 冻结保证金
if err := h.freezeMargin(strategy.ID, price, quantity); err != nil {
    h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to freeze margin")
    return
}
```

**Step 3: 验证编译**

```bash
cd /home/wurong/workspaces/maker-arena/backend && go build ./...
```

Expected: 编译成功

---

## Task 3: 取消订单时释放保证金

**Files:**
- Modify: `backend/internal/handlers/exchange.go:320-335`

**Step 1: 在 CancelOrder 中释放冻结保证金**

在更新订单状态为 canceled 后（第328行后）添加：

```go
// 释放冻结保证金
if err := h.unfreezeMargin(strategy.ID, order.Price, order.Quantity); err != nil {
    fmt.Printf("Failed to unfreeze margin: %v\n", err)
}
```

**Step 2: 验证编译**

```bash
cd /home/wurong/workspaces/maker-arena/backend && go build ./...
```

Expected: 编译成功

---

## Task 4: 成交时释放冻结保证金

**Files:**
- Modify: `backend/internal/engine/matching.go`

**Step 1: 找到订单成交处理位置**

搜索 FillOrder 或 OrderFilled 相关代码，找到成交后更新数据库的位置。

**Step 2: 添加释放冻结保证金逻辑**

在订单成交后，需要：
1. 获取订单的 price 和 quantity
2. 调用 handler 的 unfreezeMargin 释放冻结保证金
3. 增加持仓保证金

注意：matching engine 需要能访问 handler 的 unfreezeMargin 方法，可能需要：
- 将 unfreezeMargin 改为 engine 包内的函数
- 或者通过回调函数传入

**Step 3: 验证编译**

```bash
cd /home/wurong/workspaces/maker-arena/backend && go build ./...
```

Expected: 编译成功

---

## Task 5: TTL订单过期时释放保证金

**Files:**
- Modify: `backend/internal/engine/matching.go`

**Step 1: 找到 TTL 过期处理位置**

搜索 TTL 相关代码，找到订单超时处理的位置。

**Step 2: 添加释放冻结保证金逻辑**

在订单状态更新为 expired/canceled 后，调用 unfreezeMargin 释放冻结保证金。

**Step 3: 验证编译**

```bash
cd /home/wurong/workspaces/maker-arena/backend && go build ./...
```

Expected: 编译成功

---

## Task 6: GetBalance 返回冻结保证金

**Files:**
- Modify: `backend/internal/handlers/exchange.go:494-530`

**Step 1: 在 GetBalanceResponse 中添加 FrozenMargin 字段**

检查 GetBalanceResponse 结构体，添加 FrozenMargin 字段：

```go
type GetBalanceResponse struct {
    Balance         string `json:"balance"`
    FrozenMargin    string `json:"frozenMargin"` // 新增
    UnrealizedPnl   string `json:"unrealizedPnl"`
    TotalEquity     string `json:"totalEquity"`
    UsedMargin      string `json:"usedMargin"`
    AvailableMargin string `json:"availableMargin"`
}
```

**Step 2: 更新 GetBalance 函数**

```go
// Calculate available margin (包含冻结保证金)
availableMargin := strategy.Balance.Sub(usedMargin).Sub(strategy.FrozenMargin)

response := GetBalanceResponse{
    Balance:         strategy.Balance.String(),
    FrozenMargin:    strategy.FrozenMargin.String(),
    UnrealizedPnl:   unrealizedPnl.String(),
    TotalEquity:     totalEquity.String(),
    UsedMargin:      usedMargin.String(),
    AvailableMargin: availableMargin.String(),
}
```

**Step 3: 验证编译**

```bash
cd /home/wurong/workspaces/maker-arena/backend && go build ./...
```

Expected: 编译成功

---

## Task 7: 强平检查包含冻结保证金

**Files:**
- Modify: `backend/internal/handlers/exchange.go:216-221`

**Step 1: 修改保证金检查逻辑**

当前代码：
```go
marginRequired := price.Mul(quantity).Div(decimal.NewFromInt(100))
if marginRequired.GreaterThan(strategy.Balance) {
    h.writeError(w, http.StatusBadRequest, "INSUFFICIENT_BALANCE", "Insufficient balance")
    return
}
```

修改为：
```go
marginRequired := price.Mul(quantity).Div(decimal.NewFromInt(100))
availableMargin := strategy.Balance.Sub(strategy.FrozenMargin)
if marginRequired.GreaterThan(availableMargin) {
    h.writeError(w, http.StatusBadRequest, "INSUFFICIENT_BALANCE", "Insufficient balance")
    return
}
```

**Step 2: 验证编译**

```bash
cd /home/wurong/workspaces/maker-arena/backend && go build ./...
```

Expected: 编译成功

---

## Task 8: 数据库迁移

**Files:**
- Create: `backend/internal/database/migration_add_frozen_margin.go`

**Step 1: 添加迁移脚本**

由于使用了 GORM 的 AutoMigrate，需要确保数据库表包含新字段。运行服务时会自动迁移，或者手动执行：

```sql
ALTER TABLE strategies ADD COLUMN frozen_margin NUMERIC(20,8) DEFAULT 0;
```

**Step 2: 验证**

```bash
cd /home/wurong/workspaces/maker-arena/backend && go run cmd/server/main.go
```

检查启动日志，确认数据库迁移成功。

---

## Task 9: 手动测试验证

**Step 1: 启动服务**

```bash
cd /home/wurong/workspaces/maker-arena/backend && go run cmd/server/main.go
```

**Step 2: 测试场景**

1. 查询初始余额（Balance=5000, FrozenMargin=0, Available=5000）
2. 创建限价挂单（价格=50000, 数量=1，保证金=500）
3. 查询余额（Balance=5000, FrozenMargin=500, Available=4500）
4. 取消订单
5. 查询余额（Balance=5000, FrozenMargin=0, Available=5000）
6. 创建多个挂单超过余额
7. 验证返回 INSUFFICIENT_BALANCE

---

## 执行方式

**Plan complete and saved to `docs/plans/2026-02-25-margin-fix-design.md`. Two execution options:**

**1. Subagent-Driven (this session)** - I dispatch fresh subagent per task, review between tasks, fast iteration

**2. Parallel Session (separate)** - Open new session with executing-plans, batch execution with checkpoints

**Which approach?**
