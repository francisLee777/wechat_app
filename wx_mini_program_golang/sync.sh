#!/bin/sh

# 远程服务器配置
REMOTE_USER="root"
REMOTE_HOST="47.97.126.66"
REMOTE_DIR="/root/wx_mini_program/"

# 同步命令
rsync -avz \
 --exclude='.*/' \
 --exclude='data/' \
 ./ "$REMOTE_USER@$REMOTE_HOST:$REMOTE_DIR"

echo "Synchronization complete."