#!/bin/bash

# 数据库重置脚本
# 用于清空数据库并重新创建表结构

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== 数据库重置脚本 ===${NC}"

# 检查配置文件
CONFIG_FILE="config.yaml"
EXAMPLE_FILE="config.example.yaml"

if [ ! -f "$CONFIG_FILE" ]; then
    echo -e "${YELLOW}配置文件 $CONFIG_FILE 不存在，正在从示例创建...${NC}"
    cp $EXAMPLE_FILE $CONFIG_FILE
    echo -e "${YELLOW}请编辑 $CONFIG_FILE 填入正确的数据库配置${NC}"
    exit 1
fi

# 从配置文件读取数据库配置
DB_HOST=$(grep -A 5 "database:" $CONFIG_FILE | grep "host:" | awk -F'"' '{print $2}')
DB_PORT=$(grep -A 5 "database:" $CONFIG_FILE | grep "port:" | awk -F'"' '{print $2}')
DB_USER=$(grep -A 5 "database:" $CONFIG_FILE | grep "user:" | awk -F'"' '{print $2}')
DB_PASSWORD=$(grep -A 5 "database:" $CONFIG_FILE | grep "password:" | awk -F'"' '{print $2}')
DB_NAME=$(grep -A 5 "database:" $CONFIG_FILE | grep "dbname:" | awk -F'"' '{print $2}')

# 导出密码（用于 psql 连接）
export PGPASSWORD="$DB_PASSWORD"

echo -e "数据库: ${GREEN}$DB_NAME${NC} @ ${GREEN}$DB_HOST:$DB_PORT${NC}"

# 确认操作
echo -e "${RED}警告: 此操作将删除数据库 $DB_NAME 中的所有数据！${NC}"
read -p "确定继续? (yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    echo "操作已取消"
    exit 0
fi

echo -e "${YELLOW}正在清空数据库...${NC}"

# 方法1: 删除并重建数据库（推荐）
echo "删除并重建数据库..."

# 断开所有连接并删除数据库
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -c "
    SELECT pg_terminate_backend(pg_stat_activity.pid)
    FROM pg_stat_activity
    WHERE pg_stat_activity.datname = '$DB_NAME'
    AND pid <> pg_backend_pid();
" 2>/dev/null || true

psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -c "DROP DATABASE IF EXISTS $DB_NAME;"

# 重建数据库
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -c "CREATE DATABASE $DB_NAME;"

echo -e "${GREEN}数据库已重建！${NC}"

# 清理环境变量
unset PGPASSWORD

echo ""
echo -e "${GREEN}=== 完成 ===${NC}"
echo -e "请重启服务以重新创建数据表"
echo "运行: cd backend && go run cmd/server/main.go"
