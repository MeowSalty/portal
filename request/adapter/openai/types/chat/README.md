# Chat Completions 自述文档

本文档描述 `request/adapter/openai/types/chat` 下的 Chat Completions 数据结构与能力范围，并对齐 OpenAI 官方 OpenAPI 文档中的 `/chat/completions` 定义。内容包括请求、非流式响应、流式响应的字段说明、规范要求、事件顺序与完整性约束，以及与官方规范的差异与支持情况。

---

## 1. 范围与对齐基准

- 覆盖类型：`Request`、`Response`、`StreamEvent` 及其关联结构。
- 对齐基准：`request/adapter/openai/types/docs/openapi.documented.yml`。
- 兼容策略：
  - 未知字段透传：`Request`、`RequestMessage`、`Message`、`Delta` 支持 `ExtraFields` 透传。
  - deprecated 字段保留兼容，但不推荐使用（如 `max_tokens`、`seed`、`function_call`、`functions`）。
  - 官方文档字段按规范对齐，扩展能力仅在工具类型等非标准范围内标注。

---

## 2. 请求（Request）

### 2.1 顶层字段

**必填字段**：

- `model`：模型标识。
- `messages`：消息列表。

**可选字段（按实现结构列出）**：

- `stream`：是否流式返回。
- `frequency_penalty`：频率惩罚。
- `logprobs`：是否返回对数概率。
- `max_completion_tokens`：最大完成 token 数。
- `max_tokens`：最大 token 数（deprecated）。
- `n`：生成候选数量。
- `presence_penalty`：存在惩罚。
- `seed`：随机种子（deprecated）。
- `store`：是否存储。
- `temperature`：采样温度。
- `top_logprobs`：top logprobs 数量（需 `logprobs=true`）。
- `top_p`：核采样。
- `parallel_tool_calls`：并行工具调用开关。
- `prompt_cache_key`：提示缓存键。
- `prompt_cache_retention`：提示缓存保留策略。
- `safety_identifier`：安全标识符。
- `user`：用户标识符。
- `audio`：音频输出配置。
- `logit_bias`：对数偏置映射（`map[string]int`）。
- `metadata`：元数据（`map[string]string`）。
- `modalities`：模态，支持 `text`、`audio`。
- `reasoning_effort`：推理努力程度（共享枚举）。
- `service_tier`：服务层级（共享枚举）。
- `stop`：停止条件（字符串或字符串数组）。
- `stream_options`：流式选项。
- `verbosity`：详细程度（共享枚举）。
- `function_call`：函数调用控制（deprecated）。
- `functions`：函数列表（deprecated）。
- `prediction`：预测输出配置。
- `response_format`：结构化输出格式。
- `tool_choice`：工具选择策略。
- `tools`：工具列表。
- `web_search_options`：网络搜索配置。

**未知字段**：

- 所有未被定义的字段将进入 `ExtraFields`，并在序列化时原样透传。

### 2.2 关键子结构

#### 2.2.1 消息结构（RequestMessage）

- `role`：`developer`/`system`/`user`/`assistant`/`tool`/`function`。
- `content`：字符串或内容片段数组。
- `name`：消息名称。
- `tool_call_id`：工具调用 ID（常用于 `tool` role）。
- `tool_calls`：工具调用数组（请求侧）。
- `function_call`：函数调用（deprecated）。
- `refusal`：拒绝内容。
- `audio`：助手音频引用（`{ id }`）。

#### 2.2.2 内容结构（MessageContent / ContentPart）

- `content` 支持：
  - `string`
  - `[]ContentPart`
- `ContentPart.type` 支持：`text`/`image_url`/`input_audio`/`file`/`refusal`。
- `image_url.url` 可选 `detail`（共享枚举）。
- `input_audio.format` 仅支持 `wav`、`mp3`。
- `file` 支持 `filename`/`file_data`/`file_id`。

#### 2.2.3 工具与工具选择

- `tools`：支持 `function`、`custom` 工具类型。
- `tool_choice`：`auto`/`none`/`required` 或对象形态（指定 `type` 和目标工具）。
- `function_call`/`functions`：deprecated，仅为兼容旧字段。

#### 2.2.4 结构化输出与预测输出

- `response_format`：`text`/`json_schema`/`json_object`。
- `prediction`：`{ type: "content", content: string | []ContentPart(text) }`。

#### 2.2.5 Web 搜索与流式选项

- `web_search_options.user_location`：仅支持 `approximate` 类型，包含 `city`、`country`、`region`、`timezone`。
- `web_search_options.search_context_size`：`small`/`medium`/`large`。
- `stream_options.include_usage`：是否包含用量统计。
- `stream_options.include_obfuscation`：流混淆。

### 2.3 约束与序列化规则

- `top_logprobs` 仅在 `logprobs=true` 时生效。
- `content` 可为 `null`，但应遵守消息角色语义。
- `ExtraFields` 透传未知字段，避免丢失官方新增字段。

---

## 3. 非流式响应（Response）

### 3.1 顶层字段

- `id`：完成 ID。
- `object`：固定为 `chat.completion`。
- `created`：创建时间戳（秒）。
- `model`：模型名称。
- `choices`：候选列表。
- `service_tier`：服务层级。
- `system_fingerprint`：系统指纹（deprecated）。
- `usage`：用量信息。

### 3.2 关键子结构

- `Choice`：
  - `index`：候选索引。
  - `finish_reason`：`stop`/`length`/`tool_calls`/`content_filter`/`function_call`。
  - `logprobs`：包含 `content` 与 `refusal` 的 token 概率。
  - `message`：消息体。
- `Message`：
  - `role` 固定为 `assistant`。
  - `content` 可为字符串或 `null`。
  - `refusal`、`function_call`（deprecated）、`tool_calls`、`annotations`、`audio`。
  - 未知字段进入 `ExtraFields`。
- `ToolCall`：`type` 支持 `function` 与 `custom`。
- `Usage`：`prompt_tokens`、`completion_tokens`、`total_tokens` 及细节字段。

### 3.3 约束与序列化规则

- `choices` 不可为 `null`，空列表需返回 `[]`。
- `logprobs`、`usage` 为可选字段。

---

## 4. 流式响应（StreamEvent）

### 4.1 顶层字段

- `id`：完成 ID（同一流内固定）。
- `object`：固定为 `chat.completion.chunk`。
- `created`：创建时间戳（同一流内固定）。
- `model`：模型名称。
- `choices`：增量候选列表。
- `service_tier`、`system_fingerprint`、`usage`：与非流式一致。

### 4.2 事件类型与字段

- `StreamChoice`：
  - `delta`：增量字段（`role`、`content`、`refusal`、`tool_calls`、`function_call`）。
  - `logprobs`：增量 logprobs。
  - `finish_reason`：可能为 `null`。
- `ToolCallChunk`：支持分段拼接 `arguments`，`name` 可能在后续段出现。

### 4.3 事件顺序与完整性要求

- 稳定字段一致性：`id`/`created`/`model`/`object` 必须一致。
- 角色与内容顺序：首个 chunk 通常包含 `delta.role=assistant`。
- 工具调用完整性：同一 `tool_calls[index]` 的 `arguments` 可分段拼接，必要时补齐 `name`。
- 完成信号：每个 `index` 在结束时出现一次 `finish_reason`；结束 chunk 可能为空 `delta`。
- usage 返回规则：仅当 `stream_options.include_usage=true` 时出现；除最后一个 chunk 外应为 `null`。
- 多候选并行：`n>1` 时允许交错返回，但索引需完整。
- SSE 终止：`[DONE]` 为传输层语义，不在结构体中建模。

---

## 5. 支持与对齐矩阵（✅/⚠️/❌）

> ✅ 完全对齐，⚠️ 部分对齐（deprecated/行为差异/强类型不足），❌ 官方有但未实现或官方无但扩展支持。

### 5.1 请求对齐

| 字段/能力 | 官方规范 | 本实现 | 说明 |
| --- | --- | --- | --- |
| `model` | ✅ | ✅ | 必填 |
| `messages` | ✅ | ✅ | 必填 |
| `stream` | ✅ | ✅ | `true` 时返回流式响应 |
| `stream_options.include_usage` | ✅ | ✅ | 仅最后一个 chunk 有用量 |
| `logprobs` / `top_logprobs` | ✅ | ✅ | 需 `logprobs=true` |
| `temperature` / `top_p` | ✅ | ✅ | 同官方语义 |
| `stop` | ✅ | ✅ | 字符串或字符串数组 |
| `n` | ✅ | ✅ | 多候选并行 |
| `max_completion_tokens` | ✅ | ✅ | 推荐使用 |
| `max_tokens` | ⚠️ deprecated | ✅ | 保留兼容 |
| `seed` | ⚠️ deprecated | ✅ | 保留兼容 |
| `response_format` | ✅ | ✅ | `text`/`json_schema`/`json_object` |
| `tools`（function/custom） | ✅ | ✅ | 官方定义范围内 |
| `tool_choice` | ✅ | ✅ | 支持字符串/对象形态 |
| `function_call` / `functions` | ⚠️ deprecated | ✅ | 兼容旧字段 |
| `web_search_options` | ✅ | ✅ | 与官方 Chat Completions 对齐 |
| `prompt_cache_key` | ✅ | ✅ | 官方字段 |
| `prompt_cache_retention` | ✅ | ✅ | 官方字段 |
| `safety_identifier` | ✅ | ✅ | 官方字段 |
| `stream_options.include_obfuscation` | ✅ | ✅ | 官方字段 |

### 5.2 非流式响应对齐

| 字段/能力 | 官方规范 | 本实现 | 说明 |
| --- | --- | --- | --- |
| `object=chat.completion` | ✅ | ✅ | 非流式对象 |
| `choices.message` | ✅ | ✅ | `role=assistant` |
| `usage` | ✅ | ✅ | 与 `CompletionUsage` 对齐 |
| `system_fingerprint` | ⚠️ deprecated | ✅ | 兼容返回 |
| `tool_calls`（custom） | ⚠️ partial | ✅ | 依赖上游实现 |
| 未知字段透传 | ✅ | ✅ | `ExtraFields` |

### 5.3 流式响应对齐

| 字段/能力 | 官方规范 | 本实现 | 说明 |
| --- | --- | --- | --- |
| `object=chat.completion.chunk` | ✅ | ✅ | 流式对象 |
| `delta.role/content/tool_calls` | ✅ | ✅ | 递增片段 |
| `finish_reason` | ✅ | ✅ | 最终 chunk 提供 |
| `usage` | ✅ | ✅ | 依赖 `include_usage` |
| 未知字段透传 | ✅ | ✅ | `ExtraFields` |

---

## 6. 差异与兼容策略

- deprecated 字段：`max_tokens`、`seed`、`function_call`、`functions` 保留兼容但不推荐使用。

---

## 7. 最小示例轮廓（结构级）

- 请求必须包含 `model` 与 `messages`。
- 非流式响应必须包含 `id`、`object`、`created`、`model`、`choices`。
- 流式响应必须包含 `id`、`object`、`created`、`model`、`choices`，并在结束时返回 `finish_reason`。
