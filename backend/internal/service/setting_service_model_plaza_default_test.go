//go:build unit

package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// 模型广场在本 fork 取代了 /models，是唯一的公开定价页，所以「没配过」必须等于开启。
// 上游默认关，而 InitializeDefaultSettings 在 registration_enabled 已存在时整体早退
// —— 存量部署永远不会被补写这条记录，键会一直缺失。若这里退化成
// `settings[key] == "true"`，所有存量部署都将没有公开定价页（后端 404 + 前端跳 /home）。
func TestModelPlazaEnabledFrom_AbsentKeyMeansEnabled(t *testing.T) {
	t.Run("键缺失视为开启", func(t *testing.T) {
		require.True(t, modelPlazaEnabledFrom(map[string]string{}))
		require.True(t, modelPlazaEnabledFrom(map[string]string{
			SettingKeyModelPlazaRequireAuth: "false",
		}), "只有别的 plaza 键存在时也算没配过")
	})

	t.Run("显式关闭仍然生效", func(t *testing.T) {
		require.False(t, modelPlazaEnabledFrom(map[string]string{
			SettingKeyModelPlazaEnabled: "false",
		}), "管理员在后台关掉的意图不能被默认值覆盖")
	})

	t.Run("显式开启", func(t *testing.T) {
		require.True(t, modelPlazaEnabledFrom(map[string]string{
			SettingKeyModelPlazaEnabled: "true",
		}))
	})

	t.Run("空串与非法值按关闭处理", func(t *testing.T) {
		// 空串是「写过但清空」，与键缺失语义不同，按字面判定 → 关闭。
		require.False(t, modelPlazaEnabledFrom(map[string]string{
			SettingKeyModelPlazaEnabled: "",
		}))
		require.False(t, modelPlazaEnabledFrom(map[string]string{
			SettingKeyModelPlazaEnabled: "yes",
		}))
	})
}
