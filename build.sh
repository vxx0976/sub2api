#!/bin/bash

set -e

# 添加常见的 Go 路径
export PATH="/opt/homebrew/bin:/usr/local/go/bin:$HOME/go/bin:$PATH"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 项目根目录
ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
FRONTEND_DIR="$ROOT_DIR/frontend"
BACKEND_DIR="$ROOT_DIR/backend"
OUTPUT_FILE="$ROOT_DIR/sub2api"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  Sub2API 构建脚本${NC}"
echo -e "${GREEN}========================================${NC}"

# 1. 打包前端
echo -e "\n${YELLOW}[1/2] 打包前端...${NC}"
cd "$FRONTEND_DIR"
if ! command -v pnpm &> /dev/null; then
    echo -e "${RED}错误: 未找到 pnpm，请先安装 pnpm${NC}"
    exit 1
fi
pnpm run build
echo -e "${GREEN}✓ 前端打包完成${NC}"

# 2. 编译后端（嵌入前端，交叉编译为 Linux amd64）
echo -e "\n${YELLOW}[2/2] 编译后端 (Linux amd64)...${NC}"
cd "$BACKEND_DIR"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags embed -ldflags="-s -w" -o "$OUTPUT_FILE" ./cmd/server
echo -e "${GREEN}✓ 后端编译完成${NC}"

# 完成
echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}  构建成功！${NC}"
echo -e "${GREEN}========================================${NC}"
echo -e "输出文件: ${YELLOW}$OUTPUT_FILE${NC}"
echo -e "文件大小: ${YELLOW}$(du -h "$OUTPUT_FILE" | cut -f1)${NC}"
echo -e "\n上传到服务器后执行:"
echo -e "  ${YELLOW}sudo systemctl stop sub2api${NC}"
echo -e "  ${YELLOW}sudo cp sub2api /opt/sub2api/sub2api${NC}"
echo -e "  ${YELLOW}sudo systemctl start sub2api${NC}"
