# 保证金计算修复设计方案

## 问题描述

挂单后可用保证金没有减少，导致可以无限挂单。

## 根因分析

1. **挂单时没有冻结保证金**：创建订单时只检查余额是否足够，但没有实际冻结保证金
2. **可用保证金计算错误**：只计算了持仓保证金，没有计算挂单冻结保证金

## 修复方案

### 1. 数据模型变更

在 Strategy 模型添加 `FrozenMargin` 字段：

```go
type Strategy struct {
    // ... 现有字段
    Balance      decimal.Decimal `gorm:"type:numeric(20,8);default:5000" json:"balance"`
    FrozenMargin decimal.Decimal `gorm:"type:numeric(20,8);default:0" json:"frozenMargin"` // 新增
}
```

### 2. 保证金计算公式

```
持仓保证金 (UsedMargin)   = Σ(持仓价值 / 杠杆)
冻结保证金 (FrozenMargin) = Σ(挂单价格 × 挂单数量 / 杠杆)
可用保证金 (Available)   = 余额 - 持仓保证金 - 冻结保证金
总权益 (TotalEquity)     = 余额 + 未实现盈亏
```

### 3. 业务流程

| 操作 | 冻结保证金变化 |
|------|--------------|
| 创建订单 | `FrozenMargin += price × quantity / 100` |
| 订单成交 | `FrozenMargin -= price × quantity / 100` (释放) + 持仓保证金增加 |
| 订单取消 | `FrozenMargin -= price × quantity / 100` (释放) |
| 订单TTL过期 | `FrozenMargin -= price × quantity / 100` (释放) |

### 4. 强平检查

强平时检查可用余额是否包含冻结保证金：

```go
availableMargin := balance.Sub(usedMargin).Sub(frozenMargin)
if availableMargin.LessThan(decimal.Zero) {
    // 触发强平
}
```

## 需要修改的文件

1. `backend/internal/models/strategy.go` - 添加 FrozenMargin 字段
2. `backend/internal/handlers/exchange.go` - 创建订单时冻结保证金
3. `backend/internal/handlers/exchange.go` - 取消订单时释放保证金
4. `backend/internal/handlers/exchange.go` - GetBalance 返回 FrozenMargin
5. `backend/internal/handlers/exchange.go` - 强平检查包含冻结保证金
6. `backend/internal/engine/matching.go` - 订单成交时释放冻结保证金
