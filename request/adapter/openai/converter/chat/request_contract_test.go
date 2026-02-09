package chat

import (
	"reflect"
	"testing"

	chatTypes "github.com/MeowSalty/portal/request/adapter/openai/types/chat"
	"github.com/MeowSalty/portal/request/adapter/openai/types/shared"
	"github.com/MeowSalty/portal/request/adapter/types"
)

func TestFromContract_PromptFallback(t *testing.T) {
	prompt := "hello"
	contract := &types.RequestContract{
		Model:  "gpt-test",
		Prompt: &prompt,
	}

	req, err := RequestFromContract(contract)
	if err != nil {
		t.Fatalf("FromContract 错误: %v", err)
	}
	if req == nil {
		t.Fatalf("FromContract 返回 nil")
	}
	if len(req.Messages) != 1 {
		t.Fatalf("意外的消息长度：%d", len(req.Messages))
	}
	if req.Messages[0].Role != chatTypes.MessageRoleUser {
		t.Fatalf("意外的角色: %s", req.Messages[0].Role)
	}
	if req.Messages[0].Content.StringValue == nil || *req.Messages[0].Content.StringValue != prompt {
		t.Fatalf("意外的提示内容")
	}
}

func TestFromContract_Complex(t *testing.T) {
	sysText := "system"
	msgText := "hi"
	imgURL := "https://example.com/img.png"
	detail := "high"
	audioData := "audio"
	audioFormat := "mp3"
	fileID := "file-id"
	fileName := "name.txt"
	fileData := "data"
	refusal := "no"
	toolID := "tool-id"
	toolName := "calc"
	toolArgs := "{\"a\":1}"
	temperature := 0.4
	presence := 0.2
	frequency := 0.1
	maxOut := 128
	seed := 7
	candidate := 2
	logprobs := true
	topLogprobs := 3
	stream := true
	includeUsage := true
	serviceTier := "auto"
	promptCacheKey := "cache-key"
	promptCacheRetention := "retention"
	store := true
	effort := "low"
	verbosity := "low"
	safetyID := "safe-id"
	parallel := false
	stopText := "stop"

	system := &types.System{
		Text: &sysText,
		VendorExtras: map[string]interface{}{
			"sys_extra": "v",
		},
	}

	parts := []types.ContentPart{
		{
			Type: "text",
			Text: &msgText,
		},
		{
			Type: "image",
			Image: &types.Image{
				URL:    &imgURL,
				Detail: &detail,
			},
		},
		{
			Type: "audio",
			Audio: &types.Audio{
				Data:   &audioData,
				Format: &audioFormat,
			},
		},
		{
			Type: "file",
			File: &types.File{
				ID:       &fileID,
				Data:     &fileData,
				Filename: &fileName,
			},
		},
		{
			Type: chatTypes.ContentPartTypeRefusal,
			VendorExtras: map[string]interface{}{
				"refusal": &refusal,
			},
		},
		{
			Type: "unknown",
			VendorExtras: map[string]interface{}{
				"original_part": chatTypes.ContentPart{
					Type: chatTypes.ContentPartTypeText,
					Text: strPtr("fallback"),
				},
			},
		},
	}

	functionCall := &chatTypes.RequestFunctionCall{Name: "f", Arguments: "{}"}
	assistantAudio := &chatTypes.AssistantAudio{ID: "audio-id"}

	message := types.Message{
		Role: chatTypes.MessageRoleUser,
		Content: types.Content{
			Parts: parts,
		},
		ToolCalls: []types.ToolCall{
			{
				ID:        &toolID,
				Name:      &toolName,
				Arguments: &toolArgs,
			},
		},
		VendorExtras: map[string]interface{}{
			"function_call": functionCall,
			"refusal":       &refusal,
			"audio":         assistantAudio,
			"extra":         "x",
		},
	}

	prediction := &chatTypes.PredictionContent{
		Type: chatTypes.PredictionTypeContent,
		Content: chatTypes.PredictionContentUnion{
			StringValue: strPtr("predict"),
		},
	}
	webSearch := &chatTypes.WebSearchOptions{
		SearchContextSize: ptrWebSearchContextSize(chatTypes.WebSearchContextSizeSmall),
	}
	requestAudio := &chatTypes.RequestAudio{Format: chatTypes.AudioFormatMP3, Voice: "alloy"}

	strict := true
	toolFuncDesc := "tool desc"
	contract := &types.RequestContract{
		Model:            "gpt-test",
		Messages:         []types.Message{message},
		System:           system,
		MaxOutputTokens:  &maxOut,
		Temperature:      &temperature,
		TopP:             nil,
		PresencePenalty:  &presence,
		FrequencyPenalty: &frequency,
		Seed:             &seed,
		Stop: &types.Stop{
			Text: &stopText,
		},
		CandidateCount: &candidate,
		Logprobs:       &logprobs,
		TopLogprobs:    &topLogprobs,
		Stream:         &stream,
		StreamOptions: &types.StreamOption{
			IncludeUsage:       &includeUsage,
			IncludeObfuscation: nil,
		},
		Metadata: map[string]interface{}{
			"ok":   "yes",
			"drop": 1,
		},
		User:                 strPtr("user1"),
		ServiceTier:          &serviceTier,
		PromptCacheKey:       &promptCacheKey,
		PromptCacheRetention: &promptCacheRetention,
		Store:                &store,
		Reasoning: &types.Reasoning{
			Effort: &effort,
		},
		ResponseFormat: &types.ResponseFormat{
			Type: string(chatTypes.ResponseFormatTypeJSONSchema),
			JSONSchema: map[string]interface{}{
				"name":        "schema",
				"description": "desc",
				"schema": map[string]interface{}{
					"type": "object",
				},
				"strict": true,
			},
		},
		Modalities: []string{"text", "audio"},
		Tools: []types.Tool{
			{
				Type: "function",
				Function: &types.Function{
					Name:        "tool",
					Description: &toolFuncDesc,
					Parameters:  map[string]interface{}{"a": "b"},
				},
				VendorExtras: map[string]interface{}{
					"strict": &strict,
				},
			},
			{
				Type: "custom",
				VendorExtras: map[string]interface{}{
					"tool": &shared.ToolCustom{
						Type:   "custom",
						Name:   strPtr("my_tool"),
						Custom: map[string]interface{}{"key": "value"},
					},
				},
			},
		},
		ToolChoice: &types.ToolChoice{
			Mode:     strPtr("function"),
			Function: strPtr("tool"),
		},
		ParallelToolCalls: &parallel,
		VendorExtras: map[string]interface{}{
			"audio":              requestAudio,
			"prediction":         prediction,
			"web_search_options": webSearch,
			"verbosity":          verbosity,
			"safety_identifier":  &safetyID,
			"logit_bias":         map[string]int{"1": 2},
			"function_call":      &chatTypes.FunctionCallUnion{Mode: ptrToolChoiceMode(chatTypes.ToolChoiceModeAuto)},
			"functions":          []shared.FunctionDefinition{{Name: "fn"}},
			"extra_field":        "extra",
		},
	}

	req, err := RequestFromContract(contract)
	if err != nil {
		t.Fatalf("FromContract 错误: %v", err)
	}
	if req == nil {
		t.Fatalf("FromContract 返回 nil")
	}
	if req.Model != contract.Model {
		t.Fatalf("意外的模型")
	}
	if req.MaxCompletionTokens == nil || *req.MaxCompletionTokens != maxOut {
		t.Fatalf("意外的 MaxCompletionTokens")
	}
	if req.PresencePenalty == nil || *req.PresencePenalty != presence {
		t.Fatalf("意外的 PresencePenalty")
	}
	if req.FrequencyPenalty == nil || *req.FrequencyPenalty != frequency {
		t.Fatalf("意外的 FrequencyPenalty")
	}
	if req.Seed == nil || *req.Seed != seed {
		t.Fatalf("意外的 Seed")
	}
	if req.Stop == nil || req.Stop.StringValue == nil || *req.Stop.StringValue != stopText {
		t.Fatalf("意外的 Stop")
	}
	if req.N == nil || *req.N != candidate {
		t.Fatalf("意外的 N")
	}
	if req.Stream == nil || *req.Stream != stream {
		t.Fatalf("意外的 Stream")
	}
	if req.StreamOptions == nil || req.StreamOptions.IncludeUsage == nil || *req.StreamOptions.IncludeUsage != includeUsage {
		t.Fatalf("意外的 StreamOptions")
	}
	if req.Metadata == nil || len(req.Metadata) != 1 || req.Metadata["ok"] != "yes" {
		t.Fatalf("意外的 Metadata")
	}
	if req.ServiceTier == nil || string(*req.ServiceTier) != serviceTier {
		t.Fatalf("意外的 ServiceTier")
	}
	if req.ReasoningEffort == nil || string(*req.ReasoningEffort) != effort {
		t.Fatalf("意外的 ReasoningEffort")
	}
	if req.ResponseFormat == nil || req.ResponseFormat.JSONSchema == nil || req.ResponseFormat.JSONSchema.JSONSchema.Name != "schema" {
		t.Fatalf("意外的 ResponseFormat")
	}
	if len(req.Modalities) != 2 || req.Modalities[1] != chatTypes.ChatModalitiesAudio {
		t.Fatalf("意外的 Modalities")
	}
	if len(req.Tools) != 2 || req.Tools[0].Function == nil || req.Tools[1].Custom == nil {
		t.Fatalf("意外的 Tools")
	}
	if req.ToolChoice == nil || req.ToolChoice.Named == nil || req.ToolChoice.Named.Function.Name != "tool" {
		t.Fatalf("意外的 ToolChoice")
	}
	if req.ParallelToolCalls == nil || *req.ParallelToolCalls != parallel {
		t.Fatalf("意外的 ParallelToolCalls")
	}
	if req.Verbosity == nil || string(*req.Verbosity) != verbosity {
		t.Fatalf("意外的 Verbosity")
	}
	if req.SafetyIdentifier == nil || *req.SafetyIdentifier != safetyID {
		t.Fatalf("意外的 SafetyIdentifier")
	}
	if req.Audio == nil || req.Prediction == nil || req.WebSearchOptions == nil {
		t.Fatalf("意外的 VendorExtras 字段")
	}
	if req.ExtraFields == nil || req.ExtraFields["extra_field"] != "extra" {
		t.Fatalf("意外的 ExtraFields")
	}

	if len(req.Messages) != 2 {
		t.Fatalf("意外的消息长度：%d", len(req.Messages))
	}
	if req.Messages[0].Role != chatTypes.MessageRoleSystem {
		t.Fatalf("系统消息未插入")
	}
	if req.Messages[0].ExtraFields == nil || req.Messages[0].ExtraFields["sys_extra"] != "v" {
		t.Fatalf("意外的系统消息扩展")
	}
	userMsg := req.Messages[1]
	if userMsg.Content.ContentParts == nil || len(userMsg.Content.ContentParts) != len(parts) {
		t.Fatalf("意外的内容片段长度")
	}
	if userMsg.ToolCalls == nil || len(userMsg.ToolCalls) != 1 {
		t.Fatalf("意外的工具调用")
	}
	if userMsg.FunctionCall == nil || userMsg.Audio == nil || userMsg.Refusal == nil {
		t.Fatalf("意外的消息级别扩展")
	}
	if userMsg.ExtraFields == nil || userMsg.ExtraFields["extra"] != "x" {
		t.Fatalf("意外的消息 ExtraFields")
	}
}

func TestToContract_Complex(t *testing.T) {
	sysText := "system"
	msgText := "hello"
	refusal := "no"
	imgURL := "https://example.com/img.png"
	detail := shared.ImageDetailHigh
	audioData := "audio"
	format := chatTypes.AudioFormatMP3
	fileID := "file-id"
	fileName := "name.txt"
	toolArgs := "{\"a\":1}"
	topLogprobs := 2
	maxCompletion := 100
	seed := 9
	candidate := 3
	presence := 0.2
	frequency := 0.3
	stopValue := "stop"
	parallel := true

	functionCall := &chatTypes.RequestFunctionCall{Name: "f", Arguments: "{}"}
	assistantAudio := &chatTypes.AssistantAudio{ID: "audio-id"}

	messages := []chatTypes.RequestMessage{
		{
			Role: chatTypes.MessageRoleSystem,
			Content: chatTypes.MessageContent{
				StringValue: &sysText,
			},
			ExtraFields: map[string]interface{}{
				"sys_extra": "v",
			},
		},
		{
			Role: chatTypes.MessageRoleUser,
			Content: chatTypes.MessageContent{
				ContentParts: []chatTypes.ContentPart{
					{
						Type: chatTypes.ContentPartTypeText,
						Text: &msgText,
					},
					{
						Type: chatTypes.ContentPartTypeImageURL,
						ImageURL: &chatTypes.ImageURL{
							URL:    imgURL,
							Detail: &detail,
						},
					},
					{
						Type: chatTypes.ContentPartTypeInputAudio,
						InputAudio: &chatTypes.InputAudio{
							Data:   audioData,
							Format: format,
						},
					},
					{
						Type: chatTypes.ContentPartTypeFile,
						File: &chatTypes.InputFile{
							FileID:   &fileID,
							Filename: &fileName,
						},
					},
					{
						Type:    chatTypes.ContentPartTypeRefusal,
						Refusal: &refusal,
					},
					{
						Type: "unknown",
					},
				},
			},
			ToolCalls: []chatTypes.RequestToolCall{
				{
					ID:   "id",
					Type: chatTypes.RequestToolCallTypeFunction,
					Function: chatTypes.RequestToolCallFunction{
						Name:      "tool",
						Arguments: toolArgs,
					},
				},
			},
			FunctionCall: functionCall,
			Refusal:      &refusal,
			Audio:        assistantAudio,
			ExtraFields: map[string]interface{}{
				"extra": "x",
			},
		},
	}

	strict := true
	req := &chatTypes.Request{
		Model:               "gpt-test",
		Messages:            messages,
		MaxCompletionTokens: &maxCompletion,
		Temperature:         nil,
		TopP:                nil,
		PresencePenalty:     &presence,
		FrequencyPenalty:    &frequency,
		Seed:                &seed,
		N:                   &candidate,
		TopLogprobs:         &topLogprobs,
		Stop:                &chatTypes.StopConfiguration{StringValue: &stopValue},
		ParallelToolCalls:   &parallel,
		ResponseFormat:      &chatTypes.FormatUnion{Text: &chatTypes.FormatText{Type: chatTypes.ResponseFormatTypeText}},
		Modalities:          []chatTypes.ChatModalities{chatTypes.ChatModalitiesText},
		Stream:              nil,
		StreamOptions:       nil,
		Logprobs:            nil,
		ToolChoice:          &shared.ToolChoiceUnion{Allowed: &shared.ToolChoiceAllowed{Type: "allowed_tools", Mode: "required", Tools: []map[string]interface{}{{"type": "function", "name": "tool"}}}},
		Tools: []chatTypes.ChatToolUnion{
			{
				Function: &chatTypes.ChatToolFunction{
					Type: "function",
					Function: shared.FunctionDefinition{
						Name:       "tool",
						Strict:     &strict,
						Parameters: map[string]interface{}{"a": "b"},
					},
				},
			},
			{
				Custom: &chatTypes.ChatToolCustom{
					Type:   "custom",
					Name:   strPtr("my_custom"),
					Custom: map[string]interface{}{"key": "value"},
				},
			},
		},
		Audio:            &chatTypes.RequestAudio{Format: chatTypes.AudioFormatMP3, Voice: "alloy"},
		Prediction:       &chatTypes.PredictionContent{Type: chatTypes.PredictionTypeContent},
		WebSearchOptions: &chatTypes.WebSearchOptions{SearchContextSize: ptrWebSearchContextSize(chatTypes.WebSearchContextSizeSmall)},
		Verbosity:        ptrVerbosity(shared.VerbosityLow),
		SafetyIdentifier: strPtr("safe"),
		LogitBias:        map[string]int{"1": 2},
		FunctionCall:     &chatTypes.FunctionCallUnion{Mode: ptrToolChoiceMode(chatTypes.ToolChoiceModeAuto)},
		Functions:        []shared.FunctionDefinition{{Name: "fn"}},
		ExtraFields: map[string]interface{}{
			"extra_field": "extra",
		},
	}

	contract, err := RequestToContract(req)
	if err != nil {
		t.Fatalf("ToContract 错误: %v", err)
	}
	if contract == nil {
		t.Fatalf("ToContract 返回 nil")
	}
	if contract.Source != types.VendorSourceOpenAIChat {
		t.Fatalf("意外的 Source")
	}
	if contract.MaxOutputTokens == nil || *contract.MaxOutputTokens != maxCompletion {
		t.Fatalf("意外的 MaxOutputTokens")
	}
	if contract.Stop == nil || contract.Stop.Text == nil || *contract.Stop.Text != stopValue {
		t.Fatalf("意外的 Stop")
	}
	if contract.CandidateCount == nil || *contract.CandidateCount != candidate {
		t.Fatalf("意外的 CandidateCount")
	}
	if contract.ParallelToolCalls == nil || *contract.ParallelToolCalls != parallel {
		t.Fatalf("意外的 ParallelToolCalls")
	}
	if contract.ToolChoice == nil || contract.ToolChoice.Mode == nil || *contract.ToolChoice.Mode != "required" {
		t.Fatalf("意外的 ToolChoice")
	}
	if len(contract.Tools) != 2 || contract.Tools[0].Function == nil || contract.Tools[1].Type != "custom" {
		t.Fatalf("意外的 Tools")
	}
	if contract.Tools[0].VendorExtras == nil || contract.Tools[0].VendorExtras["strict"] != true {
		t.Fatalf("意外的 Tool VendorExtras")
	}
	if contract.VendorExtras == nil || contract.VendorExtrasSource == nil {
		t.Fatalf("意外的 VendorExtras")
	}
	if contract.VendorExtras["extra_field"] != "extra" {
		t.Fatalf("意外的 ExtraFields 合并")
	}
	if contract.System == nil || contract.System.Text == nil || *contract.System.Text != sysText {
		t.Fatalf("意外的 System")
	}
	if contract.System.VendorExtras == nil || contract.System.VendorExtras["sys_extra"] != "v" {
		t.Fatalf("意外的 System VendorExtras")
	}
	if len(contract.Messages) != 1 {
		t.Fatalf("意外的消息长度：%d", len(contract.Messages))
	}
	msg := contract.Messages[0]
	if msg.VendorExtras == nil || msg.VendorExtras["extra"] != "x" {
		t.Fatalf("意外的消息 VendorExtras")
	}
	if len(msg.ToolCalls) != 1 || msg.ToolCalls[0].Arguments == nil || *msg.ToolCalls[0].Arguments != toolArgs {
		t.Fatalf("意外的 ToolCalls")
	}
	if len(msg.Content.Parts) != 6 {
		t.Fatalf("意外的 Content Parts")
	}
	if msg.Content.Parts[4].VendorExtras == nil || msg.Content.Parts[4].VendorExtras["refusal"] != &refusal {
		t.Fatalf("意外的 Refusal Part")
	}
	if msg.Content.Parts[5].VendorExtras == nil || msg.Content.Parts[5].VendorExtras["original_part"] == nil {
		t.Fatalf("意外的 Unknown Part")
	}

	if !reflect.DeepEqual(contract.Modalities, []string{"text"}) {
		t.Fatalf("意外的 Modalities")
	}
}

func TestToContract_Nil(t *testing.T) {
	contract, err := RequestToContract(nil)
	if err != nil {
		t.Fatalf("ToContract(nil) 错误: %v", err)
	}
	if contract != nil {
		t.Fatalf("ToContract(nil) 应返回 nil")
	}
}

func TestFromContract_Nil(t *testing.T) {
	req, err := RequestFromContract(nil)
	if err != nil {
		t.Fatalf("FromContract(nil) 错误: %v", err)
	}
	if req != nil {
		t.Fatalf("FromContract(nil) 应返回 nil")
	}
}

func strPtr(s string) *string {
	return &s
}

func ptrVerbosity(v shared.VerbosityLevel) *shared.VerbosityLevel {
	return &v
}

func ptrToolChoiceMode(v chatTypes.ToolChoiceMode) *chatTypes.ToolChoiceMode {
	return &v
}

func ptrWebSearchContextSize(v chatTypes.WebSearchContextSize) *chatTypes.WebSearchContextSize {
	return &v
}
