#!/bin/bash

# ============================================
# 排班管理系统 V2.0.1 - 多平台发布构建脚本
# ============================================
#
# 用法:
#   ./release/build-release.sh [版本号]
#
# 示例:
#   ./release/build-release.sh v2.0.1
#
# 前置要求:
#   - Go 1.21+
#   - Node.js 18+
#   - zip (打包用)
#
# 输出:
#   release/v2.0.1/
#   ├── linux-amd64/     ... + .zip
#   ├── linux-arm64/     ... + .zip
#   ├── windows-amd64/   ... + .zip
#   ├── windows-arm64/   ... + .zip
#   └── frontend.zip     (纯前端包)
# ============================================

set -e

VERSION=${1:-"v2.0.1"}
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
OUTPUT_DIR="$SCRIPT_DIR/$VERSION"

echo "================================"
echo "  排班管理系统 $VERSION 构建脚本"
echo "================================"
echo "项目目录: $PROJECT_DIR"
echo "输出目录: $OUTPUT_DIR"
echo ""

# ---- 构建前端 ----
echo "[1/5] 构建前端..."
cd "$PROJECT_DIR/frontend-v2"
npm install --silent 2>/dev/null
npm run build 2>&1 | grep -E "(error|built in)" || true
echo "前端构建完成"
echo ""

# ---- 构建后端（多平台） ----
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
    "windows/arm64"
)

echo "[2/5] 构建后端（多平台）..."
for PLATFORM in "${PLATFORMS[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$PLATFORM"
    OUTPUT_NAME="schedule-server"
    if [ "$GOOS" = "windows" ]; then
        OUTPUT_NAME="schedule-server.exe"
    fi

    echo "  - 编译 $GOOS/$GOARCH ..."
    cd "$PROJECT_DIR/backend"
    CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH \
        go build -ldflags="-s -w -X main.version=$VERSION" \
        -o "$OUTPUT_DIR/${GOOS}-${GOARCH}/$OUTPUT_NAME" \
        ./cmd/server/main.go 2>&1
done
echo "后端构建完成"
echo ""

# ---- 复制公共文件 ----
echo "[3/5] 复制公共文件..."

for PLATFORM in "${PLATFORMS[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$PLATFORM"
    PLATFORM_DIR="$OUTPUT_DIR/${GOOS}-${GOARCH}"

    # 前端资源
    mkdir -p "$PLATFORM_DIR/dist"
    cp -r "$PROJECT_DIR/frontend-v2/dist/"* "$PLATFORM_DIR/dist/"

    # 配置文件
    mkdir -p "$PLATFORM_DIR/configs"
    cp "$PROJECT_DIR/backend/configs/config.example.yaml" "$PLATFORM_DIR/configs/config.example.yaml"

    # 文档
    mkdir -p "$PLATFORM_DIR/docs"
    cp "$PROJECT_DIR/docs/"*.md "$PLATFORM_DIR/docs/"
done
echo "公共文件复制完成"
echo ""

# ---- 打包 ----
echo "[4/5] 打包..."

cd "$OUTPUT_DIR"

# 各平台 zip
for PLATFORM in "${PLATFORMS[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$PLATFORM"
    ZIP_NAME="schedule-system-${VERSION}-${GOOS}-${GOARCH}.zip"
    echo "  - 打包 $ZIP_NAME ..."
    cd "$OUTPUT_DIR/${GOOS}-${GOARCH}"
    zip -rq "$OUTPUT_DIR/$ZIP_NAME" . \
        -x "*.exe" 2>/dev/null || true
    # 重新打包包含 exe（Windows）
    if [ "$GOOS" = "windows" ]; then
        rm -f "$OUTPUT_DIR/$ZIP_NAME"
        zip -rq "$OUTPUT_DIR/$ZIP_NAME" .
    fi
done

# 纯前端包
echo "  - 打包前端..."
cd "$PROJECT_DIR/frontend-v2"
zip -rq "$OUTPUT_DIR/schedule-system-${VERSION}-frontend.zip" dist/

echo "打包完成"
echo ""

# ---- 汇总 ----
echo "[5/5] 汇总"
echo ""
echo "================================"
echo "  构建完成! ($VERSION)"
echo "================================"
echo ""
echo "输出目录: $OUTPUT_DIR/"
echo ""
echo "文件列表:"
ls -lh "$OUTPUT_DIR/" | grep -v "^total"
echo ""
echo "各平台内容:"
for PLATFORM in "${PLATFORMS[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$PLATFORM"
    echo ""
    echo "  [$GOOS/$GOARCH]"
    ls -lh "$OUTPUT_DIR/${GOOS}-${GOARCH}/" 2>/dev/null | grep -v "^total" | grep -v "^d" | awk '{print "    " $NF " (" $5 ")"}'
done
echo ""
echo "部署步骤:"
echo "  1. 将对应平台目录上传到服务器"
echo "  2. 解压或直接使用"
echo "  3. 修改 configs/config.yaml 配置数据库"
echo "  4. 运行 start.sh (Linux) 或 start.bat (Windows)"
echo ""
