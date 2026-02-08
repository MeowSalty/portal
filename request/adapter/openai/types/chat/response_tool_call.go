package chat

// ToolCallType 表示工具调用类型
type ToolCallType string

const (
	ToolCallTypeFunction ToolCallType = "function"
	ToolCallTypeCustom   ToolCallType = "custom"
)

// MessageToolCall 表示工具调用（非流式）
type MessageToolCall struct {
	ID       string            `json:"id"`
	Type     ToolCallType      `json:"type"`
	Function *ToolCallFunction `json:"function,omitempty"`
	Custom   *ToolCallCustom   `json:"custom,omitempty"`
}

// ToolCallCustom 表示自定义工具调用信息
// 对应 type=custom。
type ToolCallCustom struct {
	Name  string `json:"name"`
	Input string `json:"input"`
}

// ToolCallFunction 表示工具调用函数（非流式）
type ToolCallFunction struct {
	Arguments string `json:"arguments"` // 参数
	Name      string `json:"name"`      // 名称
}

// ToolCallChunk 表示工具调用（流式增量）
type ToolCallChunk struct {
	Index    int                    `json:"index"`
	ID       *string                `json:"id,omitempty"`
	Type     *ToolCallType          `json:"type,omitempty"`
	Function *ToolCallChunkFunction `json:"function,omitempty"`
}

// ToolCallChunkFunction 表示工具调用函数（流式增量）
type ToolCallChunkFunction struct {
	Arguments *string `json:"arguments,omitempty"` // 参数（流式响应中逐步累积）
	Name      *string `json:"name,omitempty"`      // 名称（流式响应中可选）
}
