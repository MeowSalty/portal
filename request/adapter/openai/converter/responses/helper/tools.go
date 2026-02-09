package helper

import (
	"github.com/MeowSalty/portal/request/adapter/openai/types/shared"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// ConvertToolsToContract 将 OpenAI Responses 工具列表转换为统一工具列表。
func ConvertToolsToContract(tools []shared.ToolUnion) ([]types.Tool, error) {
	result := make([]types.Tool, 0, len(tools))

	for _, tool := range tools {
		if tool.Function != nil {
			// 标准函数工具
			contractTool := types.Tool{
				Type: "function",
			}

			// 处理两种形态：嵌套 Function 或顶层字段
			if tool.Function.Function.Name != "" {
				// Chat Completions 形态（嵌套）
				contractTool.Function = &types.Function{
					Name:        tool.Function.Function.Name,
					Description: tool.Function.Function.Description,
					Parameters:  tool.Function.Function.Parameters,
				}

				// Strict 放入 VendorExtras
				if tool.Function.Function.Strict != nil {
					contractTool.VendorExtras = make(map[string]interface{})
					source := types.VendorSourceOpenAIResponse
					contractTool.VendorExtrasSource = &source
					contractTool.VendorExtras["strict"] = *tool.Function.Function.Strict
				}
			} else if tool.Function.Name != nil {
				// Responses 形态（顶层）
				contractTool.Function = &types.Function{
					Name:        *tool.Function.Name,
					Description: tool.Function.Description,
					Parameters:  tool.Function.Parameters,
				}

				if tool.Function.Strict != nil {
					contractTool.VendorExtras = make(map[string]interface{})
					source := types.VendorSourceOpenAIResponse
					contractTool.VendorExtrasSource = &source
					contractTool.VendorExtras["strict"] = *tool.Function.Strict
				}
			}

			result = append(result, contractTool)
		} else {
			// 其他工具类型放入 VendorExtras
			contractTool := types.Tool{}
			contractTool.VendorExtras = make(map[string]interface{})
			source := types.VendorSourceOpenAIResponse
			contractTool.VendorExtrasSource = &source

			if tool.FileSearch != nil {
				contractTool.Type = "file_search"
				contractTool.VendorExtras["tool"] = tool.FileSearch
			} else if tool.ComputerUse != nil {
				contractTool.Type = "computer_use_preview"
				contractTool.VendorExtras["tool"] = tool.ComputerUse
			} else if tool.WebSearch != nil {
				contractTool.Type = "web_search"
				contractTool.VendorExtras["tool"] = tool.WebSearch
			} else if tool.MCP != nil {
				contractTool.Type = "mcp"
				contractTool.VendorExtras["tool"] = tool.MCP
			} else if tool.CodeInterpreter != nil {
				contractTool.Type = "code_interpreter"
				contractTool.VendorExtras["tool"] = tool.CodeInterpreter
			} else if tool.ImageGen != nil {
				contractTool.Type = "image_generation"
				contractTool.VendorExtras["tool"] = tool.ImageGen
			} else if tool.LocalShell != nil {
				contractTool.Type = "local_shell"
				contractTool.VendorExtras["tool"] = tool.LocalShell
			} else if tool.FunctionShell != nil {
				contractTool.Type = "shell"
				contractTool.VendorExtras["tool"] = tool.FunctionShell
			} else if tool.Custom != nil {
				contractTool.Type = "custom"
				contractTool.VendorExtras["tool"] = tool.Custom
			} else if tool.WebSearchPreview != nil {
				contractTool.Type = "web_search_preview"
				contractTool.VendorExtras["tool"] = tool.WebSearchPreview
			} else if tool.ApplyPatch != nil {
				contractTool.Type = "apply_patch"
				contractTool.VendorExtras["tool"] = tool.ApplyPatch
			}

			result = append(result, contractTool)
		}
	}

	return result, nil
}

// ConvertToolsFromContract 将统一工具列表转换为 OpenAI Responses 工具列表。
func ConvertToolsFromContract(tools []types.Tool) ([]shared.ToolUnion, error) {
	result := make([]shared.ToolUnion, 0, len(tools))

	for _, tool := range tools {
		toolUnion := shared.ToolUnion{}

		if tool.Type == "function" && tool.Function != nil {
			// 标准函数工具 - Responses 形态（顶层字段）
			toolFunc := &shared.ToolFunction{
				Type:        "function",
				Name:        &tool.Function.Name,
				Description: tool.Function.Description,
				Parameters:  tool.Function.Parameters,
			}

			// 从 VendorExtras 恢复 Strict
			if tool.VendorExtras != nil {
				if strict, ok := tool.VendorExtras["strict"].(*bool); ok {
					toolFunc.Strict = strict
				}
			}

			toolUnion.Function = toolFunc
		} else {
			// 从 VendorExtras 恢复其他工具类型
			if tool.VendorExtras != nil {
				switch tool.Type {
				case "file_search":
					if t, ok := tool.VendorExtras["tool"].(*shared.ToolFileSearch); ok {
						toolUnion.FileSearch = t
					}
				case "computer_use_preview":
					if t, ok := tool.VendorExtras["tool"].(*shared.ToolComputerUsePreview); ok {
						toolUnion.ComputerUse = t
					}
				case "web_search":
					if t, ok := tool.VendorExtras["tool"].(*shared.ToolWebSearch); ok {
						toolUnion.WebSearch = t
					}
				case "mcp":
					if t, ok := tool.VendorExtras["tool"].(*shared.ToolMCP); ok {
						toolUnion.MCP = t
					}
				case "code_interpreter":
					if t, ok := tool.VendorExtras["tool"].(*shared.ToolCodeInterpreter); ok {
						toolUnion.CodeInterpreter = t
					}
				case "image_generation":
					if t, ok := tool.VendorExtras["tool"].(*shared.ToolImageGen); ok {
						toolUnion.ImageGen = t
					}
				case "local_shell":
					if t, ok := tool.VendorExtras["tool"].(*shared.ToolLocalShell); ok {
						toolUnion.LocalShell = t
					}
				case "shell":
					if t, ok := tool.VendorExtras["tool"].(*shared.ToolFunctionShell); ok {
						toolUnion.FunctionShell = t
					}
				case "custom":
					if t, ok := tool.VendorExtras["tool"].(*shared.ToolCustom); ok {
						toolUnion.Custom = t
					}
				case "web_search_preview":
					if t, ok := tool.VendorExtras["tool"].(*shared.ToolWebSearchPreview); ok {
						toolUnion.WebSearchPreview = t
					}
				case "apply_patch":
					if t, ok := tool.VendorExtras["tool"].(*shared.ToolApplyPatch); ok {
						toolUnion.ApplyPatch = t
					}
				}
			}
		}

		result = append(result, toolUnion)
	}

	return result, nil
}

// ConvertToolChoiceToContract 将 OpenAI Responses 工具选择转换为统一工具选择。
func ConvertToolChoiceToContract(toolChoice *shared.ToolChoiceUnion) (*types.ToolChoice, error) {
	if toolChoice == nil {
		return nil, nil
	}

	result := &types.ToolChoice{}

	if toolChoice.Auto != nil {
		// 字符串模式（auto/none/required）
		mode := *toolChoice.Auto
		result.Mode = &mode
	} else if toolChoice.Named != nil {
		// 命名函数选择
		mode := "function"
		result.Mode = &mode
		result.Function = &toolChoice.Named.Function.Name
	} else if toolChoice.Allowed != nil {
		// 允许的工具列表
		mode := toolChoice.Allowed.Mode
		result.Mode = &mode

		// 提取允许的工具名称
		if len(toolChoice.Allowed.Tools) > 0 {
			allowed := make([]string, 0, len(toolChoice.Allowed.Tools))
			for _, t := range toolChoice.Allowed.Tools {
				if name, ok := t["name"].(string); ok {
					allowed = append(allowed, name)
				}
			}
			if len(allowed) > 0 {
				result.Allowed = allowed
			}
		}
	} else {
		// 其他类型（NamedCustom/NamedMCP/Hosted 等）
		// 提取 type 作为 mode
		if toolChoice.NamedCustom != nil {
			mode := toolChoice.NamedCustom.Type
			result.Mode = &mode
		} else if toolChoice.NamedMCP != nil {
			mode := toolChoice.NamedMCP.Type
			result.Mode = &mode
		} else if toolChoice.Hosted != nil {
			mode := toolChoice.Hosted.Type
			result.Mode = &mode
		} else if toolChoice.ApplyPatch != nil {
			mode := toolChoice.ApplyPatch.Type
			result.Mode = &mode
		} else if toolChoice.Shell != nil {
			mode := toolChoice.Shell.Type
			result.Mode = &mode
		}
	}

	return result, nil
}

// ConvertToolChoiceFromContract 将统一工具选择转换为 OpenAI Responses 工具选择。
func ConvertToolChoiceFromContract(toolChoice *types.ToolChoice) (*shared.ToolChoiceUnion, error) {
	if toolChoice == nil {
		return nil, nil
	}

	result := &shared.ToolChoiceUnion{}

	if toolChoice.Mode != nil {
		mode := *toolChoice.Mode

		// 简单模式字符串（auto/none/required）
		if mode == "auto" || mode == "none" || mode == "required" {
			result.Auto = &mode
		} else if mode == "function" && toolChoice.Function != nil {
			// 命名函数选择
			result.Named = &shared.ToolChoiceNamed{
				Type: "function",
			}
			result.Named.Function.Name = *toolChoice.Function
		} else if len(toolChoice.Allowed) > 0 {
			// 允许的工具列表
			result.Allowed = &shared.ToolChoiceAllowed{
				Type:  "allowed_tools",
				Mode:  mode,
				Tools: make([]map[string]interface{}, len(toolChoice.Allowed)),
			}
			for i, name := range toolChoice.Allowed {
				result.Allowed.Tools[i] = map[string]interface{}{
					"type": "function",
					"name": name,
				}
			}
		} else {
			// 其他类型（hosted/custom/mcp 等）
			switch mode {
			case "file_search", "web_search_preview", "computer_use_preview",
				"web_search_preview_2025_03_11", "image_generation", "code_interpreter":
				result.Hosted = &shared.ToolChoiceHosted{
					Type: mode,
				}
			case "custom":
				result.NamedCustom = &shared.ToolChoiceNamedCustom{
					Type: mode,
				}
				if toolChoice.Function != nil {
					result.NamedCustom.Custom.Name = *toolChoice.Function
				}
			case "mcp":
				result.NamedMCP = &shared.ToolChoiceNamedMCP{
					Type: mode,
				}
			case "apply_patch":
				result.ApplyPatch = &shared.ToolChoiceApplyPatch{
					Type: mode,
				}
			case "shell":
				result.Shell = &shared.ToolChoiceShell{
					Type: mode,
				}
			default:
				// 默认作为 auto 模式
				result.Auto = &mode
			}
		}
	}

	return result, nil
}
