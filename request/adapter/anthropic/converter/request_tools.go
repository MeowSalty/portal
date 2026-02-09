package converter

import (
	"github.com/MeowSalty/portal/logger"
	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// convertToolCallFromContract 从 Contract 转换工具调用。
func convertToolCallFromContract(tc *types.ToolCall) (*anthropicTypes.ContentBlockParam, error) {
	if tc == nil || tc.ID == nil || tc.Name == nil {
		return nil, nil
	}

	toolUseBlock := &anthropicTypes.ToolUseBlockParam{
		Type: anthropicTypes.ContentBlockTypeToolUse,
		ID:   *tc.ID,
		Name: *tc.Name,
	}

	// 解析 Arguments
	if tc.Arguments != nil {
		input, err := DeserializeToolInput(*tc.Arguments)
		if err == nil && input != nil {
			toolUseBlock.Input = input
		}
	}

	return &anthropicTypes.ContentBlockParam{
		ToolUse: toolUseBlock,
	}, nil
}

// convertToolsFromContract 从 Contract 转换工具列表。
func convertToolsFromContract(tools []types.Tool, vendorExtras map[string]interface{}) ([]anthropicTypes.ToolUnion, error) {
	result := make([]anthropicTypes.ToolUnion, 0, len(tools))

	for _, tool := range tools {
		if tool.Type == "function" && tool.Function != nil {
			// 转换为自定义工具
			customTool := &anthropicTypes.Tool{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
			}

			// 转换 Parameters 为 InputSchema
			if tool.Function.Parameters != nil {
				if schema, ok := tool.Function.Parameters.(anthropicTypes.InputSchema); ok {
					customTool.InputSchema = schema
				} else if schemaMap, ok := tool.Function.Parameters.(map[string]interface{}); ok {
					customTool.InputSchema = anthropicTypes.InputSchema{
						Type:       anthropicTypes.InputSchemaTypeObject,
						Properties: schemaMap,
					}
				}
			}

			// 从 VendorExtras 恢复 CacheControl
			if tool.VendorExtras != nil {
				var cc anthropicTypes.CacheControlEphemeral
				if found, err := GetVendorExtra("cache_control", tool.VendorExtras, &cc); err == nil && found {
					customTool.CacheControl = &cc
				} else if err != nil {
					logger.Default().Warn("读取工具缓存配置失败", "error", err)
				} else if fallback, ok := tool.VendorExtras["cache_control"].(*anthropicTypes.CacheControlEphemeral); ok {
					customTool.CacheControl = fallback
				}
			}

			result = append(result, anthropicTypes.ToolUnion{
				Custom: customTool,
			})
		}
	}

	// 从 VendorExtras 恢复特殊工具类型
	var toolsExtras []map[string]interface{}
	if found, err := GetVendorExtra("tools_extras", vendorExtras, &toolsExtras); err == nil && found {
		// keep toolsExtras
	} else if err != nil {
		logger.Default().Warn("读取特殊工具扩展失败", "error", err)
	} else if vendorExtras != nil {
		if fallback, ok := vendorExtras["tools_extras"].([]map[string]interface{}); ok {
			toolsExtras = fallback
		}
	}

	if len(toolsExtras) > 0 {
		for _, extra := range toolsExtras {
			toolType, _ := extra["type"].(string)
			switch toolType {
			case "bash_20250124":
				if tool, ok := extra["tool"].(*anthropicTypes.ToolBash20250124); ok {
					result = append(result, anthropicTypes.ToolUnion{
						Bash20250124: tool,
					})
				}
			case "text_editor_20250124":
				if tool, ok := extra["tool"].(*anthropicTypes.ToolTextEditor20250124); ok {
					result = append(result, anthropicTypes.ToolUnion{
						TextEditor20250124: tool,
					})
				}
			case "text_editor_20250429":
				if tool, ok := extra["tool"].(*anthropicTypes.ToolTextEditor20250429); ok {
					result = append(result, anthropicTypes.ToolUnion{
						TextEditor20250429: tool,
					})
				}
			case "text_editor_20250728":
				if tool, ok := extra["tool"].(*anthropicTypes.ToolTextEditor20250728); ok {
					result = append(result, anthropicTypes.ToolUnion{
						TextEditor20250728: tool,
					})
				}
			case "web_search_20250305":
				if tool, ok := extra["tool"].(*anthropicTypes.WebSearchTool20250305); ok {
					result = append(result, anthropicTypes.ToolUnion{
						WebSearch20250305: tool,
					})
				}
			}
		}
	}

	return result, nil
}

// convertToolChoiceFromContract 从 Contract 转换工具选择。
func convertToolChoiceFromContract(toolChoice *types.ToolChoice, parallelToolCalls *bool) (*anthropicTypes.ToolChoiceParam, error) {
	if toolChoice == nil || toolChoice.Mode == nil {
		return nil, nil
	}

	result := &anthropicTypes.ToolChoiceParam{}

	// 计算 DisableParallelToolUse
	var disableParallel *bool
	if parallelToolCalls != nil {
		disabled := !*parallelToolCalls
		disableParallel = &disabled
	}

	switch *toolChoice.Mode {
	case "auto":
		result.Auto = &anthropicTypes.ToolChoiceAuto{
			Type:                   anthropicTypes.ToolChoiceTypeAuto,
			DisableParallelToolUse: disableParallel,
		}
	case "any":
		result.Any = &anthropicTypes.ToolChoiceAny{
			Type:                   anthropicTypes.ToolChoiceTypeAny,
			DisableParallelToolUse: disableParallel,
		}
	case "tool":
		if toolChoice.Function != nil {
			result.Tool = &anthropicTypes.ToolChoiceTool{
				Type:                   anthropicTypes.ToolChoiceTypeTool,
				Name:                   *toolChoice.Function,
				DisableParallelToolUse: disableParallel,
			}
		}
	case "none":
		result.None = &anthropicTypes.ToolChoiceNone{
			Type: anthropicTypes.ToolChoiceTypeNone,
		}
	}

	return result, nil
}

// convertToolsToContract 转换工具列表。
func convertToolsToContract(tools []anthropicTypes.ToolUnion) ([]types.Tool, []map[string]interface{}, error) {
	result := make([]types.Tool, 0, len(tools))
	vendorExtras := make([]map[string]interface{}, 0)

	for _, tool := range tools {
		if tool.Custom != nil {
			// 标准自定义工具
			contractTool := types.Tool{
				Type: "function",
				Function: &types.Function{
					Name:        tool.Custom.Name,
					Description: tool.Custom.Description,
					Parameters:  tool.Custom.InputSchema,
				},
			}

			// CacheControl 放入 VendorExtras
			if tool.Custom.CacheControl != nil {
				contractTool.VendorExtras = make(map[string]interface{})
				if err := SaveVendorExtra("cache_control", tool.Custom.CacheControl, contractTool.VendorExtras); err != nil {
					logger.Default().Warn("保存工具缓存配置失败", "error", err)
				}
				SetVendorSource(&contractTool, string(types.VendorSourceAnthropic))
			}

			result = append(result, contractTool)
		} else {
			// 其他特殊工具类型（bash、text_editor、web_search）放入 VendorExtras
			extras := make(map[string]interface{})

			if tool.Bash20250124 != nil {
				extras["type"] = "bash_20250124"
				extras["tool"] = tool.Bash20250124
			} else if tool.TextEditor20250124 != nil {
				extras["type"] = "text_editor_20250124"
				extras["tool"] = tool.TextEditor20250124
			} else if tool.TextEditor20250429 != nil {
				extras["type"] = "text_editor_20250429"
				extras["tool"] = tool.TextEditor20250429
			} else if tool.TextEditor20250728 != nil {
				extras["type"] = "text_editor_20250728"
				extras["tool"] = tool.TextEditor20250728
			} else if tool.WebSearch20250305 != nil {
				extras["type"] = "web_search_20250305"
				extras["tool"] = tool.WebSearch20250305
			}

			if len(extras) > 0 {
				vendorExtras = append(vendorExtras, extras)
			}
		}
	}

	return result, vendorExtras, nil
}

// convertToolChoiceToContract 转换工具选择。
func convertToolChoiceToContract(toolChoice *anthropicTypes.ToolChoiceParam) (*types.ToolChoice, *bool, error) {
	if toolChoice == nil {
		return nil, nil, nil
	}

	result := &types.ToolChoice{}
	var parallelToolCalls *bool

	if toolChoice.Auto != nil {
		mode := "auto"
		result.Mode = &mode

		if toolChoice.Auto.DisableParallelToolUse != nil {
			enabled := !*toolChoice.Auto.DisableParallelToolUse
			parallelToolCalls = &enabled
		}
	} else if toolChoice.Any != nil {
		mode := "any"
		result.Mode = &mode

		if toolChoice.Any.DisableParallelToolUse != nil {
			enabled := !*toolChoice.Any.DisableParallelToolUse
			parallelToolCalls = &enabled
		}
	} else if toolChoice.Tool != nil {
		mode := "tool"
		result.Mode = &mode
		result.Function = &toolChoice.Tool.Name

		if toolChoice.Tool.DisableParallelToolUse != nil {
			enabled := !*toolChoice.Tool.DisableParallelToolUse
			parallelToolCalls = &enabled
		}
	} else if toolChoice.None != nil {
		mode := "none"
		result.Mode = &mode
	}

	return result, parallelToolCalls, nil
}
