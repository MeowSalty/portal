package types

// StreamEvent 表示 Gemini API 的 streamGenerateContent 流式响应块。
// 文档中流式块与非流式响应结构一致，均为 GenerateContentResponse。
// 使用类型别名确保流式与非流式响应严格对齐。
type StreamEvent = Response
