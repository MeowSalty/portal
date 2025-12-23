package request

import (
	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/request/adapter"
	"github.com/MeowSalty/portal/types"
	"github.com/valyala/fasthttp"
)

// 辅助函数和工具方法

// checkResponseError 检查响应中的错误
func (p *Request) checkResponseError(response *types.Response) error {
	log := p.logger

	// 检查 Choices 是否为空
	if len(response.Choices) == 0 {
		log.Error("响应中 Choices 为空")
		return errors.ErrEmptyResponse.WithContext("error_from", "upstream").WithContext("response", response)
	}

	for i, choice := range response.Choices {
		if choice.Error != nil && choice.Error.Code != fasthttp.StatusOK {
			log.Error("响应中检测到错误",
				"choice_index", i,
				"error_code", choice.Error.Code,
				"error_message", choice.Error.Message,
			)
			return errors.NewWithHTTPStatus(errors.ErrCodeStreamError, "流处理错误", choice.Error.Code).
				WithContext("error_from", "upstream").
				WithContext("choice_index", i).
				WithContext("error_message", choice.Error.Message)
		}
	}

	log.Debug("响应检查通过", "choices_count", len(response.Choices))
	return nil
}

// getAdapter 获取适配器
func (p *Request) getAdapter(format string) (*adapter.Adapter, error) {
	log := p.logger

	log.Debug("获取适配器", "format", format)

	adapter, err := adapter.GetAdapter(format)
	if err != nil {
		log.Error("适配器未找到",
			"format", format,
			"error", err,
		)
		return nil, errors.ErrAdapterNotFound.WithContext("format", format).WithCause(err)
	}

	log.Debug("适配器获取成功",
		"format", format,
		"adapter_name", adapter.Name(),
	)
	return adapter, nil
}
