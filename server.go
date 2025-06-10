package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/net/http2"
)

type ProxyServer struct {
	config     *ProxyConfig
	httpServer *http.Server
	mux        *http.ServeMux
}

func NewProxyServer(config *ProxyConfig) *ProxyServer {
	log.Printf("正在创建代理服务器，端口: %d", config.Port)

	mux := http.NewServeMux()
	proxy := &ProxyServer{
		config: config,
		mux:    mux,
	}

	proxy.setupRoutes()

	// 构建监听地址
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	if config.Host == "" {
		addr = fmt.Sprintf(":%d", config.Port) // 默认localhost
	}

	proxy.httpServer = &http.Server{
		Addr:              addr,
		Handler:           proxy.mux,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	if err := http2.ConfigureServer(proxy.httpServer, &http2.Server{}); err != nil {
		log.Printf("警告：无法启用HTTP/2支持: %v", err)
	}

	log.Printf("✓ 代理服务器创建完成")
	return proxy
}

func (ps *ProxyServer) setupRoutes() {
	log.Printf("正在设置API路由...")

	ps.mux.HandleFunc("/health", ps.handleHealth)
	ps.mux.HandleFunc("/v1/chat/completions", ps.handleChatCompletions)
	ps.mux.HandleFunc("/v1/models", ps.handleModels)
	ps.mux.HandleFunc("/v1/usage", ps.handleUsage)
	ps.mux.HandleFunc("/", ps.handleRoot)

	log.Printf("✓ API路由设置完成")
}

func (ps *ProxyServer) Start() error {
	log.Printf("🚀 启动代理服务器...")
	
	// 显示监听地址
	host := ps.config.Host
	if host == "" {
		host = "localhost"
	}
	log.Printf("📡 监听地址: http://%s:%d", host, ps.config.Port)
	log.Printf("🔧 API端点: http://%s:%d/v1/chat/completions", host, ps.config.Port)
	log.Printf("📋 模型列表: http://%s:%d/v1/models", host, ps.config.Port)
	log.Printf("❤️  健康检查: http://%s:%d/health", host, ps.config.Port)

	return ps.httpServer.ListenAndServe()
}

func (ps *ProxyServer) handleCORS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Length")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
}

func (ps *ProxyServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	ps.handleCORS(w, r)
	if r.Method == "OPTIONS" {
		return
	}

	log.Printf("收到健康检查请求")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	healthInfo := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
		"service":   "deepseek-proxy",
		"uptime":    time.Since(startTime).Seconds(),
	}

	if err := writeJSONResponse(w, healthInfo); err != nil {
		log.Printf("写入健康检查响应失败: %v", err)
	}
}

func (ps *ProxyServer) handleRoot(w http.ResponseWriter, r *http.Request) {
	ps.handleCORS(w, r)
	if r.Method == "OPTIONS" {
		return
	}

	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	log.Printf("收到根路径访问请求")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	host := ps.config.Host
	if host == "" {
		host = "localhost"
	}

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
        <code>http://` + host + `:` + fmt.Sprintf("%d", ps.config.Port) + `/v1</code>
        
        <h2>📚 支持的模型：</h2>
        <ul>` + ps.getSupportedModelsHTML() + `</ul>
    </div>
</body>
</html>`

	w.Write([]byte(welcomeHTML))
}

func (ps *ProxyServer) getSupportedModelsHTML() string {
	models := GetSupportedModels()
	html := ""
	for _, model := range models {
		html += fmt.Sprintf("<li><code>%s</code></li>", model)
	}
	return html
}

var startTime = time.Now()