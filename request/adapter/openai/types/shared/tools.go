package shared

import "encoding/json"

// ToolUnion 表示工具联合类型
type ToolUnion struct {
	Function         *ToolFunction
	FileSearch       *ToolFileSearch
	ComputerUse      *ToolComputerUsePreview
	WebSearch        *ToolWebSearch
	MCP              *ToolMCP
	CodeInterpreter  *ToolCodeInterpreter
	ImageGen         *ToolImageGen
	LocalShell       *ToolLocalShell
	FunctionShell    *ToolFunctionShell
	Custom           *ToolCustom
	WebSearchPreview *ToolWebSearchPreview
	ApplyPatch       *ToolApplyPatch
}

// ToolFunction 表示函数工具
// 支持 Chat Completions 与 Responses 形态。
type ToolFunction struct {
	Type        string             `json:"type"`
	Function    FunctionDefinition `json:"function,omitempty"`
	Name        *string            `json:"name,omitempty"`
	Description *string            `json:"description,omitempty"`
	Parameters  interface{}        `json:"parameters,omitempty"`
	Strict      *bool              `json:"strict,omitempty"`
}

// ToolCustom 表示自定义工具
// 支持 custom 字段与顶层字段两种形态。
type ToolCustom struct {
	Type        string      `json:"type"`
	Custom      interface{} `json:"custom,omitempty"`
	Name        *string     `json:"name,omitempty"`
	Description *string     `json:"description,omitempty"`
	Format      interface{} `json:"format,omitempty"`
}

// ToolFileSearch 表示文件搜索工具
type ToolFileSearch struct {
	Type           string              `json:"type"`
	VectorStoreIDs []string            `json:"vector_store_ids,omitempty"`
	MaxNumResults  *int                `json:"max_num_results,omitempty"`
	RankingOptions *ToolRankingOptions `json:"ranking_options,omitempty"`
	Filters        interface{}         `json:"filters,omitempty"`
}

// ToolRankingOptions 表示检索排序选项
type ToolRankingOptions struct {
	Ranker         *string               `json:"ranker,omitempty"`
	ScoreThreshold *float64              `json:"score_threshold,omitempty"`
	HybridSearch   *ToolHybridSearchOpts `json:"hybrid_search,omitempty"`
}

// ToolHybridSearchOpts 表示混合检索权重
type ToolHybridSearchOpts struct {
	EmbeddingWeight *float64 `json:"embedding_weight,omitempty"`
	TextWeight      *float64 `json:"text_weight,omitempty"`
}

// ToolComputerUsePreview 表示电脑使用预览工具
type ToolComputerUsePreview struct {
	Type          string `json:"type"`
	Environment   string `json:"environment"`
	DisplayWidth  int    `json:"display_width"`
	DisplayHeight int    `json:"display_height"`
}

// ToolWebSearch 表示网页搜索工具
type ToolWebSearch struct {
	Type              string                 `json:"type"`
	Filters           *ToolWebSearchFilters  `json:"filters,omitempty"`
	UserLocation      *ToolWebSearchLocation `json:"user_location,omitempty"`
	SearchContextSize *string                `json:"search_context_size,omitempty"`
}

// ToolWebSearchFilters 表示网页搜索过滤器
type ToolWebSearchFilters struct {
	AllowedDomains []string `json:"allowed_domains,omitempty"`
}

// ToolWebSearchLocation 表示网页搜索位置
// type 固定为 approximate。
type ToolWebSearchLocation struct {
	Type     string  `json:"type"`
	Country  *string `json:"country,omitempty"`
	Region   *string `json:"region,omitempty"`
	City     *string `json:"city,omitempty"`
	Timezone *string `json:"timezone,omitempty"`
}

// ToolWebSearchPreview 表示网页搜索预览工具
type ToolWebSearchPreview struct {
	Type              string              `json:"type"`
	UserLocation      *ToolApproxLocation `json:"user_location,omitempty"`
	SearchContextSize *string             `json:"search_context_size,omitempty"`
}

// ToolApproxLocation 表示近似位置
type ToolApproxLocation struct {
	Type     string  `json:"type"`
	Country  *string `json:"country,omitempty"`
	Region   *string `json:"region,omitempty"`
	City     *string `json:"city,omitempty"`
	Timezone *string `json:"timezone,omitempty"`
}

// ToolMCP 表示 MCP 工具
type ToolMCP struct {
	Type              string            `json:"type"`
	ServerLabel       string            `json:"server_label"`
	ServerURL         *string           `json:"server_url,omitempty"`
	ConnectorID       *string           `json:"connector_id,omitempty"`
	Authorization     *string           `json:"authorization,omitempty"`
	ServerDescription *string           `json:"server_description,omitempty"`
	Headers           map[string]string `json:"headers,omitempty"`
	AllowedTools      interface{}       `json:"allowed_tools,omitempty"`
}

// ToolCodeInterpreter 表示代码解释器工具
type ToolCodeInterpreter struct {
	Type      string                        `json:"type"`
	Container *ToolCodeInterpreterContainer `json:"container,omitempty"`
}

// ToolCodeInterpreterContainer 表示代码解释器容器联合类型
type ToolCodeInterpreterContainer struct {
	ID   *string
	Auto *ToolCodeInterpreterContainerAuto
}

// ToolCodeInterpreterContainerAuto 表示自动容器配置
type ToolCodeInterpreterContainerAuto struct {
	Type        string   `json:"type"`
	FileIDs     []string `json:"file_ids,omitempty"`
	MemoryLimit *string  `json:"memory_limit,omitempty"`
}

// ToolImageGen 表示图片生成工具
type ToolImageGen struct {
	Type              string  `json:"type"`
	Model             *string `json:"model,omitempty"`
	Quality           *string `json:"quality,omitempty"`
	Size              *string `json:"size,omitempty"`
	OutputFormat      *string `json:"output_format,omitempty"`
	OutputCompression *int    `json:"output_compression,omitempty"`
	Moderation        *string `json:"moderation,omitempty"`
}

// ToolLocalShell 表示本地 shell 工具
type ToolLocalShell struct {
	Type string `json:"type"`
}

// ToolFunctionShell 表示 shell 工具
type ToolFunctionShell struct {
	Type string `json:"type"`
}

// ToolApplyPatch 表示 apply_patch 工具
type ToolApplyPatch struct {
	Type string `json:"type"`
}

// MarshalJSON 实现 ToolUnion 的自定义 JSON 序列化
func (t ToolUnion) MarshalJSON() ([]byte, error) {
	switch {
	case t.Function != nil:
		return json.Marshal(t.Function)
	case t.FileSearch != nil:
		return json.Marshal(t.FileSearch)
	case t.ComputerUse != nil:
		return json.Marshal(t.ComputerUse)
	case t.WebSearch != nil:
		return json.Marshal(t.WebSearch)
	case t.MCP != nil:
		return json.Marshal(t.MCP)
	case t.CodeInterpreter != nil:
		return json.Marshal(t.CodeInterpreter)
	case t.ImageGen != nil:
		return json.Marshal(t.ImageGen)
	case t.LocalShell != nil:
		return json.Marshal(t.LocalShell)
	case t.FunctionShell != nil:
		return json.Marshal(t.FunctionShell)
	case t.Custom != nil:
		return json.Marshal(t.Custom)
	case t.WebSearchPreview != nil:
		return json.Marshal(t.WebSearchPreview)
	case t.ApplyPatch != nil:
		return json.Marshal(t.ApplyPatch)
	default:
		return json.Marshal(nil)
	}
}

// UnmarshalJSON 实现 ToolUnion 的自定义 JSON 反序列化
func (t *ToolUnion) UnmarshalJSON(data []byte) error {
	// 解析到通用 map 以检查 type 字段
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	typeVal, ok := raw["type"].(string)
	if !ok {
		return nil
	}

	switch typeVal {
	case "function":
		var function ToolFunction
		if err := json.Unmarshal(data, &function); err == nil {
			t.Function = &function
			return nil
		}
	case "file_search":
		var tool ToolFileSearch
		if err := json.Unmarshal(data, &tool); err == nil {
			t.FileSearch = &tool
			return nil
		}
	case "computer_use_preview":
		var tool ToolComputerUsePreview
		if err := json.Unmarshal(data, &tool); err == nil {
			t.ComputerUse = &tool
			return nil
		}
	case "web_search", "web_search_2025_08_26":
		var tool ToolWebSearch
		if err := json.Unmarshal(data, &tool); err == nil {
			t.WebSearch = &tool
			return nil
		}
	case "mcp":
		var tool ToolMCP
		if err := json.Unmarshal(data, &tool); err == nil {
			t.MCP = &tool
			return nil
		}
	case "code_interpreter":
		var tool ToolCodeInterpreter
		if err := json.Unmarshal(data, &tool); err == nil {
			t.CodeInterpreter = &tool
			return nil
		}
	case "image_generation":
		var tool ToolImageGen
		if err := json.Unmarshal(data, &tool); err == nil {
			t.ImageGen = &tool
			return nil
		}
	case "local_shell":
		var tool ToolLocalShell
		if err := json.Unmarshal(data, &tool); err == nil {
			t.LocalShell = &tool
			return nil
		}
	case "shell":
		var tool ToolFunctionShell
		if err := json.Unmarshal(data, &tool); err == nil {
			t.FunctionShell = &tool
			return nil
		}
	case "custom":
		var tool ToolCustom
		if err := json.Unmarshal(data, &tool); err == nil {
			t.Custom = &tool
			return nil
		}
	case "web_search_preview", "web_search_preview_2025_03_11":
		var tool ToolWebSearchPreview
		if err := json.Unmarshal(data, &tool); err == nil {
			t.WebSearchPreview = &tool
			return nil
		}
	case "apply_patch":
		var tool ToolApplyPatch
		if err := json.Unmarshal(data, &tool); err == nil {
			t.ApplyPatch = &tool
			return nil
		}
	}

	return nil
}

// MarshalJSON 实现 ToolCodeInterpreterContainer 的自定义 JSON 序列化
func (c ToolCodeInterpreterContainer) MarshalJSON() ([]byte, error) {
	if c.ID != nil {
		return json.Marshal(c.ID)
	}
	if c.Auto != nil {
		return json.Marshal(c.Auto)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON 实现 ToolCodeInterpreterContainer 的自定义 JSON 反序列化
func (c *ToolCodeInterpreterContainer) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	var id string
	if err := json.Unmarshal(data, &id); err == nil {
		c.ID = &id
		return nil
	}

	var auto ToolCodeInterpreterContainerAuto
	if err := json.Unmarshal(data, &auto); err == nil {
		c.Auto = &auto
		return nil
	}

	return nil
}
