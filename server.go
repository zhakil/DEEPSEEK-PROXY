package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/net/http2"
)

// ProxyServer 代理服务器的主要结构体
// 这是我们整个代理系统的核心，就像是一个智能的交通指挥官
type ProxyServer struct {
	config     *ProxyConfig   // 服务器配置信息
	httpServer *http.Server   // 底层的HTTP服务器
	mux        *http.ServeMux // 请求路由器，决定不同的请求去哪里处理
}

// NewProxyServer 创建一个新的代理服务器实例
// 这个函数就像是一个工厂，负责组装我们代理服务器的所有组件
func NewProxyServer(config *ProxyConfig) *ProxyServer {
	log.Printf("正在创建代理服务器，端口: %d", config.Port)

	// 创建路由器
	mux := http.NewServeMux()

	// 创建代理服务器实例
	proxy := &ProxyServer{
		config: config,
		mux:    mux,
	}

	// 设置路由规则
	// 这些路由就像是道路标志，告诉不同的请求应该去哪里
	proxy.setupRoutes()

	// 创建HTTP服务器，配置超时和其他参数
	proxy.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: proxy.mux,

		// 超时配置很重要，防止恶意客户端占用服务器资源
		ReadTimeout:       30 * time.Second,  // 读取请求的最大时间
		WriteTimeout:      30 * time.Second,  // 写入响应的最大时间
		ReadHeaderTimeout: 10 * time.Second,  // 读取请求头的最大时间
		IdleTimeout:       120 * time.Second, // 保持连接的最大空闲时间

		// 限制请求大小，防止过大的请求导致内存问题
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// 启用HTTP/2支持，这能提供更好的性能
	// HTTP/2支持多路复用，就像高速公路的多车道一样
	if err := http2.ConfigureServer(proxy.httpServer, &http2.Server{}); err != nil {
		log.Printf("警告：无法启用HTTP/2支持: %v", err)
	}

	log.Printf("✓ 代理服务器创建完成")
	return proxy
}

// setupRoutes 配置所有的请求路由
// 这个方法定义了我们的代理服务器可以处理哪些类型的请求
func (ps *ProxyServer) setupRoutes() {
	log.Printf("正在设置API路由...")
	
	// 健康检查端点，让外部系统可以检查服务器是否正常运行
	ps.mux.HandleFunc("/health", ps.handleHealth)
	
	// OpenAI兼容的聊天完成端点，这是最重要的端点
	ps.mux.HandleFunc("/v1/chat/completions", ps.handleChatCompletions)
	
	// 模型列表端点，返回支持的模型列表
	ps.mux.HandleFunc("/v1/models", ps.handleModels)
	
	// 根路径处理（包含CORS和欢迎页面逻辑）
	ps.mux.HandleFunc("/", ps.handleRoot)
	
	log.Printf("✓ API路由设置完成")
}

// Start 启动代理服务器
// 这个方法让我们的服务器开始监听和处理请求
func (ps *ProxyServer) Start() error {
	log.Printf("🚀 启动代理服务器...")
	log.Printf("📡 监听地址: http://localhost:%d", ps.config.Port)
	log.Printf("🔧 API端点: http://localhost:%d/v1/chat/completions", ps.config.Port)
	log.Printf("📋 模型列表: http://localhost:%d/v1/models", ps.config.Port)
	log.Printf("❤️  健康检查: http://localhost:%d/health", ps.config.Port)

	// 开始监听请求，这是一个阻塞操作
	return ps.httpServer.ListenAndServe()
}

// handleCORS 处理跨域资源共享（CORS）
// 这个处理器确保我们的API可以被来自不同域名的网页应用调用
func (ps *ProxyServer) handleCORS(w http.ResponseWriter, r *http.Request) {
	// 设置CORS头部，允许跨域访问
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Length")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	// 如果是OPTIONS请求（CORS预检），直接返回成功
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
}

// handleHealth 健康检查处理器
// 这个端点让运维人员和监控系统可以检查服务器状态
func (ps *ProxyServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	// 设置CORS头部
	ps.handleCORS(w, r)

	// 如果是OPTIONS请求，已经在handleCORS中处理了
	if r.Method == "OPTIONS" {
		return
	}

	log.Printf("收到健康检查请求")

	// 返回服务器状态信息
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// 创建健康状态响应
	healthInfo := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
		"service":   "deepseek-proxy",
		"uptime":    time.Since(startTime).Seconds(),
	}

	// 将健康信息写入响应
	if err := writeJSONResponse(w, healthInfo); err != nil {
		log.Printf("写入健康检查响应失败: %v", err)
	}
}

// handleRoot 根路径处理器
// 当有人访问我们的根URL时，显示欢迎信息和使用说明
func (ps *ProxyServer) handleRoot(w http.ResponseWriter, r *http.Request) {
	// 设置CORS头部
	ps.handleCORS(w, r)

	// 如果是OPTIONS请求，已经在handleCORS中处理了
	if r.Method == "OPTIONS" {
		return
	}

	// 只处理根路径的GET请求
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	log.Printf("收到根路径访问请求")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	// 返回友好的HTML欢迎页面
	welcomeHTML := `
<!DOCTYPE html>
<html>
<head>
    <title>DeepSeek API 代理服务器</title>
    <meta charset="utf-8">
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #f5f5f5; }
        .container { background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; border-bottom: 2px solid #007bff; padding-bottom: 10px; }
        .endpoint { background: #f8f9fa; padding: 15px; border-radius: 4px; margin: 10px 0; }
        .status { color: #28a745; font-weight: bold; }
        code { background: #e9ecef; padding: 2px 4px; border-radius: 3px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 DeepSeek API 代理服务器</h1>
        <p class="status">✅ 服务器运行正常</p>
        
        <h2>📡 可用端点：</h2>
        <div class="endpoint">
            <strong>聊天完成：</strong><br>
            <code>POST /v1/chat/completions</code><br>
            与OpenAI ChatGPT API完全兼容
        </div>
        
        <div class="endpoint">
            <strong>模型列表：</strong><br>
            <code>GET /v1/models</code><br>
            获取支持的AI模型列表
        </div>
        
        <div class="endpoint">
            <strong>健康检查：</strong><br>
            <code>GET /health</code><br>
            检查服务器运行状态
        </div>
        
        <h2>🔧 使用方法：</h2>
        <p>将你的OpenAI客户端基础URL设置为：</p>
        <code>http://localhost:` + fmt.Sprintf("%d", ps.config.Port) + `/v1</code>
        
        <h2>📚 支持的模型：</h2>
        <ul>` + ps.getSupportedModelsHTML() + `</ul>
    </div>
</body>
</html>`

	w.Write([]byte(welcomeHTML))
}

// getSupportedModelsHTML 获取支持模型的HTML列表
func (ps *ProxyServer) getSupportedModelsHTML() string {
	models := GetSupportedModels()
	html := ""
	for _, model := range models {
		html += fmt.Sprintf("<li><code>%s</code></li>", model)
	}
	return html
}

// 记录服务器启动时间，用于计算运行时长
var startTime = time.Now()
