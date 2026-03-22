package request

import (
	"context"
	"encoding/json"
	stdErrors "errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	portalErrors "github.com/MeowSalty/portal/errors"
)

const (
	// requestLogLongFieldMaxLength 是请求日志中长文本字段的最大长度。
	requestLogLongFieldMaxLength = 4000
)

var (
	upstreamRequestIDRegex = regexp.MustCompile(`(?i)\(\s*request id:\s*([^\)]+)\s*\)`)
)

// RequestLog 表示单个请求的统计信息
type RequestLog struct {
	ID uint `json:"id"` // 唯一标识符

	// 请求基本信息
	Timestamp         time.Time `json:"timestamp"`                     // 请求时间
	ModelName         string    `json:"model_name"`                    // 模型名称
	OriginalModelName string    `json:"original_model_name,omitempty"` // 原始模型名称（用户请求中的模型名称）
	IsStream          bool      `json:"is_stream"`
	IsNative          bool      `json:"is_native"`

	// 通道信息
	PlatformID uint `json:"platform_id"` // 平台 ID
	APIKeyID   uint `json:"api_key_id"`  // 密钥 ID
	ModelID    uint `json:"model_id"`    // 模型 ID

	// 耗时信息
	Duration      time.Duration  `json:"duration"`                  // 总用时
	FirstByteTime *time.Duration `json:"first_byte_time,omitempty"` // 首字用时（仅流式）

	// 结果状态
	Success bool `json:"success"` // 是否成功

	// Deprecated: ErrorMsg 为展示型错误信息（失败时），后续将逐步下线；
	// 请优先使用结构化错误字段（如 ErrorCode/ErrorLevel/HTTPStatus/ErrorFrom 及上游错误字段）。
	ErrorMsg *string `json:"error_msg,omitempty"`

	// CauseMessage 为底层原因文本，用于排障与审计；非稳定展示字段。
	CauseMessage *string `json:"cause_message,omitempty"`

	// 结构化错误字段（建议前端优先消费）。
	ErrorCode  *string `json:"error_code,omitempty"`
	ErrorLevel *string `json:"error_level,omitempty"`
	HTTPStatus *int    `json:"http_status,omitempty"`
	ErrorFrom  *string `json:"error_from,omitempty"`

	// 上游错误字段（若能从 response_body 解析到）。
	UpstreamErrorType    *string `json:"upstream_error_type,omitempty"`
	UpstreamErrorCode    *string `json:"upstream_error_code,omitempty"`
	UpstreamErrorParam   *string `json:"upstream_error_param,omitempty"`
	UpstreamErrorMessage *string `json:"upstream_error_message,omitempty"`
	UpstreamRequestID    *string `json:"upstream_request_id,omitempty"`

	// response_body 解析状态与兜底。
	ResponseBodyIsJSON *bool   `json:"response_body_is_json,omitempty"`
	ResponseBodyRaw    *string `json:"response_body_raw,omitempty"`

	// Token 使用统计
	PromptTokens     *int `json:"prompt_tokens"`     // 提示 Token 数
	CompletionTokens *int `json:"completion_tokens"` // 完成 Token 数
	TotalTokens      *int `json:"total_tokens"`      // 总 Token 数

	// 以下字段仅用于运行时日志上下文，不持久化到存储。
	errorClassifyExplain      string
	errorClassifyMatchedRules string
}

// recordRequestLog 记录请求统计信息
func (p *Request) recordRequestLog(
	requestLog *RequestLog,
	firstByteTime *time.Time,
	success bool,
) {
	// 创建带有请求上下文的日志记录器
	log := p.logger.With(
		"platform_id", requestLog.PlatformID,
		"model_id", requestLog.ModelID,
		"api_key_id", requestLog.APIKeyID,
		"is_stream", requestLog.IsStream,
		"is_native", requestLog.IsNative,
		"model_name", requestLog.ModelName,
	)

	// 计算耗时
	requestDuration := time.Since(requestLog.Timestamp)
	requestLog.Duration = requestDuration
	requestLog.Success = success

	debugArgs := []any{
		"duration", requestDuration.String(),
		"success", success,
	}

	// 如果记录了首字节时间，则计算首字节耗时
	if firstByteTime != nil && !firstByteTime.IsZero() {
		firstByteDuration := firstByteTime.Sub(requestLog.Timestamp)
		requestLog.FirstByteTime = &firstByteDuration
		debugArgs = append(debugArgs, "first_byte_time", firstByteDuration.String())
	}

	// 将运行态耗时与 Token 统计收敛为结束摘要（调试级别）
	if requestLog.PromptTokens != nil && requestLog.CompletionTokens != nil && requestLog.TotalTokens != nil {
		debugArgs = append(debugArgs,
			"prompt_tokens", *requestLog.PromptTokens,
			"completion_tokens", *requestLog.CompletionTokens,
			"total_tokens", *requestLog.TotalTokens,
		)
	}
	if requestLog.errorClassifyExplain != "" {
		debugArgs = append(debugArgs, "error_classify_explain", requestLog.errorClassifyExplain)
	}
	if requestLog.errorClassifyMatchedRules != "" {
		debugArgs = append(debugArgs, "error_classify_matched_rules", requestLog.errorClassifyMatchedRules)
	}
	log.Debug("请求结束摘要", debugArgs...)

	// 保存到数据库
	err := p.repo.CreateRequestLog(context.Background(), requestLog)
	if err != nil {
		// 审计层仅记录自身持久化异常，不重复输出业务请求失败。
		log.Error("audit_log_persist_failed", "error", err)
	}
}

// fillRequestLogErrorFields 将 error 中的关键信息填充到 RequestLog 的结构化错误字段。
//
// 说明：
// 1. 始终保留 ErrorMsg 作为展示/排障文本。
// 2. 优先从 response_body(JSON) 解析上游错误字段。
// 3. response_body 解析失败时，将原文写入 ResponseBodyRaw。
func fillRequestLogErrorFields(log *RequestLog, err error) {
	if log == nil || err == nil {
		return
	}

	errMsg := err.Error()
	log.ErrorMsg = &errMsg

	if causeMsg := extractCauseMessage(err); causeMsg != "" && causeMsg != errMsg {
		causeMsg = clipLongField(causeMsg)
		log.CauseMessage = &causeMsg
	}

	var portalErr *portalErrors.Error
	if !portalErrors.As(err, &portalErr) {
		return
	}

	if portalErr.Code != "" {
		code := string(portalErr.Code)
		log.ErrorCode = &code
	}

	if portalErr.HTTPStatus != nil {
		status := *portalErr.HTTPStatus
		log.HTTPStatus = &status
	}

	if level := errorLevelToString(portalErrors.GetErrorLevel(err)); level != "" {
		log.ErrorLevel = &level
	}

	ctx := portalErr.Context
	if len(ctx) == 0 {
		return
	}

	if s, ok := contextValueToString(ctx["error_from"]); ok {
		log.ErrorFrom = &s
	}

	classifierInput := portalErrors.ClassifierInput{
		Code:    portalErr.Code,
		Message: portalErr.Message,
	}
	if classifierInput.Message == "" {
		classifierInput.Message = errMsg
	}

	if portalErr.HTTPStatus != nil {
		classifierInput.HTTPStatus = *portalErr.HTTPStatus
		classifierInput.HTTPResponseReceived = true
	}

	if errorFrom := portalErrors.GetErrorFrom(err); errorFrom != "" {
		classifierInput.ErrorFrom = errorFrom
	}

	if log.CauseMessage != nil {
		classifierInput.CauseMessage = *log.CauseMessage
	}

	// response_body 之外的上下文字段可作为兜底来源。
	if s, ok := contextValueToString(ctx["error_type"]); ok {
		log.UpstreamErrorType = &s
		classifierInput.ErrorType = s
	}
	if s, ok := contextValueToString(ctx["error_code"]); ok {
		log.UpstreamErrorCode = &s
		classifierInput.VendorCode = s
	}
	if s, ok := contextValueToString(ctx["error_message"]); ok {
		s = clipLongField(s)
		log.UpstreamErrorMessage = &s
		classifierInput.ErrorMessage = s
		if requestID := extractUpstreamRequestID(s); requestID != "" {
			log.UpstreamRequestID = &requestID
		}
	}

	responseBody, ok := contextValueToString(ctx["response_body"])
	if !ok {
		fillRequestLogClassificationSummary(log, classifierInput)
		return
	}

	classifierInput.ResponseBody = responseBody
	upstreamFields := parseAndFillResponseBody(log, responseBody)
	if upstreamFields.ErrorType != "" {
		classifierInput.ErrorType = upstreamFields.ErrorType
	}
	if upstreamFields.ErrorCode != "" {
		classifierInput.VendorCode = upstreamFields.ErrorCode
	}
	if upstreamFields.ErrorMessage != "" {
		classifierInput.ErrorMessage = upstreamFields.ErrorMessage
	}

	fillRequestLogClassificationSummary(log, classifierInput)
}

type requestLogUpstreamFields struct {
	ErrorType    string
	ErrorCode    string
	ErrorMessage string
}

func fillRequestLogClassificationSummary(log *RequestLog, input portalErrors.ClassifierInput) {
	if log == nil {
		return
	}

	result := portalErrors.ClassifyError(input)
	if log.ErrorFrom == nil {
		source := string(result.Source.Value)
		if source != "" {
			log.ErrorFrom = &source
		}
	}

	log.errorClassifyExplain = strings.TrimSpace(result.Explain)
	if len(result.MatchedRules) == 0 {
		log.errorClassifyMatchedRules = ""
		return
	}

	log.errorClassifyMatchedRules = strings.Join(result.MatchedRules, ",")
}

// extractCauseMessage 获取错误链最底层 cause 的文本。
func extractCauseMessage(err error) string {
	if err == nil {
		return ""
	}

	cause := err
	for {
		next := stdErrors.Unwrap(cause)
		if next == nil {
			break
		}
		cause = next
	}

	if cause == err {
		return ""
	}

	return cause.Error()
}

// parseAndFillResponseBody 解析 response_body 并填充上游错误字段。
func parseAndFillResponseBody(log *RequestLog, responseBody string) requestLogUpstreamFields {
	upstreamFields := requestLogUpstreamFields{}

	responseBody = strings.TrimSpace(responseBody)
	if responseBody == "" {
		return upstreamFields
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(responseBody), &payload); err != nil {
		isJSON := false
		log.ResponseBodyIsJSON = &isJSON
		raw := clipLongField(responseBody)
		log.ResponseBodyRaw = &raw
		return upstreamFields
	}

	isJSON := true
	log.ResponseBodyIsJSON = &isJSON

	extracted := false
	if errValue, ok := payload["error"]; ok {
		switch errObj := errValue.(type) {
		case map[string]any:
			if s, ok := contextValueToString(errObj["type"]); ok {
				log.UpstreamErrorType = &s
				upstreamFields.ErrorType = s
				extracted = true
			}
			if s, ok := contextValueToString(errObj["code"]); ok {
				log.UpstreamErrorCode = &s
				upstreamFields.ErrorCode = s
				extracted = true
			}
			if s, ok := contextValueToString(errObj["param"]); ok {
				log.UpstreamErrorParam = &s
				extracted = true
			}
			if s, ok := contextValueToString(errObj["message"]); ok {
				s = clipLongField(s)
				log.UpstreamErrorMessage = &s
				upstreamFields.ErrorMessage = s
				extracted = true
				if requestID := extractUpstreamRequestID(s); requestID != "" {
					log.UpstreamRequestID = &requestID
				}
			}
		case string:
			if s, ok := contextValueToString(errObj); ok {
				s = clipLongField(s)
				log.UpstreamErrorMessage = &s
				upstreamFields.ErrorMessage = s
				extracted = true
				if requestID := extractUpstreamRequestID(s); requestID != "" {
					log.UpstreamRequestID = &requestID
				}
			}
		}
	}

	if extracted {
		log.ResponseBodyRaw = nil
		return upstreamFields
	}

	raw := clipLongField(responseBody)
	log.ResponseBodyRaw = &raw
	return upstreamFields
}

// errorLevelToString 将错误层级枚举转换为对前端稳定的字符串。
func errorLevelToString(level portalErrors.ErrorLevel) string {
	switch level {
	case portalErrors.ErrorLevelPlatform:
		return "platform"
	case portalErrors.ErrorLevelKey:
		return "key"
	case portalErrors.ErrorLevelModel:
		return "model"
	default:
		return ""
	}
}

// contextValueToString 将上下文值转换为字符串，并兼容数值型 code。
func contextValueToString(value any) (string, bool) {
	if value == nil {
		return "", false
	}

	var str string
	switch v := value.(type) {
	case string:
		str = v
	case json.Number:
		str = v.String()
	case float64:
		if math.Trunc(v) == v {
			str = strconv.FormatInt(int64(v), 10)
		} else {
			str = strconv.FormatFloat(v, 'f', -1, 64)
		}
	case float32:
		fv := float64(v)
		if math.Trunc(fv) == fv {
			str = strconv.FormatInt(int64(fv), 10)
		} else {
			str = strconv.FormatFloat(fv, 'f', -1, 64)
		}
	default:
		str = fmt.Sprintf("%v", v)
	}

	str = strings.TrimSpace(str)
	if str == "" {
		return "", false
	}

	return str, true
}

// clipLongField 对长文本字段做长度限制，避免日志体积膨胀。
func clipLongField(s string) string {
	runes := []rune(s)
	if len(runes) <= requestLogLongFieldMaxLength {
		return s
	}
	return string(runes[:requestLogLongFieldMaxLength])
}

// extractUpstreamRequestID 从错误消息中提取 request id。
func extractUpstreamRequestID(message string) string {
	if message == "" {
		return ""
	}

	matches := upstreamRequestIDRegex.FindStringSubmatch(message)
	if len(matches) < 2 {
		return ""
	}

	return strings.TrimSpace(matches[1])
}
