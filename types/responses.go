package types

import (
	"encoding/json"
)

// Response 表示 API 响应结构
type Response struct {
	ID                string   `json:"id"`                           // 响应 ID
	Choices           []Choice `json:"choices"`                      // 根据是否设置 "stream" 为 "true" 以及传入的是 "messages" 还是 "prompt"，会得到不同的输出形状
	Created           int64    `json:"created"`                      // Unix 时间戳
	Model             string   `json:"model"`                        // 模型名称
	Object            string   `json:"object"`                       // 对象类型：'chat.completion' 或 'chat.completion.chunk'
	SystemFingerprint *string  `json:"system_fingerprint,omitempty"` // 仅当提供商支持时存在

	// 使用情况数据总是为非流式返回。
	// 在流式传输时，您将在最后获得一个使用情况对象，伴随一个空的 choices 数组。
	Usage *ResponseUsage `json:"usage,omitempty"`
}

// ResponseUsage 表示 API 使用情况统计
type ResponseUsage struct {
	// 包括图像和工具（如果有）
	PromptTokens int `json:"prompt_tokens"`
	// 生成的令牌数
	CompletionTokens int `json:"completion_tokens"`
	// 上述两个字段的总和
	TotalTokens int `json:"total_tokens"`
}

// Choice 表示选择项，可以是非聊天选择、非流式选择或流式选择
type Choice struct {
	FinishReason       *string          `json:"finish_reason,omitempty"`        // 完成原因
	NativeFinishReason *string          `json:"native_finish_reason,omitempty"` // 原生完成原因
	Text               *string          `json:"text,omitempty"`                 // 文本内容（仅用于非聊天选择）
	Message            *ResponseMessage `json:"message,omitempty"`              // 消息内容（仅用于非流式选择）
	Delta              *Delta           `json:"delta,omitempty"`                // 增量内容（仅用于流式选择）
	Error              *ErrorResponse   `json:"error,omitempty"`                // 错误信息
}

// Message 表示非流式选择中的消息
type ResponseMessage struct {
	Content   *string    `json:"content,omitempty"`    // 消息内容
	Role      string     `json:"role"`                 // 角色
	ToolCalls []ToolCall `json:"tool_calls,omitempty"` // 工具调用
}

// Delta 表示流式选择中的增量内容
type Delta struct {
	Content   *string    `json:"content,omitempty"`    // 增量内容
	Role      *string    `json:"role,omitempty"`       // 角色（可选）
	ToolCalls []ToolCall `json:"tool_calls,omitempty"` // 工具调用
}

// ErrorResponse 表示错误响应
type ErrorResponse struct {
	Code     int                    `json:"code"`               // 错误代码，参见 "Error Handling" 部分
	Message  string                 `json:"message"`            // 错误消息
	Metadata map[string]interface{} `json:"metadata,omitempty"` // 包含额外的错误信息，如提供商详细信息、原始错误消息等
}

// ToolCall 表示工具调用
type ToolCall struct {
	ID       string       `json:"id"`       // 工具调用 ID
	Type     string       `json:"type"`     // 类型，固定为 'function'
	Function FunctionCall `json:"function"` // 函数调用
}

// FunctionCall 表示函数调用
type FunctionCall struct {
	Name      string `json:"name"`      // 函数名称
	Arguments string `json:"arguments"` // 函数参数
}

// MarshalJSON 实现 Choice 的自定义 JSON 序列化
func (c Choice) MarshalJSON() ([]byte, error) {
	type Alias Choice
	// aux := &struct {
	// 	*Alias
	// }{
	// 	Alias: (*Alias)(&c),
	// }

	// 根据内容类型确定如何序列化
	if c.Text != nil {
		// 非聊天选择
		return json.Marshal(struct {
			FinishReason *string        `json:"finish_reason,omitempty"`
			Text         *string        `json:"text,omitempty"`
			Error        *ErrorResponse `json:"error,omitempty"`
		}{
			FinishReason: c.FinishReason,
			Text:         c.Text,
			Error:        c.Error,
		})
	} else if c.Message != nil {
		// 非流式选择
		return json.Marshal(struct {
			FinishReason       *string          `json:"finish_reason,omitempty"`
			NativeFinishReason *string          `json:"native_finish_reason,omitempty"`
			Message            *ResponseMessage `json:"message,omitempty"`
			Error              *ErrorResponse   `json:"error,omitempty"`
		}{
			FinishReason:       c.FinishReason,
			NativeFinishReason: c.NativeFinishReason,
			Message:            c.Message,
			Error:              c.Error,
		})
	} else if c.Delta != nil {
		// 流式选择
		return json.Marshal(struct {
			FinishReason       *string        `json:"finish_reason,omitempty"`
			NativeFinishReason *string        `json:"native_finish_reason,omitempty"`
			Delta              *Delta         `json:"delta,omitempty"`
			Error              *ErrorResponse `json:"error,omitempty"`
		}{
			FinishReason:       c.FinishReason,
			NativeFinishReason: c.NativeFinishReason,
			Delta:              c.Delta,
			Error:              c.Error,
		})
	} else {
		// 默认序列化
		return json.Marshal(struct {
			FinishReason       *string          `json:"finish_reason,omitempty"`
			NativeFinishReason *string          `json:"native_finish_reason,omitempty"`
			Text               *string          `json:"text,omitempty"`
			Message            *ResponseMessage `json:"message,omitempty"`
			Delta              *Delta           `json:"delta,omitempty"`
			Error              *ErrorResponse   `json:"error,omitempty"`
		}{
			FinishReason:       c.FinishReason,
			NativeFinishReason: c.NativeFinishReason,
			Text:               c.Text,
			Message:            c.Message,
			Delta:              c.Delta,
			Error:              c.Error,
		})
	}

	// return json.Marshal(aux)
}

// UnmarshalJSON 实现 Choice 的自定义 JSON 反序列化
func (c *Choice) UnmarshalJSON(data []byte) error {
	// 首先尝试解析为非聊天选择
	var nonChat struct {
		FinishReason *string        `json:"finish_reason"`
		Text         *string        `json:"text"`
		Error        *ErrorResponse `json:"error"`
	}
	if err := json.Unmarshal(data, &nonChat); err == nil && nonChat.Text != nil {
		c.FinishReason = nonChat.FinishReason
		c.Text = nonChat.Text
		c.Error = nonChat.Error
		return nil
	}

	// 尝试解析为非流式选择
	var nonStreaming struct {
		FinishReason       *string          `json:"finish_reason"`
		NativeFinishReason *string          `json:"native_finish_reason"`
		Message            *ResponseMessage `json:"message"`
		Error              *ErrorResponse   `json:"error"`
	}
	if err := json.Unmarshal(data, &nonStreaming); err == nil && nonStreaming.Message != nil {
		c.FinishReason = nonStreaming.FinishReason
		c.NativeFinishReason = nonStreaming.NativeFinishReason
		c.Message = nonStreaming.Message
		c.Error = nonStreaming.Error
		return nil
	}

	// 尝试解析为流式选择
	var streaming struct {
		FinishReason       *string        `json:"finish_reason"`
		NativeFinishReason *string        `json:"native_finish_reason"`
		Delta              *Delta         `json:"delta"`
		Error              *ErrorResponse `json:"error"`
	}
	if err := json.Unmarshal(data, &streaming); err == nil && streaming.Delta != nil {
		c.FinishReason = streaming.FinishReason
		c.NativeFinishReason = streaming.NativeFinishReason
		c.Delta = streaming.Delta
		c.Error = streaming.Error
		return nil
	}

	return nil
}
