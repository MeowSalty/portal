package request

import "context"

type RequestLogRepository interface {
	CreateRequestLog(ctx context.Context, log *RequestLog) error
}
