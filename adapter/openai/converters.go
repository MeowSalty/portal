package openai

import (
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
			Role:       msg.Role,
			Name:       &msg.Name,
			ToolCallID: &msg.ToolCallID,
		}

		// 转换 Content
		if contentStr, ok := msg.Content.(string); ok {
			result.Messages[i].Content = coreTypes.MessageContent{
				StringValue: &contentStr,
			}
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
		// 这里需要根据实际的 StopField 类型进行转换
	}

	// 转换 LogitBias
	for k, v := range req.LogitBias {
		// 注意：这里需要处理 string 到 int 的键转换
		// 由于 map[string]int 和 map[int]float64 之间的键类型不同，需要进行转换
		// 这里简化处理，实际应用中可能需要更复杂的转换逻辑
		_ = k
		_ = v
	}

	// 转换 Tools
	for i, tool := range req.Tools {
		result.Tools[i] = coreTypes.Tool{
			Type: tool.Type,
		}
		// 注意：由于 coreTypes.Tool 和 openaiTypes.Tool 结构不同，需要进一步处理 Function 字段
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
			requestMsg.Content = msg.Content.ContentParts
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
	if req.Stop.StringValue != nil || req.Stop.StringArray != nil {
		// 这里需要根据实际的 StopField 类型进行转换
	}

	// 转换 LogitBias
	for k, v := range req.LogitBias {
		// 注意：这里需要处理 int 到 string 的键转换
		// 由于 map[int]float64 和 map[string]int 之间的键类型不同，需要进行转换
		// 这里简化处理，实际应用中可能需要更复杂的转换逻辑
		_ = k
		_ = v
	}

	// 转换 Tools
	for i, tool := range req.Tools {
		result.Tools[i] = openaiTypes.Tool{
			Type: tool.Type,
		}
		// 注意：由于 coreTypes.Tool 和 openaiTypes.Tool 结构不同，需要进一步处理 Function 字段
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
			content := respChoice.Delta.Content
			role := respChoice.Delta.Role
			choice.Delta = &coreTypes.Delta{
				Content: &content,
				Role:    &role,
			}
		} else {
			// 非流式响应，转换 Message
			choice.Message = &coreTypes.ResponseMessage{
				Role:    respChoice.Message.Role,
				Content: respChoice.Message.Content,
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
		ID:          resp.ID,
		Model:       resp.Model,
		Object:      resp.Object,
		Created:     int(resp.Created),
		ServiceTier: resp.SystemFingerprint, // 这里可能需要调整
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

	// 转换 Choices（只处理第一个 Choice）
	if len(resp.Choices) > 0 {
		choice := resp.Choices[0]
		result.Choices = []openaiTypes.Choices{
			{
				FinishReason: choice.FinishReason,
			},
		}

		// 根据 Object 类型判断是流式还是非流式
		if resp.Object == "chat.completion.chunk" {
			// 流式响应，转换 Delta
			result.Choices[0].Delta = &openaiTypes.Delta{
				Content: func() string {
					if choice.Delta != nil && choice.Delta.Content != nil {
						return *choice.Delta.Content
					}
					return ""
				}(),
				Role: func() string {
					if choice.Delta != nil && choice.Delta.Role != nil {
						return *choice.Delta.Role
					}
					return ""
				}(),
			}
		} else {
			// 非流式响应，转换 Message
			result.Choices[0].Message = &openaiTypes.ResponseMessage{
				Role: func() string {
					if choice.Message != nil {
						return choice.Message.Role
					}
					return ""
				}(),
				Content: func() *string {
					if choice.Message != nil {
						return choice.Message.Content
					}
					return nil
				}(),
			}
		}
	}

	return result
}
