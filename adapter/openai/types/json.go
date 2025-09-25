package types

import (
	"encoding/json"
	"fmt"
)

// MarshalJSON 实现 json.Marshaler 接口
func (r *ChatCompletionRequest) MarshalJSON() ([]byte, error) {
	type Alias ChatCompletionRequest
	aux := &struct {
		Stop interface{} `json:"stop,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}

	if r.Stop != nil {
		aux.Stop = r.Stop.Value
	}

	return json.Marshal(aux)
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (r *ChatCompletionRequest) UnmarshalJSON(data []byte) error {
	type Alias ChatCompletionRequest
	aux := &struct {
		Stop json.RawMessage `json:"stop,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.Stop != nil {
		r.Stop = &StopField{}
		// 尝试解析为字符串
		var str string
		if err := json.Unmarshal(aux.Stop, &str); err == nil {
			r.Stop.Value = str
			return nil
		}

		// 尝试解析为字符串数组
		var strArray []string
		if err := json.Unmarshal(aux.Stop, &strArray); err == nil {
			r.Stop.Value = strArray
			return nil
		}

		// 如果解析失败，将原始数据存储为字符串
		r.Stop.Value = string(aux.Stop)
	}

	return nil
}

// StopField 是一个自定义类型，用于处理 Stop 字段，它可以是字符串或字符串数组
type StopField struct {
	Value interface{}
}

// MarshalJSON 实现 json.Marshaler 接口
func (s *StopField) MarshalJSON() ([]byte, error) {
	if s.Value == nil {
		return []byte("null"), nil
	}
	return json.Marshal(s.Value)
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (s *StopField) UnmarshalJSON(data []byte) error {
	// 尝试解析为字符串
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		s.Value = str
		return nil
	}

	// 尝试解析为字符串数组
	var strArray []string
	if err := json.Unmarshal(data, &strArray); err == nil {
		s.Value = strArray
		return nil
	}

	// 如果都失败了，保持原数据
	s.Value = nil
	return nil
}

// Get 返回 Stop 字段的值
func (s *StopField) Get() interface{} {
	return s.Value
}

// Set 设置 Stop 字段的值
func (s *StopField) Set(value interface{}) {
	s.Value = value
}

// IsString 判断 Stop 字段是否为字符串
func (s *StopField) IsString() bool {
	_, ok := s.Value.(string)
	return ok
}

// IsStringSlice 判断 Stop 字段是否为字符串数组
func (s *StopField) IsStringSlice() bool {
	_, ok := s.Value.([]string)
	return ok
}

// MarshalJSON 实现 Message 的自定义序列化
func (m *RequestMessage) MarshalJSON() ([]byte, error) {
	type Alias RequestMessage
	aux := &struct {
		Content interface{} `json:"content"`
		*Alias
	}{
		Content: m.Content,
		Alias:   (*Alias)(m),
	}
	return json.Marshal(aux)
}

// UnmarshalJSON 实现 Message 的自定义反序列化
func (m *RequestMessage) UnmarshalJSON(data []byte) error {
	type Alias RequestMessage
	aux := &struct {
		Content json.RawMessage `json:"content"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// 尝试解析为字符串
	var str string
	if err := json.Unmarshal(aux.Content, &str); err == nil {
		m.Content = str
		return nil
	}

	// 尝试解析为内容部分数组
	var contentParts []json.RawMessage
	if err := json.Unmarshal(aux.Content, &contentParts); err == nil {
		var parts []interface{}

		for _, part := range contentParts {
			var partMap map[string]interface{}
			if err := json.Unmarshal(part, &partMap); err != nil {
				return err
			}

			partType, ok := partMap["type"].(string)
			if !ok {
				return fmt.Errorf("内容部分缺少 type 字段")
			}

			switch partType {
			case "text":
				var textPart TextContentPart
				if err := json.Unmarshal(part, &textPart); err != nil {
					return err
				}
				parts = append(parts, textPart)
			case "image_url":
				var imagePart ImageContentPart
				if err := json.Unmarshal(part, &imagePart); err != nil {
					return err
				}
				parts = append(parts, imagePart)
			case "input_audio":
				var audioPart AudioContentPart
				if err := json.Unmarshal(part, &audioPart); err != nil {
					return err
				}
				parts = append(parts, audioPart)
			case "file":
				var filePart FileContentPart
				if err := json.Unmarshal(part, &filePart); err != nil {
					return err
				}
				parts = append(parts, filePart)
			default:
				// 对于未知类型，保留为原始 JSON
				parts = append(parts, partMap)
			}
		}

		m.Content = parts
		return nil
	}

	// 如果解析失败，将原始数据存储为字符串
	m.Content = string(aux.Content)
	return nil
}
