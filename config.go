package main

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// 全局配置变量
// 这个变量会在整个应用程序中被使用，存储所有的配置信息
var GlobalConfig *ProxyConfig

// 初始化配置
// init函数在程序启动时自动运行，负责加载所有必要的配置
func init() {
	log.Printf("开始初始化代理配置...")

	// 尝试加载.env文件
	// .env文件包含敏感信息如API密钥，不应该提交到版本控制系统
	if err := godotenv.Load(); err != nil {
		log.Printf("警告：无法加载.env文件，将使用环境变量: %v", err)
	}

	// 初始化全局配置
	GlobalConfig = &ProxyConfig{
		Port:           getEnvAsInt("PORT", 9000),                                       // 默认端口9000
		DeepSeekAPIKey: getEnvAsString("DEEPSEEK_API_KEY", ""),                          // DeepSeek API密钥
		DeepSeekModel:  getEnvAsString("DEEPSEEK_MODEL", "deepseek-chat"),               // 默认模型
		Endpoint:       getEnvAsString("DEEPSEEK_ENDPOINT", "https://api.deepseek.com"), // API端点
	}

	// 验证必需的配置
	validateConfig(GlobalConfig)

	log.Printf("配置初始化完成:")
	log.Printf("  - 监听端口: %d", GlobalConfig.Port)
	log.Printf("  - DeepSeek模型: %s", GlobalConfig.DeepSeekModel)
	log.Printf("  - API端点: %s", GlobalConfig.Endpoint)
	log.Printf("  - API密钥状态: %s", maskAPIKey(GlobalConfig.DeepSeekAPIKey))
}

// 从环境变量获取字符串值，如果不存在则使用默认值
// 这个函数让我们的配置更加灵活，可以通过环境变量轻松修改
func getEnvAsString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// 从环境变量获取整数值，如果不存在或无效则使用默认值
// 修复：现在真正实现了字符串到整数的转换
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		// 使用 strconv.Atoi 进行字符串到整数的转换
		if intValue, err := strconv.Atoi(value); err == nil {
			log.Printf("从环境变量 %s 读取到整数值: %d", key, intValue)
			return intValue
		} else {
			log.Printf("警告：环境变量 %s 的值 '%s' 不是有效整数，使用默认值 %d", key, value, defaultValue)
		}
	}
	return defaultValue
}

// 验证配置的有效性
// 确保所有必需的配置项都已正确设置
func validateConfig(config *ProxyConfig) {
	if config.DeepSeekAPIKey == "" {
		log.Fatal("错误：DEEPSEEK_API_KEY 环境变量是必需的，请设置你的DeepSeek API密钥")
	}

	if config.Port <= 0 || config.Port > 65535 {
		log.Fatal("错误：端口号必须在1-65535之间")
	}

	if config.Endpoint == "" {
		log.Fatal("错误：DeepSeek API端点不能为空")
	}

	log.Printf("✓ 配置验证通过")
}

// 隐藏API密钥的敏感部分，只显示前几位和后几位
// 这样在日志中既能确认密钥存在，又不会泄露完整密钥
func maskAPIKey(apiKey string) string {
	if apiKey == "" {
		return "未设置"
	}

	if len(apiKey) < 8 {
		return "已设置（格式可能有误）"
	}

	// 显示前4位和后4位，中间用*代替
	return apiKey[:4] + "****" + apiKey[len(apiKey)-4:]
}

// 获取支持的模型列表
// 这个函数返回我们代理支持的所有模型名称
func GetSupportedModels() []string {
	return []string{
		"gpt-4o",            // 映射到DeepSeek的高级模型
		"gpt-4",             // 映射到DeepSeek Chat
		"gpt-3.5-turbo",     // 映射到DeepSeek Chat
		"deepseek-chat",     // DeepSeek原生聊天模型
		"deepseek-coder",    // DeepSeek代码专用模型
		"deepseek-reasoner", // DeepSeek推理模型 - 新增！
		"o3",                // 映射到DeepSeek Reasoner（兼容OpenAI o1）
		"o4-mini",           // 映射到DeepSeek Reasoner
	}
}

// 将OpenAI模型名映射到DeepSeek模型名
// 这是翻译过程的关键部分，确保不同的模型请求能正确路由
func MapModelName(openaiModel string) string {
	// 模型映射表，就像一个字典，告诉我们如何翻译模型名称
	modelMapping := map[string]string{
		"gpt-4o":            GlobalConfig.DeepSeekModel, // 最新的GPT-4模型映射
		"gpt-4":             GlobalConfig.DeepSeekModel, // GPT-4映射
		"gpt-3.5-turbo":     GlobalConfig.DeepSeekModel, // GPT-3.5映射
		"deepseek-chat":     "deepseek-chat",            // 保持原名
		"deepseek-coder":    "deepseek-coder",           // 保持原名
		"deepseek-reasoner": "deepseek-reasoner",        // 推理模型保持原名
		"o3":                "deepseek-reasoner",        // OpenAI o1映射到推理模型
		"o4-mini":           "deepseek-reasoner",        // o1 mini版映射到推理模型
	}

	// 如果找到映射，使用映射的模型名；否则使用默认模型
	if mappedModel, exists := modelMapping[openaiModel]; exists {
		log.Printf("模型映射: %s -> %s", openaiModel, mappedModel)
		return mappedModel
	}

	log.Printf("未知模型 %s，使用默认模型: %s", openaiModel, GlobalConfig.DeepSeekModel)
	return GlobalConfig.DeepSeekModel
}

// 检查模型是否支持工具调用
// 不是所有模型都支持函数调用功能，这个函数帮我们判断
func ModelSupportsTools(modelName string) bool {
	// 支持工具调用的模型列表
	toolSupportedModels := map[string]bool{
		"deepseek-chat":  true,
		"deepseek-coder": true,
		"gpt-4o":         true,
		"gpt-4":          true,
	}

	supported, exists := toolSupportedModels[modelName]
	if !exists {
		return false // 默认不支持
	}

	return supported
}
