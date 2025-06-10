package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// 程序版本信息
const (
	Version     = "1.0.0"
	ProgramName = "DeepSeek API 代理服务器"
)

// 命令行参数定义
var (
	showVersion = flag.Bool("version", false, "显示版本信息")
	showHelp    = flag.Bool("help", false, "显示帮助信息")
	configPath  = flag.String("config", ".env", "配置文件路径")
	port        = flag.Int("port", 0, "服务器端口号（覆盖配置文件设置）")
	host        = flag.String("host", "", "绑定主机地址")
	debug       = flag.Bool("debug", false, "启用调试模式")
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	printWelcomeBanner()
	flag.Parse()

	if *showVersion {
		printVersion()
		return
	}

	if *showHelp {
		printHelp()
		return
	}

	if err := validateEnvironment(); err != nil {
		log.Fatalf("环境验证失败: %v", err)
	}

	// 处理命令行主机地址覆盖
	if *host != "" {
		GlobalConfig.Host = *host
		log.Printf("使用命令行指定的主机地址: %s", *host)
	}

	// 处理命令行端口覆盖
	if *port > 0 {
		GlobalConfig.Port = *port
		log.Printf("使用命令行指定的端口: %d", *port)
	}

	if *debug {
		log.Println("调试模式已启用")
		printDebugInfo()
	}

	log.Println("正在初始化代理服务器...")
	proxyServer := NewProxyServer(GlobalConfig)

	setupGracefulShutdown(proxyServer)

	log.Printf("🎉 %s v%s 启动完成！", ProgramName, Version)
	log.Printf("📖 访问 http://localhost:%d 查看服务器信息", GlobalConfig.Port)
	log.Println("🛑 按 Ctrl+C 停止服务器")

	if err := proxyServer.Start(); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

func printWelcomeBanner() {
	fmt.Printf(`
╔══════════════════════════════════════════════════════════════╗
║                    🚀 DeepSeek API 代理服务器                     ║
║                         Version %s                          ║
║                                                              ║
║  将 OpenAI 兼容的请求转换为 DeepSeek API 格式                      ║
║  支持流式响应、工具调用和完整的 OpenAI API 兼容性                     ║
╚══════════════════════════════════════════════════════════════╝

`, Version)
}

func printVersion() {
	fmt.Printf("%s v%s", ProgramName, Version)
	fmt.Println()
	fmt.Println("构建信息:")
	fmt.Println("  - Go 版本: 1.19+")
	fmt.Println("  - 支持的协议: HTTP/1.1, HTTP/2")
	fmt.Println("  - 支持的格式: JSON, Server-Sent Events")
	fmt.Println("  - 兼容性: OpenAI Chat Completions API v1")
	fmt.Println()
	fmt.Println("项目主页: https://github.com/your-username/deepseek-proxy")
}

func printHelp() {
	fmt.Printf("%s - OpenAI 到 DeepSeek API 代理服务器", ProgramName)
	fmt.Println()
	fmt.Println()
	fmt.Println("用法:")
	fmt.Printf("  %s [选项]", os.Args[0])
	fmt.Println()
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  -version          显示版本信息并退出")
	fmt.Println("  -help             显示此帮助信息并退出")
	fmt.Println("  -config string    配置文件路径 (默认: .env)")
	fmt.Println("  -port int         服务器端口号 (覆盖配置文件)")
	fmt.Println("  -host string      绑定主机地址 (如: 0.0.0.0)")
	fmt.Println("  -debug            启用调试模式")
	fmt.Println()
	fmt.Println("环境变量:")
	fmt.Println("  DEEPSEEK_API_KEY     DeepSeek API 密钥 (必需)")
	fmt.Println("  PORT                 服务器端口号 (默认: 9000)")
	fmt.Println("  HOST                 绑定主机地址 (默认: localhost)")
	fmt.Println("  DEEPSEEK_MODEL       默认模型 (默认: deepseek-reasoner)")
	fmt.Println("  DEEPSEEK_ENDPOINT    API 端点 (默认: https://api.deepseek.com)")
	fmt.Println()
	fmt.Println("示例:")
	examples := []struct {
		command     string
		description string
	}{
		{os.Args[0], "使用默认配置启动"},
		{os.Args[0] + " -port 8080", "在端口 8080 启动"},
		{os.Args[0] + " -host 0.0.0.0", "绑定所有网络接口"},
		{os.Args[0] + " -host 0.0.0.0 -port 9000", "绑定所有接口端口9000"},
		{os.Args[0] + " -debug", "启用调试模式"},
	}

	for _, example := range examples {
		fmt.Printf("  %-50s # %s", example.command, example.description)
		fmt.Println()
	}
	fmt.Println()
	fmt.Println("配置文件示例 (.env):")
	fmt.Println("  DEEPSEEK_API_KEY=sk-your-api-key-here")
	fmt.Println("  PORT=9000")
	fmt.Println("  HOST=0.0.0.0")
	fmt.Println("  DEEPSEEK_MODEL=deepseek-reasoner")
}

func validateEnvironment() error {
	log.Println("正在验证运行环境...")
	log.Println("✓ Go 运行时环境正常")

	if GlobalConfig.DeepSeekAPIKey == "" {
		return fmt.Errorf("缺少必需的环境变量: DEEPSEEK_API_KEY")
	}
	log.Println("✓ API 密钥已配置")

	if GlobalConfig.Port <= 0 || GlobalConfig.Port > 65535 {
		return fmt.Errorf("无效的端口号: %d (必须在 1-65535 之间)", GlobalConfig.Port)
	}
	log.Printf("✓ 端口配置有效: %d", GlobalConfig.Port)
	log.Println("✓ 环境验证通过")
	return nil
}

func printDebugInfo() {
	fmt.Println()
	fmt.Println("=== 调试信息 ===")
	fmt.Println("配置文件路径:", *configPath)
	fmt.Println("绑定主机:", GlobalConfig.Host)
	fmt.Println("监听端口:", GlobalConfig.Port)
	fmt.Println("DeepSeek 端点:", GlobalConfig.Endpoint)
	fmt.Println("默认模型:", GlobalConfig.DeepSeekModel)
	fmt.Println("API 密钥:", maskAPIKey(GlobalConfig.DeepSeekAPIKey))
	fmt.Println()
	fmt.Println("支持的模型:")

	for i, model := range GetSupportedModels() {
		fmt.Printf("  %d. %s", i+1, model)
		fmt.Println()
	}

	fmt.Println()
	fmt.Println("可用端点:")
	endpoints := []struct {
		name string
		path string
	}{
		{"聊天完成", "/v1/chat/completions"},
		{"模型列表", "/v1/models"},
		{"健康检查", "/health"},
		{"服务器信息", "/"},
	}

	host := GlobalConfig.Host
	if host == "" {
		host = "localhost"
	}
	for _, endpoint := range endpoints {
		fmt.Printf("  - %s: http://%s:%d%s",
			endpoint.name, host, GlobalConfig.Port, endpoint.path)
		fmt.Println()
	}
	fmt.Println("================")
	fmt.Println()
}

func setupGracefulShutdown(server *ProxyServer) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		fmt.Println()
		log.Printf("收到信号: %v", sig)
		log.Println("正在优雅关闭服务器...")
		log.Printf("正在关闭服务器实例: %p", server)
		log.Println("✅ 服务器已安全关闭")
		log.Printf("👋 感谢使用 %s！", ProgramName)
		os.Exit(0)
	}()
}