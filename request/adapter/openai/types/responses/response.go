package responses

import (
	"encoding/json"

	portalErrors "github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/request/adapter/openai/types/shared"
)

// ResponseInstructions 表示响应中的系统指令
// 支持 string（文本指令）或 []InputItem（输入项数组）的联合类型
// 参考：https://platform.openai.com/docs/api-reference/responses/object
type ResponseInstructions struct {
	// String 表示文本指令
	String *string `json:"-"`
	// List 表示输入项数组
	List *[]InputItem `json:"-"`
}

// MarshalJSON 实现 ResponseInstructions 的自定义 JSON 序列化
// 优先序列化 String，其次序列化 List
func (r ResponseInstructions) MarshalJSON() ([]byte, error) {
	// 优先使用 String
	if r.String != nil {
		return json.Marshal(r.String)
	}
	// 其次使用 List
	if r.List != nil {
		return json.Marshal(r.List)
	}
	// 都为空则序列化为 null
	return json.Marshal(nil)
}

// UnmarshalJSON 实现 ResponseInstructions 的自定义 JSON 反序列化
// 支持 string 或 []InputItem
func (r *ResponseInstructions) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	// 尝试解析为 string
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*r = ResponseInstructions{String: &str}
		return nil
	}

	// 尝试解析为 []InputItem
	var list []InputItem
	if err := json.Unmarshal(data, &list); err == nil {
		*r = ResponseInstructions{List: &list}
		return nil
	}

	return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "instructions 必须是 string 或 []InputItem")
}

// NewResponseInstructionsFromString 创建文本指令
func NewResponseInstructionsFromString(text string) *ResponseInstructions {
	return &ResponseInstructions{String: &text}
}

// NewResponseInstructionsFromList 创建输入项数组指令
func NewResponseInstructionsFromList(list []InputItem) *ResponseInstructions {
	return &ResponseInstructions{List: &list}
}

// Response 表示 OpenAI Responses API 响应
// 字段与官方 Responses 文档对齐（非流式）。
// 参考：https://platform.openai.com/docs/api-reference/responses/object
type Response struct {
	// 必需字段
	ID                string                  `json:"id"`                  // 响应的唯一标识符
	Object            string                  `json:"object"`              // 对象类型，始终为 "response"
	CreatedAt         int64                   `json:"created_at"`          // 创建时间戳（秒）
	Model             string                  `json:"model"`               // 使用的模型 ID
	Output            []OutputItem            `json:"output"`              // 模型生成的内容项数组
	ParallelToolCalls bool                    `json:"parallel_tool_calls"` // 是否允许并行工具调用
	Metadata          map[string]string       `json:"metadata"`            // 元数据键值对
	ToolChoice        *shared.ToolChoiceUnion `json:"tool_choice"`         // 工具选择策略
	Tools             []shared.ToolUnion      `json:"tools"`               // 可用工具列表

	// 可选字段
	Status               *string               `json:"status,omitempty"`                 // 响应状态：completed, failed, in_progress, cancelled, queued, incomplete
	CompletedAt          *int64                `json:"completed_at,omitempty"`           // 完成时间戳（秒）
	Error                *ResponseError        `json:"error,omitempty"`                  // 错误对象
	IncompleteDetails    *IncompleteDetails    `json:"incomplete_details,omitempty"`     // 未完成原因详情
	Instructions         *ResponseInstructions `json:"instructions,omitempty"`           // 系统指令（string 或 []InputItem，强类型）
	Usage                *Usage                `json:"usage,omitempty"`                  // Token 使用情况
	Conversation         *ConversationRef      `json:"conversation,omitempty"`           // 关联的对话引用
	PreviousResponseID   *string               `json:"previous_response_id,omitempty"`   // 前一个响应的 ID
	Reasoning            *ResponseReasoning    `json:"reasoning,omitempty"`              // 推理信息
	Background           *bool                 `json:"background,omitempty"`             // 是否在后台运行
	MaxOutputTokens      *int                  `json:"max_output_tokens,omitempty"`      // 最大输出 token 数
	MaxToolCalls         *int                  `json:"max_tool_calls,omitempty"`         // 最大工具调用次数
	Text                 *TextConfig           `json:"text,omitempty"`                   // 文本配置
	TopP                 *float64              `json:"top_p,omitempty"`                  // 核采样参数
	Temperature          *float64              `json:"temperature,omitempty"`            // 采样温度
	Truncation           *string               `json:"truncation,omitempty"`             // 截断策略：auto, disabled
	User                 *string               `json:"user,omitempty"`                   // 用户标识符（已弃用）
	SafetyIdentifier     *string               `json:"safety_identifier,omitempty"`      // 安全标识符
	PromptCacheKey       *string               `json:"prompt_cache_key,omitempty"`       // 提示缓存键
	ServiceTier          *string               `json:"service_tier,omitempty"`           // 服务层级
	PromptCacheRetention *string               `json:"prompt_cache_retention,omitempty"` // 提示缓存保留策略
	TopLogprobs          *int                  `json:"top_logprobs,omitempty"`           // 返回的顶部 logprobs 数量
}

// ResponseError 表示 Responses API 错误对象。
type ResponseError struct {
	Code    string  `json:"code"`            // 错误代码
	Message string  `json:"message"`         // 错误描述
	Type    *string `json:"type,omitempty"`  // 错误类型
	Param   *string `json:"param,omitempty"` // 相关参数
}

// IncompleteDetails 表示未完成响应的原因。
type IncompleteDetails struct {
	Reason *string `json:"reason,omitempty"` // 未完成原因：max_output_tokens, content_filter
}

// ResponseReasoning 表示响应中的推理信息。
type ResponseReasoning struct {
	Effort  *string `json:"effort,omitempty"`  // 推理努力程度
	Summary *string `json:"summary,omitempty"` // 推理摘要
}

// OutputSummaryPart 表示推理摘要片段。
type OutputSummaryPart struct {
	Type string `json:"type"`           // 类型：summary_text
	Text string `json:"text,omitempty"` // 摘要文本
}

// LogProb 表示 token 对数概率。
type LogProb struct {
	Token       string       `json:"token"`        // Token 文本
	Logprob     float64      `json:"logprob"`      // 对数概率
	Bytes       []int        `json:"bytes"`        // UTF-8 字节表示
	TopLogprobs []TopLogProb `json:"top_logprobs"` // 顶部对数概率
}

// TopLogProb 表示顶部 token 对数概率。
type TopLogProb struct {
	Token   string  `json:"token"`   // Token 文本
	Logprob float64 `json:"logprob"` // 对数概率
	Bytes   []int   `json:"bytes"`   // UTF-8 字节表示
}

// ResponseLogProb 表示响应文本 token 对数概率。
type ResponseLogProb struct {
	Token       string               `json:"token"`        // Token 文本
	Logprob     float64              `json:"logprob"`      // 对数概率
	TopLogprobs []ResponseTopLogProb `json:"top_logprobs"` // 顶部对数概率
}

// ResponseTopLogProb 表示响应文本顶部 token 对数概率。
type ResponseTopLogProb struct {
	Token   string  `json:"token"`   // Token 文本
	Logprob float64 `json:"logprob"` // 对数概率
}

// Usage 表示 Responses 使用情况
// 所有字段都是必需的。
type Usage struct {
	InputTokens         int                 `json:"input_tokens"`          // 输入 token 数
	InputTokensDetails  InputTokensDetails  `json:"input_tokens_details"`  // 输入 token 详情
	OutputTokens        int                 `json:"output_tokens"`         // 输出 token 数
	OutputTokensDetails OutputTokensDetails `json:"output_tokens_details"` // 输出 token 详情
	TotalTokens         int                 `json:"total_tokens"`          // 总 token 数
}

// InputTokensDetails 表示输入 token 细节。
type InputTokensDetails struct {
	CachedTokens int `json:"cached_tokens"` // 缓存的 token 数
}

// OutputTokensDetails 表示输出 token 细节。
type OutputTokensDetails struct {
	ReasoningTokens int `json:"reasoning_tokens"` // 推理 token 数
}

// ConversationRef 表示对话引用信息
// 用于 response.conversation 字段。
type ConversationRef struct {
	ID string `json:"id"` // 对话的唯一 ID
}

// ==================== Shell Call 相关类型 ====================

// FunctionShellAction 表示 Shell 工具调用的动作结构。
// 参考：https://platform.openai.com/docs/api-reference/responses/object#responses/object-output
type FunctionShellAction struct {
	Commands        []string `json:"commands"`          // 要执行的命令列表
	TimeoutMs       *int     `json:"timeout_ms"`        // 超时时间（毫秒）
	MaxOutputLength *int     `json:"max_output_length"` // 最大输出长度
}

// FunctionShellCallOutputContent 表示 Shell 工具调用的输出内容项。
type FunctionShellCallOutputContent struct {
	Type    string `json:"type"`    // 输出类型：text, error, logs
	Content string `json:"content"` // 输出内容
}

// ==================== Apply Patch 相关类型 ====================

// ApplyPatchToolCallOperation 表示应用补丁工具调用的操作结构。
// 使用 discriminator 模式，根据 type 字段区分不同操作类型。
type ApplyPatchToolCallOperation struct {
	Type string `json:"type"` // 操作类型：create, delete, update

	// Create 操作字段
	Create *ApplyPatchCreateOperation `json:"create,omitempty"`

	// Delete 操作字段
	Delete *ApplyPatchDeleteOperation `json:"delete,omitempty"`

	// Update 操作字段
	Update *ApplyPatchUpdateOperation `json:"update,omitempty"`
}

// ApplyPatchCreateOperation 表示创建文件操作。
type ApplyPatchCreateOperation struct {
	Path    string `json:"path"`    // 文件路径
	Content string `json:"content"` // 文件内容
}

// ApplyPatchDeleteOperation 表示删除文件操作。
type ApplyPatchDeleteOperation struct {
	Path string `json:"path"` // 文件路径
}

// ApplyPatchUpdateOperation 表示更新文件操作。
type ApplyPatchUpdateOperation struct {
	Path    string `json:"path"`    // 文件路径
	Content string `json:"content"` // 文件内容
}

// ==================== MCP 相关类型 ====================

// MCPListToolsTool 表示 MCP 列出工具中的单个工具项。
type MCPListToolsTool struct {
	Name        string                 `json:"name"`                   // 工具名称
	Description string                 `json:"description"`            // 工具描述
	InputSchema map[string]interface{} `json:"input_schema,omitempty"` // 输入模式（可选）
}

// ==================== Reasoning 相关类型 ====================

// ReasoningTextContent 表示推理文本内容项。
type ReasoningTextContent struct {
	Type    string `json:"type"`    // 内容类型：text
	Content string `json:"content"` // 文本内容
}

// ==================== Code Interpreter 相关类型 ====================

// CodeInterpreterOutput 表示代码解释器输出项。
// 使用 discriminator 模式，根据 type 字段区分不同输出类型。
type CodeInterpreterOutput struct {
	Type string `json:"type"` // 输出类型：logs, image

	// Logs 输出字段
	Logs *CodeInterpreterLogsOutput `json:"logs,omitempty"`

	// Image 输出字段
	Image *CodeInterpreterImageOutput `json:"image,omitempty"`
}

// CodeInterpreterLogsOutput 表示代码解释器的日志输出。
type CodeInterpreterLogsOutput struct {
	Text string `json:"text"` // 日志文本
}

// CodeInterpreterImageOutput 表示代码解释器的图像输出。
type CodeInterpreterImageOutput struct {
	FileID string `json:"file_id"` // 文件 ID
}

// OutputContentPart 表示输出内容片段的联合类型。
// 使用多指针 oneof 结构，仅子结构体携带 `type` 字段。
//
// 强类型约束：OutputText 和 Refusal 互斥，仅能有一个非空。
type OutputContentPart struct {
	OutputText *OutputTextContent `json:"-"` // 文本输出内容
	Refusal    *RefusalContent    `json:"-"` // 拒绝内容
}

// UnmarshalJSON 实现 OutputContentPart 的自定义反序列化。
// 先读取 type 字段，再按类型反序列化到对应指针。
func (o *OutputContentPart) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	var base struct {
		Type OutputMessageContentType `json:"type"`
	}
	if err := json.Unmarshal(data, &base); err != nil {
		return err
	}
	if base.Type == "" {
		return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "输出内容片段类型为空")
	}

	*o = OutputContentPart{}

	switch base.Type {
	case OutputMessageContentTypeOutputText:
		var content OutputTextContent
		if err := json.Unmarshal(data, &content); err != nil {
			return err
		}
		content.Type = OutputMessageContentTypeOutputText
		o.OutputText = &content
	case OutputMessageContentTypeRefusal:
		var content RefusalContent
		if err := json.Unmarshal(data, &content); err != nil {
			return err
		}
		content.Type = OutputMessageContentTypeRefusal
		o.Refusal = &content
	default:
		return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "输出内容片段类型不支持")
	}

	return nil
}

// MarshalJSON 实现 OutputContentPart 的自定义序列化。
// 要求仅一个指针非空；两者皆空或同时非空报错。
func (o OutputContentPart) MarshalJSON() ([]byte, error) {
	// 互斥校验：仅能有一个非空
	if o.OutputText == nil && o.Refusal == nil {
		return nil, portalErrors.New(portalErrors.ErrCodeInvalidArgument, "输出内容片段不能为空")
	}
	if o.OutputText != nil && o.Refusal != nil {
		return nil, portalErrors.New(portalErrors.ErrCodeInvalidArgument, "输出内容片段类型冲突，不能同时存在 output_text 和 refusal")
	}

	// 序列化非空指针
	if o.OutputText != nil {
		// 确保 annotations 为空数组而非 null
		if o.OutputText.Annotations == nil {
			o.OutputText.Annotations = []Annotation{}
		}
		return json.Marshal(o.OutputText)
	}
	if o.Refusal != nil {
		return json.Marshal(o.Refusal)
	}

	return nil, portalErrors.New(portalErrors.ErrCodeInvalidArgument, "输出内容片段类型不支持")
}
