/**
 * UI 组件
 * 负责渲染各种 UI 组件
 */

const Components = {
    /**
     * 格式化数字
     */
    formatNumber(value, decimals = 2) {
        if (value === null || value === undefined || isNaN(value)) {
            return '-';
        }
        return Number(value).toFixed(decimals);
    },

    /**
     * 格式化百分比
     */
    formatPercent(value, decimals = 2) {
        if (value === null || value === undefined || isNaN(value)) {
            return '-';
        }
        const sign = value >= 0 ? '+' : '';
        return `${sign}${Number(value).toFixed(decimals)}%`;
    },

    /**
     * 格式化时间
     */
    formatTime(timestamp) {
        if (!timestamp) return '-';
        const date = new Date(timestamp);
        return date.toLocaleString('zh-CN', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit'
        });
    },

    /**
     * 格式化时间（简短）
     */
    formatTimeShort(timestamp) {
        if (!timestamp) return '-';
        const date = new Date(timestamp);
        return date.toLocaleTimeString('zh-CN', {
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit'
        });
    },

    /**
     * 渲染策略列表项
     */
    renderStrategyItem(strategy, rank, isActive = false) {
        const returnClass = strategy.returnRate >= 0 ? 'positive' : 'negative';
        const returnText = this.formatPercent(strategy.returnRate);
        const rankClass = rank <= 3 ? `top-${rank}` : '';

        return `
            <div class="strategy-item ${isActive ? 'active' : ''}" data-strategy-id="${strategy.id}">
                <div class="strategy-rank ${rankClass}">#${rank}</div>
                <div class="strategy-info">
                    <div class="strategy-name">${strategy.name || '未命名策略'}</div>
                    <div class="strategy-balance">余额: ${this.formatNumber(strategy.balance)} USDC</div>
                </div>
                <div class="strategy-return ${returnClass}">${returnText}</div>
            </div>
        `;
    },

    /**
     * 渲染策略列表
     */
    renderStrategyList(strategies, activeStrategyId = null) {
        if (!strategies || strategies.length === 0) {
            return '<div class="empty-state">暂无策略</div>';
        }

        return strategies.map((strategy, index) =>
            this.renderStrategyItem(strategy, index + 1, strategy.id === activeStrategyId)
        ).join('');
    },

    /**
     * 渲染持仓表格行
     */
    renderPositionRow(position, ticker = {}) {
        const side = position.side || 'long';
        const sideClass = side === 'long' ? 'long' : 'short';
        const sideText = side === 'long' ? '多' : '空';

        const currentPrice = ticker.price || position.currentPrice || 0;
        const unrealizedPnl = position.unrealizedPnl || 0;
        const pnlClass = unrealizedPnl >= 0 ? 'positive' : 'negative';

        return `
            <tr>
                <td>${position.symbol || '-'}</td>
                <td class="${sideClass}">${sideText}</td>
                <td>${this.formatNumber(position.quantity)}</td>
                <td>${this.formatNumber(position.entryPrice)}</td>
                <td>${this.formatNumber(currentPrice)}</td>
                <td>${this.formatNumber(position.liquidationPrice)}</td>
                <td class="${pnlClass}">${this.formatNumber(unrealizedPnl)}</td>
            </tr>
        `;
    },

    /**
     * 渲染持仓表格
     */
    renderPositionsTable(positions, tickerMap = {}) {
        if (!positions || positions.length === 0) {
            return '<tr><td colspan="7" class="empty-state">暂无持仓</td></tr>';
        }

        return positions.map(position =>
            this.renderPositionRow(position, tickerMap[position.symbol])
        ).join('');
    },

    /**
     * 渲染成交记录表格行
     */
    renderFillRow(fill) {
        const side = fill.side || 'buy';
        const sideClass = side === 'buy' ? 'long' : 'short';
        const sideText = side === 'buy' ? '买入' : '卖出';

        return `
            <tr>
                <td>${this.formatTime(fill.timestamp || fill.time)}</td>
                <td>${fill.symbol || '-'}</td>
                <td class="${sideClass}">${sideText}</td>
                <td>${this.formatNumber(fill.price)}</td>
                <td>${this.formatNumber(fill.quantity)}</td>
                <td>${this.formatNumber(fill.fee)}</td>
            </tr>
        `;
    },

    /**
     * 渲染成交记录表格
     */
    renderFillsTable(fills) {
        if (!fills || fills.length === 0) {
            return '<tr><td colspan="6" class="empty-state">暂无成交记录</td></tr>';
        }

        return fills.map(fill => this.renderFillRow(fill)).join('');
    },

    /**
     * 渲染订单状态
     */
    renderOrderStatus(status) {
        const statusMap = {
            'pending': { text: '挂单中', class: 'status-pending' },
            'filled': { text: '已成交', class: 'status-filled' },
            'cancelled': { text: '已取消', class: 'status-cancelled' },
            'liquidated': { text: '已强平', class: 'status-liquidated' },
            'partially_filled': { text: '部分成交', class: 'status-pending' }
        };

        const statusInfo = statusMap[status] || { text: status, class: '' };
        return `<span class="${statusInfo.class}">${statusInfo.text}</span>`;
    },

    /**
     * 渲染订单表格行
     */
    renderOrderRow(order) {
        const side = order.side || 'buy';
        const sideText = side === 'buy' ? '买入' : '卖出';
        const sideClass = side === 'buy' ? 'long' : 'short';

        return `
            <tr>
                <td>${this.formatTime(order.timestamp || order.time)}</td>
                <td>${order.symbol || '-'}</td>
                <td>${order.type === 'limit' ? '限价' : '市价'}</td>
                <td class="${sideClass}">${sideText}</td>
                <td>${this.formatNumber(order.price)}</td>
                <td>${this.formatNumber(order.quantity)}</td>
                <td>${this.formatNumber(order.filledQuantity || 0)}</td>
                <td>${this.renderOrderStatus(order.status)}</td>
            </tr>
        `;
    },

    /**
     * 渲染订单表格
     */
    renderOrdersTable(orders, canCancel = false) {
        if (!orders || orders.length === 0) {
            return '<tr><td colspan="8" class="empty-state">暂无订单</td></tr>';
        }

        if (canCancel) {
            return this.renderOrdersTableWithAction(orders, canCancel);
        }

        return orders.map(order => this.renderOrderRow(order)).join('');
    },

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

    /**
     * 渲染强平记录表格行
     */
    renderLiquidationRow(liquidation) {
        const side = liquidation.side || 'long';
        const sideClass = side === 'long' ? 'long' : 'short';
        const sideText = side === 'long' ? '多' : '空';

        return `
            <tr>
                <td>${this.formatTime(liquidation.timestamp || liquidation.time)}</td>
                <td>${liquidation.symbol || '-'}</td>
                <td class="${sideClass}">${sideText}</td>
                <td>${this.formatNumber(liquidation.liquidationPrice)}</td>
                <td>${this.formatNumber(liquidation.quantity)}</td>
            </tr>
        `;
    },

    /**
     * 渲染强平记录表格
     */
    renderLiquidationsTable(liquidations) {
        if (!liquidations || liquidations.length === 0) {
            return '<tr><td colspan="5" class="empty-state">暂无强平记录</td></tr>';
        }

        return liquidations.map(l => this.renderLiquidationRow(l)).join('');
    },

    /**
     * 渲染分页组件
     */
    renderPagination(currentPage, totalPages, onPageChange) {
        if (totalPages <= 1) return '';

        let pages = [];
        let startPage = Math.max(1, currentPage - 2);
        let endPage = Math.min(totalPages, currentPage + 2);

        if (startPage > 1) {
            pages.push(1);
            if (startPage > 2) pages.push('...');
        }

        for (let i = startPage; i <= endPage; i++) {
            pages.push(i);
        }

        if (endPage < totalPages) {
            if (endPage < totalPages - 1) pages.push('...');
            pages.push(totalPages);
        }

        return `
            <div class="pagination">
                <button ${currentPage === 1 ? 'disabled' : ''} data-page="${currentPage - 1}">上一页</button>
                ${pages.map(page => {
                    if (page === '...') {
                        return '<span class="page-info">...</span>';
                    }
                    return `<button class="${page === currentPage ? 'active' : ''}" data-page="${page}">${page}</button>`;
                }).join('')}
                <button ${currentPage === totalPages ? 'disabled' : ''} data-page="${currentPage + 1}">下一页</button>
                <span class="page-info">${currentPage} / ${totalPages}</span>
            </div>
        `;
    },

    /**
     * 渲染行情列表
     */
    renderTickers(tickers) {
        if (!tickers || tickers.length === 0) {
            return '<div class="loading">暂无行情</div>';
        }

        return tickers.map(ticker => {
            const change = ticker.change24h || 0;
            const changeClass = change >= 0 ? 'positive' : 'negative';
            const changeText = this.formatPercent(change);

            return `
                <div class="ticker-item">
                    <span class="ticker-symbol">${ticker.symbol}</span>
                    <span class="ticker-price">$${this.formatNumber(ticker.price)}</span>
                    <span class="ticker-change ${changeClass}">${changeText}</span>
                    <span class="ticker-time">${this.formatTimeShort(ticker.updatedAt)}</span>
                </div>
            `;
        }).join('');
    },

    /**
     * 渲染系统统计
     */
    renderStatistics(stats) {
        return {
            strategyCount: stats.totalStrategies || stats.strategyCount || 0,
            tradeCount: stats.totalTrades || stats.tradeCount || 0,
            orderCount: stats.pendingOrders || stats.orderCount || 0
        };
    }
};

// 导出到全局
window.Components = Components;
