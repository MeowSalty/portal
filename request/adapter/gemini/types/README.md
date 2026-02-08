# Gemini GenerateContent 自述文档

本文档描述 `request/adapter/gemini/types` 目录下对 Gemini GenerateContent API 的**请求**、**非流式响应**与**流式响应**结构的实现支持情况、字段约束与事件语义，并对照官方格式文档进行对齐说明。

适用范围：仅覆盖类型层面的结构、序列化/反序列化约束与事件语义，不包含转换器或业务逻辑。

---

## 1. 范围与对齐基准

- 覆盖类型：`Request`、`Response`、`StreamEvent` 及其关联结构，见 [`request/adapter/gemini/types/request.go`](request/adapter/gemini/types/request.go:7)、[`request/adapter/gemini/types/response.go`](request/adapter/gemini/types/response.go:3)、[`request/adapter/gemini/types/stream.go`](request/adapter/gemini/types/stream.go:3)。
- 对齐基准：`request/adapter/gemini/types/docs/rest.json`（Generative Language API v1beta，revision 20260115），见 [`request/adapter/gemini/types/docs/rest.json`](request/adapter/gemini/types/docs/rest.json:5)。
- 兼容策略：
  - 未知字段透传：通过 `ExtraFields` 字段捕获顶层未知字段，支持 round-trip 序列化。
  - deprecated 字段保留兼容：未主动引入额外 deprecated 字段；按官方字段命名与枚举值保留。
  - 官方未定义但本实现支持的扩展字段需明确标注（见差异与兼容策略）。

---

## 2. 请求（Request）

### 2.1 顶层字段

**必填字段（官方语义）**：

- `contents`：对话内容列表，见 [`Content`](request/adapter/gemini/types/request.go:27)。在类型层面为必填字段（未 `omitempty`）。

**可选字段（按实现结构）**：

- `systemInstruction`：开发者设置的系统指令（内容结构与 `Content` 相同）。
- `generationConfig`：生成配置。
- `safetySettings`：安全设置列表。
- `tools`：工具定义列表。
- `toolConfig`：工具配置。
- `cachedContent`：缓存内容引用。

**路径/URL 传递字段**：

- `model`：官方在 `GenerateContentRequest` 中要求 `model`，但本实现通过 URL 传递，且在 JSON 中被忽略（`json:"-"`），见 [`Request.Model`](request/adapter/gemini/types/request.go:8)。

**未知字段**：

- 通过 `ExtraFields` 字段捕获顶层未知字段，支持 round-trip 序列化，见 [`Request.ExtraFields`](request/adapter/gemini/types/request.go:26)。
- 显式字段优先：如果未知字段名与显式字段冲突，显式字段值优先。
- 未知字段仅作用于顶层，子结构（如 `Content`、`Part`）的未知字段会被忽略。

### 2.2 关键子结构

#### 2.2.1 `Content`

- `role`：仅支持 `user` 或 `model`。
- `parts`：内容片段数组（必填）。

对应官方定义见 [`Content`](request/adapter/gemini/types/docs/rest.json:2354)。

#### 2.2.2 `Part`（内容片段）

`Part` 为 oneof 类型，基础内容只允许设置一种：

- `text`
- `inlineData`
- `functionResponse`
- `functionCall`
- `fileData`
- `executableCode`
- `codeExecutionResult`

同时允许附带的可选字段：`videoMetadata`、`thought`、`thoughtSignature`、`partMetadata`、`mediaResolution`。

序列化由 [`Part.MarshalJSON`](request/adapter/gemini/types/request.go:344) 保证 oneof 约束与字段输出顺序。

#### 2.2.3 工具与工具选择

- `tools`：支持 `functionDeclarations`、`googleSearchRetrieval`、`codeExecution`、`googleSearch`、`computerUse`、`urlContext`、`fileSearch`、`mcpServers`、`googleMaps`，见 [`Tool`](request/adapter/gemini/types/request.go:181)。
- `toolConfig.functionCallingConfig.mode`：类型为 string，可接受官方枚举（`AUTO`/`ANY`/`NONE`/`VALIDATED` 等），但当前实现不做枚举校验。

#### 2.2.4 生成配置（`GenerationConfig`）

- 采样参数：`temperature` / `topP` / `topK` / `seed`。
- 输出控制：`maxOutputTokens` / `stopSequences` / `candidateCount`。
- JSON/Schema 输出：`responseMimeType` / `responseSchema` / `_responseJsonSchema`。
- 思考与多模态：`thinkingConfig` / `responseModalities` / `speechConfig` / `imageConfig` / `mediaResolution`。

与官方字段对齐见 [`GenerationConfig`](request/adapter/gemini/types/docs/rest.json:3225)。

### 2.3 约束与序列化规则

- oneof 约束：`Part` 仅允许设置一种基础内容字段，其他类型会在序列化时被忽略，见 [`Part.MarshalJSON`](request/adapter/gemini/types/request.go:344)。
- `model` 字段：官方在请求体中要求 `model`，本实现通过 URL 传递，不序列化到 JSON。
- `content` 必须包含至少一个 `parts` 元素（业务层需保证）。

---

## 3. 非流式响应（Response）

### 3.1 顶层字段

- `candidates`：候选响应数组（可能为空）。
- `promptFeedback`：提示反馈。
- `usageMetadata`：用量信息。
- `modelVersion` / `responseId` / `modelStatus`：模型状态与响应标识。

对应结构见 [`Response`](request/adapter/gemini/types/response.go:4) 与官方 `GenerateContentResponse` 定义 [`GenerateContentResponse`](request/adapter/gemini/types/docs/rest.json:3474)。

### 3.2 关键子结构

- `Candidate`：响应内容与终止原因、引用与归因信息，见 [`Candidate`](request/adapter/gemini/types/response.go:123)。
- `Content` / `Part`：与请求侧复用相同结构。
- `UsageMetadata`：包含 token 统计与模态细分，见 [`UsageMetadata`](request/adapter/gemini/types/response.go:145)。

### 3.3 约束与序列化规则

- `finishReason` 的枚举值由常量定义，见 [`FinishReason`](request/adapter/gemini/types/response.go:13)。
- `candidates` 为空但 `promptFeedback.blockReason` 设置时表示请求被拦截（与官方行为一致）。

---

## 4. 流式响应（StreamEvent）

### 4.1 顶层字段

- `StreamEvent` 为 `Response` 的类型别名，流式块结构与非流式响应一致，见 [`StreamEvent`](request/adapter/gemini/types/stream.go:3)。

### 4.2 事件类型与字段

- Gemini 的 REST 流式响应每个 chunk 是一个 `GenerateContentResponse` 结构体，字段语义与非流式一致。

### 4.3 事件顺序与完整性要求

- `candidates[].content.parts` 可按 chunk 增量拼接（具体顺序由服务端控制）。
- `finishReason` 在完成时出现，需由调用方确认完整性。

---

## 5. 支持与对齐矩阵（✅/⚠️/❌）

> ✅ 完全对齐，⚠️ 部分对齐（行为差异/强类型不足），❌ 官方有但未实现或官方无但扩展支持。

### 5.1 请求对齐

| 字段/能力                                                   | 官方规范 | 本实现 | 说明                                                    |
| ----------------------------------------------------------- | -------- | ------ | ------------------------------------------------------- |
| `contents`                                                  | ✅       | ✅     | 与官方 `GenerateContentRequest.contents` 对齐。         |
| `model`                                                     | ✅       | ⚠️     | 官方在请求体；本实现通过 URL 传递，不参与 JSON 序列化。 |
| `systemInstruction` / `generationConfig` / `safetySettings` | ✅       | ✅     | 字段与结构对齐。                                        |
| `tools` / `toolConfig`                                      | ✅       | ✅     | 工具种类齐全，枚举不强校验。                            |
| 未知字段透传                                                | ⚠️       | ✅     | 通过 `ExtraFields` 捕获顶层未知字段，支持 round-trip。  |

### 5.2 非流式响应对齐

| 字段/能力                 | 官方规范 | 本实现 | 说明                               |
| ------------------------- | -------- | ------ | ---------------------------------- |
| `GenerateContentResponse` | ✅       | ✅     | 结构与字段对齐。                   |
| `finishReason` 枚举       | ✅       | ✅     | 覆盖官方枚举常量。                 |
| `UsageMetadata` 细分字段  | ✅       | ✅     | 覆盖 prompt/candidate/token 细分。 |

### 5.3 流式响应对齐

| 字段/能力      | 官方规范 | 本实现 | 说明                                   |
| -------------- | -------- | ------ | -------------------------------------- |
| chunk 结构一致 | ✅       | ✅     | `StreamEvent = Response`。             |
| 事件顺序约束   | ⚠️       | ⚠️     | 事件顺序由服务端保证，本实现不强校验。 |

---

## 6. 差异与兼容策略

- `model` 字段位置：官方 `GenerateContentRequest` 中有 `model` 字段，但本实现通过 URL 传递并在 JSON 中忽略，见 [`Request.Model`](request/adapter/gemini/types/request.go:8)。
- 未知字段：通过 `ExtraFields` 字段捕获顶层未知字段，支持 round-trip 序列化，见 [`Request.ExtraFields`](request/adapter/gemini/types/request.go:26)。
- `FunctionDeclaration.description`：官方为必填，本实现为可选字段，见 [`FunctionDeclaration`](request/adapter/gemini/types/request.go:255)。
- `FunctionCallingConfig.mode`：类型为 string，不强制枚举校验，见 [`FunctionCallingConfig`](request/adapter/gemini/types/request.go:284)。

---

## 7. 最小示例轮廓（结构级）

- 请求必须包含：`contents`（且 `content.parts` 至少一个）。
- 非流式响应必须包含：`candidates`（可为空，但若为空应配合 `promptFeedback`）。
- 流式响应必须包含：每个 chunk 为 `GenerateContentResponse` 结构，完成时出现 `finishReason`。
