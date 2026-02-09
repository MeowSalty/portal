package converter

import (
	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/logger"
	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// ResponseToContract 将 Anthropic 响应转换为统一的 ResponseContract。
func ResponseToContract(resp *anthropicTypes.Response, log logger.Logger) (*types.ResponseContract, error) {
	if resp == nil {
		return nil, nil
	}

	if log == nil {
		log = logger.NewNopLogger()
	}

	contract := &types.ResponseContract{
		Source: types.VendorSourceAnthropic,
		ID:     resp.ID,
		Extras: make(map[string]interface{}),
	}

	// 转换顶层字段
	objType := string(resp.Type)
	contract.Object = &objType
	contract.Model = &resp.Model

	// 转换 StopSequence 到 Extras
	if resp.StopSequence != nil {
		if err := SaveVendorExtra("anthropic.stop_sequence", *resp.StopSequence, contract.Extras); err != nil {
			log.Warn("保存 StopSequence 失败", "error", err)
		}
	}

	// 转换 Usage
	if resp.Usage != nil {
		usage, err := convertUsageToContract(resp.Usage)
		if err != nil {
			log.Error("转换 Usage 失败", "error", err)
			return nil, errors.Wrap(errors.ErrCodeInternal, "转换 Usage 失败", err)
		}
		contract.Usage = usage
	}

	// 创建单个 Choice（Anthropic 只有一个响应）
	choice := types.ResponseChoice{
		Extras: make(map[string]interface{}),
	}

	// 转换 FinishReason
	if resp.StopReason != nil {
		finishReason := mapStopReasonToFinishReason(*resp.StopReason)
		choice.FinishReason = &finishReason
		choice.NativeFinishReason = resp.StopReason
	}

	// 转换 Message
	message, err := convertResponseToMessage(resp, log)
	if err != nil {
		log.Error("转换 Message 失败", "error", err)
		return nil, errors.Wrap(errors.ErrCodeInternal, "转换 Message 失败", err)
	}
	choice.Message = message

	contract.Choices = []types.ResponseChoice{choice}

	return contract, nil
}

// ResponseErrorToContract 将 Anthropic 错误响应转换为 ResponseContract。
func ResponseErrorToContract(errResp *anthropicTypes.ErrorResponse) (*types.ResponseContract, error) {
	if errResp == nil {
		return nil, nil
	}

	contract := &types.ResponseContract{
		Source: types.VendorSourceAnthropic,
		Extras: make(map[string]interface{}),
	}

	// 转换错误信息
	errType := errResp.Error.Type
	errMsg := errResp.Error.Message
	contract.Error = &types.ResponseError{
		Code:    &errType,
		Type:    &errType,
		Message: &errMsg,
		Extras:  make(map[string]interface{}),
	}

	return contract, nil
}
