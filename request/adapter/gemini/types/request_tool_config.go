package types

// ToolConfig 表示工具配置
type ToolConfig struct {
	FunctionCallingConfig *FunctionCallingConfig `json:"functionCallingConfig,omitempty"` // 函数调用配置
	RetrievalConfig       *RetrievalConfig       `json:"retrievalConfig,omitempty"`
}

// RetrievalConfig 表示检索配置
type RetrievalConfig struct {
	LatLng       *LatLng `json:"latLng,omitempty"`
	LanguageCode *string `json:"languageCode,omitempty"`
}

// LatLng 表示地理坐标
type LatLng struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// FunctionCallingConfig 表示函数调用配置
type FunctionCallingConfig struct {
	Mode                 string   `json:"mode,omitempty"` // 模式：AUTO 或 ANY
	AllowedFunctionNames []string `json:"allowedFunctionNames,omitempty"`
}
