package chat

// WebSearchContextSize 表示 Web 搜索上下文大小
type WebSearchContextSize = string

const (
	WebSearchContextSizeSmall  WebSearchContextSize = "small"
	WebSearchContextSizeMedium WebSearchContextSize = "medium"
	WebSearchContextSizeLarge  WebSearchContextSize = "large"
)

// WebSearchOptions 表示网络搜索选项
type WebSearchOptions struct {
	UserLocation      *UserLocation         `json:"user_location,omitempty"`       // 用户位置
	SearchContextSize *WebSearchContextSize `json:"search_context_size,omitempty"` // 搜索上下文大小
}

// UserLocationType 表示用户位置类型
// 固定为 approximate。
type UserLocationType = string

const (
	UserLocationTypeApproximate UserLocationType = "approximate"
)

// UserLocation 表示用户位置
type UserLocation struct {
	Type        UserLocationType     `json:"type"`                  // 类型
	Approximate *ApproximateLocation `json:"approximate,omitempty"` // 近似位置
}

// ApproximateLocation 表示近似位置
type ApproximateLocation struct {
	City     *string `json:"city,omitempty"`     // 城市
	Country  *string `json:"country,omitempty"`  // 国家
	Region   *string `json:"region,omitempty"`   // 地区
	Timezone *string `json:"timezone,omitempty"` // 时区
}
