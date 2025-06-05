# 🚀 DeepSeek API 代理服务器

> **让 Cursor IDE 和任何 OpenAI 兼容应用使用 DeepSeek 强大 AI 模型的智能桥梁**

## 📖 项目概述

### 这个项目解决什么问题？

想象一下这样的场景：你在使用 Cursor IDE 编程，想要利用 DeepSeek 的强大 AI 能力（特别是其出色的推理模型），但是 Cursor 只支持 OpenAI API 格式。这就像你有一台只能播放 CD 的音响，但你手里只有磁带——两者都很棒，但就是不兼容。

这个代理服务器就是那个"格式转换器"，它坐在中间，接收 OpenAI 格式的请求，将其转换为 DeepSeek API 能理解的格式，然后将响应再转换回来。整个过程对于客户端来说是完全透明的。

### 为什么选择 DeepSeek？

DeepSeek 提供了多个优势：

**成本效益**：相比 OpenAI GPT-4，DeepSeek 提供了更有竞争力的价格
**推理能力**：DeepSeek-Reasoner 模型在复杂推理任务上表现出色
**代码专精**：DeepSeek-Coder 专门为编程任务优化
**无地域限制**：在某些地区可能比 OpenAI 更容易访问

## 🎯 核心功能

### 完整的 API 兼容性
我们的代理服务器不是简单的请求转发器，而是一个智能的协议翻译器。它理解两种 API 之间的细微差别，并妥善处理各种边缘情况。

### 流式响应支持
当你在 Cursor 中询问一个复杂问题时，你不希望等待几十秒才看到完整答案。流式响应让你能实时看到 AI 的思考过程，就像在和一个真人对话一样。

### 推理模型集成
DeepSeek-Reasoner 是一个特殊的模型，它不仅给出答案，还会展示完整的推理过程。我们的代理正确处理了这些推理内容，让你能看到 AI 是如何一步步得出结论的。

### 工具调用支持
现代 AI 不仅仅是对话，还能调用外部工具和函数。我们的代理完整支持这些高级功能，让你能在项目中使用更复杂的 AI 工作流。

## 🛠️ 技术架构

### 系统设计理念

这个代理服务器采用了模块化设计，每个组件都有明确的职责：

**配置管理器** (`config.go`)：负责加载和验证所有配置，确保系统有正确的运行参数
**类型系统** (`types.go`)：定义了完整的数据结构，确保请求和响应的类型安全
**服务器核心** (`server.go`)：处理 HTTP 连接、路由和并发管理
**请求处理器** (`handlers.go`)：实现具体的业务逻辑，包括协议转换
**工具库** (`utils.go`)：提供通用功能，如 JSON 处理、错误处理、日志记录

### 请求处理流程

让我们跟踪一个完整的请求是如何被处理的：

1. **接收阶段**：Cursor 发送 OpenAI 格式的请求到我们的代理
2. **验证阶段**：检查 API 密钥和请求格式的有效性  
3. **转换阶段**：将 OpenAI 请求结构转换为 DeepSeek 格式
4. **代理阶段**：向 DeepSeek API 发送转换后的请求
5. **响应阶段**：接收 DeepSeek 的响应并转换回 OpenAI 格式
6. **返回阶段**：将最终响应发送给 Cursor

这个过程中的每一步都经过精心设计，确保信息不丢失，格式完全兼容。

## ⚡ 快速开始

### 第一步：环境准备

在开始之前，确保你的开发环境满足要求：

```bash
# 检查 Go 版本（需要 1.19 或更高版本）
go version

# 如果没有安装 Go，请访问 https://golang.org/dl/ 下载
```

### 第二步：获取项目

```bash
# 克隆项目（如果从 Git 仓库获取）
git clone <repository-url>
cd deepseek-proxy

# 或者如果你已经有了源代码文件，确保它们都在同一个目录中
```

### 第三步：配置设置

```bash
# 复制配置模板
cp .env.example .env

# 编辑配置文件
nano .env  # 或使用你喜欢的编辑器
```

在 `.env` 文件中，你需要设置：

```bash
# 必需配置
DEEPSEEK_API_KEY=sk-your-actual-api-key-here

# 可选配置（有合理的默认值）
PORT=9000
DEEPSEEK_MODEL=deepseek-reasoner
DEEPSEEK_ENDPOINT=https://api.deepseek.com
```

**重要提示**：你需要在 [DeepSeek 平台](https://platform.deepseek.com/) 注册账户并获取 API 密钥。这个密钥是连接 DeepSeek 服务的凭证。

### 第四步：构建和启动

```bash
# 下载依赖
go mod tidy

# 构建项目
go build -o deepseek-proxy .

# 启动服务器
./deepseek-proxy
```

如果一切配置正确，你应该看到类似这样的输出：

```
🚀 启动 DeepSeek API 代理服务器
📡 监听地址: http://localhost:9000
🔧 API端点: http://localhost:9000/v1/chat/completions
```

### 第五步：验证安装

在浏览器中访问 http://localhost:9000，你应该看到一个欢迎页面，确认服务器正在运行。

## 🎮 在 Cursor IDE 中使用

### 配置步骤

1. **打开 Cursor 设置**：按 `Cmd/Ctrl + ,` 进入设置界面
2. **找到模型配置**：查找 "Models" 或 "AI" 相关设置
3. **配置 OpenAI API**：
   - **基础 URL**：`http://localhost:9000/v1`
   - **API 密钥**：你的 DeepSeek API 密钥
   - **模型**：`deepseek-reasoner`（推荐）或 `gpt-4o`

### 为什么这样配置？

**基础 URL**：这告诉 Cursor 将请求发送到你的本地代理，而不是直接发送到 OpenAI
**API 密钥**：虽然 Cursor 认为这是 OpenAI 密钥，但我们的代理会用它来访问 DeepSeek
**模型名称**：我们支持多种模型映射，让你可以使用熟悉的 OpenAI 模型名称

### 使用体验

配置完成后，在 Cursor 的 Composer 中提问时，你实际上是在与 DeepSeek 的 AI 对话。如果你选择了推理模型，你甚至可能看到 AI 的完整思考过程，这在调试复杂逻辑或学习新概念时非常有价值。

## 📡 API 参考

### 聊天完成端点

```http
POST /v1/chat/completions
Content-Type: application/json
Authorization: Bearer YOUR_DEEPSEEK_API_KEY

{
  "model": "deepseek-reasoner",
  "messages": [
    {
      "role": "system", 
      "content": "你是一个有用的编程助手。"
    },
    {
      "role": "user", 
      "content": "解释一下 Go 语言的 goroutine 工作原理。"
    }
  ],
  "stream": true,
  "temperature": 0.7,
  "max_tokens": 2000
}
```

### 支持的模型

我们的代理支持智能模型映射：

| 客户端请求的模型 | 实际使用的 DeepSeek 模型 | 用途说明 |
|-----------------|------------------------|---------|
| `deepseek-reasoner` | `deepseek-reasoner` | 复杂推理任务，显示思考过程 |
| `o1`, `o1-preview` | `deepseek-reasoner` | OpenAI o1 兼容映射 |
| `gpt-4o`, `gpt-4` | `deepseek-chat` | 通用对话和问答 |
| `deepseek-coder` | `deepseek-coder` | 代码生成和调试专用 |
| `gpt-3.5-turbo` | `deepseek-chat` | 快速响应场景 |

### 推理模型的特殊功能

当使用推理模型时，响应可能包含额外的 `reasoning_content` 字段，这包含了 AI 的完整思考过程：

```json
{
  "choices": [{
    "message": {
      "role": "assistant",
      "content": "最终答案在这里...",
      "reasoning_content": "让我一步步分析这个问题：首先..."
    }
  }]
}
```

## 🔧 高级配置

### 环境变量详解

```bash
# 核心配置
DEEPSEEK_API_KEY=sk-xxx    # DeepSeek API 密钥（必需）
PORT=9000                  # 服务器监听端口
DEEPSEEK_ENDPOINT=https://api.deepseek.com  # API 端点

# 模型配置
DEEPSEEK_MODEL=deepseek-reasoner  # 默认模型

# 调试配置
DEBUG=false               # 调试模式开关
LOG_LEVEL=info           # 日志级别：debug, info, warn, error
```

### 命令行选项

```bash
# 显示帮助信息
./deepseek-proxy -help

# 显示版本信息
./deepseek-proxy -version

# 使用调试模式启动
./deepseek-proxy -debug

# 指定不同端口
./deepseek-proxy -port 8080

# 使用自定义配置文件
./deepseek-proxy -config /path/to/custom.env
```

### 性能调优

服务器默认配置已经针对大多数使用场景进行了优化：

**连接超时**：30秒读取/写入超时，防止长时间挂起
**HTTP/2 支持**：启用多路复用，提高并发性能
**连接池**：复用连接，减少延迟
**内存限制**：限制请求大小，防止内存耗尽

## 🧪 测试和验证

### 自动化测试

项目包含完整的测试套件：

```bash
# 运行基础测试
./test.sh

# 测试推理模型专项功能
./start.sh test-reasoner
```

### 手动测试

测试健康检查：
```bash
curl http://localhost:9000/health
```

测试模型列表：
```bash
curl http://localhost:9000/v1/models
```

测试聊天完成：
```bash
curl -X POST http://localhost:9000/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "model": "deepseek-reasoner",
    "messages": [{"role": "user", "content": "Hello!"}],
    "max_tokens": 100
  }'
```

## 🐛 故障排除

### 常见问题和解决方案

**问题：服务器启动失败，提示端口被占用**

解决方案：
```bash
# 检查端口使用情况
netstat -an | grep 9000

# 更改端口配置
echo "PORT=9001" >> .env
```

**问题：API 密钥验证失败**

检查清单：
- 确保 `.env` 文件中的密钥格式正确
- 验证密钥在 DeepSeek 平台上是否有效
- 确认没有多余的空格或换行符

**问题：Cursor 无法连接到代理**

调试步骤：
1. 确认代理服务器正在运行：`curl http://localhost:9000/health`
2. 检查防火墙设置
3. 验证 Cursor 中的基础 URL 配置：`http://localhost:9000/v1`

**问题：响应速度慢**

优化建议：
- 检查网络连接到 DeepSeek API 的延迟
- 考虑使用较小的 `max_tokens` 设置
- 启用调试模式查看详细的请求处理时间

### 调试技巧

启用详细日志：
```bash
DEBUG=true ./deepseek-proxy -debug
```

这会显示：
- 每个请求的详细处理过程
- 模型映射信息
- DeepSeek API 的响应时间
- 错误的完整堆栈跟踪

## 🔒 安全考虑

### API 密钥安全

**最佳实践**：
- 永远不要将 `.env` 文件提交到版本控制系统
- 定期轮换 API 密钥
- 在生产环境中使用环境变量而不是文件

**权限控制**：
```bash
# 设置 .env 文件的适当权限
chmod 600 .env
```

### 网络安全

在生产环境中：
- 考虑使用反向代理（如 Nginx）
- 启用 HTTPS 
- 实施请求速率限制
- 添加 IP 白名单功能

## 📈 生产部署

### Docker 部署

创建 Dockerfile：
```dockerfile
FROM golang:1.19-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o deepseek-proxy .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/deepseek-proxy .
EXPOSE 9000
CMD ["./deepseek-proxy"]
```

构建和运行：
```bash
docker build -t deepseek-proxy .
docker run -p 9000:9000 --env-file .env deepseek-proxy
```

### 系统服务

创建 systemd 服务文件 `/etc/systemd/system/deepseek-proxy.service`：
```ini
[Unit]
Description=DeepSeek API Proxy
After=network.target

[Service]
Type=simple
User=nobody
WorkingDirectory=/opt/deepseek-proxy
ExecStart=/opt/deepseek-proxy/deepseek-proxy
EnvironmentFile=/opt/deepseek-proxy/.env
Restart=always

[Install]
WantedBy=multi-user.target
```

## 🤝 贡献指南

### 开发环境设置

1. Fork 项目仓库
2. 创建功能分支：`git checkout -b feature/amazing-feature`
3. 安装开发依赖：`go mod tidy`
4. 运行测试：`go test ./...`
5. 提交更改：`git commit -m 'Add amazing feature'`
6. 推送分支：`git push origin feature/amazing-feature`
7. 创建 Pull Request

### 代码风格

我们遵循标准的 Go 代码风格：
- 使用 `gofmt` 格式化代码
- 遵循 Go 的命名约定
- 为公共函数添加文档注释
- 保持函数简洁，单一职责

### 添加新功能

如果你想添加对新 AI 服务商的支持：

1. 在 `types.go` 中添加新的请求/响应结构
2. 在 `config.go` 中添加配置选项
3. 在 `handlers.go` 中实现转换逻辑
4. 更新模型映射表
5. 添加相应的测试用例

## 📚 学习资源

### 相关技术文档

- [Go 语言官方文档](https://golang.org/doc/)
- [OpenAI API 参考](https://platform.openai.com/docs/api-reference)
- [DeepSeek API 文档](https://platform.deepseek.com/api-docs)
- [HTTP/2 协议说明](https://http2.github.io/)

### 项目架构深入理解

如果你想深入理解这个项目的设计模式：

**代理模式**：整个项目是代理模式的经典实现，它代表客户端与远程服务交互
**适配器模式**：协议转换部分使用了适配器模式，使不兼容的接口能够协同工作
**工厂模式**：HTTP 客户端的创建使用了工厂模式，确保一致的配置
**策略模式**：不同模型的处理逻辑使用了策略模式

## 📞 支持和社区

### 获取帮助

如果你遇到问题：

1. 首先查看本文档的故障排除部分
2. 检查项目的 Issues 页面，看是否有类似问题
3. 如果问题未解决，创建新的 Issue，请包含：
   - 详细的错误信息
   - 你的配置信息（隐藏敏感数据）
   - 重现问题的步骤
   - 你的操作系统和 Go 版本

### 社区贡献

我们欢迎各种形式的贡献：
- 代码改进和 bug 修复
- 文档完善和翻译
- 功能建议和讨论
- 使用体验分享

---

## 🎉 结语

这个 DeepSeek API 代理服务器不仅仅是一个技术工具，它是连接不同 AI 生态系统的桥梁。通过使用它，你可以在喜爱的开发环境中享受 DeepSeek 强大 AI 模型的能力，无需改变现有的工作流程。

特别是 DeepSeek-Reasoner 模型的推理能力，能让你看到 AI 是如何一步步解决复杂问题的，这对于学习、调试和理解复杂概念都非常有价值。

希望这个项目能帮助你在 AI 辅助编程的道路上走得更远！

**🚀 现在就开始你的 DeepSeek + Cursor 之旅吧！**