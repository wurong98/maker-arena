/**
 * 主应用逻辑
 * 负责页面初始化、事件处理、数据加载等
 */

const App = {
    // 当前选中的策略
    currentStrategy: null,

    // 策略列表
    strategies: [],

    // 行情数据
    tickerMap: {},

    // API Key 状态
    apiKey: '',

    // 保证金率（100倍杠杆）
    marginRatio: 0.01,

    // Chart 实例
    charts: {
        return: null,
        balance: null,
        positionValue: null
    },

    // 分页状态
    pagination: {
        fills: { page: 1, totalPages: 1 },
        orders: { page: 1, totalPages: 1 },
        liquidations: { page: 1, totalPages: 1 }
    },

    // 定时器
    tickerTimer: null,

    /**
     * 初始化应用
     */
    async init() {
        console.log('Initializing App...');

        // 绑定事件
        this.bindEvents();

        // 加载初始数据
        await Promise.all([
            this.loadStrategies(),
            this.loadStatistics(),
            this.loadTickers()
        ]);

        // 启动行情轮询
        this.startTickerPolling();

        console.log('App initialized');
    },

    /**
     * 绑定事件
     */
    bindEvents() {
        // Tab 切换
        document.querySelectorAll('.tab-btn').forEach(btn => {
            btn.addEventListener('click', (e) => this.handleTabSwitch(e));
        });

        // 策略列表点击
        document.getElementById('strategyListContainer').addEventListener('click', (e) => {
            const strategyItem = e.target.closest('.strategy-item');
            if (strategyItem) {
                const strategyId = strategyItem.dataset.strategyId;
                this.selectStrategy(strategyId);
            }
        });

        // 分页点击（事件委托）
        document.addEventListener('click', (e) => {
            if (e.target.classList.contains('pagination button')) {
                const page = parseInt(e.target.dataset.page);
                if (page) {
                    this.handlePageChange(e.target.closest('.pagination').id, page);
                }
            }
        });

        // 绑定下单事件
        this.bindOrderEvents();
    },

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

    /**
     * 处理 Tab 切换
     */
    handleTabSwitch(e) {
        const tabName = e.target.dataset.tab;

        // 更新按钮状态
        document.querySelectorAll('.tab-btn').forEach(btn => btn.classList.remove('active'));
        e.target.classList.add('active');

        // 更新面板显示
        document.querySelectorAll('.tab-panel').forEach(panel => panel.classList.remove('active'));
        document.getElementById(`tab-${tabName}`).classList.add('active');

        // 加载对应数据
        if (this.currentStrategy) {
            this.loadTabData(tabName);
        }
    },

    /**
     * 处理分页变化
     */
    handlePageChange(paginationId, page) {
        const tabName = paginationId.replace('Pagination', '');

        if (tabName === 'fills') {
            this.pagination.fills.page = page;
            this.loadFills();
        } else if (tabName === 'orders') {
            this.pagination.orders.page = page;
            this.loadOrders();
        } else if (tabName === 'liquidations') {
            this.pagination.liquidations.page = page;
            this.loadLiquidations();
        }
    },

    /**
     * 加载策略列表
     */
    async loadStrategies() {
        try {
            const response = await ApiClient.getStrategies(1, 50);
            this.strategies = response.data || response.strategies || [];

            // 按收益率排序
            this.strategies.sort((a, b) => (b.returnRate || 0) - (a.returnRate || 0));

            // 渲染列表
            document.getElementById('strategyListContainer').innerHTML =
                Components.renderStrategyList(this.strategies, this.currentStrategy?.id);

            // 自动选择第一个策略
            if (this.strategies.length > 0 && !this.currentStrategy) {
                this.selectStrategy(this.strategies[0].id);
            }
        } catch (error) {
            console.error('Failed to load strategies:', error);
            document.getElementById('strategyListContainer').innerHTML =
                '<div class="empty-state">加载失败，请刷新页面</div>';
        }
    },

    /**
     * 加载系统统计
     */
    async loadStatistics() {
        try {
            const stats = await ApiClient.getStatistics();
            const formatted = Components.renderStatistics(stats);

            document.getElementById('strategyCount').textContent = formatted.strategyCount;
            document.getElementById('tradeCount').textContent = formatted.tradeCount;
            document.getElementById('orderCount').textContent = formatted.orderCount;
        } catch (error) {
            console.error('Failed to load statistics:', error);
        }
    },

    /**
     * 加载行情
     */
    async loadTickers() {
        try {
            const response = await ApiClient.getMarketTicker();
            const tickerList = response.tickers || [];

            // 更新行情映射
            tickerList.forEach(ticker => {
                this.tickerMap[ticker.symbol] = ticker;
            });

            // 渲染行情
            document.getElementById('tickerContainer').innerHTML =
                Components.renderTickers(tickerList);
        } catch (error) {
            console.error('Failed to load tickers:', error);
            document.getElementById('tickerContainer').innerHTML =
                '<div class="loading">行情加载失败</div>';
        }
    },

    /**
     * 启动行情轮询
     */
    startTickerPolling() {
        // 每 5 秒刷新一次行情
        this.tickerTimer = setInterval(() => {
            this.loadTickers();
        }, 5000);
    },

    /**
     * 选择策略
     */
    async selectStrategy(strategyId) {
        const strategy = this.strategies.find(s => s.id === strategyId);
        if (!strategy) return;

        this.currentStrategy = strategy;

        // 更新策略列表选中状态
        document.querySelectorAll('.strategy-item').forEach(item => {
            item.classList.toggle('active', item.dataset.strategyId === strategyId);
        });

        // 更新顶部策略信息
        document.getElementById('selectedStrategyName').textContent = strategy.name || '未命名策略';
        document.getElementById('strategyBalance').textContent = `${Components.formatNumber(strategy.balance)} USDC`;

        const returnText = Components.formatPercent(strategy.returnRate);
        const returnEl = document.getElementById('strategyReturn');
        returnEl.textContent = returnText;
        returnEl.className = `value ${strategy.returnRate >= 0 ? 'positive' : 'negative'}`;

        // 获取当前激活的 Tab
        const activeTab = document.querySelector('.tab-btn.active').dataset.tab;

        // 加载 Tab 数据
        this.loadTabData(activeTab);

        // 加载 API Key
        this.apiKey = ApiClient.getApiKey(strategy.id);

        // 更新下单按钮状态
        this.updateOrderButtonState();
    },

    /**
     * 加载 Tab 数据
     */
    async loadTabData(tabName) {
        if (!this.currentStrategy) return;

        switch (tabName) {
            case 'positions':
                this.loadPositions();
                break;
            case 'fills':
                this.loadFills();
                break;
            case 'orders':
                this.loadOrders();
                break;
            case 'snapshots':
                this.loadSnapshots();
                break;
            case 'liquidations':
                this.loadLiquidations();
                break;
        }
    },

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

    /**
     * 绑定下单相关事件
     */
    bindOrderEvents() {
        // 下单按钮点击（使用事件委托）
        document.addEventListener('click', (e) => {
            if (e.target && e.target.id === 'createOrderBtn') {
                this.handleCreateOrderClick();
            }
        });

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

        // 刷新订单列表（显示取消按钮）
        this.loadOrders();

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

    /**
     * 加载持仓数据
     */
    async loadPositions() {
        try {
            const positions = await ApiClient.getPosition(this.currentStrategy.id);
            const positionList = positions.data || positions.positions || [];

            document.getElementById('positionsTable').innerHTML =
                Components.renderPositionsTable(positionList, this.tickerMap);

            // 使用后端返回的可用保证金
            const balanceData = await ApiClient.getBalance(this.currentStrategy.id);
            const availableBalance = balanceData.availableMargin || balanceData.available || 0;
            document.getElementById('strategyAvailableBalance').textContent =
                `${Components.formatNumber(availableBalance)} USDC`;
        } catch (error) {
            console.error('Failed to load positions:', error);
            document.getElementById('positionsTable').innerHTML =
                '<tr><td colspan="7" class="empty-state">加载失败</td></tr>';
        }
    },

    /**
     * 加载成交记录
     */
    async loadFills() {
        try {
            const { page } = this.pagination.fills;
            const fills = await ApiClient.getFills(this.currentStrategy.id, page, 20);
            const fillList = fills.data || fills.fills || [];

            document.getElementById('fillsTable').innerHTML =
                Components.renderFillsTable(fillList);

            // 渲染分页
            const totalPages = fills.totalPages || Math.ceil((fills.total || 0) / 20);
            this.pagination.fills.totalPages = totalPages;
            document.getElementById('fillsPagination').innerHTML =
                Components.renderPagination(page, totalPages, 'fills');
        } catch (error) {
            console.error('Failed to load fills:', error);
            document.getElementById('fillsTable').innerHTML =
                '<tr><td colspan="6" class="empty-state">加载失败</td></tr>';
        }
    },

    /**
     * 加载订单
     */
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

    /**
     * 加载快照（图表）
     */
    async loadSnapshots() {
        try {
            // 加载账户快照
            const accountSnapshots = await ApiClient.getAccountSnapshots(this.currentStrategy.id);
            const accountData = accountSnapshots.data || accountSnapshots || [];

            // 加载持仓快照
            const positionSnapshots = await ApiClient.getPositionSnapshots(this.currentStrategy.id);
            const positionData = positionSnapshots.data || positionSnapshots || [];

            // 渲染图表
            this.renderCharts(accountData, positionData);
        } catch (error) {
            console.error('Failed to load snapshots:', error);
        }
    },

    /**
     * 渲染图表
     */
    renderCharts(accountSnapshots, positionSnapshots) {
        const chartOptions = {
            responsive: true,
            maintainAspectRatio: true,
            plugins: {
                legend: {
                    display: false
                }
            },
            scales: {
                x: {
                    grid: {
                        color: '#30363d'
                    },
                    ticks: {
                        color: '#8b949e'
                    }
                },
                y: {
                    grid: {
                        color: '#30363d'
                    },
                    ticks: {
                        color: '#8b949e'
                    }
                }
            }
        };

        // 收益曲线
        this.renderSingleChart('returnChart', 'return', accountSnapshots, chartOptions);

        // 余额曲线
        this.renderSingleChart('balanceChart', 'balance', accountSnapshots, chartOptions);

        // 持仓价值曲线
        this.renderSingleChart('positionValueChart', 'positionValue', positionSnapshots, chartOptions);
    },

    /**
     * 渲染单个图表
     */
    renderSingleChart(canvasId, dataKey, data, options) {
        const ctx = document.getElementById(canvasId);
        if (!ctx) return;

        // 销毁旧图表
        if (this.charts[dataKey]) {
            this.charts[dataKey].destroy();
        }

        const labels = data.map(d => {
            const time = d.timestamp || d.time;
            return time ? Components.formatTimeShort(time) : '';
        });

        const values = data.map(d => d[dataKey] || 0);

        const color = dataKey === 'return' ? '#f7931a' : '#0ecb81';

        this.charts[dataKey] = new Chart(ctx, {
            type: 'line',
            data: {
                labels: labels,
                datasets: [{
                    data: values,
                    borderColor: color,
                    backgroundColor: color + '20',
                    fill: true,
                    tension: 0.4,
                    pointRadius: 0,
                    pointHoverRadius: 4
                }]
            },
            options: options
        });
    },

    /**
     * 加载强平记录
     */
    async loadLiquidations() {
        try {
            const { page } = this.pagination.liquidations;
            const liquidations = await ApiClient.getLiquidations(this.currentStrategy.id, page, 20);
            const liquidationList = liquidations.data || liquidations.liquidations || [];

            document.getElementById('liquidationsTable').innerHTML =
                Components.renderLiquidationsTable(liquidationList);

            // 渲染分页
            const totalPages = liquidations.totalPages || Math.ceil((liquidations.total || 0) / 20);
            this.pagination.liquidations.totalPages = totalPages;
            document.getElementById('liquidationsPagination').innerHTML =
                Components.renderPagination(page, totalPages, 'liquidations');
        } catch (error) {
            console.error('Failed to load liquidations:', error);
            document.getElementById('liquidationsTable').innerHTML =
                '<tr><td colspan="5" class="empty-state">加载失败</td></tr>';
        }
    }
};

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
    App.init();
});

// 导出到全局
window.App = App;
