package types

import (
	"encoding/json"
)

// Content 表示内容部分
type Content struct {
	Role  string `json:"role,omitempty"` // 角色：user 或 model
	Parts []Part `json:"parts"`          // 内容部分
}

// Part 表示内容部分，可以是文本或内联数据
type Part struct {
	Text       *string     `json:"text,omitempty"`       // 文本内容
	InlineData *InlineData `json:"inlineData,omitempty"` // 内联数据（如图像）
	// 函数调用响应（用于工具调用）
	FunctionResponse *FunctionResponse `json:"functionResponse,omitempty"`
	// 函数调用（模型生成的工具调用）
	FunctionCall *FunctionCall `json:"functionCall,omitempty"`
	// URI 形式的数据
	FileData *FileData `json:"fileData,omitempty"`
	// 可执行代码
	ExecutableCode *ExecutableCode `json:"executableCode,omitempty"`
	// 代码执行结果
	CodeExecutionResult *CodeExecutionResult `json:"codeExecutionResult,omitempty"`
	// 视频元数据
	VideoMetadata *VideoMetadata `json:"videoMetadata,omitempty"`
	// 是否为思考内容
	Thought *bool `json:"thought,omitempty"`
	// 思考签名
	ThoughtSignature *string `json:"thoughtSignature,omitempty"`
	// Part 自定义元数据
	PartMetadata map[string]interface{} `json:"partMetadata,omitempty"`
	// 媒体分辨率
	MediaResolution *MediaResolution `json:"mediaResolution,omitempty"`
}

// InlineData 表示内联数据
type InlineData struct {
	MimeType string `json:"mimeType"` // MIME 类型
	Data     string `json:"data"`     // Base64 编码的数据
}

// FileData 表示 URI 数据
type FileData struct {
	MimeType *string `json:"mimeType,omitempty"` // MIME 类型
	FileURI  string  `json:"fileUri"`            // 文件 URI
}

// ExecutableCode 表示可执行代码
type ExecutableCode struct {
	Language string `json:"language"` // 语言
	Code     string `json:"code"`     // 代码内容
}

// CodeExecutionResult 表示代码执行结果
type CodeExecutionResult struct {
	Outcome string `json:"outcome"`          // 执行结果
	Output  string `json:"output,omitempty"` // 标准输出或错误输出
}

// VideoMetadata 表示视频元数据
type VideoMetadata struct {
	StartOffset *string  `json:"startOffset,omitempty"` // 开始偏移
	EndOffset   *string  `json:"endOffset,omitempty"`   // 结束偏移
	FPS         *float64 `json:"fps,omitempty"`         // 帧率
}

// MediaResolution 表示媒体分辨率
type MediaResolution struct {
	Level string `json:"level,omitempty"` // 分辨率级别
}

// MarshalJSON 实现 Part 的自定义 JSON 序列化
func (p Part) MarshalJSON() ([]byte, error) {
	result := make(map[string]interface{})

	// 根据内容类型确定序列化方式
	if p.Text != nil {
		result["text"] = *p.Text
	} else if p.InlineData != nil {
		result["inlineData"] = p.InlineData
	} else if p.FunctionResponse != nil {
		result["functionResponse"] = p.FunctionResponse
	} else if p.FunctionCall != nil {
		result["functionCall"] = p.FunctionCall
	} else if p.FileData != nil {
		result["fileData"] = p.FileData
	} else if p.ExecutableCode != nil {
		result["executableCode"] = p.ExecutableCode
	} else if p.CodeExecutionResult != nil {
		result["codeExecutionResult"] = p.CodeExecutionResult
	}

	if p.VideoMetadata != nil {
		result["videoMetadata"] = p.VideoMetadata
	}
	if p.Thought != nil {
		result["thought"] = *p.Thought
	}
	if p.ThoughtSignature != nil {
		result["thoughtSignature"] = *p.ThoughtSignature
	}
	if p.PartMetadata != nil {
		result["partMetadata"] = p.PartMetadata
	}
	if p.MediaResolution != nil {
		result["mediaResolution"] = p.MediaResolution
	}

	if len(result) == 0 {
		return json.Marshal(struct{}{})
	}

	return json.Marshal(result)
}
