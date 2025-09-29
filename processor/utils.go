package processor

import (
	"fmt"
	"time"

	"github.com/MeowSalty/portal/types"
	"github.com/valyala/fasthttp"
)

// 辅助函数和工具方法

// checkResponseError 检查响应中的错误
func (p *RequestProcessor) checkResponseError(response *types.Response) error {
	for _, choice := range response.Choices {
		if choice.Error != nil && choice.Error.Code != fasthttp.StatusOK {
			message := sanitizeErrorMessage(choice.Error.Message)
			return fmt.Errorf("stream error: code=%d, message=%s", choice.Error.Code, message)
		}
	}
	return nil
}

// getAdapter 获取适配器
func (p *RequestProcessor) getAdapter(format string) (types.Adapter, error) {
	adapter, ok := p.Adapters[format]
	if !ok {
		return nil, fmt.Errorf("适配器未找到：%s", format)
	}
	return adapter, nil
}

// createRetryContext 创建重试上下文
func (p *RequestProcessor) createRetryContext(channels []*types.Channel, startTime time.Time) *retryContext {
	ctx := &retryContext{
		allChannels:  channels,
		startTime:    startTime,
		channelIndex: make(map[string]int, len(channels)),
	}

	// 构建索引映射以优化移除操作
	for i, ch := range channels {
		key := fmt.Sprintf("%d_%d_%d", ch.Platform.ID, ch.Model.ID, ch.APIKey.ID)
		ctx.channelIndex[key] = i
	}

	return ctx
}

// removeChannel 移除通道
func (p *RequestProcessor) removeChannel(retryCtx *retryContext, target *types.Channel) []*types.Channel {
	key := fmt.Sprintf("%d_%d_%d", target.Platform.ID, target.Model.ID, target.APIKey.ID)

	if idx, exists := retryCtx.channelIndex[key]; exists {
		// 使用最后一个元素替换要删除的元素
		lastIdx := len(retryCtx.allChannels) - 1
		if idx != lastIdx {
			retryCtx.allChannels[idx] = retryCtx.allChannels[lastIdx]
			// 更新索引
			lastKey := fmt.Sprintf("%d_%d_%d",
				retryCtx.allChannels[idx].Platform.ID,
				retryCtx.allChannels[idx].Model.ID,
				retryCtx.allChannels[idx].APIKey.ID)
			retryCtx.channelIndex[lastKey] = idx
		}

		// 删除最后一个元素和索引
		retryCtx.allChannels = retryCtx.allChannels[:lastIdx]
		delete(retryCtx.channelIndex, key)
	}

	return retryCtx.allChannels
}

// isHTMLContent 检查错误信息是否包含 HTML 页面内容
func isHTMLContent(errorMsg string) bool {
	return len(errorMsg) > 15 && (errorMsg[:15] == "<!DOCTYPE html>" || errorMsg[:5] == "<html")
}

// sanitizeError 检查错误中是否包含 HTML 页面内容，如果包含，则屏蔽 HTML 内容并返回
func sanitizeError(error error) error {
	if isHTMLContent(error.Error()) {
		return fmt.Errorf("[HTML content filtered]")
	}
	return error
}

// sanitizeErrorMessage 检查错误信息中是否包含 HTML 页面内容，如果包含，则屏蔽 HTML 内容并返回
func sanitizeErrorMessage(errorMsg string) string {
	if isHTMLContent(errorMsg) {
		return "[HTML content filtered]"
	}
	return errorMsg
}
