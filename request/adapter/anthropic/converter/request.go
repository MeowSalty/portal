package converter

import (
	"strings"

	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
	"github.com/MeowSalty/portal/routing"
	coreTypes "github.com/MeowSalty/portal/types"
)

// ConvertRequest 将核心请求转换为 Anthropic 请求
func ConvertRequest(request *coreTypes.Request, channel *routing.Channel) *anthropicTypes.Request {
	anthropicReq := &anthropicTypes.Request{
		Model:     request.Model,
		Messages:  convertMessages(request.Messages),
		MaxTokens: getMaxTokens(request.MaxTokens),
		Stream:    request.Stream,
	}

	// 设置可选参数
	if request.Temperature != nil {
		anthropicReq.Temperature = request.Temperature
	}

	if request.TopP != nil {
		anthropicReq.TopP = request.TopP
	}

	if request.TopK != nil {
		anthropicReq.TopK = request.TopK
	}

	// 转换停止序列
	if request.Stop.StringValue != nil {
		anthropicReq.StopSequences = []string{*request.Stop.StringValue}
	} else if request.Stop.StringArray != nil {
		anthropicReq.StopSequences = request.Stop.StringArray
	}

	// 提取系统消息
	anthropicReq.System, anthropicReq.Messages = extractSystemMessage(anthropicReq.Messages)

	// 转换工具
	if len(request.Tools) > 0 {
		anthropicReq.Tools = convertTools(request.Tools)
		anthropicReq.ToolChoice = convertToolChoice(request.ToolChoice)
	}

	// 添加用户元数据
	if request.User != nil {
		anthropicReq.Metadata = &anthropicTypes.Metadata{
			UserID: request.User,
		}
	}

	return anthropicReq
}

// convertMessages 转换消息列表
func convertMessages(messages []coreTypes.Message) []anthropicTypes.InputMessage {
	result := make([]anthropicTypes.InputMessage, 0, len(messages))

	for _, msg := range messages {
		// 跳过系统消息，将在后面单独处理
		if msg.Role == "system" {
			continue
		}

		anthropicMsg := anthropicTypes.InputMessage{
			Role: msg.Role,
		}

		// 转换消息内容
		if msg.Content.StringValue != nil {
			anthropicMsg.Content = *msg.Content.StringValue
		} else if msg.Content.ContentParts != nil {
			anthropicMsg.Content = convertContentParts(msg.Content.ContentParts)
		}

		result = append(result, anthropicMsg)
	}

	return result
}

// convertContentParts 转换内容部分
func convertContentParts(parts []coreTypes.ContentPart) []anthropicTypes.ContentBlock {
	blocks := make([]anthropicTypes.ContentBlock, 0, len(parts))

	for _, part := range parts {
		block := anthropicTypes.ContentBlock{
			Type: part.Type,
		}

		if part.Type == "text" && part.Text != nil {
			block.Text = part.Text
		} else if part.Type == "image_url" && part.ImageURL != nil {
			// 转换图像 URL 为 Anthropic 格式
			block.Type = "image"
			block.Source = convertImageURL(part.ImageURL)
		}

		blocks = append(blocks, block)
	}

	return blocks
}

// convertImageURL 转换图像 URL
func convertImageURL(imageURL *coreTypes.ImageURL) *anthropicTypes.ImageSource {
	// 检查是否是 base64 编码
	if strings.HasPrefix(imageURL.URL, "data:") {
		// 解析 data URL: data:image/png;base64,xxxxx
		parts := strings.SplitN(imageURL.URL, ",", 2)
		if len(parts) == 2 {
			// 提取 media type
			mediaParts := strings.Split(parts[0], ";")
			mediaType := strings.TrimPrefix(mediaParts[0], "data:")

			return &anthropicTypes.ImageSource{
				Type:      "base64",
				MediaType: mediaType,
				Data:      parts[1],
			}
		}
	}

	// 对于普通 URL,Anthropic 不直接支持，需要先下载转换为 base64
	// 这里返回 nil，实际使用时应该先处理
	return nil
}

// extractSystemMessage 提取系统消息
func extractSystemMessage(messages []anthropicTypes.InputMessage) (interface{}, []anthropicTypes.InputMessage) {
	var systemContent string
	filtered := make([]anthropicTypes.InputMessage, 0, len(messages))

	for _, msg := range messages {
		if msg.Role == "system" {
			// 提取系统消息内容
			if str, ok := msg.Content.(string); ok {
				if systemContent != "" {
					systemContent += "\n\n"
				}
				systemContent += str
			}
		} else {
			filtered = append(filtered, msg)
		}
	}

	if systemContent != "" {
		return systemContent, filtered
	}

	return nil, filtered
}

// convertTools 转换工具定义
func convertTools(tools []coreTypes.Tool) []anthropicTypes.Tool {
	anthropicTools := make([]anthropicTypes.Tool, 0, len(tools))

	for _, tool := range tools {
		anthropicTool := anthropicTypes.Tool{
			Name:        tool.Function.Name,
			Description: tool.Function.Description,
			InputSchema: convertInputSchema(tool.Function.Parameters),
		}

		anthropicTools = append(anthropicTools, anthropicTool)
	}

	return anthropicTools
}

// convertInputSchema 转换输入 schema
func convertInputSchema(params interface{}) anthropicTypes.InputSchema {
	schema := anthropicTypes.InputSchema{
		Type: "object",
	}

	// 尝试解析参数
	if paramsMap, ok := params.(map[string]interface{}); ok {
		if props, ok := paramsMap["properties"].(map[string]interface{}); ok {
			schema.Properties = props
		}
		if required, ok := paramsMap["required"].([]interface{}); ok {
			schema.Required = make([]string, 0, len(required))
			for _, r := range required {
				if str, ok := r.(string); ok {
					schema.Required = append(schema.Required, str)
				}
			}
		}
	}

	return schema
}

// convertToolChoice 转换工具选择
func convertToolChoice(toolChoice coreTypes.ToolChoice) interface{} {
	if toolChoice.Mode != nil {
		switch *toolChoice.Mode {
		case "auto":
			return anthropicTypes.ToolChoiceAuto{Type: "auto"}
		case "none":
			// Anthropic 没有 "none" 选项，返回 nil 表示不使用工具
			return nil
		}
	}

	if toolChoice.Function != nil {
		return anthropicTypes.ToolChoiceTool{
			Type: "tool",
			Name: toolChoice.Function.Function.Name,
		}
	}

	// 默认为 auto
	return anthropicTypes.ToolChoiceAuto{Type: "auto"}
}

// getMaxTokens 获取最大 token 数，如果未设置则返回默认值
func getMaxTokens(maxTokens *int) int {
	if maxTokens != nil {
		return *maxTokens
	}
	// Anthropic 要求必须设置 max_tokens，默认返回 1024
	return 1024
}
