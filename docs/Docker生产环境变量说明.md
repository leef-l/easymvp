# Docker 生产环境变量说明

本文说明 `docker/prod/docker-compose.yml` 里用到的环境变量，及 Linux / Windows 下的常见设置方式。

## 需要设置的变量

生产环境至少建议设置这些变量：

- `REDIS_PASSWORD`
- `MYSQL_PASSWORD`
- `JWT_SECRET`

按需设置的变量：

- `DB_HOST`
- `DB_PORT`
- `MYSQL_USER`
- `MYSQL_DATABASE`
- `SYSTEM_PORT`
- `FRONTEND_PORT`
- `VITE_GLOB_API_URL`

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

这种方式只对当前 shell 生效，关闭终端后失效。

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

可以在 `docker/prod/` 下自建一个 `.env` 文件，例如：

```env
REDIS_PASSWORD=your-redis-password
MYSQL_PASSWORD=your-mysql-password
JWT_SECRET=your-jwt-secret
DB_HOST=host.docker.internal
DB_PORT=3306
MYSQL_USER=easymvp
MYSQL_DATABASE=easymvp
SYSTEM_PORT=8000
FRONTEND_PORT=5555
VITE_GLOB_API_URL=http://localhost:8000
```

然后执行：

```bash
docker compose --env-file docker/prod/.env -f docker/prod/docker-compose.yml up -d
```

这是最推荐的方式，便于管理和复用。

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

这种方式只对当前 PowerShell 窗口生效。

### 方式二：CMD 当前窗口临时生效

```cmd
set REDIS_PASSWORD=your-redis-password
set MYSQL_PASSWORD=your-mysql-password
set JWT_SECRET=your-jwt-secret
docker compose -f docker/prod/docker-compose.yml up -d
```

这种方式只对当前 CMD 窗口生效。

### 方式三：系统环境变量

在 Windows 图形界面中：

1. 打开“系统属性”
2. 进入“高级” → “环境变量”
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

设置完成后，关闭并重新打开终端，再执行：

```powershell
docker compose -f docker/prod/docker-compose.yml up -d
```

### 方式四：使用 `.env` 文件

在 `docker/prod/` 下创建 `.env`：

```env
REDIS_PASSWORD=your-redis-password
MYSQL_PASSWORD=your-mysql-password
JWT_SECRET=your-jwt-secret
DB_HOST=host.docker.internal
DB_PORT=3306
MYSQL_USER=easymvp
MYSQL_DATABASE=easymvp
SYSTEM_PORT=8000
FRONTEND_PORT=5555
VITE_GLOB_API_URL=http://localhost:8000
```

然后执行：

```powershell
docker compose --env-file docker/prod/.env -f docker/prod/docker-compose.yml up -d
```

这也是 Windows 下最推荐的方式。

## 推荐做法

开发环境：

- 使用 `docker/dev/.env`

生产环境：

- 使用 `docker/prod/.env`
- 不要把真实密码和密钥提交到 git

## 验证环境变量是否生效

Linux:

```bash
echo $REDIS_PASSWORD
echo $MYSQL_PASSWORD
echo $JWT_SECRET
```

PowerShell:

```powershell
$env:REDIS_PASSWORD
$env:MYSQL_PASSWORD
$env:JWT_SECRET
```

CMD:

```cmd
echo %REDIS_PASSWORD%
echo %MYSQL_PASSWORD%
echo %JWT_SECRET%
```
