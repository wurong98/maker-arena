# 挂单成交检查实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 在创建限价单时检查最新成交价格，如果会立刻成交则返回错误

**Architecture:** 在 ExchangeHandler.CreateOrder 方法中添加检查逻辑，获取 ticker 并判断是否满足成交条件

**Tech Stack:** Go, GORM

---

### Task 1: 添加挂单成交检查逻辑

**Files:**
- Modify: `backend/internal/handlers/exchange.go:186-204`

**Step 1: 查看当前 CreateOrder 代码中限价单价格验证的位置**

当前代码在第 184-204 行验证价格，需要在这之后添加成交检查逻辑。

**Step 2: 添加成交检查逻辑**

在价格验证通过后（orderType == "limit" 分支内），添加以下代码：

```go
// 检查挂单是否会导致立即成交
ticker := h.matchingEngine.GetTicker(req.Symbol)
if ticker == nil {
    h.writeError(w, http.StatusBadRequest, "NO_MARKET_DATA", "No market data available for this symbol")
    return
}

// 检查是否会立即成交
if order.Side == "buy" {
    // 买单：价格上穿时成交（当前价格 > 挂单价 且 上一笔价格 <= 挂单价）
    if ticker.Price.GreaterThan(price) && ticker.PreviousPrice.LessThanOrEqual(price) {
        h.writeError(w, http.StatusBadRequest, "ORDER_WOULD_IMMEDIATELY_FILL", "Buy order would immediately fill")
        return
    }
} else if order.Side == "sell" {
    // 卖单：价格下穿时成交（当前价格 < 挂单价 且 上一笔价格 >= 挂单价）
    if ticker.Price.LessThan(price) && ticker.PreviousPrice.GreaterThanOrEqual(price) {
        h.writeError(w, http.StatusBadRequest, "ORDER_WOULD_IMMEDIATELY_FILL", "Sell order would immediately fill")
        return
    }
}
```

**Step 3: 编译检查**

```bash
cd backend && go build ./...
```

预期：无编译错误

**Step 4: 提交**

```bash
git add backend/internal/handlers/exchange.go
git commit -m "$(cat <<'EOF'
feat: 添加挂单成交检查逻辑

挂单时检查最新成交价格，如果会立刻成交则返回错误。
- 买单：当前价格 > 挂单价 且 上一笔价格 <= 挂单价时拒绝
- 卖单：当前价格 < 挂单价 且 上一笔价格 >= 挂单价时拒绝

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
EOF
)"
```

---

### Task 2: 测试验证

**Step 1: 启动后端服务**

```bash
cd backend && go run cmd/server/main.go
```

**Step 2: 手动测试**

使用 curl 或 Postman 测试以下场景：

1. **应该拒绝的场景**（会立即成交）：
   - 买单：当前价格 > 挂单价，且上一笔价格 <= 挂单价
   - 卖单：当前价格 < 挂单价，且上一笔价格 >= 挂单价

2. **应该允许的场景**：
   - 买单：当前价格 <= 挂单价
   - 卖单：当前价格 >= 挂单价

**Step 3: 提交测试结果**
