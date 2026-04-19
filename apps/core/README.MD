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

## 发布

- GitHub 校验流：`.github/workflows/v3-guard.yml`
- GitHub core 发布流：`.github/workflows/core-release.yml`
- 打 `core-v*` tag 后会产出 Linux / Windows 二进制，并推送 `ghcr.io/leef-l/easymvp-core`

## 运行配置

- 默认配置文件：`manifest/config/config.yaml`
- 默认数据目录：`./var`
- 默认监听端口：`8000`

## 健康检查

- `GET /api/v3/system/healthz`

## Kubernetes

- 开发 overlay：`manifest/deploy/kustomize/overlays/develop`
- 生产 overlay：`manifest/deploy/kustomize/overlays/production`
- 生产密钥样例：`manifest/deploy/kustomize/overlays/production/secret.env.example`
