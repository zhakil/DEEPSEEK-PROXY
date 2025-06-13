# 🚀 DeepSeek API 代理服务器

> **让任何 OpenAI 兼容应用直接使用 DeepSeek 强大AI模型的高效解决方案**

## 核心价值

**解决的实际问题：** Cursor IDE、ChatGPT客户端等工具只支持OpenAI API格式，无法直接使用DeepSeek的高性价比AI模型。

**技术方案：** 实时协议转换代理，零配置集成，完全透明的API兼容层。

**经济效益：** DeepSeek API成本比OpenAI低60-80%，性能相当甚至更优。

## 快速开始

### 1. 环境准备

```bash
# 确保Go 1.19+已安装
go version

# 获取项目代码
git clone <repository-url>
cd deepseek-proxy
```

### 2. 配置设置

```bash
# 创建配置文件
cp .env.example .env

# 编辑配置 - 只需设置API密钥
DEEPSEEK_API_KEY=sk-your-deepseek-api-key-here
PORT=9000
HOST=0.0.0.0
# PROXY_URL=http://127.0.0.1:10808 # Optional. The URL of the proxy server to use for outgoing requests to the DeepSeek API. e.g., http://127.0.0.1:10808 or socks5://127.0.0.1:10809 (Note: Go's default HTTP client supports HTTP/HTTPS and SOCKS5 proxies).
```

### 3. 启动服务

#### 环境变量

除了`.env`文件，所有配置也可以通过环境变量设置：

- `DEEPSEEK_API_KEY`: 必需。您的 DeepSeek API 密钥。
- `PORT`: 可选。代理服务器监听的端口，默认为 `9000`。
- `HOST`: 可选。代理服务器绑定的主机地址，默认为 `""` (空字符串，表示 `localhost`)。设置为 `0.0.0.0` 可以监听所有网络接口。
- `PROXY_URL`: 可选。用于向 DeepSeek API 发出请求的代理服务器的 URL。
  - 示例: `PROXY_URL=http://127.0.0.1:10808` 或 `PROXY_URL=socks5://127.0.0.1:10809`
  - 注意: Go 的默认 HTTP 客户端支持 HTTP/HTTPS 和 SOCKS5 代理。
- `DEEPSEEK_MODEL`: 可选。默认使用的 DeepSeek 模型，默认为 `deepseek-reasoner`。
- `DEEPSEEK_ENDPOINT`: 可选。DeepSeek API 的端点URL，默认为 `https://api.deepseek.com`。

### 3. 启动服务

```bash
# 构建并启动
go build . && ./deepseek-proxy -host 0.0.0.0 -port 9000

# 验证运行
curl http://localhost:9000/health
```

### 4. 客户端配置

**Cursor IDE：**
- API Base URL: `http://0.0.0.0:9000/v1`
- API Key: `你的DeepSeek密钥`
- Model: `gpt-4o`

**OpenAI SDK：**
```python
import openai
client = openai.OpenAI(
    base_url="http://0.0.0.0:9000/v1",
    api_key="sk-your-deepseek-key"
)
```

## 核心特性

### 🔄 协议转换
- **OpenAI → DeepSeek** 实时格式转换
- **零延迟** 请求处理
- **完整兼容** Chat Completions API

### 🧠 DeepSeek-Reasoner 集成
- **推理过程可视化** - 查看AI思考步骤
- **复杂问题解决** - 数学、逻辑、编程推理
- **透明化AI决策** - 理解每个答案的产生过程

### 🌊 流式响应支持
- **实时输出** - 字符逐步显示
- **低延迟体验** - 响应即时可见
- **中断保护** - 连接异常自动恢复

### 🔧 网络兼容性
- **私网绕过** - 解决Cursor等客户端的网络限制
- **灵活绑定** - 支持localhost/0.0.0.0/自定义IP
- **跨平台部署** - Windows/Linux/macOS

## 命令行工具

### 基础用法
```bash
./deepseek-proxy                           # 默认配置启动
./deepseek-proxy -port 8080                # 指定端口
./deepseek-proxy -host 0.0.0.0             # 绑定所有接口
./deepseek-proxy -host 0.0.0.0 -port 9000  # 完整配置
./deepseek-proxy -debug                     # 调试模式
```

### 测试工具
```bash
# 健康检查
curl http://localhost:9000/health

# API测试
curl -X POST http://localhost:9000/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-your-key" \
  -d '{"model":"gpt-4o","messages":[{"role":"user","content":"test"}]}'

# 自动化测试
./test.sh
```

## 模型映射策略

### 统一映射 - 专注推理能力
所有模型请求统一映射到 `deepseek-reasoner`，获得最佳推理性能：

| 客户端请求 | 实际调用 | 优势 |
|-----------|----------|------|
| `gpt-4o` | `deepseek-reasoner` | 显示完整推理过程 |
| `gpt-4` | `deepseek-reasoner` | 复杂问题分析能力 |
| `o3/o3-mini` | `deepseek-reasoner` | 直接对标推理模型 |
| `claude-*` | `deepseek-reasoner` | 跨平台兼容 |

### 推理模型特殊功能
```json
{
  "choices": [{
    "message": {
      "content": "最终答案...",
      "reasoning_content": "详细的推理过程：第一步...第二步..."
    }
  }]
}
```

## 实际应用场景

### Cursor IDE 增强编程
```bash
# 启动代理
./deepseek-proxy -host 0.0.0.0

# Cursor配置后直接使用
# 获得代码推理、bug分析、算法优化等能力
```

### API开发调试
```bash
# 替换OpenAI调用，无需修改代码
# 原：openai.api_base = "https://api.openai.com/v1"
# 新：openai.api_base = "http://localhost:9000/v1"
```

### 批量处理任务
```bash
# 利用DeepSeek性价比优势处理大量请求
# 成本降低60-80%，性能保持同等水平
```

## 故障排除

### 网络连接问题
```bash
# 403 Private Network Error
./deepseek-proxy -host 0.0.0.0  # 绑定公网接口

# 端口冲突
netstat -an | findstr :9000     # 检查端口占用
./deepseek-proxy -port 9001     # 使用其他端口
```

### API密钥验证
```bash
# 验证密钥有效性
curl -H "Authorization: Bearer sk-your-key" \
     https://api.deepseek.com/v1/models

# 检查配置加载
./deepseek-proxy -debug
```

### Cursor集成问题
1. **Base URL必须是HTTP（非HTTPS）**
2. **API Key使用DeepSeek密钥**
3. **模型选择gpt-4o即可**
4. **重启Cursor应用生效配置**

## 性能优化

### 服务器配置
```bash
# 生产环境推荐配置
HOST=0.0.0.0
PORT=9000
DEEPSEEK_MODEL=deepseek-reasoner
LOG_LEVEL=warn
```

### 并发处理
- **HTTP/2多路复用** - 单连接处理多请求
- **连接池优化** - 复用TCP连接
- **智能超时** - 30秒读写，2分钟空闲

### 监控指标
```bash
# 实时监控
curl http://localhost:9000/v1/usage

# 健康状态
curl http://localhost:9000/health
```

## 生产部署

### Docker 部署
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
CMD ["./deepseek-proxy", "-host", "0.0.0.0"]
```

### 系统服务
```bash
# systemd服务
sudo cp deepseek-proxy.service /etc/systemd/system/
sudo systemctl enable deepseek-proxy
sudo systemctl start deepseek-proxy
```

## 技术架构

### 核心组件
- **协议转换器** - OpenAI ↔ DeepSeek格式互转
- **HTTP代理** - 高性能请求转发
- **流式处理** - Server-Sent Events支持
- **错误处理** - 完善的异常恢复机制

### 设计原则
1. **零配置** - 开箱即用
2. **高性能** - 低延迟转发
3. **强兼容** - 完整API支持
4. **易扩展** - 模块化架构

## 成本对比

| 服务商 | 模型 | 输入价格/1M tokens | 输出价格/1M tokens | 推理能力 |
|--------|------|------------------|------------------|----------|
| OpenAI | GPT-4o | $5.00 | $15.00 | ⭐⭐⭐ |
| OpenAI | o1-preview | $15.00 | $60.00 | ⭐⭐⭐⭐⭐ |
| DeepSeek | Reasoner | $0.14 | $0.28 | ⭐⭐⭐⭐⭐ |

**节省成本：95%+**，**推理能力：相当或更优**

## 开发指南

### 项目结构
```
deepseek-proxy/
├── main.go          # 程序入口
├── config.go        # 配置管理
├── server.go        # HTTP服务器
├── handlers.go      # 请求处理
├── types.go         # 数据结构
├── utils.go         # 工具函数
└── .env.example     # 配置模板
```

### 贡献代码
1. Fork项目
2. 创建功能分支
3. 提交Pull Request
4. 通过代码审查

## 许可证

MIT License - 自由使用、修改、分发

## 支持

- **问题反馈**: [GitHub Issues](github.com/your-username/deepseek-proxy/issues)
- **功能建议**: 创建Feature Request
- **技术讨论**: [Discussions](github.com/your-username/deepseek-proxy/discussions)

---

**让AI编程更高效，让成本控制更精准。** 🚀