#!/bin/bash

# DeepSeek API 代理服务器启动脚本
# 这个脚本提供了一个简单的方式来启动服务器

echo "🚀 启动 DeepSeek API 代理服务器（支持推理模型）"
echo "=============================================="

# 检查命令行参数
if [ "$1" = "test-reasoner" ]; then
    echo "🧠 运行推理模型专项测试模式"
    TEST_REASONER=true
else
    TEST_REASONER=false
fi

# 检查 .env 文件是否存在
if [ ! -f ".env" ]; then
    echo "❌ 错误: .env 文件不存在"
    echo "请创建 .env 文件并配置你的 DeepSeek API 密钥"
    echo ""
    echo "示例 .env 文件内容:"
    echo "DEEPSEEK_API_KEY=sk-your-deepseek-api-key-here"
    echo "PORT=9000"
    echo "DEEPSEEK_MODEL=deepseek-reasoner"
    exit 1
fi

# 检查 Go 是否已安装
if ! command -v go &> /dev/null; then
    echo "❌ 错误: Go 编程语言未安装"
    echo "请访问 https://golang.org/dl/ 下载并安装 Go"
    exit 1
fi

echo "✅ 环境检查通过"

# 下载依赖（如果需要）
echo "📦 检查并安装依赖..."
go mod tidy

# 构建程序
echo "🔨 构建程序..."
if go build -o deepseek-proxy .; then
    echo "✅ 构建成功"
else
    echo "❌ 构建失败"
    exit 1
fi

if [ "$TEST_REASONER" = true ]; then
    # 推理测试模式
    echo "🧠 构建推理测试程序..."
    if go build -o test-reasoner test_reasoner.go; then
        echo "✅ 推理测试程序构建成功"
    else
        echo "❌ 推理测试程序构建失败"
        exit 1
    fi
    
    # 启动服务器（后台运行）
    echo "🚀 启动代理服务器（后台模式）..."
    ./deepseek-proxy &
    SERVER_PID=$!
    
    # 等待服务器启动
    echo "⏳ 等待服务器启动..."
    sleep 5
    
    # 运行推理测试
    echo "🧠 运行DeepSeek-Reasoner推理能力测试..."
    go run test_reasoner.go
    
    # 停止服务器
    echo "🛑 停止服务器..."
    kill $SERVER_PID
    
    # 清理
    rm -f deepseek-proxy test-reasoner
    
else
    # 正常启动模式
    echo "🚀 启动服务器..."
    echo "📖 访问 http://localhost:9000 查看服务器信息"
    echo "🧠 支持DeepSeek-Reasoner推理模型"
    echo "🛑 按 Ctrl+C 停止服务器"
    echo ""
    echo "💡 提示：运行 './start.sh test-reasoner' 来测试推理功能"
    echo ""

    # 运行程序
    ./deepseek-proxy

    # 清理
    echo ""
    echo "🧹 清理临时文件..."
    rm -f deepseek-proxy
fi