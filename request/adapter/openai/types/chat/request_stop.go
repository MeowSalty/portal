package chat

import (
	"encoding/json"
)

// StopConfiguration 表示停止条件
// 支持字符串或字符串数组。
type StopConfiguration struct {
	StringValue *string
	StringArray []string
}

// MarshalJSON 实现 StopConfiguration 的自定义 JSON 序列化
func (s StopConfiguration) MarshalJSON() ([]byte, error) {
	if s.StringValue != nil {
		return json.Marshal(s.StringValue)
	}
	if s.StringArray != nil {
		return json.Marshal(s.StringArray)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON 实现 StopConfiguration 的自定义 JSON 反序列化
func (s *StopConfiguration) UnmarshalJSON(data []byte) error {
	// 首先尝试反序列化为字符串
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		s.StringValue = &str
		return nil
	}

	// 尝试反序列化为字符串数组
	var arr []string
	if err := json.Unmarshal(data, &arr); err == nil {
		s.StringArray = arr
		return nil
	}

	return nil
}
