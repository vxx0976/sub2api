package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const nationalPlatformRenameMigration = "226_rename_national_platforms_to_upstream_naming.sql"

// 226 把国产平台标识改成上游命名（moonshot→kimi、glm→zhipu，qwen/seedance 下线）。
// 迁移一旦发布就被 checksum 锁死、无法原地修，因此这里把最危险的几条性质钉在 CI 里。
func TestMigration226OrdersDataConversionBeforeCheckTightening(t *testing.T) {
	content, err := FS.ReadFile(nationalPlatformRenameMigration)
	require.NoError(t, err)
	sql := string(content)

	addConstraint := strings.Index(sql, "ADD CONSTRAINT user_platform_quotas_platform_check")
	require.NotEqual(t, -1, addConstraint, "收紧后的 CHECK 必须存在")

	// ADD CONSTRAINT 会立即校验存量行：任何一条数据转换排在它之后，
	// 存量 moonshot/glm/qwen/seedance 行都会让整个迁移事务回滚 → 应用起不来。
	mustPrecedeAddConstraint := []string{
		"UPDATE accounts SET platform = 'kimi'",
		"UPDATE groups SET platform = 'kimi'",
		"UPDATE user_platform_quotas",
		"DELETE FROM user_platform_quotas WHERE platform IN ('qwen', 'seedance')",
	}
	for _, stmt := range mustPrecedeAddConstraint {
		idx := strings.Index(sql, stmt)
		require.NotEqual(t, -1, idx, "缺少数据转换语句: %s", stmt)
		require.Less(t, idx, addConstraint,
			"数据转换必须排在 ADD CONSTRAINT 之前，否则迁移回滚、启动失败: %s", stmt)
	}

	// 重跑安全：DROP CONSTRAINT IF EXISTS 必须在 ADD 之前。
	dropConstraint := strings.Index(sql, "DROP CONSTRAINT IF EXISTS user_platform_quotas_platform_check")
	require.NotEqual(t, -1, dropConstraint)
	require.Less(t, dropConstraint, addConstraint)

	// 顺序约束的另一半（同样致命，且**空库测不出来**）：
	// 旧 CHECK 在数据转换时仍然在位，而它的白名单里没有 kimi / zhipu
	// （158 的终态是 10 个旧平台名）。若 DROP 排在改名 UPDATE 之后，
	// 第一条 `SET platform = 'kimi'` 就会撞 SQLSTATE 23514，整个事务型迁移回滚，
	// 应用启动失败且重启无限复现。全新库里这些 UPDATE 影响 0 行、一路全绿，
	// 所以只有这条静态断言能挡住它——别指望集成测试。
	firstConversion := addConstraint
	for _, stmt := range mustPrecedeAddConstraint {
		if idx := strings.Index(sql, stmt); idx != -1 && idx < firstConversion {
			firstConversion = idx
		}
	}
	require.Less(t, dropConstraint, firstConversion,
		"旧 CHECK 必须在任何数据转换之前卸掉：它的白名单里没有 kimi/zhipu，"+
			"改名 UPDATE 会当场违约(23514)、整个迁移回滚、应用起不来")

	// 事务型迁移：不得含 CONCURRENTLY（那需要 _notx 后缀）。
	require.NotContains(t, strings.ToUpper(sql), "CONCURRENTLY")
	require.False(t, strings.HasSuffix(nationalPlatformRenameMigration, "_notx.sql"))
}

// 终态白名单必须与 service.AllowedQuotaPlatforms / ent schema 逐字一致（8 个平台）。
// 三处任意一处漂移都会让「校验通过但 INSERT 撞 DB CHECK」的静默故障重演
// （150/156/158 已经为同一类问题各修过一次）。
func TestMigration226FinalPlatformWhitelist(t *testing.T) {
	content, err := FS.ReadFile(nationalPlatformRenameMigration)
	require.NoError(t, err)
	sql := string(content)

	for _, p := range []string{"anthropic", "openai", "gemini", "antigravity", "grok", "kimi", "zhipu", "deepseek"} {
		require.Contains(t, sql, "'"+p+"'", "终态白名单缺少平台 %s", p)
	}

	// 收紧后的 CHECK 里绝不能再出现下线/改名前的标识。
	addConstraint := strings.Index(sql, "ADD CONSTRAINT user_platform_quotas_platform_check")
	require.NotEqual(t, -1, addConstraint)
	tail := sql[addConstraint:]
	for _, gone := range []string{"'moonshot'", "'glm'", "'qwen'", "'seedance'"} {
		require.NotContains(t, tail, gone, "收紧后的 CHECK 仍含旧平台标识 %s", gone)
	}
}

// 边界：只改平台标识，绝不碰模型名。channels.model_mapping 是
// {平台: {源模型: 目标模型}} 的嵌套结构，只允许重命名外层平台键。
func TestMigration226DoesNotRewriteModelNames(t *testing.T) {
	content, err := FS.ReadFile(nationalPlatformRenameMigration)
	require.NoError(t, err)
	sql := string(content)

	// 只看可执行语句：注释里为了说明「平台标识 vs 模型名」的区别会举模型名的例子。
	executable := stripSQLLineComments(sql)

	// 模型名字面量（moonshot-v1-*、glm-*、kimi-k*）不得出现在任何转换语句里。
	for _, modelish := range []string{"moonshot-v1", "glm-5", "glm-4", "kimi-k2", "kimi-k3", "qwen-max", "qwen-plus"} {
		require.NotContains(t, executable, modelish,
			"迁移不得引用模型名 %s——本迁移只改平台标识", modelish)
	}

	// model_mapping 只能整体搬迁外层键：内层 map 必须原样取值。
	require.Contains(t, sql, "jsonb_build_object('kimi', model_mapping -> 'moonshot')")
	require.Contains(t, sql, "jsonb_build_object('zhipu', model_mapping -> 'glm')")
}

// stripSQLLineComments 去掉整行 `--` 注释，只留可执行 SQL。
// 本迁移不含行内 `--`（字符串字面量里也没有），因此逐行前缀判断即可。
func stripSQLLineComments(sql string) string {
	var b strings.Builder
	for _, line := range strings.Split(sql, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "--") {
			continue
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	return b.String()
}
