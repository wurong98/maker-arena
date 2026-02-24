# 模拟交易系统前端实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 构建 HTML/CSS/JS 前端页面，展示交易界面和策略管理

**Architecture:**
- 纯原生 HTML + CSS + JavaScript
- 使用 Chart.js (CDN) 绘制图表
- fetch API 调用后端接口

**Tech Stack:** HTML5, CSS3, JavaScript (ES6+), Chart.js

---

## Task 1: 项目结构初始化

**Files:**
- Create: `frontend/index.html` (交易界面)
- Create: `frontend/strategy.html` (策略管理界面)
- Create: `frontend/css/style.css`
- Create: `frontend/js/app.js`
- Create: `frontend/js/api.js`
- Create: `frontend/js/components.js`

**Step 1: 创建目录结构**

```bash
mkdir -p frontend/css frontend/js
```

**Step 2: 提交**

```bash
git add frontend/
git commit -m "feat: 初始化前端项目结构"
```

---

## Task 2: 基础样式与布局

**Files:**
- Modify: `frontend/css/style.css`

**Step 1: 创建样式**

```css
:root {
    --bg-primary: #0d1117;
    --bg-secondary: #161b22;
    --border-color: #30363d;
    --accent: #f7931a;
    --green: #0ecb81;
    --red: #f6465d;
    --text-primary: #e6edf3;
    --text-secondary: #8b949e;
}

* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    background: var(--bg-primary);
    color: var(--text-primary);
    min-width: 1024px;
}

/* Header */
.header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 24px;
    background: var(--bg-secondary);
    border-bottom: 1px solid var(--border-color);
}

/* Layout */
.container {
    display: flex;
    height: calc(100vh - 120px);
}

.sidebar {
    width: 320px;
    background: var(--bg-secondary);
    border-right: 1px solid var(--border-color);
    overflow-y: auto;
}

.content {
    flex: 1;
    display: flex;
    flex-direction: column;
}

/* Footer */
.footer {
    display: flex;
    gap: 24px;
    padding: 8px 24px;
    background: var(--bg-secondary);
    border-top: 1px solid var(--border-color);
    font-size: 12px;
}

/* Tables */
table {
    width: 100%;
    border-collapse: collapse;
}

th, td {
    padding: 12px;
    text-align: left;
    border-bottom: 1px solid var(--border-color);
}

th {
    background: var(--bg-secondary);
    font-weight: 600;
}

/* Colors */
.text-green { color: var(--green); }
.text-red { color: var(--red); }
.text-accent { color: var(--accent); }

/* Buttons */
button {
    padding: 8px 16px;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-size: 14px;
}

.btn-primary {
    background: var(--accent);
    color: white;
}

.btn-secondary {
    background: var(--bg-secondary);
    color: var(--text-primary);
    border: 1px solid var(--border-color);
}
```

**Step 2: 提交**

```bash
git add frontend/css/style.css
git commit -m "feat: 添加基础样式"
```

---

## Task 3: API 客户端

**Files:**
- Modify: `frontend/js/api.js`

**Step 1: 创建 api.js**

```javascript
const API_BASE = 'http://localhost:8080/api/v1';

const API = {
    // 策略
    getStrategies(page = 1, limit = 20) {
        return fetch(`${API_BASE}/strategies?page=${page}&limit=${limit}`)
            .then(res => res.json());
    },

    getStrategy(id) {
        return fetch(`${API_BASE}/strategies/${id}`)
            .then(res => res.json());
    },

    createStrategy(data) {
        return fetch(`${API_BASE}/strategies`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data)
        }).then(res => res.json());
    },

    updateStrategy(id, data) {
        return fetch(`${API_BASE}/strategies/${id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data)
        }).then(res => res.json());
    },

    deleteStrategy(id) {
        return fetch(`${API_BASE}/strategies/${id}`, {
            method: 'DELETE'
        }).then(res => res.json());
    },

    // 交易
    getPosition(strategyId, symbol) {
        const url = `${API_BASE}/exchange/getPosition?strategy_id=${strategyId}${symbol ? '&symbol=' + symbol : ''}`;
        return fetch(url).then(res => res.json());
    },

    getOrders(strategyId, page = 1, limit = 20) {
        return fetch(`${API_BASE}/exchange/getOrders?strategy_id=${strategyId}&page=${page}&limit=${limit}`)
            .then(res => res.json());
    },

    getBalance(strategyId) {
        return fetch(`${API_BASE}/exchange/getBalance?strategy_id=${strategyId}`)
            .then(res => res.json());
    },

    // 数据
    getFills(strategyId, page = 1, limit = 20) {
        return fetch(`${API_BASE}/fills?strategy_id=${strategyId}&page=${page}&limit=${limit}`)
            .then(res => res.json());
    },

    getAccountSnapshots(strategyId, limit = 100) {
        return fetch(`${API_BASE}/snapshots/account?strategy_id=${strategyId}&limit=${limit}`)
            .then(res => res.json());
    },

    getPositionSnapshots(strategyId, limit = 100) {
        return fetch(`${API_BASE}/snapshots/position?strategy_id=${strategyId}&limit=${limit}`)
            .then(res => res.json());
    },

    getLiquidations(strategyId, page = 1, limit = 20) {
        return fetch(`${API_BASE}/liquidations?strategy_id=${strategyId}&page=${page}&limit=${limit}`)
            .then(res => res.json());
    },

    getTicker() {
        return fetch(`${API_BASE}/market/ticker`)
            .then(res => res.json());
    },

    getStatistics() {
        return fetch(`${API_BASE}/statistics`)
            .then(res => res.json());
    }
};
```

**Step 2: 提交**

```bash
git add frontend/js/api.js
git commit -m "feat: 实现 API 客户端"
```

---

## Task 4: 交易界面 - Header 与 Footer

**Files:**
- Modify: `frontend/index.html`

**Step 1: 创建 index.html 基础结构**

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>模拟交易训练系统</title>
    <link rel="stylesheet" href="css/style.css">
    <script src="https://cdn.jsdelivr.net/npm/chart.js@4.4.0/dist/chart.umd.min.js"></script>
</head>
<body>
    <header class="header">
        <div class="logo">Maker Arena</div>
        <div class="stats">
            <span>策略: <span id="totalStrategies">-</span></span>
            <span>成交: <span id="totalFills">-</span></span>
            <span>挂单: <span id="openOrders">-</span></span>
        </div>
    </header>

    <div class="container">
        <aside class="sidebar" id="strategyList">
            <!-- 策略列表 -->
        </aside>

        <main class="content">
            <div class="content-header">
                <h2 id="selectedStrategy">选择策略</h2>
                <div class="strategy-info">
                    <span>余额: <span id="strategyBalance">-</span></span>
                    <span>收益率: <span id="strategyReturn">-</span></span>
                </div>
            </div>

            <div class="tabs">
                <button class="tab active" data-tab="positions">持仓</button>
                <button class="tab" data-tab="fills">成交记录</button>
                <button class="tab" data-tab="orders">订单</button>
                <button class="tab" data-tab="snapshots">快照</button>
                <button class="tab" data-tab="liquidations">强平记录</button>
            </div>

            <div class="tab-content" id="tabContent">
                <!-- Tab 内容 -->
            </div>
        </main>
    </div>

    <footer class="footer" id="tickerFooter">
        <!-- 行情 -->
    </footer>

    <script src="js/api.js"></script>
    <script src="js/components.js"></script>
    <script src="js/app.js"></script>
</body>
</html>
```

**Step 2: 提交**

```bash
git add frontend/index.html
git commit -m "feat: 创建交易界面 HTML"
```

---

## Task 5: 策略列表组件

**Files:**
- Modify: `frontend/js/components.js`

**Step 1: 创建策略列表渲染函数**

```javascript
function renderStrategyList(strategies) {
    const container = document.getElementById('strategyList');
    if (!strategies || strategies.length === 0) {
        container.innerHTML = '<p class="empty">暂无策略</p>';
        return;
    }

    const html = strategies.map((s, i) => {
        const returnPercent = ((parseFloat(s.balance) - 5000) / 5000 * 100).toFixed(2);
        const returnClass = returnPercent >= 0 ? 'text-green' : 'text-red';

        return `
            <div class="strategy-item" data-id="${s.id}">
                <div class="strategy-rank">#${i + 1}</div>
                <div class="strategy-name">${s.name}</div>
                <div class="strategy-balance">${s.balance}</div>
                <div class="strategy-return ${returnClass}">${returnPercent}%</div>
            </div>
        `;
    }).join('');

    container.innerHTML = html;

    // 绑定点击事件
    container.querySelectorAll('.strategy-item').forEach(item => {
        item.addEventListener('click', () => selectStrategy(item.dataset.id));
    });
}
```

**Step 2: 提交**

```bash
git add frontend/js/components.js
git commit -m "feat: 实现策略列表组件"
```

---

## Task 6: 持仓 Tab

**Files:**
- Modify: `frontend/js/components.js`

**Step 1: 创建持仓表格渲染函数**

```javascript
function renderPositions(positions) {
    if (!positions || positions.length === 0) {
        return '<p class="empty">暂无持仓</p>';
    }

    return `
        <table>
            <thead>
                <tr>
                    <th>交易对</th>
                    <th>方向</th>
                    <th>数量</th>
                    <th>开仓价</th>
                    <th>当前价</th>
                    <th>强平价</th>
                    <th>未实现盈亏</th>
                </tr>
            </thead>
            <tbody>
                ${positions.map(p => {
                    const sideClass = p.side === 'long' ? 'text-green' : 'text-red';
                    const pnlClass = parseFloat(p.unrealizedPnl) >= 0 ? 'text-green' : 'text-red';
                    return `
                        <tr>
                            <td>${p.symbol}</td>
                            <td class="${sideClass}">${p.side === 'long' ? '多' : '空'}</td>
                            <td>${p.quantity}</td>
                            <td>${p.entryPrice}</td>
                            <td>${p.currentPrice}</td>
                            <td>${p.liquidationPrice}</td>
                            <td class="${pnlClass}">${p.unrealizedPnl}</td>
                        </tr>
                    `;
                }).join('')}
            </tbody>
        </table>
    `;
}
```

**Step 2: 提交**

```bash
git add frontend/js/components.js
git commit -m "feat: 实现持仓 Tab"
```

---

## Task 7: 成交记录 Tab

**Files:**
- Modify: `frontend/js/components.js`

**Step 1: 创建成交记录渲染函数**

```javascript
function renderFills(fills) {
    if (!fills || fills.length === 0) {
        return '<p class="empty">暂无成交记录</p>';
    }

    return `
        <table>
            <thead>
                <tr>
                    <th>时间</th>
                    <th>交易对</th>
                    <th>方向</th>
                    <th>价格</th>
                    <th>数量</th>
                    <th>手续费</th>
                </tr>
            </thead>
            <tbody>
                ${fills.map(f => {
                    const sideClass = f.side === 'buy' ? 'text-green' : 'text-red';
                    return `
                        <tr>
                            <td>${new Date(f.createdAt).toLocaleString()}</td>
                            <td>${f.symbol}</td>
                            <td class="${sideClass}">${f.side === 'buy' ? '买入' : '卖出'}</td>
                            <td>${f.price}</td>
                            <td>${f.quantity}</td>
                            <td>${f.fee}</td>
                        </tr>
                    `;
                }).join('')}
            </tbody>
        </table>
        ${renderPagination('fills')}
    `;
}
```

**Step 2: 提交**

```bash
git add frontend/js/components.js
git commit -m "feat: 实现成交记录 Tab"
```

---

## Task 8: 订单 Tab

**Files:**
- Modify: `frontend/js/components.js`

**Step 1: 创建订单渲染函数**

```javascript
function renderOrders(orders) {
    if (!orders || orders.length === 0) {
        return '<p class="empty">暂无订单</p>';
    }

    const statusMap = {
        'open中',
        'filled': '已成交',
        'canceled': '已取消',
       ': '挂单 'liquidated': '已强平'
    };

    return `
        <table>
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
                </tr>
            </thead>
            <tbody>
                ${orders.map(o => {
                    const sideClass = o.side === 'buy' ? 'text-green' : 'text-red';
                    return `
                        <tr>
                            <td>${new Date(o.createdAt).toLocaleString()}</td>
                            <td>${o.symbol}</td>
                            <td>${o.type === 'limit' ? '限价' : '市价'}</td>
                            <td class="${sideClass}">${o.side === 'buy' ? '买入' : '卖出'}</td>
                            <td>${o.price}</td>
                            <td>${o.quantity}</td>
                            <td>${o.filledQuantity}</td>
                            <td>${statusMap[o.status] || o.status}</td>
                        </tr>
                    `;
                }).join('')}
            </tbody>
        </table>
    `;
}
```

**Step 2: 提交**

```bash
git add frontend/js/components.js
git commit -m "feat: 实现订单 Tab"
```

---

## Task 9: 快照 Tab (Chart.js)

**Files:**
- Modify: `frontend/js/components.js`

**Step 1: 创建快照渲染函数**

```javascript
let balanceChart = null;
let equityChart = null;

function renderSnapshots(accountSnapshots, positionSnapshots) {
    const hasData = accountSnapshots && accountSnapshots.length > 0;

    if (!hasData) {
        return '<p class="empty">暂无快照数据</p>';
    }

    // 过滤最近 24 小时
    const now = new Date();
    const dayAgo = new Date(now.getTime() - 24 * 60 * 60 * 1000);
    const filtered = accountSnapshots.filter(s => new Date(s.createdAt) >= dayAgo);

    const labels = filtered.map(s => new Date(s.createdAt).toLocaleTimeString());
    const balances = filtered.map(s => s.balance);
    const equities = filtered.map(s => s.totalEquity);

    return `
        <div class="charts">
            <div class="chart-container">
                <h3>余额曲线</h3>
                <canvas id="balanceChart"></canvas>
            </div>
            <div class="chart-container">
                <h3>收益曲线</h3>
                <canvas id="equityChart"></canvas>
            </div>
        </div>
        <script>
            renderCharts('${JSON.stringify(labels)}', '${JSON.stringify(balances)}', '${JSON.stringify(equities)}');
        </script>
    `;
}

function renderCharts(labels, balances, equities) {
    const ctx1 = document.getElementById('balanceChart');
    const ctx2 = document.getElementById('equityChart');

    if (balanceChart) balanceChart.destroy();
    if (equityChart) equityChart.destroy();

    balanceChart = new Chart(ctx1, {
        type: 'line',
        data: {
            labels: JSON.parse(labels),
            datasets: [{
                label: '余额',
                data: JSON.parse(balances),
                borderColor: '#f7931a',
                fill: false
            }]
        }
    });

    equityChart = new Chart(ctx2, {
        type: 'line',
        data: {
            labels: JSON.parse(labels),
            datasets: [{
                label: '总资产',
                data: JSON.parse(equities),
                borderColor: '#0ecb81',
                fill: false
            }]
        }
    });
}
```

**Step 2: 提交**

```bash
git add frontend/js/components.js
git commit -m "feat: 实现快照 Tab (Chart.js)"
```

---

## Task 10: 强平记录 Tab

**Files:**
- Modify: `frontend/js/components.js`

**Step 1: 创建强平记录渲染函数**

```javascript
function renderLiquidations(liquidations) {
    if (!liquidations || liquidations.length === 0) {
        return '<p class="empty">暂无强平记录</p>';
    }

    return `
        <table>
            <thead>
                <tr>
                    <th>时间</th>
                    <th>交易对</th>
                    <th>方向</th>
                    <th>强平价格</th>
                    <th>数量</th>
                </tr>
            </thead>
            <tbody>
                ${liquidations.map(l => {
                    const sideClass = l.side === 'long' ? 'text-green' : 'text-red';
                    return `
                        <tr>
                            <td>${new Date(l.createdAt).toLocaleString()}</td>
                            <td>${l.symbol}</td>
                            <td class="${sideClass}">${l.side === 'long' ? '多' : '空'}</td>
                            <td>${l.liquidationPrice}</td>
                            <td>${l.quantity}</td>
                        </tr>
                    `;
                }).join('')}
            </tbody>
        </table>
    `;
}
```

**Step 2: 提交**

```bash
git add frontend/js/components.js
git commit -m "feat: 实现强平记录 Tab"
```

---

## Task 11: 底部行情栏

**Files:**
- Modify: `frontend/js/components.js`

**Step 1: 创建行情渲染函数**

```javascript
function renderTickers(tickers) {
    if (!tickers || tickers.length === 0) {
        return '';
    }

    return tickers.map(t => {
        const changeClass = parseFloat(t.priceChangePercent24h) >= 0 ? 'text-green' : 'text-red';
        return `
            <span class="ticker">
                <span class="symbol">${t.symbol.toUpperCase()}</span>
                <span class="price">$${t.price}</span>
                <span class="change ${changeClass}">${t.priceChangePercent24h}%</span>
            </span>
        `;
    }).join('');
}
```

**Step 2: 提交**

```bash
git add frontend/js/components.js
git commit -m "feat: 实现底部行情栏"
```

---

## Task 12: 主应用逻辑

**Files:**
- Modify: `frontend/js/app.js`

**Step 1: 创建 app.js**

```javascript
let currentStrategyId = null;
let currentTab = 'positions';

// 初始化
async function init() {
    await loadStatistics();
    await loadStrategies();
    await loadTickers();

    // 5秒轮询行情
    setInterval(loadTickers, 5000);
}

// 加载统计
async function loadStatistics() {
    try {
        const data = await API.getStatistics();
        document.getElementById('totalStrategies').textContent = data.totalStrategies;
        document.getElementById('totalFills').textContent = data.totalFills;
        document.getElementById('openOrders').textContent = data.openOrders;
    } catch (e) {
        console.error('Failed to load statistics:', e);
    }
}

// 加载策略列表
async function loadStrategies() {
    try {
        const data = await API.getStrategies();
        renderStrategyList(data.strategies);
    } catch (e) {
        console.error('Failed to load strategies:', e);
    }
}

// 选择策略
async function selectStrategy(id) {
    currentStrategyId = id;

    // 更新高亮
    document.querySelectorAll('.strategy-item').forEach(item => {
        item.classList.toggle('active', item.dataset.id === id);
    });

    // 加载策略信息
    const strategy = await API.getStrategy(id);
    document.getElementById('selectedStrategy').textContent = strategy.name;

    const returnPercent = ((parseFloat 5000)(strategy.balance) - / 5000 * 100).toFixed(2);
    document.getElementById('strategyBalance').textContent = strategy.balance;
    document.getElementById('strategyReturn').textContent = returnPercent + '%';

    // 加载当前 tab
    await loadTab(currentTab);
}

// 加载 Tab
async function loadTab(tab) {
    currentTab = tab;
    const content = document.getElementById('tabContent');

    if (!currentStrategyId) {
        content.innerHTML = '<p class="empty">请选择策略</p>';
        return;
    }

    try {
        switch (tab) {
            case 'positions':
                const positions = await API.getPosition(currentStrategyId);
                content.innerHTML = renderPositions(positions);
                break;
            case 'fills':
                const fills = await API.getFills(currentStrategyId);
                content.innerHTML = renderFills(fills.fills);
                break;
            case 'orders':
                const orders = await API.getOrders(currentStrategyId);
                content.innerHTML = renderOrders(orders);
                break;
            case 'snapshots':
                const [account, position] = await Promise.all([
                    API.getAccountSnapshots(currentStrategyId),
                    API.getPositionSnapshots(currentStrategyId)
                ]);
                content.innerHTML = renderSnapshots(account.snapshots, position.snapshots);
                break;
            case 'liquidations':
                const liquidations = await API.getLiquidations(currentStrategyId);
                content.innerHTML = renderLiquidations(liquidations.liquidations);
                break;
        }
    } catch (e) {
        content.innerHTML = '<p class="error">加载失败</p>';
    }
}

// 加载行情
async function loadTickers() {
    try {
        const data = await API.getTicker();
        document.getElementById('tickerFooter').innerHTML = renderTickers(data.tickers);
    } catch (e) {
        console.error('Failed to load tickers:', e);
    }
}

// Tab 切换
document.querySelectorAll('.tab').forEach(tab => {
    tab.addEventListener('click', () => {
        document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
        tab.classList.add('active');
        loadTab(tab.dataset.tab);
    });
});

// 启动
init();
```

**Step 2: 提交**

```bash
git add frontend/js/app.js
git commit -m "feat: 实现主应用逻辑"
```

---

## Task 13: 策略管理界面

**Files:**
- Modify: `frontend/strategy.html`

**Step 1: 创建 strategy.html**

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <title>策略管理 - Maker Arena</title>
    <link rel="stylesheet" href="css/style.css">
</head>
<body>
    <header class="header">
        <div class="logo">策略管理</div>
        <button class="btn-primary" onclick="showCreateModal()">创建策略</button>
    </header>

    <main class="content">
        <table>
            <thead>
                <tr>
                    <th>#</th>
                    <th>名称</th>
                    <th>API Key</th>
                    <th>余额</th>
                    <th>收益率</th>
                    <th>状态</th>
                    <th>创建时间</th>
                    <th>操作</th>
                </tr>
            </thead>
            <tbody id="strategyTable">
                <!-- 策略列表 -->
            </tbody>
        </table>
    </main>

    <!-- 模态框 -->
    <div id="modal" class="modal hidden">
        <div class="modal-content">
            <h3 id="modalTitle">创建策略</h3>
            <form id="strategyForm">
                <input type="hidden" id="strategyId">
                <div class="form-group">
                    <label>名称</label>
                    <input type="text" id="name" required>
                </div>
                <div class="form-group">
                    <label>描述</label>
                    <textarea id="description"></textarea>
                </div>
                <div class="form-group">
                    <label>初始资金</label>
                    <input type="number" id="balance" value="5000">
                </div>
                <div class="form-group">
                    <label>API Key (可选)</label>
                    <input type="text" id="apiKey" placeholder="留空自动生成">
                </div>
                <div class="form-actions">
                    <button type="submit" class="btn-primary">保存</button>
                    <button type="button" class="btn-secondary" onclick="closeModal()">取消</button>
                </div>
            </form>
        </div>
    </div>

    <script src="js/api.js"></script>
    <script src="js/strategy.js"></script>
</body>
</html>
```

**Step 2: 创建 strategy.js**

```javascript
async function loadStrategies() {
    const data = await API.getStrategies();
    renderStrategyTable(data.strategies);
}

function renderStrategyTable(strategies) {
    const tbody = document.getElementById('strategyTable');
    tbody.innerHTML = strategies.map((s, i) => {
        const returnPercent = ((parseFloat(s.balance) - 5000) / 5000 * 100).toFixed(2);
        const returnClass = returnPercent >= 0 ? 'text-green' : 'text-red';
        return `
            <tr>
                <td>${i + 1}</td>
                <td>${s.name}</td>
                <td>${s.apiKey.substring(0, 8)}...</td>
                <td>${s.balance}</td>
                <td class="${returnClass}">${returnPercent}%</td>
                <td>${s.enabled ? '启用' : '禁用'}</td>
                <td>${new Date(s.createdAt).toLocaleString()}</td>
                <td>
                    <button class="btn-secondary" onclick="editStrategy('${s.id}')">编辑</button>
                    <button class="btn-secondary" onclick="deleteStrategy('${s.id}')">删除</button>
                </td>
            </tr>
        `;
    }).join('');
}

// CRUD 操作...
```

**Step 3: 提交**

```bash
git add frontend/strategy.html frontend/js/strategy.js
git commit -m "feat: 实现策略管理界面"
```

---

## Task 14: 分页组件

**Files:**
- Modify: `frontend/js/components.js`

**Step 1: 创建分页渲染函数**

```javascript
function renderPagination(type) {
    // 简单的分页组件
    return `
        <div class="pagination">
            <button onclick="changePage('${type}', -1)">上一页</button>
            <span>第 <span id="page-${type}">1</span> 页</span>
            <button onclick="changePage('${type}', 1)">下一页</button>
        </div>
    `;
}

async function changePage(type, delta) {
    const pageEl = document.getElementById(`page-${type}`);
    const currentPage = parseInt(pageEl.textContent);
    const newPage = currentPage + delta;

    if (newPage < 1) return;

    // 重新加载数据
    if (type === 'fills') {
        const data = await API.getFills(currentStrategyId, newPage);
        document.getElementById('tabContent').innerHTML = renderFills(data.fills);
    }
}
```

**Step 2: 提交**

```bash
git add frontend/js/components.js
git commit -m "feat: 添加分页组件"
```

---

## Task 15: 样式优化与响应式

**Files:**
- Modify: `frontend/css/style.css`

**Step 1: 补充样式**

```css
/* 策略列表 */
.strategy-item {
    display: flex;
    padding: 12px 16px;
    border-bottom: 1px solid var(--border-color);
    cursor: pointer;
    transition: background 0.2s;
}

.strategy-item:hover {
    background: rgba(255, 255, 255, 0.05);
}

.strategy-item.active {
    background: rgba(247, 147, 26, 0.1);
    border-left: 3px solid var(--accent);
}

.strategy-rank { width: 40px; color: var(--text-secondary); }
.strategy-name { flex: 1; }
.strategy-balance { text-align: right; }
.strategy-return { width: 60px; text-align: right; }

/* Tabs */
.tabs {
    display: flex;
    border-bottom: 1px solid var(--border-color);
}

.tab {
    padding: 12px 24px;
    background: transparent;
    color: var(--text-secondary);
    border: none;
    border-bottom: 2px solid transparent;
    cursor: pointer;
}

.tab.active {
    color: var(--text-primary);
    border-bottom-color: var(--accent);
}

/* Modal */
.modal {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
}

.modal.hidden { display: none; }

.modal-content {
    background: var(--bg-secondary);
    padding: 24px;
    border-radius: 8px;
    width: 400px;
}

/* Charts */
.charts {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 24px;
    padding: 24px;
}

.chart-container {
    background: var(--bg-secondary);
    padding: 16px;
    border-radius: 8px;
}
```

**Step 2: 提交**

```bash
git add frontend/css/style.css
git commit -m "feat: 完善样式"
```

---

## Task 16: 测试与验证

**Step 1: 测试前端**

- 打开浏览器访问 `http://localhost:8080`
- 检查页面是否正常加载
- 测试策略列表点击
- 测试 Tab 切换

**Step 2: 提交**

```bash
git commit -m "test: 前端完成"
```

---

## 计划完成

前端实现计划已完成，包含 16 个任务：

1. 项目结构初始化
2. 基础样式与布局
3. API 客户端
4. Header 与 Footer
5. 策略列表组件
6. 持仓 Tab
7. 成交记录 Tab
8. 订单 Tab
9. 快照 Tab (Chart.js)
10. 强平记录 Tab
11. 底部行情栏
12. 主应用逻辑
13. 策略管理界面
14. 分页组件
15. 样式优化
16. 测试与验证

**Plan complete and saved to `docs/plans/2026-02-25-frontend-implementation-plan.md`. Two execution options:**

**1. Subagent-Driven (this session)** - I dispatch fresh subagent per task, review between tasks, fast iteration

**2. Parallel Session (separate)** - Open new session with executing-plans, batch execution with checkpoints

**Which approach?**
