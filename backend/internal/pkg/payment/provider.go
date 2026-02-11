package payment

import (
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/google/wire"
)

// ProvideAlipayPayment creates an AlipayPayment instance from config
func ProvideAlipayPayment(cfg *config.Config) (*AlipayPayment, error) {
	return NewAlipayPayment(cfg.Payment.Alipay)
}

// ProviderSet is the Wire provider set for payment package
var ProviderSet = wire.NewSet(
	ProvideAlipayPayment,
)
