package converter

import (
	"encoding/json"

	portalErrors "github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/logger"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// SaveVendorExtra 将 Anthropic 特有字段保存到 VendorExtras 中。
//
// - key: 字段名，使用 anthropic. 前缀（顶层字段）或直接使用（工具级别）
// - value: 要保存的值，使用 json.RawMessage 保存原始 JSON
// - extras: 目标 extras map
func SaveVendorExtra(key string, value interface{}, extras map[string]interface{}) error {
	if extras == nil {
		return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "VendorExtras 不能为空")
	}
	if key == "" {
		return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "字段名不能为空")
	}
	if value == nil {
		return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "字段值不能为空")
	}

	switch raw := value.(type) {
	case json.RawMessage:
		return SaveVendorExtraRaw(key, raw, extras)
	case []byte:
		return SaveVendorExtraRaw(key, json.RawMessage(raw), extras)
	case string:
		return SaveVendorExtraRaw(key, json.RawMessage([]byte(raw)), extras)
	default:
		encoded, err := json.Marshal(value)
		if err != nil {
			return portalErrors.Wrap(portalErrors.ErrCodeInternal, "序列化供应商扩展失败", err)
		}
		return SaveVendorExtraRaw(key, json.RawMessage(encoded), extras)
	}
}

// SaveVendorExtraRaw 直接保存原始 JSON 字节到 VendorExtras。
func SaveVendorExtraRaw(key string, raw json.RawMessage, extras map[string]interface{}) error {
	if extras == nil {
		return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "VendorExtras 不能为空")
	}
	if key == "" {
		return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "字段名不能为空")
	}
	if len(raw) == 0 {
		return nil
	}

	data := make([]byte, len(raw))
	copy(data, raw)
	extras[key] = json.RawMessage(data)
	return nil
}

// SetVendorSource 设置供应商来源标识。
func SetVendorSource(contract interface{}, source string) {
	if contract == nil {
		return
	}

	vendorSource := types.VendorSource(source)
	var vendorSourcePtr *types.VendorSource
	if source != "" {
		vendorSourcePtr = &vendorSource
	}

	switch target := contract.(type) {
	case *types.RequestContract:
		target.Source = vendorSource
		target.VendorExtrasSource = vendorSourcePtr
	case *types.System:
		target.VendorExtrasSource = vendorSourcePtr
	case *types.Message:
		target.VendorExtrasSource = vendorSourcePtr
	case *types.ContentPart:
		target.VendorExtrasSource = vendorSourcePtr
	case *types.Tool:
		target.VendorExtrasSource = vendorSourcePtr
	case *types.Reasoning:
		target.VendorExtrasSource = vendorSourcePtr
	case *types.StreamOption:
		target.VendorExtrasSource = vendorSourcePtr
	default:
		return
	}
}

// GetVendorExtra 从 VendorExtras 中获取 Anthropic 特有字段。
//
// - key: 字段名，使用 anthropic. 前缀
// - extras: 源 extras map
// - target: 目标指针，用于反序列化
//
// 返回值：是否找到该字段。
func GetVendorExtra(key string, extras map[string]interface{}, target interface{}) (found bool, err error) {
	if extras == nil {
		return false, nil
	}
	if key == "" {
		return false, portalErrors.New(portalErrors.ErrCodeInvalidArgument, "字段名不能为空")
	}
	if target == nil {
		return false, portalErrors.New(portalErrors.ErrCodeInvalidArgument, "目标不能为空")
	}

	value, ok := extras[key]
	if !ok {
		return false, nil
	}

	var raw []byte
	switch v := value.(type) {
	case json.RawMessage:
		raw = v
	case []byte:
		raw = v
	case string:
		raw = []byte(v)
	default:
		encoded, marshalErr := json.Marshal(v)
		if marshalErr != nil {
			return true, portalErrors.Wrap(portalErrors.ErrCodeInternal, "序列化供应商扩展失败", marshalErr)
		}
		raw = encoded
	}

	if unmarshalErr := json.Unmarshal(raw, target); unmarshalErr != nil {
		return true, portalErrors.Wrap(portalErrors.ErrCodeInternal, "反序列化供应商扩展失败", unmarshalErr)
	}

	return true, nil
}

func getStringExtra(key string, extras map[string]interface{}) (string, bool) {
	var value string
	if found, err := GetVendorExtra(key, extras, &value); err == nil && found {
		return value, true
	}
	if extras == nil {
		return "", false
	}
	fallback, ok := extras[key].(string)
	return fallback, ok
}

func getBoolExtra(key string, extras map[string]interface{}) (bool, bool) {
	var value bool
	if found, err := GetVendorExtra(key, extras, &value); err == nil && found {
		return value, true
	}
	if extras == nil {
		return false, false
	}
	fallback, ok := extras[key].(bool)
	return fallback, ok
}

func getIntExtra(key string, extras map[string]interface{}) (int, bool) {
	var value int
	if found, err := GetVendorExtra(key, extras, &value); err == nil && found {
		return value, true
	}
	if extras == nil {
		return 0, false
	}
	if fallback, ok := extras[key].(int); ok {
		return fallback, true
	}
	if fallback, ok := extras[key].(float64); ok {
		return int(fallback), true
	}
	return 0, false
}

// GetVendorExtraRaw 从 VendorExtras 中获取原始 JSON 字节。
func GetVendorExtraRaw(key string, extras map[string]interface{}) (json.RawMessage, bool) {
	if extras == nil {
		return nil, false
	}

	value, ok := extras[key]
	if !ok {
		return nil, false
	}

	switch v := value.(type) {
	case json.RawMessage:
		data := make([]byte, len(v))
		copy(data, v)
		return json.RawMessage(data), true
	case []byte:
		data := make([]byte, len(v))
		copy(data, v)
		return json.RawMessage(data), true
	case string:
		return json.RawMessage([]byte(v)), true
	default:
		encoded, err := json.Marshal(v)
		if err != nil {
			return nil, false
		}
		return json.RawMessage(encoded), true
	}
}

// GetVendorSource 获取供应商来源标识。
func GetVendorSource(contract interface{}) string {
	if contract == nil {
		return ""
	}

	switch target := contract.(type) {
	case *types.RequestContract:
		if target.Source != "" {
			return string(target.Source)
		}
		if target.VendorExtrasSource != nil {
			return string(*target.VendorExtrasSource)
		}
	case types.RequestContract:
		if target.Source != "" {
			return string(target.Source)
		}
		if target.VendorExtrasSource != nil {
			return string(*target.VendorExtrasSource)
		}
	case *types.System:
		if target.VendorExtrasSource != nil {
			return string(*target.VendorExtrasSource)
		}
	case *types.Message:
		if target.VendorExtrasSource != nil {
			return string(*target.VendorExtrasSource)
		}
	case *types.ContentPart:
		if target.VendorExtrasSource != nil {
			return string(*target.VendorExtrasSource)
		}
	case *types.Tool:
		if target.VendorExtrasSource != nil {
			return string(*target.VendorExtrasSource)
		}
	case *types.Reasoning:
		if target.VendorExtrasSource != nil {
			return string(*target.VendorExtrasSource)
		}
	case *types.StreamOption:
		if target.VendorExtrasSource != nil {
			return string(*target.VendorExtrasSource)
		}
	default:
		return ""
	}

	return ""
}

// IsAnthropicVendor 判断是否为 Anthropic 供应商。
func IsAnthropicVendor(contract interface{}) bool {
	return GetVendorSource(contract) == string(types.VendorSourceAnthropic)
}

// SafeMarshal 安全序列化，失败时记录 WARN 日志并返回 nil。
func SafeMarshal(v interface{}) json.RawMessage {
	if v == nil {
		logger.Default().Warn("序列化失败", "error", portalErrors.New(portalErrors.ErrCodeInvalidArgument, "序列化对象为空"))
		return nil
	}

	data, err := json.Marshal(v)
	if err != nil {
		logger.Default().Warn("序列化失败", "error", err)
		return nil
	}

	return json.RawMessage(data)
}

// SafeUnmarshal 安全反序列化，失败时记录 WARN 日志并返回 false。
func SafeUnmarshal(data []byte, v interface{}) bool {
	if len(data) == 0 {
		logger.Default().Warn("反序列化失败", "error", portalErrors.New(portalErrors.ErrCodeInvalidArgument, "反序列化数据为空"))
		return false
	}
	if v == nil {
		logger.Default().Warn("反序列化失败", "error", portalErrors.New(portalErrors.ErrCodeInvalidArgument, "反序列化目标为空"))
		return false
	}

	if err := json.Unmarshal(data, v); err != nil {
		logger.Default().Warn("反序列化失败", "error", err)
		return false
	}

	return true
}
