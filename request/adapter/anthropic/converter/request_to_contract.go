package converter

import (
	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// RequestToContract 将 Anthropic 请求转换为统一的 RequestContract。
func RequestToContract(req *anthropicTypes.Request) (*types.RequestContract, error) {
	if req == nil {
		return nil, nil
	}

	contract := &types.RequestContract{
		Source: types.VendorSourceAnthropic,
		Model:  req.Model,
	}

	// 转换 Messages
	if len(req.Messages) > 0 {
		messages, err := convertMessagesToContract(req.Messages)
		if err != nil {
			return nil, err
		}
		contract.Messages = messages
	}

	// 转换 System
	if req.System != nil {
		system, err := convertSystemToContract(req.System)
		if err != nil {
			return nil, err
		}
		contract.System = system
	}

	// 转换采样参数
	contract.MaxOutputTokens = &req.MaxTokens
	contract.Temperature = req.Temperature
	contract.TopP = req.TopP
	contract.TopK = req.TopK

	// 转换 StopSequences
	if len(req.StopSequences) > 0 {
		contract.Stop = &types.Stop{
			List: req.StopSequences,
		}
	}

	// 转换流式配置
	contract.Stream = req.Stream

	// 转换 Metadata
	if req.Metadata != nil {
		contract.Metadata = make(map[string]interface{})
		if req.Metadata.UserID != nil {
			contract.Metadata["user_id"] = *req.Metadata.UserID
		}
	}

	// 转换 ServiceTier
	if req.ServiceTier != nil {
		serviceTier := string(*req.ServiceTier)
		contract.ServiceTier = &serviceTier
	}

	// 转换 Tools
	if len(req.Tools) > 0 {
		tools, vendorExtras, err := convertToolsToContract(req.Tools)
		if err != nil {
			return nil, err
		}
		contract.Tools = tools

		// 合并工具的 VendorExtras
		if len(vendorExtras) > 0 {
			if contract.VendorExtras == nil {
				contract.VendorExtras = make(map[string]interface{})
			}
			contract.VendorExtras["tools_extras"] = vendorExtras
		}
	}

	// 转换 ToolChoice
	if req.ToolChoice != nil {
		toolChoice, parallelToolCalls, err := convertToolChoiceToContract(req.ToolChoice)
		if err != nil {
			return nil, err
		}
		contract.ToolChoice = toolChoice
		contract.ParallelToolCalls = parallelToolCalls
	}

	// 转换 Thinking
	if req.Thinking != nil {
		reasoning, err := convertThinkingToContract(req.Thinking)
		if err != nil {
			return nil, err
		}
		contract.Reasoning = reasoning
	}

	// 初始化 VendorExtras
	if contract.VendorExtras == nil {
		contract.VendorExtras = make(map[string]interface{})
	}
	source := types.VendorSourceAnthropic
	contract.VendorExtrasSource = &source

	return contract, nil
}
