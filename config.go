package main

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// 全局配置变量
var GlobalConfig *ProxyConfig

// 初始化配置
func init() {
	log.Printf("开始初始化代理配置...")

	if err := godotenv.Load(); err != nil {
		log.Printf("警告：无法加载.env文件，将使用环境变量: %v", err)
	}

	// 初始化全局配置
	GlobalConfig = &ProxyConfig{
		Port:           getEnvAsInt("PORT", 9000),
		Host:           getEnvAsString("HOST", ""),                                       // 默认空字符串表示localhost
		DeepSeekAPIKey: getEnvAsString("DEEPSEEK_API_KEY", ""),
		DeepSeekModel:  getEnvAsString("DEEPSEEK_MODEL", "deepseek-reasoner"),           // 默认使用推理模型
		Endpoint:       getEnvAsString("DEEPSEEK_ENDPOINT", "https://api.deepseek.com"),
		ProxyURL:       getEnvAsString("PROXY_URL", ""),
	}

	validateConfig(GlobalConfig)

	log.Printf("配置初始化完成:")
	log.Printf("  - 绑定主机: %s", getDisplayHost(GlobalConfig.Host))
	log.Printf("  - 监听端口: %d", GlobalConfig.Port)
	log.Printf("  - DeepSeek模型: %s", GlobalConfig.DeepSeekModel)
	log.Printf("  - API端点: %s", GlobalConfig.Endpoint)
	log.Printf("  - API密钥状态: %s", maskAPIKey(GlobalConfig.DeepSeekAPIKey))
	if GlobalConfig.ProxyURL != "" {
		log.Printf("  - Proxy URL: %s", GlobalConfig.ProxyURL)
	}
}

// getDisplayHost 获取用于显示的主机地址
func getDisplayHost(host string) string {
	if host == "" {
		return "localhost"
	}
	if host == "0.0.0.0" {
		return "所有网络接口 (0.0.0.0)"
	}
	return host
}

// 从环境变量获取字符串值
func getEnvAsString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// 从环境变量获取整数值
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			log.Printf("从环境变量读取 %s: %d", key, intValue)
			return intValue
		}
		log.Printf("警告：环境变量 %s 的值 '%s' 不是有效整数，使用默认值 %d", key, value, defaultValue)
	}
	return defaultValue
}

// 验证配置的有效性
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

// 隐藏API密钥的敏感部分
func maskAPIKey(apiKey string) string {
	if apiKey == "" {
		return "未设置"
	}

	if len(apiKey) < 8 {
		return "已设置（格式可能有误）"
	}

	return apiKey[:4] + "****" + apiKey[len(apiKey)-4:]
}

// 获取支持的模型列表
func GetSupportedModels() []string {
	return []string{
		"gpt-4o",
		"gpt-4",
		"gpt-3.5-turbo",
		"deepseek-chat",
		"deepseek-coder",
		"deepseek-reasoner",
		"o3",
		"o3-preview",
		"o3-mini",
		"o4-mini",
	}
}

// 将OpenAI模型名映射到DeepSeek模型名
func MapModelName(openaiModel string) string {
	// 统一映射到推理模型
	modelMapping := map[string]string{
		"gpt-4o":            "deepseek-reasoner",
		"gpt-4":             "deepseek-reasoner",
		"gpt-3.5-turbo":     "deepseek-reasoner",
		"deepseek-chat":     "deepseek-reasoner",
		"deepseek-coder":    "deepseek-reasoner",
		"deepseek-reasoner": "deepseek-reasoner",
		"o3":                "deepseek-reasoner",
		"o3-preview":        "deepseek-reasoner",
		"o3-mini":           "deepseek-reasoner",
		"o4-mini":           "deepseek-reasoner",
	}

	if mappedModel, exists := modelMapping[openaiModel]; exists {
		log.Printf("模型映射: %s -> %s", openaiModel, mappedModel)
		return mappedModel
	}

	log.Printf("未知模型 %s，使用默认模型: %s", openaiModel, GlobalConfig.DeepSeekModel)
	return GlobalConfig.DeepSeekModel
}

// 检查模型是否支持工具调用
func ModelSupportsTools(modelName string) bool {
	toolSupportedModels := map[string]bool{
		"deepseek-chat":     true,
		"deepseek-coder":    true,
		"deepseek-reasoner": true,
		"gpt-4o":            true,
		"gpt-4":             true,
	}

	supported, exists := toolSupportedModels[modelName]
	if !exists {
		return false
	}

	return supported
}