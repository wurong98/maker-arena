/**
 * API 客户端
 * 负责与后端 API 通信
 */

const API_BASE_URL = '/api/v1';

const ApiClient = {
    /**
     * 通用请求方法
     */
    async request(endpoint, options = {}) {
        const url = `${API_BASE_URL}${endpoint}`;
        const config = {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            },
            ...options
        };

        try {
            const response = await fetch(url, config);
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            return await response.json();
        } catch (error) {
            console.error(`API Error [${endpoint}]:`, error);
            throw error;
        }
    },

    /**
     * GET 请求
     */
    async get(endpoint, params = {}) {
        const queryString = new URLSearchParams(params).toString();
        const url = queryString ? `${endpoint}?${queryString}` : endpoint;
        return this.request(url, { method: 'GET' });
    },

    /**
     * POST 请求
     */
    async post(endpoint, data = {}, customHeaders = {}) {
        return this.request(endpoint, {
            method: 'POST',
            body: JSON.stringify(data),
            headers: customHeaders
        });
    },

    /**
     * PUT 请求
     */
    async put(endpoint, data = {}) {
        return this.request(endpoint, {
            method: 'PUT',
            body: JSON.stringify(data)
        });
    },

    /**
     * DELETE 请求
     */
    async delete(endpoint, customHeaders = {}) {
        return this.request(endpoint, { method: 'DELETE', headers: customHeaders });
    },

    // ===== API 接口 =====

    /**
     * 获取策略列表
     * @param {number} page - 页码
     * @param {number} limit - 每页数量
     */
    async getStrategies(page = 1, limit = 20) {
        return this.get('/strategies', { page, limit });
    },

    /**
     * 创建策略
     * @param {Object} strategy - 策略数据
     * @param {string} strategy.name - 策略名称
     * @param {string} strategy.description - 描述
     * @param {string} strategy.balance - 初始资金
     * @param {string} strategy.apiKey - API 密钥
     * @param {string} adminPassword - 管理员密码
     */
    async createStrategy(strategy, adminPassword) {
        return this.post('/strategies', strategy, {
            headers: { 'X-Admin-Password': adminPassword }
        });
    },

    /**
     * 删除策略
     * @param {string} id - 策略 ID
     * @param {string} adminPassword - 管理员密码
     */
    async deleteStrategy(id, adminPassword) {
        return this.delete(`/strategies/${id}`, {
            headers: { 'X-Admin-Password': adminPassword }
        });
    },

    /**
     * 获取系统统计
     */
    async getStatistics() {
        return this.get('/statistics');
    },

    /**
     * 获取持仓
     * @param {string} strategyId - 策略 ID
     */
    async getPosition(strategyId) {
        return this.get('/exchange/getPosition', { strategyId });
    },

    /**
     * 获取成交记录
     * @param {string} strategyId - 策略 ID
     * @param {number} page - 页码
     * @param {number} limit - 每页数量
     */
    async getFills(strategyId, page = 1, limit = 20) {
        return this.get('/fills', { strategyId, page, limit });
    },

    /**
     * 获取订单
     * @param {string} strategyId - 策略 ID
     * @param {number} page - 页码
     * @param {number} limit - 每页数量
     */
    async getOrders(strategyId, page = 1, limit = 20) {
        return this.get('/exchange/getOrders', { strategyId, page, limit });
    },

    /**
     * 获取账户快照（余额曲线）
     * @param {string} strategyId - 策略 ID
     */
    async getAccountSnapshots(strategyId) {
        return this.get('/snapshots/account', { strategyId });
    },

    /**
     * 获取持仓快照（持仓价值曲线）
     * @param {string} strategyId - 策略 ID
     */
    async getPositionSnapshots(strategyId) {
        return this.get('/snapshots/position', { strategyId });
    },

    /**
     * 获取强平记录
     * @param {string} strategyId - 策略 ID
     * @param {number} page - 页码
     * @param {number} limit - 每页数量
     */
    async getLiquidations(strategyId, page = 1, limit = 20) {
        return this.get('/liquidations', { strategyId, page, limit });
    },

    /**
     * 获取行情
     */
    async getMarketTicker() {
        return this.get('/market/ticker');
    }
};

// 导出到全局
window.ApiClient = ApiClient;
