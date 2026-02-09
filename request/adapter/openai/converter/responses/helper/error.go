package helper

import (
	responsesTypes "github.com/MeowSalty/portal/request/adapter/openai/types/responses"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// ConvertErrorToContract 将 OpenAI Responses 错误转换为统一错误。
func ConvertErrorToContract(respErr *responsesTypes.ResponseError) *types.ResponseError {
	if respErr == nil {
		return nil
	}

	code := respErr.Code
	message := respErr.Message
	return &types.ResponseError{
		Code:    &code,
		Message: &message,
		Type:    respErr.Type,
		Param:   respErr.Param,
		Extras:  make(map[string]interface{}),
	}
}

// ConvertErrorFromContract 将统一错误转换为 OpenAI Responses 错误。
func ConvertErrorFromContract(contractErr *types.ResponseError) *responsesTypes.ResponseError {
	if contractErr == nil {
		return nil
	}

	respErr := &responsesTypes.ResponseError{}

	if contractErr.Code != nil {
		respErr.Code = *contractErr.Code
	}
	if contractErr.Message != nil {
		respErr.Message = *contractErr.Message
	}
	respErr.Type = contractErr.Type
	respErr.Param = contractErr.Param

	return respErr
}

// ConvertStreamErrorToResponseError 将流错误转换为 ResponseError。
func ConvertStreamErrorToResponseError(err *types.StreamErrorPayload) *responsesTypes.ResponseError {
	if err == nil {
		return nil
	}

	return &responsesTypes.ResponseError{
		Code:    err.Code,
		Message: err.Message,
		Type:    &err.Type,
		Param:   &err.Param,
	}
}

// ConvertResponseErrorToStreamError 将 ResponseError 转换为流错误。
func ConvertResponseErrorToStreamError(err *responsesTypes.ResponseError) *types.StreamErrorPayload {
	if err == nil {
		return nil
	}

	result := &types.StreamErrorPayload{
		Message: err.Message,
		Code:    err.Code,
		Raw:     make(map[string]interface{}),
	}

	if err.Type != nil {
		result.Type = *err.Type
	}
	if err.Param != nil {
		result.Param = *err.Param
	}

	return result
}
