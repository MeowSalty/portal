package errors

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestUnicodeSafeTruncate(t *testing.T) {
	// 测试包含不同类型字符的字符串截断
	testCases := []struct {
		name        string
		input       string
		shouldTrunc bool
	}{
		{
			name:        "Long Chinese string",
			input:       strings.Repeat("这是一个很长的中文字符串测试", 50),
			shouldTrunc: true,
		},
		{
			name:        "Long Emoji string",
			input:       strings.Repeat("Hello 🌍 World 🚀 ", 50),
			shouldTrunc: true,
		},
		{
			name:        "Long mixed string",
			input:       strings.Repeat("Hello 世界 World 测试 ", 50),
			shouldTrunc: true,
		},
		{
			name:        "Short string",
			input:       "短字符串",
			shouldTrunc: false,
		},
		{
			name:        "Long ASCII string",
			input:       strings.Repeat("This is a long ASCII string ", 50),
			shouldTrunc: true,
		},
		{
			name:        "Edge case - exactly at limit",
			input:       strings.Repeat("a", maxContextValueLength),
			shouldTrunc: false,
		},
		{
			name:        "Edge case - one over limit",
			input:       strings.Repeat("a", maxContextValueLength+1),
			shouldTrunc: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := New(ErrCodeInternal, "test error").
				WithContext("test_key", tc.input)

			errStr := err.Error()

			// 验证错误字符串是有效的 UTF-8
			if !utf8.ValidString(errStr) {
				t.Errorf("结果错误字符串不是有效的 UTF-8: %q", errStr)
			}

			// 检查是否包含截断后缀
			hasTruncateSuffix := strings.Contains(errStr, truncateSuffix)
			if tc.shouldTrunc && !hasTruncateSuffix {
				t.Errorf("预期字符串被截断但未找到截断后缀")
			}
			if !tc.shouldTrunc && hasTruncateSuffix {
				t.Errorf("预期字符串不被截断但找到了截断后缀")
			}

			// 如果被截断，验证截断后的总长度不超过限制
			if tc.shouldTrunc {
				// 从错误字符串中提取上下文值
				contextStart := strings.Index(errStr, "test_key=")
				if contextStart == -1 {
					t.Fatal("未找到上下文键")
				}
				contextStart += len("test_key=")

				// 找到值的结束位置（逗号或右括号）
				contextEnd := strings.IndexAny(errStr[contextStart:], ",}")
				if contextEnd == -1 {
					t.Fatal("未找到上下文值结束位置")
				}

				contextValue := errStr[contextStart : contextStart+contextEnd]

				// 验证截断后的值长度不超过 maxContextValueLength
				if len(contextValue) > maxContextValueLength {
					t.Errorf("截断后的值长度 %d 超过了最大限制 %d", len(contextValue), maxContextValueLength)
				}
			}
		})
	}
}

func TestUnicodeTruncateAtBoundary(t *testing.T) {
	// 测试在多字节字符边界处的截断
	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "Chinese characters at boundary",
			input: strings.Repeat("中", maxContextValueLength/3+10), // 中文字符占 3 字节
		},
		{
			name:  "Emoji at boundary",
			input: strings.Repeat("🎉", maxContextValueLength/4+10), // Emoji 占 4 字节
		},
		{
			name:  "Mixed multibyte",
			input: strings.Repeat("a 中🎉", maxContextValueLength/8+10),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := New(ErrCodeInternal, "test error").
				WithContext("key", tc.input)

			errStr := err.Error()

			// 最重要的检查：确保结果是有效的 UTF-8
			if !utf8.ValidString(errStr) {
				t.Errorf("截断后的字符串包含无效的 UTF-8 序列")

				// 打印详细信息帮助调试
				for i, r := range errStr {
					if r == utf8.RuneError {
						t.Logf("在位置 %d 发现无效的 UTF-8 字符", i)
					}
				}
			}
		})
	}
}
