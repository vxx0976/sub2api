package payment

import (
	"log"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/google/wire"
)

// ProvideAlipayPayment creates an AlipayPayment instance from config
func ProvideAlipayPayment(cfg *config.Config) (*AlipayPayment, error) {
	return NewAlipayPayment(cfg.Payment.Alipay)
}

// ProvideAlipayMonitor creates and starts the AlipayMonitor
func ProvideAlipayMonitor(alipay *AlipayPayment, orderMatcher OrderMatcher, cfg *config.Config) *AlipayMonitor {
	if !cfg.Payment.Enabled {
		log.Println("[AlipayMonitor] Payment disabled, skipping monitor start")
		return nil
	}
	monitor := NewAlipayMonitor(alipay, orderMatcher)
	monitor.Start()
	return monitor
}

// ProviderSet is the Wire provider set for payment package
var ProviderSet = wire.NewSet(
	ProvideAlipayPayment,
	ProvideAlipayMonitor,
)
