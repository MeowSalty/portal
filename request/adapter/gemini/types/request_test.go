package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRequestExtraFields_UnknownFieldCapture 测试未知字段捕获
func TestRequestExtraFields_UnknownFieldCapture(t *testing.T) {
	jsonStr := `{
		"contents": [{
			"role": "user",
			"parts": [{"text": "hello"}]
		}],
		"unknownField": "value",
		"anotherUnknown": 123
	}`

	var req Request
	err := json.Unmarshal([]byte(jsonStr), &req)
	require.NoError(t, err)

	// 验证显式字段正确解析
	assert.Len(t, req.Contents, 1)
	assert.Equal(t, "user", req.Contents[0].Role)
	assert.Len(t, req.Contents[0].Parts, 1)
	assert.NotNil(t, req.Contents[0].Parts[0].Text)
	assert.Equal(t, "hello", *req.Contents[0].Parts[0].Text)

	// 验证未知字段被捕获
	assert.Len(t, req.ExtraFields, 2)
	assert.Contains(t, req.ExtraFields, "unknownField")
	assert.Contains(t, req.ExtraFields, "anotherUnknown")
}

// TestRequestExtraFields_ExplicitFieldPriority 测试显式字段优先
func TestRequestExtraFields_ExplicitFieldPriority(t *testing.T) {
	jsonStr := `{
		"contents": [{
			"role": "user",
			"parts": [{"text": "hello"}]
		}],
		"systemInstruction": {
			"role": "user",
			"parts": [{"text": "system"}]
		}
	}`

	var req Request
	err := json.Unmarshal([]byte(jsonStr), &req)
	require.NoError(t, err)

	// 验证显式字段正确解析
	assert.NotNil(t, req.SystemInstruction)
	assert.Equal(t, "user", req.SystemInstruction.Role)

	// 验证 ExtraFields 为空（因为所有字段都是显式字段）
	assert.Nil(t, req.ExtraFields)
}

// TestRequestExtraFields_RoundTrip 测试 round-trip 序列化
func TestRequestExtraFields_RoundTrip(t *testing.T) {
	originalJSON := `{
		"contents": [{
			"role": "user",
			"parts": [{"text": "hello"}]
		}],
		"unknownField": "value",
		"anotherUnknown": 123
	}`

	var req Request
	err := json.Unmarshal([]byte(originalJSON), &req)
	require.NoError(t, err)

	// 序列化回 JSON
	output, err := json.Marshal(req)
	require.NoError(t, err)

	// 解析回结构体
	var req2 Request
	err = json.Unmarshal(output, &req2)
	require.NoError(t, err)

	// 验证显式字段一致
	assert.Len(t, req2.Contents, 1)
	assert.Equal(t, "user", req2.Contents[0].Role)

	// 验证未知字段被保留
	assert.Len(t, req2.ExtraFields, 2)
	assert.Contains(t, req2.ExtraFields, "unknownField")
	assert.Contains(t, req2.ExtraFields, "anotherUnknown")
}

// TestRequestExtraFields_EmptyExtraFields 测试空 ExtraFields
func TestRequestExtraFields_EmptyExtraFields(t *testing.T) {
	jsonStr := `{
		"contents": [{
			"role": "user",
			"parts": [{"text": "hello"}]
		}]
	}`

	var req Request
	err := json.Unmarshal([]byte(jsonStr), &req)
	require.NoError(t, err)

	// 验证 ExtraFields 为 nil
	assert.Nil(t, req.ExtraFields)

	// 序列化应该保持简洁
	output, err := json.Marshal(req)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(output, &result)
	require.NoError(t, err)

	// 验证输出中只有显式字段
	assert.Contains(t, result, "contents")
	assert.NotContains(t, result, "ExtraFields")
}

// TestRequestExtraFields_MarshalWithExtraFields 测试序列化时合并 ExtraFields
func TestRequestExtraFields_MarshalWithExtraFields(t *testing.T) {
	req := Request{
		Contents: []Content{
			{
				Role: "user",
				Parts: []Part{
					{Text: stringPtr("hello")},
				},
			},
		},
		ExtraFields: map[string]json.RawMessage{
			"customField": json.RawMessage(`"customValue"`),
		},
	}

	output, err := json.Marshal(req)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(output, &result)
	require.NoError(t, err)

	// 验证显式字段存在
	assert.Contains(t, result, "contents")

	// 验证 ExtraFields 被合并
	assert.Contains(t, result, "customField")
	assert.Equal(t, "customValue", result["customField"])
}

// TestRequestExtraFields_PartMarshalJSONCompatibility 测试与 Part.MarshalJSON() 兼容性
func TestRequestExtraFields_PartMarshalJSONCompatibility(t *testing.T) {
	jsonStr := `{
		"contents": [{
			"role": "user",
			"parts": [
				{"text": "hello"},
				{"inlineData": {"mimeType": "image/png", "data": "base64data"}},
				{"functionCall": {"name": "test", "args": {}}}
			]
		}]
	}`

	var req Request
	err := json.Unmarshal([]byte(jsonStr), &req)
	require.NoError(t, err)

	// 验证 Parts 正确解析
	assert.Len(t, req.Contents[0].Parts, 3)
	assert.NotNil(t, req.Contents[0].Parts[0].Text)
	assert.NotNil(t, req.Contents[0].Parts[1].InlineData)
	assert.NotNil(t, req.Contents[0].Parts[2].FunctionCall)

	// 序列化
	output, err := json.Marshal(req)
	require.NoError(t, err)

	// 验证序列化结果包含正确的 Part 字段
	var result map[string]interface{}
	err = json.Unmarshal(output, &result)
	require.NoError(t, err)

	contents := result["contents"].([]interface{})
	content := contents[0].(map[string]interface{})
	parts := content["parts"].([]interface{})

	// 验证第一个 Part 只有 text 字段
	part0 := parts[0].(map[string]interface{})
	assert.Contains(t, part0, "text")
	assert.NotContains(t, part0, "inlineData")

	// 验证第二个 Part 只有 inlineData 字段
	part1 := parts[1].(map[string]interface{})
	assert.Contains(t, part1, "inlineData")
	assert.NotContains(t, part1, "text")
}

// 辅助函数：创建字符串指针
func stringPtr(s string) *string {
	return &s
}
