package portal

// NativeOption 定义原生请求的可选配置函数。
type NativeOption func(*nativeOptions)

// nativeOptions 存储原生请求的所有可选配置。
type nativeOptions struct {
	compatMode bool // 是否启用兼容模式
}

// applyNativeOptions 应用所有选项并返回配置。
func applyNativeOptions(opts []NativeOption) *nativeOptions {
	options := &nativeOptions{}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

// WithCompatMode 启用兼容模式。
//
// 当原生端点不可用时，自动降级到默认端点，
// 通过 Contract 归一格式中转完成请求，并将响应转回原生格式。
func WithCompatMode() NativeOption {
	return func(o *nativeOptions) {
		o.compatMode = true
	}
}
