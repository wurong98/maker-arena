/**
 * 策略管理逻辑
 * 负责策略管理页面的数据加载和交互
 */

const StrategyApp = {
    // 策略列表
    strategies: [],

    // 当前选中的策略
    selectedStrategy: null,

    /**
     * 初始化应用
     */
    async init() {
        console.log('Initializing Strategy App...');

        // 绑定事件
        this.bindEvents();

        // 加载策略列表
        await this.loadStrategies();

        console.log('Strategy App initialized');
    },

    /**
     * 绑定事件
     */
    bindEvents() {
        // 策略列表点击
        document.getElementById('strategyManagementList').addEventListener('click', (e) => {
            const strategyItem = e.target.closest('.strategy-item');
            if (strategyItem) {
                const strategyId = strategyItem.dataset.strategyId;
                this.selectStrategy(strategyId);
            }
        });
    },

    /**
     * 加载策略列表
     */
    async loadStrategies() {
        try {
            const response = await ApiClient.getStrategies(1, 100);
            this.strategies = response.data || response.strategies || [];

            // 渲染策略列表
            this.renderStrategyList();
        } catch (error) {
            console.error('Failed to load strategies:', error);
            document.getElementById('strategyManagementList').innerHTML =
                '<div class="empty-state">加载失败，请刷新页面</div>';
        }
    },

    /**
     * 渲染策略列表
     */
    renderStrategyList() {
        if (this.strategies.length === 0) {
            document.getElementById('strategyManagementList').innerHTML =
                '<div class="empty-state">暂无策略</div>';
            return;
        }

        const html = this.strategies.map(strategy => {
            const returnClass = strategy.returnRate >= 0 ? 'positive' : 'negative';
            const returnText = Components.formatPercent(strategy.returnRate);
            const isActive = this.selectedStrategy?.id === strategy.id;

            return `
                <div class="strategy-item ${isActive ? 'active' : ''}" data-strategy-id="${strategy.id}">
                    <div class="strategy-info">
                        <div class="strategy-name">${strategy.name || '未命名策略'}</div>
                        <div class="strategy-balance">
                            余额: ${Components.formatNumber(strategy.balance)} USDT
                            <span class="strategy-return ${returnClass}" style="margin-left: 10px;">${returnText}</span>
                        </div>
                    </div>
                </div>
            `;
        }).join('');

        document.getElementById('strategyManagementList').innerHTML = html;
    },

    /**
     * 选择策略
     */
    async selectStrategy(strategyId) {
        const strategy = this.strategies.find(s => s.id === strategyId);
        if (!strategy) return;

        this.selectedStrategy = strategy;

        // 更新列表选中状态
        this.renderStrategyList();

        // 显示详情面板
        document.getElementById('strategyDetailPanel').style.display = 'block';

        // 渲染详情
        this.renderStrategyDetail(strategy);
    },

    /**
     * 渲染策略详情
     */
    renderStrategyDetail(strategy) {
        // 策略名称
        document.getElementById('detailName').textContent = strategy.name || '未命名策略';

        // 状态
        const statusEl = document.getElementById('detailStatus');
        const isActive = strategy.status === 'active' || strategy.isActive;
        statusEl.textContent = isActive ? '运行中' : '已停止';
        statusEl.className = `detail-value ${isActive ? 'active' : 'inactive'}`;

        // 初始资金
        document.getElementById('detailInitialBalance').textContent =
            `${Components.formatNumber(strategy.initialBalance || strategy.initial_balance)} USDT`;

        // 当前余额
        document.getElementById('detailBalance').textContent =
            `${Components.formatNumber(strategy.balance)} USDT`;

        // 收益率
        const returnEl = document.getElementById('detailReturn');
        const returnRate = strategy.returnRate || 0;
        returnEl.textContent = Components.formatPercent(returnRate);
        returnEl.className = `detail-value ${returnRate >= 0 ? 'positive' : 'negative'}`;

        // 创建时间
        document.getElementById('detailCreatedAt').textContent =
            Components.formatTime(strategy.createdAt || strategy.created_at || strategy.createdTime);

        // 总成交数
        document.getElementById('detailTotalTrades').textContent =
            strategy.totalTrades || strategy.total_trades || 0;

        // 总强平次数
        document.getElementById('detailTotalLiquidations').textContent =
            strategy.totalLiquidations || strategy.total_liquidations || 0;
    }
};

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
    StrategyApp.init();
});

// 导出到全局
window.StrategyApp = StrategyApp;
