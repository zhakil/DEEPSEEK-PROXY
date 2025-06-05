#!/bin/bash

# 简化的测试构建脚本
# 这个脚本会逐步验证每个组件，帮助我们定位问题

echo "🔍 Go语言代理服务器 - 诊断构建脚本"
echo "================================="

# 第一步：检查Go环境
echo "📋 步骤1: 检查Go环境"
if ! command -v go &> /dev/null; then
    echo "❌ Go编程语言未安装"
    exit 1
fi

GO_VERSION=$(go version)
echo "✅ Go版本: $GO_VERSION"

# 第二步：检查必需文件
echo ""
echo "📋 步骤2: 检查项目文件"
REQUIRED_FILES=("main.go" "types.go" "config.go" "server.go" "handlers.go" "utils.go" "go.mod")

for file in "${REQUIRED_FILES[@]}"; do
    if [ -f "$file" ]; then
        echo "✅ $file 存在"
    else
        echo "❌ $file 缺失"
        MISSING_FILES=true
    fi
done

if [ "$MISSING_FILES" = true ]; then
    echo "❌ 项目文件不完整，请确保所有必需文件都存在"
    exit 1
fi

# 第三步：检查包声明一致性
echo ""
echo "📋 步骤3: 检查包声明"
for file in *.go; do
    if [ "$file" != "test_client.go" ] && [ "$file" != "test_reasoner.go" ]; then
        PACKAGE=$(head -1 "$file" | grep "package")
        echo "📄 $file: $PACKAGE"
    fi
done

# 第四步：检查语法错误
echo ""
echo "📋 步骤4: 检查语法"
if go fmt ./...; then
    echo "✅ 语法格式检查通过"
else
    echo "❌ 存在语法格式问题"
    exit 1
fi

# 第五步：逐个验证文件编译
echo ""
echo "📋 步骤5: 逐个验证核心文件"

# 验证类型定义
echo "🔍 验证 types.go..."
if go build -o /dev/null types.go 2>/dev/null; then
    echo "✅ types.go 编译成功"
else
    echo "⚠️ types.go 可能存在问题"
fi

# 验证配置模块
echo "🔍 验证 config.go..."
if go build -o /dev/null config.go types.go 2>/dev/null; then
    echo "✅ config.go 编译成功"
else
    echo "⚠️ config.go 可能存在问题"
fi

# 验证工具函数
echo "🔍 验证 utils.go..."
if go build -o /dev/null utils.go types.go 2>/dev/null; then
    echo "✅ utils.go 编译成功"
else
    echo "⚠️ utils.go 可能存在问题"
fi

# 第六步：检查Go模块
echo ""
echo "📋 步骤6: 检查依赖"
if go mod tidy; then
    echo "✅ 依赖检查完成"
else
    echo "❌ 依赖问题"
    exit 1
fi

# 第七步：尝试完整构建
echo ""
echo "📋 步骤7: 尝试完整构建"
echo "🔨 开始构建..."

if go build -v -o deepseek-proxy-test . 2>&1; then
    echo "🎉 构建成功！"
    echo "✅ 生成的可执行文件: deepseek-proxy-test"
    
    # 清理测试文件
    rm -f deepseek-proxy-test
    
    echo ""
    echo "📖 下一步："
    echo "1. 创建 .env 文件并配置你的 DEEPSEEK_API_KEY"
    echo "2. 运行 go run . 启动服务器"
    echo "3. 测试基本功能"
    
else
    echo "❌ 构建失败"
    echo ""
    echo "🔧 可能的解决方案："
    echo "1. 检查上面的错误信息"
    echo "2. 确保所有Go文件都以 'package main' 开头"
    echo "3. 确保 handlers.go 包含完整的方法定义"
    echo "4. 检查是否有语法错误或拼写错误"
    exit 1
fi