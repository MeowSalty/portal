package types

// StopReason 停止原因。
type StopReason = string

const (
	StopReasonEndTurn   StopReason = "end_turn"
	StopReasonMaxTokens StopReason = "max_tokens"
	StopReasonStopSeq   StopReason = "stop_sequence"
	StopReasonToolUse   StopReason = "tool_use"
	StopReasonPauseTurn StopReason = "pause_turn"
	StopReasonRefusal   StopReason = "refusal"
)

// ResponseType 响应对象类型。
type ResponseType string

const (
	ResponseTypeMessage ResponseType = "message"
)

// ResponseServiceTier 响应服务层级。
type ResponseServiceTier = string

const (
	ResponseServiceTierStandard ResponseServiceTier = "standard"
	ResponseServiceTierPriority ResponseServiceTier = "priority"
	ResponseServiceTierBatch    ResponseServiceTier = "batch"
)

// ResponseContentBlockType 响应内容块类型。
type ResponseContentBlockType string

const (
	ResponseContentBlockText                ResponseContentBlockType = "text"
	ResponseContentBlockThinking            ResponseContentBlockType = "thinking"
	ResponseContentBlockRedactedThinking    ResponseContentBlockType = "redacted_thinking"
	ResponseContentBlockToolUse             ResponseContentBlockType = "tool_use"
	ResponseContentBlockServerToolUse       ResponseContentBlockType = "server_tool_use"
	ResponseContentBlockWebSearchToolResult ResponseContentBlockType = "web_search_tool_result"
)
