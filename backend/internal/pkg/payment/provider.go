package payment

import (
	"log"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/google/wire"
)

// ProvideAlipayPayment creates an AlipayPayment instance
// config.yaml 作为 fallback，settings 表优先（运行时可改）
func ProvideAlipayPayment(cfg *config.Config, settings SettingGetter) (*AlipayPayment, error) {
	return NewAlipayPayment(cfg.Payment.Alipay, settings)
}

// ProvideAlipayMonitor 创建并启动监控协程
// 注意：监控始终启动，运行时按 settings.alimpay_enabled 控制是否真正工作
func ProvideAlipayMonitor(alipay *AlipayPayment, orderMatcher OrderMatcher) *AlipayMonitor {
	monitor := NewAlipayMonitor(alipay, orderMatcher)
	monitor.Start()
	log.Println("[AlipayMonitor] Started (gated by settings.alimpay_enabled at runtime)")
	return monitor
}

// ProvideUsdtPayment 创建 USDT(TRC20) 收款配置持有者
// config.yaml 作为 fallback，settings 表优先（运行时可改）
func ProvideUsdtPayment(cfg *config.Config, settings SettingStore) (*UsdtPayment, error) {
	return NewUsdtPayment(cfg.Payment.Usdt, settings)
}

// ProvideUsdtMonitor 创建并启动 USDT 链上监控协程
// 始终启动，运行时按 settings.usdt_enabled 控制是否真正工作
func ProvideUsdtMonitor(usdt *UsdtPayment, matcher UsdtOrderMatcher) *UsdtMonitor {
	monitor := NewUsdtMonitor(usdt, matcher)
	monitor.Start()
	return monitor
}

// ProviderSet is the Wire provider set for payment package
var ProviderSet = wire.NewSet(
	ProvideAlipayPayment,
	ProvideAlipayMonitor,
	ProvideUsdtPayment,
	ProvideUsdtMonitor,
)
