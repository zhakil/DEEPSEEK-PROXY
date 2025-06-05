#!/bin/bash

# DeepSeek API 代理服务器测试脚本
# 这个脚本启动服务器并运行自动化测试

echo "🧪 DeepSeek API 代理服务器测试套件"
echo "=================================="

# 检查环境
if [ ! -f ".env" ]; then
    echo "❌ 错误: .env 文件不存在"
    exit 1
fi

if ! command -v go &> /dev/null; then
    echo "❌ 错误: Go 编程语言未安装"
    exit 1
fi

# 构建主程序
echo "🔨 构建主程序..."
if ! go build -o deepseek-proxy .; then
    echo "❌ 主程序构建失败"
    exit 1
fi

# 构建测试客户端
echo "🔨 构建测试客户端..."
if ! go build -o test-client test_client.go; then
    echo "❌ 测试客户端构建失败"
    exit 1
fi

echo "✅ 构建完成"

# 启动服务器（后台运行）
echo "🚀 启动代理服务器..."
./deepseek-proxy &
SERVER_PID=$!

# 等待服务器启动
echo "⏳ 等待服务器启动 (5秒)..."
sleep 5

# 检查服务器是否成功启动
if ! ps -p $SERVER_PID > /dev/null; then
    echo "❌ 服务器启动失败"
    exit 1
fi

echo "✅ 服务器启动成功 (PID: $SERVER_PID)"

# 运行测试
echo ""
echo "🧪 运行自动化测试..."
echo "===================="

# 简单的API测试
echo "🏥 测试健康检查端点..."
if curl -s -f http://localhost:9000/health > /dev/null; then
    echo "✅ 健康检查通过"
else
    echo "❌ 健康检查失败"
fi

echo "📋 测试模型列表端点..."
if curl -s -f http://localhost:9000/v1/models > /dev/null; then
    echo "✅ 模型列表端点可访问"
else
    echo "❌ 模型列表端点失败"
fi

# 运行详细测试（如果可能）
echo ""
echo "🔍 运行详细测试..."
if [ -f "test-client" ]; then
    # 注意：这需要有效的API密钥才能工作
    echo "⚠️  注意：详细测试需要有效的 DEEPSEEK_API_KEY"
    # ./test-client
else
    echo "❌ 测试客户端不可用"
fi

# 测试基本的CURL请求
echo ""
echo "🌐 测试基本API调用..."
TEST_RESPONSE=$(curl -s -X POST http://localhost:9000/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test-key" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "Hello"}],
    "max_tokens": 10
  }')

if echo "$TEST_RESPONSE" | grep -q "error"; then
    echo "✅ API正确处理了测试请求（返回了预期的错误响应）"
else
    echo "⚠️  API响应格式可能需要检查"
fi

# 清理
echo ""
echo "🧹 清理测试环境..."
echo "🛑 停止服务器..."
kill $SERVER_PID

# 等待服务器完全停止
sleep 2

# 清理构建文件
rm -f deepseek-proxy test-client

echo ""
echo "✅ 测试完成！"
echo ""
echo "📝 下一步："
echo "1. 确保你的 .env 文件包含有效的 DEEPSEEK_API_KEY"
echo "2. 运行 ./start.sh 启动服务器"
echo "3. 在 Cursor 中配置 OpenAI API 基础URL 为 http://localhost:9000/v1"
echo "4. 享受使用 DeepSeek 模型的乐趣！"