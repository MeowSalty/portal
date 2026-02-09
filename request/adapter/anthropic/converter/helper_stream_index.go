package converter

import (
	"github.com/MeowSalty/portal/request/adapter/types"
)

// fillMissingIndices 补齐缺失的索引字段
//
// 根据 plans/stream-index-repair-plan.md B.2 章节实现：
//   - sequence_number：若为 0 -> ctx.NextSequence()
//   - output_index：若缺失（< 0）-> ctx.EnsureOutputIndex(message_id)，0 值保持不变
//   - item_id：若有 tool_use.id -> 用 tool_use.id；否则 message_id（message_start/stop）；content_block 事件用 message_id:content_index 组合
//   - content_index：来自 event.Index（content_block*）；若缺失（< 0）-> ctx.EnsureContentIndex(item_id, -1)，0 值保持不变
//   - 缺失索引统一使用 -1 作为哨兵值，与合法 0 区分
func fillMissingIndices(contract *types.StreamEventContract, messageID string, outputIndex int, toolCalls []types.StreamToolCall, contentIndex int, ctx types.StreamIndexContext) {
	// 补齐 sequence_number
	if contract.SequenceNumber == 0 {
		contract.SequenceNumber = ctx.NextSequence()
	}

	// 补齐 output_index（仅处理负值，0 值保持不变）
	if contract.OutputIndex < 0 {
		if outputIndex > 0 {
			contract.OutputIndex = outputIndex
		} else if messageID != "" {
			contract.OutputIndex = ctx.EnsureOutputIndex(messageID)
		}
	}

	// 补齐 item_id
	if contract.ItemID == "" {
		// 优先使用 tool_call id
		if len(toolCalls) > 0 && toolCalls[0].ID != "" {
			contract.ItemID = toolCalls[0].ID
		} else if messageID != "" {
			// 使用 message_id 作为 item_id
			contract.ItemID = messageID
		} else if contentIndex >= 0 {
			// content_block 事件使用 message_id:content_index 组合
			key := types.BuildStreamIndexKey(contract.ResponseID, contract.OutputIndex, contentIndex)
			contract.ItemID = ctx.EnsureItemID(key)
		} else {
			// 尝试从上下文获取 item_id
			if ctxItemID := ctx.GetItemID(); ctxItemID != "" {
				contract.ItemID = ctxItemID
			}
		}
	}

	// 补齐 content_index（仅处理负值，0 值保持不变）
	if contract.ContentIndex < 0 {
		if contentIndex >= 0 {
			// 使用 event.Index 作为 content_index
			contract.ContentIndex = ctx.EnsureContentIndex(contract.ItemID, contentIndex)
		} else {
			// 若缺失，使用 EnsureContentIndex
			contract.ContentIndex = ctx.EnsureContentIndex(contract.ItemID, -1)
		}
	}
}
