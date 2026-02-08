package responses

import (
	"encoding/json"

	portalErrors "github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/request/adapter/openai/types/shared"
)

// InputContent 表示输入内容联合类型
// 支持文本、图片、文件。
//
// 强类型约束：InputContent 只能承载一种具体类型，使用 oneof 指针字段实现互斥。
type InputContent struct {
	// Text 表示文本输入内容（type == "input_text"）
	Text *InputTextContent `json:"-"`
	// Image 表示图片输入内容（type == "input_image"）
	Image *InputImageContent `json:"-"`
	// File 表示文件输入内容（type == "input_file"）
	File *InputFileContent `json:"-"`
}

// MarshalJSON 实现 InputContent 的自定义 JSON 序列化
// 严格 oneof：只能设置一种类型，否则返回错误
func (c InputContent) MarshalJSON() ([]byte, error) {
	// 统计非空字段数量
	count := 0
	if c.Text != nil {
		count++
	}
	if c.Image != nil {
		count++
	}
	if c.File != nil {
		count++
	}

	// count==0 序列化为 null
	if count == 0 {
		return json.Marshal(nil)
	}

	// count>1 返回错误
	if count > 1 {
		return nil, portalErrors.New(portalErrors.ErrCodeInvalidArgument, "input content 只能设置一种类型")
	}

	// count==1 直接序列化该字段
	switch {
	case c.Text != nil:
		return json.Marshal(c.Text)
	case c.Image != nil:
		return json.Marshal(c.Image)
	case c.File != nil:
		return json.Marshal(c.File)
	default:
		return json.Marshal(nil)
	}
}

// UnmarshalJSON 实现 InputContent 的自定义 JSON 反序列化
// 严格 oneof：必须有 type 字段，按 type 分支反序列化到对应字段
func (c *InputContent) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	typeVal, ok := raw["type"].(string)
	if !ok || typeVal == "" {
		return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "input content 缺少 type 字段")
	}

	*c = InputContent{}

	switch typeVal {
	case InputContentTypeText:
		var text InputTextContent
		if err := json.Unmarshal(data, &text); err != nil {
			return err
		}
		c.Text = &text
	case InputContentTypeImage:
		var image InputImageContent
		if err := json.Unmarshal(data, &image); err != nil {
			return err
		}
		c.Image = &image
	case InputContentTypeFile:
		var file InputFileContent
		if err := json.Unmarshal(data, &file); err != nil {
			return err
		}
		c.File = &file
	default:
		return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "input content 类型不支持："+typeVal)
	}

	return nil
}

// InputTextContent 表示文本输入内容
type InputTextContent struct {
	Type InputContentType `json:"type"`
	Text string           `json:"text"`
}

// InputImageContent 表示图片输入内容
type InputImageContent struct {
	Type     InputContentType    `json:"type"`
	ImageURL *string             `json:"image_url,omitempty"`
	FileID   *string             `json:"file_id,omitempty"`
	Detail   *shared.ImageDetail `json:"detail"`
}

// MarshalJSON 实现 InputImageContent 的自定义 JSON 序列化
// 确保 detail 字段在为空时序列化为 "auto"
func (c InputImageContent) MarshalJSON() ([]byte, error) {
	// 创建一个临时结构体用于序列化
	type Alias InputImageContent

	// 如果 Detail 为空，设置为 auto
	if c.Detail == nil {
		auto := shared.ImageDetailAuto
		c.Detail = &auto
	}

	return json.Marshal((Alias)(c))
}

// UnmarshalJSON 实现 InputImageContent 的自定义 JSON 反序列化
// 确保 detail 字段在缺失时设置为 auto
func (c *InputImageContent) UnmarshalJSON(data []byte) error {
	type Alias InputImageContent
	alias := (*Alias)(c)

	if err := json.Unmarshal(data, alias); err != nil {
		return err
	}

	// 如果 Detail 为空，设置为 auto
	if c.Detail == nil {
		auto := shared.ImageDetailAuto
		c.Detail = &auto
	}

	return nil
}

// InputFileContent 表示文件输入内容
type InputFileContent struct {
	Type     InputContentType `json:"type"`
	FileID   *string          `json:"file_id,omitempty"`
	Filename *string          `json:"filename,omitempty"`
	FileData *string          `json:"file_data,omitempty"`
	FileURL  *string          `json:"file_url,omitempty"`
}
