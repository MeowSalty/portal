# Portal

Portal 是一个用 Go 语言编写的 AI 网关管理模块，提供多平台 AI 服务的统一接入、负载均衡、健康检查和请求统计功能。该模块旨在简化对多个 AI 平台的访问，提供高可用性和智能路由功能。

## 功能特性

- **统一接入**: 支持多种 AI 平台（虽然现在只有 OpenAI）的统一接入
- **负载均衡**: 提供多种通道选择策略（随机、最近最少使用等）
- **健康检查**: 实时监控各平台和通道的健康状态
- **请求统计**: 记录和统计请求数据，便于分析和监控
- **优雅停机**: 支持优雅停机，确保正在处理的请求能正常完成

## 安装

使用 Go modules 安装 Portal：

```bash
go get github.com/MeowSalty/portal
```

## 使用方法

### 1. 初始化 GatewayManager

```go
import (
    "context"
    "log/slog"
    "time"

    "github.com/MeowSalty/portal"
    "github.com/MeowSalty/portal/types"
)

// 创建配置
config := &portal.Config{
    Repo:  yourDataRepository, // 实现 types.DataRepository 接口
    Logger: slog.Default(),
}

// 创建 GatewayManager 实例
ctx := context.Background()
manager, err := portal.New(ctx,
    portal.WithRepository(yourDataRepository),
    portal.WithSelectorStrategy(portal.RandomSelectorStrategy), // 或 portal.LRUSelectorStrategy
    portal.WithHealthSyncInterval(time.Minute),
    portal.WithLogger(slog.Default()),
)
if err != nil {
    log.Fatal("Failed to create GatewayManager:", err)
}
```

### 2. 处理请求

```go
// 创建请求
request := &types.Request{
    Model: "gpt-3.5-turbo",
    Messages: []types.Message{
        {
            Role:    "user",
            Content: "Hello, world!",
        },
    },
}

// 处理聊天完成请求
response, err := manager.ChatCompletion(context.Background(), request)
if err != nil {
    log.Fatal("Failed to process request:", err)
}

// 处理流式聊天完成请求
stream, err := manager.ChatCompletionStream(context.Background(), request)
if err != nil {
    log.Fatal("Failed to process stream request:", err)
}

for resp := range stream {
    // 处理流式响应
    fmt.Print(resp.Choices[0].Message.Content)
}
```

### 3. 查询统计信息

```go
// 查询请求统计
stats, err := manager.QueryStats(context.Background(), &types.StatsQueryParams{
    Limit: 100,
})
if err != nil {
    log.Fatal("Failed to query stats:", err)
}

// 统计请求计数
summary, err := manager.CountStats(context.Background(), &types.StatsQueryParams{})
if err != nil {
    log.Fatal("Failed to count stats:", err)
}
```

### 4. 优雅停机

```go
// 优雅停机，等待最多 30 秒
err := manager.Shutdown(30 * time.Second)
if err != nil {
    log.Fatal("Shutdown error:", err)
}
```

## 包结构

```tree
portal/
├── adapter/           # 适配器相关
│   ├── openai/        # OpenAI 适配器实现
│   └── registry.go    # 适配器注册机制
├── health/            # 健康检查模块
├── selector/          # 通道选择策略
├── stats/             # 统计模块
├── types/             # 数据类型定义
├── processor/         # 请求处理逻辑（新增）
├── channel/           # 通道管理（新增）
├── config.go          # 配置定义
└── portal.go          # 核心管理器
```

## 核心概念

### 适配器 (Adapter)

适配器负责与特定 AI 平台的 API 交互，目前支持 OpenAI 格式的适配器。

### 选择器 (Selector)

选择器决定在多个可用通道中选择哪个通道来处理请求，支持随机选择和最近最少使用 (LRU) 选择策略。

### 健康管理 (Health)

健康管理系统监控各平台和通道的健康状态，确保请求只会被发送到健康的通道。

### 统计 (Stats)

统计模块记录和分析请求数据，提供查询和统计功能。
