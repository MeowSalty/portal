package responses

import (
	"encoding/json"
	"testing"

	responsesTypes "github.com/MeowSalty/portal/request/adapter/openai/types/responses"
	"github.com/MeowSalty/portal/request/adapter/openai/types/shared"
	"github.com/MeowSalty/portal/request/adapter/types"
	"github.com/stretchr/testify/assert"
)

func TestRequestFromContract(t *testing.T) {
	t.Run("当请求为 nil 时应返回 nil", func(t *testing.T) {
		req, err := RequestFromContract(nil)
		assert.NoError(t, err)
		assert.Nil(t, req)
	})

	t.Run("应优先使用 Prompt 生成输入", func(t *testing.T) {
		prompt := "hello"
		messageText := "message"
		contract := &types.RequestContract{
			Prompt: &prompt,
			Messages: []types.Message{
				{
					Role: "user",
					Content: types.Content{
						Text: &messageText,
					},
				},
			},
		}

		req, err := RequestFromContract(contract)
		assert.NoError(t, err)
		if assert.NotNil(t, req) && assert.NotNil(t, req.Input) {
			assert.Equal(t, prompt, *req.Input.StringValue)
			assert.Empty(t, req.Input.Items)
		}
	})

	t.Run("应将消息转换为输入项", func(t *testing.T) {
		text := "你好"
		contract := &types.RequestContract{
			Messages: []types.Message{
				{
					Role:    "user",
					Content: types.Content{Text: &text},
				},
			},
		}

		req, err := RequestFromContract(contract)
		assert.NoError(t, err)
		if assert.NotNil(t, req) && assert.NotNil(t, req.Input) {
			assert.Nil(t, req.Input.StringValue)
			if assert.Len(t, req.Input.Items, 1) {
				assert.NotNil(t, req.Input.Items[0].Message)
				assert.Equal(t, responsesTypes.ResponseMessageRoleUser, req.Input.Items[0].Message.Role)
				// 断言 Content 列表内容
				if req.Input.Items[0].Message.Content.List != nil {
					if assert.Len(t, *req.Input.Items[0].Message.Content.List, 1) {
						assert.NotNil(t, (*req.Input.Items[0].Message.Content.List)[0].Text)
						assert.Equal(t, responsesTypes.InputContentTypeText, (*req.Input.Items[0].Message.Content.List)[0].Text.Type)
						assert.Equal(t, text, (*req.Input.Items[0].Message.Content.List)[0].Text.Text)
					}
				}
			}
		}
	})

	t.Run("应转换系统指令与采样参数", func(t *testing.T) {
		systemText := "系统"
		maxTokens := 120
		temperature := 0.4
		topP := 0.8
		topLogprobs := 2
		contract := &types.RequestContract{
			System: &types.System{
				Text: &systemText,
			},
			MaxOutputTokens: &maxTokens,
			Temperature:     &temperature,
			TopP:            &topP,
			TopLogprobs:     &topLogprobs,
		}

		req, err := RequestFromContract(contract)
		assert.NoError(t, err)
		if assert.NotNil(t, req) {
			assert.Equal(t, systemText, *req.Instructions)
			assert.Equal(t, maxTokens, *req.MaxOutputTokens)
			assert.Equal(t, temperature, *req.Temperature)
			assert.Equal(t, topP, *req.TopP)
			assert.Equal(t, topLogprobs, *req.TopLogprobs)
		}
	})

	t.Run("应转换流式配置与元数据", func(t *testing.T) {
		stream := true
		includeObfuscation := true
		contract := &types.RequestContract{
			Stream: &stream,
			StreamOptions: &types.StreamOption{
				IncludeObfuscation: &includeObfuscation,
			},
			Metadata: map[string]interface{}{
				"ok":   "yes",
				"drop": 1,
			},
		}

		req, err := RequestFromContract(contract)
		assert.NoError(t, err)
		if assert.NotNil(t, req) {
			assert.Equal(t, true, *req.Stream)
			assert.NotNil(t, req.StreamOptions)
			assert.Equal(t, true, *req.StreamOptions.IncludeObfuscation)
			if assert.NotNil(t, req.Metadata) {
				assert.Equal(t, "yes", req.Metadata["ok"])
				_, exists := req.Metadata["drop"]
				assert.False(t, exists)
			}
		}
	})

	t.Run("应转换服务层级与缓存字段", func(t *testing.T) {
		user := "user"
		serviceTier := "auto"
		cacheKey := "cache"
		retention := "24h"
		store := true
		contract := &types.RequestContract{
			User:                 &user,
			ServiceTier:          &serviceTier,
			PromptCacheKey:       &cacheKey,
			PromptCacheRetention: &retention,
			Store:                &store,
		}

		req, err := RequestFromContract(contract)
		assert.NoError(t, err)
		if assert.NotNil(t, req) {
			assert.Equal(t, user, *req.User)
			assert.Equal(t, serviceTier, string(*req.ServiceTier))
			assert.Equal(t, cacheKey, *req.PromptCacheKey)
			assert.Equal(t, responsesTypes.PromptCacheRetention(retention), *req.PromptCacheRetention)
			assert.Equal(t, store, *req.Store)
		}
	})

	t.Run("应转换 Reasoning 与 ResponseFormat", func(t *testing.T) {
		effort := "medium"
		summary := "detailed"
		generateSummary := "concise"
		description := "结构化输出"
		strict := true
		contract := &types.RequestContract{
			Reasoning: &types.Reasoning{
				Effort:          &effort,
				Summary:         &summary,
				GenerateSummary: &generateSummary,
			},
			ResponseFormat: &types.ResponseFormat{
				Type: string(responsesTypes.TextResponseFormatTypeJSONSchema),
				JSONSchema: map[string]interface{}{
					"name":        "schema",
					"description": description,
					"schema": map[string]interface{}{
						"type": "object",
					},
					"strict": strict,
				},
			},
		}

		req, err := RequestFromContract(contract)
		assert.NoError(t, err)
		if assert.NotNil(t, req) {
			if assert.NotNil(t, req.Reasoning) {
				assert.Equal(t, effort, string(*req.Reasoning.Effort))
				assert.Equal(t, summary, string(*req.Reasoning.Summary))
				assert.Equal(t, generateSummary, string(*req.Reasoning.GenerateSummary))
			}
			if assert.NotNil(t, req.Text) && assert.NotNil(t, req.Text.Format) && assert.NotNil(t, req.Text.Format.JSONSchema) {
				assert.Equal(t, "schema", req.Text.Format.JSONSchema.Name)
				assert.Equal(t, description, *req.Text.Format.JSONSchema.Description)
				assert.Equal(t, strict, *req.Text.Format.JSONSchema.Strict)
			}
		}
	})

	t.Run("应转换工具、工具选择与并行调用", func(t *testing.T) {
		strict := true
		mode := "required"
		parallel := true
		contract := &types.RequestContract{
			Tools: []types.Tool{
				{
					Type: "function",
					Function: &types.Function{
						Name:       "search",
						Parameters: map[string]interface{}{"type": "object"},
					},
					VendorExtras: map[string]interface{}{
						"strict": &strict,
					},
				},
			},
			ToolChoice:        &types.ToolChoice{Mode: &mode},
			ParallelToolCalls: &parallel,
		}

		req, err := RequestFromContract(contract)
		assert.NoError(t, err)
		if assert.NotNil(t, req) {
			if assert.Len(t, req.Tools, 1) {
				assert.NotNil(t, req.Tools[0].Function)
				assert.Equal(t, "search", *req.Tools[0].Function.Name)
				assert.Equal(t, strict, *req.Tools[0].Function.Strict)
			}
			if assert.NotNil(t, req.ToolChoice) && assert.NotNil(t, req.ToolChoice.Auto) {
				assert.Equal(t, mode, *req.ToolChoice.Auto)
			}
			assert.Equal(t, parallel, *req.ParallelToolCalls)
		}
	})

	t.Run("应恢复供应商附加字段并保留扩展字段", func(t *testing.T) {
		include := []string{"file_search_call.results"}
		truncation := "auto"
		conversationID := "conv_1"
		promptID := "prompt_1"
		previousResponseID := "resp_1"
		background := true
		maxToolCalls := 2
		safetyID := "safe-1"
		verbosity := "high"
		contract := &types.RequestContract{
			VendorExtras: map[string]interface{}{
				"include":              include,
				"truncation":           truncation,
				"conversation":         &responsesTypes.ConversationUnion{StringValue: &conversationID},
				"prompt":               &responsesTypes.PromptTemplate{ID: promptID},
				"previous_response_id": &previousResponseID,
				"background":           &background,
				"max_tool_calls":       &maxToolCalls,
				"safety_identifier":    &safetyID,
				"verbosity":            verbosity,
				"extra":                "value",
			},
		}

		req, err := RequestFromContract(contract)
		assert.NoError(t, err)
		if assert.NotNil(t, req) {
			assert.Equal(t, include, []string(req.Include))
			assert.Equal(t, responsesTypes.TruncationStrategy(truncation), *req.Truncation)
			assert.NotNil(t, req.Conversation)
			assert.Equal(t, conversationID, *req.Conversation.StringValue)
			assert.NotNil(t, req.Prompt)
			assert.Equal(t, promptID, req.Prompt.ID)
			assert.Equal(t, previousResponseID, *req.PreviousResponseID)
			assert.Equal(t, background, *req.Background)
			assert.Equal(t, maxToolCalls, *req.MaxToolCalls)
			assert.Equal(t, safetyID, *req.SafetyIdentifier)
			assert.NotNil(t, req.Text)
			assert.Equal(t, shared.VerbosityLevel(verbosity), *req.Text.Verbosity)
			assert.Equal(t, json.RawMessage(`"value"`), req.ExtraFields["extra"])
		}
	})
}
