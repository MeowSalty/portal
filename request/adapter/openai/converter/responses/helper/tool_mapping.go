package helper

import (
	"encoding/json"

	"github.com/MeowSalty/portal/request/adapter/types"
)

// convertToolOutput 转换工具输出为 ToolResult。
// 根据 Output content conversion rules：
// - 如果 output 是 JSON 字符串，放入 Content
// - 否则放入 Payload["output"]
func convertToolOutput(callID *string, name *string, status string, output json.RawMessage) *types.ToolResult {
	result := &types.ToolResult{
		ID:   callID,
		Name: name,
	}

	// 尝试解析为字符串
	var strValue string
	if err := json.Unmarshal(output, &strValue); err == nil {
		// 是 JSON 字符串，放入 Content
		result.Content = &strValue
	} else {
		// 不是字符串，放入 Payload
		result.Payload = map[string]interface{}{
			"status": status,
			"output": output,
		}
		// Content 保持为 nil
	}

	return result
}
