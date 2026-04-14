package adapter

import (
	"encoding/json"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/logger"
	"github.com/MeowSalty/portal/request/adapter/anthropic/converter"
	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
	adapterTypes "github.com/MeowSalty/portal/request/adapter/types"
	"github.com/MeowSalty/portal/routing"
)

// Anthropic Anthropic 提供商实现
type Anthropic struct {
	logger logger.Logger
}

// init 函数注册 Anthropic 提供商
func init() {
	RegisterProviderFactory("anthropic", func() Provider {
		return NewAnthropicProvider()
	})
}

// NewAnthropicProvider 创建新的 Anthropic 提供商
func NewAnthropicProvider() *Anthropic {
	return &Anthropic{}
}

// Name 返回提供商名称
func (p *Anthropic) Name() string {
	return "anthropic"
}

// CreateRequest 创建 Anthropic 请求
func (p *Anthropic) CreateRequest(request *adapterTypes.RequestContract, channel *routing.Channel) (interface{}, error) {
	request.Model = channel.ModelName
	return converter.RequestFromContract(request)
}

// ParseResponse 解析 Anthropic 响应
func (p *Anthropic) ParseResponse(variant string, responseData []byte) (*adapterTypes.ResponseContract, error) {
	// 首先检查是否是错误响应
	var errorResp anthropicTypes.ErrorResponse
	if err := json.Unmarshal(responseData, &errorResp); err == nil && errorResp.Type == "error" {
		return &adapterTypes.ResponseContract{
			Error: &adapterTypes.ResponseError{
				Code:    &errorResp.Error.Type,
				Message: &errorResp.Error.Message,
				Type:    &errorResp.Error.Type,
			},
		}, nil
	}

	var response anthropicTypes.Response
	if err := json.Unmarshal(responseData, &response); err != nil {
		return nil, err
	}
	return converter.ResponseToContract(&response, p.logger)
}

// ParseStreamResponse 解析 Anthropic 流式响应
func (p *Anthropic) ParseStreamResponse(variant string, ctx adapterTypes.StreamIndexContext, responseData []byte) ([]*adapterTypes.StreamEventContract, error) {
	var event anthropicTypes.StreamEvent
	if err := json.Unmarshal(responseData, &event); err != nil {
		return nil, err
	}

	converted, err := converter.StreamEventToContract(&event, ctx, p.logger)
	if err != nil {
		return nil, err
	}
	if converted == nil {
		return nil, nil
	}
	return []*adapterTypes.StreamEventContract{converted}, nil
}

// APIEndpoint 返回 API 端点
func (p *Anthropic) APIEndpoint(variant string, model string, stream bool, config ...string) string {
	defaultEndpoint := "/v1/messages"

	// 如果没有提供 config，使用默认端点
	if len(config) == 0 || config[0] == "" {
		return defaultEndpoint
	}

	c := config[0]

	// 如果 config 以 "/" 结尾，视为前缀，拼接默认端点
	if len(c) > 0 && c[len(c)-1] == '/' {
		return c + defaultEndpoint
	}

	// 其他情况，视为完整路径
	return c
}

// Headers 返回特定头部
func (p *Anthropic) Headers(key string) map[string]string {
	headers := map[string]string{
		"x-api-key":         key,
		"anthropic-version": "2023-06-01",
		"Content-Type":      "application/json",
	}

	return headers
}

// SupportsStreaming 是否支持流式传输
func (p *Anthropic) SupportsStreaming() bool {
	return true
}

// SupportsNative 返回是否支持原生 API 调用
func (p *Anthropic) SupportsNative() bool {
	return true
}

// BuildNativeRequest 构建原生请求
func (p *Anthropic) BuildNativeRequest(channel *routing.Channel, payload any) (body any, err error) {
	if req, ok := payload.(*anthropicTypes.Request); ok {
		req.Model = channel.ModelName
		return req, nil
	}
	return nil, errors.New(errors.ErrCodeInvalidArgument, "无效的请求类型，期望 anthropicTypes.Request")
}

// ParseNativeResponse 解析原生响应
func (p *Anthropic) ParseNativeResponse(variant string, raw []byte) (any, error) {
	var response anthropicTypes.Response
	if err := json.Unmarshal(raw, &response); err != nil {
		return nil, errors.Wrap(errors.ErrCodeInvalidArgument, "解析 Anthropic 响应失败", err)
	}
	return &response, nil
}

// ParseNativeStreamEvent 解析原生流事件
func (p *Anthropic) ParseNativeStreamEvent(variant string, raw []byte) (any, error) {
	var event anthropicTypes.StreamEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		return nil, errors.Wrap(errors.ErrCodeInvalidArgument, "解析 Anthropic 流事件失败", err)
	}
	return &event, nil
}

// ExtractUsageFromNativeStreamEvent 从原生流事件中提取使用统计信息
func (p *Anthropic) ExtractUsageFromNativeStreamEvent(variant string, event any) *adapterTypes.ResponseUsage {
	streamEvent, ok := event.(*anthropicTypes.StreamEvent)
	if !ok {
		return nil
	}
	if streamEvent.MessageDelta == nil || streamEvent.MessageDelta.Usage == nil {
		return nil
	}
	usage := streamEvent.MessageDelta.Usage
	if usage.InputTokens == nil || usage.OutputTokens == nil {
		return nil
	}
	totalTokens := *usage.InputTokens + *usage.OutputTokens
	return &adapterTypes.ResponseUsage{
		InputTokens:  usage.InputTokens,
		OutputTokens: usage.OutputTokens,
		TotalTokens:  &totalTokens,
	}
}

// IdentifyStreamEventSignal 识别 Anthropic 原生流事件的信号类型。
//
// Anthropic 的完成信号识别规则：
//   - message_stop 事件为消息级完成信号（IsCompletionSignal + IsTerminalEvent）
//     表示整个消息生成结束，可视为 completed
//   - message_delta 包含 stop_reason 时为终止事件（IsCompletionSignal + IsTerminalEvent）
//     stop_reason 提供具体完成原因（end_turn/tool_use/max_tokens/stop_sequence/pause_turn/refusal）
//   - content_block_stop 事件仅为单个内容块结束，不能误判为整体完成
//     它既不是 IsCompletionSignal 也不是 IsTerminalEvent
//   - content_block_delta 包含文本或工具调用增量时为有效输出
//   - error 事件为流级错误信号
func (p *Anthropic) IdentifyStreamEventSignal(variant string, event any) StreamEventSignal {
	signal := StreamEventSignal{}

	streamEvent, ok := event.(*anthropicTypes.StreamEvent)
	if !ok {
		return signal
	}

	// 检查 message_stop 事件（消息级完成信号）
	// message_stop 表示整个消息生成结束，是明确的完成信号。
	// 注意：message_stop 通常在 message_delta（含 stop_reason）之后到达，
	// 两者都会设置 IsCompletionSignal，但 FinishReason 以先到达的 message_delta 为准。
	if streamEvent.MessageStop != nil {
		signal.IsCompletionSignal = true
		signal.IsTerminalEvent = true
		// 仅在未已有 FinishReason 时设置默认值，
		// 避免覆盖 message_delta 中更具体的 stop_reason
		if signal.FinishReason == "" {
			signal.FinishReason = "stop"
		}
	}

	// 检查 message_delta 事件（包含 stop_reason）
	// message_delta 中的 stop_reason 是 Anthropic 协议的完成原因载体，
	// 比 message_stop 更早到达且包含更具体的完成原因。
	if streamEvent.MessageDelta != nil {
		if streamEvent.MessageDelta.Delta.StopReason != nil {
			signal.IsCompletionSignal = true
			signal.IsTerminalEvent = true
			signal.FinishReason = string(*streamEvent.MessageDelta.Delta.StopReason)
		}
	}

	// 注意：content_block_stop 不设置 IsCompletionSignal 或 IsTerminalEvent
	// content_block_stop 仅表示单个内容块（如文本块、工具调用块）结束，
	// 并不代表整个消息生成完成。在多内容块场景中，一个块结束后还有后续块。
	// 此处仅标记 HasValidOutput，不误判为整体完成。

	// 检查 content_block_delta 事件（有效输出）
	if streamEvent.ContentBlockDelta != nil {
		delta := streamEvent.ContentBlockDelta.Delta
		// 文本增量
		if delta.Text != nil && delta.Text.Text != "" {
			signal.HasValidOutput = true
		}
		// 思考增量
		if delta.Thinking != nil && delta.Thinking.Thinking != "" {
			signal.HasValidOutput = true
		}
		// 工具调用 JSON 增量
		if delta.InputJSON != nil && delta.InputJSON.PartialJSON != "" {
			signal.HasValidOutput = true
		}
	}

	// 检查 content_block_start 事件（工具调用开始）
	if streamEvent.ContentBlockStart != nil {
		if streamEvent.ContentBlockStart.ContentBlock.ToolUse != nil {
			signal.HasValidOutput = true
		}
	}

	// 检查错误事件
	if streamEvent.Error != nil {
		signal.IsCompletionSignal = true
		signal.IsTerminalEvent = true
		signal.FinishReason = "error"
	}

	return signal
}
