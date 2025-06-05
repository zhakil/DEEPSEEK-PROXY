# 🚀 DeepSeek API 代理服务器

一个高性能的API代理服务器，让你能够在Cursor IDE和其他OpenAI兼容的应用程序中使用DeepSeek的强大AI模型。

## ✨ 特性

- 🔄 **完整的OpenAI API兼容性** - 无缝替换OpenAI API
- 🌊 **流式响应支持** - 实时查看AI生成过程
- 🛠️ **工具调用支持** - 支持函数调用和工具使用
- 🔒 **安全的API密钥管理** - 环境变量配置，保护你的密钥
- ⚡ **HTTP/2支持** - 现代化的网络协议，更快的连接
- 🌐 **CORS支持** - 支持来自浏览器的跨域请求
- 📊 **健康检查和监控** - 内置状态监控端点
- 🐳 **容器化就绪** - 支持Docker部署

## 🎯 主要用途

这个代理服务器主要为以下场景设计：

1. **Cursor IDE用户** - 在Cursor的Composer中使用DeepSeek模型
2. **开发者工具集成** - 将DeepSeek集成到现有的OpenAI生态系统
3. **成本优化** - 使用DeepSeek的高性价比模型替代OpenAI
4. **本地化部署** - 在自己的基础设施上运行API代理

## 📋 系统要求

- **Go 1.19+** - 用于编译和运行代理服务器
- **DeepSeek API密钥** - 从 [DeepSeek平台](https://platform.deepseek.com/) 获取
- **网络连接** - 用于与DeepSeek API通信

## 🚀 快速开始

### 1️⃣ 克隆或下载项目

```bash
# 如果从GitHub克隆
git clone https://github.com/your-username/deepseek-proxy.git
cd deepseek-proxy

# 或者直接创建项目目录
mkdir deepseek-proxy
cd deepseek-proxy
# 然后复制所有源代码文件到此目录
```

### 2️⃣ 安装依赖

```bash
# 初始化Go模块（如果需要）
go mod init deepseek-proxy

# 安装依赖
go mod tidy
```

### 3️⃣ 配置环境

```bash
# 复制环境配置模板
cp .env.example .env

# 编辑 .env 文件，添加你的API密钥
nano .env
```

在 `.env` 文件中设置：

```bash
DEEPSEEK_API_KEY=sk-your-deepseek-api-key-here
PORT=9000
DEEPSEEK_MODEL=deepseek-chat
DEEPSEEK_ENDPOINT=https://api.deepseek.com
```

### 4️⃣ 启动服务器

```bash
# 使用启动脚本（推荐）
./start.sh

# 或者直接运行
go run .

# 或者构建后运行
go build -o deepseek-proxy .
./deepseek-proxy
```

### 5️⃣ 验证安装

访问 http://localhost:9000 查看服务器状态，或运行测试：

```bash
./test.sh
```

## 🎮 在Cursor中使用

1. 打开Cursor IDE
2. 进入设置 (Cmd/Ctrl + ,)
3. 找到"模型"或"AI"设置
4. 设置OpenAI API配置：
   - **基础URL**: `http://localhost:9000/v1`
   - **API密钥**: 你的DeepSeek API密钥
   - **模型**: `gpt-4o` 或其他支持的模型

现在你可以在Cursor的Composer中享受DeepSeek模型的强大功能！

## 📡 API端点

### 聊天完成
```http
POST /v1/chat/completions
Content-Type: application/json
Authorization: Bearer YOUR_DEEPSEEK_API_KEY

{
  "model": "gpt-4o",
  "messages": [
    {"role": "user", "content": "Hello, world!"}
  ],
  "temperature": 0.7,
  "max_tokens": 1000
}
```

### 模型列表
```http
GET /v1/models
Authorization: Bearer YOUR_DEEPSEEK_API_KEY
```

### 健康检查
```http
GET /health
```

## 🛠️ 支持的模型

| OpenAI模型名 | 映射到DeepSeek模型 | 说明 |
|-------------|------------------|------|
| `gpt-4o` | `deepseek-chat` | 最新的高性能模型 |
| `gpt-4` | `deepseek-chat` | 通用聊天模型 |
| `gpt-3.5-turbo` | `deepseek-chat` | 快速响应模型 |
| `deepseek-chat` | `deepseek-chat` | DeepSeek原生模型 |
| `deepseek-coder` | `deepseek-coder` | 代码专用模型 |

## ⚙️ 高级配置

### 命令行参数

```bash
./deepseek-proxy -help                    # 显示帮助
./deepseek-proxy -version                 # 显示版本
./deepseek-proxy -port 8080              # 指定端口
./deepseek-proxy -debug                  # 启用调试模式
./deepseek-proxy -config /path/to/.env   # 指定配置文件
```

### 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `DEEPSEEK_API_KEY` | (必需) | DeepSeek API密钥 |
| `PORT` | `9000` | 服务器监听端口 |
| `DEEPSEEK_MODEL` | `deepseek-chat` | 默认使用的模型 |
| `DEEPSEEK_ENDPOINT` | `https://api.deepseek.com` | API端点地址 |
| `DEBUG` | `false` | 调试模式开关 |

## 🔧 开发和扩展

### 项目结构

```
deepseek-proxy/
├── main.go          # 程序入口点
├── types.go         # 数据类型定义
├── config.go        # 配置管理
├── server.go        # HTTP服务器
├── handlers.go      # 请求处理器
├── utils.go         # 工具函数
├── test_client.go   # 测试客户端
├── .env.example     # 环境配置模板
├── start.sh         # 启动脚本
├── test.sh          # 测试脚本
└── README.md        # 项目文档
```

### 添加新的AI服务商

要支持新的AI服务商（如Claude、Gemini等），你需要：

1. 在 `types.go` 中添加新的请求/响应结构
2. 在 `config.go` 中添加服务商配置
3. 在 `handlers.go` 中添加转换逻辑
4. 更新模型映射表

### 测试

运行完整的测试套件：

```bash
# 运行自动化测试
./test.sh

# 手动测试特定功能
curl -X POST http://localhost:9000/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{"model":"gpt-4o","messages":[{"role":"user","content":"Hello"}]}'
```

## 🐛 故障排除

### 常见问题

**Q: 服务器启动失败，提示"端口已被占用"**
A: 更改 `.env` 文件中的 `PORT` 设置，或使用 `-port` 参数指定不同端口。

**Q: 收到"无效的API密钥"错误**
A: 检查 `.env` 文件中的 `DEEPSEEK_API_KEY` 是否正确设置，确保没有多余的空格。

**Q: Cursor无法连接到代理**
A: 确保代理服务器正在运行，检查防火墙设置，验证基础URL配置。

**Q: 响应速度慢**
A: 这可能是网络延迟导致的，代理服务器只是转发请求，响应速度主要取决于DeepSeek API的响应时间。

### 调试模式

启用调试模式获取详细日志：

```bash
./deepseek-proxy -debug
```

或在 `.env` 文件中设置：

```bash
DEBUG=true
```

## 📜 许可证

本项目采用MIT许可证。详见 [LICENSE](LICENSE) 文件。

## 🤝 贡献

欢迎贡献代码！请遵循以下步骤：

1. Fork 本仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 开启 Pull Request

## 🙏 致谢

- [DeepSeek](https://www.deepseek.com/) - 提供强大的AI模型
- [Cursor](https://cursor.sh/) - 激发了这个项目的创建
- Go社区 - 提供了优秀的工具和库

## 📞 支持

如果你遇到问题或有建议，请：

1. 查看 [故障排除](#-故障排除) 部分
2. 搜索已有的 [Issues](https://github.com/your-username/deepseek-proxy/issues)
3. 创建新的 Issue 详细描述你的问题

---

**🎉 享受使用DeepSeek模型的乐趣吧！**