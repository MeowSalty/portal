package responses

import "github.com/MeowSalty/portal/request/adapter/openai/types/shared"

// Request 表示 OpenAI Responses API 请求
// 仅保留当前需要的字段，其他字段通过 ExtraFields 透传。
type Request struct {
	Model           string                  `json:"model"`
	Input           interface{}             `json:"input"`
	Stream          *bool                   `json:"stream,omitempty"`
	MaxOutputTokens *int                    `json:"max_output_tokens,omitempty"`
	Temperature     *float64                `json:"temperature,omitempty"`
	TopP            *float64                `json:"top_p,omitempty"`
	Tools           []shared.ToolUnion      `json:"tools,omitempty"`
	ToolChoice      *shared.ToolChoiceUnion `json:"tool_choice,omitempty"`
	ExtraFields     map[string]interface{}  `json:"-"`
}

// InputItem 表示 Responses API 的输入项
// 使用 messages 形式时，input 为该结构数组。
type InputItem struct {
	Role    string      `json:"role"`
	Content []InputPart `json:"content"`
}

// InputPart 表示 Responses API 的输入内容片段
// 支持 input_text 与 input_image。
type InputPart struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

// ImageURL 表示 Responses API 输入图像信息
type ImageURL struct {
	URL    string  `json:"url"`
	Detail *string `json:"detail,omitempty"`
}
