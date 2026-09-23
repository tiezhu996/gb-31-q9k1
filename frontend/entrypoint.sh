#!/bin/sh
set -e
# Next.js standalone 使用 HOSTNAME 决定监听地址，这里强制 0.0.0.0 供 nginx 反代
export HOSTNAME=0.0.0.0
export PORT=3000
node /app/server.js &
NGINX_PID=0
trap 'kill $NGINX_PID 2>/dev/null || true; exit 0' TERM INT
nginx -g 'daemon off;' &
NGINX_PID=$!
wait
