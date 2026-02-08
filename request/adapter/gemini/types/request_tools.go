package types

// Tool 表示工具定义
type Tool struct {
	FunctionDeclarations  []FunctionDeclaration  `json:"functionDeclarations,omitempty"` // 函数声明
	GoogleSearchRetrieval *GoogleSearchRetrieval `json:"googleSearchRetrieval,omitempty"`
	CodeExecution         *CodeExecution         `json:"codeExecution,omitempty"`
	GoogleSearch          *GoogleSearch          `json:"googleSearch,omitempty"`
	ComputerUse           *ComputerUse           `json:"computerUse,omitempty"`
	URLContext            *UrlContext            `json:"urlContext,omitempty"`
	FileSearch            *FileSearch            `json:"fileSearch,omitempty"`
	MCPServers            []McpServer            `json:"mcpServers,omitempty"`
	GoogleMaps            *GoogleMaps            `json:"googleMaps,omitempty"`
}

// GoogleSearchRetrieval 表示 Google 搜索检索配置
type GoogleSearchRetrieval struct {
	DynamicRetrievalConfig *DynamicRetrievalConfig `json:"dynamicRetrievalConfig,omitempty"`
}

// DynamicRetrievalConfig 表示动态检索配置
type DynamicRetrievalConfig struct {
	Mode             string   `json:"mode"`
	DynamicThreshold *float64 `json:"dynamicThreshold,omitempty"`
}

// CodeExecution 表示代码执行工具
type CodeExecution struct{}

// GoogleSearch 表示 Google 搜索工具
type GoogleSearch struct {
	TimeRangeFilter *Interval `json:"timeRangeFilter,omitempty"`
}

// Interval 表示时间范围
type Interval struct {
	StartTime *string `json:"startTime,omitempty"`
	EndTime   *string `json:"endTime,omitempty"`
}

// ComputerUse 表示计算机使用工具
type ComputerUse struct {
	Environment                 string   `json:"environment"`
	ExcludedPredefinedFunctions []string `json:"excludedPredefinedFunctions,omitempty"`
}

// UrlContext 表示 URL 上下文工具
type UrlContext struct{}

// FileSearch 表示文件检索工具
type FileSearch struct {
	FileSearchStoreNames []string `json:"fileSearchStoreNames"`
	TopK                 *int     `json:"topK,omitempty"`
	MetadataFilter       *string  `json:"metadataFilter,omitempty"`
}

// McpServer 表示 MCP 服务器
type McpServer struct {
	StreamableHTTPTransport *StreamableHttpTransport `json:"streamableHttpTransport,omitempty"`
	Name                    string                   `json:"name"`
}

// StreamableHttpTransport 表示可流式 HTTP 传输
type StreamableHttpTransport struct {
	URL              string            `json:"url"`
	Headers          map[string]string `json:"headers,omitempty"`
	Timeout          *string           `json:"timeout,omitempty"`
	SSEReadTimeout   *string           `json:"sseReadTimeout,omitempty"`
	TerminateOnClose *bool             `json:"terminateOnClose,omitempty"`
}

// GoogleMaps 表示 Google Maps 工具
type GoogleMaps struct {
	EnableWidget *bool `json:"enableWidget,omitempty"`
}
