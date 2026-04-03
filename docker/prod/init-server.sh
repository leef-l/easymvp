#!/bin/bash
# EasyMVP 服务器初始化脚本
# 在目标服务器上执行一次，创建 systemd 服务和 nginx 配置
# 使用: bash init-server.sh

set -e

DEPLOY_DIR=/www/wwwroot/mvp.easytestdev.online
DOMAIN=mvp.easytestdev.online

echo "===== EasyMVP 服务器初始化 ====="

# 1. 创建部署目录
echo "[1/3] 创建部署目录..."
for app in system ai mvp admin; do
    mkdir -p $DEPLOY_DIR/$app
done

# 2. 创建 Systemd 服务
echo "[2/3] 创建 Systemd 服务..."

# mvp-system (port 9000)
cat > /etc/systemd/system/mvp-system.service <<'EOF'
[Unit]
Description=EasyMVP system
After=network.target mysql.service

[Service]
Type=simple
WorkingDirectory=/www/wwwroot/mvp.easytestdev.online/system
ExecStart=/www/wwwroot/mvp.easytestdev.online/system/system
Restart=always
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

# mvp-ai (port 9001)
cat > /etc/systemd/system/mvp-ai.service <<'EOF'
[Unit]
Description=EasyMVP ai
After=network.target mysql.service

[Service]
Type=simple
WorkingDirectory=/www/wwwroot/mvp.easytestdev.online/ai
ExecStart=/www/wwwroot/mvp.easytestdev.online/ai/ai
Restart=always
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

# mvp-mvp (port 9002)
cat > /etc/systemd/system/mvp-mvp.service <<'EOF'
[Unit]
Description=EasyMVP mvp
After=network.target mysql.service

[Service]
Type=simple
WorkingDirectory=/www/wwwroot/mvp.easytestdev.online/mvp
ExecStart=/www/wwwroot/mvp.easytestdev.online/mvp/mvp
Restart=always
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable mvp-system mvp-ai mvp-mvp

echo "  mvp-system (port 9000) ✓"
echo "  mvp-ai     (port 9001) ✓"
echo "  mvp-mvp    (port 9002) ✓"

# 3. 生成 Nginx 扩展配置（不覆盖宝塔主配置）
echo "[3/3] 生成 Nginx 配置..."

NGINX_EXT_DIR=/www/server/panel/vhost/nginx/extension/$DOMAIN
mkdir -p $NGINX_EXT_DIR

cat > $NGINX_EXT_DIR/proxy.conf <<'NGINX'
# EasyMVP 反向代理配置
# system 服务 (port 9000)
location /api/system/ {
    proxy_pass http://127.0.0.1:9000;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_connect_timeout 30s;
    proxy_read_timeout 120s;
    proxy_send_timeout 120s;
}

# ai 服务 (port 9001)
location /api/ai/ {
    proxy_pass http://127.0.0.1:9001;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_connect_timeout 30s;
    proxy_read_timeout 120s;
    proxy_send_timeout 120s;
}

# mvp 服务 (port 9002) — SSE 需要长连接
location /api/mvp/ {
    proxy_pass http://127.0.0.1:9002;
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
echo "  system : 9000  →  /api/system/"
echo "  ai     : 9001  →  /api/ai/"
echo "  mvp    : 9002  →  /api/mvp/"
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
