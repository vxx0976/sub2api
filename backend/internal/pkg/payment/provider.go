package payment

import (
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/google/wire"
)

// ProvideMusePayment creates a MusePayment instance from config
func ProvideMusePayment(cfg *config.Config) *MusePayment {
	return NewMusePayment(cfg.Payment.Muse)
}

// ProviderSet is the Wire provider set for payment package
var ProviderSet = wire.NewSet(
	ProvideMusePayment,
)
