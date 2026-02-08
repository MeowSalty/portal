package types

import (
	"encoding/json"
	"fmt"
)

// ImageMediaType 图片媒体类型。
type ImageMediaType string

const (
	ImageMediaTypeJPEG ImageMediaType = "image/jpeg"
	ImageMediaTypePNG  ImageMediaType = "image/png"
	ImageMediaTypeGIF  ImageMediaType = "image/gif"
	ImageMediaTypeWEBP ImageMediaType = "image/webp"
)

// ImageSourceType 图片来源类型。
type ImageSourceType string

const (
	ImageSourceTypeBase64 ImageSourceType = "base64"
	ImageSourceTypeURL    ImageSourceType = "url"
)

// Base64ImageSource base64 图片来源。
type Base64ImageSource struct {
	Type      ImageSourceType `json:"type"`       // "base64"
	MediaType ImageMediaType  `json:"media_type"` // 图片类型
	Data      string          `json:"data"`       // base64 数据
}

// URLImageSource URL 图片来源。
type URLImageSource struct {
	Type ImageSourceType `json:"type"` // "url"
	URL  string          `json:"url"`  // 图片 URL
}

// ImageSource 图像来源联合类型。
type ImageSource struct {
	Base64 *Base64ImageSource
	URL    *URLImageSource
}

// MarshalJSON 实现 ImageSource 的序列化。
func (s ImageSource) MarshalJSON() ([]byte, error) {
	count := 0
	var payload any
	set := func(v any) {
		count++
		payload = v
	}
	if s.Base64 != nil {
		set(s.Base64)
	}
	if s.URL != nil {
		set(s.URL)
	}
	if count == 0 {
		return json.Marshal(nil)
	}
	if count > 1 {
		return nil, fmt.Errorf("图片来源只能设置一种类型")
	}
	return json.Marshal(payload)
}

// UnmarshalJSON 实现 ImageSource 的反序列化。
func (s *ImageSource) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	type typeHolder struct {
		Type ImageSourceType `json:"type"`
	}
	var t typeHolder
	if err := json.Unmarshal(data, &t); err != nil {
		return fmt.Errorf("图片来源解析失败：%w", err)
	}

	switch t.Type {
	case ImageSourceTypeBase64:
		var v Base64ImageSource
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("base64 图片来源解析失败：%w", err)
		}
		s.Base64 = &v
	case ImageSourceTypeURL:
		var v URLImageSource
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("URL 图片来源解析失败：%w", err)
		}
		s.URL = &v
	default:
		return fmt.Errorf("不支持的图片来源类型: %s", t.Type)
	}

	return nil
}

// ImageBlockParam 图像内容块。
type ImageBlockParam struct {
	Type         ContentBlockType       `json:"type"`                    // "image"
	Source       ImageSource            `json:"source"`                  // 图像来源
	CacheControl *CacheControlEphemeral `json:"cache_control,omitempty"` // 缓存控制
}
