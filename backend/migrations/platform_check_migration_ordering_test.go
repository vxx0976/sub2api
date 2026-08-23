package migrations

import (
	"io/fs"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// 任何重建 user_platform_quotas.platform CHECK 的迁移，都必须排在改名迁移 226 之后。
//
// runner 用 sort.Strings（纯文件名字典序）决定执行顺序。226 是有状态迁移：
// 先 DROP 旧 CHECK → 把 moonshot→kimi / glm→zhipu 改名并删掉 qwen/seedance 行 →
// 最后才 ADD 收敛到 8 平台的新 CHECK。任何排在它**之前**、又无条件
// ADD CONSTRAINT 8 平台白名单的迁移，都会在「尚未跑过 226、库里还有旧平台行」的部署上
// 当场 23514 → 事务型迁移整体回滚 → 容器起不来，且永远轮不到能清洗数据的 226。
//
// 本轮合并上游时就踩过一次：上游的 user_platform_quotas_add_cn_providers 编号是 224，
// 字典序排在 226 之前，已改号为 229。空库和已跑过 226 的库都测不出来
// （前者无存量行、后者数据已干净），只有这条静态断言能提前挡住。
// knownHistoricalPlatformCheckMigrations 是已发布、无法改号的历史例外，见下方引用处说明。
var knownHistoricalPlatformCheckMigrations = map[string]bool{
	"157_user_platform_quotas_add_grok.sql": true,
}

func TestPlatformCheckMigrationsComeAfterRename(t *testing.T) {
	const renameMigration = "226_rename_national_platforms_to_upstream_naming.sql"

	names, err := fs.Glob(FS, "*.sql")
	require.NoError(t, err)
	sort.Strings(names) // 与 runner 的排序方式一致

	renameIdx := -1
	for i, n := range names {
		if n == renameMigration {
			renameIdx = i
		}
	}
	require.NotEqual(t, -1, renameIdx, "找不到改名迁移 %s", renameMigration)

	for i, name := range names {
		if name == renameMigration {
			continue
		}
		content, err := FS.ReadFile(name)
		require.NoError(t, err)
		sql := string(content)
		if !strings.Contains(sql, "ADD CONSTRAINT user_platform_quotas_platform_check") {
			continue
		}
		// 只看「收敛到新 8 平台」的迁移。150/156/157/158 是历史迁移，它们的白名单里
		// 还带着 moonshot/glm/qwen/seedance，排在 226 之前是正确的——226 的职责正是
		// 收敛它们。危险的只有那些白名单已经**不含**旧平台名、却排在 226 之前的迁移：
		// 它们会在数据清洗之前就把约束收紧。
		legacyPlatformNames := false
		for _, legacy := range []string{"'moonshot'", "'glm'", "'qwen'", "'seedance'"} {
			if strings.Contains(sql, legacy) {
				legacyPlatformNames = true
				break
			}
		}
		if legacyPlatformNames {
			continue
		}
		// 已发布的历史例外：157 是上游当年误把白名单收回 5 个平台的那次（紧接着由 158
		// 恢复成 10 个）。它早已在所有库执行完毕、受 checksum 保护无法改号，且后面的 158
		// 立刻放宽回去，不构成「先收紧后清洗」的危险序列。
		if knownHistoricalPlatformCheckMigrations[name] {
			continue
		}
		require.Greater(t, i, renameIdx,
			"%s 重建了 user_platform_quotas 的 platform CHECK，但字典序排在改名迁移 %s 之前。\n"+
				"在尚未跑过改名迁移的库上，它会先于数据清洗执行 ADD CONSTRAINT，"+
				"撞上存量的 moonshot/glm/qwen/seedance 行（SQLSTATE 23514），"+
				"整个迁移事务回滚、应用无法启动且重启无限复现。请把它改号到 %s 之后。",
			name, renameMigration, renameMigration)
	}
}
