package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

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
	
	log.Printf("[%s] 收到请求: 模型=%s, 消息数=%d, 流式=%v", 
		requestID, openaiReq.Model, len(openaiReq.Messages), openaiReq.Stream)
	
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
		ps.handleNormalResponse(w, r, deepseekReq, openaiReq.Model, requestID)
	}
}

// convertToDeepSeekRequest 将OpenAI请求转换为DeepSeek格式
// 这个函数是翻译过程的核心，处理两种API格式之间的所有差异
func (ps *ProxyServer) convertToDeepSeekRequest(openaiReq ChatRequest, requestID string) (*DeepSeekRequest, error) {
	log.Printf("[%s] 开始转换请求格式", requestID)
	
	// 映射模型名称，将OpenAI的模型名转换为DeepSeek的模型名
	deepseekModel := MapModelName(openaiReq.Model)
	log.Printf("[%s] 模型映射: %s -> %s", requestID, openaiReq.Model, deepseekModel)
	
	// 检查是否使用推理模型
	// 推理模型有特殊的处理需求，我们需要记录这个信息
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
	// 工具调用让AI可以调用外部函数来获取信息或执行操作
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
func (ps *ProxyServer) handleNormalResponse(w http.ResponseWriter, r *http.Request, 
	deepseekReq *DeepSeekRequest, originalModel, requestID string) {
	
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

// sendRequestToDeepSeek 向DeepSeek API发送普通请求
// 这个函数负责与DeepSeek API的实际通信
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
	
	// 设置请求头部
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+ps.config.DeepSeekAPIKey)
	httpReq.Header.Set("User-Agent", "DeepSeek-Proxy/1.0.0")
	
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
		return nil, fmt.Errorf("DeepSeek API返回错误 %d: %s", resp.StatusCode, string(body))
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
	
	// 设置流式请求的头部
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+ps.config.DeepSeekAPIKey)
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("User-Agent", "DeepSeek-Proxy/1.0.0")
	
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

// processStreamingData 处理流式数据
// 这个函数负责读取DeepSeek的流式响应并转换为OpenAI格式
func (ps *ProxyServer) processStreamingData(w http.ResponseWriter, reader io.Reader, 
	flusher http.Flusher, originalModel, requestID string, ctx context.Context) {
	
	log.Printf("[%s] 开始处理流式数据", requestID)
	
	// 这是一个简化的流式处理实现
	// 在实际应用中，你需要正确解析Server-Sent Events格式
	
	buffer := make([]byte, 4096)
	for {
		select {
		case <-ctx.Done():
			log.Printf("[%s] 客户端连接已断开", requestID)
			return
		default:
			// 读取数据
			n, err := reader.Read(buffer)
			if err != nil {
				if err == io.EOF {
					log.Printf("[%s] 流式数据读取完成", requestID)
				} else {
					log.Printf("[%s] 流式数据读取错误: %v", requestID, err)
				}
				return
			}
			
			// 写入数据到客户端
			if _, writeErr := w.Write(buffer[:n]); writeErr != nil {
				log.Printf("[%s] 写入流式数据失败: %v", requestID, writeErr)
				return
			}
			
			// 立即发送数据
			flusher.Flush()
		}
	}
}

// convertToOpenAIResponse 将DeepSeek响应转换为OpenAI格式
// 这确保客户端接收到的响应格式与OpenAI API完全兼容
func (ps *ProxyServer) convertToOpenAIResponse(deepseekResp *DeepSeekResponse, originalModel, requestID string) map[string]interface{} {
	log.Printf("[%s] 转换响应格式", requestID)
	
	// 检查是否是推理模型的响应
	// 推理模型的响应可能包含reasoning_content字段
	isReasoningModel := originalModel == "deepseek-reasoner" || originalModel == "o1" || originalModel == "o1-preview" || originalModel == "o1-mini"
	
	// 创建OpenAI格式的响应
	openaiResp := map[string]interface{}{
		"id":      deepseekResp.ID,
		"object":  "chat.completion",
		"created": deepseekResp.Created,
		"model":   originalModel, // 使用客户端请求的模型名
		"choices": deepseekResp.Choices,
		"usage":   deepseekResp.Usage,
	}
	
	// 如果是推理模型，我们需要特殊处理reasoning_content
	if isReasoningModel {
		log.Printf("[%s] 处理推理模型响应，检查reasoning_content", requestID)
		
		// 检查choices中是否包含reasoning_content
		if choices, ok := deepseekResp.Choices.([]interface{}); ok {
			for i, choice := range choices {
				if choiceMap, ok := choice.(map[string]interface{}); ok {
					if message, ok := choiceMap["message"].(map[string]interface{}); ok {
						// 如果存在reasoning_content，确保它被正确传递
						if reasoningContent, exists := message["reasoning_content"]; exists {
							log.Printf("[%s] 发现推理内容，长度: %d字符", requestID, len(reasoningContent.(string)))
							// reasoning_content会自动保留在响应中
						} else {
							log.Printf("[%s] 该推理模型响应不包含reasoning_content", requestID)
						}
					}
				}
			}
		}
	}
	
	log.Printf("[%s] 响应格式转换完成", requestID)
	return openaiResp
}

// handleModels 处理模型列表请求
// 返回我们代理服务器支持的所有模型列表
func (ps *ProxyServer) handleModels(w http.ResponseWriter, r *http.Request) {
	logRequest(r, "模型列表")
	
	// 设置CORS头部
	ps.handleCORS(w, r)
	
	if r.Method == "OPTIONS" {
		return
	}
	
	// 只接受GET请求
	if r.Method != "GET" {
		handleError(w, fmt.Errorf("不支持的请求方法: %s", r.Method), 
			http.StatusMethodNotAllowed, "方法检查")
		return
	}
	
	log.Printf("返回支持的模型列表")
	
	// 创建模型列表响应
	models := GetSupportedModels()
	modelsData := make([]Model, len(models))
	
	currentTime := time.Now().Unix()
	for i, modelName := range models {
		modelsData[i] = Model{
			ID:      modelName,
			Object:  "model",
			Created: currentTime,
			OwnedBy: "deepseek-proxy",
		}
	}
	
	response := ModelsResponse{
		Object: "list",
		Data:   modelsData,
	}
	
	// 返回响应
	if err := writeJSONResponse(w, response); err != nil {
		log.Printf("写入模型列表响应失败: %v", err)
		return
	}
	
	log.Printf("模型列表返回成功，共 %d 个模型", len(models))
}