-- 238: 关掉「开启了模型白名单但一条模型都没选」的历史脏数据。
--
-- 背景：235 把 groups.models_list_config 更名为 model_allowlist，同时把语义从
-- 「只过滤 /v1/models 的展示」升级为「同时做请求准入」（middleware/group_model_allowlist.go
-- 挂在每一条网关链上）。235/236 都只搬数据、不改内容。
--
-- 旧语义下 `{"enabled":true,"models":[]}` 是**彻底的 no-op**——旧的
-- CustomModelsListEnabled() 要求 enabled && len(models) > 0，且过滤函数首行就是
-- 「列表为空则原样返回全部模型」。新语义下同一行数据的含义变成「白名单为空集」：
-- GroupModelAllowlist.Allows() 对任何模型都返回 false，该分组下每一把 Key 的
-- 每一个网关请求（/v1/messages、/v1/chat/completions、Gemini、Codex …）都会被
-- 中间件拦成 404 Model "x" is not available for this group，/v1/models 同时返回空。
--
-- 管理端新的归一化函数已经把这种组合判成 400 配置错误（normalizeGroupModelAllowlist），
-- 也就是说它从此是**存不进去**的状态；库里残留的同款旧行只会造成整组静默停摆。
-- 这里统一翻成 enabled=false，等价于恢复它在旧语义下的实际行为（no-op）。
--
-- 只动 enabled 字段，models 数组原样保留，管理员想启用时重新勾选即可。
UPDATE groups
   SET model_allowlist = jsonb_set(model_allowlist, '{enabled}', 'false'::jsonb)
 WHERE (model_allowlist->>'enabled')::boolean IS TRUE
   AND COALESCE(jsonb_array_length(model_allowlist->'models'), 0) = 0;
