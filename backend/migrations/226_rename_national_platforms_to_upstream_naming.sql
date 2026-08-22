-- 国产平台标识对齐上游命名：moonshot → kimi、glm → zhipu；qwen / seedance 整体下线。
--
-- 背景
--   本 fork 先于上游实现了国产 OpenAI 兼容供应商，平台标识取名 moonshot/glm/qwen/seedance；
--   上游后来自行实现了同类平台，标识为 kimi/zhipu/deepseek。两套命名并存导致每轮
--   main → dev 合并都要在平台枚举、调度、配额、前端徽章等处反复解冲突。站长决定改用上游命名：
--     moonshot → kimi      （Kimi 月之暗面 / Moonshot）
--     glm      → zhipu     （智谱 GLM / bigmodel）
--     deepseek → deepseek  （值不变，仅 Go 标识符大小写改为 PlatformDeepseek，DB 无需动）
--     qwen     → 删除      （上游无此平台）
--     seedance → 删除      （上游无此平台）
--
-- 为什么可以直接删 qwen / seedance
--   生产实况：accounts / groups(active) 均 0 行、composite_model_routes 0 行、历史零调用；
--   usage_logs 没有 platform 列（按 account_id 关联），历史账单与用量统计不受影响。
--   唯一有存量的是 user_platform_quotas 里注册时自动预填的默认配额行（qwen 37、seedance 37），
--   这些行从未生效过（没有账号就没有调用），且第 3 步收紧后的 CHECK 不再允许这两个值，
--   必须在收紧前删掉。
--
-- 为什么顺序不能颠倒（本文件最关键的一条）
--   ADD CONSTRAINT ... CHECK 会立即校验存量行。若先收紧 CHECK 再改数据，存量
--   moonshot/glm/qwen/seedance 行会让 ADD CONSTRAINT 当场失败 → 整个迁移事务回滚 →
--   启动时迁移报错、应用起不来。所以严格按「先转换数据（1、2 步），最后重建 CHECK（3 步）」编排。
--
-- 边界：只改「平台标识」，绝不碰模型名
--   moonshot/glm/kimi/qwen 这些字符串在本项目里有两种语境：
--     (a) 平台标识：accounts.platform / groups.platform / 平台枚举 / 按平台维度的配置与指标；
--     (b) 模型名：moonshot-v1-8k、kimi-k3、glm-5.1、channels.model_mapping 内层的
--         src→dst 模型映射、channel_model_pricing.models 数组、¥ 计价表匹配用的模型名……
--   本迁移只动 (a)；所有 (b) 一律原样保留。特别地：
--     - channels.model_mapping 是 {平台: {源模型: 目标模型}} 的嵌套结构，本迁移只改**外层平台键**，
--       内层模型映射整体搬迁不做任何改写；
--     - channel_model_pricing / channel_account_stats_model_pricing 只改 platform 列，
--       models 数组（模型名）不动；
--     - channel_monitor_v2_config.platforms 里每项只改 "platform" 字段，"models" 数组不动。
--
-- 不动的东西
--   - composite_model_routes.target_platform：复合分组只支持
--     anthropic/openai/gemini/antigravity/grok（既有 CHECK 已经把国产平台挡在外面，
--     生产 0 行），无需改名，也不放宽 CHECK。
--   - accounts / groups 里万一存在的 qwen/seedance 行：不删，但**停用**（见第 1 步末尾）。
--     账号/分组挂着用量、订单、外键，删除是不可逆的破坏性操作，不适合在迁移里替站长做决定
--     （生产实测为 0 行，这两条只是防其它部署）。但也不能原样留成 active——理由见下。
--   - ops_alert_rules / ops_alert_silences 里 qwen/seedance 的告警作用域：同理只改名不删，
--     指向已下线平台的规则本身就不会触发。
--
-- 幂等：所有语句都带「WHERE 旧值」或 IF EXISTS 守卫，可重复执行且第二次为空操作。


-- ---------------------------------------------------------------------------
-- 0. 先卸掉旧 CHECK —— 必须排在任何数据转换之前
--
--     这是「顺序」问题的另一半，与第 4 步的注意事项互为镜像，两边都会致命：
--       - 第 4 步方向：ADD CONSTRAINT 立即校验存量行 → 数据必须**先**转换完；
--       - 本步方向：旧 CHECK 在转换时**仍然在位**，而它的白名单里没有 kimi / zhipu
--         （158 留下的终态是 anthropic/openai/gemini/antigravity/deepseek/
--          moonshot/glm/qwen/seedance/grok 这 10 个旧名）→ 第 1 步那条把 moonshot
--         改成 kimi 的 UPDATE 会当场撞 SQLSTATE 23514。
--     226 是事务型迁移（整文件一个事务，见 migrations_runner.go），任一条失败即整体回滚，
--     ApplyMigrations 报错 → 容器起不来，且重启无限复现同一错误。
--
--     ⚠️ 这个坑在空库上测不出来：全新库里 user_platform_quotas 没有 moonshot 行，
--     UPDATE 影响 0 行，001→226 一路全绿。只有带存量数据的真实库才炸，
--     所以别拿「集成测试通过」当它安全的证据。
--
--     代价：ACCESS EXCLUSIVE 锁从事务开头持到提交（而非只在末尾）。这几张表都很小，
--     且 runner 的 SET LOCAL lock_timeout 仍然生效，可接受。
-- ---------------------------------------------------------------------------
ALTER TABLE user_platform_quotas
    DROP CONSTRAINT IF EXISTS user_platform_quotas_platform_check;


-- ---------------------------------------------------------------------------
-- 1. 配置 / 业务表：platform 标识改名
-- ---------------------------------------------------------------------------

-- 账号与分组（生产：moonshot 各 1 行，glm 0 行）。
-- groups 上的 enqueue_group_auth_cache_invalidation 触发器会检测到 platform 变化，
-- 自动把受影响 api_key 的缓存快照写进 auth_cache_invalidation_outbox，无需手工失效。
UPDATE accounts SET platform = 'kimi'  WHERE platform = 'moonshot';
UPDATE accounts SET platform = 'zhipu' WHERE platform = 'glm';

UPDATE groups SET platform = 'kimi'  WHERE platform = 'moonshot';
UPDATE groups SET platform = 'zhipu' WHERE platform = 'glm';

-- 渠道（余额/计价管理）及其按平台维度的计价规则。
UPDATE channels SET platform = 'kimi'  WHERE platform = 'moonshot';
UPDATE channels SET platform = 'zhipu' WHERE platform = 'glm';

UPDATE channel_model_pricing SET platform = 'kimi'  WHERE platform = 'moonshot';
UPDATE channel_model_pricing SET platform = 'zhipu' WHERE platform = 'glm';

UPDATE channel_account_stats_model_pricing SET platform = 'kimi'  WHERE platform = 'moonshot';
UPDATE channel_account_stats_model_pricing SET platform = 'zhipu' WHERE platform = 'glm';

-- channels.model_mapping：{平台: {源模型: 目标模型}}，只重命名**外层平台键**。
-- 若同时存在旧键和新键（理论上不会发生），保留已存在的新键、丢弃旧键，避免键冲突。
UPDATE channels
SET model_mapping = (
        (model_mapping - 'moonshot' - 'glm')
        || CASE WHEN model_mapping ? 'moonshot' AND NOT (model_mapping ? 'kimi')
                THEN jsonb_build_object('kimi', model_mapping -> 'moonshot')
                ELSE '{}'::jsonb END
        || CASE WHEN model_mapping ? 'glm' AND NOT (model_mapping ? 'zhipu')
                THEN jsonb_build_object('zhipu', model_mapping -> 'glm')
                ELSE '{}'::jsonb END
    )
WHERE model_mapping IS NOT NULL
  AND jsonb_typeof(model_mapping) = 'object'
  AND (model_mapping ?| ARRAY['moonshot', 'glm']);

-- user_platform_quotas：先删掉「改名后会和已有活跃行撞唯一索引」的旧行
-- （userplatformquota_user_id_platform_uq 是 deleted_at IS NULL 的部分唯一索引；
--  正常情况下不会有 kimi/zhipu 行，这段只是保证重跑/异常态下不会因唯一冲突中止迁移），
-- 再统一改名，最后删除 qwen / seedance 的预填行。
DELETE FROM user_platform_quotas q
WHERE q.platform IN ('moonshot', 'glm')
  AND q.deleted_at IS NULL
  AND EXISTS (
        SELECT 1
        FROM user_platform_quotas k
        WHERE k.user_id = q.user_id
          AND k.deleted_at IS NULL
          AND k.platform = CASE q.platform WHEN 'moonshot' THEN 'kimi' ELSE 'zhipu' END
  );

UPDATE user_platform_quotas
SET platform = CASE platform WHEN 'moonshot' THEN 'kimi' ELSE 'zhipu' END,
    updated_at = NOW()
WHERE platform IN ('moonshot', 'glm');

DELETE FROM user_platform_quotas WHERE platform IN ('qwen', 'seedance');

-- 残留的 qwen / seedance 账号与分组必须停用，否则升级后会把 API Key 发给错误的上游。
--
--   本次改动从 isOpenAICompatPlatform 里摘掉了 qwen/seedance。此后 platform='qwen' 的分组
--   在 /v1/messages、/chat/completions 上都不再命中 OpenAI 兼容分支，会 fall through 到
--   **Anthropic 网关**；而 Anthropic 这条路径上没有平台白名单兜底：
--     matchingPlatforms('qwen') 原样返回 ['qwen'] → 按 platform 直接查出该账号；
--     账号没配 model_mapping 就视为「支持所有模型」→ 被选中；
--     buildUpstreamRequest 取 account.GetBaseURL()，credentials.base_url 为空时
--     默认回落 https://api.anthropic.com（旧代码是靠已被删除的 GetQwenBaseURL 回落到
--     DashScope 的，那个默认值随平台一起没了）。
--   净效果：阿里云百炼 / 火山方舟的 API Key 会被当作 x-api-key 发给 Anthropic —— 凭证外泄。
--
--   停用而非删除：status='disabled' 已足以阻断（可调度账号查询硬性要求 status='active'），
--   且完全可逆——站长把平台迁走后自行清理或改配到别的平台即可。
--   只动 active 行，重跑为空操作。
UPDATE accounts
SET status = 'disabled', updated_at = NOW()
WHERE platform IN ('qwen', 'seedance')
  AND status = 'active';

UPDATE groups
SET status = 'disabled', updated_at = NOW()
WHERE platform IN ('qwen', 'seedance')
  AND status = 'active';


-- ---------------------------------------------------------------------------
-- 2. JSON 配置里的平台标识
-- ---------------------------------------------------------------------------

-- 2.1 settings 表里的默认平台配额：
--     key = 'default_platform_quotas'（系统层）
--     key LIKE 'auth_source_default_%_platform_quotas'（各登录来源层）
--     值是 {平台: {daily/weekly/monthly}} 的 JSON 对象。
--     必须一起改：这些 map 是注册时预填 user_platform_quotas 的数据源，
--     留着 moonshot/qwen 键会（a）让新注册用户重新写入被 CHECK 拒绝的平台行，
--     （b）让后台保存设置时撞 validateDefaultPlatformQuotaMap 的白名单校验直接报错。
--     逐行处理并对非法 JSON 容错跳过，避免个别脏值把整个迁移（和启动）拖挂。
DO $$
DECLARE
    r        RECORD;
    parsed   JSONB;
    rewritten JSONB;
BEGIN
    FOR r IN
        SELECT id, key, value
        FROM settings
        WHERE key = 'default_platform_quotas'
           OR key LIKE 'auth\_source\_default\_%\_platform\_quotas'
    LOOP
        IF r.value IS NULL OR btrim(r.value) = '' THEN
            CONTINUE;
        END IF;

        BEGIN
            parsed := r.value::jsonb;
        EXCEPTION WHEN others THEN
            RAISE NOTICE '226: skip settings key % (value is not valid JSON)', r.key;
            CONTINUE;
        END;

        IF jsonb_typeof(parsed) <> 'object'
           OR NOT (parsed ?| ARRAY['moonshot', 'glm', 'qwen', 'seedance']) THEN
            CONTINUE;
        END IF;

        SELECT COALESCE(
                   jsonb_object_agg(
                       CASE e.key
                           WHEN 'moonshot' THEN 'kimi'
                           WHEN 'glm'      THEN 'zhipu'
                           ELSE e.key
                       END,
                       e.value
                   ),
                   '{}'::jsonb
               )
          INTO rewritten
          FROM jsonb_each(parsed) AS e
         WHERE e.key NOT IN ('qwen', 'seedance');

        UPDATE settings
        SET value = rewritten::text,
            updated_at = NOW()
        WHERE id = r.id;
    END LOOP;
END $$;

-- 2.2 error_passthrough_rules.platforms：平台标识字符串数组（服务端按小写比较）。
--     改名 moonshot/glm，并剔除 qwen/seedance；保持原有顺序，模型无关。
--
--     ⚠️ COALESCE 的兜底是 platforms（原值）而**不是** '[]'：空数组在这张表里不是
--     「不匹配任何平台」，而是 platformMatchesCached() 的
--         if len(rule.lowerPlatforms) == 0 { return true }
--     ——**匹配所有平台**。若某条规则的 platforms 恰好只有 qwen/seedance，剔除后留下空数组，
--     这条原本只作用于一个已下线平台的规则会静默升级成全平台生效，改变 Anthropic/OpenAI
--     等主力平台的错误透传行为（该不该把上游错误体透给客户端、该不该触发故障转移）。
--     jsonb_agg 在空集上返回 NULL，因此这里让 COALESCE 回落到原值：整条规则原样保留。
--     保留是安全的——不会再有任何请求带 platform='qwen'，规则在运行时恒不命中，
--     管理员也能在后台看到它引用了已下线平台并自行清理。
--     （只含 moonshot/glm 的规则不受影响：那是改名不是剔除，聚合结果非空。）
UPDATE error_passthrough_rules
SET platforms = COALESCE(
        (
            SELECT jsonb_agg(
                       CASE lower(e.value)
                           WHEN 'moonshot' THEN 'kimi'
                           WHEN 'glm'      THEN 'zhipu'
                           ELSE e.value
                       END
                       ORDER BY e.ord
                   )
            FROM jsonb_array_elements_text(platforms) WITH ORDINALITY AS e(value, ord)
            WHERE lower(e.value) NOT IN ('qwen', 'seedance')
        ),
        platforms
    ),
    updated_at = NOW()
WHERE jsonb_typeof(platforms) = 'array'
  AND EXISTS (
        SELECT 1
        FROM jsonb_array_elements_text(platforms) AS t(value)
        WHERE lower(t.value) IN ('moonshot', 'glm', 'qwen', 'seedance')
  );

-- 2.3 channel_monitor_v2_config.platforms：[{"platform":..,"enabled":..,"models":[..]}]。
--     只改每项的 "platform" 字段并剔除 qwen/seedance 项；"models" 是模型名，原样保留。
--     同时 version + 1：该列是后台保存的乐观锁版本号，带外改动后让仍停留在旧版本的
--     管理页保存时冲突失败、强制刷新，避免把旧平台列表覆盖回来。
UPDATE channel_monitor_v2_config c
SET platforms = COALESCE(
        (
            SELECT jsonb_agg(
                       CASE e.value ->> 'platform'
                           WHEN 'moonshot' THEN jsonb_set(e.value, '{platform}', '"kimi"'::jsonb)
                           WHEN 'glm'      THEN jsonb_set(e.value, '{platform}', '"zhipu"'::jsonb)
                           ELSE e.value
                       END
                       ORDER BY e.ord
                   )
            FROM jsonb_array_elements(c.platforms) WITH ORDINALITY AS e(value, ord)
            WHERE COALESCE(e.value ->> 'platform', '') NOT IN ('qwen', 'seedance')
        ),
        '[]'::jsonb
    ),
    version = c.version + 1,
    updated_at = NOW()
WHERE jsonb_typeof(c.platforms) = 'array'
  AND EXISTS (
        SELECT 1
        FROM jsonb_array_elements(c.platforms) AS t(value)
        WHERE t.value ->> 'platform' IN ('moonshot', 'glm', 'qwen', 'seedance')
  );

-- 2.4 ops_alert_rules.filters->>'platform'：告警规则的平台作用域。只改名。
UPDATE ops_alert_rules
SET filters = jsonb_set(
        filters,
        '{platform}',
        to_jsonb(CASE filters ->> 'platform' WHEN 'moonshot' THEN 'kimi' ELSE 'zhipu' END)
    ),
    updated_at = NOW()
WHERE filters IS NOT NULL
  AND jsonb_typeof(filters) = 'object'
  AND filters ->> 'platform' IN ('moonshot', 'glm');


-- ---------------------------------------------------------------------------
-- 3. 观测 / 历史数据表：只改名，不删行
--     ops_* 与 channel_monitor_v2_* 存的是历史观测数据。改名是为了让旧数据在新命名下
--     仍能按平台聚合出来（否则 moonshot 的历史曲线会凭空断掉）；qwen/seedance 的历史点
--     即便存在也保留原值，删掉只会让历史不可读。
--     这些表大多带含 platform 的唯一键/主键，理论上不会与新命名撞键（新命名的行只可能在
--     本迁移之后产生）；万一撞上（例如 schema_migrations 被清空后在混合数据上重跑），
--     捕获 unique_violation 跳过该表而不是让启动失败。
-- ---------------------------------------------------------------------------
DO $$
DECLARE
    t     TEXT;
    tbls  TEXT[] := ARRAY[
        'ops_metrics_hourly',
        'ops_metrics_daily',
        'ops_error_logs',
        'ops_system_metrics',
        'ops_system_logs',
        'ops_alert_silences',
        'channel_monitor_v2_metrics_1m',
        'channel_monitor_v2_user_metrics_1m',
        'channel_monitor_v2_error_metrics_1m',
        'channel_monitor_v2_latency_histograms_1m',
        'channel_monitor_v2_metrics_rollup',
        'channel_monitor_v2_user_metrics_rollup',
        'channel_monitor_v2_error_metrics_rollup',
        'channel_monitor_v2_latency_histograms_rollup'
    ];
BEGIN
    FOREACH t IN ARRAY tbls LOOP
        -- 表或 platform 列不存在（历史部署裁剪过监控模块）时跳过，不让迁移硬失败。
        IF to_regclass(t) IS NULL THEN
            CONTINUE;
        END IF;
        IF NOT EXISTS (
            SELECT 1
            FROM information_schema.columns
            WHERE table_schema = current_schema()
              AND table_name = t
              AND column_name = 'platform'
        ) THEN
            CONTINUE;
        END IF;

        BEGIN
            EXECUTE format(
                'UPDATE %I SET platform = CASE platform WHEN %L THEN %L ELSE %L END '
                || 'WHERE platform IN (%L, %L)',
                t, 'moonshot', 'kimi', 'zhipu', 'moonshot', 'glm'
            );
        EXCEPTION WHEN unique_violation THEN
            RAISE NOTICE '226: skip platform rename on % (unique_violation: new-name rows already exist)', t;
        END;
    END LOOP;
END $$;


-- ---------------------------------------------------------------------------
-- 4. 收紧 user_platform_quotas.platform 的 CHECK（必须在数据转换之后）
--     终态 = 上游 5 平台 + 保留的 3 个国产平台 = 8 个：
--     anthropic / openai / gemini / antigravity / grok / kimi / zhipu / deepseek。
--     历史沿革：142 建表 4 平台 → 150 放宽到 8（含 moonshot/glm/seedance）→ 156 加 qwen（9）
--     → 157（上游）误收回 5 → 158 恢复 10。本迁移是这条约束的新终态，
--     150/156/158 里写着旧平台名的 CHECK 到此全部被覆盖（已发布迁移受 checksum 保护，
--     不能原地改，只能像这样用新迁移收敛）。
--     平台列表须与 backend/internal/service/domain_constants.go 的 AllowedQuotaPlatforms 一致。
-- ---------------------------------------------------------------------------
-- 预检：ADD CONSTRAINT 会立即校验存量行，失败会让整个迁移回滚、应用起不来。
-- 这里先把「不在新白名单内的残留值」变成一条能直接看懂的报错，而不是裸的约束冲突。
-- 正常路径下第 1 步已经把 moonshot/glm 改名、qwen/seedance 删掉，这里应当为空。
DO $$
DECLARE
    leftover TEXT;
BEGIN
    SELECT string_agg(DISTINCT platform, ', ')
      INTO leftover
      FROM user_platform_quotas
     WHERE platform NOT IN (
               'anthropic', 'openai', 'gemini', 'antigravity', 'grok',
               'kimi', 'zhipu', 'deepseek'
           );
    IF leftover IS NOT NULL THEN
        RAISE EXCEPTION
            '226: user_platform_quotas 仍有不在新白名单内的 platform 值 (%)，请人工处理后再重跑迁移',
            leftover;
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'user_platform_quotas_platform_check'
          AND conrelid = 'user_platform_quotas'::regclass
    ) THEN
        ALTER TABLE user_platform_quotas
            ADD CONSTRAINT user_platform_quotas_platform_check
            CHECK (platform IN (
                'anthropic', 'openai', 'gemini', 'antigravity', 'grok',
                'kimi', 'zhipu', 'deepseek'
            ));
    END IF;
END $$;
