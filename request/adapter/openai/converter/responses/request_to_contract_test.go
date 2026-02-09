package responses

import (
	"encoding/json"
	"testing"

	responsesTypes "github.com/MeowSalty/portal/request/adapter/openai/types/responses"
	"github.com/MeowSalty/portal/request/adapter/openai/types/shared"
	"github.com/MeowSalty/portal/request/adapter/types"
	"github.com/stretchr/testify/assert"
)

func TestRequestToContract(t *testing.T) {
	t.Run("当请求为 nil 时应返回 nil", func(t *testing.T) {
		contract, err := RequestToContract(nil)
		assert.NoError(t, err)
		assert.Nil(t, contract)
	})

	t.Run("应正确转换基本字段", func(t *testing.T) {
		model := "gpt-4"
		maxTokens := 100
		temp := 0.7
		topP := 0.9
		stream := true
		user := "test-user"

		req := &responsesTypes.Request{
			Model:           &model,
			MaxOutputTokens: &maxTokens,
			Temperature:     &temp,
			TopP:            &topP,
			Stream:          &stream,
			User:            &user,
		}

		contract, err := RequestToContract(req)
		assert.NoError(t, err)
		assert.NotNil(t, contract)
		assert.Equal(t, types.VendorSourceOpenAIResponse, contract.Source)
		assert.Equal(t, model, contract.Model)
		assert.Equal(t, maxTokens, *contract.MaxOutputTokens)
		assert.Equal(t, temp, *contract.Temperature)
		assert.Equal(t, topP, *contract.TopP)
		assert.Equal(t, stream, *contract.Stream)
		assert.Equal(t, user, *contract.User)
	})

	t.Run("应将输入字符串转换为提示", func(t *testing.T) {
		inputStr := "Hello, world!"
		req := &responsesTypes.Request{
			Input: &responsesTypes.InputUnion{
				StringValue: &inputStr,
			},
		}

		contract, err := RequestToContract(req)
		assert.NoError(t, err)
		assert.Equal(t, inputStr, *contract.Prompt)
		assert.Empty(t, contract.Messages)
	})

	t.Run("应将输入项转换为消息", func(t *testing.T) {
		content := "你好"
		parts := []responsesTypes.InputContent{
			{
				Text: &responsesTypes.InputTextContent{
					Type: responsesTypes.InputContentTypeText,
					Text: content,
				},
			},
		}
		items := []responsesTypes.InputItem{
			{
				Message: &responsesTypes.InputMessage{
					Type:    responsesTypes.InputItemTypeMessage,
					Role:    responsesTypes.ResponseMessageRoleUser,
					Content: responsesTypes.NewInputMessageContentFromList(parts),
				},
			},
		}
		req := &responsesTypes.Request{
			Input: &responsesTypes.InputUnion{
				Items: items,
			},
		}

		contract, err := RequestToContract(req)
		assert.NoError(t, err)
		assert.Empty(t, contract.Prompt)
		if assert.Len(t, contract.Messages, 1) {
			assert.Equal(t, "user", contract.Messages[0].Role)
			if assert.NotNil(t, contract.Messages[0].Content.Text) {
				assert.Equal(t, content, *contract.Messages[0].Content.Text)
			}
		}
	})

	t.Run("应转换多种输入项类型", func(t *testing.T) {
		text := "hello"
		imageURL := "https://example.com/image.png"
		fileID := "file_123"
		fileData := "ZmlsZQ=="
		filename := "doc.txt"
		fileURL := "https://example.com/doc.txt"
		detail := shared.ImageDetailHigh

		parts := []responsesTypes.InputContent{
			{
				Text: &responsesTypes.InputTextContent{
					Type: responsesTypes.InputContentTypeText,
					Text: text,
				},
			},
			{
				Image: &responsesTypes.InputImageContent{
					Type:     responsesTypes.InputContentTypeImage,
					ImageURL: &imageURL,
					FileID:   &fileID,
					Detail:   &detail,
				},
			},
			{
				File: &responsesTypes.InputFileContent{
					Type:     responsesTypes.InputContentTypeFile,
					FileID:   &fileID,
					FileData: &fileData,
					Filename: &filename,
					FileURL:  &fileURL,
				},
			},
		}
		items := []responsesTypes.InputItem{
			{
				Message: &responsesTypes.InputMessage{
					Type:    responsesTypes.InputItemTypeMessage,
					Role:    responsesTypes.ResponseMessageRoleUser,
					Content: responsesTypes.NewInputMessageContentFromList(parts),
				},
			},
			{
				ItemReference: &responsesTypes.ItemReferenceParam{ID: "item_1"},
			},
			{
				FunctionCall: &responsesTypes.FunctionToolCall{
					CallID:    "call_1",
					Name:      "lookup",
					Arguments: "{\"q\":\"go\"}",
					Status:    "completed",
				},
			},
			{
				FunctionCallOutput: &responsesTypes.InputFunctionToolCallOutput{
					CallID: "call_1",
					Output: json.RawMessage(`"ok"`),
				},
			},
		}

		req := &responsesTypes.Request{
			Input: &responsesTypes.InputUnion{Items: items},
		}

		contract, err := RequestToContract(req)
		assert.NoError(t, err)
		if assert.Len(t, contract.Messages, 3) {
			assert.Equal(t, "user", contract.Messages[0].Role)
			if assert.Len(t, contract.Messages[0].Content.Parts, 3) {
				assert.Equal(t, "input_text", contract.Messages[0].Content.Parts[0].Type)
				assert.Equal(t, text, *contract.Messages[0].Content.Parts[0].Text)
				assert.Equal(t, "input_image", contract.Messages[0].Content.Parts[1].Type)
				if assert.NotNil(t, contract.Messages[0].Content.Parts[1].Image) {
					assert.Equal(t, imageURL, *contract.Messages[0].Content.Parts[1].Image.URL)
					assert.Equal(t, "high", *contract.Messages[0].Content.Parts[1].Image.Detail)
				}
				if assert.NotNil(t, contract.Messages[0].Content.Parts[1].VendorExtras) {
					assert.Equal(t, fileID, contract.Messages[0].Content.Parts[1].VendorExtras["file_id"])
				}
				assert.Equal(t, "input_file", contract.Messages[0].Content.Parts[2].Type)
				if assert.NotNil(t, contract.Messages[0].Content.Parts[2].File) {
					assert.Equal(t, fileID, *contract.Messages[0].Content.Parts[2].File.ID)
					assert.Equal(t, fileData, *contract.Messages[0].Content.Parts[2].File.Data)
					assert.Equal(t, filename, *contract.Messages[0].Content.Parts[2].File.Filename)
					assert.Equal(t, fileURL, *contract.Messages[0].Content.Parts[2].File.URL)
				}
			}

			assert.Equal(t, "assistant", contract.Messages[1].Role)
			if assert.Len(t, contract.Messages[1].Content.Parts, 1) {
				part := contract.Messages[1].Content.Parts[0]
				assert.Equal(t, "tool_call", part.Type)
				if assert.NotNil(t, part.ToolCall) {
					assert.Equal(t, "function", *part.ToolCall.Type)
					assert.Equal(t, "call_1", *part.ToolCall.ID)
					assert.Equal(t, "lookup", *part.ToolCall.Name)
					assert.Equal(t, "{\"q\":\"go\"}", *part.ToolCall.Arguments)
					if assert.NotNil(t, part.ToolCall.Payload) {
						assert.Equal(t, "completed", part.ToolCall.Payload["status"])
					}
				}
			}

			assert.Equal(t, "tool", contract.Messages[2].Role)
			if assert.Len(t, contract.Messages[2].Content.Parts, 1) {
				part := contract.Messages[2].Content.Parts[0]
				assert.Equal(t, "tool_result", part.Type)
				if assert.NotNil(t, part.ToolResult) {
					assert.Equal(t, "call_1", *part.ToolResult.ID)
					assert.Equal(t, "ok", *part.ToolResult.Content)
					assert.Nil(t, part.ToolResult.Payload)
				}
			}
		}
	})

	t.Run("应将指令转换为系统消息", func(t *testing.T) {
		instructions := "Be helpful and concise."
		req := &responsesTypes.Request{
			Instructions: &instructions,
		}

		contract, err := RequestToContract(req)
		assert.NoError(t, err)
		if assert.NotNil(t, contract.System) {
			assert.NotNil(t, contract.System.Text)
			assert.Equal(t, instructions, *contract.System.Text)
		}
	})

	t.Run("应转换流选项", func(t *testing.T) {
		includeObfuscation := true
		req := &responsesTypes.Request{
			StreamOptions: &responsesTypes.StreamOptions{
				IncludeObfuscation: &includeObfuscation,
			},
		}

		contract, err := RequestToContract(req)
		assert.NoError(t, err)
		assert.True(t, *contract.StreamOptions.IncludeObfuscation)
	})

	t.Run("应转换响应格式与文本详细度", func(t *testing.T) {
		description := "结构化输出"
		strict := true
		format := &responsesTypes.TextFormatUnion{
			JSONSchema: &responsesTypes.TextFormatJSONSchema{
				Type:        responsesTypes.TextResponseFormatTypeJSONSchema,
				Name:        "schema_name",
				Description: &description,
				Schema: map[string]interface{}{
					"type": "object",
				},
				Strict: &strict,
			},
		}
		verbosity := shared.VerbosityHigh
		textConfig := &responsesTypes.TextConfig{
			Format:    format,
			Verbosity: &verbosity,
		}

		req := &responsesTypes.Request{
			Text: textConfig,
		}

		contract, err := RequestToContract(req)
		assert.NoError(t, err)
		if assert.NotNil(t, contract.ResponseFormat) {
			assert.Equal(t, "json_schema", contract.ResponseFormat.Type)
			if assert.NotNil(t, contract.ResponseFormat.JSONSchema) {
				schema, ok := contract.ResponseFormat.JSONSchema.(map[string]interface{})
				if assert.True(t, ok) {
					assert.Equal(t, "schema_name", schema["name"])
					assert.Equal(t, "结构化输出", schema["description"])
					if schemaObj, ok := schema["schema"].(map[string]interface{}); assert.True(t, ok) {
						assert.Equal(t, "object", schemaObj["type"])
					}
					assert.Equal(t, true, schema["strict"])
				}
			}
		}
		assert.Equal(t, "high", contract.VendorExtras["verbosity"])
	})

	t.Run("应转换工具、工具选择与并行调用", func(t *testing.T) {
		strict := true
		functionName := "search"
		req := &responsesTypes.Request{
			Tools: []shared.ToolUnion{
				{
					Function: &shared.ToolFunction{
						Type: "function",
						Name: &functionName,
						Parameters: map[string]interface{}{
							"type": "object",
						},
						Strict: &strict,
					},
				},
			},
			ToolChoice: &shared.ToolChoiceUnion{
				Auto: func() *string {
					mode := "required"
					return &mode
				}(),
			},
			ParallelToolCalls: func() *bool {
				value := true
				return &value
			}(),
		}

		contract, err := RequestToContract(req)
		assert.NoError(t, err)
		if assert.Len(t, contract.Tools, 1) {
			assert.Equal(t, "function", contract.Tools[0].Type)
			if assert.NotNil(t, contract.Tools[0].Function) {
				assert.Equal(t, "search", contract.Tools[0].Function.Name)
			}
			if assert.NotNil(t, contract.Tools[0].VendorExtras) {
				assert.Equal(t, true, contract.Tools[0].VendorExtras["strict"])
			}
		}
		if assert.NotNil(t, contract.ToolChoice) {
			assert.Equal(t, "required", *contract.ToolChoice.Mode)
		}
		assert.Equal(t, true, *contract.ParallelToolCalls)
	})

	t.Run("应转换缓存与存储字段", func(t *testing.T) {
		promptCacheKey := "cache-key"
		retention := responsesTypes.PromptCacheRetention24h
		store := true
		req := &responsesTypes.Request{
			PromptCacheKey:       &promptCacheKey,
			PromptCacheRetention: &retention,
			Store:                &store,
		}

		contract, err := RequestToContract(req)
		assert.NoError(t, err)
		assert.Equal(t, "cache-key", *contract.PromptCacheKey)
		assert.Equal(t, "24h", *contract.PromptCacheRetention)
		assert.Equal(t, true, *contract.Store)
	})

	t.Run("应转换元数据", func(t *testing.T) {
		metadata := map[string]string{"key1": "value1", "key2": "value2"}
		req := &responsesTypes.Request{
			Metadata: metadata,
		}

		contract, err := RequestToContract(req)
		assert.NoError(t, err)
		assert.Len(t, contract.Metadata, 2)
		assert.Equal(t, "value1", contract.Metadata["key1"])
		assert.Equal(t, "value2", contract.Metadata["key2"])
	})

	t.Run("应转换服务层级", func(t *testing.T) {
		serviceTier := "auto"
		req := &responsesTypes.Request{
			ServiceTier: &serviceTier,
		}

		contract, err := RequestToContract(req)
		assert.NoError(t, err)
		assert.Equal(t, "auto", *contract.ServiceTier)
	})

	t.Run("应处理供应商附加字段", func(t *testing.T) {
		include := responsesTypes.IncludeList{responsesTypes.IncludeFileSearchResults}
		truncation := responsesTypes.TruncationStrategyAuto
		conversationID := "conv_123456"
		background := true
		previousResponseID := "resp_abc"
		maxToolCalls := 2
		safetyIdentifier := "safe-1"
		promptTemplate := &responsesTypes.PromptTemplate{
			ID: "prompt-123",
		}

		req := &responsesTypes.Request{
			Include:            include,
			Truncation:         &truncation,
			Conversation:       &responsesTypes.ConversationUnion{StringValue: &conversationID},
			Background:         &background,
			Prompt:             promptTemplate,
			PreviousResponseID: &previousResponseID,
			MaxToolCalls:       &maxToolCalls,
			SafetyIdentifier:   &safetyIdentifier,
			ExtraFields:        map[string]json.RawMessage{"extra_field": json.RawMessage(`"extra_value"`)},
		}

		contract, err := RequestToContract(req)
		assert.NoError(t, err)
		assert.NotNil(t, contract.VendorExtras)
		assert.Equal(t, types.VendorSourceOpenAIResponse, *contract.VendorExtrasSource)
		assert.Equal(t, include, contract.VendorExtras["include"])
		assert.Equal(t, "auto", contract.VendorExtras["truncation"])
		if assert.NotNil(t, contract.VendorExtras["conversation"]) {
			conversation, ok := contract.VendorExtras["conversation"].(*responsesTypes.ConversationUnion)
			if assert.True(t, ok) {
				assert.NotNil(t, conversation.StringValue)
				assert.Equal(t, conversationID, *conversation.StringValue)
			}
		}
		if assert.NotNil(t, contract.VendorExtras["background"]) {
			backgroundValue, ok := contract.VendorExtras["background"].(*bool)
			if assert.True(t, ok) {
				assert.Equal(t, background, *backgroundValue)
			}
		}
		if assert.NotNil(t, contract.VendorExtras["prompt"]) {
			prompt, ok := contract.VendorExtras["prompt"].(*responsesTypes.PromptTemplate)
			if assert.True(t, ok) {
				assert.Equal(t, "prompt-123", prompt.ID)
			}
		}
		if assert.NotNil(t, contract.VendorExtras["previous_response_id"]) {
			value, ok := contract.VendorExtras["previous_response_id"].(*string)
			if assert.True(t, ok) {
				assert.Equal(t, "resp_abc", *value)
			}
		}
		if assert.NotNil(t, contract.VendorExtras["max_tool_calls"]) {
			value, ok := contract.VendorExtras["max_tool_calls"].(*int)
			if assert.True(t, ok) {
				assert.Equal(t, 2, *value)
			}
		}
		if assert.NotNil(t, contract.VendorExtras["safety_identifier"]) {
			value, ok := contract.VendorExtras["safety_identifier"].(*string)
			if assert.True(t, ok) {
				assert.Equal(t, "safe-1", *value)
			}
		}
		assert.Equal(t, "extra_value", contract.VendorExtras["extra_field"])
	})

	t.Run("应处理推理字段", func(t *testing.T) {
		reasoningEffort := "medium"
		summary := responsesTypes.ReasoningSummaryDetailed
		generateSummary := responsesTypes.ReasoningSummaryConcise

		req := &responsesTypes.Request{
			Reasoning: &responsesTypes.Reasoning{
				Effort:          &reasoningEffort,
				Summary:         &summary,
				GenerateSummary: &generateSummary,
			},
		}

		contract, err := RequestToContract(req)
		assert.NoError(t, err)
		assert.Equal(t, "medium", *contract.Reasoning.Effort)
		assert.Equal(t, "detailed", *contract.Reasoning.Summary)
		assert.Equal(t, "concise", *contract.Reasoning.GenerateSummary)
	})

	t.Run("应转换 TopLogprobs", func(t *testing.T) {
		topLogprobs := 4
		req := &responsesTypes.Request{
			TopLogprobs: &topLogprobs,
		}

		contract, err := RequestToContract(req)
		assert.NoError(t, err)
		assert.Equal(t, 4, *contract.TopLogprobs)
	})
}
