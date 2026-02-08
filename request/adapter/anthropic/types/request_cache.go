package types

// CacheControlType 缓存控制类型。
type CacheControlType string

const (
	CacheControlTypeEphemeral CacheControlType = "ephemeral"
)

// CacheControlTTL 缓存 TTL。
type CacheControlTTL string

const (
	CacheControlTTL5m CacheControlTTL = "5m"
	CacheControlTTL1h CacheControlTTL = "1h"
)

// CacheControlEphemeral 临时缓存控制。
type CacheControlEphemeral struct {
	Type CacheControlType `json:"type"`          // "ephemeral"
	TTL  *CacheControlTTL `json:"ttl,omitempty"` // "5m" 或 "1h"
}
