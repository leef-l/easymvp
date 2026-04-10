# Workflow V2 Snake Sample

一个用于 Workflow V2 真实链路验证的全栈贪吃蛇样例：

- `frontend/`：React + TypeScript + Vite，包含贪吃蛇游戏、后端健康状态和排行榜面板
- `backend/`：GoFrame v2 API，提供游戏配置、健康检查和排行榜存储

## 启动方式

### 1. 启动后端

```bash
cd backend
go mod tidy
go run .
```

默认端口：`http://127.0.0.1:18080`

### 2. 启动前端

```bash
cd frontend
npm install
npm run dev
```

默认端口：`http://127.0.0.1:5173`

前端已通过 Vite 代理将 `/api/*` 转发到后端。

## 验证命令

```bash
cd backend && go test ./...
cd frontend && npm run lint
cd frontend && npm run test
cd frontend && npm run build
```
