package main

// === OpenAI兼容的请求结构 ===
// 这是客户端（如Cursor）发送给我们的请求格式
type ChatRequest struct {
	Model       string      `json:"model"`                 // 请求的模型名称，如"gpt-4"
	Messages    []Message   `json:"messages"`              // 对话消息列表
	Stream      bool        `json:"stream"`                // 是否启用流式响应
	Temperature *float64    `json:"temperature,omitempty"` // 控制响应的随机性，0-2之间
	MaxTokens   *int        `json:"max_tokens,omitempty"`  // 最大生成的token数量
	Tools       []Tool      `json:"tools,omitempty"`       // 可调用的工具函数列表
	ToolChoice  interface{} `json:"tool_choice,omitempty"` // 工具选择策略
	Functions   []Function  `json:"functions,omitempty"`   // 兼容旧版本的functions格式
}

// === 消息结构 ===
// 对话中的每一条消息，包含角色和内容
type Message struct {
	Role             string     `json:"role"`                       // 消息角色：user/assistant/system/tool
	Content          string     `json:"content"`                    // 消息的实际内容
	ReasoningContent string     `json:"reasoning_content,omitempty"` // DeepSeek-Reasoner的推理过程内容
	ToolCalls        []ToolCall `json:"tool_calls,omitempty"`       // AI助手调用的工具列表
	ToolCallID       string     `json:"tool_call_id,omitempty"`     // 工具响应消息的关联ID
	Name             string     `json:"name,omitempty"`             // 消息发送者的名称
}

// === 工具相关结构 ===
// 定义AI可以调用的外部函数
type Tool struct {
	Type     string   `json:"type"`     // 工具类型，通常是"function"
	Function Function `json:"function"` // 具体的函数定义
}

// 函数的详细定义
type Function struct {
	Name        string      `json:"name"`        // 函数名称
	Description string      `json:"description"` // 函数的用途描述
	Parameters  interface{} `json:"parameters"`  // 函数参数的JSON Schema定义
}

// AI调用工具时的具体调用信息
type ToolCall struct {
	ID       string `json:"id"`   // 工具调用的唯一标识符
	Type     string `json:"type"` // 调用类型，通常是"function"
	Function struct {
		Name      string `json:"name"`      // 要调用的函数名
		Arguments string `json:"arguments"` // JSON格式的函数参数
	} `json:"function"`
}

// === DeepSeek API特定结构 ===
// 发送给DeepSeek API的请求格式
type DeepSeekRequest struct {
	Model       string    `json:"model"`                 // DeepSeek的模型名称
	Messages    []Message `json:"messages"`              // 转换后的消息列表
	Stream      bool      `json:"stream"`                // 是否启用流式响应
	Temperature float64   `json:"temperature,omitempty"` // 温度参数
	MaxTokens   int       `json:"max_tokens,omitempty"`  // 最大token数
	Tools       []Tool    `json:"tools,omitempty"`       // 工具定义
	ToolChoice  string    `json:"tool_choice,omitempty"` // 工具选择策略字符串
}

// DeepSeek API的响应结构
type DeepSeekResponse struct {
	ID      string `json:"id"`      // 响应唯一标识
	Object  string `json:"object"`  // 对象类型
	Created int64  `json:"created"` // 创建时间戳
	Model   string `json:"model"`   // 使用的模型
	Choices []struct {
		Index        int     `json:"index"`         // 选择索引
		Message      Message `json:"message"`       // 生成的消息
		FinishReason string  `json:"finish_reason"` // 完成原因
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`     // 输入使用的token数
		CompletionTokens int `json:"completion_tokens"` // 输出使用的token数
		TotalTokens      int `json:"total_tokens"`      // 总计token数
	} `json:"usage"`
}

// === 模型列表相关结构 ===
// 模型信息
type Model struct {
	ID      string `json:"id"`       // 模型ID
	Object  string `json:"object"`   // 对象类型，通常是"model"
	Created int64  `json:"created"`  // 创建时间
	OwnedBy string `json:"owned_by"` // 拥有者
}

// 模型列表响应
type ModelsResponse struct {
	Object string  `json:"object"` // 对象类型，通常是"list"
	Data   []Model `json:"data"`   // 模型列表
}

// === 配置管理结构 ===
// 代理服务器的全局配置
type ProxyConfig struct {
	Port          int    `json:"port"`           // 监听端口
	DeepSeekAPIKey string `json:"deepseek_key"`   // DeepSeek API密钥
	DeepSeekModel  string `json:"deepseek_model"` // 默认使用的DeepSeek模型
	Endpoint      string `json:"endpoint"`       // DeepSeek API端点
}

// === 流式响应结构 ===
// 用于处理流式响应的数据块
type StreamChunk struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}