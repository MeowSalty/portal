package types

import (
	"encoding/json"
)

//  消息类型定义

type TextContentPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ImageUrl struct {
	URL     string `json:"url"`
	Details string `json:"details,omitempty"`
}

type ImageContentPart struct {
	Type     string   `json:"type"`
	ImageUrl ImageUrl `json:"image_url"`
}

type InputAudio struct {
	Data   string `json:"data"`
	Format string `json:"format"`
}

type AudioContentPart struct {
	Type       string     `json:"type"` // 始终是 `input_audio`
	InputAudio InputAudio `json:"input_audio"`
}

type File struct {
	FileID   string `json:"file_id"`
	FileName string `json:"filename"`
	FileData string `json:"file_data"`
}

type FileContentPart struct {
	Type string `json:"type"` // 始终是 `file`
	File File   `json:"file"`
}

// RequestMessage 表示聊天完成请求中的消息
type RequestMessage struct {
	Role       string      `json:"role,omitempty"`
	Name       string      `json:"name,omitempty"`
	Content    interface{} `json:"content"` // 支持字符串或内容部分数组
	ToolCallID string      `json:"tool_call_id,omitempty"`
	Refusal    string      `json:"refusal,omitempty"`
	ToolCalls  *Tool       `json:"tool_calls,omitempty"`
	Audio      *struct {
		ID string `json:"id"`
	} `json:"audio,omitempty"`
}

//  响应格式

type JsonSchema struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Schema      json.RawMessage `json:"schema,omitempty"`
	Strict      bool            `json:"strict,omitempty"`
}

// ResponseFormat 结构体定义了响应格式
type ResponseFormat struct {
	Type       string      `json:"type"` // 类型，"text"、"json_object"、"json_schema"
	JsonSchema *JsonSchema `json:"json_schema,omitempty"`
}

//  中的工具

// Function 结构体定义了一个函数工具
type Function struct {
	Name        string          `json:"name"`                  // 函数名称
	Description string          `json:"description,omitempty"` // 函数描述
	Parameters  json.RawMessage `json:"parameters"`            // 函数参数
	Strict      bool            `json:"strict,omitempty"`
	Arguments   string          `json:"arguments,omitempty"`
}

type Grammar struct {
	Definition string `json:"definition"`
	Syntax     string `json:"syntax"`
}

type Format struct {
	Type    string   `json:"type"`
	Grammar *Grammar `json:"grammar,omitempty"`
}

type Custom struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Format      *Format `json:"parameters"`
	Input       string  `json:"input,omitempty"`
}

// Tool 结构体定义了模型可以使用的工具
type Tool struct {
	ID       string    `json:"id,omitempty"`
	Type     string    `json:"type"`               // 工具类型，function 或 custom
	Function *Function `json:"function,omitempty"` // 工具函数
	Custom   *Custom   `json:"custom,omitempty"`
}

// ChatCompletionRequest 表示  聊天补全 API 的请求体
//
// 存在暂未完全支持的参数：
//
//   - metadata
//   - modalities
//   - parallel_tool_calls
//   - prediction
//   - prompt_cache_key
//   - safety_identifier
//   - service_tier
//   - store
//   - stream_options
//   - tool_choice
//   - web_search_options
//
// 处于废弃状态且不打算支持的参数：
//   - function_call
//   - functions
//   - max_tokens
//   - user
//
// 请参阅 https://platform.openai.com/docs/api-reference/chat/create
type ChatCompletionRequest struct {
	// 必要参数
	Model    string           `json:"model"`    // 模型名称
	Messages []RequestMessage `json:"messages"` // 消息列表

	// 可选参数
	FrequencyPenalty    float64         `json:"frequency_penalty,omitempty"`
	LogitBias           map[string]int  `json:"logit_bias,omitempty"`
	LogProbs            bool            `json:"logprobs,omitempty"`
	MaxCompletionTokens int             `json:"max_completion_tokens,omitempty"`
	N                   int             `json:"n,omitempty"`
	PresencePenalty     float64         `json:"presence_penalty,omitempty"`
	ReasoningEffort     string          `json:"reasoning_effort,omitempty"`
	ResponseFormat      *ResponseFormat `json:"response_format,omitempty"`
	Seed                int             `json:"seed,omitempty"`
	Stop                *StopField      `json:"stop,omitempty"`   // 停止序列
	Stream              bool            `json:"stream,omitempty"` // 是否使用流式响应
	Temperature         float64         `json:"temperature,omitempty"`
	Tools               []Tool          `json:"tools,omitempty"`
	TopLogProbs         int             `json:"top_logprobs,omitempty"`
	TopP                float64         `json:"top_p,omitempty"`
	Verbosity           string          `json:"verbosity,omitempty"`
	Audio               *struct {
		Format string `json:"format"`
		Voice  string `json:"voice"`
	} `json:"audio,omitempty"`
}

type UrlCitations struct {
	Url        string `json:"url"`
	Title      string `json:"title"`
	StartIndex int    `json:"start_index"`
	EndIndex   int    `json:"end_index"`
}

type Annotations struct {
	Type         string        `json:"type"`
	UrlCitations *UrlCitations `json:"url_citations,omitempty"`
}

type ResponseMessage struct {
	Role        string        `json:"role"`
	Content     *string       `json:"content"`
	Refusal     *string       `json:"refusal"`
	Annotations []Annotations `json:"annotations"`
}

type Delta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content"`
}

type ResponseLogprobs struct {
	Content *struct {
		Bytes       *[]byte `json:"bytes"`
		LogProb     float64 `json:"logprob"`
		Token       string  `json:"token"`
		TopLogProbs []struct {
			Bytes   *[]byte `json:"bytes"`
			LogProb float64 `json:"logprob"`
			Token   string  `json:"token"`
		} `json:"top_logprobs"`
	} `json:"content"`
	Refusal *string `json:"refusal"`
}

type Choices struct {
	Index        int               `json:"index"`
	Message      *ResponseMessage  `json:"message,omitempty"` // 非流时存在
	Delta        *Delta            `json:"delta,omitempty"`   // 流时存在
	Logprobs     *ResponseLogprobs `json:"logprobs,omitempty"`
	FinishReason *string           `json:"finish_reason"`
}

type Usage struct {
	PromptTokens        int `json:"prompt_tokens"`
	CompletionTokens    int `json:"completion_tokens"`
	TotalTokens         int `json:"total_tokens"`
	PromptTokensDetails struct {
		CachedTokens int `json:"cached_tokens"`
		AudioTokens  int `json:"audio_tokens"`
	} `json:"prompt_tokens_details"`
	CompletionTokensDetails struct {
		ReasoningTokens          int `json:"reasoning_tokens"`
		AudioTokens              int `json:"audio_tokens"`
		AcceptedPredictionTokens int `json:"accepted_prediction_tokens"`
		RejectedPredictionTokens int `json:"rejected_prediction_tokens"`
	} `json:"completion_tokens_details"`
}

// ChatCompletionResponse 表示来自  API 的响应
type ChatCompletionResponse struct {
	ID                string    `json:"id"`
	Object            string    `json:"object"`
	Created           int       `json:"created"`
	Model             string    `json:"model"`
	Choices           []Choices `json:"choices"`
	Usage             *Usage    `json:"usage,omitempty"` // 非流时存在
	ServiceTier       *string   `json:"service_tier,omitempty"`
	SystemFingerprint string    `json:"system_fingerprint,omitempty"`
}
