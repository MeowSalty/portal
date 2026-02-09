package converter

import (
	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// convertSystemFromContract 从 Contract 转换系统指令。
func convertSystemFromContract(system *types.System) (*anthropicTypes.SystemParam, error) {
	if system == nil {
		return nil, nil
	}

	result := &anthropicTypes.SystemParam{}

	if system.Text != nil {
		result.StringValue = system.Text
	} else if len(system.Parts) > 0 {
		blocks := make([]anthropicTypes.TextBlockParam, 0, len(system.Parts))
		for _, part := range system.Parts {
			if part.Text != nil {
				block := anthropicTypes.TextBlockParam{
					Type: anthropicTypes.ContentBlockTypeText,
					Text: *part.Text,
				}

				// 从 VendorExtras 恢复 CacheControl
				if part.VendorExtras != nil {
					if cc, ok := part.VendorExtras["cache_control"].(*anthropicTypes.CacheControlEphemeral); ok {
						block.CacheControl = cc
					}
				}

				blocks = append(blocks, block)
			}
		}
		result.Blocks = blocks
	}

	return result, nil
}

// convertSystemToContract 转换系统指令。
func convertSystemToContract(system *anthropicTypes.SystemParam) (*types.System, error) {
	if system == nil {
		return nil, nil
	}

	result := &types.System{}

	if system.StringValue != nil {
		result.Text = system.StringValue
	} else if len(system.Blocks) > 0 {
		parts := make([]types.ContentPart, 0, len(system.Blocks))
		for _, block := range system.Blocks {
			part, err := convertTextBlockToContract(&block)
			if err != nil {
				return nil, err
			}
			parts = append(parts, *part)
		}
		result.Parts = parts
	}

	return result, nil
}
