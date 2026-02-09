package types

import (
	"fmt"
	"sync"
)

// StreamIndexContext 流索引上下文接口
//
// 用于在流式响应转换过程中生成和维护稳定的索引值，确保同一流内的索引一致性。
type StreamIndexContext interface {
	// NextSequence 获取下一个序列号
	//
	// 从 1 开始递增，用于生成 sequence_number。
	NextSequence() int

	// EnsureItemID 确保获取稳定的 item_id
	//
	// 对相同的 key 返回相同的 item_id，保证同一流内一致性。
	// key 通常由 BuildStreamIndexKey 生成。
	EnsureItemID(key string) string

	// EnsureOutputIndex 确保获取稳定的 output_index
	//
	// 对相同的 key 返回相同的 output_index，保证同一流内一致性。
	EnsureOutputIndex(key string) int

	// EnsureContentIndex 确保获取稳定的 content_index
	//
	// 对相同的 item_id 返回稳定的索引值。
	// 若 hinted >= 0 则优先使用 hinted 并缓存，否则生成新索引。
	EnsureContentIndex(itemID string, hinted int) int

	// EnsureAnnotationIndex 确保获取稳定的 annotation_index
	//
	// 对相同的 item_id 返回稳定的索引值。
	// 若 hinted >= 0 则优先使用 hinted 并缓存，否则生成新索引。
	EnsureAnnotationIndex(itemID string, hinted int) int

	// SetMessageID 设置当前流的 message_id
	//
	// 用于在 message_start 事件中保存 message_id，供后续事件使用。
	SetMessageID(messageID string)

	// GetMessageID 获取当前流的 message_id
	//
	// 返回最近设置的 message_id，若未设置则返回空字符串。
	GetMessageID() string

	// SetItemID 设置当前流的 item_id
	//
	// 用于保存特定事件的 item_id，供后续事件使用。
	SetItemID(itemID string)

	// GetItemID 获取当前流的 item_id
	//
	// 返回最近设置的 item_id，若未设置则返回空字符串。
	GetItemID() string
}

// BuildStreamIndexKey 构建流索引键
//
// 按 responseID:outputIndex:contentIndex 格式生成 key，用于标识流中的唯一位置。
func BuildStreamIndexKey(responseID string, outputIndex int, contentIndex int) string {
	return fmt.Sprintf("%s:%d:%d", responseID, outputIndex, contentIndex)
}

// defaultStreamIndexContext StreamIndexContext 的默认实现
type defaultStreamIndexContext struct {
	mu sync.RWMutex

	// sequence 当前序列号，从 1 开始递增
	sequence int

	// itemIDCache key 到 item_id 的映射缓存
	itemIDCache map[string]string

	// outputIndexCache key 到 output_index 的映射缓存
	outputIndexCache map[string]int

	// contentIndexCache item_id 到 content_index 的映射缓存
	contentIndexCache map[string]int

	// annotationIndexCache item_id 到 annotation_index 的映射缓存
	annotationIndexCache map[string]int

	// messageID 当前流的 message_id
	messageID string

	// itemID 当前流的 item_id
	itemID string
}

// NewStreamIndexContext 创建新的流索引上下文
func NewStreamIndexContext() StreamIndexContext {
	return &defaultStreamIndexContext{
		sequence:             0,
		itemIDCache:          make(map[string]string),
		outputIndexCache:     make(map[string]int),
		contentIndexCache:    make(map[string]int),
		annotationIndexCache: make(map[string]int),
		messageID:            "",
		itemID:               "",
	}
}

// NextSequence 获取下一个序列号
func (c *defaultStreamIndexContext) NextSequence() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.sequence++
	return c.sequence
}

// EnsureItemID 确保获取稳定的 item_id
func (c *defaultStreamIndexContext) EnsureItemID(key string) string {
	c.mu.RLock()
	if itemID, exists := c.itemIDCache[key]; exists {
		c.mu.RUnlock()
		return itemID
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	// 双重检查，防止并发竞争
	if itemID, exists := c.itemIDCache[key]; exists {
		return itemID
	}

	// 生成新的 item_id，使用 key 作为基础
	itemID := fmt.Sprintf("item_%s", key)
	c.itemIDCache[key] = itemID
	return itemID
}

// EnsureOutputIndex 确保获取稳定的 output_index
func (c *defaultStreamIndexContext) EnsureOutputIndex(key string) int {
	c.mu.RLock()
	if index, exists := c.outputIndexCache[key]; exists {
		c.mu.RUnlock()
		return index
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	// 双重检查
	if index, exists := c.outputIndexCache[key]; exists {
		return index
	}

	// 生成新的 output_index，从 0 开始
	index := len(c.outputIndexCache)
	c.outputIndexCache[key] = index
	return index
}

// EnsureContentIndex 确保获取稳定的 content_index
func (c *defaultStreamIndexContext) EnsureContentIndex(itemID string, hinted int) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 若 hinted >= 0，优先使用 hinted 并缓存
	if hinted >= 0 {
		if existingIndex, exists := c.contentIndexCache[itemID]; exists {
			return existingIndex
		}
		c.contentIndexCache[itemID] = hinted
		return hinted
	}

	// 否则生成新索引
	if index, exists := c.contentIndexCache[itemID]; exists {
		return index
	}

	index := len(c.contentIndexCache)
	c.contentIndexCache[itemID] = index
	return index
}

// EnsureAnnotationIndex 确保获取稳定的 annotation_index
func (c *defaultStreamIndexContext) EnsureAnnotationIndex(itemID string, hinted int) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 若 hinted >= 0，优先使用 hinted 并缓存
	if hinted >= 0 {
		if existingIndex, exists := c.annotationIndexCache[itemID]; exists {
			return existingIndex
		}
		c.annotationIndexCache[itemID] = hinted
		return hinted
	}

	// 否则生成新索引
	if index, exists := c.annotationIndexCache[itemID]; exists {
		return index
	}

	index := len(c.annotationIndexCache)
	c.annotationIndexCache[itemID] = index
	return index
}

// SetMessageID 设置当前流的 message_id
func (c *defaultStreamIndexContext) SetMessageID(messageID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.messageID = messageID
}

// GetMessageID 获取当前流的 message_id
func (c *defaultStreamIndexContext) GetMessageID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.messageID
}

// SetItemID 设置当前流的 item_id
func (c *defaultStreamIndexContext) SetItemID(itemID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.itemID = itemID
}

// GetItemID 获取当前流的 item_id
func (c *defaultStreamIndexContext) GetItemID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.itemID
}
