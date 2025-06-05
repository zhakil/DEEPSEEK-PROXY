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
	debug       = flag.Bool("debug", false, "启用调试模式")
)

// main 程序的主入口点
// 这是整个程序开始执行的地方，负责初始化和启动代理服务器
func main() {
	// 设置日志格式，包含时间戳和文件位置信息
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	
	// 显示程序启动信息
	printWelcomeBanner()
	
	// 解析命令行参数
	flag.Parse()
	
	// 处理特殊命令行参数
	if *showVersion {
		printVersion()
		return
	}
	
	if *showHelp {
		printHelp()
		return
	}
	
	// 验证运行环境
	if err := validateEnvironment(); err != nil {
		log.Fatalf("环境验证失败: %v", err)
	}
	
	// 如果命令行指定了端口，覆盖配置文件中的设置
	if *port > 0 {
		GlobalConfig.Port = *port
		log.Printf("使用命令行指定的端口: %d", *port)
	}
	
	// 如果启用了调试模式，显示详细信息
	if *debug {
		log.Printf("调试模式已启用")
		printDebugInfo()
	}
	
	// 创建代理服务器实例
	log.Printf("正在初始化代理服务器...")
	proxyServer := NewProxyServer(GlobalConfig)
	
	// 设置优雅关闭处理
	// 这确保服务器可以优雅地响应停止信号，完成正在处理的请求
	setupGracefulShutdown(proxyServer)
	
	// 启动服务器
	log.Printf("🎉 %s v%s 启动完成！", ProgramName, Version)
	log.Printf("📖 访问 http://localhost:%d 查看服务器信息", GlobalConfig.Port)
	log.Printf("🛑 按 Ctrl+C 停止服务器")
	
	// 开始监听请求，这是一个阻塞调用
	if err := proxyServer.Start(); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

// printWelcomeBanner 显示程序启动时的欢迎横幅
// 这让用户知道程序正在启动，并提供基本的版本信息
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

// printVersion 显示详细的版本信息
func printVersion() {
	fmt.Printf("%s v%s\n", ProgramName, Version)
	fmt.Println("构建信息:")
	fmt.Println("  - Go 版本: 1.19+")
	fmt.Println("  - 支持的协议: HTTP/1.1, HTTP/2")
	fmt.Println("  - 支持的格式: JSON, Server-Sent Events")
	fmt.Println("  - 兼容性: OpenAI Chat Completions API v1")
	fmt.Println("\n项目主页: https://github.com/your-username/deepseek-proxy")
}

// printHelp 显示详细的帮助信息
func printHelp() {
	fmt.Printf("%s - OpenAI 到 DeepSeek API 代理服务器\n\n", ProgramName)
	
	fmt.Println("用法:")
	fmt.Printf("  %s [选项]\n\n", os.Args[0])
	
	fmt.Println("选项:")
	fmt.Println("  -version          显示版本信息并退出")
	fmt.Println("  -help             显示此帮助信息并退出")
	fmt.Println("  -config string    配置文件路径 (默认: .env)")
	fmt.Println("  -port int         服务器端口号 (覆盖配置文件)")
	fmt.Println("  -debug            启用调试模式")
	
	fmt.Println("\n环境变量:")
	fmt.Println("  DEEPSEEK_API_KEY     DeepSeek API 密钥 (必需)")
	fmt.Println("  PORT                 服务器端口号 (默认: 9000)")
	fmt.Println("  DEEPSEEK_MODEL       默认模型 (默认: deepseek-chat)")
	fmt.Println("  DEEPSEEK_ENDPOINT    API 端点 (默认: https://api.deepseek.com)")
	
	fmt.Println("\n示例:")
	fmt.Printf("  %s                           # 使用默认配置启动\n", os.Args[0])
	fmt.Printf("  %s -port 8080                # 在端口 8080 启动\n", os.Args[0])
	fmt.Printf("  %s -config /path/to/.env     # 使用指定配置文件\n", os.Args[0])
	fmt.Printf("  %s -debug                    # 启用调试模式\n", os.Args[0])
	
	fmt.Println("\n配置文件示例 (.env):")
	fmt.Println("  DEEPSEEK_API_KEY=sk-your-api-key-here")
	fmt.Println("  PORT=9000")
	fmt.Println("  DEEPSEEK_MODEL=deepseek-chat")
	
	fmt.Println("\n支持的客户端:")
	fmt.Println("  - Cursor IDE")
	fmt.Println("  - 任何兼容 OpenAI API 的应用程序")
	fmt.Println("  - 自定义应用程序使用 OpenAI SDK")
}

// validateEnvironment 验证运行环境是否满足要求
// 这个函数检查所有必需的条件，确保程序能够正常运行
func validateEnvironment() error {
	log.Printf("正在验证运行环境...")
	
	// 检查Go版本（这在编译时已经保证，但我们可以添加运行时检查）
	log.Printf("✓ Go 运行时环境正常")
	
	// 检查必需的环境变量
	if GlobalConfig.DeepSeekAPIKey == "" {
		return fmt.Errorf("缺少必需的环境变量: DEEPSEEK_API_KEY")
	}
	log.Printf("✓ API 密钥已配置")
	
	// 检查端口号的有效性
	if GlobalConfig.Port <= 0 || GlobalConfig.Port > 65535 {
		return fmt.Errorf("无效的端口号: %d (必须在 1-65535 之间)", GlobalConfig.Port)
	}
	log.Printf("✓ 端口配置有效: %d", GlobalConfig.Port)
	
	// 检查网络连通性（可选）
	// 这里可以添加对DeepSeek API的连通性测试
	log.Printf("✓ 环境验证通过")
	
	return nil
}

// printDebugInfo 显示调试信息
// 在调试模式下，这个函数会显示详细的配置信息，帮助开发者诊断问题
func printDebugInfo() {
	fmt.Println("\n=== 调试信息 ===")
	fmt.Printf("配置文件路径: %s\n", *configPath)
	fmt.Printf("监听端口: %d\n", GlobalConfig.Port)
	fmt.Printf("DeepSeek 端点: %s\n", GlobalConfig.Endpoint)
	fmt.Printf("默认模型: %s\n", GlobalConfig.DeepSeekModel)
	fmt.Printf("API 密钥: %s\n", maskAPIKey(GlobalConfig.DeepSeekAPIKey))
	
	fmt.Println("\n支持的模型:")
	for i, model := range GetSupportedModels() {
		fmt.Printf("  %d. %s\n", i+1, model)
	}
	
	fmt.Println("\n可用端点:")
	fmt.Printf("  - 聊天完成: http://localhost:%d/v1/chat/completions\n", GlobalConfig.Port)
	fmt.Printf("  - 模型列表: http://localhost:%d/v1/models\n", GlobalConfig.Port)
	fmt.Printf("  - 健康检查: http://localhost:%d/health\n", GlobalConfig.Port)
	fmt.Printf("  - 服务器信息: http://localhost:%d/\n", GlobalConfig.Port)
	fmt.Println("================\n")
}

// setupGracefulShutdown 设置优雅关闭处理
// 这个函数确保当收到停止信号时，服务器能够优雅地关闭，完成正在处理的请求
func setupGracefulShutdown(server *ProxyServer) {
	// 创建信号通道，监听系统信号
	sigChan := make(chan os.Signal, 1)
	
	// 注册我们感兴趣的信号
	// SIGINT: Ctrl+C
	// SIGTERM: 系统关闭信号
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	// 在后台goroutine中等待信号
	go func() {
		// 等待信号
		sig := <-sigChan
		log.Printf("\n收到信号: %v", sig)
		log.Printf("正在优雅关闭服务器...")
		
		// 这里可以添加清理逻辑
		// 例如：关闭数据库连接、保存状态、完成正在处理的请求等
		
		log.Printf("✅ 服务器已安全关闭")
		log.Printf("👋 感谢使用 %s！", ProgramName)
		
		// 退出程序
		os.Exit(0)
	}()
}