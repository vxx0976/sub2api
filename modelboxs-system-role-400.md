# modelboxs：`/v1/responses` 收到 `type:"message"` 的 system 消息时 400

> 发现于 2026-10-19，mayi.one 的 GPT-Pro-mb15（账号 290）已暂停调度，修好后重新打开。

## 现象

modelboxs 的 `/v1/responses` 一碰到 **带 `"type":"message"` 的 system 角色消息**，就返回：

```
HTTP 400
{"error":{"message":"Upstream request failed","type":"upstream_error"}}
```

- 不分模型：gpt-6-sol、gpt-6.1-sol、gpt-6-luna 都会出现。
- 与 JSON 模式（`text.format`）无关，不带 format 也会挂。
- 同一个请求体发给其他 sub2api 中转（ai.zhangrui、pigcode）都能返回 200。
- modelboxs 自己的 `/v1/chat/completions` 收到带 system 的请求是正常的。

## 最小复现

只改一处变量，结果如下：

| `input` 写法 | 结果 |
|---|---|
| `{"type":"message","role":"system","content":"..."}` | **400** |
| `{"type":"message","role":"system","content":[{"type":"input_text","text":"..."}]}` | **400** |
| `{"role":"system","content":"..."}`（不带 type） | 200 |
| `{"type":"message","role":"developer","content":"..."}` | 200 |
| 只有 `{"type":"message","role":"user",...}` | 200 |

```bash
curl -sS https://modelboxs.com/v1/responses -H "Authorization: Bearer $KEY" -H 'Content-Type: application/json' -d '{"model":"gpt-6.1-sol","input":[{"type":"message","role":"system","content":"Be brief"},{"type":"message","role":"user","content":"Say OK"}],"stream":true,"store":false}'
```

预期返回 200 流式输出，实际返回 400 "Upstream request failed"。

下面这个是 mayi.one 把 chat/completions 转成 Responses 后真实发出去的请求体，也会 400：

```json
{"model":"gpt-6-sol","input":[{"type":"message","role":"system","content":"Reply in JSON"},{"type":"message","role":"user","content":"Say OK as {\"r\":...}"}],"stream":true,"include":["reasoning.encrypted_content"],"store":false,"text":{"format":{"type":"json_object"}}}
```

## 影响

- 所有客户端经 `/v1/chat/completions` 发来、带 system 提示词的请求都会挂。这类请求会被转成 Responses，并生成上面那种 system 消息。
- 典型受害者：qmhlan@gmail.com 的生图工具（tiktok-ai-tool-v2）。它生图前先用 chat 写提示词，这一步挂了，用户看到的是"生图失败"。
- Codex 原生走 `/responses`，用的是 developer 角色，不受影响。所以这个号之前看起来一直正常。
- 这个 400 是普通的 400，mayi.one 不会因此换号，请求直接失败。

## 排查方向（还没在 modelboxs 代码上核实）

ChatGPT 内部的 Codex 端点不接受 `role:"system"`，所以 sub2api 需要先改写再发给上游。改写逻辑在：

- `backend/internal/service/openai_codex_transform.go` 的 `extractSystemMessagesFromInput`（约 1446 行）。
- 由 `applyCodexOAuthTransformWithOptions` 调用（约 303 行）。

它会把 system 消息的文本放进 `instructions`，并把 role 改成 `developer`。

不带 `type` 的能过、带 `type:"message"` 的不能过，说明 modelboxs 在发往 ChatGPT 之前，有一条路径把带 type 的 system 消息漏过去了。值得查的地方：

1. modelboxs 部署的版本里，这个函数（或它之前的某个 input 归一化步骤）是不是按 `type=="message"` 走了另一个分支，没改 role。
2. 原生 `/v1/responses` 入口转发给 OAuth 账号的路径，是否和 chat→responses 的桥接路径走了不同的改写逻辑。例如 passthrough 或 Responses Lite 模式下会跳过这个改写。
3. 已排除：modelboxs 后台的「强制 Responses」开关不在实际承接这批请求的账号上，所以不是那条转换路径。先查清 gpt-6-sol / gpt-6.1-sol 在 modelboxs 上落到哪个账号（OAuth 还是 API Key、上游是谁），再看那个账号的转发路径怎么处理 `type:"message"` 的 system。
4. 400 被包成笼统的 "Upstream request failed"，真实错误在上游。打开 modelboxs 的 ops_error_logs 或上游错误详情，看 ChatGPT 返回的原始 message。大概率是 `Unsupported role: system` 一类。

## 修好后的验证

1. 先跑上面的 curl，再把 `role` 换成 `developer` 跑一遍，两次都应返回 200。
2. 回到 mayi.one 恢复 GPT-Pro-mb15 的调度，然后用分组 29 的 key 发这个请求，应返回 200：

```bash
curl -sS https://mayi.one/v1/chat/completions -H "Authorization: Bearer $KEY" -H 'Content-Type: application/json' -d '{"model":"gpt-6-sol","messages":[{"role":"system","content":"Reply in JSON"},{"role":"user","content":"Say OK as {\"r\":...}"}],"response_format":{"type":"json_object"}}'
```

3. 查 usage_logs，确认这条请求落在了 GPT-Pro-mb15 上。

## 别做的事

不要把 mayi.one 上 mb15 的「Responses API 支持」从"强制 Responses"改回自动。

- modelboxs 只支持流式返回，自动探测会把它误判成不支持 Responses。
- 一旦判成不支持，mb15 就会被排除在 Codex 的 `/v1/responses` 流量之外，而这是它绝大部分的请求。
