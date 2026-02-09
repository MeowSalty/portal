package converter

import (
	"encoding/json"

	geminiTypes "github.com/MeowSalty/portal/request/adapter/gemini/types"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// FromContract 将 RequestContract 转换为 Gemini 请求。
func FromContract(contract *types.RequestContract) (*geminiTypes.Request, error) {
	if contract == nil {
		return nil, nil
	}

	req := &geminiTypes.Request{
		Model: contract.Model,
	}

	// 转换 Messages -> Contents
	if len(contract.Messages) > 0 {
		contents, err := convertMessagesFromContract(contract.Messages)
		if err != nil {
			return nil, err
		}
		req.Contents = contents
	} else if contract.Prompt != nil {
		// 若 Prompt 不为空且 Messages 为空，构造单条 user content
		req.Contents = []geminiTypes.Content{
			{
				Role: "user",
				Parts: []geminiTypes.Part{
					{
						Text: contract.Prompt,
					},
				},
			},
		}
	}

	// 转换 System -> SystemInstruction
	if contract.System != nil {
		systemInstruction, err := convertSystemFromContract(contract.System)
		if err != nil {
			return nil, err
		}
		req.SystemInstruction = systemInstruction
	}

	// 转换 GenerationConfig
	req.GenerationConfig = &geminiTypes.GenerationConfig{}
	convertGenerationConfigFromContract(contract, req.GenerationConfig)

	// 转换 Tools
	if len(contract.Tools) > 0 {
		tools, err := convertToolsFromContract(contract.Tools, contract.VendorExtras)
		if err != nil {
			return nil, err
		}
		req.Tools = tools
	}

	// 转换 ToolChoice -> ToolConfig
	if contract.ToolChoice != nil {
		req.ToolConfig = &geminiTypes.ToolConfig{}
		convertToolChoiceFromContract(contract.ToolChoice, req.ToolConfig)
	}

	// 从 VendorExtras 恢复特有字段
	if contract.VendorExtras != nil {
		if safetySettings, ok := contract.VendorExtras["safetySettings"].([]geminiTypes.SafetySetting); ok {
			req.SafetySettings = safetySettings
		}
		if cachedContent, ok := contract.VendorExtras["cachedContent"].(*string); ok {
			req.CachedContent = cachedContent
		}
		if retrievalConfig, ok := contract.VendorExtras["retrievalConfig"].(*geminiTypes.RetrievalConfig); ok {
			if req.ToolConfig == nil {
				req.ToolConfig = &geminiTypes.ToolConfig{}
			}
			req.ToolConfig.RetrievalConfig = retrievalConfig
		}
	}

	return req, nil
}

// convertMessagesFromContract 从 Contract 转换消息列表。
func convertMessagesFromContract(messages []types.Message) ([]geminiTypes.Content, error) {
	result := make([]geminiTypes.Content, 0, len(messages))

	for _, msg := range messages {
		content := geminiTypes.Content{
			Role: msg.Role,
		}

		// 转换 Content
		if msg.Content.Text != nil {
			content.Parts = []geminiTypes.Part{
				{
					Text: msg.Content.Text,
				},
			}
		} else if len(msg.Content.Parts) > 0 || len(msg.ToolCalls) > 0 {
			parts, err := convertContentPartsFromContract(msg.Content.Parts, msg.ToolCalls)
			if err != nil {
				return nil, err
			}
			content.Parts = parts
		}

		result = append(result, content)
	}

	return result, nil
}

// convertSystemFromContract 从 Contract 转换系统指令。
func convertSystemFromContract(system *types.System) (*geminiTypes.Content, error) {
	if system == nil {
		return nil, nil
	}

	content := &geminiTypes.Content{}

	if system.Text != nil {
		content.Parts = []geminiTypes.Part{
			{
				Text: system.Text,
			},
		}
	} else if len(system.Parts) > 0 {
		parts := make([]geminiTypes.Part, 0, len(system.Parts))
		for _, part := range system.Parts {
			geminiPart, err := convertContentPartFromContract(&part)
			if err != nil {
				return nil, err
			}
			if geminiPart != nil {
				parts = append(parts, *geminiPart)
			}
		}
		content.Parts = parts
	}

	return content, nil
}

// convertContentPartsFromContract 从 Contract 转换内容块列表。
func convertContentPartsFromContract(parts []types.ContentPart, toolCalls []types.ToolCall) ([]geminiTypes.Part, error) {
	result := make([]geminiTypes.Part, 0, len(parts)+len(toolCalls))

	// 转换 ContentParts
	for _, part := range parts {
		geminiPart, err := convertContentPartFromContract(&part)
		if err != nil {
			return nil, err
		}
		if geminiPart != nil {
			result = append(result, *geminiPart)
		}
	}

	// 转换 ToolCalls
	for _, tc := range toolCalls {
		geminiPart, err := convertToolCallFromContract(&tc)
		if err != nil {
			return nil, err
		}
		if geminiPart != nil {
			result = append(result, *geminiPart)
		}
	}

	return result, nil
}

// convertContentPartFromContract 从 Contract 转换单个内容块。
func convertContentPartFromContract(part *types.ContentPart) (*geminiTypes.Part, error) {
	if part == nil {
		return nil, nil
	}

	geminiPart := &geminiTypes.Part{}

	switch part.Type {
	case "text":
		if part.Text != nil {
			geminiPart.Text = part.Text
		}

	case "image":
		if part.Image != nil {
			if part.Image.Data != nil {
				// 内联数据
				geminiPart.InlineData = &geminiTypes.InlineData{
					Data: *part.Image.Data,
				}
				if part.Image.MIME != nil {
					geminiPart.InlineData.MimeType = *part.Image.MIME
				} else {
					geminiPart.InlineData.MimeType = "image/jpeg"
				}
			} else if part.Image.URL != nil {
				// 文件 URI
				geminiPart.FileData = &geminiTypes.FileData{
					FileURI: *part.Image.URL,
				}
				if part.Image.MIME != nil {
					geminiPart.FileData.MimeType = part.Image.MIME
				}
			}
		}

	case "audio":
		if part.Audio != nil {
			if part.Audio.Data != nil {
				// 内联数据
				geminiPart.InlineData = &geminiTypes.InlineData{
					Data: *part.Audio.Data,
				}
				if part.Audio.MIME != nil {
					geminiPart.InlineData.MimeType = *part.Audio.MIME
				} else {
					geminiPart.InlineData.MimeType = "audio/mpeg"
				}
			} else if part.VendorExtras != nil {
				// 从 VendorExtras 恢复 fileUri
				if fileUri, ok := part.VendorExtras["fileUri"].(string); ok {
					geminiPart.FileData = &geminiTypes.FileData{
						FileURI: fileUri,
					}
					if part.Audio.MIME != nil {
						geminiPart.FileData.MimeType = part.Audio.MIME
					}
				}
			}
		}

	case "video":
		if part.Video != nil {
			if part.Video.Data != nil {
				// 内联数据
				geminiPart.InlineData = &geminiTypes.InlineData{
					Data: *part.Video.Data,
				}
				if part.Video.MIME != nil {
					geminiPart.InlineData.MimeType = *part.Video.MIME
				} else {
					geminiPart.InlineData.MimeType = "video/mp4"
				}
			} else if part.Video.URL != nil {
				// 文件 URI
				geminiPart.FileData = &geminiTypes.FileData{
					FileURI: *part.Video.URL,
				}
				if part.Video.MIME != nil {
					geminiPart.FileData.MimeType = part.Video.MIME
				}
			}

			// 恢复视频元数据
			if part.Video.Start != nil || part.Video.End != nil || part.Video.FPS != nil {
				geminiPart.VideoMetadata = &geminiTypes.VideoMetadata{
					StartOffset: part.Video.Start,
					EndOffset:   part.Video.End,
					FPS:         part.Video.FPS,
				}
			}
		}

	case "file":
		if part.File != nil {
			if part.File.Data != nil {
				// 内联数据
				geminiPart.InlineData = &geminiTypes.InlineData{
					Data: *part.File.Data,
				}
				if part.File.MIME != nil {
					geminiPart.InlineData.MimeType = *part.File.MIME
				} else {
					geminiPart.InlineData.MimeType = "application/octet-stream"
				}
			} else if part.File.URL != nil {
				// 文件 URI
				geminiPart.FileData = &geminiTypes.FileData{
					FileURI: *part.File.URL,
				}
				if part.File.MIME != nil {
					geminiPart.FileData.MimeType = part.File.MIME
				}
			}
		}

	case "tool_result":
		if part.ToolResult != nil {
			geminiPart.FunctionResponse = &geminiTypes.FunctionResponse{
				Name: *part.ToolResult.Name,
			}

			if part.ToolResult.ID != nil {
				geminiPart.FunctionResponse.ID = part.ToolResult.ID
			}

			// 解析 Content 为 Response
			if part.ToolResult.Content != nil {
				var response map[string]interface{}
				if err := json.Unmarshal([]byte(*part.ToolResult.Content), &response); err == nil {
					geminiPart.FunctionResponse.Response = response
				} else {
					// 如果解析失败，作为字符串包装
					geminiPart.FunctionResponse.Response = map[string]interface{}{
						"result": *part.ToolResult.Content,
					}
				}
			}

			// 从 Payload 恢复其他字段
			if part.ToolResult.Payload != nil {
				if parts, ok := part.ToolResult.Payload["parts"].([]geminiTypes.FunctionResponsePart); ok {
					geminiPart.FunctionResponse.Parts = parts
				}
				if willContinue, ok := part.ToolResult.Payload["willContinue"].(*bool); ok {
					geminiPart.FunctionResponse.WillContinue = willContinue
				}
				if scheduling, ok := part.ToolResult.Payload["scheduling"].(*string); ok {
					geminiPart.FunctionResponse.Scheduling = scheduling
				}
			}
		}

	case "executable_code":
		// 从 VendorExtras 恢复
		if part.VendorExtras != nil {
			if ec, ok := part.VendorExtras["executableCode"].(*geminiTypes.ExecutableCode); ok {
				geminiPart.ExecutableCode = ec
			}
		}

	case "code_execution_result":
		// 从 VendorExtras 恢复
		if part.VendorExtras != nil {
			if cer, ok := part.VendorExtras["codeExecutionResult"].(*geminiTypes.CodeExecutionResult); ok {
				geminiPart.CodeExecutionResult = cer
			}
		}

	default:
		return nil, nil
	}

	// 恢复 Part 级别的特有字段
	if part.VendorExtras != nil {
		if videoMetadata, ok := part.VendorExtras["videoMetadata"].(*geminiTypes.VideoMetadata); ok {
			geminiPart.VideoMetadata = videoMetadata
		}
		if thought, ok := part.VendorExtras["thought"].(*bool); ok {
			geminiPart.Thought = thought
		}
		if thoughtSignature, ok := part.VendorExtras["thoughtSignature"].(*string); ok {
			geminiPart.ThoughtSignature = thoughtSignature
		}
		if partMetadata, ok := part.VendorExtras["partMetadata"].(map[string]interface{}); ok {
			geminiPart.PartMetadata = partMetadata
		}
		if mediaResolution, ok := part.VendorExtras["mediaResolution"].(*geminiTypes.MediaResolution); ok {
			geminiPart.MediaResolution = mediaResolution
		}
	}

	return geminiPart, nil
}

// convertToolCallFromContract 从 Contract 转换工具调用。
func convertToolCallFromContract(tc *types.ToolCall) (*geminiTypes.Part, error) {
	if tc == nil || tc.Name == nil {
		return nil, nil
	}

	part := &geminiTypes.Part{
		FunctionCall: &geminiTypes.FunctionCall{
			Name: *tc.Name,
		},
	}

	if tc.ID != nil {
		part.FunctionCall.ID = tc.ID
	}

	// 解析 Arguments
	if tc.Arguments != nil {
		var args map[string]interface{}
		if err := json.Unmarshal([]byte(*tc.Arguments), &args); err == nil {
			part.FunctionCall.Args = args
		}
	}

	return part, nil
}

// convertGenerationConfigFromContract 从 Contract 转换生成配置。
func convertGenerationConfigFromContract(contract *types.RequestContract, gc *geminiTypes.GenerationConfig) {
	// 采样参数
	gc.MaxOutputTokens = contract.MaxOutputTokens
	gc.Temperature = contract.Temperature
	gc.TopP = contract.TopP
	gc.TopK = contract.TopK
	gc.Seed = contract.Seed
	gc.PresencePenalty = contract.PresencePenalty
	gc.FrequencyPenalty = contract.FrequencyPenalty
	gc.CandidateCount = contract.CandidateCount

	// StopSequences
	if contract.Stop != nil {
		if len(contract.Stop.List) > 0 {
			gc.StopSequences = contract.Stop.List
		} else if contract.Stop.Text != nil {
			gc.StopSequences = []string{*contract.Stop.Text}
		}
	}

	// Logprobs
	gc.ResponseLogprobs = contract.Logprobs
	gc.Logprobs = contract.TopLogprobs

	// ResponseModalities
	if len(contract.Modalities) > 0 {
		gc.ResponseModalities = contract.Modalities
	}

	// ResponseFormat
	if contract.ResponseFormat != nil {
		if contract.ResponseFormat.MimeType != nil {
			gc.ResponseMimeType = contract.ResponseFormat.MimeType
		} else {
			// 根据 Type 设置 MimeType
			switch contract.ResponseFormat.Type {
			case "json":
				mimeType := "application/json"
				gc.ResponseMimeType = &mimeType
			case "text":
				mimeType := "text/plain"
				gc.ResponseMimeType = &mimeType
			}
		}

		// 转换 JSONSchema
		if contract.ResponseFormat.JSONSchema != nil {
			if schema, ok := contract.ResponseFormat.JSONSchema.(*geminiTypes.Schema); ok {
				gc.ResponseSchema = schema
			} else {
				// 尝试作为 interface{} 使用
				gc.ResponseJSONSchemaRaw = contract.ResponseFormat.JSONSchema
			}
		}
	}

	// Reasoning -> ThinkingConfig
	if contract.Reasoning != nil {
		gc.ThinkingConfig = &geminiTypes.ThinkingConfig{
			IncludeThoughts: contract.Reasoning.IncludeThoughts,
			ThinkingBudget:  contract.Reasoning.Budget,
			ThinkingLevel:   contract.Reasoning.Level,
		}
	}

	// 从 VendorExtras 恢复特有字段
	if contract.VendorExtras != nil {
		if enableEnhancedCivicAnswers, ok := contract.VendorExtras["enableEnhancedCivicAnswers"].(*bool); ok {
			gc.EnableEnhancedCivicAnswers = enableEnhancedCivicAnswers
		}
		if speechConfig, ok := contract.VendorExtras["speechConfig"].(*geminiTypes.SpeechConfig); ok {
			gc.SpeechConfig = speechConfig
		}
		if imageConfig, ok := contract.VendorExtras["imageConfig"].(*geminiTypes.ImageConfig); ok {
			gc.ImageConfig = imageConfig
		}
		if mediaResolution, ok := contract.VendorExtras["mediaResolution"].(*string); ok {
			gc.MediaResolution = mediaResolution
		}
	}
}

// convertToolsFromContract 从 Contract 转换工具列表。
func convertToolsFromContract(tools []types.Tool, vendorExtras map[string]interface{}) ([]geminiTypes.Tool, error) {
	result := make([]geminiTypes.Tool, 0)

	// 收集所有 FunctionDeclarations
	functionDeclarations := make([]geminiTypes.FunctionDeclaration, 0)

	for _, tool := range tools {
		if tool.Type == "function" && tool.Function != nil {
			fd := geminiTypes.FunctionDeclaration{
				Name:        tool.Function.Name,
				Description: "",
			}

			if tool.Function.Description != nil {
				fd.Description = *tool.Function.Description
			}

			// 转换 Parameters
			if tool.Function.Parameters != nil {
				if schema, ok := tool.Function.Parameters.(*geminiTypes.Schema); ok {
					fd.Parameters = schema
				} else {
					// 尝试作为 interface{} 使用
					fd.ParametersJSONSchema = tool.Function.Parameters
				}
			}

			// 转换 ResponseSchema
			if tool.Function.ResponseSchema != nil {
				if schema, ok := tool.Function.ResponseSchema.(*geminiTypes.Schema); ok {
					fd.Response = schema
				} else {
					fd.ResponseJSONSchema = tool.Function.ResponseSchema
				}
			}

			// 从 VendorExtras 恢复 Behavior
			if tool.VendorExtras != nil {
				if behavior, ok := tool.VendorExtras["behavior"].(*string); ok {
					fd.Behavior = behavior
				}
			}

			functionDeclarations = append(functionDeclarations, fd)
		}
	}

	// 如果有 FunctionDeclarations，创建一个 Tool
	if len(functionDeclarations) > 0 {
		result = append(result, geminiTypes.Tool{
			FunctionDeclarations: functionDeclarations,
		})
	}

	// 从 VendorExtras 恢复特殊工具类型
	if vendorExtras != nil {
		if toolsExtras, ok := vendorExtras["tools_extras"].([]map[string]interface{}); ok {
			for _, extra := range toolsExtras {
				toolType, _ := extra["type"].(string)
				tool := geminiTypes.Tool{}

				switch toolType {
				case "googleSearchRetrieval":
					if gsr, ok := extra["tool"].(*geminiTypes.GoogleSearchRetrieval); ok {
						tool.GoogleSearchRetrieval = gsr
						result = append(result, tool)
					}
				case "codeExecution":
					if ce, ok := extra["tool"].(*geminiTypes.CodeExecution); ok {
						tool.CodeExecution = ce
						result = append(result, tool)
					}
				case "googleSearch":
					if gs, ok := extra["tool"].(*geminiTypes.GoogleSearch); ok {
						tool.GoogleSearch = gs
						result = append(result, tool)
					}
				case "computerUse":
					if cu, ok := extra["tool"].(*geminiTypes.ComputerUse); ok {
						tool.ComputerUse = cu
						result = append(result, tool)
					}
				case "urlContext":
					if uc, ok := extra["tool"].(*geminiTypes.UrlContext); ok {
						tool.URLContext = uc
						result = append(result, tool)
					}
				case "fileSearch":
					if fs, ok := extra["tool"].(*geminiTypes.FileSearch); ok {
						tool.FileSearch = fs
						result = append(result, tool)
					}
				case "mcpServers":
					if mcps, ok := extra["tool"].([]geminiTypes.McpServer); ok {
						tool.MCPServers = mcps
						result = append(result, tool)
					}
				case "googleMaps":
					if gm, ok := extra["tool"].(*geminiTypes.GoogleMaps); ok {
						tool.GoogleMaps = gm
						result = append(result, tool)
					}
				}
			}
		}
	}

	return result, nil
}

// convertToolChoiceFromContract 从 Contract 转换工具选择。
func convertToolChoiceFromContract(toolChoice *types.ToolChoice, toolConfig *geminiTypes.ToolConfig) {
	if toolChoice == nil {
		return
	}

	toolConfig.FunctionCallingConfig = &geminiTypes.FunctionCallingConfig{}

	// 转换 Mode
	if toolChoice.Mode != nil {
		toolConfig.FunctionCallingConfig.Mode = *toolChoice.Mode
	}

	// 转换 Allowed
	if len(toolChoice.Allowed) > 0 {
		toolConfig.FunctionCallingConfig.AllowedFunctionNames = toolChoice.Allowed
	} else if toolChoice.Function != nil {
		// 如果指定了单个函数，设置为 AllowedFunctionNames
		toolConfig.FunctionCallingConfig.AllowedFunctionNames = []string{*toolChoice.Function}
	}
}
