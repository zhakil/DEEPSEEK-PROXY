package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TestClient 测试客户端结构
type TestClient struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

// NewTestClient 创建新的测试客户端
func NewTestClient(baseURL, apiKey string) *TestClient {
	return &TestClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// TestChatCompletion 测试聊天完成功能
func (tc *TestClient) TestChatCompletion() error {
	fmt.Println("🧪 测试聊天完成功能...")

	// 创建测试请求
	request := map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]interface{}{
			{
				"role":    "system",
				"content": "你是一个有用的AI助手。请用中文回答问题。",
			},
			{
				"role":    "user",
				"content": "请介绍一下Go语言的特点，用简洁的语言回答。",
			},
		},
		"temperature": 0.7,
		"max_tokens":  500,
		"stream":      false,
	}

	// 发送请求
	response, err := tc.sendRequest("/v1/chat/completions", request)
	if err != nil {
		return fmt.Errorf("发送聊天完成请求失败: %w", err)
	}

	// 解析响应
	var chatResponse map[string]interface{}
	if err := json.Unmarshal(response, &chatResponse); err != nil {
		return fmt.Errorf("解析聊天响应失败: %w", err)
	}

	// 提取AI回复
	choices, ok := chatResponse["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return fmt.Errorf("响应中没有找到choices")
	}

	firstChoice, ok := choices[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("无法解析第一个choice")
	}

	message, ok := firstChoice["message"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("无法解析message")
	}

	content, ok := message["content"].(string)
	if !ok {
		return fmt.Errorf("无法解析content")
	}

	fmt.Printf("✅ 聊天完成测试成功！\n")
	fmt.Printf("🤖 AI回复: %s\n\n", content)

	return nil
}

// TestStreamingCompletion 测试流式聊天完成
func (tc *TestClient) TestStreamingCompletion() error {
	fmt.Println("🌊 测试流式聊天完成功能...")

	// 创建流式测试请求
	request := map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": "请写一首关于编程的短诗。",
			},
		},
		"temperature": 0.8,
		"max_tokens":  200,
		"stream":      true,
	}

	// 发送流式请求
	fmt.Printf("📡 正在发送流式请求...\n")

	// 创建HTTP请求
	reqBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequest("POST", tc.baseURL+"/v1/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+tc.apiKey)
	httpReq.Header.Set("Accept", "text/event-stream")

	// 发送请求
	resp, err := tc.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("发送流式请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("流式请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("🎯 开始接收流式响应:\n")
	fmt.Printf("💭 ")

	// 读取流式响应（简化处理）
	buffer := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("读取流式响应失败: %w", err)
		}

		// 简单地输出接收到的数据（在实际应用中需要解析SSE格式）
		fmt.Printf("%s", string(buffer[:n]))
	}

	fmt.Printf("\n✅ 流式聊天完成测试成功！\n\n")
	return nil
}

// TestModels 测试模型列表功能
func (tc *TestClient) TestModels() error {
	fmt.Println("📋 测试模型列表功能...")

	// 发送GET请求获取模型列表
	httpReq, err := http.NewRequest("GET", tc.baseURL+"/v1/models", nil)
	if err != nil {
		return fmt.Errorf("创建模型列表请求失败: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+tc.apiKey)

	resp, err := tc.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("发送模型列表请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("模型列表请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取模型列表响应失败: %w", err)
	}

	// 解析响应
	var modelsResponse map[string]interface{}
	if err := json.Unmarshal(body, &modelsResponse); err != nil {
		return fmt.Errorf("解析模型列表响应失败: %w", err)
	}

	// 提取模型列表
	data, ok := modelsResponse["data"].([]interface{})
	if !ok {
		return fmt.Errorf("无法解析模型数据")
	}

	fmt.Printf("✅ 模型列表测试成功！找到 %d 个模型:\n", len(data))
	for i, model := range data {
		if modelMap, ok := model.(map[string]interface{}); ok {
			if id, ok := modelMap["id"].(string); ok {
				fmt.Printf("  %d. %s\n", i+1, id)
			}
		}
	}
	fmt.Println()

	return nil
}

// TestHealth 测试健康检查功能
func (tc *TestClient) TestHealth() error {
	fmt.Println("❤️  测试健康检查功能...")

	httpReq, err := http.NewRequest("GET", tc.baseURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("创建健康检查请求失败: %w", err)
	}

	resp, err := tc.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("发送健康检查请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("健康检查失败，状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取健康检查响应失败: %w", err)
	}

	var healthResponse map[string]interface{}
	if err := json.Unmarshal(body, &healthResponse); err != nil {
		return fmt.Errorf("解析健康检查响应失败: %w", err)
	}

	status, ok := healthResponse["status"].(string)
	if !ok || status != "healthy" {
		return fmt.Errorf("服务器状态不健康: %v", status)
	}

	fmt.Printf("✅ 健康检查测试成功！服务器状态: %s\n\n", status)
	return nil
}

// sendRequest 发送通用请求
func (tc *TestClient) sendRequest(endpoint string, data interface{}) ([]byte, error) {
	// 序列化请求数据
	reqBody, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequest("POST", tc.baseURL+endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+tc.apiKey)

	// 发送请求
	resp, err := tc.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("发送HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	return body, nil
}


