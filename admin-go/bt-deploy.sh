#!/bin/bash
set -e

# ============================================
# GBaseAdmin 宝塔面板部署脚本
# 依赖: 宝塔 Nginx + 宝塔 MySQL
# ============================================

# ---------- 配置区 ----------
DOMAIN="pw.easytestdev.online"          # 替换为你的域名
DEPLOY_DIR="/www/wwwroot/${DOMAIN}"
FRONTEND_DIR="/www/wwwroot/${DOMAIN}"
DB_NAME="gbaseadmin"
DB_USER="gbaseadmin"
DB_PASS="gbaseadmin123"
DB_HOST="127.0.0.1"
DB_PORT="3306"
JWT_SECRET="gbaseadmin-secret-key-2026"

APPS=("system" "play" "upload")
PORTS=("8000" "8001" "8002")

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SOURCE_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# 加载 Go 环境（sudo 下 PATH 可能丢失）
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
if ! command -v go &>/dev/null; then
  echo "错误: 找不到 go 命令，请确认 Go 已安装" && exit 1
fi

# ---------- 颜色 ----------
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info()  { echo -e "${GREEN}[INFO]${NC} $1"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }

# ---------- 1. 创建目录 ----------
info "创建部署目录..."
mkdir -p "$DEPLOY_DIR"
mkdir -p "$FRONTEND_DIR"
for app in "${APPS[@]}"; do
  mkdir -p "$DEPLOY_DIR/$app/manifest/config"
  mkdir -p "$DEPLOY_DIR/$app/resource/upload"
done

# ---------- 2. 停止旧服务 ----------
info "停止旧服务..."
for app in "${APPS[@]}"; do
  if systemctl is-active "gba-${app}" &>/dev/null; then
    systemctl stop "gba-${app}"
    info "gba-${app} 已停止"
  fi
done

# ---------- 3. 编译 ----------
info "开始编译..."
cd "$SCRIPT_DIR"
for i in "${!APPS[@]}"; do
  app="${APPS[$i]}"
  info "编译 $app ..."
  cd "app/$app"
  CGO_ENABLED=0 GOOS=linux go build -o "$DEPLOY_DIR/$app/$app" .
  chmod +x "$DEPLOY_DIR/$app/$app"
  cd "$SCRIPT_DIR"
  info "$app 编译完成"
done

# ---------- 4. 生成配置文件 ----------
info "生成配置文件..."
for i in "${!APPS[@]}"; do
  app="${APPS[$i]}"
  port="${PORTS[$i]}"
  conf="$DEPLOY_DIR/$app/manifest/config/config.yaml"
  if [ -f "$conf" ]; then
    warn "$app 配置已存在，跳过"
    continue
  fi
  cat > "$conf" <<EOF
server:
  address: ":${port}"
  openapiPath: "/api.json"
  swaggerPath: "/swagger"

logger:
  level: "warning"
  stdout: true
  path: "$DEPLOY_DIR/$app/logs"

database:
  default:
    link: "mysql:${DB_USER}:${DB_PASS}@tcp(${DB_HOST}:${DB_PORT})/${DB_NAME}"
    debug: false
    charset: "utf8mb4"

jwt:
  secret: "${JWT_SECRET}"
  expire: 24
EOF
  info "$app 配置已生成"
done

# play 额外添加 redis 配置（如需要取消注释）
cat >> "$DEPLOY_DIR/play/manifest/config/config.yaml" <<EOF

redis:
  default:
    address: "127.0.0.1:6379"
    pass: ""
    db: 1
EOF

# ---------- 5. 初始化数据库 ----------
info "检查数据库..."
# 用 root 检查数据库是否存在（DB_USER 可能尚无权限）
if mysql -u root -p"$DB_PASS" -h"$DB_HOST" -P"$DB_PORT" -e "USE $DB_NAME" 2>/dev/null; then
  warn "数据库 $DB_NAME 已存在，跳过初始化"
else
  info "创建数据库并导入..."
  mysql -u root -p"$DB_PASS" -h"$DB_HOST" -P"$DB_PORT" -e "CREATE DATABASE IF NOT EXISTS \`$DB_NAME\` DEFAULT CHARSET utf8mb4 COLLATE utf8mb4_general_ci;"
  mysql -u root -p"$DB_PASS" -h"$DB_HOST" -P"$DB_PORT" -e "CREATE USER IF NOT EXISTS '$DB_USER'@'%' IDENTIFIED BY '$DB_PASS';"
  mysql -u root -p"$DB_PASS" -h"$DB_HOST" -P"$DB_PORT" -e "CREATE USER IF NOT EXISTS '$DB_USER'@'localhost' IDENTIFIED BY '$DB_PASS';"
  mysql -u root -p"$DB_PASS" -h"$DB_HOST" -P"$DB_PORT" -e "GRANT ALL ON \`$DB_NAME\`.* TO '$DB_USER'@'%'; GRANT ALL ON \`$DB_NAME\`.* TO '$DB_USER'@'localhost'; FLUSH PRIVILEGES;"
  if [ -f "$SCRIPT_DIR/codegen/sql/init.sql" ]; then
    mysql -u root -p"$DB_PASS" -h"$DB_HOST" -P"$DB_PORT" --default-character-set=utf8mb4 "$DB_NAME" < "$SCRIPT_DIR/codegen/sql/init.sql"
    info "数据库导入完成"
  else
    warn "init.sql 不存在，请手动导入数据库"
  fi
fi

# ---------- 6. 创建 systemd 服务 ----------
info "创建 systemd 服务..."
for i in "${!APPS[@]}"; do
  app="${APPS[$i]}"
  cat > "/etc/systemd/system/gba-${app}.service" <<EOF
[Unit]
Description=GBaseAdmin ${app}
After=network.target mysql.service

[Service]
Type=simple
WorkingDirectory=$DEPLOY_DIR/$app
ExecStart=$DEPLOY_DIR/$app/$app
Restart=always
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF
done

systemctl daemon-reload
for app in "${APPS[@]}"; do
  systemctl enable "gba-${app}"
  systemctl restart "gba-${app}"
  info "gba-${app} 服务已启动"
done

# ---------- 7. 生成宝塔 Nginx 反向代理配置 ----------
info "生成 Nginx 配置..."
NGINX_CONF="/www/server/panel/vhost/nginx/${DOMAIN}.conf"
cat > "$NGINX_CONF" <<'NGINXEOF'
server {
    listen 80;
    server_name DOMAIN_PLACEHOLDER;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name DOMAIN_PLACEHOLDER;
    root FRONTEND_PLACEHOLDER;
    index index.html;

    ssl_certificate /www/server/panel/vhost/cert/DOMAIN_PLACEHOLDER/fullchain.pem;
    ssl_certificate_key /www/server/panel/vhost/cert/DOMAIN_PLACEHOLDER/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers on;

    client_max_body_size 50m;

    # 静态资源缓存
    location ~* \.(js|css|woff2?|ttf|eot)$ {
        expires 30d;
        add_header Cache-Control "public";
        access_log off;
    }

    # 图片静态访问（优先上传目录，回退前端目录）
    location ~* \.(png|jpg|jpeg|gif|ico|svg|bmp|webp)$ {
        root DEPLOY_PLACEHOLDER/upload/resource;
        try_files $uri @frontend_static;
        expires 30d;
        add_header Cache-Control "public";
        access_log off;
    }

    location @frontend_static {
        root FRONTEND_PLACEHOLDER;
        expires 30d;
        add_header Cache-Control "public";
        access_log off;
    }

    # 上传文件静态访问
    location ^~ /upload/ {
        alias DEPLOY_PLACEHOLDER/upload/resource/upload/;
        expires 30d;
        add_header Cache-Control "public";
        access_log off;
    }

    # play 后端 API（具体路径优先匹配）
    location /api/play/ {
        proxy_pass http://127.0.0.1:8001;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_connect_timeout 10s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # play WAP端 API（直接透传，后端有独立 /api/playapi 路由）
    location /api/playapi/ {
        proxy_pass http://127.0.0.1:8001;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_connect_timeout 10s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # upload 后端 API
    location /api/upload/ {
        proxy_pass http://127.0.0.1:8002;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_connect_timeout 10s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # system 后端 API（兜底，匹配其余 /api/ 请求）
    location /api/ {
        proxy_pass http://127.0.0.1:8000;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_connect_timeout 10s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # 管理端 Vue Router history 模式
    location /admin/ {
        try_files $uri $uri/ /admin/index.html;
        location = /admin/index.html {
            add_header Cache-Control "no-cache, no-store, must-revalidate";
        }
    }

    # WAP端 Vue Router history 模式
    location /wap/ {
        try_files $uri $uri/ /wap/index.html;
        location = /wap/index.html {
            add_header Cache-Control "no-cache, no-store, must-revalidate";
        }
    }

    # 根路径跳转到 WAP
    location / {
        return 302 /wap/;
    }

    # 安全 Header
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;

    error_page 404 /index.html;
}
NGINXEOF

# 替换占位符
sed -i "s|DOMAIN_PLACEHOLDER|${DOMAIN}|g" "$NGINX_CONF"
sed -i "s|FRONTEND_PLACEHOLDER|${FRONTEND_DIR}|g" "$NGINX_CONF"
sed -i "s|DEPLOY_PLACEHOLDER|${DEPLOY_DIR}|g" "$NGINX_CONF"

# 测试并重载 Nginx
nginx -t && nginx -s reload
info "Nginx 配置已生效"

# ---------- 8. 完成 ----------
echo ""
info "========================================="
info "部署完成！"
info "========================================="
info "后端部署目录: $DEPLOY_DIR"
info "前端部署目录: $FRONTEND_DIR"
info "服务管理:"
for app in "${APPS[@]}"; do
  info "  systemctl status/restart gba-${app}"
done
info ""
info "前端构建后将 dist 内容复制到 $FRONTEND_DIR:"
info "  cd $SOURCE_DIR/vue-vben-admin && pnpm build"
info "  cp -rf apps/web-antd/dist/* $FRONTEND_DIR/"
info "========================================="
