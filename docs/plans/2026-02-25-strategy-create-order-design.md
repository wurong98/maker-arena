# 主页策略下单功能设计

**日期**: 2026-02-25

## 需求概述

在主页(index.html)的每个策略上添加创建限价单功能，支持取消订单，显示可用保证金。

## 需求详情

1. **认证方式**: 用户填写 API Key（兼容 ccxt），存储在 localStorage，无 API Key 不可创建/取消订单
2. **下单入口**: 选中策略后，在策略信息栏显示"下单"按钮
3. **功能**:
   - 创建限价单
   - 支持取消订单
   - 显示可用保证金

## 设计方案

### 1. API Key 存储

- 使用 localStorage 存储策略 API Key
- 存储键名: `strategy_api_keys`
- 数据结构: `{ "策略ID": "API Key", ... }`

### 2. 前端 UI 变更

#### index.html

在策略信息栏添加：
- "可用保证金"显示
- "下单"按钮

```html
<div class="strategy-stats">
    <div class="strategy-stat">
        <span class="label">余额</span>
        <span class="value" id="strategyBalance">-</span>
    </div>
    <div class="strategy-stat">
        <span class="label">可用保证金</span>
        <span class="value" id="strategyAvailableBalance">-</span>
    </div>
    <div class="strategy-stat">
        <span class="label">收益率</span>
        <span class="value" id="strategyReturn">-</span>
    </div>
    <button class="btn btn-primary" id="createOrderBtn">下单</button>
</div>
```

#### components.js

新增方法：
- `renderOrderRowWithAction(order, canCancel)` - 订单行含取消按钮
- `renderCreateOrderModal(tickerMap)` - 下单表单模态框
- `renderApiKeyModal()` - API Key 设置模态框

### 3. 业务逻辑

#### app.js

1. **API Key 管理**:
   - 从 localStorage 读取 API Key
   - 检查当前策略是否有 API Key
   - 无 API Key 时禁用下单/取消按钮

2. **可用保证金计算**:
   - 加载持仓时计算保证金占用
   - 可用保证金 = 余额 - 持仓保证金

3. **下单流程**:
   - 点击"下单"按钮 → 检查 API Key → 弹出下单表单 → 提交 → 调用 API → 刷新订单列表

4. **取消订单流程**:
   - 点击"取消"按钮 → 确认 → 调用 API → 刷新订单列表

### 4. API 接口

#### api.js 新增方法

```javascript
// 获取 API Key
getApiKey(strategyId) {
    const keys = JSON.parse(localStorage.getItem('strategy_api_keys') || '{}');
    return keys[strategyId];
}

// 保存 API Key
setApiKey(strategyId, apiKey) {
    const keys = JSON.parse(localStorage.getItem('strategy_api_keys') || '{}');
    keys[strategyId] = apiKey;
    localStorage.setItem('strategy_api_keys', JSON.stringify(keys));
}

// 创建订单
async createOrder(strategyId, orderData) {
    const apiKey = this.getApiKey(strategyId);
    return this.post('/exchange/createOrder', orderData, { 'X-API-Key': apiKey });
}

// 取消订单
async cancelOrder(strategyId, orderId) {
    const apiKey = this.getApiKey(strategyId);
    return this.post('/exchange/cancelOrder', { orderId }, { 'X-API-Key': apiKey });
}
```

### 5. 模态框设计

#### API Key 设置模态框

首次点击下单时，若无 API Key，弹出此框：
- 输入框：API Key
- 确认/取消按钮

#### 下单表单模态框

- 交易对：下拉选择（从行情列表获取）
- 方向：买入/卖出 选择
- 类型：限价（固定）
- 价格：输入框
- 数量：输入框
- 提交/取消按钮

### 6. 后端接口

无需改动。后端 API `POST /exchange/createOrder` 和 `POST /exchange/cancelOrder` 已支持 `X-API-Key` 认证。

## 文件变更清单

1. `frontend/index.html` - 添加可用保证金显示和下单按钮
2. `frontend/js/components.js` - 新增渲染方法
3. `frontend/js/app.js` - 添加下单/取消逻辑
4. `frontend/js/api.js` - 添加 API 方法

## 验收标准

1. 选中策略后显示"可用保证金"
2. 无 API Key 时，下单/取消按钮禁用或提示设置
3. 可以创建限价单（限价单不会立即成交）
4. 可以在订单 Tab 取消订单
5. API Key 存储在 localStorage，刷新页面后保留
