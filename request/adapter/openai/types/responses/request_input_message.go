package responses

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	portalErrors "github.com/MeowSalty/portal/errors"
)

var normalizedMessageIDCounter uint64

// InputMessage 表示简化输入消息
// role 支持 user/system/developer/assistant。
// content 支持 string 或 []InputContent 的联合类型。
type InputMessage struct {
	Type    InputItemType       `json:"type,omitempty"`
	ID      *string             `json:"id,omitempty"`
	Role    ResponseMessageRole `json:"role"`
	Content InputMessageContent `json:"content"`
}

// InputMessageContent 表示输入消息内容的联合类型
// 支持 string（文本输入）或 []InputContent（内容列表）。
type InputMessageContent struct {
	// String 表示文本输入
	String *string `json:"-"`
	// List 表示内容列表
	List *[]InputContent `json:"-"`
}

// MarshalJSON 实现 InputMessageContent 的自定义 JSON 序列化
// 优先序列化 String，其次序列化 List
func (c InputMessageContent) MarshalJSON() ([]byte, error) {
	// 优先使用 String
	if c.String != nil {
		return json.Marshal(c.String)
	}
	// 其次使用 List
	if c.List != nil {
		return json.Marshal(c.List)
	}
	// 都为空则序列化为 null
	return json.Marshal(nil)
}

// UnmarshalJSON 实现 InputMessageContent 的自定义 JSON 反序列化
// 支持 string 或 []InputContent
func (c *InputMessageContent) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	// 尝试解析为 string
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*c = InputMessageContent{String: &str}
		return nil
	}

	// 尝试解析为 []InputContent
	var list []InputContent
	if err := json.Unmarshal(data, &list); err == nil {
		*c = InputMessageContent{List: &list}
		return nil
	}

	return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "input message content 必须是 string 或 []InputContent")
}

// NewInputMessageContentFromString 创建文本输入内容
func NewInputMessageContentFromString(text string) InputMessageContent {
	return InputMessageContent{String: &text}
}

// NewInputMessageContentFromList 创建内容列表输入
func NewInputMessageContentFromList(list []InputContent) InputMessageContent {
	return InputMessageContent{List: &list}
}

// ItemReferenceParam 表示引用项
// 用于 item_reference。
type ItemReferenceParam struct {
	Type *InputItemType `json:"type,omitempty"`
	ID   string         `json:"id"`
}

// shouldNormalizeToOutputMessage 判断是否为错误格式的输出消息输入。
//
// 条件：content 为列表，且每个元素的 type 都是 output_text/refusal。
// 参考文档：https://developers.openai.com/api/reference/resources/responses/methods/create
func shouldNormalizeToOutputMessage(raw map[string]interface{}) bool {
	contentValue, ok := raw["content"]
	if !ok {
		return false
	}

	contentItems, ok := contentValue.([]interface{})
	if !ok || len(contentItems) == 0 {
		return false
	}

	for _, item := range contentItems {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			return false
		}
		itemType, ok := itemMap["type"].(string)
		if !ok {
			return false
		}
		if itemType != string(OutputMessageContentTypeOutputText) && itemType != string(OutputMessageContentTypeRefusal) {
			return false
		}
	}

	return true
}

// normalizeOutputMessageRaw 将错误格式的输入 message 补齐为 OutputMessage 结构。
//
// 适用场景：部分第三方库把模型输出回填到 input 中，但缺失 status/type 字段。
// 此处作为临时兼容逻辑补齐：status=completed，type=message。
//
// 参考文档：
// https://developers.openai.com/api/reference/resources/responses/methods/create
func normalizeOutputMessageRaw(raw map[string]interface{}) map[string]interface{} {
	normalized := make(map[string]interface{}, len(raw)+2)
	for key, value := range raw {
		normalized[key] = value
	}

	if idValue, ok := normalized["id"]; !ok || idValue == nil || idValue == "" {
		normalized["id"] = generateOutputMessageID()
	}
	if _, ok := normalized["type"]; !ok {
		normalized["type"] = OutputItemTypeMessage
	}
	if _, ok := normalized["status"]; !ok {
		normalized["status"] = "completed"
	}

	return normalized
}

func generateOutputMessageID() string {
	counter := atomic.AddUint64(&normalizedMessageIDCounter, 1)
	return fmt.Sprintf("msg_auto_%d_%d", time.Now().UnixNano(), counter)
}
