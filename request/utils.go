package request

import (
	"fmt"

	"github.com/MeowSalty/portal/request/adapter"
	"github.com/MeowSalty/portal/types"
	"github.com/valyala/fasthttp"
)

// 辅助函数和工具方法

// checkResponseError 检查响应中的错误
func (p *Request) checkResponseError(response *types.Response) error {
	for _, choice := range response.Choices {
		if choice.Error != nil && choice.Error.Code != fasthttp.StatusOK {
			return fmt.Errorf("stream error: code=%d, message=%s", choice.Error.Code, choice.Error.Message)
		}
	}
	return nil
}

// getAdapter 获取适配器
func (p *Request) getAdapter(format string) (*adapter.Adapter, error) {
	adapter, err := adapter.GetAdapter(format)
	if err != nil {
		return nil, fmt.Errorf("适配器未找到：%s", format)
	}
	return adapter, nil
}
