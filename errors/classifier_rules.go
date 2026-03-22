package errors

func boolPtr(v bool) *bool {
	return &v
}

// DefaultClassificationRules 返回默认分类规则。
func DefaultClassificationRules() []ClassificationRule {
	rules := []ClassificationRule{
		// source 规则
		{
			ID:       "source-explicit-client",
			Enabled:  true,
			Stage:    ClassificationStageSource,
			Priority: 100,
			Conditions: RuleConditions{
				ErrorFrom: []ErrorFromValue{ErrorFromClient},
			},
			Decision: RuleDecision{
				Source: ErrorFromClient,
			},
			Confidence: ConfidenceHigh,
			Reason:     "显式错误来源为 client",
		},
		{
			ID:       "source-explicit-upstream",
			Enabled:  true,
			Stage:    ClassificationStageSource,
			Priority: 100,
			Conditions: RuleConditions{
				ErrorFrom: []ErrorFromValue{ErrorFromUpstream},
			},
			Decision: RuleDecision{
				Source: ErrorFromUpstream,
			},
			Confidence: ConfidenceHigh,
			Reason:     "显式错误来源为 upstream",
		},
		{
			ID:       "source-explicit-server",
			Enabled:  true,
			Stage:    ClassificationStageSource,
			Priority: 100,
			Conditions: RuleConditions{
				ErrorFrom: []ErrorFromValue{ErrorFromServer},
			},
			Decision: RuleDecision{
				Source: ErrorFromServer,
			},
			Confidence: ConfidenceHigh,
			Reason:     "显式错误来源为 server",
		},
		{
			ID:       "source-explicit-gateway",
			Enabled:  true,
			Stage:    ClassificationStageSource,
			Priority: 100,
			Conditions: RuleConditions{
				ErrorFrom: []ErrorFromValue{ErrorFromGateway},
			},
			Decision: RuleDecision{
				Source: ErrorFromGateway,
			},
			Confidence: ConfidenceHigh,
			Reason:     "显式错误来源为 gateway",
		},
		{
			ID:       "source-client-canceled",
			Enabled:  true,
			Stage:    ClassificationStageSource,
			Priority: 95,
			Conditions: RuleConditions{
				Codes:       []ErrorCode{ErrCodeAborted},
				AnyContains: []string{"cancel", "取消"},
			},
			Decision: RuleDecision{
				Source: ErrorFromClient,
			},
			Confidence: ConfidenceHigh,
			Reason:     "命中客户端取消信号",
		},
		{
			ID:       "source-upstream-keywords",
			Enabled:  true,
			Stage:    ClassificationStageSource,
			Priority: 90,
			Conditions: RuleConditions{
				AnyContains: []string{"upstream", "provider", "vendor", "bad_response_status_code", "do_request_failed"},
			},
			Decision: RuleDecision{
				Source: ErrorFromUpstream,
			},
			Confidence: ConfidenceHigh,
			Reason:     "命中上游特征关键词",
		},
		{
			ID:       "source-http-response-received",
			Enabled:  true,
			Stage:    ClassificationStageSource,
			Priority: 30,
			Conditions: RuleConditions{
				HTTPResponseReceived: boolPtr(true),
			},
			Decision: RuleDecision{
				Source: ErrorFromServer,
			},
			Confidence: ConfidenceMedium,
			Reason:     "已收到上游 HTTP 响应，默认归类 server",
		},
		{
			ID:       "source-fallback-gateway",
			Enabled:  true,
			Stage:    ClassificationStageSource,
			Priority: 0,
			Conditions: RuleConditions{
				HTTPResponseReceived: boolPtr(false),
			},
			Decision: RuleDecision{
				Source: ErrorFromGateway,
			},
			Confidence: ConfidenceLow,
			Reason:     "未收到 HTTP 响应，兜底归类 gateway",
		},

		// level 规则
		{
			ID:       "level-auth-status",
			Enabled:  true,
			Stage:    ClassificationStageLevel,
			Priority: 100,
			Conditions: RuleConditions{
				HTTPStatuses: []int{401, 403},
				AnyContains:  []string{"api key", "invalid key", "unauthorized", "authentication", "permission", "token", "鉴权", "认证", "权限", "密钥"},
			},
			Decision: RuleDecision{
				Level: ErrorLevelKey,
			},
			Confidence: ConfidenceHigh,
			Reason:     "认证/权限状态码与密钥信号联合命中",
		},
		{
			ID:       "level-auth-code",
			Enabled:  true,
			Stage:    ClassificationStageLevel,
			Priority: 95,
			Conditions: RuleConditions{
				Codes: []ErrorCode{ErrCodeAuthenticationFailed, ErrCodePermissionDenied},
			},
			Decision: RuleDecision{
				Level: ErrorLevelKey,
			},
			Confidence: ConfidenceHigh,
			Reason:     "显式认证类错误码命中",
		},
		{
			ID:       "level-quota-with-key",
			Enabled:  true,
			Stage:    ClassificationStageLevel,
			Priority: 88,
			Conditions: RuleConditions{
				HTTPStatuses: []int{429},
				AnyContains:  []string{"quota", "配额", "额度", "欠费", "billing"},
				AllContains:  []string{"key"},
			},
			Decision: RuleDecision{
				Level: ErrorLevelKey,
			},
			Confidence: ConfidenceHigh,
			Reason:     "额度类错误且包含 key 信号",
		},
		{
			ID:       "level-platform-strong-keywords",
			Enabled:  true,
			Stage:    ClassificationStageLevel,
			Priority: 85,
			Conditions: RuleConditions{
				AnyContains: []string{"渠道", "路由", "节点", "平台内部", "channel unavailable", "route", "proxy", "backend", "platform"},
			},
			Decision: RuleDecision{
				Level: ErrorLevelPlatform,
			},
			Confidence: ConfidenceHigh,
			Reason:     "命中平台级高置信度关键词",
		},
		{
			ID:       "level-model-not-found",
			Enabled:  true,
			Stage:    ClassificationStageLevel,
			Priority: 80,
			Conditions: RuleConditions{
				HTTPStatuses: []int{400, 404},
				AnyContains:  []string{"模型", "model", "model not found", "model unavailable", "unsupported model", "unknown model"},
			},
			Decision: RuleDecision{
				Level: ErrorLevelModel,
			},
			Confidence: ConfidenceHigh,
			Reason:     "模型关键词与状态码联合命中",
		},
		{
			ID:       "level-model-keywords",
			Enabled:  true,
			Stage:    ClassificationStageLevel,
			Priority: 60,
			Conditions: RuleConditions{
				AnyContains: []string{"模型", "model", "unsupported model", "unknown model", "model unavailable"},
			},
			Decision: RuleDecision{
				Level: ErrorLevelModel,
			},
			Confidence: ConfidenceMedium,
			Reason:     "命中模型关键词",
		},
		{
			ID:       "level-upstream-default-model",
			Enabled:  true,
			Stage:    ClassificationStageLevel,
			Priority: 40,
			Conditions: RuleConditions{
				ErrorFrom: []ErrorFromValue{ErrorFromUpstream},
			},
			Decision: RuleDecision{
				Level: ErrorLevelModel,
			},
			Confidence: ConfidenceMedium,
			Reason:     "来源 upstream，保守归类 model",
		},
		{
			ID:       "level-ambiguous-timeout",
			Enabled:  true,
			Stage:    ClassificationStageLevel,
			Priority: 20,
			Conditions: RuleConditions{
				AnyContains: []string{"gateway timeout", "timeout", "超时"},
			},
			Decision: RuleDecision{
				Level: ErrorLevelModel,
			},
			Confidence: ConfidenceLow,
			Reason:     "仅命中模糊超时文本，保守归类 model",
		},
		{
			ID:         "level-fallback-model",
			Enabled:    true,
			Stage:      ClassificationStageLevel,
			Priority:   0,
			Conditions: RuleConditions{},
			Decision: RuleDecision{
				Level: ErrorLevelModel,
			},
			Confidence: ConfidenceLow,
			Reason:     "缺少高置信度平台/密钥证据，兜底归类 model",
		},

		// resource 规则
		{
			ID:       "resource-auth-status",
			Enabled:  true,
			Stage:    ClassificationStageResource,
			Priority: 100,
			Conditions: RuleConditions{
				HTTPStatuses: []int{401, 403},
				AnyContains:  []string{"api key", "invalid key", "unauthorized", "authentication", "permission", "token", "鉴权", "认证", "权限", "密钥"},
			},
			Decision: RuleDecision{
				Resource: ErrorResourceAPIKey,
			},
			Confidence: ConfidenceHigh,
			Reason:     "认证/权限状态码与密钥信号联合命中",
		},
		{
			ID:       "resource-auth-code",
			Enabled:  true,
			Stage:    ClassificationStageResource,
			Priority: 95,
			Conditions: RuleConditions{
				Codes: []ErrorCode{ErrCodeAuthenticationFailed, ErrCodePermissionDenied},
			},
			Decision: RuleDecision{
				Resource: ErrorResourceAPIKey,
			},
			Confidence: ConfidenceHigh,
			Reason:     "显式认证类错误码命中",
		},
		{
			ID:       "resource-quota-with-key",
			Enabled:  true,
			Stage:    ClassificationStageResource,
			Priority: 88,
			Conditions: RuleConditions{
				HTTPStatuses: []int{429},
				AnyContains:  []string{"quota", "配额", "额度", "欠费", "billing"},
				AllContains:  []string{"key"},
			},
			Decision: RuleDecision{
				Resource: ErrorResourceAPIKey,
			},
			Confidence: ConfidenceHigh,
			Reason:     "额度类错误且包含 key 信号",
		},
		{
			ID:       "resource-platform-strong-keywords",
			Enabled:  true,
			Stage:    ClassificationStageResource,
			Priority: 85,
			Conditions: RuleConditions{
				AnyContains: []string{"渠道", "路由", "节点", "平台内部", "channel unavailable", "route", "proxy", "backend", "platform"},
			},
			Decision: RuleDecision{
				Resource: ErrorResourcePlatform,
			},
			Confidence: ConfidenceHigh,
			Reason:     "命中平台级高置信度关键词",
		},
		{
			ID:       "resource-model-not-found",
			Enabled:  true,
			Stage:    ClassificationStageResource,
			Priority: 80,
			Conditions: RuleConditions{
				HTTPStatuses: []int{400, 404},
				AnyContains:  []string{"模型", "model", "model not found", "model unavailable", "unsupported model", "unknown model"},
			},
			Decision: RuleDecision{
				Resource: ErrorResourceModel,
			},
			Confidence: ConfidenceHigh,
			Reason:     "模型关键词与状态码联合命中",
		},
		{
			ID:       "resource-model-keywords",
			Enabled:  true,
			Stage:    ClassificationStageResource,
			Priority: 60,
			Conditions: RuleConditions{
				AnyContains: []string{"模型", "model", "unsupported model", "unknown model", "model unavailable"},
			},
			Decision: RuleDecision{
				Resource: ErrorResourceModel,
			},
			Confidence: ConfidenceMedium,
			Reason:     "命中模型关键词",
		},
		{
			ID:       "resource-upstream-default-model",
			Enabled:  true,
			Stage:    ClassificationStageResource,
			Priority: 40,
			Conditions: RuleConditions{
				ErrorFrom: []ErrorFromValue{ErrorFromUpstream},
			},
			Decision: RuleDecision{
				Resource: ErrorResourceModel,
			},
			Confidence: ConfidenceMedium,
			Reason:     "来源 upstream，保守归类 model 资源",
		},
		{
			ID:       "resource-ambiguous-timeout",
			Enabled:  true,
			Stage:    ClassificationStageResource,
			Priority: 20,
			Conditions: RuleConditions{
				AnyContains: []string{"gateway timeout", "timeout", "超时"},
			},
			Decision: RuleDecision{
				Resource: ErrorResourceModel,
			},
			Confidence: ConfidenceLow,
			Reason:     "仅命中模糊超时文本，保守归类 model 资源",
		},
		{
			ID:         "resource-fallback-model",
			Enabled:    true,
			Stage:      ClassificationStageResource,
			Priority:   0,
			Conditions: RuleConditions{},
			Decision: RuleDecision{
				Resource: ErrorResourceModel,
			},
			Confidence: ConfidenceLow,
			Reason:     "缺少高置信度平台/密钥证据，兜底归类 model 资源",
		},
	}

	cloned := make([]ClassificationRule, len(rules))
	copy(cloned, rules)
	return cloned
}
