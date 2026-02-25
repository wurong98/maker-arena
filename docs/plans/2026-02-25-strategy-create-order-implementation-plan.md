# 主页策略下单功能实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 在主页(index.html)的每个策略上添加创建限价单功能，支持取消订单，显示可用保证金。

**Architecture:** 前端纯 UI 改动，修改 index.html、components.js、app.js、api.js 四个文件。后端 API 已存在，无需改动。

**Tech Stack:** HTML, JavaScript, localStorage

---

## Task 1: 修改 index.html - 添加可用保证金显示和下单按钮

**Files:**
- Modify: `frontend/index.html:54-68`

**Step 1: 添加可用保证金显示和下单按钮**

在策略信息栏的 `.strategy-stats` 中添加可用保证金显示和下单按钮：

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

**Step 2: 在订单表格添加操作列**

修改订单表格表头，添加"操作"列：

```html
<thead>
    <tr>
        <th>时间</th>
        <th>交易对</th>
        <th>类型</th>
        <th>方向</th>
        <th>价格</th>
        <th>数量</th>
        <th>已成交</th>
        <th>状态</th>
        <th>操作</th>
    </tr>
</thead>
```

**Step 3: 添加模态框容器**

在 `</body>` 前添加两个模态框：

```html
<!-- API Key 设置模态框 -->
<div class="modal" id="apiKeyModal">
    <div class="modal-content">
        <div class="modal-header">
            <h3>设置 API Key</h3>
            <span class="modal-close" id="closeApiKeyModal">&times;</span>
        </div>
        <div class="modal-body">
            <p>请输入策略的 API Key（兼容 ccxt）</p>
            <div class="form-group">
                <label>API Key</label>
                <input type="text" id="apiKeyInput" placeholder="请输入 API Key">
            </div>
            <div class="form-actions">
                <button type="button" class="btn btn-outline" id="cancelApiKey">取消</button>
                <button type="button" class="btn btn-primary" id="saveApiKey">保存</button>
            </div>
        </div>
    </div>
</div>

<!-- 下单模态框 -->
<div class="modal" id="createOrderModal">
    <div class="modal-content">
        <div class="modal-header">
            <h3>创建订单</h3>
            <span class="modal-close" id="closeCreateOrderModal">&times;</span>
        </div>
        <div class="modal-body">
            <form id="createOrderForm">
                <div class="form-group">
                    <label>交易对</label>
                    <select id="orderSymbol" required></select>
                </div>
                <div class="form-group">
                    <label>方向</label>
                    <select id="orderSide" required>
                        <option value="buy">买入</option>
                        <option value="sell">卖出</option>
                    </select>
                </div>
                <div class="form-group">
                    <label>类型</label>
                    <input type="text" value="限价" disabled>
                </div>
                <div class="form-group">
                    <label>价格</label>
                    <input type="number" id="orderPrice" step="0.01" required placeholder="请输入价格">
                </div>
                <div class="form-group">
                    <label>数量</label>
                    <input type="number" id="orderQuantity" step="0.001" required placeholder="请输入数量">
                </div>
                <div class="form-actions">
                    <button type="button" class="btn btn-outline" id="cancelCreateOrder">取消</button>
                    <button type="submit" class="btn btn-primary">创建</button>
                </div>
            </form>
        </div>
    </div>
</div>
```

---

## Task 2: 修改 api.js - 添加 API Key 管理和订单相关方法

**Files:**
- Modify: `frontend/js/api.js`

**Step 1: 在 ApiClient 对象末尾添加新方法**

```javascript
// ===== API Key 管理 =====

/**
 * 获取 API Key
 * @param {string} strategyId - 策略 ID
 */
getApiKey(strategyId) {
    const keys = JSON.parse(localStorage.getItem('strategy_api_keys') || '{}');
    return keys[strategyId] || '';
},

/**
 * 保存 API Key
 * @param {string} strategyId - 策略 ID
 * @param {string} apiKey - API Key
 */
setApiKey(strategyId, apiKey) {
    const keys = JSON.parse(localStorage.getItem('strategy_api_keys') || '{}');
    keys[strategyId] = apiKey;
    localStorage.setItem('strategy_api_keys', JSON.stringify(keys));
},

/**
 * 创建订单
 * @param {string} strategyId - 策略 ID
 * @param {Object} orderData - 订单数据
 */
async createOrder(strategyId, orderData) {
    const apiKey = this.getApiKey(strategyId);
    if (!apiKey) {
        throw new Error('请先设置 API Key');
    }
    return this.post('/exchange/createOrder', orderData, { 'X-API-Key': apiKey });
},

/**
 * 取消订单
 * @param {string} strategyId - 策略 ID
 * @param {string} orderId - 订单 ID
 */
async cancelOrder(strategyId, orderId) {
    const apiKey = this.getApiKey(strategyId);
    if (!apiKey) {
        throw new Error('请先设置 API Key');
    }
    return this.post('/exchange/cancelOrder', { orderId }, { 'X-API-Key': apiKey });
},
```

---

## Task 3: 修改 components.js - 添加渲染方法

**Files:**
- Modify: `frontend/js/components.js`

**Step 1: 添加 API Key 模态框渲染方法**

在 Components 对象中添加：

```javascript
/**
 * 渲染 API Key 设置模态框
 */
renderApiKeyModal() {
    return `
        <div class="modal active" id="apiKeyModal">
            <div class="modal-content">
                <div class="modal-header">
                    <h3>设置 API Key</h3>
                    <span class="modal-close" id="closeApiKeyModal">&times;</span>
                </div>
                <div class="modal-body">
                    <p>请输入策略的 API Key（兼容 ccxt）</p>
                    <div class="form-group">
                        <label>API Key</label>
                        <input type="text" id="apiKeyInput" placeholder="请输入 API Key">
                    </div>
                    <div class="form-actions">
                        <button type="button" class="btn btn-outline" id="cancelApiKey">取消</button>
                        <button type="button" class="btn btn-primary" id="saveApiKey">保存</button>
                    </div>
                </div>
            </div>
        </div>
    `;
},

/**
 * 渲染下单模态框
 */
renderCreateOrderModal(tickerMap = {}) {
    const symbols = Object.keys(tickerMap);
    const symbolOptions = symbols.map(s =>
        `<option value="${s}">${s}</option>`
    ).join('');

    return `
        <div class="modal active" id="createOrderModal">
            <div class="modal-content">
                <div class="modal-header">
                    <h3>创建订单</h3>
                    <span class="modal-close" id="closeCreateOrderModal">&times;</span>
                </div>
                <div class="modal-body">
                    <form id="createOrderForm">
                        <div class="form-group">
                            <label>交易对</label>
                            <select id="orderSymbol" required>${symbolOptions}</select>
                        </div>
                        <div class="form-group">
                            <label>方向</label>
                            <select id="orderSide" required>
                                <option value="buy">买入</option>
                                <option value="sell">卖出</option>
                            </select>
                        </div>
                        <div class="form-group">
                            <label>类型</label>
                            <input type="text" value="限价" disabled>
                        </div>
                        <div class="form-group">
                            <label>价格</label>
                            <input type="number" id="orderPrice" step="0.01" required placeholder="请输入价格">
                        </div>
                        <div class="form-group">
                            <label>数量</label>
                            <input type="number" id="orderQuantity" step="0.001" required placeholder="请输入数量">
                        </div>
                        <div class="form-actions">
                            <button type="button" class="btn btn-outline" id="cancelCreateOrder">取消</button>
                            <button type="submit" class="btn btn-primary">创建</button>
                        </div>
                    </form>
                </div>
            </div>
        </div>
    `;
},

/**
 * 渲染订单表格行（带操作按钮）
 */
renderOrderRowWithAction(order, canCancel = false) {
    const side = order.side || 'buy';
    const sideText = side === 'buy' ? '买入' : '卖出';
    const sideClass = side === 'buy' ? 'long' : 'short';

    const status = order.status || 'pending';
    const canCancelOrder = canCancel && (status === 'pending' || status === 'partially_filled');

    return `
        <tr data-order-id="${order.id}">
            <td>${this.formatTime(order.timestamp || order.time)}</td>
            <td>${order.symbol || '-'}</td>
            <td>${order.type === 'limit' ? '限价' : '市价'}</td>
            <td class="${sideClass}">${sideText}</td>
            <td>${this.formatNumber(order.price)}</td>
            <td>${this.formatNumber(order.quantity)}</td>
            <td>${this.formatNumber(order.filledQuantity || 0)}</td>
            <td>${this.renderOrderStatus(order.status)}</td>
            <td>
                ${canCancelOrder ? `<button class="btn btn-sm btn-outline cancel-order-btn" data-order-id="${order.id}">取消</button>` : '-'}
            </td>
        </tr>
    `;
},

/**
 * 渲染订单表格（带操作）
 */
renderOrdersTableWithAction(orders, canCancel = false) {
    if (!orders || orders.length === 0) {
        return '<tr><td colspan="9" class="empty-state">暂无订单</td></tr>';
    }

    return orders.map(order =>
        this.renderOrderRowWithAction(order, canCancel)
    ).join('');
},
```

**Step 2: 更新 renderOrdersTable 方法支持 canCancel 参数**

将原有的 `renderOrdersTable` 方法签名改为支持 canCancel 参数（向后兼容）：

```javascript
renderOrdersTable(orders, canCancel = false) {
    if (!orders || orders.length === 0) {
        return '<tr><td colspan="8" class="empty-state">暂无订单</td></tr>';
    }

    if (canCancel) {
        return this.renderOrdersTableWithAction(orders, canCancel);
    }

    return orders.map(order => this.renderOrderRow(order)).join('');
},
```

---

## Task 4: 修改 app.js - 添加下单/取消逻辑

**Files:**
- Modify: `frontend/js/app.js`

**Step 1: 添加 API Key 状态和保证金计算**

在 App 对象中添加：

```javascript
// API Key 状态
apiKey: '',

// 保证金率（100倍杠杆）
marginRatio: 0.01,
```

**Step 2: 添加 calculateAvailableBalance 方法**

```javascript
/**
 * 计算可用保证金
 * 可用保证金 = 余额 - 持仓保证金占用
 */
calculateAvailableBalance(balance, positions) {
    if (!positions || positions.length === 0) {
        return balance;
    }

    let usedMargin = 0;
    positions.forEach(pos => {
        const positionValue = parseFloat(pos.quantity) * parseFloat(pos.entryPrice);
        usedMargin += positionValue * this.marginRatio;
    });

    return balance - usedMargin;
},
```

**Step 3: 修改 selectStrategy 方法 - 加载 API Key 并更新 UI**

在 `selectStrategy` 方法末尾添加：

```javascript
// 加载 API Key
this.apiKey = ApiClient.getApiKey(strategy.id);

// 更新下单按钮状态
this.updateOrderButtonState();
```

**Step 4: 添加 updateOrderButtonState 方法**

```javascript
/**
 * 更新下单按钮状态
 */
updateOrderButtonState() {
    const btn = document.getElementById('createOrderBtn');
    if (!btn) return;

    if (!this.apiKey) {
        btn.textContent = '设置 API Key';
        btn.classList.add('btn-outline');
        btn.classList.remove('btn-primary');
    } else {
        btn.textContent = '下单';
        btn.classList.add('btn-primary');
        btn.classList.remove('btn-outline');
    }
},
```

**Step 5: 添加下单相关事件处理方法**

```javascript
/**
 * 绑定下单相关事件
 */
bindOrderEvents() {
    // 下单按钮点击
    const createOrderBtn = document.getElementById('createOrderBtn');
    if (createOrderBtn) {
        createOrderBtn.addEventListener('click', () => this.handleCreateOrderClick());
    }

    // 订单取消按钮（事件委托）
    document.getElementById('ordersTable').addEventListener('click', (e) => {
        if (e.target.classList.contains('cancel-order-btn')) {
            const orderId = e.target.dataset.orderId;
            this.handleCancelOrder(orderId);
        }
    });
},

/**
 * 处理下单按钮点击
 */
handleCreateOrderClick() {
    if (!this.apiKey) {
        this.showApiKeyModal();
    } else {
        this.showCreateOrderModal();
    }
},

/**
 * 显示 API Key 设置模态框
 */
showApiKeyModal() {
    // 移除已存在的模态框
    const existingModal = document.getElementById('apiKeyModal');
    if (existingModal) existingModal.remove();

    // 添加模态框到页面
    const modalHtml = Components.renderApiKeyModal();
    document.body.insertAdjacentHTML('beforeend', modalHtml);

    // 绑定事件
    document.getElementById('closeApiKeyModal').addEventListener('click', () => this.hideApiKeyModal());
    document.getElementById('cancelApiKey').addEventListener('click', () => this.hideApiKeyModal());
    document.getElementById('saveApiKey').addEventListener('click', () => this.saveApiKey());

    // 回显已有 API Key
    const input = document.getElementById('apiKeyInput');
    if (this.apiKey) {
        input.value = this.apiKey;
    }
},

/**
 * 隐藏 API Key 设置模态框
 */
hideApiKeyModal() {
    const modal = document.getElementById('apiKeyModal');
    if (modal) modal.remove();
},

/**
 * 保存 API Key
 */
saveApiKey() {
    const input = document.getElementById('apiKeyInput');
    const apiKey = input.value.trim();

    if (!apiKey) {
        alert('请输入 API Key');
        return;
    }

    // 保存到 localStorage
    ApiClient.setApiKey(this.currentStrategy.id, apiKey);
    this.apiKey = apiKey;

    // 更新按钮状态
    this.updateOrderButtonState();

    // 隐藏模态框
    this.hideApiKeyModal();

    // 显示下单模态框
    this.showCreateOrderModal();
},

/**
 * 显示下单模态框
 */
showCreateOrderModal() {
    // 移除已存在的模态框
    const existingModal = document.getElementById('createOrderModal');
    if (existingModal) existingModal.remove();

    // 添加模态框到页面
    const modalHtml = Components.renderCreateOrderModal(this.tickerMap);
    document.body.insertAdjacentHTML('beforeend', modalHtml);

    // 绑定事件
    document.getElementById('closeCreateOrderModal').addEventListener('click', () => this.hideCreateOrderModal());
    document.getElementById('cancelCreateOrder').addEventListener('click', () => this.hideCreateOrderModal());
    document.getElementById('createOrderForm').addEventListener('submit', (e) => this.handleCreateOrderSubmit(e));
},

/**
 * 隐藏下单模态框
 */
hideCreateOrderModal() {
    const modal = document.getElementById('createOrderModal');
    if (modal) modal.remove();
},

/**
 * 处理创建订单提交
 */
async handleCreateOrderSubmit(e) {
    e.preventDefault();

    const orderData = {
        symbol: document.getElementById('orderSymbol').value,
        side: document.getElementById('orderSide').value,
        type: 'limit',
        quantity: document.getElementById('orderQuantity').value,
        price: document.getElementById('orderPrice').value
    };

    try {
        await ApiClient.createOrder(this.currentStrategy.id, orderData);

        // 隐藏模态框
        this.hideCreateOrderModal();

        // 刷新订单列表
        this.loadOrders();

        // 显示成功提示
        alert('订单创建成功');
    } catch (error) {
        console.error('创建订单失败:', error);
        alert('创建订单失败: ' + (error.message || '未知错误'));
    }
},

/**
 * 处理取消订单
 */
async handleCancelOrder(orderId) {
    if (!confirm('确定要取消该订单吗？')) {
        return;
    }

    try {
        await ApiClient.cancelOrder(this.currentStrategy.id, orderId);

        // 刷新订单列表
        this.loadOrders();

        // 显示成功提示
        alert('订单已取消');
    } catch (error) {
        console.error('取消订单失败:', error);
        alert('取消订单失败: ' + (error.message || '未知错误'));
    }
},
```

**Step 6: 修改 loadPositions 方法 - 计算并显示可用保证金**

```javascript
async loadPositions() {
    try {
        const positions = await ApiClient.getPosition(this.currentStrategy.id);
        const positionList = positions.data || positions.positions || [];

        document.getElementById('positionsTable').innerHTML =
            Components.renderPositionsTable(positionList, this.tickerMap);

        // 计算并显示可用保证金
        const balance = this.currentStrategy.balance || 0;
        const availableBalance = this.calculateAvailableBalance(balance, positionList);
        document.getElementById('strategyAvailableBalance').textContent =
            `${Components.formatNumber(availableBalance)} USDC`;
    } catch (error) {
        console.error('Failed to load positions:', error);
        document.getElementById('positionsTable').innerHTML =
            '<tr><td colspan="7" class="empty-state">加载失败</td></tr>';
    }
},
```

**Step 7: 修改 loadOrders 方法 - 传入 canCancel 参数**

```javascript
async loadOrders() {
    try {
        const { page } = this.pagination.orders;
        const orders = await ApiClient.getOrders(this.currentStrategy.id, page, 20);
        const orderList = orders.data || orders.orders || [];

        const canCancel = !!this.apiKey;
        document.getElementById('ordersTable').innerHTML =
            Components.renderOrdersTable(orderList, canCancel);

        // 渲染分页
        const totalPages = orders.totalPages || Math.ceil((orders.total || 0) / 20);
        this.pagination.orders.totalPages = totalPages;
        document.getElementById('ordersPagination').innerHTML =
            Components.renderPagination(page, totalPages, 'orders');
    } catch (error) {
        console.error('Failed to load orders:', error);
        document.getElementById('ordersTable').innerHTML =
            '<tr><td colspan="9" class="empty-state">加载失败</td></tr>';
    }
},
```

**Step 8: 修改 init 方法 - 调用 bindOrderEvents**

在 `init` 方法的 `this.bindEvents()` 之后添加：

```javascript
// 绑定下单事件
this.bindOrderEvents();
```

---

## Task 5: 验证实现

**Step 1: 检查所有文件是否正确修改**

运行以下命令检查：
```bash
# 检查 index.html 是否包含可用保证金和下单按钮
grep -n "strategyAvailableBalance\|createOrderBtn" frontend/index.html

# 检查 api.js 是否包含 API Key 方法
grep -n "getApiKey\|setApiKey\|createOrder\|cancelOrder" frontend/js/api.js

# 检查 components.js 是否包含新方法
grep -n "renderApiKeyModal\|renderCreateOrderModal\|renderOrderRowWithAction" frontend/js/components.js

# 检查 app.js 是否包含新方法
grep -n "bindOrderEvents\|handleCreateOrderClick\|handleCancelOrder" frontend/js/app.js
```

**Step 2: 提交代码**

```bash
git add frontend/index.html frontend/js/api.js frontend/js/components.js frontend/js/app.js
git commit -m "feat: 添加主页策略下单功能

- 添加可用保证金显示和下单按钮
- 添加 API Key 管理（localStorage）
- 支持创建限价单
- 支持取消订单

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```
