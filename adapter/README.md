# Adapter 包

Adapter 包提供了一个统一的接口来集成不同的 AI 服务提供商，如 OpenAI、Anthropic 等。该包通过统一的 `Provider` 接口抽象了不同提供商的 API 差异，使得添加新的 AI 服务提供商变得简单且标准化。

## 架构概述

### 核心接口

#### `Provider` 接口

```go
type Provider interface {
    Name() string
    CreateRequest(request *coreTypes.Request, channel *coreTypes.Channel) (interface{}, error)
    ParseResponse(responseData []byte) (*coreTypes.Response, error)
    ParseStreamResponse(responseData []byte) (*coreTypes.Response, error)
    APIEndpoint() string
    Headers(channel *coreTypes.Channel) map[string]string
    SupportsStreaming() bool
}
```

#### `Adapter` 结构体

统一适配器实现，封装了 HTTP 客户端和提供商特定的逻辑。

### 注册机制

通过 `registry.go` 实现了提供商的注册和发现机制：

- `RegisterProviderFactory`: 注册新的提供商工厂
- `NewAdapterRegistry`: 创建适配器注册表
- `GetAdapter`: 获取指定名称的适配器

## 如何添加新的适配器

### 步骤 1: 目录结构

创建以下目录结构：

```text
adapter/
├── newprovider/
│   ├── converter/
│   │   ├── request.go
│   │   └── response.go
│   └── types/
│       ├── request.go
│       └── response.go
├── newprovider.go
```

### 步骤 2: 创建提供商实现

创建一个新的 Go 文件，例如 `adapter/newprovider.go`，实现 `Provider` 接口：

```go
package adapter

import (
    "encoding/json"
    "log/slog"

    "github.com/MeowSalty/portal/adapter/newprovider/converter"
    newproviderTypes "github.com/MeowSalty/portal/adapter/newprovider/types"
    coreTypes "github.com/MeowSalty/portal/types"
)

// NewProvider 新提供商实现
type NewProvider struct {
    logger *slog.Logger
}

// init 函数注册新提供商
func init() {
    RegisterProviderFactory("NewProvider", func(logger *slog.Logger) Provider {
        return NewNewProvider(logger)
    })
}

// NewNewProvider 创建新的提供商实例
func NewNewProvider(logger *slog.Logger) *NewProvider {
    if logger == nil {
        logger = slog.Default()
    }
    return &NewProvider{
        logger: logger.WithGroup("newprovider"),
    }
}

// Name 返回提供商名称
func (p *NewProvider) Name() string {
    return "newprovider"
}

// CreateRequest 创建新提供商请求
func (p *NewProvider) CreateRequest(request *coreTypes.Request, channel *coreTypes.Channel) (interface{}, error) {
    return converter.ConvertRequest(request, channel), nil
}

// ParseResponse 解析新提供商响应
func (p *NewProvider) ParseResponse(responseData []byte) (*coreTypes.Response, error) {
    var response newproviderTypes.Response
    if err := json.Unmarshal(responseData, &response); err != nil {
        return nil, err
    }
    return converter.ConvertCoreResponse(&response), nil
}

// ParseStreamResponse 解析新提供商流式响应
func (p *NewProvider) ParseStreamResponse(responseData []byte) (*coreTypes.Response, error) {
    var chunk newproviderTypes.Response
    if err := json.Unmarshal(responseData, &chunk); err != nil {
        return nil, err
    }
    return converter.ConvertCoreResponse(&chunk), nil
}

// APIEndpoint 返回 API 端点
func (p *NewProvider) APIEndpoint() string {
    return "/v1/chat/completions" // 根据实际 API 端点调整
}

// Headers 返回特定头部
func (p *NewProvider) Headers(channel *coreTypes.Channel) map[string]string {
    headers := map[string]string{
        "Authorization": "Bearer " + channel.APIKey.Value,
        "Content-Type":  "application/json",
        // 添加提供商特定的头部
    }
    return headers
}

// SupportsStreaming 是否支持流式传输
func (p *NewProvider) SupportsStreaming() bool {
    return true // 根据实际情况调整
}
```

### 步骤 3: 创建类型定义

在 `adapter/newprovider/types/` 目录下创建类型定义文件：

#### `request.go`

```go
package types

// Request 新提供商的请求结构
type Request struct {
    Model       string    `json:"model"`
    Messages    []Message `json:"messages"`
    MaxTokens   int       `json:"max_tokens,omitempty"`
    Temperature float64   `json:"temperature,omitempty"`
    Stream      bool      `json:"stream,omitempty"`
    // 添加其他提供商特定的字段
}

// Message 消息结构
type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}
```

#### `response.go`

```go
package types

// Response 新提供商的响应结构
type Response struct {
    ID      string   `json:"id"`
    Object  string   `json:"object"`
    Created int64    `json:"created"`
    Model   string   `json:"model"`
    Choices []Choice `json:"choices"`
    Usage   Usage    `json:"usage"`
}

// Choice 选择项
type Choice struct {
    Index        int     `json:"index"`
    Message      Message `json:"message"`
    FinishReason string  `json:"finish_reason"`
}

// Usage 使用情况
type Usage struct {
    PromptTokens     int `json:"prompt_tokens"`
    CompletionTokens int `json:"completion_tokens"`
    TotalTokens      int `json:"total_tokens"`
}
```

### 步骤 4: 创建转换器

在 `adapter/newprovider/converter/` 目录下创建转换器：

#### `request.go`

```go
package converter

import (
    coreTypes "github.com/MeowSalty/portal/types"
    newproviderTypes "github.com/MeowSalty/portal/adapter/newprovider/types"
)

// ConvertRequest 将核心请求转换为新提供商特定请求
func ConvertRequest(request *coreTypes.Request, channel *coreTypes.Channel) *newproviderTypes.Request {
    messages := make([]newproviderTypes.Message, len(request.Messages))
    for i, msg := range request.Messages {
        messages[i] = newproviderTypes.Message{
            Role:    msg.Role,
            Content: msg.Content,
        }
    }

    return &newproviderTypes.Request{
        Model:       request.Model,
        Messages:    messages,
        MaxTokens:   request.MaxTokens,
        Temperature: request.Temperature,
        Stream:      request.Stream,
    }
}
```

#### `response.go`

```go
package converter

import (
    coreTypes "github.com/MeowSalty/portal/types"
    newproviderTypes "github.com/MeowSalty/portal/adapter/newprovider/types"
)

// ConvertCoreResponse 将新提供商响应转换为核心响应
func ConvertCoreResponse(response *newproviderTypes.Response) *coreTypes.Response {
    choices := make([]coreTypes.Choice, len(response.Choices))
    for i, choice := range response.Choices {
        choices[i] = coreTypes.Choice{
            Index:   choice.Index,
            Message: coreTypes.Message{
                Role:    choice.Message.Role,
                Content: choice.Message.Content,
            },
            FinishReason: choice.FinishReason,
        }
    }

    return &coreTypes.Response{
        ID:      response.ID,
        Object:  response.Object,
        Created: response.Created,
        Model:   response.Model,
        Choices: choices,
        Usage: coreTypes.Usage{
            PromptTokens:     response.Usage.PromptTokens,
            CompletionTokens: response.Usage.CompletionTokens,
            TotalTokens:      response.Usage.TotalTokens,
        },
    }
}
```

## 故障排除

### 常见问题

1. **提供商未注册**

   - 确保在 `init()` 函数中调用 `RegisterProviderFactory`

2. **HTTP 错误**

   - 检查 `Headers()` 方法是否正确实现
   - 验证 API 密钥和端点配置

3. **解析错误**

   - 确保响应结构与提供商 API 文档匹配
   - 检查 JSON 标签是否正确

4. **流式传输问题**
   - 验证 `SupportsStreaming()` 返回正确值
   - 检查流式响应格式

## 扩展性

Adapter 包设计为高度可扩展，可以轻松添加新的 AI 服务提供商。通过实现标准的 `Provider` 接口，新的提供商可以无缝集成到现有系统中。

如需进一步定制，可以考虑扩展 `Adapter` 结构体或添加新的接口方法来支持特定提供商的高级功能。
