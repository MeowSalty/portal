package converter

import (
	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/logger"
	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// ResponseFromContract 将 ResponseContract 转换回 Anthropic 响应。
func ResponseFromContract(contract *types.ResponseContract, log logger.Logger) (*anthropicTypes.Response, error) {
	if contract == nil {
		return nil, nil
	}

	if log == nil {
		log = logger.NewNopLogger()
	}

	resp := &anthropicTypes.Response{
		ID:   contract.ID,
		Type: anthropicTypes.ResponseTypeMessage,
		Role: anthropicTypes.RoleAssistant,
	}

	// 转换 Model
	if contract.Model != nil {
		resp.Model = *contract.Model
	}

	// 从 Extras 恢复 StopSequence
	if stopSeq, ok := getStringExtra("anthropic.stop_sequence", contract.Extras); ok {
		resp.StopSequence = &stopSeq
	}

	// 转换 Usage
	if contract.Usage != nil {
		usage, err := convertUsageFromContract(contract.Usage)
		if err != nil {
			log.Error("转换 Usage 失败", "error", err)
			return nil, errors.Wrap(errors.ErrCodeInternal, "转换 Usage 失败", err)
		}
		resp.Usage = usage
	}

	// 从第一个 Choice 恢复响应内容
	if len(contract.Choices) > 0 {
		choice := contract.Choices[0]

		// 转换 FinishReason
		if choice.FinishReason != nil {
			stopReason := mapFinishReasonToStopReason(*choice.FinishReason)
			resp.StopReason = &stopReason
		}

		// 转换 Message
		if choice.Message != nil {
			content, err := convertMessageToResponseContent(choice.Message, log)
			if err != nil {
				log.Error("转换 Message 失败", "error", err)
				return nil, errors.Wrap(errors.ErrCodeInternal, "转换 Message 失败", err)
			}
			resp.Content = content
		}
	}

	return resp, nil
}

// ResponseErrorFromContract 将 ResponseContract 错误转换回 Anthropic 错误响应。
func ResponseErrorFromContract(contract *types.ResponseContract) (*anthropicTypes.ErrorResponse, error) {
	if contract == nil || contract.Error == nil {
		return nil, nil
	}

	errResp := &anthropicTypes.ErrorResponse{
		Type:  "error",
		Error: anthropicTypes.Error{},
	}

	if contract.Error.Type != nil {
		errResp.Error.Type = *contract.Error.Type
	}
	if contract.Error.Message != nil {
		errResp.Error.Message = *contract.Error.Message
	}

	return errResp, nil
}
