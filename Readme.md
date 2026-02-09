# Portal

Portal 是一个用 Go 语言编写的 AI 网关管理模块，提供多平台 AI 服务的统一接入、负载均衡、健康检查和请求统计功能。该模块旨在简化对多个 AI 平台的访问，提供高可用性和智能路由功能。

## 功能特性

- **统一接入**: 支持多种 AI 平台（包括 OpenAI 和 Gemini）的统一接入
- **智能路由**: 基于模型名称和健康状态自动选择最佳通道
- **健康检查**: 实时监控各平台和通道的健康状态
- **会话管理**: 提供请求会话生命周期管理和优雅停机
- **错误处理**: 结构化错误码和上下文信息
- **性能监控**: 请求统计和性能指标收集

## 安装

使用 Go modules 安装 Portal：

```bash
go get github.com/MeowSalty/portal
```

## 使用方法

### 1. 初始化 Portal

```go
import (
    "context"
    "log/slog"
    "time"

    "github.com/MeowSalty/portal"
    "github.com/MeowSalty/portal/routing"
    "github.com/MeowSalty/portal/request"
    "github.com/MeowSalty/portal/routing/health"
)

// 创建配置
cfg := &portal.Config{
    PlatformRepo: yourPlatformRepo,    // 实现 routing.PlatformRepository
    ModelRepo:    yourModelRepo,      // 实现 routing.ModelRepository
    KeyRepo:      yourKeyRepo,       // 实现 routing.KeyRepository
    HealthRepo:   yourHealthRepo,    // 实现 health.HealthRepository
    LogRepo:      yourLogRepo,       // 实现 request.RequestLogRepository
}

// 创建 Portal 实例
portal, err := portal.New(cfg)
if err != nil {
    log.Fatal("Failed to create Portal:", err)
}
```

### 2. 处理请求

```go
// 创建请求
request := &types.RequestContract{
    Model: "gpt-3.5-turbo",
    Messages: []types.Message{
        {
            Role:    "user",
            Content: "Hello, world!",
        },
    },
}

// 处理聊天完成请求
response, err := portal.ChatCompletion(context.Background(), request)
if err != nil {
    log.Fatal("Failed to process request:", err)
}

// 处理流式聊天完成请求
stream, err := portal.ChatCompletionStream(context.Background(), request)
if err != nil {
    log.Fatal("Failed to process stream request:", err)
}

for resp := range stream {
    // 处理流式响应
    fmt.Print(resp.Choices[0].Message.Content)
}
```

### 3. 优雅停机

```go
// 优雅停机，等待最多 30 秒
err := portal.Close(30 * time.Second)
if err != nil {
    log.Fatal("Shutdown error:", err)
}
```

## 包结构

```tree
portal/
├── request/           # 请求处理模块
│   ├── adapter/       # 平台适配器实现
│   │   ├── openai/    # OpenAI 适配器
│   │   ├── gemini/    # Gemini 适配器
│   │   └── registry.go # 适配器注册机制
│   └── request.go     # 核心请求处理逻辑
├── routing/            # 路由管理模块
│   ├── health/        # 健康检查实现
│   ├── selector/      # 通道选择策略
│   └── routing.go     # 核心路由逻辑
├── session/           # 会话管理模块
├── types/             # 数据类型定义
├── errors/            # 错误处理模块
└── portal.go          # 核心入口
```

## 核心概念

### 适配器 (Adapter)

适配器负责与特定 AI 平台的 API 交互，目前支持 OpenAI 和 Gemini 格式的适配器。

### 路由 (Routing)

路由模块负责根据模型名称查找可用通道，并基于健康状态选择最佳通道。

### 会话 (Session)

会话管理模块处理请求的生命周期，包括优雅停机和上下文取消。

### 错误处理 (Error Handling)

提供结构化错误码和上下文信息，便于问题诊断和监控。

### 性能监控 (Performance Monitoring)

收集和分析请求统计信息，提供性能洞察和优化建议。
