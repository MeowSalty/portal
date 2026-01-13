package responses

import (
	"encoding/json"

	"github.com/MeowSalty/portal/request/adapter/openai/types/shared"
)

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
	Type     string                 `json:"type"`
	Text     string                 `json:"text,omitempty"`
	ImageURL *ImageURL              `json:"image_url,omitempty"`
	Raw      map[string]interface{} `json:"-"`
}

// MarshalJSON 实现 InputPart 的自定义 JSON 序列化
func (p InputPart) MarshalJSON() ([]byte, error) {
	if len(p.Raw) == 0 {
		type Alias InputPart
		aux := Alias(p)
		return json.Marshal(aux)
	}

	result := make(map[string]interface{})
	for key, value := range p.Raw {
		result[key] = value
	}

	if p.Type != "" {
		result["type"] = p.Type
	}
	if p.Text != "" {
		result["text"] = p.Text
	}
	if p.ImageURL != nil {
		imageURL := map[string]interface{}{
			"url": p.ImageURL.URL,
		}
		if p.ImageURL.Detail != nil {
			imageURL["detail"] = *p.ImageURL.Detail
		}
		result["image_url"] = imageURL
	}

	return json.Marshal(result)
}

// UnmarshalJSON 实现 InputPart 的自定义 JSON 反序列化
func (p *InputPart) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	type Alias InputPart
	aux := &struct{ *Alias }{Alias: (*Alias)(p)}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	if len(raw) > 0 {
		p.Raw = raw
	}

	return nil
}

// ImageURL 表示 Responses API 输入图像信息
type ImageURL struct {
	URL    string  `json:"url"`
	Detail *string `json:"detail,omitempty"`
}
