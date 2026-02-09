package helper

import (
	responsesTypes "github.com/MeowSalty/portal/request/adapter/openai/types/responses"
	"github.com/MeowSalty/portal/request/adapter/openai/types/shared"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// ConvertResponseFormatToTextFormat 从 Contract 转换响应格式为文本格式。
func ConvertResponseFormatToTextFormat(format *types.ResponseFormat) (*responsesTypes.TextFormatUnion, *shared.VerbosityLevel, error) {
	if format == nil {
		return nil, nil, nil
	}

	result := &responsesTypes.TextFormatUnion{}

	switch format.Type {
	case string(responsesTypes.TextResponseFormatTypeText):
		result.Text = &responsesTypes.TextFormat{
			Type: responsesTypes.TextResponseFormatTypeText,
		}

	case string(responsesTypes.TextResponseFormatTypeJSONSchema):
		if format.JSONSchema != nil {
			jsonSchemaSpec := responsesTypes.TextFormatJSONSchema{
				Type: responsesTypes.TextResponseFormatTypeJSONSchema,
			}

			// 从 JSONSchema 中提取字段
			if schemaMap, ok := format.JSONSchema.(map[string]interface{}); ok {
				if name, ok := schemaMap["name"].(string); ok {
					jsonSchemaSpec.Name = name
				}
				if desc, ok := schemaMap["description"].(string); ok {
					jsonSchemaSpec.Description = &desc
				}
				if schema, ok := schemaMap["schema"].(map[string]interface{}); ok {
					jsonSchemaSpec.Schema = schema
				}
				if strict, ok := schemaMap["strict"].(bool); ok {
					jsonSchemaSpec.Strict = &strict
				}
			}

			result.JSONSchema = &jsonSchemaSpec
		}

	case string(responsesTypes.TextResponseFormatTypeJSONObject):
		result.JSONObject = &responsesTypes.TextFormatJSONObject{
			Type: responsesTypes.TextResponseFormatTypeJSONObject,
		}
	}

	return result, nil, nil
}

// convertTextFormatToResponseFormat 转换文本格式为响应格式。
func ConvertTextFormatToResponseFormat(format *responsesTypes.TextFormatUnion) (*types.ResponseFormat, error) {
	if format == nil {
		return nil, nil
	}

	result := &types.ResponseFormat{}

	if format.Text != nil {
		result.Type = string(format.Text.Type)
	} else if format.JSONSchema != nil {
		result.Type = string(format.JSONSchema.Type)
		// 构造 JSONSchema
		schema := map[string]interface{}{
			"name":   format.JSONSchema.Name,
			"schema": format.JSONSchema.Schema,
		}
		if format.JSONSchema.Description != nil {
			schema["description"] = *format.JSONSchema.Description
		}
		if format.JSONSchema.Strict != nil {
			schema["strict"] = *format.JSONSchema.Strict
		}
		result.JSONSchema = schema
	} else if format.JSONObject != nil {
		result.Type = string(format.JSONObject.Type)
	}

	return result, nil
}
