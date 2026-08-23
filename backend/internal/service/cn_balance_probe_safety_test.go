package service

import "testing"

// 余额探测的失败必须表现为「不确定」而不是「余额 0」。
//
// 下游 checkOne 只看 Success：Success=true 且余额低于阈值(默认 0.5) 就会
// SetTempUnschedulable 20 分钟，10 分钟后再探又是 0 → 无限循环停调。
// 线上只有一个 Kimi 账号，踩中即整组 no available account。
func TestKimiBalanceProbeDoesNotTreatUnparsableAsZero(t *testing.T) {
	// cnParseF64 对缺失/非数值返回 (0,false)——旧代码丢掉了这个 ok。
	for _, v := range []any{nil, "", "abc", map[string]any{}} {
		if _, ok := cnParseF64(v); ok {
			t.Errorf("cnParseF64(%v) 不应报告解析成功", v)
		}
	}
	if got, ok := cnParseF64(12.5); !ok || got != 12.5 {
		t.Errorf("cnParseF64(12.5) = (%v,%v)，期望 (12.5,true)", got, ok)
	}
}

// Kimi 的余额端点只有 Moonshot 开放平台有。账号 base_url 指向别处时必须不探测，
// 否则会把 api_key 发给一台运维从未为此配置过的主机，而且拿不到余额字段又会
// 踩上面那条「解析失败当 0」的坑。生产的 Kimi 账号 base_url 正是 api.kimi.com。
func TestKimiBalanceURLSkipsNonMoonshotHosts(t *testing.T) {
	mk := func(platform, baseURL string) *Account {
		a := &Account{Platform: platform, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "k"}}
		if baseURL != "" {
			a.Credentials["base_url"] = baseURL
		}
		return a
	}
	cases := []struct {
		name string
		acc  *Account
		want string
	}{
		{"未配 base_url → 用官方端点", mk(PlatformKimi, ""), "https://api.moonshot.cn/v1/users/me/balance"},
		{"官方 base_url → 用官方端点", mk(PlatformKimi, "https://api.moonshot.cn/v1"), "https://api.moonshot.cn/v1/users/me/balance"},
		{"生产的 Kimi For Coding → 不探测", mk(PlatformKimi, "https://api.kimi.com/coding/v1"), ""},
		{"中转 → 不探测", mk(PlatformKimi, "https://relay.example.com/v1"), ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := cnBalanceURL(tc.acc); got != tc.want {
				t.Errorf("cnBalanceURL = %q，期望 %q", got, tc.want)
			}
		})
	}
}
