#!/bin/bash
# EasyMVP 服务器初始化脚本
# 在目标服务器上执行一次，创建 systemd 服务和 nginx 配置
# 使用: bash init-server.sh

set -e

DEPLOY_DIR=/www/wwwroot/mvp.easytestdev.online
DOMAIN=mvp.easytestdev.online
SYSTEM_PORT=${SYSTEM_PORT:-10041}
AI_PORT=${AI_PORT:-10042}
MVP_PORT=${MVP_PORT:-10043}

echo "===== EasyMVP 服务器初始化 ====="
echo "目标端口：system=${SYSTEM_PORT} ai=${AI_PORT} mvp=${MVP_PORT}"

if command -v ss >/dev/null 2>&1; then
    echo "[0/3] 检查端口占用..."
    for port in "$SYSTEM_PORT" "$AI_PORT" "$MVP_PORT"; do
        if ss -ltn | awk 'NR>1 {print $4}' | grep -q ":${port}\$"; then
            echo "  端口 ${port} 已被占用"
        else
            echo "  端口 ${port} 可用"
        fi
    done
fi

# 1. 创建部署目录
echo "[1/3] 创建部署目录..."
for app in system ai mvp admin; do
    mkdir -p $DEPLOY_DIR/$app
done

# 2. 创建 Systemd 服务
echo "[2/3] 创建 Systemd 服务..."

# mvp-system
cat > /etc/systemd/system/mvp-system.service <<EOF
[Unit]
Description=EasyMVP system
After=network.target mysql.service

[Service]
Type=simple
WorkingDirectory=$DEPLOY_DIR/system
Environment=GF_GCFG_PATH=$DEPLOY_DIR/system/manifest/config
Environment=GF_GCFG_FILE=config.yaml
Environment=GF_SERVER_ADDRESS=:$SYSTEM_PORT
Environment=TZ=Asia/Shanghai
ExecStart=$DEPLOY_DIR/system/system
Restart=always
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

# mvp-ai
cat > /etc/systemd/system/mvp-ai.service <<EOF
[Unit]
Description=EasyMVP ai
After=network.target mysql.service

[Service]
Type=simple
WorkingDirectory=$DEPLOY_DIR/ai
Environment=GF_GCFG_PATH=$DEPLOY_DIR/ai/manifest/config
Environment=GF_GCFG_FILE=config.yaml
Environment=GF_SERVER_ADDRESS=:$AI_PORT
Environment=TZ=Asia/Shanghai
ExecStart=$DEPLOY_DIR/ai/ai
Restart=always
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

# mvp-mvp
cat > /etc/systemd/system/mvp-mvp.service <<EOF
[Unit]
Description=EasyMVP mvp
After=network.target mysql.service

[Service]
Type=simple
WorkingDirectory=$DEPLOY_DIR/mvp
Environment=GF_GCFG_PATH=$DEPLOY_DIR/mvp/manifest/config
Environment=GF_GCFG_FILE=config.yaml
Environment=GF_SERVER_ADDRESS=:$MVP_PORT
Environment=TZ=Asia/Shanghai
ExecStart=$DEPLOY_DIR/mvp/mvp
Restart=always
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable mvp-system mvp-ai mvp-mvp

echo "  mvp-system (port ${SYSTEM_PORT}) ✓"
echo "  mvp-ai     (port ${AI_PORT}) ✓"
echo "  mvp-mvp    (port ${MVP_PORT}) ✓"

# 3. 生成 Nginx 扩展配置（不覆盖宝塔主配置）
echo "[3/3] 生成 Nginx 配置..."

NGINX_EXT_DIR=/www/server/panel/vhost/nginx/extension/$DOMAIN
mkdir -p $NGINX_EXT_DIR

cat > $NGINX_EXT_DIR/proxy.conf <<NGINX
# EasyMVP 反向代理配置
# system 服务
location /api/system/ {
    proxy_pass http://127.0.0.1:${SYSTEM_PORT};
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_connect_timeout 30s;
    proxy_read_timeout 120s;
    proxy_send_timeout 120s;
}

# ai 服务
location /api/ai/ {
    proxy_pass http://127.0.0.1:${AI_PORT};
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_connect_timeout 30s;
    proxy_read_timeout 120s;
    proxy_send_timeout 120s;
}

# mvp 服务 — SSE 需要长连接
location /api/mvp/ {
    proxy_pass http://127.0.0.1:${MVP_PORT};
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_connect_timeout 30s;
    proxy_read_timeout 600s;
    proxy_send_timeout 120s;
    proxy_buffering off;
    proxy_cache off;
    chunked_transfer_encoding on;
}

# 前端 (Vue Vben Admin)
location /admin/ {
    alias /www/wwwroot/mvp.easytestdev.online/admin/;
    index index.html;
    try_files $uri $uri/ /admin/index.html;
}

# 默认首页重定向到管理后台
location = / {
    return 302 /admin/;
}
NGINX

nginx -t && nginx -s reload

echo ""
echo "===== 初始化完成 ====="
echo ""
echo "端口分配："
echo "  system : ${SYSTEM_PORT} →  /api/system/"
echo "  ai     : ${AI_PORT} →  /api/ai/"
echo "  mvp    : ${MVP_PORT} →  /api/mvp/"
echo "  前端   : Nginx →  /admin/"
echo ""
echo "管理命令："
echo "  systemctl start|stop|restart mvp-system"
echo "  systemctl start|stop|restart mvp-ai"
echo "  systemctl start|stop|restart mvp-mvp"
echo ""
echo "日志查看："
echo "  journalctl -u mvp-system -f"
echo "  journalctl -u mvp-ai -f"
echo "  journalctl -u mvp-mvp -f"
