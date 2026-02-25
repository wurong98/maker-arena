# 挂单成交检查设计方案

## 需求

挂单需要检查最新成交价格，如果会立刻成交则返回错误，不允许挂单。

## 成交规则回顾

- 买单（buy）：价格上穿挂单价时成交（从 <= 挂单价变为 > 挂单价）
- 卖单（sell）：价格下穿挂单价时成交（从 >= 挂单价变为 < 挂单价）

## 检查逻辑

在创建限价单时，获取当前 ticker，检查是否满足成交条件：

```go
// 获取 ticker
ticker := matchingEngine.GetTicker(symbol)
if ticker == nil {
    return error("No market data available")
}

// 买单检查：如果当前价格 > 挂单价 且 上一笔价格 <= 挂单价，则会立即成交
if side == "buy" && ticker.Price.GreaterThan(price) && ticker.PreviousPrice.LessThanOrEqual(price) {
    return error("Order would immediately fill")
}

// 卖单检查：如果当前价格 < 挂单价 且 上一笔价格 >= 挂单价，则会立即成交
if side == "sell" && ticker.Price.LessThan(price) && ticker.PreviousPrice.GreaterThanOrEqual(price) {
    return error("Order would immediately fill")
}
```

## 边界情况处理

| 情况 | 处理 |
|------|------|
| ticker 不存在（无行情数据） | 拒绝挂单，返回错误 |
| PreviousPrice 为 0（首次行情） | 使用当前价格进行比较 |

## 市价单处理

市价单（market order）不允许挂单，在订单类型验证时已拒绝。

## 需要修改的文件

`backend/internal/handlers/exchange.go` - 在 CreateOrder 方法中添加检查逻辑

## 错误码

- `ORDER_WOULD_IMMEDIATELY_FILL` - 订单会立即成交
- `NO_MARKET_DATA` - 无市场数据
