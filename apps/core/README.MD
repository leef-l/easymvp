# EasyMVP Core

V3 本地核心服务。

## 本地开发

```bash
go test ./...
go run main.go
```

## 本地构建

```bash
make build
make build-linux
make build-windows
```

## Docker

```bash
make docker-build
```

## 运行配置

- 默认配置文件：`manifest/config/config.yaml`
- 默认数据目录：`./var`
- 默认监听端口：`8000`

## 健康检查

- `GET /api/v3/system/healthz`
