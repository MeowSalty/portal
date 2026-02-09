package converter

import (
	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// RequestFromContract 将 RequestContract 转换为 Anthropic 请求。
func RequestFromContract(contract *types.RequestContract) (*anthropicTypes.Request, error) {
	if contract == nil {
		return nil, nil
	}

	req := &anthropicTypes.Request{
		Model: contract.Model,
	}

	// 转换 Messages
	if len(contract.Messages) > 0 {
		messages, err := convertMessagesFromContract(contract.Messages)
		if err != nil {
			return nil, err
		}
		req.Messages = messages
	} else if contract.Prompt != nil {
		// 若 Prompt 不为空且 Messages 为空，构造单条 user message
		req.Messages = []anthropicTypes.Message{
			{
				Role: anthropicTypes.RoleUser,
				Content: anthropicTypes.MessageContentParam{
					StringValue: contract.Prompt,
				},
			},
		}
	}

	// 转换 System
	if contract.System != nil {
		system, err := convertSystemFromContract(contract.System)
		if err != nil {
			return nil, err
		}
		req.System = system
	}

	// 转换采样参数
	if contract.MaxOutputTokens != nil {
		req.MaxTokens = *contract.MaxOutputTokens
	}
	req.Temperature = contract.Temperature
	req.TopP = contract.TopP
	req.TopK = contract.TopK

	// 转换 Stop
	if contract.Stop != nil {
		if len(contract.Stop.List) > 0 {
			req.StopSequences = contract.Stop.List
		} else if contract.Stop.Text != nil {
			req.StopSequences = []string{*contract.Stop.Text}
		}
	}

	// 转换流式配置
	req.Stream = contract.Stream

	// 转换 Metadata
	if len(contract.Metadata) > 0 {
		req.Metadata = &anthropicTypes.Metadata{}
		if userID, ok := contract.Metadata["user_id"].(string); ok {
			req.Metadata.UserID = &userID
		}
	}

	// 转换 ServiceTier
	if contract.ServiceTier != nil {
		serviceTier := anthropicTypes.ServiceTier(*contract.ServiceTier)
		req.ServiceTier = &serviceTier
	}

	// 转换 Tools
	if len(contract.Tools) > 0 {
		tools, err := convertToolsFromContract(contract.Tools, contract.VendorExtras)
		if err != nil {
			return nil, err
		}
		req.Tools = tools
	}

	// 转换 ToolChoice
	if contract.ToolChoice != nil {
		toolChoice, err := convertToolChoiceFromContract(contract.ToolChoice, contract.ParallelToolCalls)
		if err != nil {
			return nil, err
		}
		req.ToolChoice = toolChoice
	}

	// 转换 Reasoning -> Thinking
	if contract.Reasoning != nil {
		thinking, err := convertReasoningFromContract(contract.Reasoning)
		if err != nil {
			return nil, err
		}
		req.Thinking = thinking
	}

	return req, nil
}
