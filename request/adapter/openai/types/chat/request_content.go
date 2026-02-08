package chat

import (
	"encoding/json"

	"github.com/MeowSalty/portal/request/adapter/openai/types/shared"
)

// MessageContent 表示消息内容
type MessageContent struct {
	StringValue  *string
	ContentParts []ContentPart
}

// ContentPartType 表示消息内容片段类型
type ContentPartType = string

const (
	ContentPartTypeText       ContentPartType = "text"
	ContentPartTypeImageURL   ContentPartType = "image_url"
	ContentPartTypeInputAudio ContentPartType = "input_audio"
	ContentPartTypeFile       ContentPartType = "file"
	ContentPartTypeRefusal    ContentPartType = "refusal"
)

// ContentPart 表示消息内容片段
type ContentPart struct {
	Type       ContentPartType `json:"type"`
	Text       *string         `json:"text,omitempty"`
	Refusal    *string         `json:"refusal,omitempty"`
	ImageURL   *ImageURL       `json:"image_url,omitempty"`
	InputAudio *InputAudio     `json:"input_audio,omitempty"`
	File       *InputFile      `json:"file,omitempty"`
}

// ImageURL 表示图像 URL 信息
type ImageURL struct {
	URL    string              `json:"url"`
	Detail *shared.ImageDetail `json:"detail,omitempty"`
}

// AudioFormat 表示输入音频格式
// 仅支持 wav 或 mp3。
type AudioFormat = string

const (
	AudioFormatWav AudioFormat = "wav"
	AudioFormatMP3 AudioFormat = "mp3"
)

// InputAudio 表示输入音频信息
type InputAudio struct {
	Data   string      `json:"data"`
	Format AudioFormat `json:"format"`
}

// InputFile 表示输入文件信息
type InputFile struct {
	Filename *string `json:"filename,omitempty"`
	FileData *string `json:"file_data,omitempty"`
	FileID   *string `json:"file_id,omitempty"`
}

// MarshalJSON 实现 MessageContent 的自定义 JSON 序列化
func (mc MessageContent) MarshalJSON() ([]byte, error) {
	if mc.StringValue != nil {
		return json.Marshal(mc.StringValue)
	}
	if mc.ContentParts != nil {
		return json.Marshal(mc.ContentParts)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON 实现 MessageContent 的自定义 JSON 反序列化
func (mc *MessageContent) UnmarshalJSON(data []byte) error {
	// 允许 content 为 null
	if string(data) == "null" {
		return nil
	}

	// 首先尝试反序列化为字符串
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		mc.StringValue = &str
		return nil
	}

	// 尝试反序列化为 ContentPart 数组
	var parts []ContentPart
	if err := json.Unmarshal(data, &parts); err == nil {
		mc.ContentParts = parts
		return nil
	}

	return nil
}
