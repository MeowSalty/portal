# OpenAI Responses 自述文档

本文档描述 `request/adapter/openai/types/responses` 目录下对 OpenAI Responses API 的**请求**、**非流式响应**与**流式响应**结构的实现支持情况、字段约束与事件语义，并对照官方 OpenAPI 文档进行对齐说明。

适用范围：仅覆盖类型层面的结构、序列化/反序列化约束与事件语义，不包含转换器或业务逻辑。

---

## 1. 范围与对齐基准

- 覆盖类型：`Request`、`Response`、`StreamEvent` 及其关联结构。
- 对齐基准：`request/adapter/openai/types/docs/openapi.documented.yml` 中的 `CreateResponse`、`InputParam`、`ResponseProperties`、`ResponseStreamEvent`、`ResponseStreamOptions` 等定义。
- 兼容策略：
  - 未知字段透传：`Request.ExtraFields` 捕获并原样序列化。
  - deprecated 字段保留兼容：如 `user` 字段仍可读写但不推荐使用。

---

## 2. 请求（Request）

### 2.1 顶层字段

**必填字段（官方语义）**：

- `model`：模型 ID。
- `input`：输入内容，支持字符串或输入项数组。

**可选字段（按实现结构）**：

- `stream`：是否开启流式响应。
- `stream_options`：流式选项，支持 `include_obfuscation`。
- `max_output_tokens`、`temperature`、`top_p`、`top_logprobs`：采样与输出控制。
- `tools`、`tool_choice`、`parallel_tool_calls`、`max_tool_calls`：工具调用能力。
- `truncation`、`text`、`store`、`include`、`metadata`。
- `instructions`、`reasoning`、`prompt`、`conversation`、`previous_response_id`。
- `safety_identifier`、`user`（已弃用）、`prompt_cache_key`、`prompt_cache_retention`、`service_tier`、`background`。

**未知字段**：

- 所有未被定义的字段进入 `ExtraFields` 并在序列化时原样透传。

### 2.2 关键子结构

#### 2.2.1 `input` 联合类型

- **字符串**：直接文本输入。
- **数组**：`[]InputItem`，支持多种输入项类型。

#### 2.2.2 `InputItem` 类型枚举（oneof）

- `message`、`item_reference`。
- 工具调用：`function_call` / `file_search_call` / `web_search_call` / `computer_call` / `code_interpreter_call` / `image_generation_call` / `local_shell_call` / `shell_call` / `apply_patch_call` / `mcp_call` / `custom_tool_call` / `mcp_list_tools`。
- 工具输出：`function_call_output` / `computer_call_output` / `local_shell_call_output` / `shell_call_output` / `apply_patch_call_output` / `custom_tool_call_output`。
- 其他：`reasoning` / `compaction` / `mcp_approval_request` / `mcp_approval_response`。

#### 2.2.3 `InputMessage` 与 `InputMessageContent`

- `role`：`user` / `system` / `developer` / `assistant`。
- `content`：支持 **string** 或 **[]InputContent**。
- `type` 可省略：当输入项缺少 `type`，默认按 `message` 解析，兼容 EasyInputMessage。

#### 2.2.4 `InputContent`（输入内容片段）

- `input_text`：文本输入。
- `input_image`：图片输入，`detail` 缺省时默认 `auto`。
- `input_file`：文件输入。
- `input_audio`：未实现。

### 2.3 约束与序列化规则

- oneof 约束：`InputItem` / `InputContent` / `OutputContentPart` / `OutputMessageContent` 只能设置一种类型。
- `InputMessageContent` 序列化时优先使用 string，若同时设置 string 与 list，string 覆盖 list。
- `InputImageContent.detail` 缺省时自动补齐为 `auto`。
- `Reasoning.summary` 与 `Reasoning.generate_summary` **互斥**：
  - 同时设置会返回错误。
  - 两者都未设置时序列化输出 `summary: null`。
- `PromptTemplate.variables` 支持 `string` / `input_text` / `input_image` / `input_file`。
- `Request.ExtraFields` 用于未知字段透传。

---

## 3. 非流式响应（Response）

### 3.1 顶层字段

- `id`：响应唯一 ID（必填）。
- `object`：固定为 `response`（必填）。
- `created_at`：创建时间戳（秒，必填）。
- `model`：模型 ID（必填）。
- `output`：输出项数组（必填）。
- `parallel_tool_calls`：是否并行调用工具（必填）。
- `metadata`：元数据（必填）。
- `tools` / `tool_choice`：工具定义与策略（必填）。

可选字段：`status` / `completed_at` / `error` / `incomplete_details` / `instructions` / `usage` / `conversation` / `previous_response_id` / `reasoning` / `background` / `max_output_tokens` / `max_tool_calls` / `text` / `top_p` / `temperature` / `truncation` / `user` / `safety_identifier` / `prompt_cache_key` / `service_tier` / `prompt_cache_retention` / `top_logprobs`。

`instructions` 支持 string / []InputItem / null，与 Response Object 文档对齐。

### 3.2 `output` 输出项（oneof）

`OutputItem` 通过 `type` 判别：

- `message`、`function_call`、`file_search_call`、`web_search_call`、`computer_call`、`reasoning`、`code_interpreter_call`、`image_generation_call`、`local_shell_call`、`shell_call`、`shell_call_output`、`apply_patch_call`、`apply_patch_call_output`、`mcp_call`、`mcp_list_tools`、`mcp_approval_request`、`custom_tool_call`、`compaction`。

### 3.3 输出消息内容与注释

`OutputMessageContent` / `OutputContentPart` 为 oneof：

- `output_text`：文本输出。
- `refusal`：拒绝内容。

约束：

- `output_text.annotations` 序列化时保证为 **空数组 `[]` 而非 `null`**。
- 互斥约束违反会返回错误。

注释（Annotations）支持：`file_citation` / `url_citation` / `container_file_citation` / `file_path`。

### 3.4 `usage`

`usage` 与 `input_tokens_details` / `output_tokens_details` 结构对齐，字段均为必填。

---

## 4. 流式响应（StreamEvent）

### 4.1 顶层字段

- 所有事件均包含 `type` 与 `sequence_number`。
- 同一流内 `response` 的 `id` / `created_at` / `model` / `object` 在 `response.*` 事件中保持一致。

### 4.2 事件类型与字段

按事件分组：

- Response 生命周期：`response.created` / `response.in_progress` / `response.completed` / `response.failed` / `response.incomplete` / `response.queued`。
- Output item：`response.output_item.added` / `response.output_item.done`。
- Content part：`response.content_part.added` / `response.content_part.done`。
- Output text：`response.output_text.delta` / `response.output_text.done` / `response.output_text.annotation.added`。
- Refusal：`response.refusal.delta` / `response.refusal.done`。
- Reasoning text：`response.reasoning_text.delta` / `response.reasoning_text.done`。
- Reasoning summary：`response.reasoning_summary_part.added` / `response.reasoning_summary_part.done` / `response.reasoning_summary_text.delta` / `response.reasoning_summary_text.done`。
- Function call：`response.function_call_arguments.delta` / `response.function_call_arguments.done`。
- Custom tool call：`response.custom_tool_call_input.delta` / `response.custom_tool_call_input.done`。
- MCP：`response.mcp_call_arguments.delta` / `response.mcp_call_arguments.done` / `response.mcp_call.in_progress` / `response.mcp_call.completed` / `response.mcp_call.failed` / `response.mcp_list_tools.in_progress` / `response.mcp_list_tools.completed` / `response.mcp_list_tools.failed`。
- File search：`response.file_search_call.in_progress` / `response.file_search_call.searching` / `response.file_search_call.completed`。
- Web search：`response.web_search_call.in_progress` / `response.web_search_call.searching` / `response.web_search_call.completed`。
- Code interpreter：`response.code_interpreter_call.in_progress` / `response.code_interpreter_call.interpreting` / `response.code_interpreter_call.completed` / `response.code_interpreter_call_code.delta` / `response.code_interpreter_call_code.done`。
- Image generation：`response.image_generation_call.in_progress` / `response.image_generation_call.generating` / `response.image_generation_call.completed` / `response.image_generation_call.partial_image`。
- Audio：`response.audio.delta` / `response.audio.done` / `response.audio.transcript.delta` / `response.audio.transcript.done`。
- Error：`error`。

### 4.3 约束与序列化规则

- `StreamEvent` 为 oneof：序列化时仅允许一个事件字段非空。
- `sequence_number` **全局单调递增**，用于排序与一致性校验。
- `response.output_text.delta` / `response.output_text.done` 的 `logprobs` 为 **必填数组**（允许空数组）。
- 多个 delta 事件可携带 `obfuscation`，与 `stream_options.include_obfuscation` 对齐。

### 4.4 事件顺序与完整性要求

1. **生命周期起始**：`response.created` 先于其他事件。
2. **处理中标记**：`response.in_progress` 可在输出项产生前出现。
3. **输出项建立**：每个 `output_item` 必须先出现 `response.output_item.added`。
4. **内容片段建立**：若输出项包含内容片段，需先发 `response.content_part.added`。
5. **增量与完成**：
   - 文本：`response.output_text.delta`\* → `response.output_text.done`。
   - 拒绝：`response.refusal.delta`\* → `response.refusal.done`。
   - 推理文本：`response.reasoning_text.delta`\* → `response.reasoning_text.done`。
   - 推理摘要文本：`response.reasoning_summary_text.delta`\* → `response.reasoning_summary_text.done`。
   - 函数/自定义/MCP 参数或输入：`*.delta`_ → `_.done`。
6. **输出项完成**：每个 `output_item` 必须以 `response.output_item.done` 结束。
7. **生命周期结束**：以 `response.completed` / `response.failed` / `response.incomplete` 之一收束。

补充要求：

- 允许跨输出项交错，但必须满足 **同一 output item 的局部顺序**（added → delta → done）。
- 若出现 `error` 事件，应视为响应异常终止，不再期待后续事件。
- `response.output_text.annotation.added` 若出现，必须指向已存在的 `output_item` 与 `content_part`。

---

## 5. 支持与对齐矩阵（✅/⚠️/❌）

> ✅ 完全对齐，⚠️ 部分对齐（deprecated/行为差异/强类型不足），❌ 官方有但未实现或官方无但扩展支持。

### 5.1 请求对齐

| 字段/能力                                                    | 官方规范 | 本实现 | 说明                                                                                   |
| ------------------------------------------------------------ | -------- | ------ | -------------------------------------------------------------------------------------- |
| `model`/`input`                                              | ✅       | ✅     | 与 `CreateResponse` 对齐。                                                             |
| `stream`/`stream_options.include_obfuscation`                | ✅       | ✅     | 支持 obfuscation 透传。                                                                |
| `reasoning`                                                  | ✅       | ✅     | `summary` 与 `generate_summary` 互斥。                                                 |
| `tools`/`tool_choice`/`parallel_tool_calls`/`max_tool_calls` | ✅       | ✅     | 结构对齐。                                                                             |
| `include`                                                    | ✅       | ✅     | 枚举值与官方一致。                                                                     |
| `context_management`                                         | ✅       | ❌     | `Request` 顶层字段缺失。                                                               |
| `input_audio`                                                | ✅       | ❌     | `InputContent` 尚未实现。                                                              |
| `image_generation` 工具扩展字段                              | ✅       | ❌     | `ToolImageGen` 缺 `action/background/input_fidelity/input_image_mask/partial_images`。 |
| `mcp.require_approval`                                       | ✅       | ❌     | `ToolMCP` 未包含 `require_approval`。                                                  |
| 未知字段透传                                                 | ⚠️       | ✅     | `ExtraFields` 透传，需上层校验。                                                       |

### 5.2 非流式响应对齐

| 字段/能力             | 官方规范 | 本实现 | 说明                                                            |
| --------------------- | -------- | ------ | --------------------------------------------------------------- |
| `Response` 主体字段   | ✅       | ✅     | 必填字段完整覆盖。                                              |
| `instructions` 强类型 | ✅       | ✅     | 使用 `ResponseInstructions`，支持 string / []InputItem / null。 |
| `usage` 结构          | ✅       | ✅     | 字段均为必填。                                                  |

### 5.3 流式响应对齐

| 字段/能力                      | 官方规范 | 本实现 | 说明                                         |
| ------------------------------ | -------- | ------ | -------------------------------------------- |
| `ResponseStreamEvent` 事件集合 | ✅       | ✅     | 覆盖 MCP、图像生成、代码解释器、音频等事件。 |
| `sequence_number`              | ✅       | ✅     | 全局单调递增要求。                           |
| `logprobs` 必填数组            | ✅       | ✅     | 允许空数组。                                 |
| `obfuscation`                  | ✅       | ✅     | 依赖 `stream_options.include_obfuscation`。  |

---

## 6. 差异与兼容策略

- deprecated 字段：`user` 保留兼容但不推荐使用。
- 未实现字段：`context_management`、`input_audio` 结构当前缺失。
- 工具字段缺失：`ToolImageGen` 未覆盖 `action/background/input_fidelity/input_image_mask/partial_images`，`ToolMCP` 缺 `require_approval`。
- 输入侧工具调用/输出存在偏差：
  - `InputFunctionShellToolCall`/`InputApplyPatchToolCall`：`id`、`status` 被实现为必填，且包含 `created_by`（文档为可选/无该字段）。
  - `InputFunctionToolCallOutput` 缺 `status`。
  - `InputComputerToolCallOutput` 缺 `acknowledged_safety_checks` 与 `status`。
  - `InputLocalShellToolCallOutput` 多出 `call_id` 且 `status` 非可选。
  - `InputFunctionShellToolCallOutput` 缺 `status`，且 `id`/`created_by` 为必填（文档可选/无该字段）。
  - `InputApplyPatchToolCallOutput` 缺 `output`。
- 弱类型字段：工具 `action`/`results` 使用 `interface{}` 承载，强类型约束不足。
- EasyInputMessage：`InputItem` 缺省 `type` 时按 `message` 解析。

---

## 7. 最小示例轮廓（结构级）

- 请求必须包含：`model`、`input`。
- 非流式响应必须包含：`id`、`object`、`created_at`、`model`、`output`。
- 流式响应必须包含：`response.created` → （若干事件） → `response.completed`/`response.failed`/`response.incomplete`，且 `sequence_number` 全局递增。
- 每个 `output_item` 必须至少包含：`response.output_item.added` 与 `response.output_item.done`。
