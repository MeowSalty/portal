package types

// VendorSource 表示供应商来源标记。
type VendorSource string

const (
	VendorSourceAnthropic      VendorSource = "anthropic"
	VendorSourceGemini         VendorSource = "google"
	VendorSourceOpenAIChat     VendorSource = "openai.chat"
	VendorSourceOpenAIResponse VendorSource = "openai.responses"
)

// RequestContract 表示统一的请求中间格式。
type RequestContract struct {
	Source VendorSource `json:"source"`

	Model string `json:"model"`

	Messages []Message `json:"messages,omitempty"`
	Prompt   *string   `json:"prompt,omitempty"`
	System   *System   `json:"system,omitempty"`

	MaxOutputTokens  *int     `json:"max_output_tokens,omitempty"`
	Temperature      *float64 `json:"temperature,omitempty"`
	TopP             *float64 `json:"top_p,omitempty"`
	TopK             *int     `json:"top_k,omitempty"`
	Stop             *Stop    `json:"stop,omitempty"`
	Seed             *int     `json:"seed,omitempty"`
	PresencePenalty  *float64 `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"`
	CandidateCount   *int     `json:"candidate_count,omitempty"`
	Logprobs         *bool    `json:"logprobs,omitempty"`
	TopLogprobs      *int     `json:"top_logprobs,omitempty"`

	Stream        *bool         `json:"stream,omitempty"`
	StreamOptions *StreamOption `json:"stream_options,omitempty"`

	Tools             []Tool      `json:"tools,omitempty"`
	ToolChoice        *ToolChoice `json:"tool_choice,omitempty"`
	ParallelToolCalls *bool       `json:"parallel_tool_calls,omitempty"`

	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`
	Modalities     []string        `json:"modalities,omitempty"`
	Reasoning      *Reasoning      `json:"reasoning,omitempty"`

	Metadata             map[string]interface{} `json:"metadata,omitempty"`
	User                 *string                `json:"user,omitempty"`
	ServiceTier          *string                `json:"service_tier,omitempty"`
	PromptCacheKey       *string                `json:"prompt_cache_key,omitempty"`
	PromptCacheRetention *string                `json:"prompt_cache_retention,omitempty"`
	Store                *bool                  `json:"store,omitempty"`

	VendorExtras       map[string]interface{} `json:"-"`
	VendorExtrasSource *VendorSource          `json:"-"`

	// 自定义 HTTP 头部（不会被序列化到请求体中）
	// 用于透传 User-Agent、Referer 等 HTTP 头部信息
	Headers map[string]string `json:"-"`
}

// System 表示系统指令。
type System struct {
	Text  *string       `json:"text,omitempty"`
	Parts []ContentPart `json:"parts,omitempty"`

	VendorExtras       map[string]interface{} `json:"-"`
	VendorExtrasSource *VendorSource          `json:"-"`
}

// Message 表示统一消息结构。
type Message struct {
	Role       string  `json:"role"`
	Name       *string `json:"name,omitempty"`
	ToolCallID *string `json:"tool_call_id,omitempty"`

	Content   Content    `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`

	VendorExtras       map[string]interface{} `json:"-"`
	VendorExtrasSource *VendorSource          `json:"-"`
}

// Content 表示文本或结构化内容。
type Content struct {
	Text  *string       `json:"text,omitempty"`
	Parts []ContentPart `json:"parts,omitempty"`
}

// ContentPart 表示多模态内容片段。
type ContentPart struct {
	Type string `json:"type"`

	Text  *string `json:"text,omitempty"`
	Image *Image  `json:"image,omitempty"`
	Audio *Audio  `json:"audio,omitempty"`
	Video *Video  `json:"video,omitempty"`
	File  *File   `json:"file,omitempty"`

	ToolCall   *ToolCall   `json:"tool_call,omitempty"`
	ToolResult *ToolResult `json:"tool_result,omitempty"`

	VendorExtras       map[string]interface{} `json:"-"`
	VendorExtrasSource *VendorSource          `json:"-"`
}

// Image 表示图像内容。
type Image struct {
	URL    *string `json:"url,omitempty"`
	Data   *string `json:"data,omitempty"`
	Detail *string `json:"detail,omitempty"`
	MIME   *string `json:"mime,omitempty"`
}

// Audio 表示音频内容。
type Audio struct {
	Data   *string `json:"data,omitempty"`
	Format *string `json:"format,omitempty"`
	MIME   *string `json:"mime,omitempty"`
}

// Video 表示视频内容。
type Video struct {
	URL   *string  `json:"url,omitempty"`
	Data  *string  `json:"data,omitempty"`
	MIME  *string  `json:"mime,omitempty"`
	FPS   *float64 `json:"fps,omitempty"`
	Start *string  `json:"start,omitempty"`
	End   *string  `json:"end,omitempty"`
}

// File 表示文件内容。
type File struct {
	ID       *string `json:"id,omitempty"`
	URL      *string `json:"url,omitempty"`
	Data     *string `json:"data,omitempty"`
	Filename *string `json:"filename,omitempty"`
	MIME     *string `json:"mime,omitempty"`
}

// Tool 表示工具定义。
type Tool struct {
	Type string `json:"type"`

	Function *Function              `json:"function,omitempty"`
	Config   map[string]interface{} `json:"config,omitempty"`

	VendorExtras       map[string]interface{} `json:"-"`
	VendorExtrasSource *VendorSource          `json:"-"`
}

// Function 表示函数工具定义。
type Function struct {
	Name           string      `json:"name"`
	Description    *string     `json:"description,omitempty"`
	Parameters     interface{} `json:"parameters,omitempty"`
	ResponseSchema interface{} `json:"response_schema,omitempty"`
}

// ToolChoice 表示工具选择。
type ToolChoice struct {
	Mode     *string  `json:"mode,omitempty"`
	Function *string  `json:"function,omitempty"`
	Allowed  []string `json:"allowed,omitempty"`
}

// ToolCall 表示工具调用。
type ToolCall struct {
	ID        *string                `json:"id,omitempty"`
	Type      *string                `json:"type,omitempty"`
	Name      *string                `json:"name,omitempty"`
	Arguments *string                `json:"arguments,omitempty"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
}

// ToolResult 表示工具调用结果。
type ToolResult struct {
	ID      *string                `json:"id,omitempty"`
	Name    *string                `json:"name,omitempty"`
	Content *string                `json:"content,omitempty"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

// Stop 表示停止条件。
type Stop struct {
	Text *string  `json:"text,omitempty"`
	List []string `json:"list,omitempty"`
}

// Reasoning 表示推理相关配置。
type Reasoning struct {
	Effort          *string `json:"effort,omitempty"`
	Summary         *string `json:"summary,omitempty"`
	GenerateSummary *string `json:"generate_summary,omitempty"`

	Budget          *int    `json:"budget,omitempty"`
	Level           *string `json:"level,omitempty"`
	IncludeThoughts *bool   `json:"include_thoughts,omitempty"`

	Mode *string `json:"mode,omitempty"`

	VendorExtras       map[string]interface{} `json:"-"`
	VendorExtrasSource *VendorSource          `json:"-"`
}

// ResponseFormat 表示响应格式。
type ResponseFormat struct {
	Type       string      `json:"type"`
	JSONSchema interface{} `json:"json_schema,omitempty"`
	MimeType   *string     `json:"mime_type,omitempty"`
}

// StreamOption 表示流式选项。
type StreamOption struct {
	IncludeUsage       *bool `json:"include_usage,omitempty"`
	IncludeObfuscation *bool `json:"include_obfuscation,omitempty"`

	VendorExtras       map[string]interface{} `json:"-"`
	VendorExtrasSource *VendorSource          `json:"-"`
}
