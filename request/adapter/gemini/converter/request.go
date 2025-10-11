package converter

import (
	"github.com/MeowSalty/portal/request/adapter/gemini/types"
	"github.com/MeowSalty/portal/routing"
	coreTypes "github.com/MeowSalty/portal/types"
)

// ConvertRequest 将核心请求转换为 Gemini 请求
func ConvertRequest(request *coreTypes.Request, channel *routing.Channel) *types.Request {
	geminiReq := &types.Request{
		Contents: convertMessagesToContents(request.Messages),
	}

	// 处理系统消息
	if systemInstruction := extractSystemInstruction(request.Messages); systemInstruction != nil {
		geminiReq.SystemInstruction = systemInstruction
	}

	// 转换温度参数
	if request.Temperature != nil {
		if geminiReq.GenerationConfig == nil {
			geminiReq.GenerationConfig = &types.GenerationConfig{}
		}
		geminiReq.GenerationConfig.Temperature = request.Temperature
	}

	// 转换 TopP 参数
	if request.TopP != nil {
		if geminiReq.GenerationConfig == nil {
			geminiReq.GenerationConfig = &types.GenerationConfig{}
		}
		geminiReq.GenerationConfig.TopP = request.TopP
	}

	// 转换最大 token 数
	if request.MaxTokens != nil {
		if geminiReq.GenerationConfig == nil {
			geminiReq.GenerationConfig = &types.GenerationConfig{}
		}
		geminiReq.GenerationConfig.MaxOutputTokens = request.MaxTokens
	}

	// 转换停止序列
	if request.Stop.StringValue != nil {
		if geminiReq.GenerationConfig == nil {
			geminiReq.GenerationConfig = &types.GenerationConfig{}
		}
		geminiReq.GenerationConfig.StopSequences = []string{*request.Stop.StringValue}
	} else if len(request.Stop.StringArray) > 0 {
		if geminiReq.GenerationConfig == nil {
			geminiReq.GenerationConfig = &types.GenerationConfig{}
		}
		geminiReq.GenerationConfig.StopSequences = request.Stop.StringArray
	}

	// 转换工具（如果存在）
	if len(request.Tools) > 0 {
		geminiReq.Tools = []types.Tool{
			{
				FunctionDeclarations: convertTools(request.Tools),
			},
		}
	}

	// 转换工具选择
	if request.ToolChoice.Mode != nil {
		if geminiReq.ToolConfig == nil {
			geminiReq.ToolConfig = &types.ToolConfig{}
		}
		if geminiReq.ToolConfig.FunctionCallingConfig == nil {
			geminiReq.ToolConfig.FunctionCallingConfig = &types.FunctionCallingConfig{}
		}

		switch *request.ToolChoice.Mode {
		case "auto":
			geminiReq.ToolConfig.FunctionCallingConfig.Mode = "AUTO"
		case "none":
			geminiReq.ToolConfig.FunctionCallingConfig.Mode = "NONE"
		}
	} else if request.ToolChoice.Function != nil {
		if geminiReq.ToolConfig == nil {
			geminiReq.ToolConfig = &types.ToolConfig{}
		}
		if geminiReq.ToolConfig.FunctionCallingConfig == nil {
			geminiReq.ToolConfig.FunctionCallingConfig = &types.FunctionCallingConfig{}
		}
		geminiReq.ToolConfig.FunctionCallingConfig.Mode = "ANY"
		// 注意：Gemini 不支持指定特定函数，只能设置模式
	}

	return geminiReq
}

// convertMessagesToContents 将核心消息转换为 Gemini 内容数组
func convertMessagesToContents(messages []coreTypes.Message) []types.Content {
	var contents []types.Content

	for _, msg := range messages {
		var geminiRole string

		// 转换角色：assistant -> model, user -> user, 其他角色需要特殊处理
		switch msg.Role {
		case "assistant":
			geminiRole = "model"
		case "user":
			geminiRole = "user"
		case "system":
			// system 消息通常作为第一个 user 消息的一部分处理
			// 这里我们将其转换为 user 角色，但需要特殊标记或合并
			geminiRole = "user"
		case "tool":
			// tool 消息需要转换为 functionResponse
			// 这里暂时跳过，需要更复杂的处理
			continue
		default:
			// 未知角色跳过
			continue
		}

		// 转换消息内容为 parts
		parts := convertMessageToParts(msg)
		if len(parts) > 0 {
			contents = append(contents, types.Content{
				Role:  geminiRole,
				Parts: parts,
			})
		}
	}

	return contents
}

// convertMessageToParts 将单个消息转换为 Gemini 部分
func convertMessageToParts(msg coreTypes.Message) []types.Part {
	var parts []types.Part

	if msg.Content.StringValue != nil {
		// 文本内容
		text := *msg.Content.StringValue
		parts = append(parts, types.Part{
			Text: &text,
		})
	} else if len(msg.Content.ContentParts) > 0 {
		// 内容部分数组
		for _, part := range msg.Content.ContentParts {
			if part.Text != nil {
				// 文本部分
				text := *part.Text
				parts = append(parts, types.Part{
					Text: &text,
				})
			} else if part.ImageURL != nil {
				// 图像部分 - 注意：Gemini 需要 base64 编码的内联数据
				// 这里只处理文本，图像需要额外处理
				// 可以添加日志或错误处理
			}
		}
	}

	return parts
}

// convertTools 将核心工具转换为 Gemini 函数声明
func convertTools(tools []coreTypes.Tool) []types.FunctionDeclaration {
	var functionDeclarations []types.FunctionDeclaration

	for _, tool := range tools {
		if tool.Type == "function" {
			// 转换参数为 JSON schema
			var parameters interface{}
			if tool.Function.Parameters != nil {
				// 如果参数已经是 map 或结构体，直接使用
				parameters = tool.Function.Parameters
			} else {
				// 创建默认的 schema
				parameters = map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				}
			}

			functionDeclarations = append(functionDeclarations, types.FunctionDeclaration{
				Name:        tool.Function.Name,
				Description: *tool.Function.Description,
				Parameters:  parameters,
			})
		}
	}

	return functionDeclarations
}

// ConvertCoreRequest 将 Gemini 请求转换为核心请求（反向转换）
func ConvertCoreRequest(geminiReq *types.Request) *coreTypes.Request {
	coreReq := &coreTypes.Request{}

	// 转换生成配置
	if geminiReq.GenerationConfig != nil {
		if geminiReq.GenerationConfig.Temperature != nil {
			coreReq.Temperature = geminiReq.GenerationConfig.Temperature
		}
		if geminiReq.GenerationConfig.TopP != nil {
			coreReq.TopP = geminiReq.GenerationConfig.TopP
		}
		if geminiReq.GenerationConfig.MaxOutputTokens != nil {
			coreReq.MaxTokens = geminiReq.GenerationConfig.MaxOutputTokens
		}
		if len(geminiReq.GenerationConfig.StopSequences) > 0 {
			if len(geminiReq.GenerationConfig.StopSequences) == 1 {
				coreReq.Stop.StringValue = &geminiReq.GenerationConfig.StopSequences[0]
			} else {
				coreReq.Stop.StringArray = geminiReq.GenerationConfig.StopSequences
			}
		}
	}

	// 转换工具
	if len(geminiReq.Tools) > 0 && len(geminiReq.Tools[0].FunctionDeclarations) > 0 {
		for _, funcDecl := range geminiReq.Tools[0].FunctionDeclarations {
			coreReq.Tools = append(coreReq.Tools, coreTypes.Tool{
				Type: "function",
				Function: coreTypes.FunctionDescription{
					Name:        funcDecl.Name,
					Description: &funcDecl.Description,
					Parameters:  funcDecl.Parameters,
				},
			})
		}
	}

	// 转换工具选择
	if geminiReq.ToolConfig != nil && geminiReq.ToolConfig.FunctionCallingConfig != nil {
		mode := geminiReq.ToolConfig.FunctionCallingConfig.Mode
		switch mode {
		case "AUTO":
			auto := "auto"
			coreReq.ToolChoice.Mode = &auto
		case "NONE":
			none := "none"
			coreReq.ToolChoice.Mode = &none
		case "ANY":
			// Gemini 的 ANY 模式对应 OpenAI 的自动模式
			auto := "auto"
			coreReq.ToolChoice.Mode = &auto
		}
	}

	// 转换消息内容
	if len(geminiReq.Contents) > 0 {
		for _, content := range geminiReq.Contents {
			if content.Role == "user" {
				for _, part := range content.Parts {
					if part.Text != nil {
						message := coreTypes.Message{
							Role: "user",
							Content: coreTypes.MessageContent{
								StringValue: part.Text,
							},
						}
						coreReq.Messages = append(coreReq.Messages, message)
					}
				}
			}
		}
	}

	return coreReq
}

// extractSystemInstruction 从消息中提取系统指令
func extractSystemInstruction(messages []coreTypes.Message) *types.Content {
	var systemParts []types.Part

	for _, msg := range messages {
		if msg.Role == "system" {
			if msg.Content.StringValue != nil {
				text := *msg.Content.StringValue
				systemParts = append(systemParts, types.Part{
					Text: &text,
				})
			} else if len(msg.Content.ContentParts) > 0 {
				for _, part := range msg.Content.ContentParts {
					if part.Text != nil {
						text := *part.Text
						systemParts = append(systemParts, types.Part{
							Text: &text,
						})
					}
				}
			}
		}
	}

	if len(systemParts) > 0 {
		return &types.Content{
			Parts: systemParts,
		}
	}

	return nil
}
