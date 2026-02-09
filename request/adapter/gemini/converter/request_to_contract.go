package converter

import (
	"encoding/json"

	geminiTypes "github.com/MeowSalty/portal/request/adapter/gemini/types"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// RequestToContract 将 Gemini 请求转换为统一的 RequestContract。
func RequestToContract(req *geminiTypes.Request) (*types.RequestContract, error) {
	if req == nil {
		return nil, nil
	}

	contract := &types.RequestContract{
		Source: types.VendorSourceGemini,
		Model:  req.Model,
	}

	// 转换 Contents -> Messages
	if len(req.Contents) > 0 {
		messages, err := convertContentsToContract(req.Contents)
		if err != nil {
			return nil, err
		}
		contract.Messages = messages
	}

	// 转换 SystemInstruction -> System
	if req.SystemInstruction != nil {
		system, err := convertSystemInstructionToContract(req.SystemInstruction)
		if err != nil {
			return nil, err
		}
		contract.System = system
	}

	// 转换 GenerationConfig
	if req.GenerationConfig != nil {
		convertGenerationConfigToContract(req.GenerationConfig, contract)
	}

	// 转换 Tools
	if len(req.Tools) > 0 {
		tools, vendorExtras, err := convertToolsToContract(req.Tools)
		if err != nil {
			return nil, err
		}
		contract.Tools = tools

		// 合并工具的 VendorExtras
		if len(vendorExtras) > 0 {
			if contract.VendorExtras == nil {
				contract.VendorExtras = make(map[string]interface{})
			}
			contract.VendorExtras["tools_extras"] = vendorExtras
		}
	}

	// 转换 ToolConfig
	if req.ToolConfig != nil {
		convertToolConfigToContract(req.ToolConfig, contract)
	}

	// 初始化 VendorExtras
	if contract.VendorExtras == nil {
		contract.VendorExtras = make(map[string]interface{})
	}
	source := types.VendorSourceGemini
	contract.VendorExtrasSource = &source

	// 特有字段放入 VendorExtras
	if len(req.SafetySettings) > 0 {
		contract.VendorExtras["safetySettings"] = req.SafetySettings
	}
	if req.CachedContent != nil {
		contract.VendorExtras["cachedContent"] = req.CachedContent
	}

	return contract, nil
}

// convertContentsToContract 转换内容列表为消息列表。
func convertContentsToContract(contents []geminiTypes.Content) ([]types.Message, error) {
	result := make([]types.Message, 0, len(contents))

	for _, content := range contents {
		msg := types.Message{
			Role: content.Role,
		}

		// 转换 Parts
		if len(content.Parts) > 0 {
			parts, toolCalls, err := convertPartsToContract(content.Parts)
			if err != nil {
				return nil, err
			}

			msg.Content = types.Content{
				Parts: parts,
			}

			if len(toolCalls) > 0 {
				msg.ToolCalls = toolCalls
			}
		}

		result = append(result, msg)
	}

	return result, nil
}

// convertSystemInstructionToContract 转换系统指令。
func convertSystemInstructionToContract(content *geminiTypes.Content) (*types.System, error) {
	if content == nil {
		return nil, nil
	}

	system := &types.System{}

	if len(content.Parts) > 0 {
		parts := make([]types.ContentPart, 0, len(content.Parts))
		for _, part := range content.Parts {
			contractPart, err := convertPartToContract(&part)
			if err != nil {
				return nil, err
			}
			if contractPart != nil {
				parts = append(parts, *contractPart)
			}
		}
		system.Parts = parts
	}

	return system, nil
}

// convertPartsToContract 转换 Part 列表。
func convertPartsToContract(parts []geminiTypes.Part) ([]types.ContentPart, []types.ToolCall, error) {
	contractParts := make([]types.ContentPart, 0, len(parts))
	toolCalls := make([]types.ToolCall, 0)

	for _, part := range parts {
		// 处理 FunctionCall
		if part.FunctionCall != nil {
			toolCall, err := convertFunctionCallToContract(part.FunctionCall)
			if err != nil {
				return nil, nil, err
			}
			toolCalls = append(toolCalls, *toolCall)
			continue
		}

		// 处理 FunctionResponse
		if part.FunctionResponse != nil {
			contractPart, err := convertFunctionResponseToContract(part.FunctionResponse)
			if err != nil {
				return nil, nil, err
			}
			contractParts = append(contractParts, *contractPart)
			continue
		}

		// 处理其他类型的 Part
		contractPart, err := convertPartToContract(&part)
		if err != nil {
			return nil, nil, err
		}
		if contractPart != nil {
			contractParts = append(contractParts, *contractPart)
		}
	}

	return contractParts, toolCalls, nil
}

// convertPartToContract 转换单个 Part。
func convertPartToContract(part *geminiTypes.Part) (*types.ContentPart, error) {
	if part == nil {
		return nil, nil
	}

	contractPart := &types.ContentPart{}

	// 文本内容
	if part.Text != nil {
		contractPart.Type = "text"
		contractPart.Text = part.Text
	} else if part.InlineData != nil {
		// 内联数据（图像、音频等）
		return convertInlineDataToContract(part.InlineData, part)
	} else if part.FileData != nil {
		// 文件数据
		return convertFileDataToContract(part.FileData, part)
	} else if part.ExecutableCode != nil {
		// 可执行代码
		contractPart.Type = "executable_code"
		contractPart.VendorExtras = make(map[string]interface{})
		source := types.VendorSourceGemini
		contractPart.VendorExtrasSource = &source
		contractPart.VendorExtras["executableCode"] = part.ExecutableCode
	} else if part.CodeExecutionResult != nil {
		// 代码执行结果
		contractPart.Type = "code_execution_result"
		contractPart.VendorExtras = make(map[string]interface{})
		source := types.VendorSourceGemini
		contractPart.VendorExtrasSource = &source
		contractPart.VendorExtras["codeExecutionResult"] = part.CodeExecutionResult
	} else {
		return nil, nil
	}

	// 处理 Part 级别的特有字段
	if part.VideoMetadata != nil || part.Thought != nil || part.ThoughtSignature != nil ||
		part.PartMetadata != nil || part.MediaResolution != nil {
		if contractPart.VendorExtras == nil {
			contractPart.VendorExtras = make(map[string]interface{})
		}
		source := types.VendorSourceGemini
		contractPart.VendorExtrasSource = &source

		if part.VideoMetadata != nil {
			contractPart.VendorExtras["videoMetadata"] = part.VideoMetadata
		}
		if part.Thought != nil {
			contractPart.VendorExtras["thought"] = part.Thought
		}
		if part.ThoughtSignature != nil {
			contractPart.VendorExtras["thoughtSignature"] = part.ThoughtSignature
		}
		if part.PartMetadata != nil {
			contractPart.VendorExtras["partMetadata"] = part.PartMetadata
		}
		if part.MediaResolution != nil {
			contractPart.VendorExtras["mediaResolution"] = part.MediaResolution
		}
	}

	return contractPart, nil
}

// convertInlineDataToContract 转换内联数据。
func convertInlineDataToContract(inlineData *geminiTypes.InlineData, part *geminiTypes.Part) (*types.ContentPart, error) {
	contractPart := &types.ContentPart{}

	// 根据 MIME 类型判断内容类型
	mimeType := inlineData.MimeType
	if len(mimeType) >= 5 && mimeType[:5] == "image" {
		contractPart.Type = "image"
		contractPart.Image = &types.Image{
			Data: &inlineData.Data,
			MIME: &inlineData.MimeType,
		}
	} else if len(mimeType) >= 5 && mimeType[:5] == "audio" {
		contractPart.Type = "audio"
		contractPart.Audio = &types.Audio{
			Data: &inlineData.Data,
			MIME: &inlineData.MimeType,
		}
	} else if len(mimeType) >= 5 && mimeType[:5] == "video" {
		contractPart.Type = "video"
		contractPart.Video = &types.Video{
			Data: &inlineData.Data,
			MIME: &inlineData.MimeType,
		}

		// 添加视频元数据
		if part.VideoMetadata != nil {
			contractPart.Video.Start = part.VideoMetadata.StartOffset
			contractPart.Video.End = part.VideoMetadata.EndOffset
			contractPart.Video.FPS = part.VideoMetadata.FPS
		}
	} else {
		// 其他类型作为文件
		contractPart.Type = "file"
		contractPart.File = &types.File{
			Data: &inlineData.Data,
			MIME: &inlineData.MimeType,
		}
	}

	return contractPart, nil
}

// convertFileDataToContract 转换文件数据。
func convertFileDataToContract(fileData *geminiTypes.FileData, part *geminiTypes.Part) (*types.ContentPart, error) {
	contractPart := &types.ContentPart{}

	// 根据 MIME 类型判断内容类型
	if fileData.MimeType != nil {
		mimeType := *fileData.MimeType
		if len(mimeType) >= 5 && mimeType[:5] == "image" {
			contractPart.Type = "image"
			contractPart.Image = &types.Image{
				URL:  &fileData.FileURI,
				MIME: fileData.MimeType,
			}
		} else if len(mimeType) >= 5 && mimeType[:5] == "audio" {
			contractPart.Type = "audio"
			contractPart.Audio = &types.Audio{
				MIME: fileData.MimeType,
			}
			contractPart.VendorExtras = make(map[string]interface{})
			source := types.VendorSourceGemini
			contractPart.VendorExtrasSource = &source
			contractPart.VendorExtras["fileUri"] = fileData.FileURI
		} else if len(mimeType) >= 5 && mimeType[:5] == "video" {
			contractPart.Type = "video"
			contractPart.Video = &types.Video{
				URL:  &fileData.FileURI,
				MIME: fileData.MimeType,
			}

			// 添加视频元数据
			if part.VideoMetadata != nil {
				contractPart.Video.Start = part.VideoMetadata.StartOffset
				contractPart.Video.End = part.VideoMetadata.EndOffset
				contractPart.Video.FPS = part.VideoMetadata.FPS
			}
		} else {
			contractPart.Type = "file"
			contractPart.File = &types.File{
				URL:  &fileData.FileURI,
				MIME: fileData.MimeType,
			}
		}
	} else {
		// 无 MIME 类型，默认作为文件
		contractPart.Type = "file"
		contractPart.File = &types.File{
			URL: &fileData.FileURI,
		}
	}

	return contractPart, nil
}

// convertFunctionCallToContract 转换函数调用。
func convertFunctionCallToContract(fc *geminiTypes.FunctionCall) (*types.ToolCall, error) {
	toolCall := &types.ToolCall{
		Name: &fc.Name,
	}

	if fc.ID != nil {
		toolCall.ID = fc.ID
	}

	// 序列化 Args 为 JSON 字符串
	if fc.Args != nil {
		argsJSON, err := json.Marshal(fc.Args)
		if err != nil {
			return nil, err
		}
		argsStr := string(argsJSON)
		toolCall.Arguments = &argsStr
	}

	toolType := "function"
	toolCall.Type = &toolType

	return toolCall, nil
}

// convertFunctionResponseToContract 转换函数响应。
func convertFunctionResponseToContract(fr *geminiTypes.FunctionResponse) (*types.ContentPart, error) {
	part := &types.ContentPart{
		Type: "tool_result",
		ToolResult: &types.ToolResult{
			Name: &fr.Name,
		},
	}

	if fr.ID != nil {
		part.ToolResult.ID = fr.ID
	}

	// 将 Response 序列化为 JSON 字符串
	if fr.Response != nil {
		responseJSON, err := json.Marshal(fr.Response)
		if err != nil {
			return nil, err
		}
		responseStr := string(responseJSON)
		part.ToolResult.Content = &responseStr
	}

	// 其他字段放入 VendorExtras
	if len(fr.Parts) > 0 || fr.WillContinue != nil || fr.Scheduling != nil {
		part.ToolResult.Payload = make(map[string]interface{})
		if len(fr.Parts) > 0 {
			part.ToolResult.Payload["parts"] = fr.Parts
		}
		if fr.WillContinue != nil {
			part.ToolResult.Payload["willContinue"] = fr.WillContinue
		}
		if fr.Scheduling != nil {
			part.ToolResult.Payload["scheduling"] = fr.Scheduling
		}
	}

	return part, nil
}

// convertGenerationConfigToContract 转换生成配置。
func convertGenerationConfigToContract(gc *geminiTypes.GenerationConfig, contract *types.RequestContract) {
	// 采样参数
	contract.MaxOutputTokens = gc.MaxOutputTokens
	contract.Temperature = gc.Temperature
	contract.TopP = gc.TopP
	contract.TopK = gc.TopK
	contract.Seed = gc.Seed
	contract.PresencePenalty = gc.PresencePenalty
	contract.FrequencyPenalty = gc.FrequencyPenalty
	contract.CandidateCount = gc.CandidateCount

	// StopSequences
	if len(gc.StopSequences) > 0 {
		contract.Stop = &types.Stop{
			List: gc.StopSequences,
		}
	}

	// Logprobs
	contract.Logprobs = gc.ResponseLogprobs
	contract.TopLogprobs = gc.Logprobs

	// ResponseModalities
	if len(gc.ResponseModalities) > 0 {
		contract.Modalities = gc.ResponseModalities
	}

	// ResponseFormat
	if gc.ResponseMimeType != nil || gc.ResponseSchema != nil || gc.ResponseJSONSchemaRaw != nil || gc.ResponseJsonSchema != nil {
		contract.ResponseFormat = &types.ResponseFormat{}

		if gc.ResponseMimeType != nil {
			contract.ResponseFormat.MimeType = gc.ResponseMimeType

			// 根据 MIME 类型设置 Type
			switch *gc.ResponseMimeType {
			case "application/json":
				contract.ResponseFormat.Type = "json"
			case "text/plain":
				contract.ResponseFormat.Type = "text"
			default:
				contract.ResponseFormat.Type = "custom"
			}
		}

		// 优先使用 ResponseSchema
		if gc.ResponseSchema != nil {
			contract.ResponseFormat.JSONSchema = gc.ResponseSchema
		} else if gc.ResponseJSONSchemaRaw != nil {
			contract.ResponseFormat.JSONSchema = gc.ResponseJSONSchemaRaw
		} else if gc.ResponseJsonSchema != nil {
			contract.ResponseFormat.JSONSchema = gc.ResponseJsonSchema
		}
	}

	// ThinkingConfig -> Reasoning
	if gc.ThinkingConfig != nil {
		contract.Reasoning = &types.Reasoning{
			IncludeThoughts: gc.ThinkingConfig.IncludeThoughts,
			Budget:          gc.ThinkingConfig.ThinkingBudget,
			Level:           gc.ThinkingConfig.ThinkingLevel,
		}
	}

	// 特有字段放入 VendorExtras
	if contract.VendorExtras == nil {
		contract.VendorExtras = make(map[string]interface{})
	}

	if gc.EnableEnhancedCivicAnswers != nil {
		contract.VendorExtras["enableEnhancedCivicAnswers"] = gc.EnableEnhancedCivicAnswers
	}
	if gc.SpeechConfig != nil {
		contract.VendorExtras["speechConfig"] = gc.SpeechConfig
	}
	if gc.ImageConfig != nil {
		contract.VendorExtras["imageConfig"] = gc.ImageConfig
	}
	if gc.MediaResolution != nil {
		contract.VendorExtras["mediaResolution"] = gc.MediaResolution
	}
}

// convertToolsToContract 转换工具列表。
func convertToolsToContract(tools []geminiTypes.Tool) ([]types.Tool, []map[string]interface{}, error) {
	result := make([]types.Tool, 0)
	vendorExtras := make([]map[string]interface{}, 0)

	for _, tool := range tools {
		// 处理 FunctionDeclarations
		if len(tool.FunctionDeclarations) > 0 {
			for _, fd := range tool.FunctionDeclarations {
				contractTool := types.Tool{
					Type: "function",
					Function: &types.Function{
						Name:        fd.Name,
						Description: &fd.Description,
					},
				}

				// 转换 Parameters
				if fd.Parameters != nil {
					contractTool.Function.Parameters = fd.Parameters
				} else if fd.ParametersJSONSchema != nil {
					contractTool.Function.Parameters = fd.ParametersJSONSchema
				}

				// 转换 Response
				if fd.Response != nil {
					contractTool.Function.ResponseSchema = fd.Response
				} else if fd.ResponseJSONSchema != nil {
					contractTool.Function.ResponseSchema = fd.ResponseJSONSchema
				}

				// Behavior 放入 VendorExtras
				if fd.Behavior != nil {
					contractTool.VendorExtras = make(map[string]interface{})
					source := types.VendorSourceGemini
					contractTool.VendorExtrasSource = &source
					contractTool.VendorExtras["behavior"] = fd.Behavior
				}

				result = append(result, contractTool)
			}
		}

		// 其他特殊工具类型放入 VendorExtras
		if tool.GoogleSearchRetrieval != nil {
			extras := map[string]interface{}{
				"type": "googleSearchRetrieval",
				"tool": tool.GoogleSearchRetrieval,
			}
			vendorExtras = append(vendorExtras, extras)
		}
		if tool.CodeExecution != nil {
			extras := map[string]interface{}{
				"type": "codeExecution",
				"tool": tool.CodeExecution,
			}
			vendorExtras = append(vendorExtras, extras)
		}
		if tool.GoogleSearch != nil {
			extras := map[string]interface{}{
				"type": "googleSearch",
				"tool": tool.GoogleSearch,
			}
			vendorExtras = append(vendorExtras, extras)
		}
		if tool.ComputerUse != nil {
			extras := map[string]interface{}{
				"type": "computerUse",
				"tool": tool.ComputerUse,
			}
			vendorExtras = append(vendorExtras, extras)
		}
		if tool.URLContext != nil {
			extras := map[string]interface{}{
				"type": "urlContext",
				"tool": tool.URLContext,
			}
			vendorExtras = append(vendorExtras, extras)
		}
		if tool.FileSearch != nil {
			extras := map[string]interface{}{
				"type": "fileSearch",
				"tool": tool.FileSearch,
			}
			vendorExtras = append(vendorExtras, extras)
		}
		if len(tool.MCPServers) > 0 {
			extras := map[string]interface{}{
				"type": "mcpServers",
				"tool": tool.MCPServers,
			}
			vendorExtras = append(vendorExtras, extras)
		}
		if tool.GoogleMaps != nil {
			extras := map[string]interface{}{
				"type": "googleMaps",
				"tool": tool.GoogleMaps,
			}
			vendorExtras = append(vendorExtras, extras)
		}
	}

	return result, vendorExtras, nil
}

// convertToolConfigToContract 转换工具配置。
func convertToolConfigToContract(tc *geminiTypes.ToolConfig, contract *types.RequestContract) {
	// FunctionCallingConfig -> ToolChoice
	if tc.FunctionCallingConfig != nil {
		fcc := tc.FunctionCallingConfig
		contract.ToolChoice = &types.ToolChoice{}

		// 转换 Mode
		if fcc.Mode != "" {
			mode := fcc.Mode
			contract.ToolChoice.Mode = &mode
		}

		// 转换 AllowedFunctionNames
		if len(fcc.AllowedFunctionNames) > 0 {
			contract.ToolChoice.Allowed = fcc.AllowedFunctionNames
		}
	}

	// RetrievalConfig 放入 VendorExtras
	if tc.RetrievalConfig != nil {
		if contract.VendorExtras == nil {
			contract.VendorExtras = make(map[string]interface{})
		}
		contract.VendorExtras["retrievalConfig"] = tc.RetrievalConfig
	}
}
