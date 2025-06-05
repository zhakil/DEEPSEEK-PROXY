package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

// enhanceRequestHeaders 为HTTP请求添加完整的浏览器伪装头部
// 这个函数就像为网络请求穿上一套完美的"伪装服"，让它看起来像来自真实的浏览器
func enhanceRequestHeaders(req *http.Request) {
	// 模拟最新版Chrome浏览器的User-Agent字符串
	// 更新为最新的Chrome版本，增强真实性
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	// 设置Accept头部，告诉服务器我们能接受什么类型的响应
	req.Header.Set("Accept", "application/json, text/plain, */*")

	// 设置语言偏好，模拟真实用户的语言环境
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7")

	// 设置编码偏好，告诉服务器我们支持的压缩方式
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")

	// DNT表示"Do Not Track"，这是现代浏览器的标准头部
	req.Header.Set("DNT", "1")

	// 连接类型设置，keep-alive可以复用TCP连接，提高效率
	req.Header.Set("Connection", "keep-alive")

	// 这些Sec-Fetch头部是现代浏览器的安全特性，帮助服务器识别请求类型
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "cross-site")

	// 添加Cache-Control头部，模拟浏览器的缓存行为
	req.Header.Set("Cache-Control", "no-cache")

	// 添加一个随机的请求ID，模拟真实应用程序的行为
	req.Header.Set("X-Request-ID", generateRandomRequestID())

	// 添加Referer头部，让请求看起来像是从一个合法的网页发起的
	req.Header.Set("Referer", "https://chat.deepseek.com/")

	// 添加Origin头部，进一步增强请求的可信度
	req.Header.Set("Origin", "https://chat.deepseek.com")

	log.Printf("已应用完整的浏览器伪装头部")
}

// generateRandomRequestID 生成一个随机的请求ID
// 这模拟了真实应用程序为每个请求分配唯一标识符的行为
func generateRandomRequestID() string {
	// 使用当前时间的纳秒部分作为随机种子，确保每次生成的ID都不同
	rand.Seed(time.Now().UnixNano())

	// 生成一个16位的随机十六进制字符串，这是常见的请求ID格式
	const chars = "0123456789abcdef"
	result := make([]byte, 16)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}

	return string(result)
}

// mapNewModelsToDeepSeek 将新的OpenAI模型映射到DeepSeek模型
// 这个函数专门处理o3和o4-mini等新模型的映射逻辑
func mapNewModelsToDeepSeek(requestedModel string) string {
	// 新的模型映射表，专门针对最新的OpenAI模型
	newModelMapping := map[string]string{
		// o3系列模型映射到DeepSeek的推理模型
		"o3":         "deepseek-reasoner",
		"o3-preview": "deepseek-reasoner",
		"o3-mini":    "deepseek-reasoner",

		// o4系列模型映射
		"o4-mini": "deepseek-reasoner", // o4-mini也使用推理模型

		// 保持对经典模型的支持
		"gpt-4o":        "deepseek-reasoner",
		"gpt-4":         "deepseek-chat",
		"gpt-3.5-turbo": "deepseek-chat",

		// DeepSeek原生模型保持不变
		"deepseek-chat":     "deepseek-chat",
		"deepseek-coder":    "deepseek-coder",
		"deepseek-reasoner": "deepseek-reasoner",
	}

	if mappedModel, exists := newModelMapping[requestedModel]; exists {
		log.Printf("新模型映射: %s -> %s", requestedModel, mappedModel)
		return mappedModel
	}

	// 如果没有找到映射，默认使用推理模型
	log.Printf("未知模型 %s，默认映射到 deepseek-reasoner", requestedModel)
	return "deepseek-reasoner"
}

// handleChatCompletions 处理聊天完成请求
// 这是我们代理服务器最重要的处理器，负责处理所有的AI对话请求
func (ps *ProxyServer) handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	// 记录请求信息，便于监控和调试
	logRequest(r, "聊天完成")

	// 设置CORS头部，允许跨域访问
	ps.handleCORS(w, r)

	// 如果是OPTIONS预检请求，直接返回
	if r.Method == "OPTIONS" {
		return
	}

	// 只接受POST请求，因为聊天完成需要发送请求体
	if r.Method != "POST" {
		handleError(w, fmt.Errorf("不支持的请求方法: %s", r.Method),
			http.StatusMethodNotAllowed, "方法检查")
		return
	}

	// 验证API密钥，确保只有授权用户可以访问
	if err := validateAPIKey(r); err != nil {
		handleError(w, err, http.StatusUnauthorized, "API密钥验证")
		return
	}

	// 生成唯一的请求ID，用于追踪整个请求过程
	requestID := generateRequestID()
	log.Printf("[%s] 开始处理聊天完成请求", requestID)

	// 解析客户端发送的OpenAI格式请求
	var openaiReq ChatRequest
	if err := readJSONRequest(r, &openaiReq); err != nil {
		handleError(w, fmt.Errorf("解析请求失败: %w", err),
			http.StatusBadRequest, "请求解析")
		return
	}

	// 详细记录请求内容，便于调试
	log.Printf("[%s] 收到的完整请求内容:", requestID)
	log.Printf("  模型: %s", openaiReq.Model)
	log.Printf("  消息数量: %d", len(openaiReq.Messages))
	log.Printf("  流式模式: %v", openaiReq.Stream)
	if openaiReq.Temperature != nil {
		log.Printf("  温度参数: %.2f", *openaiReq.Temperature)
	}
	if openaiReq.MaxTokens != nil {
		log.Printf("  最大Token: %d", *openaiReq.MaxTokens)
	}

	// 将OpenAI请求转换为DeepSeek格式
	deepseekReq, err := ps.convertToDeepSeekRequest(openaiReq, requestID)
	if err != nil {
		handleError(w, fmt.Errorf("请求转换失败: %w", err),
			http.StatusInternalServerError, "请求转换")
		return
	}

	// 发送请求到DeepSeek API并处理响应
	if openaiReq.Stream {
		// 处理流式响应，适用于需要实时显示生成过程的场景
		ps.handleStreamingResponse(w, r, deepseekReq, openaiReq.Model, requestID)
	} else {
		// 处理普通响应，等待完整结果后一次性返回
		ps.handleNormalResponse(w, deepseekReq, openaiReq.Model, requestID)
	}
}

// convertToDeepSeekRequest 将OpenAI请求转换为DeepSeek格式
// 这个函数是翻译过程的核心，处理两种API格式之间的所有差异
func (ps *ProxyServer) convertToDeepSeekRequest(openaiReq ChatRequest, requestID string) (*DeepSeekRequest, error) {
	log.Printf("[%s] 开始转换请求格式", requestID)

	// 使用新的模型映射函数
	deepseekModel := mapNewModelsToDeepSeek(openaiReq.Model)
	log.Printf("[%s] 模型映射: %s -> %s", requestID, openaiReq.Model, deepseekModel)

	// 检查是否使用推理模型
	isReasoningModel := deepseekModel == "deepseek-reasoner"
	if isReasoningModel {
		log.Printf("[%s] 使用DeepSeek推理模型，将提供完整的思考过程", requestID)
	}

	// 创建DeepSeek请求结构
	deepseekReq := &DeepSeekRequest{
		Model:    deepseekModel,
		Messages: convertMessagesFormat(openaiReq.Messages),
		Stream:   openaiReq.Stream,
	}

	// 处理可选参数
	// 注意：DeepSeek-Reasoner模型不支持temperature等采样参数
	if !isReasoningModel {
		// 只为非推理模型设置采样参数
		if openaiReq.Temperature != nil {
			deepseekReq.Temperature = *openaiReq.Temperature
			log.Printf("[%s] 设置温度参数: %.2f", requestID, *openaiReq.Temperature)
		} else {
			deepseekReq.Temperature = 0.7
		}
	} else {
		// 推理模型忽略temperature设置
		if openaiReq.Temperature != nil {
			log.Printf("[%s] 推理模型忽略温度参数设置", requestID)
		}
	}

	// 最大令牌数控制生成文本的长度
	if openaiReq.MaxTokens != nil {
		deepseekReq.MaxTokens = *openaiReq.MaxTokens
		log.Printf("[%s] 设置最大令牌数: %d", requestID, *openaiReq.MaxTokens)
	}

	// 处理工具调用功能
	if len(openaiReq.Tools) > 0 {
		deepseekReq.Tools = openaiReq.Tools
		deepseekReq.ToolChoice = convertToolChoice(openaiReq.ToolChoice)
		log.Printf("[%s] 设置工具: %d个工具, 选择策略: %s",
			requestID, len(openaiReq.Tools), deepseekReq.ToolChoice)
	} else if len(openaiReq.Functions) > 0 {
		// 处理旧版本的Functions格式（向后兼容）
		tools := make([]Tool, len(openaiReq.Functions))
		for i, fn := range openaiReq.Functions {
			tools[i] = Tool{
				Type:     "function",
				Function: fn,
			}
		}
		deepseekReq.Tools = tools
		deepseekReq.ToolChoice = convertToolChoice(openaiReq.ToolChoice)
		log.Printf("[%s] 转换Functions为Tools: %d个函数", requestID, len(openaiReq.Functions))
	}

	log.Printf("[%s] 请求转换完成", requestID)
	return deepseekReq, nil
}

// handleNormalResponse 处理普通（非流式）响应
// 这种方式等待DeepSeek完全生成响应后，一次性返回给客户端
func (ps *ProxyServer) handleNormalResponse(w http.ResponseWriter, deepseekReq *DeepSeekRequest, originalModel, requestID string) {
	log.Printf("[%s] 处理普通响应模式", requestID)

	// 向DeepSeek发送请求
	deepseekResp, err := ps.sendRequestToDeepSeek(deepseekReq, requestID)
	if err != nil {
		handleError(w, fmt.Errorf("DeepSeek请求失败: %w", err),
			http.StatusBadGateway, "DeepSeek通信")
		return
	}

	// 将DeepSeek响应转换为OpenAI格式
	openaiResp := ps.convertToOpenAIResponse(deepseekResp, originalModel, requestID)

	// 返回响应给客户端
	w.Header().Set("Content-Type", "application/json")
	if err := writeJSONResponse(w, openaiResp); err != nil {
		log.Printf("[%s] 写入响应失败: %v", requestID, err)
		return
	}

	log.Printf("[%s] 普通响应处理完成", requestID)
}

// sendRequestToDeepSeek 向DeepSeek API发送普通请求
// 这个函数负责与DeepSeek API的实际通信，现在包含完整的浏览器伪装
func (ps *ProxyServer) sendRequestToDeepSeek(req *DeepSeekRequest, requestID string) (*DeepSeekResponse, error) {
	log.Printf("[%s] 向DeepSeek发送请求", requestID)

	// 将请求序列化为JSON
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 创建HTTP请求
	url := ps.config.Endpoint + "/v1/chat/completions"
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置基础的请求头部
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+ps.config.DeepSeekAPIKey)

	// *** 关键改进：应用完整的浏览器伪装 ***
	enhanceRequestHeaders(httpReq)
	log.Printf("[%s] 已应用浏览器伪装头部", requestID)

	// 发送请求
	client := createHTTPClient()
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errorMsg := fmt.Sprintf("DeepSeek API返回错误 %d: %s", resp.StatusCode, string(body))
		log.Printf("[%s] %s", requestID, errorMsg)
		return nil, fmt.Errorf(errorMsg)
	}

	// 解析响应
	var deepseekResp DeepSeekResponse
	if err := json.NewDecoder(resp.Body).Decode(&deepseekResp); err != nil {
		return nil, fmt.Errorf("解析DeepSeek响应失败: %w", err)
	}

	log.Printf("[%s] DeepSeek响应接收成功", requestID)
	return &deepseekResp, nil
}

// sendStreamingRequestToDeepSeek 向DeepSeek API发送流式请求
// 现在也包含完整的浏览器伪装功能
func (ps *ProxyServer) sendStreamingRequestToDeepSeek(req *DeepSeekRequest, requestID string) (*http.Response, error) {
	log.Printf("[%s] 向DeepSeek发送流式请求", requestID)

	// 序列化请求
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 创建HTTP请求
	url := ps.config.Endpoint + "/v1/chat/completions"
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置基础头部
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+ps.config.DeepSeekAPIKey)
	httpReq.Header.Set("Accept", "text/event-stream")

	// *** 关键改进：为流式请求也应用浏览器伪装 ***
	enhanceRequestHeaders(httpReq)
	log.Printf("[%s] 已为流式请求应用浏览器伪装头部", requestID)

	// 发送请求
	client := createHTTPClient()
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("发送流式请求失败: %w", err)
	}

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("DeepSeek API返回错误 %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("[%s] DeepSeek流式响应开始接收", requestID)
	return resp, nil
}

// handleStreamingResponse 处理流式响应
// 这种方式实时传输DeepSeek的生成过程，让用户看到文字逐步出现
func (ps *ProxyServer) handleStreamingResponse(w http.ResponseWriter, r *http.Request,
	deepseekReq *DeepSeekRequest, originalModel, requestID string) {

	log.Printf("[%s] 处理流式响应模式", requestID)

	// 设置流式响应的HTTP头部
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	// 获取Flusher接口，用于实时发送数据
	flusher, ok := w.(http.Flusher)
	if !ok {
		handleError(w, fmt.Errorf("服务器不支持流式响应"),
			http.StatusInternalServerError, "流式响应检查")
		return
	}

	// 向DeepSeek发送流式请求
	resp, err := ps.sendStreamingRequestToDeepSeek(deepseekReq, requestID)
	if err != nil {
		handleError(w, fmt.Errorf("DeepSeek流式请求失败: %w", err),
			http.StatusBadGateway, "DeepSeek流式通信")
		return
	}
	defer resp.Body.Close()

	// 创建上下文用于处理客户端断开连接
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// 处理流式数据
	ps.processStreamingData(w, resp.Body, flusher, originalModel, requestID, ctx)

	log.Printf("[%s] 流式响应处理完成", requestID)
}

// processStreamingData 处理流式数据
// 这个函数负责读取DeepSeek的流式响应并转换为OpenAI格式
func (ps *ProxyServer) processStreamingData(w http.ResponseWriter, reader io.Reader,
	flusher http.Flusher, originalModel, requestID string, ctx context.Context) {

	log.Printf("[%s] 开始处理流式数据", requestID)

	// 创建一个扫描器来逐行读取SSE数据
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			log.Printf("[%s] 客户端连接已断开", requestID)
			return
		default:
			line := scanner.Text()

			// 处理Server-Sent Events格式
			if strings.HasPrefix(line, "data: ") {
				// 提取JSON数据部分
				dataContent := strings.TrimPrefix(line, "data: ")

				// 检查是否是结束标记
				if dataContent == "[DONE]" {
					fmt.Fprintf(w, "data: [DONE]\n\n")
					flusher.Flush()
					log.Printf("[%s] 流式数据传输完成", requestID)
					return
				}

				// 转换DeepSeek流式响应为OpenAI格式
				if dataContent != "" {
					convertedData := ps.convertStreamChunk(dataContent, originalModel, requestID)
					if convertedData != "" {
						fmt.Fprintf(w, "data: %s\n\n", convertedData)
						flusher.Flush()
					}
				}
			} else if line == "" {
				continue
			} else {
				fmt.Fprintf(w, "%s\n", line)
				flusher.Flush()
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("[%s] 流式数据读取错误: %v", requestID, err)
	}

	log.Printf("[%s] 流式数据处理完成", requestID)
}

// convertStreamChunk 转换单个流式数据块
func (ps *ProxyServer) convertStreamChunk(dataContent, originalModel, requestID string) string {
	var deepSeekChunk map[string]interface{}
	if err := json.Unmarshal([]byte(dataContent), &deepSeekChunk); err != nil {
		log.Printf("[%s] 解析流式数据块失败: %v", requestID, err)
		return ""
	}

	// 转换模型名称为客户端请求的原始模型名
	if model, exists := deepSeekChunk["model"]; exists {
		deepSeekChunk["model"] = originalModel
		log.Printf("[%s] 转换流式块模型名: %v -> %s", requestID, model, originalModel)
	}

	convertedData, err := json.Marshal(deepSeekChunk)
	if err != nil {
		log.Printf("[%s] 序列化转换后的流式数据失败: %v", requestID, err)
		return ""
	}

	return string(convertedData)
}

// convertToOpenAIResponse 将DeepSeek响应转换为OpenAI格式
func (ps *ProxyServer) convertToOpenAIResponse(deepseekResp *DeepSeekResponse, originalModel, requestID string) map[string]interface{} {
	log.Printf("[%s] 转换响应格式", requestID)

	// 检查是否是推理模型
	isReasoningModel := originalModel == "o3" || originalModel == "o3-preview" || originalModel == "o3-mini" ||
		originalModel == "o4-mini" || originalModel == "deepseek-reasoner"

	// 处理Choices字段
	var processedChoices []interface{}

	for _, choice := range deepseekResp.Choices {
		processedChoice := map[string]interface{}{
			"index":         choice.Index,
			"finish_reason": choice.FinishReason,
		}

		messageMap := map[string]interface{}{
			"role":    choice.Message.Role,
			"content": choice.Message.Content,
		}

		// 如果是推理模型，检查是否有reasoning_content
		if isReasoningModel && choice.Message.ReasoningContent != "" {
			messageMap["reasoning_content"] = choice.Message.ReasoningContent
			log.Printf("[%s] 发现推理内容，长度: %d字符", requestID, len(choice.Message.ReasoningContent))
		}

		// 处理工具调用
		if len(choice.Message.ToolCalls) > 0 {
			messageMap["tool_calls"] = choice.Message.ToolCalls
		}

		if choice.Message.Name != "" {
			messageMap["name"] = choice.Message.Name
		}
		if choice.Message.ToolCallID != "" {
			messageMap["tool_call_id"] = choice.Message.ToolCallID
		}

		processedChoice["message"] = messageMap
		processedChoices = append(processedChoices, processedChoice)
	}

	// 创建OpenAI格式的响应
	openaiResp := map[string]interface{}{
		"id":      deepseekResp.ID,
		"object":  "chat.completion",
		"created": deepseekResp.Created,
		"model":   originalModel, // 使用客户端请求的模型名
		"choices": processedChoices,
		"usage":   deepseekResp.Usage,
	}

	if isReasoningModel {
		log.Printf("[%s] 推理模型响应处理完成", requestID)
	}

	log.Printf("[%s] 响应格式转换完成", requestID)
	return openaiResp
}
