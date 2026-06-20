package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/payment"
)

// AdminUsdtChainConfig 单条链的后台配置（GET，敏感字段脱敏）。
type AdminUsdtChainConfig struct {
	Enabled    bool   `json:"enabled"`
	Address    string `json:"address"`
	HasAPIKey  bool   `json:"has_api_key"`
	APIBaseURL string `json:"api_base_url"`
}

// AdminUsdtConfig USDT 多链后台配置 DTO。
// 敏感字段 per-chain api_key 在 GET 时只返回 has_api_key。
type AdminUsdtConfig struct {
	Enabled                bool                            `json:"enabled"` // 主开关
	ManualRate             float64                         `json:"manual_rate"`
	RateAutoFetch          bool                            `json:"rate_auto_fetch"`
	RateMarkup             float64                         `json:"rate_markup"`
	AmountOffset           float64                         `json:"amount_offset"`
	AmountTolerance        float64                         `json:"amount_tolerance"`
	ConfirmSeconds         int                             `json:"confirm_seconds"`
	MonitorIntervalSeconds int                             `json:"monitor_interval_seconds"`
	QueryMinutesBack       int                             `json:"query_minutes_back"`
	OrderTimeoutSeconds    int                             `json:"order_timeout_seconds"`
	Chains                 map[string]AdminUsdtChainConfig `json:"chains"`
}

// AdminUsdtChainConfigUpdate 单条链的更新（nil=不变；api_key 空字符串=保留原值）。
type AdminUsdtChainConfigUpdate struct {
	Enabled    *bool   `json:"enabled"`
	Address    *string `json:"address"`
	APIKey     *string `json:"api_key"`
	APIBaseURL *string `json:"api_base_url"`
}

// AdminUsdtConfigUpdate 更新请求（nil 指针=不变）。
type AdminUsdtConfigUpdate struct {
	Enabled                *bool                                 `json:"enabled"`
	ManualRate             *float64                              `json:"manual_rate"`
	RateAutoFetch          *bool                                 `json:"rate_auto_fetch"`
	RateMarkup             *float64                              `json:"rate_markup"`
	AmountOffset           *float64                              `json:"amount_offset"`
	AmountTolerance        *float64                              `json:"amount_tolerance"`
	ConfirmSeconds         *int                                  `json:"confirm_seconds"`
	MonitorIntervalSeconds *int                                  `json:"monitor_interval_seconds"`
	QueryMinutesBack       *int                                  `json:"query_minutes_back"`
	OrderTimeoutSeconds    *int                                  `json:"order_timeout_seconds"`
	Chains                 map[string]AdminUsdtChainConfigUpdate `json:"chains"`
}

// GetAdminUsdtConfig 读取当前配置（敏感字段脱敏）。
func (s *UsdtOrderService) GetAdminUsdtConfig(ctx context.Context) (*AdminUsdtConfig, error) {
	get := func(k string) string {
		v, _ := s.settingRepo.GetValue(ctx, k)
		return v
	}
	parseFloat := func(v string) float64 {
		f, _ := strconv.ParseFloat(v, 64)
		return f
	}
	parseInt := func(v string) int {
		i, _ := strconv.Atoi(v)
		return i
	}

	cfg := &AdminUsdtConfig{
		Enabled:                get(payment.SettingKeyUsdtEnabled) == "true",
		ManualRate:             parseFloat(get(payment.SettingKeyUsdtManualRate)),
		RateAutoFetch:          get(payment.SettingKeyUsdtRateAutoFetch) == "true",
		RateMarkup:             parseFloat(get(payment.SettingKeyUsdtRateMarkup)),
		AmountOffset:           parseFloat(get(payment.SettingKeyUsdtAmountOffset)),
		AmountTolerance:        parseFloat(get(payment.SettingKeyUsdtAmountTolerance)),
		ConfirmSeconds:         parseInt(get(payment.SettingKeyUsdtConfirmSeconds)),
		MonitorIntervalSeconds: parseInt(get(payment.SettingKeyUsdtMonitorIntervalSeconds)),
		QueryMinutesBack:       parseInt(get(payment.SettingKeyUsdtQueryMinutesBack)),
		OrderTimeoutSeconds:    parseInt(get(payment.SettingKeyUsdtOrderTimeoutSeconds)),
		Chains:                 make(map[string]AdminUsdtChainConfig, len(payment.SupportedChains)),
	}
	for _, chain := range payment.SupportedChains {
		cfg.Chains[chain] = AdminUsdtChainConfig{
			Enabled:    get(payment.UsdtChainSettingKey(chain, "enabled")) == "true",
			Address:    get(payment.UsdtChainSettingKey(chain, "address")),
			HasAPIKey:  get(payment.UsdtChainSettingKey(chain, "api_key")) != "",
			APIBaseURL: get(payment.UsdtChainSettingKey(chain, "api_base_url")),
		}
	}
	return cfg, nil
}

// UpdateAdminUsdtConfig 更新配置（nil 不变；per-chain api_key 空字符串保留原值）。保存后 Reload。
func (s *UsdtOrderService) UpdateAdminUsdtConfig(ctx context.Context, req *AdminUsdtConfigUpdate) error {
	if req == nil {
		return nil
	}
	set := func(key, value string) error {
		return s.settingRepo.Set(ctx, key, value)
	}

	if req.Enabled != nil {
		if err := set(payment.SettingKeyUsdtEnabled, strconv.FormatBool(*req.Enabled)); err != nil {
			return err
		}
	}
	if req.ManualRate != nil {
		if err := set(payment.SettingKeyUsdtManualRate, strconv.FormatFloat(*req.ManualRate, 'f', -1, 64)); err != nil {
			return err
		}
	}
	if req.RateAutoFetch != nil {
		if err := set(payment.SettingKeyUsdtRateAutoFetch, strconv.FormatBool(*req.RateAutoFetch)); err != nil {
			return err
		}
	}
	if req.RateMarkup != nil {
		if err := set(payment.SettingKeyUsdtRateMarkup, strconv.FormatFloat(*req.RateMarkup, 'f', -1, 64)); err != nil {
			return err
		}
	}
	if req.AmountOffset != nil {
		if err := set(payment.SettingKeyUsdtAmountOffset, strconv.FormatFloat(*req.AmountOffset, 'f', -1, 64)); err != nil {
			return err
		}
	}
	if req.AmountTolerance != nil {
		if err := set(payment.SettingKeyUsdtAmountTolerance, strconv.FormatFloat(*req.AmountTolerance, 'f', -1, 64)); err != nil {
			return err
		}
	}
	if req.ConfirmSeconds != nil {
		if err := set(payment.SettingKeyUsdtConfirmSeconds, strconv.Itoa(*req.ConfirmSeconds)); err != nil {
			return err
		}
	}
	if req.MonitorIntervalSeconds != nil {
		if err := set(payment.SettingKeyUsdtMonitorIntervalSeconds, strconv.Itoa(*req.MonitorIntervalSeconds)); err != nil {
			return err
		}
	}
	if req.QueryMinutesBack != nil {
		if err := set(payment.SettingKeyUsdtQueryMinutesBack, strconv.Itoa(*req.QueryMinutesBack)); err != nil {
			return err
		}
	}
	if req.OrderTimeoutSeconds != nil {
		if err := set(payment.SettingKeyUsdtOrderTimeoutSeconds, strconv.Itoa(*req.OrderTimeoutSeconds)); err != nil {
			return err
		}
	}

	// per-chain
	for chain, cu := range req.Chains {
		if !payment.IsSupportedChain(chain) {
			return fmt.Errorf("unsupported chain: %s", chain)
		}
		adapter := s.usdt.Adapter(chain)
		if cu.Address != nil {
			addr := *cu.Address
			if addr != "" && adapter != nil && !adapter.ValidateAddress(addr) {
				return fmt.Errorf("invalid %s address", chain)
			}
			if err := set(payment.UsdtChainSettingKey(chain, "address"), addr); err != nil {
				return err
			}
		}
		if cu.Enabled != nil {
			if err := set(payment.UsdtChainSettingKey(chain, "enabled"), strconv.FormatBool(*cu.Enabled)); err != nil {
				return err
			}
		}
		if cu.APIBaseURL != nil {
			if err := set(payment.UsdtChainSettingKey(chain, "api_base_url"), *cu.APIBaseURL); err != nil {
				return err
			}
		}
		if cu.APIKey != nil && *cu.APIKey != "" {
			if err := set(payment.UsdtChainSettingKey(chain, "api_key"), *cu.APIKey); err != nil {
				return err
			}
		}
	}

	if s.usdt != nil {
		s.usdt.Reload(ctx)
	}
	return nil
}
