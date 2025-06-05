package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ReasonerTestClient 专门用于测试DeepSeek-Reasoner功能的客户端
// 这个客户端设计用来验证推理模型的特殊功能，包括推理过程的展示
type ReasonerTestClient struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

// NewReasonerTestClient 创建推理模型测试客户端
func NewReasonerTestClient(baseURL, apiKey string) *ReasonerTestClient {
	return &ReasonerTestClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		client: &http.Client{
			Timeout: 120 * time.Second, // 推理模型需要更长时间思考
		},
	}
}

// TestMathReasoning 测试数学推理能力
// 这个测试验证推理模型在复杂数学问题上的表现
func (rtc *ReasonerTestClient) TestMathReasoning() error {
	fmt.Println("🧮 测试数学推理能力...")

	// 选择一个需要多步推理的数学问题
	mathProblem := `解决这个数学问题：
	
一个圆形花园的直径是14米。园丁想在花园周围建造一条2米宽的小径。
请计算：
1. 原花园的面积
2. 包含小径后的总面积  
3. 小径本身的面积

请详细展示你的推理过程。`

	request := map[string]interface{}{
		"model": "deepseek-reasoner", // 明确使用推理模型
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": mathProblem,
			},
		},
		"max_tokens": 4000,
		"stream":     false,
	}

	response, err := rtc.sendRequest("/v1/chat/completions", request)
	if err != nil {
		return fmt.Errorf("数学推理测试失败: %w", err)
	}

	return rtc.analyzeReasoningResponse(response, "数学推理")
}

// TestLogicalPuzzle 测试逻辑推理能力
// 这个测试评估模型在复杂逻辑问题上的分析能力
func (rtc *ReasonerTestClient) TestLogicalPuzzle() error {
	fmt.Println("🧩 测试逻辑推理能力...")

	logicalPuzzle := `逻辑推理题：

有五个人（Alice、Bob、Charlie、Diana、Eve）坐成一排。已知：
1. Alice不坐在Bob旁边
2. Charlie坐在Diana的右边  
3. Eve不坐在任何一端
4. Bob坐在第二个位置
5. Diana不坐在Charlie旁边，除非Charlie在她右边

请找出每个人的准确位置，并解释你的推理过程。`

	request := map[string]interface{}{
		"model": "o1", // 测试o1模型映射
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": logicalPuzzle,
			},
		},
		"max_tokens": 3000,
		"stream":     false,
	}

	response, err := rtc.sendRequest("/v1/chat/completions", request)
	if err != nil {
		return fmt.Errorf("逻辑推理测试失败: %w", err)
	}

	return rtc.analyzeReasoningResponse(response, "逻辑推理")
}

// TestCodeDebugging 测试代码调试推理
// 验证模型在代码分析和问题诊断方面的推理能力
func (rtc *ReasonerTestClient) TestCodeDebugging() error {
	fmt.Println("🐛 测试代码调试推理...")

	codeDebugProblem := `分析下面的Python代码并找出问题：

def calculate_average(numbers):
    total = 0
    count = 0
    for num in numbers:
        total += num
        count += 1
    return total / count

# 测试用例
test_cases = [
    [1, 2, 3, 4, 5],
    [],
    [10, 20, 30],
    None
]

for case in test_cases:
    print(f"平均值: {calculate_average(case)}")

请：
1. 识别所有潜在问题
2. 解释为什么会出现这些问题
3. 提供修复建议
4. 展示你的分析思路`

	request := map[string]interface{}{
		"model": "o1-preview", // 测试另一个映射
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": codeDebugProblem,
			},
		},
		"max_tokens": 4000,
		"stream":     false,
	}

	response, err := rtc.sendRequest("/v1/chat/completions", request)
	if err != nil {
		return fmt.Errorf("代码调试测试失败: %w", err)
	}

	return rtc.analyzeReasoningResponse(response, "代码调试")
}

// analyzeReasoningResponse 分析推理响应的质量
// 这个函数检查响应是否包含预期的推理内容
func (rtc *ReasonerTestClient) analyzeReasoningResponse(responseBody []byte, testType string) error {
	var response map[string]interface{}
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return fmt.Errorf("解析%s响应失败: %w", testType, err)
	}

	// 检查响应结构
	choices, ok := response["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return fmt.Errorf("%s响应中没有找到choices", testType)
	}

	firstChoice, ok := choices[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("无法解析%s的第一个choice", testType)
	}

	message, ok := firstChoice["message"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("无法解析%s的message", testType)
	}

	// 检查是否包含推理过程
	reasoningContent, hasReasoning := message["reasoning_content"].(string)
	finalContent, hasFinal := message["content"].(string)

	fmt.Printf("✅ %s测试结果分析:\n", testType)

	if hasReasoning {
		fmt.Printf("🧠 推理过程长度: %d 字符\n", len(reasoningContent))
		fmt.Printf("🎯 推理过程预览: %s...\n", truncateString(reasoningContent, 200))
	} else {
		fmt.Printf("⚠️  警告：没有找到推理过程内容\n")
	}

	if hasFinal {
		fmt.Printf("📝 最终答案长度: %d 字符\n", len(finalContent))
		fmt.Printf("💡 最终答案预览: %s...\n", truncateString(finalContent, 300))
	} else {
		return fmt.Errorf("%s响应中没有最终答案", testType)
	}

	// 验证推理质量指标
	if hasReasoning && len(reasoningContent) > 100 {
		fmt.Printf("✨ 推理质量：优秀（包含详细思考过程）\n")
	} else if hasReasoning {
		fmt.Printf("⚡ 推理质量：简化（推理过程较短）\n")
	} else {
		fmt.Printf("❓ 推理质量：未知（缺少推理过程）\n")
	}

	fmt.Println()
	return nil
}

// sendRequest 发送HTTP请求到代理服务器
func (rtc *ReasonerTestClient) sendRequest(endpoint string, data interface{}) ([]byte, error) {
	reqBody, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequest("POST", rtc.baseURL+endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+rtc.apiKey)

	resp, err := rtc.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("发送HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	return body, nil
}
