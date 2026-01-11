package converter

import (
	openaiChatConverter "github.com/MeowSalty/portal/request/adapter/openai/converter/chat"
	openaiResponsesConverter "github.com/MeowSalty/portal/request/adapter/openai/converter/responses"
	openaiChat "github.com/MeowSalty/portal/request/adapter/openai/types/chat"
	openaiResponses "github.com/MeowSalty/portal/request/adapter/openai/types/responses"
	"github.com/MeowSalty/portal/routing"
	coreTypes "github.com/MeowSalty/portal/types"
)

// Deprecated: 请使用 request/adapter/openai/converter/chat 包。
func ConvertRequest(request *coreTypes.Request, channel *routing.Channel) interface{} {
	return openaiChatConverter.ConvertRequest(request, channel)
}

// Deprecated: 请使用 request/adapter/openai/converter/chat 包。
func ConvertCoreRequest(openaiReq *openaiChat.Request) *coreTypes.Request {
	return openaiChatConverter.ConvertCoreRequest(openaiReq)
}

// Deprecated: 请使用 request/adapter/openai/converter/chat 包。
func ConvertCoreResponse(openaiResp *openaiChat.Response) *coreTypes.Response {
	return openaiChatConverter.ConvertCoreResponse(openaiResp)
}

// Deprecated: 请使用 request/adapter/openai/converter/chat 包。
func ConvertResponse(resp *coreTypes.Response) *openaiChat.Response {
	return openaiChatConverter.ConvertResponse(resp)
}

// Deprecated: 请使用 request/adapter/openai/converter/responses 包。
func ConvertResponsesRequest(request *coreTypes.Request, channel *routing.Channel) interface{} {
	return openaiResponsesConverter.ConvertRequest(request, channel)
}

// Deprecated: 请使用 request/adapter/openai/converter/responses 包。
func ConvertResponsesCoreResponse(resp *openaiResponses.Response) *coreTypes.Response {
	return openaiResponsesConverter.ConvertCoreResponse(resp)
}

// Deprecated: 请使用 request/adapter/openai/converter/responses 包。
func ConvertResponsesStreamEvent(event *openaiResponses.ResponsesStreamEvent) *coreTypes.Response {
	return openaiResponsesConverter.ConvertStreamEvent(event)
}
