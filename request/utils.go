package request

import (
	"strconv"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/request/adapter"
	"github.com/MeowSalty/portal/request/adapter/types"
	"github.com/valyala/fasthttp"
)

// 辅助函数和工具方法

// checkResponseError 检查响应中的错误
func (p *Request) checkResponseError(response *types.StreamEventContract) error {
	log := p.logger

	if response == nil {
		log.Error("响应为空")
		return errors.ErrEmptyResponse.WithContext("error_from", "upstream")
	}

	if response.Type == types.StreamEventError && response.Error == nil {
		log.Error("响应错误事件缺少错误信息")
		return errors.NewWithHTTPStatus(errors.ErrCodeStreamError, "流处理错误", fasthttp.StatusInternalServerError).
			WithContext("error_from", "upstream")
	}

	if response.Error != nil {
		statusCode := fasthttp.StatusInternalServerError
		if response.Error.Code != "" {
			if parsed, err := strconv.Atoi(response.Error.Code); err == nil {
				statusCode = parsed
			}
		}

		log.Error("响应中检测到错误",
			"error_type", response.Error.Type,
			"error_code", response.Error.Code,
			"error_message", response.Error.Message,
		)
		return errors.NewWithHTTPStatus(errors.ErrCodeStreamError, "流处理错误", statusCode).
			WithContext("error_from", "upstream").
			WithContext("error_type", response.Error.Type).
			WithContext("error_code", response.Error.Code).
			WithContext("error_message", response.Error.Message)
	}

	log.Debug("响应检查通过", "event_type", response.Type)
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
