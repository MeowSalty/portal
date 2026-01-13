package chat

import (
	"encoding/json"
	"strconv"

	openaiChat "github.com/MeowSalty/portal/request/adapter/openai/types/chat"
	openaiShared "github.com/MeowSalty/portal/request/adapter/openai/types/shared"
	"github.com/MeowSalty/portal/routing"
	coreTypes "github.com/MeowSalty/portal/types"
)

// ConvertRequest 将核心请求转换为 OpenAI 请求
func ConvertRequest(request *coreTypes.Request, channel *routing.Channel) interface{} {
	openaiReq := &openaiChat.Request{
		Model:    channel.ModelName,
		Messages: make([]openaiChat.RequestMessage, len(request.Messages)),
	}

	// 处理流参数
	if request.Stream != nil {
		openaiReq.Stream = request.Stream
		// 当开启流式传输时，默认开启 include_usage
		if *request.Stream {
			trueValue := true
			openaiReq.StreamOptions = &openaiChat.StreamOptions{
				IncludeUsage: &trueValue,
			}
		}
	}

	// 处理温度参数
	if request.Temperature != nil {
		openaiReq.Temperature = request.Temperature
	}

	// 处理 TopP 参数
	if request.TopP != nil {
		openaiReq.TopP = request.TopP
	}

	// 处理最大 token 数
	if request.MaxTokens != nil {
		openaiReq.MaxTokens = request.MaxTokens
	}

	// 处理停止序列
	if request.Stop.StringValue != nil {
		openaiReq.Stop = &openaiChat.StopUnion{
			StringValue: request.Stop.StringValue,
		}
	} else if len(request.Stop.StringArray) > 0 {
		openaiReq.Stop = &openaiChat.StopUnion{
			StringArray: request.Stop.StringArray,
		}
	}

	// 转换消息
	for i, msg := range request.Messages {
		openaiMsg := openaiChat.RequestMessage{
			Role: msg.Role,
			Name: msg.Name,
		}

		// 转换消息内容
		if msg.Content.StringValue != nil {
			// 字符串内容
			openaiMsg.Content = *msg.Content.StringValue
		} else if len(msg.Content.ContentParts) > 0 {
			// 内容部分数组 - OpenAI 使用字符串数组或结构化内容
			contentParts := make([]interface{}, 0, len(msg.Content.ContentParts))
			for _, part := range msg.Content.ContentParts {
				if part.Text != nil {
					contentParts = append(contentParts, map[string]interface{}{
						"type": "text",
						"text": *part.Text,
					})
					continue
				}
				if part.ImageURL != nil {
					imageContent := map[string]interface{}{
						"type": "image_url",
						"image_url": map[string]interface{}{
							"url": part.ImageURL.URL,
						},
					}
					if part.ImageURL.Detail != nil {
						imageContent["image_url"].(map[string]interface{})["detail"] = *part.ImageURL.Detail
					}
					contentParts = append(contentParts, imageContent)
					continue
				}
				if len(part.ExtraFields) > 0 {
					contentParts = append(contentParts, part.ExtraFields)
				}
			}
			openaiMsg.Content = contentParts
		}

		// 转换消息额外字段
		if msg.ExtraFieldsFormat != nil && *msg.ExtraFieldsFormat == "openai" && len(msg.ExtraFields) > 0 {
			openaiMsg.ExtraFields = make(map[string]interface{})
			for key, value := range msg.ExtraFields {
				openaiMsg.ExtraFields[key] = value
			}
		}

		openaiReq.Messages[i] = openaiMsg
	}

	// 转换工具（如果存在）
	if len(request.Tools) > 0 {
		openaiReq.Tools = make([]openaiShared.ToolUnion, len(request.Tools))
		for i, tool := range request.Tools {
			// 将参数转换为 JSON schema
			var parameters json.RawMessage
			if tool.Function.Parameters != nil {
				if paramsBytes, err := json.Marshal(tool.Function.Parameters); err == nil {
					parameters = json.RawMessage(paramsBytes)
				}
			}

			functionDef := openaiShared.FunctionDefinition{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				Parameters:  parameters,
			}

			openaiReq.Tools[i] = openaiShared.ToolUnion{
				Function: &openaiShared.ToolFunction{
					Type:     "function",
					Function: functionDef,
				},
			}
		}
	}

	// 转换工具选择
	if request.ToolChoice.Mode != nil {
		switch *request.ToolChoice.Mode {
		case "none":
			openaiReq.ToolChoice = &openaiShared.ToolChoiceUnion{
				Auto: func() *string { s := "none"; return &s }(),
			}
		case "auto":
			openaiReq.ToolChoice = &openaiShared.ToolChoiceUnion{
				Auto: func() *string { s := "auto"; return &s }(),
			}
		}
	} else if request.ToolChoice.Function != nil {
		openaiReq.ToolChoice = &openaiShared.ToolChoiceUnion{
			Named: &openaiShared.ToolChoiceNamed{
				Type: "function",
				Function: struct {
					Name string `json:"name"`
				}{
					Name: request.ToolChoice.Function.Function.Name,
				},
			},
		}
	}

	// 转换额外字段
	if request.ExtraFieldsFormat != nil && *request.ExtraFieldsFormat == "openai" && len(request.ExtraFields) > 0 {
		openaiReq.ExtraFields = make(map[string]interface{})
		for key, value := range request.ExtraFields {
			openaiReq.ExtraFields[key] = value
		}
	}

	return openaiReq
}

// ConvertCoreRequest 将 OpenAI 请求转换为核心请求
func ConvertCoreRequest(openaiReq *openaiChat.Request) *coreTypes.Request {
	coreReq := &coreTypes.Request{
		Model: openaiReq.Model,
	}

	// 转换流参数
	if openaiReq.Stream != nil {
		coreReq.Stream = openaiReq.Stream
	}

	// 转换温度参数
	if openaiReq.Temperature != nil {
		coreReq.Temperature = openaiReq.Temperature
	}

	// 转换 TopP 参数
	if openaiReq.TopP != nil {
		coreReq.TopP = openaiReq.TopP
	}

	// 转换最大 token 数
	if openaiReq.MaxTokens != nil {
		coreReq.MaxTokens = openaiReq.MaxTokens
	}

	// 转换停止序列
	if openaiReq.Stop != nil {
		// OpenAI 的 StopUnion 可以是字符串或字符串数组，需要转换为核心的 Stop 类型
		if openaiReq.Stop.StringValue != nil {
			coreReq.Stop.StringValue = openaiReq.Stop.StringValue
		} else if openaiReq.Stop.StringArray != nil {
			coreReq.Stop.StringArray = openaiReq.Stop.StringArray
		}
	}

	// 转换消息
	if len(openaiReq.Messages) > 0 {
		coreReq.Messages = make([]coreTypes.Message, len(openaiReq.Messages))
		for i, msg := range openaiReq.Messages {
			coreMsg := coreTypes.Message{
				Role: msg.Role,
				Name: msg.Name,
			}

			// 转换消息内容
			switch content := msg.Content.(type) {
			case string:
				// 字符串内容
				coreMsg.Content.StringValue = &content
			case []interface{}:
				// 结构化内容部分
				contentParts := make([]coreTypes.ContentPart, 0, len(content))
				for _, part := range content {
					partMap, ok := part.(map[string]interface{})
					if !ok {
						continue
					}
					partType, _ := partMap["type"].(string)
					switch partType {
					case "text":
						if text, ok := partMap["text"].(string); ok {
							contentParts = append(contentParts, coreTypes.ContentPart{
								Type:        "text",
								Text:        &text,
								ExtraFields: partMap,
							})
							continue
						}
					case "image_url":
						if imageURL, ok := partMap["image_url"].(map[string]interface{}); ok {
							if url, ok := imageURL["url"].(string); ok {
								imagePart := coreTypes.ContentPart{
									Type: "image_url",
									ImageURL: &coreTypes.ImageURL{
										URL: url,
									},
									ExtraFields: partMap,
								}
								if detail, ok := imageURL["detail"].(string); ok {
									imagePart.ImageURL.Detail = &detail
								}
								contentParts = append(contentParts, imagePart)
								continue
							}
						}
					case "input_text":
						text, _ := partMap["text"].(string)
						contentParts = append(contentParts, coreTypes.ContentPart{
							Type:        "text",
							Text:        &text,
							ExtraFields: partMap,
						})
						continue
					case "input_image":
						if imageURL, ok := partMap["image_url"].(map[string]interface{}); ok {
							url, _ := imageURL["url"].(string)
							var detail *string
							if detailValue, ok := imageURL["detail"].(string); ok {
								detail = &detailValue
							}
							contentParts = append(contentParts, coreTypes.ContentPart{
								Type: "image_url",
								ImageURL: &coreTypes.ImageURL{
									URL:    url,
									Detail: detail,
								},
								ExtraFields: partMap,
							})
							continue
						}
					}
					contentParts = append(contentParts, coreTypes.ContentPart{
						Type:        partType,
						ExtraFields: partMap,
					})
				}
				coreMsg.Content.ContentParts = contentParts
			}
			// 转换消息额外字段
			if len(msg.ExtraFields) > 0 {
				coreMsg.ExtraFields = make(map[string]interface{})
				for key, value := range msg.ExtraFields {
					coreMsg.ExtraFields[key] = value
				}
				format := "openai"
				coreMsg.ExtraFieldsFormat = &format
			}

			coreReq.Messages[i] = coreMsg
		}
	}

	// 转换工具（如果存在）
	if len(openaiReq.Tools) > 0 {
		coreReq.Tools = make([]coreTypes.Tool, len(openaiReq.Tools))
		for i, tool := range openaiReq.Tools {
			if tool.Function != nil {
				// 转换函数参数为 JSON schema
				var parameters interface{}
				if tool.Function.Function.Parameters != nil {
					var params map[string]interface{}
					if paramBytes, ok := tool.Function.Function.Parameters.([]byte); ok {
						json.Unmarshal(paramBytes, &params)
						parameters = params
					} else if paramMap, ok := tool.Function.Function.Parameters.(map[string]interface{}); ok {
						parameters = paramMap
					}
				}

				coreReq.Tools[i] = coreTypes.Tool{
					Type: "function",
					Function: coreTypes.FunctionDescription{
						Name:        tool.Function.Function.Name,
						Description: tool.Function.Function.Description,
						Parameters:  parameters,
					},
				}
			}
			// 注意：这里只处理了 function 类型的工具，custom 类型的工具需要额外处理
		}
	}

	// 转换工具选择
	if openaiReq.ToolChoice != nil {
		if openaiReq.ToolChoice.Auto != nil {
			// "auto" 或 "none" 模式
			mode := *openaiReq.ToolChoice.Auto
			coreReq.ToolChoice.Mode = &mode
		} else if openaiReq.ToolChoice.Named != nil {
			// 命名函数选择
			coreReq.ToolChoice.Function = &coreTypes.FunctionToolChoice{
				Type: "function",
				Function: coreTypes.FunctionNameChoice{
					Name: openaiReq.ToolChoice.Named.Function.Name,
				},
			}
		}
		// 注意：这里只处理了基本的工具选择类型，其他类型需要额外处理
	}

	// 转换其他可选参数
	if openaiReq.FrequencyPenalty != nil {
		coreReq.FrequencyPenalty = openaiReq.FrequencyPenalty
	}
	if openaiReq.PresencePenalty != nil {
		coreReq.PresencePenalty = openaiReq.PresencePenalty
	}
	if openaiReq.Seed != nil {
		seed := int(*openaiReq.Seed)
		coreReq.Seed = &seed
	}
	if openaiReq.User != nil {
		coreReq.User = openaiReq.User
	}
	if openaiReq.LogitBias != nil {
		// 转换 logit_bias 格式（OpenAI 使用 string→int64，核心使用 int→float64）
		coreReq.LogitBias = make(map[int]float64)
		for k, v := range openaiReq.LogitBias {
			// 这里需要将字符串键转换为整数键
			// 注意：实际实现可能需要更复杂的键转换逻辑
			var key int
			if n, err := strconv.Atoi(k); err == nil {
				key = n
			}
			coreReq.LogitBias[key] = float64(v)
		}
	}

	// 转换额外字段
	if len(openaiReq.ExtraFields) > 0 {
		coreReq.ExtraFields = make(map[string]interface{})
		for key, value := range openaiReq.ExtraFields {
			coreReq.ExtraFields[key] = value
		}
		// 设置来源格式为 "openai"
		format := "openai"
		coreReq.ExtraFieldsFormat = &format
	}

	return coreReq
}
