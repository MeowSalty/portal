package responses

// InputItemType 表示 Responses 输入项类型
type InputItemType = string

const (
	// 消息类型
	InputItemTypeMessage InputItemType = "message"

	// 引用类型
	InputItemTypeItemReference InputItemType = "item_reference"

	// 工具调用类型
	InputItemTypeFunctionCall        InputItemType = "function_call"
	InputItemTypeFileSearchCall      InputItemType = "file_search_call"
	InputItemTypeWebSearchCall       InputItemType = "web_search_call"
	InputItemTypeComputerCall        InputItemType = "computer_call"
	InputItemTypeCodeInterpreterCall InputItemType = "code_interpreter_call"
	InputItemTypeImageGenCall        InputItemType = "image_generation_call"
	InputItemTypeLocalShellCall      InputItemType = "local_shell_call"
	InputItemTypeFunctionShellCall   InputItemType = "shell_call"
	InputItemTypeApplyPatchCall      InputItemType = "apply_patch_call"
	InputItemTypeMCPCall             InputItemType = "mcp_call"
	InputItemTypeCustomToolCall      InputItemType = "custom_tool_call"
	InputItemTypeMCPListTools        InputItemType = "mcp_list_tools"

	// 工具调用输出类型
	InputItemTypeFunctionCallOutput      InputItemType = "function_call_output"
	InputItemTypeComputerCallOutput      InputItemType = "computer_call_output"
	InputItemTypeLocalShellCallOutput    InputItemType = "local_shell_call_output"
	InputItemTypeFunctionShellCallOutput InputItemType = "shell_call_output"
	InputItemTypeApplyPatchCallOutput    InputItemType = "apply_patch_call_output"
	InputItemTypeCustomToolCallOutput    InputItemType = "custom_tool_call_output"

	// 其他类型
	InputItemTypeReasoning           InputItemType = "reasoning"
	InputItemTypeCompaction          InputItemType = "compaction"
	InputItemTypeMCPApprovalRequest  InputItemType = "mcp_approval_request"
	InputItemTypeMCPApprovalResponse InputItemType = "mcp_approval_response"
)

// ResponseMessageRole 表示 Responses 消息角色
type ResponseMessageRole = string

const (
	ResponseMessageRoleUser      ResponseMessageRole = "user"
	ResponseMessageRoleSystem    ResponseMessageRole = "system"
	ResponseMessageRoleDeveloper ResponseMessageRole = "developer"
	ResponseMessageRoleAssistant ResponseMessageRole = "assistant"
)

// InputContentType 表示 Responses 输入内容类型
type InputContentType = string

const (
	InputContentTypeText  InputContentType = "input_text"
	InputContentTypeImage InputContentType = "input_image"
	InputContentTypeFile  InputContentType = "input_file"
)

// TruncationStrategy 表示截断策略
type TruncationStrategy = string

const (
	TruncationStrategyAuto     TruncationStrategy = "auto"
	TruncationStrategyDisabled TruncationStrategy = "disabled"
)

// IncludeEnum 表示 include 参数
// 参考 OpenAI Responses 文档。
type IncludeEnum = string

// IncludeList 表示 include 列表
// 与 []string 兼容，便于与转换器交互。
type IncludeList = []IncludeEnum

const (
	IncludeFileSearchResults          IncludeEnum = "file_search_call.results"
	IncludeWebSearchResults           IncludeEnum = "web_search_call.results"
	IncludeWebSearchActionSources     IncludeEnum = "web_search_call.action.sources"
	IncludeMessageInputImageURL       IncludeEnum = "message.input_image.image_url"
	IncludeComputerCallOutputImageURL IncludeEnum = "computer_call_output.output.image_url"
	IncludeCodeInterpreterOutputs     IncludeEnum = "code_interpreter_call.outputs"
	IncludeReasoningEncryptedContent  IncludeEnum = "reasoning.encrypted_content"
	IncludeMessageOutputTextLogprobs  IncludeEnum = "message.output_text.logprobs"
)

// PromptCacheRetention 表示提示缓存保留策略
type PromptCacheRetention = string

const (
	PromptCacheRetentionInMemory PromptCacheRetention = "in-memory"
	PromptCacheRetention24h      PromptCacheRetention = "24h"
)

// ReasoningSummary 表示推理摘要级别
type ReasoningSummary = string

const (
	ReasoningSummaryAuto     ReasoningSummary = "auto"
	ReasoningSummaryConcise  ReasoningSummary = "concise"
	ReasoningSummaryDetailed ReasoningSummary = "detailed"
)

// TextResponseFormatType 表示文本响应格式类型
type TextResponseFormatType = string

const (
	TextResponseFormatTypeText       TextResponseFormatType = "text"
	TextResponseFormatTypeJSONSchema TextResponseFormatType = "json_schema"
	TextResponseFormatTypeJSONObject TextResponseFormatType = "json_object"
)
