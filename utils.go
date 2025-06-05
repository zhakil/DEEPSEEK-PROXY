package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// writeJSONResponse 将数据以JSON格式写入HTTP响应
// 这个函数就像是一个智能的翻译官，把Go的数据结构转换成JSON格式发送给客户端
func writeJSONResponse(w http.ResponseWriter, data interface{}) error {
	// 设置正确的内容类型，告诉客户端这是JSON数据
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	
	// 将数据转换为JSON格式
	// json.Marshal就像是一个打包机，把复杂的数据结构打包成JSON字符串
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("JSON序列化失败: %v", err)
		// 修复：错误字符串改为小写开头
		return fmt.Errorf("json序列化失败: %w", err)
	}
	
	// 写入响应
	if _, err := w.Write(jsonData); err != nil {
		log.Printf("写入响应失败: %v", err)
		return fmt.Errorf("写入响应失败: %w", err)
	}
	
	return nil
}

// readJSONRequest 从HTTP请求中读取并解析JSON数据
// 这个函数就像是一个拆包员，把客户端发送的JSON数据解包成Go的数据结构
func readJSONRequest(r *http.Request, target interface{}) error {
	// 读取请求体中的所有数据
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("读取请求体失败: %w", err)
	}
	
	// 记录原始请求数据，便于调试
	log.Printf("收到JSON请求: %s", string(body))
	
	// 将JSON数据解析到目标结构体中
	if err := json.Unmarshal(body, target); err != nil {
		// 修复：错误字符串改为小写开头
		return fmt.Errorf("json解析失败: %w", err)
	}
	
	return nil
}

// validateAPIKey 验证API密钥的有效性
// 这个函数就像是门卫，检查来访者是否有正确的通行证
func validateAPIKey(r *http.Request) error {
	// 从Authorization头部获取API密钥
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		// 修复：错误字符串改为小写开头
		return fmt.Errorf("缺少authorization头部")
	}
	
	// 检查是否是Bearer令牌格式
	if !strings.HasPrefix(authHeader, "Bearer ") {
		// 修复：错误字符串改为小写开头
		return fmt.Errorf("authorization头部格式错误，应该是 'Bearer <token>'")
	}
	
	// 提取实际的API密钥
	providedKey := strings.TrimPrefix(authHeader, "Bearer ")
	if providedKey == "" {
		// 修复：错误字符串改为小写开头
		return fmt.Errorf("api密钥为空")
	}
	
	// 验证API密钥是否与配置中的密钥匹配
	// 在实际应用中，你可能需要更复杂的验证逻辑
	if providedKey != GlobalConfig.DeepSeekAPIKey {
		// 修复：错误字符串改为小写开头
		return fmt.Errorf("无效的api密钥")
	}
	
	return nil
}

// convertMessagesFormat 转换消息格式以适配DeepSeek API
// 这是翻译过程的核心函数，处理OpenAI和DeepSeek之间的格式差异
func convertMessagesFormat(messages []Message) []Message {
	log.Printf("开始转换 %d 条消息格式", len(messages))
	
	convertedMessages := make([]Message, 0, len(messages))
	
	for i, msg := range messages {
		log.Printf("处理消息 %d: 角色=%s", i, msg.Role)
		
		// 创建转换后的消息副本
		convertedMsg := Message{
			Role:       msg.Role,
			Content:    msg.Content,
			Name:       msg.Name,
			ToolCallID: msg.ToolCallID,
		}
		
		// 处理特殊的角色转换
		// OpenAI使用"function"角色，而DeepSeek使用"tool"角色
		if msg.Role == "function" {
			convertedMsg.Role = "tool"
			log.Printf("将function角色转换为tool角色")
		}
		
		// 处理工具调用
		if len(msg.ToolCalls) > 0 {
			log.Printf("处理 %d 个工具调用", len(msg.ToolCalls))
			convertedMsg.ToolCalls = make([]ToolCall, len(msg.ToolCalls))
			
			for j, toolCall := range msg.ToolCalls {
				convertedMsg.ToolCalls[j] = ToolCall{
					ID:   toolCall.ID,
					Type: "function", // 确保类型正确
					Function: struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					}{
						Name:      toolCall.Function.Name,
						Arguments: toolCall.Function.Arguments,
					},
				}
				log.Printf("转换工具调用 %d: %s", j, toolCall.Function.Name)
			}
		}
		
		convertedMessages = append(convertedMessages, convertedMsg)
	}
	
	log.Printf("消息格式转换完成，共处理 %d 条消息", len(convertedMessages))
	return convertedMessages
}

// convertToolChoice 转换工具选择策略
// 不同的API对工具选择有不同的表示方式，这个函数处理这些差异
func convertToolChoice(choice interface{}) string {
	if choice == nil {
		return "auto" // 默认策略
	}
	
	// 如果是字符串类型（auto、none等）
	if str, ok := choice.(string); ok {
		switch str {
		case "auto", "none":
			return str
		default:
			log.Printf("未知的工具选择策略: %s，使用默认值auto", str)
			return "auto"
		}
	}
	
	// 如果是复杂对象（指定特定函数）
	if choiceMap, ok := choice.(map[string]interface{}); ok {
		if choiceType, exists := choiceMap["type"]; exists && choiceType == "function" {
			// DeepSeek可能不支持指定特定函数，转换为auto
			log.Printf("指定函数的工具选择策略转换为auto")
			return "auto"
		}
	}
	
	log.Printf("无法识别的工具选择策略，使用默认值auto")
	return "auto"
}

// logRequest 记录请求信息，用于调试和监控
// 这个函数帮助我们了解代理服务器接收到的请求情况
func logRequest(r *http.Request, requestType string) {
	// 获取客户端IP地址
	clientIP := getClientIP(r)
	
	// 记录请求的基本信息
	log.Printf("=== %s 请求 ===", requestType)
	log.Printf("客户端IP: %s", clientIP)
	log.Printf("请求方法: %s", r.Method)
	log.Printf("请求路径: %s", r.URL.Path)
	log.Printf("User-Agent: %s", r.Header.Get("User-Agent"))
	
	// 如果有查询参数，也记录下来
	if r.URL.RawQuery != "" {
		log.Printf("查询参数: %s", r.URL.RawQuery)
	}
}

// getClientIP 获取客户端的真实IP地址
// 在代理环境中，需要检查特殊的头部来获取真实IP
func getClientIP(r *http.Request) string {
	// 检查X-Forwarded-For头部（常用于反向代理）
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For可能包含多个IP，取第一个
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	
	// 检查X-Real-IP头部（Nginx常用）
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	
	// 如果没有代理头部，使用RemoteAddr
	// RemoteAddr格式通常是 "IP:Port"，我们只要IP部分
	if addr := r.RemoteAddr; addr != "" {
		if colonPos := strings.LastIndex(addr, ":"); colonPos != -1 {
			return addr[:colonPos]
		}
		return addr
	}
	
	return "unknown"
}

// createHTTPClient 创建用于与DeepSeek API通信的HTTP客户端
// 这个客户端配置了适当的超时和其他参数，确保可靠的通信
func createHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 60 * time.Second, // 总请求超时时间
		Transport: &http.Transport{
			// 连接超时配置
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
			
			// 连接池配置，提高性能
			MaxIdleConns:        100,              // 最大空闲连接数
			MaxIdleConnsPerHost: 10,               // 每个主机的最大空闲连接数
			IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间
		},
	}
}

// handleError 统一的错误处理函数
// 这个函数确保所有的错误都以一致的格式返回给客户端
func handleError(w http.ResponseWriter, err error, statusCode int, context string) {
	log.Printf("错误 [%s]: %v", context, err)
	
	// 创建错误响应
	errorResponse := map[string]interface{}{
		"error": map[string]interface{}{
			"message": err.Error(),
			"type":    "api_error",
			"code":    statusCode,
		},
		"timestamp": time.Now().Unix(),
	}
	
	// 设置错误状态码
	w.WriteHeader(statusCode)
	
	// 写入错误响应
	if writeErr := writeJSONResponse(w, errorResponse); writeErr != nil {
		log.Printf("写入错误响应失败: %v", writeErr)
		// 如果连错误响应都写不了，只能返回简单的文本错误
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
	}
}

// truncateString 截断字符串用于日志显示
// 当字符串太长时，这个函数帮助我们只显示前面的部分，避免日志过于冗长
func truncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength] + "..."
}

// generateRequestID 生成唯一的请求ID
// 每个请求都应该有一个唯一标识符，便于追踪和调试
func generateRequestID() string {
	// 使用时间戳和简单的随机数生成ID
	// 在生产环境中，你可能需要更复杂的UUID生成算法
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("req_%d", timestamp)
}