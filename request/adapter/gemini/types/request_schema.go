package types

// Schema 表示 JSON schema
type Schema struct {
	Type             string            `json:"type"`                       // 类型
	Format           string            `json:"format,omitempty"`           // 格式
	Title            *string           `json:"title,omitempty"`            // 标题
	Description      string            `json:"description,omitempty"`      // 描述
	Nullable         *bool             `json:"nullable,omitempty"`         // 是否可为空
	Enum             []string          `json:"enum,omitempty"`             // 枚举值
	Items            *Schema           `json:"items,omitempty"`            // 数组项 schema
	MaxItems         *string           `json:"maxItems,omitempty"`         // 最大元素数量
	MinItems         *string           `json:"minItems,omitempty"`         // 最小元素数量
	Properties       map[string]Schema `json:"properties,omitempty"`       // 属性
	Required         []string          `json:"required,omitempty"`         // 必需属性
	MinProperties    *string           `json:"minProperties,omitempty"`    // 最小属性数量
	MaxProperties    *string           `json:"maxProperties,omitempty"`    // 最大属性数量
	Minimum          *float64          `json:"minimum,omitempty"`          // 最小值
	Maximum          *float64          `json:"maximum,omitempty"`          // 最大值
	MinLength        *string           `json:"minLength,omitempty"`        // 最小长度
	MaxLength        *string           `json:"maxLength,omitempty"`        // 最大长度
	Pattern          *string           `json:"pattern,omitempty"`          // 正则约束
	Example          interface{}       `json:"example,omitempty"`          // 示例
	AnyOf            []Schema          `json:"anyOf,omitempty"`            // 任一满足
	PropertyOrdering []string          `json:"propertyOrdering,omitempty"` // 属性顺序
	Default          interface{}       `json:"default,omitempty"`          // 默认值
}
