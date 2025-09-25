package openai

import (
	"encoding/json"
	"fmt"

	openaiTypes "github.com/MeowSalty/portal/adapter/openai/types"
	coreTypes "github.com/MeowSalty/portal/types"
)

// ChatCompletionRequestToRequest 将 OpenAI 的 ChatCompletionRequest 转换为内部 Request
func ChatCompletionRequestToRequest(req *openaiTypes.ChatCompletionRequest) *coreTypes.Request {
	if req == nil {
		return nil
	}

	// 初始化默认值
	stream := req.Stream
	maxCompletionTokens := req.MaxCompletionTokens
	temperature := req.Temperature
	seed := req.Seed
	topP := req.TopP
	topLogprobs := req.TopLogProbs
	frequencyPenalty := req.FrequencyPenalty
	presencePenalty := req.PresencePenalty

	result := &coreTypes.Request{
		Model:            req.Model,
		Messages:         make([]coreTypes.Message, len(req.Messages)),
		Stream:           &stream,
		Temperature:      &temperature,
		MaxTokens:        &maxCompletionTokens,
		Seed:             &seed,
		TopP:             &topP,
		TopLogprobs:      &topLogprobs,
		FrequencyPenalty: &frequencyPenalty,
		PresencePenalty:  &presencePenalty,
		LogitBias:        make(map[int]float64),
		Tools:            make([]coreTypes.Tool, len(req.Tools)),
	}

	// 转换 Messages
	for i, msg := range req.Messages {
		result.Messages[i] = coreTypes.Message{
			Role: msg.Role,
		}

		// 正确处理可选字段指针
		if msg.Name != "" {
			result.Messages[i].Name = &msg.Name
		}

		if msg.ToolCallID != "" {
			result.Messages[i].ToolCallID = &msg.ToolCallID
		}

		// 转换 Content
		if contentStr, ok := msg.Content.(string); ok {
			result.Messages[i].Content = coreTypes.MessageContent{
				StringValue: &contentStr,
			}
		} else if contentParts, ok := msg.Content.([]interface{}); ok {
			// 如果 Content 是内容部分数组，则进行转换
			parts := make([]coreTypes.ContentPart, len(contentParts))
			for j, part := range contentParts {
				if textPart, ok := part.(map[string]interface{}); ok && textPart["type"] == "text" {
					if text, ok := textPart["text"].(string); ok {
						parts[j] = coreTypes.ContentPart{
							Type: "text",
							Text: &text,
						}
					}
				}
				// 可以在此处添加对其他类型（如 image_url）的支持
			}
			result.Messages[i].Content.ContentParts = parts
		}
	}

	// 转换 ResponseFormat
	if req.ResponseFormat != nil && req.ResponseFormat.Type != "" {
		responseFormat := coreTypes.ResponseFormat{
			Type: req.ResponseFormat.Type,
		}
		result.ResponseFormat = &responseFormat
	}

	// 转换 Stop
	if req.Stop != nil {
		if stopValue := req.Stop.Get(); stopValue != nil {
			switch v := stopValue.(type) {
			case string:
				if v != "" {
					result.Stop.StringValue = &v
				}
			case []string:
				if len(v) > 0 {
					result.Stop.StringArray = v
				}
			}
		}
	}

	// 转换 LogitBias
	for k, v := range req.LogitBias {
		// 将 string 类型的键转换为 int 类型
		var intKey int
		fmt.Sscanf(k, "%d", &intKey)
		result.LogitBias[intKey] = float64(v)
	}

	// 转换 Tools
	for i, tool := range req.Tools {
		result.Tools[i] = coreTypes.Tool{
			Type: tool.Type,
		}
		if tool.Function != nil {
			result.Tools[i].Function = coreTypes.FunctionDescription{
				Name:        tool.Function.Name,
				Description: &tool.Function.Description,
				Parameters:  tool.Function.Parameters,
			}
		}
	}

	return result
}

// RequestToChatCompletionRequest 将内部 Request 转换为 OpenAI 的 ChatCompletionRequest
func RequestToChatCompletionRequest(req *coreTypes.Request) *openaiTypes.ChatCompletionRequest {
	if req == nil {
		return nil
	}

	// 过滤掉 content 为空的消息
	validMessages := make([]coreTypes.Message, 0, len(req.Messages))
	for _, msg := range req.Messages {
		// 如果 content 是字符串且不为空，或者 content 是内容部分数组且不为空，则保留该消息
		if (msg.Content.StringValue != nil && *msg.Content.StringValue != "") ||
			(len(msg.Content.ContentParts) > 0) {
			validMessages = append(validMessages, msg)
		}
	}

	result := &openaiTypes.ChatCompletionRequest{
		Model:    req.Model,
		Messages: make([]openaiTypes.RequestMessage, len(validMessages)),
		Stream:   *req.Stream,
		Temperature: func() float64 {
			if req.Temperature != nil {
				return *req.Temperature
			}
			return 0.0
		}(),
		MaxCompletionTokens: func() int {
			if req.MaxTokens != nil {
				return *req.MaxTokens
			}
			return 0
		}(),
		Seed: func() int {
			if req.Seed != nil {
				return *req.Seed
			}
			return 0
		}(),
		TopP: func() float64 {
			if req.TopP != nil {
				return *req.TopP
			}
			return 0.0
		}(),
		TopLogProbs: func() int {
			if req.TopLogprobs != nil {
				return *req.TopLogprobs
			}
			return 0
		}(),
		FrequencyPenalty: func() float64 {
			if req.FrequencyPenalty != nil {
				return *req.FrequencyPenalty
			}
			return 0.0
		}(),
		PresencePenalty: func() float64 {
			if req.PresencePenalty != nil {
				return *req.PresencePenalty
			}
			return 0.0
		}(),
		LogitBias: make(map[string]int),
		Tools:     make([]openaiTypes.Tool, len(req.Tools)),
	}

	// 转换 Messages
	for i, msg := range validMessages {
		requestMsg := openaiTypes.RequestMessage{
			Role: msg.Role,
		}

		if msg.Name != nil {
			requestMsg.Name = *msg.Name
		}

		if msg.ToolCallID != nil {
			requestMsg.ToolCallID = *msg.ToolCallID
		}

		// 转换 Content
		if msg.Content.StringValue != nil {
			requestMsg.Content = *msg.Content.StringValue
		} else if msg.Content.ContentParts != nil {
			// 这里需要将 ContentParts 转换为适当的格式
			// 由于 RequestMessage.Content 是 interface{}类型，我们可以直接赋值
			contentParts := make([]interface{}, len(msg.Content.ContentParts))
			for j, part := range msg.Content.ContentParts {
				if part.Text != nil {
					contentParts[j] = map[string]interface{}{
						"type": "text",
						"text": *part.Text,
					}
				}
				// 可以根据需要添加对其他类型内容部分的支持
			}
			requestMsg.Content = contentParts
		}

		result.Messages[i] = requestMsg
	}

	// 转换 ResponseFormat
	if req.ResponseFormat != nil {
		result.ResponseFormat = &openaiTypes.ResponseFormat{
			Type: req.ResponseFormat.Type,
		}
	}

	// 转换 Stop
	if req.Stop.StringValue != nil && *req.Stop.StringValue != "" {
		result.Stop = &openaiTypes.StopField{}
		result.Stop.Value = *req.Stop.StringValue
	} else if len(req.Stop.StringArray) > 0 {
		result.Stop = &openaiTypes.StopField{}
		result.Stop.Value = req.Stop.StringArray
	}

	// 转换 LogitBias
	for k, v := range req.LogitBias {
		// 注意：这里需要处理 int 到 string 的键转换
		// 由于 map[int]float64 和 map[string]int 之间的键类型不同，需要进行转换
		result.LogitBias[fmt.Sprintf("%d", k)] = int(v)
	}

	// 转换 Tools
	for i, tool := range req.Tools {
		result.Tools[i] = openaiTypes.Tool{
			Type: tool.Type,
		}

		// 转换 Function 字段
		if tool.Function.Name != "" {
			// 将 interface{} 类型的 Parameters 转换为 json.RawMessage
			var parameters json.RawMessage
			if tool.Function.Parameters != nil {
				if paramsBytes, err := json.Marshal(tool.Function.Parameters); err == nil {
					parameters = json.RawMessage(paramsBytes)
				}
			}

			result.Tools[i].Function = &openaiTypes.Function{
				Name: tool.Function.Name,
				Description: func() string {
					if tool.Function.Description != nil {
						return *tool.Function.Description
					}
					return ""
				}(),
				Parameters: parameters,
			}
		}
	}

	return result
}

// ChatCompletionResponseToResponse 将 OpenAI 的 ChatCompletionResponse 转换为内部 Response
func ChatCompletionResponseToResponse(resp *openaiTypes.ChatCompletionResponse) *coreTypes.Response {
	if resp == nil {
		return nil
	}

	result := &coreTypes.Response{
		ID:                resp.ID,
		Model:             resp.Model,
		Object:            resp.Object,
		Created:           int64(resp.Created),
		Choices:           make([]coreTypes.Choice, len(resp.Choices)), // 根据实际的 Choices 数组长度创建
		SystemFingerprint: &resp.SystemFingerprint,
	}

	// 转换 Usage（如果存在）
	if resp.Usage != nil && (resp.Usage.PromptTokens != 0 || resp.Usage.CompletionTokens != 0 || resp.Usage.TotalTokens != 0) {
		result.Usage = &coreTypes.ResponseUsage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		}
	}

	// 转换 Choices
	for i, respChoice := range resp.Choices {
		choice := coreTypes.Choice{
			FinishReason: respChoice.FinishReason,
		}

		// 根据 Object 类型判断是流式还是非流式
		if resp.Object == "chat.completion.chunk" {
			// 流式响应，转换 Delta
			if respChoice.Delta != nil {
				choice.Delta = &coreTypes.Delta{}

				// 只有当 Content 不为空时才设置 Content 指针
				if respChoice.Delta.Content != "" {
					content := respChoice.Delta.Content
					choice.Delta.Content = &content
				}

				// 只有当 Role 不为空时才设置 Role 指针
				if respChoice.Delta.Role != "" {
					role := respChoice.Delta.Role
					choice.Delta.Role = &role
				}
			}
		} else {
			// 非流式响应，转换 Message
			if respChoice.Message != nil {
				choice.Message = &coreTypes.ResponseMessage{
					Role:    respChoice.Message.Role,
					Content: respChoice.Message.Content,
				}

				choice.Message.Role = respChoice.Message.Role
				choice.Message.Content = respChoice.Message.Content
			}
		}

		result.Choices[i] = choice
	}

	return result
}

// ResponseToChatCompletionResponse 将内部 Response 转换为 OpenAI 的 ChatCompletionResponse
func ResponseToChatCompletionResponse(resp *coreTypes.Response) *openaiTypes.ChatCompletionResponse {
	if resp == nil {
		return nil
	}

	result := &openaiTypes.ChatCompletionResponse{
		ID:      resp.ID,
		Model:   resp.Model,
		Object:  resp.Object,
		Created: int(resp.Created),
		SystemFingerprint: func() string {
			if resp.SystemFingerprint != nil {
				return *resp.SystemFingerprint
			}
			return ""
		}(),
	}

	// 转换 Usage（如果存在）
	if resp.Usage != nil {
		result.Usage = &openaiTypes.Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		}
	}

	// 转换 Choices（处理所有 Choices）
	result.Choices = make([]openaiTypes.Choices, len(resp.Choices))
	for i, choice := range resp.Choices {
		openaiChoice := openaiTypes.Choices{
			Index:        i,
			FinishReason: choice.FinishReason,
		}

		// 根据 Object 类型判断是流式还是非流式
		if resp.Object == "chat.completion.chunk" {
			// 流式响应，转换 Delta
			delta := &openaiTypes.Delta{}
			if choice.Delta != nil {
				if choice.Delta.Content != nil {
					delta.Content = *choice.Delta.Content
				}
				if choice.Delta.Role != nil {
					delta.Role = *choice.Delta.Role
				}
			}
			openaiChoice.Delta = delta
		} else {
			// 非流式响应，转换 Message
			message := &openaiTypes.ResponseMessage{}
			if choice.Message != nil {
				message.Role = choice.Message.Role
				message.Content = choice.Message.Content
			}
			openaiChoice.Message = message
		}

		result.Choices[i] = openaiChoice
	}

	return result
}
