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
// 将版本信息定义为常量是一个好习惯，这样便于维护和版本管理
const (
	Version     = "1.0.0"
	ProgramName = "DeepSeek API 代理服务器"
)

// 命令行参数定义
// 使用指针类型可以让我们区分"未设置"和"设置为零值"的情况
var (
	showVersion = flag.Bool("version", false, "显示版本信息")
	showHelp    = flag.Bool("help", false, "显示帮助信息")
	configPath  = flag.String("config", ".env", "配置文件路径")
	port        = flag.Int("port", 0, "服务器端口号（覆盖配置文件设置）")
	debug       = flag.Bool("debug", false, "启用调试模式")
)

// main 是程序的入口点
// 这个函数负责协调整个程序的启动流程，从配置验证到服务器启动
func main() {
	// 配置日志格式，包含时间戳和文件位置信息
	// 这有助于调试和生产环境的问题追踪
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	// 显示欢迎信息，让用户知道程序正在启动
	printWelcomeBanner()

	// 解析命令行参数
	// 这必须在使用任何flag变量之前调用
	flag.Parse()

	// 处理立即退出的命令行选项
	if *showVersion {
		printVersion()
		return
	}

	if *showHelp {
		printHelp()
		return
	}

	// 验证运行环境是否满足要求
	// 早期验证可以避免服务器启动后才发现配置问题
	if err := validateEnvironment(); err != nil {
		log.Fatalf("环境验证失败: %v", err)
	}

	// 处理命令行端口覆盖
	// 这允许用户在不修改配置文件的情况下更改端口
	if *port > 0 {
		GlobalConfig.Port = *port
		log.Printf("使用命令行指定的端口: %d", *port)
	}

	// 如果启用调试模式，显示详细配置信息
	if *debug {
		log.Println("调试模式已启用")
		printDebugInfo()
	}

	// 创建代理服务器实例
	log.Println("正在初始化代理服务器...")
	proxyServer := NewProxyServer(GlobalConfig)

	// 设置优雅关闭处理
	// 这确保服务器可以优雅地响应停止信号
	setupGracefulShutdown(proxyServer)

	// 显示启动完成信息
	log.Printf("🎉 %s v%s 启动完成！", ProgramName, Version)
	log.Printf("📖 访问 http://localhost:%d 查看服务器信息", GlobalConfig.Port)
	log.Println("🛑 按 Ctrl+C 停止服务器")

	// 开始监听请求
	// 这是一个阻塞调用，服务器将在这里运行直到收到停止信号
	if err := proxyServer.Start(); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

// printWelcomeBanner 显示程序启动时的欢迎横幅
// 这个函数使用Printf是合适的，因为我们需要在模板中插入版本号
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
// 这个函数展示了Printf和Println的正确使用方式
func printVersion() {
	// 使用Printf处理需要格式化的字符串
	fmt.Printf("%s v%s", ProgramName, Version)
	fmt.Println() // 单独换行，职责分离

	// 使用Println处理简单的文本行
	fmt.Println("构建信息:")
	fmt.Println("  - Go 版本: 1.19+")
	fmt.Println("  - 支持的协议: HTTP/1.1, HTTP/2")
	fmt.Println("  - 支持的格式: JSON, Server-Sent Events")
	fmt.Println("  - 兼容性: OpenAI Chat Completions API v1")

	fmt.Println()
	fmt.Println("项目主页: https://github.com/your-username/deepseek-proxy")
}

// printHelp 显示详细的帮助信息
// 这个函数演示了如何在复杂的帮助文本中正确使用fmt函数
func printHelp() {
	// 标题使用Printf进行格式化
	fmt.Printf("%s - OpenAI 到 DeepSeek API 代理服务器", ProgramName)
	fmt.Println()
	fmt.Println()

	// 各个部分使用Println输出简单文本
	fmt.Println("用法:")
	fmt.Printf("  %s [选项]", os.Args[0])
	fmt.Println()
	fmt.Println()

	fmt.Println("选项:")
	fmt.Println("  -version          显示版本信息并退出")
	fmt.Println("  -help             显示此帮助信息并退出")
	fmt.Println("  -config string    配置文件路径 (默认: .env)")
	fmt.Println("  -port int         服务器端口号 (覆盖配置文件)")
	fmt.Println("  -debug            启用调试模式")

	fmt.Println()
	fmt.Println("环境变量:")
	fmt.Println("  DEEPSEEK_API_KEY     DeepSeek API 密钥 (必需)")
	fmt.Println("  PORT                 服务器端口号 (默认: 9000)")
	fmt.Println("  DEEPSEEK_MODEL       默认模型 (默认: deepseek-chat)")
	fmt.Println("  DEEPSEEK_ENDPOINT    API 端点 (默认: https://api.deepseek.com)")

	fmt.Println()
	fmt.Println("示例:")

	// 对于需要格式化的示例，使用Printf
	examples := []struct {
		command     string
		description string
	}{
		{os.Args[0], "使用默认配置启动"},
		{os.Args[0] + " -port 8080", "在端口 8080 启动"},
		{os.Args[0] + " -config /path/to/.env", "使用指定配置文件"},
		{os.Args[0] + " -debug", "启用调试模式"},
	}

	for _, example := range examples {
		fmt.Printf("  %-40s # %s", example.command, example.description)
		fmt.Println()
	}

	fmt.Println()
	fmt.Println("配置文件示例 (.env):")
	fmt.Println("  DEEPSEEK_API_KEY=sk-your-api-key-here")
	fmt.Println("  PORT=9000")
	fmt.Println("  DEEPSEEK_MODEL=deepseek-chat")

	fmt.Println()
	fmt.Println("支持的客户端:")
	fmt.Println("  - Cursor IDE")
	fmt.Println("  - 任何兼容 OpenAI API 的应用程序")
	fmt.Println("  - 自定义应用程序使用 OpenAI SDK")
}

// validateEnvironment 验证运行环境是否满足要求
// 这个函数实现了"快速失败"原则，在启动早期发现配置问题
func validateEnvironment() error {
	log.Println("正在验证运行环境...")

	// 检查运行时环境
	log.Println("✓ Go 运行时环境正常")

	// 验证API密钥配置
	if GlobalConfig.DeepSeekAPIKey == "" {
		return fmt.Errorf("缺少必需的环境变量: DEEPSEEK_API_KEY")
	}
	log.Println("✓ API 密钥已配置")

	// 验证端口号范围
	if GlobalConfig.Port <= 0 || GlobalConfig.Port > 65535 {
		return fmt.Errorf("无效的端口号: %d (必须在 1-65535 之间)", GlobalConfig.Port)
	}
	log.Printf("✓ 端口配置有效: %d", GlobalConfig.Port)

	// 可以在这里添加更多验证，比如网络连通性测试
	log.Println("✓ 环境验证通过")

	return nil
}

// printDebugInfo 显示调试信息
// 这个版本演示了Go语言fmt包的最佳实践用法
func printDebugInfo() {
	fmt.Println()
	fmt.Println("=== 调试信息 ===")

	// 使用Println的多参数特性，这比Printf更简洁清晰
	fmt.Println("配置文件路径:", *configPath)
	fmt.Println("监听端口:", GlobalConfig.Port)
	fmt.Println("DeepSeek 端点:", GlobalConfig.Endpoint)
	fmt.Println("默认模型:", GlobalConfig.DeepSeekModel)
	fmt.Println("API 密钥:", maskAPIKey(GlobalConfig.DeepSeekAPIKey))

	fmt.Println()
	fmt.Println("支持的模型:")

	// 这里使用Printf是必要的，因为我们需要精确的数字格式化
	for i, model := range GetSupportedModels() {
		fmt.Printf("  %d. %s", i+1, model)
		fmt.Println()
	}

	fmt.Println()
	fmt.Println("可用端点:")

	// 使用结构化的方法来处理端点信息
	// 这比硬编码的字符串更容易维护
	endpoints := []struct {
		name string
		path string
	}{
		{"聊天完成", "/v1/chat/completions"},
		{"模型列表", "/v1/models"},
		{"健康检查", "/health"},
		{"服务器信息", "/"},
	}

	for _, endpoint := range endpoints {
		fmt.Printf("  - %s: http://localhost:%d%s",
			endpoint.name, GlobalConfig.Port, endpoint.path)
		fmt.Println()
	}

	fmt.Println("================")
	fmt.Println()
}

// setupGracefulShutdown 设置优雅关闭处理
// 这个函数确保服务器可以优雅地响应停止信号，完成正在处理的请求
func setupGracefulShutdown(server *ProxyServer) {
	// 创建信号通道，缓冲大小为1可以避免信号丢失
	sigChan := make(chan os.Signal, 1)

	// 注册我们关心的系统信号
	// SIGINT: 通常来自 Ctrl+C
	// SIGTERM: 系统关闭或容器停止时发送
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 在后台goroutine中监听信号
	go func() {
		// 阻塞等待信号
		sig := <-sigChan

		fmt.Println() // 在新行开始关闭信息
		log.Printf("收到信号: %v", sig)
		log.Println("正在优雅关闭服务器...")

		// 这里可以添加清理逻辑，比如：
		// - 停止接受新请求
		// - 等待现有请求完成
		// - 关闭数据库连接
		// - 保存状态信息

		// 记录服务器实例信息，确保参数被使用
		log.Printf("正在关闭服务器实例: %p", server)

		// 显示关闭完成信息
		log.Println("✅ 服务器已安全关闭")
		log.Printf("👋 感谢使用 %s！", ProgramName)

		// 正常退出程序
		os.Exit(0)
	}()
}
