# Anthropic Messages 自述文档

本文档描述 `request/adapter/anthropic/types` 目录下对 Anthropic Messages API 的**请求**、**非流式响应**与**流式响应**结构的实现支持情况、字段约束与事件语义，并对照官方文档进行对齐说明。

适用范围：仅覆盖类型层面的结构、序列化/反序列化约束与事件语义，不包含转换器或业务逻辑。

---

## 1. 范围与对齐基准

- 覆盖类型：`Request`、`Response`、`StreamEvent` 及其关联结构。
- 对齐基准：
  - `request/adapter/anthropic/types/docs/messages.md` 中的 `Messages` 与 `Message`、`ContentBlockParam` 等定义。
  - `request/adapter/anthropic/types/docs/streaming.md` 中的 `Streaming Messages` 事件类型与顺序说明。
- 兼容策略：
  - 未知字段透传：`Request` 顶层支持 `ExtraFields` 透传未知字段，子结构（如 `Message`、`ContentBlock`）不透传。
  - 显式字段优先：未知字段与显式字段同名时，以显式字段值为准。
  - deprecated 字段保留兼容：以官方文档为准，本实现不主动引入额外 deprecated 字段。

---

## 2. 请求（Request）

### 2.1 顶层字段

**必填字段（官方语义）**：

- `model`：模型名称。
- `messages`：输入消息列表。
- `max_tokens`：最大生成 token。

**可选字段（按实现结构）**：

- `metadata`：请求元数据，包含 `user_id`。
- `service_tier`：服务层级，支持 `auto` / `standard_only`。
- `stop_sequences`：停止序列。
- `stream`：是否开启流式响应。
- `system`：system prompt，支持 string 或 `[]TextBlockParam`。
- `temperature`、`top_k`、`top_p`：采样参数。
- `thinking`：扩展思考配置，支持 `enabled` / `disabled`。
- `tools`：工具定义列表。
- `tool_choice`：工具选择策略。

**未知字段**：

- 提供 `ExtraFields` 透传容器，未知字段会被收集并保留（仅顶层）。

### 2.2 关键子结构

#### 2.2.1 `Message`

- `role`：仅支持 `user` / `assistant`。
- `content`：支持 string 或 `[]ContentBlockParam`。

#### 2.2.2 `ContentBlockParam`（请求内容块）

内容块为 oneof 类型，仅允许设置一种：

- `text`：`TextBlockParam`。
- `image`：`ImageBlockParam`，支持 `base64` 或 `url`。
- `document`：`DocumentBlockParam`，支持 `base64` / `text` / `content` / `url` 来源。
- `search_result`：`SearchResultBlockParam`。
- `thinking`：`ThinkingBlockParam`。
- `redacted_thinking`：`RedactedThinkingBlockParam`。
- `tool_use`：`ToolUseBlockParam`。
- `tool_result`：`ToolResultBlockParam`。
- `server_tool_use`：`ServerToolUseBlockParam`。
- `web_search_tool_result`：`WebSearchToolResultBlockParam`。

#### 2.2.3 文本与引用

- `TextBlockParam.citations` 支持五类引用：
  - `char_location` / `page_location` / `content_block_location` / `web_search_result_location` / `search_result_location`。
- `CitationsConfigParam` 支持 `enabled` 作为引用开关。

#### 2.2.4 工具与工具选择

- `tools` 支持 `custom` 与多版本内置工具：`bash_20250124`、`text_editor_20250124`、`text_editor_20250429`、`text_editor_20250728`、`web_search_20250305`。
- `tool_choice` 支持 `auto` / `any` / `tool` / `none`，并可设置 `disable_parallel_tool_use`。

### 2.3 约束与序列化规则

- oneof 约束：
  - `MessageContentParam`、`SystemParam`、`ContentBlockParam`、`DocumentSource`、`ImageSource`、`ToolResultContentParam` 等均只允许设置一种具体类型。
- `content`、`system`、`tool_result.content` 等字段序列化时优先使用 string，若设置了 string 则忽略数组。
- 工具选择与思考配置均采用严格枚举，解析时遇到未知类型会返回错误。

---

## 3. 非流式响应（Response）

### 3.1 顶层字段

- `id`：消息 ID。
- `type`：固定为 `message`。
- `role`：固定为 `assistant`。
- `content`：响应内容块数组。
- `model`：模型名称。
- `stop_reason`：停止原因，可能为 `end_turn` / `max_tokens` / `stop_sequence` / `tool_use` / `pause_turn` / `refusal`。
- `stop_sequence`：命中的停止序列。
- `usage`：使用统计（可选）。

### 3.2 关键子结构

- `ResponseContentBlock` 为 oneof：
  - `text` / `thinking` / `redacted_thinking` / `tool_use` / `server_tool_use` / `web_search_tool_result`。
- `TextCitation` 引用联合类型与请求侧一致。
- `Usage` 提供输入/输出 token 及 cache、server tool 等细分字段，均为指针以区分缺失与 0。

### 3.3 约束与序列化规则

- oneof 约束：`ResponseContentBlock` 与 `TextCitation` 仅允许设置一种具体类型。
- `content` 必须为数组，空值应为 `[]` 而非 `null`（由调用方保证）。

---

## 4. 流式响应（StreamEvent）

### 4.1 顶层字段

- `StreamEvent` 为 oneof，支持：
  - `message_start` / `message_delta` / `message_stop`
  - `content_block_start` / `content_block_delta` / `content_block_stop`
  - `ping` / `error`

### 4.2 事件类型与字段

- 生命周期：`message_start` → `message_delta` → `message_stop`。
- 内容块：`content_block_start` → `content_block_delta`\* → `content_block_stop`。
- 增量类型：
  - `text_delta`
  - `input_json_delta`（工具输入 JSON 片段）
  - `thinking_delta`
  - `signature_delta`
  - `citations_delta`
- 错误事件：`error` 携带 `ErrorResponse`。

### 4.3 事件顺序与完整性要求

- 同一流内 `message_start` 的 `message.id` / `model` / `type` 在后续事件保持一致。
- 每个内容块必须完整经历 `start` → `delta`\* → `stop`。
- `input_json_delta.partial_json` 支持分段拼接，完成后需在 `content_block_stop` 处收束。
- `message_delta.usage` 为累积值；当未返回 `usage` 时视为未提供。
- 出现 `error` 事件时视为流异常终止。

---

## 5. 支持与对齐矩阵（✅/⚠️/❌）

> ✅ 完全对齐，⚠️ 部分对齐（deprecated/行为差异/强类型不足），❌ 官方有但未实现或官方无但扩展支持。

### 5.1 请求对齐

| 字段/能力                                      | 官方规范 | 本实现 | 说明                                                  |
| ---------------------------------------------- | -------- | ------ | ----------------------------------------------------- |
| `model` / `messages` / `max_tokens`            | ✅       | ✅     | 必填字段对齐。                                        |
| `system`                                       | ✅       | ✅     | string 或文本块数组。                                 |
| `metadata` / `service_tier` / `stop_sequences` | ✅       | ✅     | 枚举与字段对齐。                                      |
| `thinking`                                     | ✅       | ✅     | `enabled` / `disabled` 联合类型。                     |
| `tools` / `tool_choice`                        | ✅       | ✅     | 工具与策略结构对齐。                                  |
| 未知字段透传                                   | ✅       | ✅     | `Request` 顶层支持 `ExtraFields` 透传，显式字段优先。 |

### 5.2 非流式响应对齐

| 字段/能力          | 官方规范 | 本实现 | 说明                            |
| ------------------ | -------- | ------ | ------------------------------- |
| `Message` 响应主体 | ✅       | ✅     | `type` 固定为 `message`。       |
| `content` 内容块   | ✅       | ✅     | 覆盖文本、思考、工具等块类型。  |
| `usage` 细分字段   | ✅       | ✅     | cache 与 server tool 统计对齐。 |

### 5.3 流式响应对齐

| 字段/能力                  | 官方规范 | 本实现 | 说明                                                                  |
| -------------------------- | -------- | ------ | --------------------------------------------------------------------- |
| 事件集合                   | ✅       | ✅     | 覆盖 `message_*`、`content_block_*`、`ping`、`error`。                |
| `content_block_delta` 类型 | ✅       | ✅     | 支持 `text` / `input_json` / `thinking` / `signature` / `citations`。 |
| `usage` 累积规则           | ✅       | ✅     | 仅在 `message_delta` 中返回。                                         |

---

## 6. 差异与兼容策略

- 未知字段：`Request` 顶层提供 `ExtraFields` 透传，未知字段会被收集并保留；子结构（如 `Message`、`ContentBlock`）不透传。未知字段与显式字段同名时，以显式字段值为准。
- 严格 oneof：请求与流式增量类型均强制互斥，解析时不允许多类型并存。
- 响应内容块：未实现图像输出块类型（官方消息响应亦以文本/工具/思考为主）。

---

## 7. 最小示例轮廓（结构级）

- 请求必须包含：`model`、`messages`、`max_tokens`。
- 非流式响应必须包含：`id`、`type`、`role`、`content`、`model`。
- 流式响应必须包含：`message_start` → 若干 `content_block_*` → `message_delta` → `message_stop`。
