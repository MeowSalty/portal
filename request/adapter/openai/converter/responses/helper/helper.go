package helper

import responsesTypes "github.com/MeowSalty/portal/request/adapter/openai/types/responses"

// normalizeStringValue 将指针字符串转换为普通字符串。
func normalizeStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

// GetSequenceNumber 从流事件中获取序列号。
func GetSequenceNumber(event *responsesTypes.StreamEvent) int {
	if event.Created != nil {
		return event.Created.SequenceNumber
	}
	if event.InProgress != nil {
		return event.InProgress.SequenceNumber
	}
	if event.Completed != nil {
		return event.Completed.SequenceNumber
	}
	if event.Failed != nil {
		return event.Failed.SequenceNumber
	}
	if event.Incomplete != nil {
		return event.Incomplete.SequenceNumber
	}
	if event.Queued != nil {
		return event.Queued.SequenceNumber
	}
	if event.OutputItemAdded != nil {
		return event.OutputItemAdded.SequenceNumber
	}
	if event.OutputItemDone != nil {
		return event.OutputItemDone.SequenceNumber
	}
	if event.ContentPartAdded != nil {
		return event.ContentPartAdded.SequenceNumber
	}
	if event.ContentPartDone != nil {
		return event.ContentPartDone.SequenceNumber
	}
	if event.OutputTextDelta != nil {
		return event.OutputTextDelta.SequenceNumber
	}
	if event.OutputTextDone != nil {
		return event.OutputTextDone.SequenceNumber
	}
	if event.OutputTextAnnotationAdded != nil {
		return event.OutputTextAnnotationAdded.SequenceNumber
	}
	if event.RefusalDelta != nil {
		return event.RefusalDelta.SequenceNumber
	}
	if event.RefusalDone != nil {
		return event.RefusalDone.SequenceNumber
	}
	if event.ReasoningTextDelta != nil {
		return event.ReasoningTextDelta.SequenceNumber
	}
	if event.ReasoningTextDone != nil {
		return event.ReasoningTextDone.SequenceNumber
	}
	if event.ReasoningSummaryPartAdded != nil {
		return event.ReasoningSummaryPartAdded.SequenceNumber
	}
	if event.ReasoningSummaryPartDone != nil {
		return event.ReasoningSummaryPartDone.SequenceNumber
	}
	if event.ReasoningSummaryTextDelta != nil {
		return event.ReasoningSummaryTextDelta.SequenceNumber
	}
	if event.ReasoningSummaryTextDone != nil {
		return event.ReasoningSummaryTextDone.SequenceNumber
	}
	if event.FunctionCallArgumentsDelta != nil {
		return event.FunctionCallArgumentsDelta.SequenceNumber
	}
	if event.FunctionCallArgumentsDone != nil {
		return event.FunctionCallArgumentsDone.SequenceNumber
	}
	if event.CustomToolCallInputDelta != nil {
		return event.CustomToolCallInputDelta.SequenceNumber
	}
	if event.CustomToolCallInputDone != nil {
		return event.CustomToolCallInputDone.SequenceNumber
	}
	if event.MCPCallArgumentsDelta != nil {
		return event.MCPCallArgumentsDelta.SequenceNumber
	}
	if event.MCPCallArgumentsDone != nil {
		return event.MCPCallArgumentsDone.SequenceNumber
	}
	if event.MCPCallCompleted != nil {
		return event.MCPCallCompleted.SequenceNumber
	}
	if event.MCPCallFailed != nil {
		return event.MCPCallFailed.SequenceNumber
	}
	if event.MCPCallInProgress != nil {
		return event.MCPCallInProgress.SequenceNumber
	}
	if event.MCPListToolsCompleted != nil {
		return event.MCPListToolsCompleted.SequenceNumber
	}
	if event.MCPListToolsFailed != nil {
		return event.MCPListToolsFailed.SequenceNumber
	}
	if event.MCPListToolsInProgress != nil {
		return event.MCPListToolsInProgress.SequenceNumber
	}
	if event.AudioDelta != nil {
		return event.AudioDelta.SequenceNumber
	}
	if event.AudioDone != nil {
		return event.AudioDone.SequenceNumber
	}
	if event.AudioTranscriptDelta != nil {
		return event.AudioTranscriptDelta.SequenceNumber
	}
	if event.AudioTranscriptDone != nil {
		return event.AudioTranscriptDone.SequenceNumber
	}
	if event.CodeInterpreterCallCodeDelta != nil {
		return event.CodeInterpreterCallCodeDelta.SequenceNumber
	}
	if event.CodeInterpreterCallCodeDone != nil {
		return event.CodeInterpreterCallCodeDone.SequenceNumber
	}
	if event.CodeInterpreterCallCompleted != nil {
		return event.CodeInterpreterCallCompleted.SequenceNumber
	}
	if event.CodeInterpreterCallInProgress != nil {
		return event.CodeInterpreterCallInProgress.SequenceNumber
	}
	if event.CodeInterpreterCallInterpreting != nil {
		return event.CodeInterpreterCallInterpreting.SequenceNumber
	}
	if event.FileSearchCallCompleted != nil {
		return event.FileSearchCallCompleted.SequenceNumber
	}
	if event.FileSearchCallInProgress != nil {
		return event.FileSearchCallInProgress.SequenceNumber
	}
	if event.FileSearchCallSearching != nil {
		return event.FileSearchCallSearching.SequenceNumber
	}
	if event.WebSearchCallCompleted != nil {
		return event.WebSearchCallCompleted.SequenceNumber
	}
	if event.WebSearchCallInProgress != nil {
		return event.WebSearchCallInProgress.SequenceNumber
	}
	if event.WebSearchCallSearching != nil {
		return event.WebSearchCallSearching.SequenceNumber
	}
	if event.ImageGenCallCompleted != nil {
		return event.ImageGenCallCompleted.SequenceNumber
	}
	if event.ImageGenCallGenerating != nil {
		return event.ImageGenCallGenerating.SequenceNumber
	}
	if event.ImageGenCallInProgress != nil {
		return event.ImageGenCallInProgress.SequenceNumber
	}
	if event.ImageGenCallPartialImage != nil {
		return event.ImageGenCallPartialImage.SequenceNumber
	}
	if event.Error != nil {
		return event.Error.SequenceNumber
	}
	return 0
}

// SetSequenceNumber 设置流事件的序列号。
func SetSequenceNumber(event *responsesTypes.StreamEvent, seq int) {
	if event.Created != nil {
		event.Created.SequenceNumber = seq
	}
	if event.InProgress != nil {
		event.InProgress.SequenceNumber = seq
	}
	if event.Completed != nil {
		event.Completed.SequenceNumber = seq
	}
	if event.Failed != nil {
		event.Failed.SequenceNumber = seq
	}
	if event.Incomplete != nil {
		event.Incomplete.SequenceNumber = seq
	}
	if event.Queued != nil {
		event.Queued.SequenceNumber = seq
	}
	if event.OutputItemAdded != nil {
		event.OutputItemAdded.SequenceNumber = seq
	}
	if event.OutputItemDone != nil {
		event.OutputItemDone.SequenceNumber = seq
	}
	if event.ContentPartAdded != nil {
		event.ContentPartAdded.SequenceNumber = seq
	}
	if event.ContentPartDone != nil {
		event.ContentPartDone.SequenceNumber = seq
	}
	if event.OutputTextDelta != nil {
		event.OutputTextDelta.SequenceNumber = seq
	}
	if event.OutputTextDone != nil {
		event.OutputTextDone.SequenceNumber = seq
	}
	if event.OutputTextAnnotationAdded != nil {
		event.OutputTextAnnotationAdded.SequenceNumber = seq
	}
	if event.RefusalDelta != nil {
		event.RefusalDelta.SequenceNumber = seq
	}
	if event.RefusalDone != nil {
		event.RefusalDone.SequenceNumber = seq
	}
	if event.ReasoningTextDelta != nil {
		event.ReasoningTextDelta.SequenceNumber = seq
	}
	if event.ReasoningTextDone != nil {
		event.ReasoningTextDone.SequenceNumber = seq
	}
	if event.ReasoningSummaryPartAdded != nil {
		event.ReasoningSummaryPartAdded.SequenceNumber = seq
	}
	if event.ReasoningSummaryPartDone != nil {
		event.ReasoningSummaryPartDone.SequenceNumber = seq
	}
	if event.ReasoningSummaryTextDelta != nil {
		event.ReasoningSummaryTextDelta.SequenceNumber = seq
	}
	if event.ReasoningSummaryTextDone != nil {
		event.ReasoningSummaryTextDone.SequenceNumber = seq
	}
	if event.FunctionCallArgumentsDelta != nil {
		event.FunctionCallArgumentsDelta.SequenceNumber = seq
	}
	if event.FunctionCallArgumentsDone != nil {
		event.FunctionCallArgumentsDone.SequenceNumber = seq
	}
	if event.CustomToolCallInputDelta != nil {
		event.CustomToolCallInputDelta.SequenceNumber = seq
	}
	if event.CustomToolCallInputDone != nil {
		event.CustomToolCallInputDone.SequenceNumber = seq
	}
	if event.MCPCallArgumentsDelta != nil {
		event.MCPCallArgumentsDelta.SequenceNumber = seq
	}
	if event.MCPCallArgumentsDone != nil {
		event.MCPCallArgumentsDone.SequenceNumber = seq
	}
	if event.MCPCallCompleted != nil {
		event.MCPCallCompleted.SequenceNumber = seq
	}
	if event.MCPCallFailed != nil {
		event.MCPCallFailed.SequenceNumber = seq
	}
	if event.MCPCallInProgress != nil {
		event.MCPCallInProgress.SequenceNumber = seq
	}
	if event.MCPListToolsCompleted != nil {
		event.MCPListToolsCompleted.SequenceNumber = seq
	}
	if event.MCPListToolsFailed != nil {
		event.MCPListToolsFailed.SequenceNumber = seq
	}
	if event.MCPListToolsInProgress != nil {
		event.MCPListToolsInProgress.SequenceNumber = seq
	}
	if event.AudioDelta != nil {
		event.AudioDelta.SequenceNumber = seq
	}
	if event.AudioDone != nil {
		event.AudioDone.SequenceNumber = seq
	}
	if event.AudioTranscriptDelta != nil {
		event.AudioTranscriptDelta.SequenceNumber = seq
	}
	if event.AudioTranscriptDone != nil {
		event.AudioTranscriptDone.SequenceNumber = seq
	}
	if event.CodeInterpreterCallCodeDelta != nil {
		event.CodeInterpreterCallCodeDelta.SequenceNumber = seq
	}
	if event.CodeInterpreterCallCodeDone != nil {
		event.CodeInterpreterCallCodeDone.SequenceNumber = seq
	}
	if event.CodeInterpreterCallCompleted != nil {
		event.CodeInterpreterCallCompleted.SequenceNumber = seq
	}
	if event.CodeInterpreterCallInProgress != nil {
		event.CodeInterpreterCallInProgress.SequenceNumber = seq
	}
	if event.CodeInterpreterCallInterpreting != nil {
		event.CodeInterpreterCallInterpreting.SequenceNumber = seq
	}
	if event.FileSearchCallCompleted != nil {
		event.FileSearchCallCompleted.SequenceNumber = seq
	}
	if event.FileSearchCallInProgress != nil {
		event.FileSearchCallInProgress.SequenceNumber = seq
	}
	if event.FileSearchCallSearching != nil {
		event.FileSearchCallSearching.SequenceNumber = seq
	}
	if event.WebSearchCallCompleted != nil {
		event.WebSearchCallCompleted.SequenceNumber = seq
	}
	if event.WebSearchCallInProgress != nil {
		event.WebSearchCallInProgress.SequenceNumber = seq
	}
	if event.WebSearchCallSearching != nil {
		event.WebSearchCallSearching.SequenceNumber = seq
	}
	if event.ImageGenCallCompleted != nil {
		event.ImageGenCallCompleted.SequenceNumber = seq
	}
	if event.ImageGenCallGenerating != nil {
		event.ImageGenCallGenerating.SequenceNumber = seq
	}
	if event.ImageGenCallInProgress != nil {
		event.ImageGenCallInProgress.SequenceNumber = seq
	}
	if event.ImageGenCallPartialImage != nil {
		event.ImageGenCallPartialImage.SequenceNumber = seq
	}
	if event.Error != nil {
		event.Error.SequenceNumber = seq
	}
}
