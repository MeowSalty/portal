package converter

import (
	"encoding/json"
	"strings"

	portalErrors "github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/logger"
)

// SerializeToolInput 将工具输入参数序列化为 JSON 字符串。
//
// 优先级：已有字符串 > Payload 映射序列化。
func SerializeToolInput(input interface{}, existingArgs *string) (string, error) {
	if existingArgs != nil {
		if strings.TrimSpace(*existingArgs) != "" {
			return *existingArgs, nil
		}
	}

	if input == nil {
		err := portalErrors.New(portalErrors.ErrCodeInvalidArgument, "工具输入为空")
		logger.Default().Warn("序列化工具输入失败", "error", err)
		return "", err
	}

	switch v := input.(type) {
	case string:
		if strings.TrimSpace(v) != "" {
			return v, nil
		}
		return "", nil
	case []byte:
		if len(v) > 0 {
			return string(v), nil
		}
		return "", nil
	case map[string]interface{}:
		return SerializePayload(v)
	default:
		data, err := json.Marshal(v)
		if err != nil {
			wrapped := portalErrors.Wrap(portalErrors.ErrCodeInternal, "序列化工具输入失败", err)
			logger.Default().Warn("序列化工具输入失败", "error", wrapped)
			return "", wrapped
		}
		return string(data), nil
	}
}

// SerializePayload 将任意 Payload 序列化为 JSON 字符串。
func SerializePayload(payload map[string]interface{}) (string, error) {
	if payload == nil {
		err := portalErrors.New(portalErrors.ErrCodeInvalidArgument, "Payload 为空")
		logger.Default().Warn("序列化 Payload 失败", "error", err)
		return "", err
	}

	data, err := json.Marshal(payload)
	if err != nil {
		wrapped := portalErrors.Wrap(portalErrors.ErrCodeInternal, "序列化 Payload 失败", err)
		logger.Default().Warn("序列化 Payload 失败", "error", wrapped)
		return "", wrapped
	}

	return string(data), nil
}

// DeserializeToolInput 将 JSON 字符串反序列化为工具输入参数。
func DeserializeToolInput(args string) (map[string]interface{}, error) {
	if strings.TrimSpace(args) == "" {
		return nil, nil
	}

	var input interface{}
	if err := json.Unmarshal([]byte(args), &input); err != nil {
		wrapped := portalErrors.Wrap(portalErrors.ErrCodeInternal, "反序列化工具输入失败", err)
		logger.Default().Warn("反序列化工具输入失败", "error", wrapped)
		return nil, wrapped
	}

	mapped, ok := input.(map[string]interface{})
	if !ok {
		err := portalErrors.New(portalErrors.ErrCodeInvalidArgument, "工具输入不是对象")
		logger.Default().Warn("反序列化工具输入失败", "error", err)
		return nil, err
	}

	return mapped, nil
}

// DeserializePayload 将 JSON 字符串反序列化为 Payload。
func DeserializePayload(jsonStr string) (map[string]interface{}, error) {
	if strings.TrimSpace(jsonStr) == "" {
		return nil, nil
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &payload); err != nil {
		wrapped := portalErrors.Wrap(portalErrors.ErrCodeInternal, "反序列化 Payload 失败", err)
		logger.Default().Warn("反序列化 Payload 失败", "error", wrapped)
		return nil, wrapped
	}

	return payload, nil
}
