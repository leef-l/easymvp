# Docker 生产环境变量说明

> 更新日期：2026-04-08

本文说明 `docker/prod/docker-compose.yml` 里实际用到的环境变量，并补充当前生产 compose 的边界。

## 当前 compose 的组成

`docker/prod/docker-compose.yml` 当前会启动：

- `redis`
- `system`
- `ai`
- `mvp`
- `frontend`
- `nginx`
- `gotools`（仅 `tools` profile）

注意：

- 这套 compose 依赖“宿主机已有 MySQL”，不会自建数据库容器
- `frontend` 当前仍以 `pnpm dev:antd` 方式运行，更接近“部署便利版”而不是严格的静态化生产方案

## 必填变量

生产环境至少要提供：

- `REDIS_PASSWORD`
- `MYSQL_PASSWORD`
- `JWT_SECRET`

## 常用可选变量

按当前 compose，常见可调参数包括：

- `DB_HOST`
- `DB_PORT`
- `MYSQL_USER`
- `MYSQL_DATABASE`
- `REDIS_PORT`
- `SYSTEM_PORT`
- `AI_PORT`
- `MVP_PORT`
- `FRONTEND_PORT`
- `NGINX_PORT`
- `VITE_GLOB_API_URL`
- `WORK_DIR`

## 推荐的 `.env` 示例

建议在 `docker/prod/.env` 中统一维护：

```env
REDIS_PASSWORD=your-redis-password
MYSQL_PASSWORD=your-mysql-password
JWT_SECRET=your-jwt-secret

DB_HOST=host.docker.internal
DB_PORT=3306
MYSQL_USER=easymvp
MYSQL_DATABASE=easymvp

REDIS_PORT=6379
SYSTEM_PORT=8000
AI_PORT=8001
MVP_PORT=8002
FRONTEND_PORT=5555
NGINX_PORT=80

VITE_GLOB_API_URL=http://localhost:8000
WORK_DIR=/www/wwwroot/projects
```

然后执行：

```bash
docker compose --env-file docker/prod/.env -f docker/prod/docker-compose.yml up -d
```

## Linux 设置方式

### 方式一：当前终端临时生效

```bash
export REDIS_PASSWORD='your-redis-password'
export MYSQL_PASSWORD='your-mysql-password'
export JWT_SECRET='your-jwt-secret'
```

然后执行：

```bash
docker compose -f docker/prod/docker-compose.yml up -d
```

### 方式二：写入当前用户 shell 配置

如果你用的是 `bash`：

```bash
echo "export REDIS_PASSWORD='your-redis-password'" >> ~/.bashrc
echo "export MYSQL_PASSWORD='your-mysql-password'" >> ~/.bashrc
echo "export JWT_SECRET='your-jwt-secret'" >> ~/.bashrc
source ~/.bashrc
```

如果你用的是 `zsh`：

```bash
echo "export REDIS_PASSWORD='your-redis-password'" >> ~/.zshrc
echo "export MYSQL_PASSWORD='your-mysql-password'" >> ~/.zshrc
echo "export JWT_SECRET='your-jwt-secret'" >> ~/.zshrc
source ~/.zshrc
```

### 方式三：使用 `.env` 文件

这是最推荐的方式，便于管理和复用：

```bash
docker compose --env-file docker/prod/.env -f docker/prod/docker-compose.yml up -d
```

## Windows 设置方式

### 方式一：PowerShell 当前窗口临时生效

```powershell
$env:REDIS_PASSWORD = 'your-redis-password'
$env:MYSQL_PASSWORD = 'your-mysql-password'
$env:JWT_SECRET = 'your-jwt-secret'
```

然后执行：

```powershell
docker compose -f docker/prod/docker-compose.yml up -d
```

### 方式二：CMD 当前窗口临时生效

```cmd
set REDIS_PASSWORD=your-redis-password
set MYSQL_PASSWORD=your-mysql-password
set JWT_SECRET=your-jwt-secret
docker compose -f docker/prod/docker-compose.yml up -d
```

### 方式三：系统环境变量

在 Windows 图形界面中：

1. 打开“系统属性”
2. 进入“高级” -> “环境变量”
3. 在“用户变量”或“系统变量”中新增：
   - `REDIS_PASSWORD`
   - `MYSQL_PASSWORD`
   - `JWT_SECRET`
4. 保存后重新打开终端

也可以用 PowerShell 持久写入当前用户环境变量：

```powershell
[System.Environment]::SetEnvironmentVariable('REDIS_PASSWORD', 'your-redis-password', 'User')
[System.Environment]::SetEnvironmentVariable('MYSQL_PASSWORD', 'your-mysql-password', 'User')
[System.Environment]::SetEnvironmentVariable('JWT_SECRET', 'your-jwt-secret', 'User')
```

### 方式四：使用 `.env` 文件

```powershell
docker compose --env-file docker/prod/.env -f docker/prod/docker-compose.yml up -d
```

## 推荐做法

开发环境：

- 使用 `docker/dev/.env`

生产环境：

- 使用 `docker/prod/.env`
- 不要把真实密码和密钥提交到 git
- 明确宿主机 MySQL 访问策略和 `WORK_DIR` 挂载路径

## 验证环境变量是否生效

Linux：

```bash
echo $REDIS_PASSWORD
echo $MYSQL_PASSWORD
echo $JWT_SECRET
```

PowerShell：

```powershell
$env:REDIS_PASSWORD
$env:MYSQL_PASSWORD
$env:JWT_SECRET
```

CMD：

```cmd
echo %REDIS_PASSWORD%
echo %MYSQL_PASSWORD%
echo %JWT_SECRET%
```
