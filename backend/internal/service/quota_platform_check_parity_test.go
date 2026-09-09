package service

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/migrations"
	"github.com/stretchr/testify/require"
)

// user_platform_quotas.platform 的 CHECK 名单必须与 AllowedQuotaPlatforms 逐字一致。
//
// 为什么需要这条测试：migrations 包里的 TestMigration226FinalPlatformWhitelist 只能断言
// SQL 里出现了某几个平台名——它把同一份名单**硬编码了第二遍**，因此挡不住漂移：
// 谁往 AllowedQuotaPlatforms 加下一个平台，那边照样绿，直到线上「后台校验通过、
// INSERT 撞 DB CHECK」才炸。这类静默故障在 150 / 156 / 157 / 158 已经各修过一次
// （157 是上游误把白名单收回 5 个，158 才恢复）。
//
// 这里从两侧的**真源**取值比对：Go 常量 AllowedQuotaPlatforms vs 嵌入的迁移 SQL。
//
// ⚠️ 不要把迁移文件名写死（历史上写死过 226，minimax 由 237 扩到 9 个平台后这条就红了）：
// 已发布迁移受 checksum 保护不能原地改，加平台**只能**追加一个新的收紧/放宽 CHECK 的迁移。
// 因此这里按迁移号取**最后一个**重建该 CHECK 的迁移作为 DB 侧终态。
// 新增平台时若忘了补迁移，最后一个 CHECK 里就没有新平台，这条照样会红。
func TestQuotaPlatformCheckMatchesAllowedQuotaPlatforms(t *testing.T) {
	entries, err := migrations.FS.ReadDir(".")
	require.NoError(t, err, "读取迁移目录失败")

	numPrefix := regexp.MustCompile(`^(\d+)_`)
	type migrationFile struct {
		seq  int
		name string
	}
	files := make([]migrationFile, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		m := numPrefix.FindStringSubmatch(entry.Name())
		if m == nil {
			continue
		}
		seq, convErr := strconv.Atoi(m[1])
		require.NoError(t, convErr, "迁移文件名前缀不是数字：%s", entry.Name())
		files = append(files, migrationFile{seq: seq, name: entry.Name()})
	}
	require.NotEmpty(t, files, "没有找到任何迁移文件")
	sort.Slice(files, func(i, j int) bool { return files[i].seq < files[j].seq })

	platformLiteral := regexp.MustCompile(`'([a-z0-9_]+)'`)

	var (
		sqlPlatforms []string
		sourceFile   string
	)
	for _, f := range files {
		content, readErr := migrations.FS.ReadFile(f.name)
		require.NoError(t, readErr, "读取迁移 %s 失败", f.name)
		sql := string(content)

		// 定位 ADD CONSTRAINT ... CHECK (platform IN (...)) 里的平台列表；
		// 同一个文件里可能重建多次，取最后一次。
		search := sql
		for {
			idx := strings.Index(search, "ADD CONSTRAINT user_platform_quotas_platform_check")
			if idx == -1 {
				break
			}
			tail := search[idx:]
			open := strings.Index(tail, "CHECK (platform IN (")
			if open == -1 {
				search = tail[len("ADD CONSTRAINT user_platform_quotas_platform_check"):]
				continue
			}
			rest := tail[open+len("CHECK (platform IN ("):]
			end := strings.Index(rest, ")")
			require.NotEqual(t, -1, end, "%s：CHECK 平台列表没有闭合括号", f.name)

			matches := platformLiteral.FindAllStringSubmatch(rest[:end], -1)
			require.NotEmpty(t, matches, "%s：CHECK 里没解析出任何平台", f.name)

			parsed := make([]string, 0, len(matches))
			for _, m := range matches {
				parsed = append(parsed, m[1])
			}
			sqlPlatforms = parsed
			sourceFile = f.name
			search = rest[end:]
		}
	}
	require.NotEmpty(t, sqlPlatforms, "所有迁移里都找不到 user_platform_quotas_platform_check 的 CHECK 名单")

	goPlatforms := append([]string(nil), AllowedQuotaPlatforms...)
	sort.Strings(goPlatforms)
	sort.Strings(sqlPlatforms)

	require.Equal(t, goPlatforms, sqlPlatforms,
		"AllowedQuotaPlatforms 与最后一个重建 CHECK 的迁移 %s 名单不一致。\n"+
			"新增/删除平台时必须同时追加一个重建 CHECK 的新迁移——"+
			"已发布迁移受 checksum 保护，不能原地修改。", sourceFile)
}
