package responses

import (
	"encoding/json"
	"fmt"

	portalErrors "github.com/MeowSalty/portal/errors"
)

// AnnotationType 定义注释类型常量
type AnnotationType string

const (
	AnnotationTypeFileCitation          AnnotationType = "file_citation"
	AnnotationTypeURLCitation           AnnotationType = "url_citation"
	AnnotationTypeContainerFileCitation AnnotationType = "container_file_citation"
	AnnotationTypeFilePath              AnnotationType = "file_path"
)

// Annotation 是注释的包装类型，使用多指针 oneof 结构。
// 强类型约束：FileCitation、URLCitation、ContainerFileCitation、FilePath 互斥，仅能有一个非空。
type Annotation struct {
	FileCitation          *FileCitationAnnotation          `json:"-"` // 文件引用注释
	URLCitation           *URLCitationAnnotation           `json:"-"` // URL 引用注释
	ContainerFileCitation *ContainerFileCitationAnnotation `json:"-"` // 容器文件引用注释
	FilePath              *FilePathAnnotation              `json:"-"` // 文件路径注释
}

// validateOneOf 验证仅有一个指针非空。
// 若全空或多于一个非空，返回错误。
func (a *Annotation) validateOneOf() error {
	count := 0
	if a.FileCitation != nil {
		count++
	}
	if a.URLCitation != nil {
		count++
	}
	if a.ContainerFileCitation != nil {
		count++
	}
	if a.FilePath != nil {
		count++
	}

	if count == 0 {
		return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "注释不能为空，必须设置一个类型")
	}
	if count > 1 {
		return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "注释类型冲突，不能同时设置多个类型")
	}

	return nil
}

// GetType 返回注释类型。
// 若没有设置任何类型，返回空字符串。
func (a *Annotation) GetType() AnnotationType {
	if a.FileCitation != nil {
		return AnnotationTypeFileCitation
	}
	if a.URLCitation != nil {
		return AnnotationTypeURLCitation
	}
	if a.ContainerFileCitation != nil {
		return AnnotationTypeContainerFileCitation
	}
	if a.FilePath != nil {
		return AnnotationTypeFilePath
	}
	return ""
}

// MarshalJSON 实现 json.Marshaler 接口。
// 要求仅一个指针非空；全空或多于一个非空报错。
func (a Annotation) MarshalJSON() ([]byte, error) {
	if err := a.validateOneOf(); err != nil {
		return nil, err
	}

	// 序列化非空指针
	if a.FileCitation != nil {
		return json.Marshal(a.FileCitation)
	}
	if a.URLCitation != nil {
		return json.Marshal(a.URLCitation)
	}
	if a.ContainerFileCitation != nil {
		return json.Marshal(a.ContainerFileCitation)
	}
	if a.FilePath != nil {
		return json.Marshal(a.FilePath)
	}

	return nil, portalErrors.New(portalErrors.ErrCodeInvalidArgument, "注释类型不支持")
}

// UnmarshalJSON 实现 json.Unmarshaler 接口。
// 先解析 type 字段，分派到对应指针字段；未知 type 报错。
func (a *Annotation) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	// 先解析出 type 字段
	var typeWrapper struct {
		Type AnnotationType `json:"type"`
	}

	if err := json.Unmarshal(data, &typeWrapper); err != nil {
		return err
	}

	if typeWrapper.Type == "" {
		return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "注释类型为空")
	}

	// 根据 type 创建对应的具体类型
	switch typeWrapper.Type {
	case AnnotationTypeFileCitation:
		var fc FileCitationAnnotation
		if err := json.Unmarshal(data, &fc); err != nil {
			return err
		}
		a.FileCitation = &fc
	case AnnotationTypeURLCitation:
		var uc URLCitationAnnotation
		if err := json.Unmarshal(data, &uc); err != nil {
			return err
		}
		a.URLCitation = &uc
	case AnnotationTypeContainerFileCitation:
		var cfc ContainerFileCitationAnnotation
		if err := json.Unmarshal(data, &cfc); err != nil {
			return err
		}
		a.ContainerFileCitation = &cfc
	case AnnotationTypeFilePath:
		var fp FilePathAnnotation
		if err := json.Unmarshal(data, &fp); err != nil {
			return err
		}
		a.FilePath = &fp
	default:
		return portalErrors.New(portalErrors.ErrCodeInvalidArgument, fmt.Sprintf("未知的注释类型: %s", typeWrapper.Type))
	}

	return nil
}

// FileCitationAnnotation 表示文件引用注释
type FileCitationAnnotation struct {
	Type     AnnotationType `json:"type"`     // 类型
	FileID   string         `json:"file_id"`  // 文件 ID
	Index    int            `json:"index"`    // 索引
	Filename string         `json:"filename"` // 文件名
}

// URLCitationAnnotation 表示 URL 引用注释
type URLCitationAnnotation struct {
	Type       AnnotationType `json:"type"`        // 类型
	URL        string         `json:"url"`         // URL
	StartIndex int            `json:"start_index"` // 起始索引
	EndIndex   int            `json:"end_index"`   // 结束索引
	Title      string         `json:"title"`       // 标题
}

// ContainerFileCitationAnnotation 表示容器文件引用注释
type ContainerFileCitationAnnotation struct {
	Type        AnnotationType `json:"type"`         // 类型
	ContainerID string         `json:"container_id"` // 容器 ID
	FileID      string         `json:"file_id"`      // 文件 ID
	StartIndex  int            `json:"start_index"`  // 起始索引
	EndIndex    int            `json:"end_index"`    // 结束索引
	Filename    string         `json:"filename"`     // 文件名
}

// FilePathAnnotation 表示文件路径注释
type FilePathAnnotation struct {
	Type   AnnotationType `json:"type"`    // 类型
	FileID string         `json:"file_id"` // 文件 ID
	Index  int            `json:"index"`   // 索引
}
