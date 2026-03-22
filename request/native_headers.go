package request

import "reflect"

// extractNativeHeaders 从原生请求 payload 中提取自定义 HTTP 头部。
//
// 提取规则：
//   - payload 必须是结构体或结构体指针
//   - 结构体中存在名为 Headers 的字段
//   - 字段类型必须为 map[string]string
//
// 返回值为拷贝后的 map，避免后续流程意外修改调用方传入的数据。
func extractNativeHeaders(payload any) map[string]string {
	if payload == nil {
		return nil
	}

	v := reflect.ValueOf(payload)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	headersField := v.FieldByName("Headers")
	if !headersField.IsValid() || !headersField.CanInterface() {
		return nil
	}

	headers, ok := headersField.Interface().(map[string]string)
	if !ok || len(headers) == 0 {
		return nil
	}

	cloned := make(map[string]string, len(headers))
	for key, value := range headers {
		cloned[key] = value
	}

	return cloned
}
