# 策略 API-Key 显示功能设计

## 需求

1. 创建策略时自动生成 API-Key（已有功能）
2. 在策略详情弹窗中显示 API-Key（新增）

## 后端改动

### 新增 API 端点

- **端点**: `GET /api/v1/strategies/{id}/api-key`
- **方法**: GET
- **请求头**: `X-Admin-Password`（必填）
- **响应**: `{"api_key": "xxx"}`

### 路由配置

文件: `backend/internal/router/router.go`

```go
api.HandleFunc("/strategies/{id}/api-key", strategyHandler.GetAPIKey).Methods("GET")
```

## 前端改动

### 策略详情弹窗

文件: `frontend/strategy.html`

- 添加 API-Key 显示区域
- 显示格式：API-Key 值 + 复制按钮

### API 调用

文件: `frontend/js/api.js`

- 添加 `getAPIKey(strategyId, adminPassword)` 方法

### 逻辑实现

文件: `frontend/js/strategy.js`

- 详情弹窗打开时自动获取 API-Key（需要验证密码）

## 交互流程

1. 用户点击策略查看详情
2. 弹出详情弹窗 + 密码输入框
3. 输入 admin password
4. 获取并显示 API-Key + 复制按钮
