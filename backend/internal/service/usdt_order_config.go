package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/payment"
)

// AdminUsdtConfig USDT 后台配置 DTO。
// 敏感字段 TronAPIKey 在 GET 时只返回 has_tron_api_key，PUT 空字符串表示保留原值。
type AdminUsdtConfig struct {
	Enabled                bool    `json:"enabled"`
	ReceivingAddress       string  `json:"receiving_address"`
	TronAPIBaseURL         string  `json:"tron_api_base_url"`
	HasTronAPIKey          bool    `json:"has_tron_api_key"`
	ManualRate             float64 `json:"manual_rate"`
	RateAutoFetch          bool    `json:"rate_auto_fetch"`
	RateMarkup             float64 `json:"rate_markup"`
	AmountOffset           float64 `json:"amount_offset"`
	ConfirmSeconds         int     `json:"confirm_seconds"`
	MonitorIntervalSeconds int     `json:"monitor_interval_seconds"`
	QueryMinutesBack       int     `json:"query_minutes_back"`
	OrderTimeoutSeconds    int     `json:"order_timeout_seconds"`
}

// AdminUsdtConfigUpdate 更新请求（nil 指针=不变；TronAPIKey 空字符串=保留原值）。
type AdminUsdtConfigUpdate struct {
	Enabled                *bool    `json:"enabled"`
	ReceivingAddress       *string  `json:"receiving_address"`
	TronAPIBaseURL         *string  `json:"tron_api_base_url"`
	TronAPIKey             *string  `json:"tron_api_key"` // 空字符串=保留原值
	ManualRate             *float64 `json:"manual_rate"`
	RateAutoFetch          *bool    `json:"rate_auto_fetch"`
	RateMarkup             *float64 `json:"rate_markup"`
	AmountOffset           *float64 `json:"amount_offset"`
	ConfirmSeconds         *int     `json:"confirm_seconds"`
	MonitorIntervalSeconds *int     `json:"monitor_interval_seconds"`
	QueryMinutesBack       *int     `json:"query_minutes_back"`
	OrderTimeoutSeconds    *int     `json:"order_timeout_seconds"`
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
	return &AdminUsdtConfig{
		Enabled:                get(payment.SettingKeyUsdtEnabled) == "true",
		ReceivingAddress:       get(payment.SettingKeyUsdtReceivingAddress),
		TronAPIBaseURL:         get(payment.SettingKeyUsdtTronAPIBaseURL),
		HasTronAPIKey:          get(payment.SettingKeyUsdtTronAPIKey) != "",
		ManualRate:             parseFloat(get(payment.SettingKeyUsdtManualRate)),
		RateAutoFetch:          get(payment.SettingKeyUsdtRateAutoFetch) == "true",
		RateMarkup:             parseFloat(get(payment.SettingKeyUsdtRateMarkup)),
		AmountOffset:           parseFloat(get(payment.SettingKeyUsdtAmountOffset)),
		ConfirmSeconds:         parseInt(get(payment.SettingKeyUsdtConfirmSeconds)),
		MonitorIntervalSeconds: parseInt(get(payment.SettingKeyUsdtMonitorIntervalSeconds)),
		QueryMinutesBack:       parseInt(get(payment.SettingKeyUsdtQueryMinutesBack)),
		OrderTimeoutSeconds:    parseInt(get(payment.SettingKeyUsdtOrderTimeoutSeconds)),
	}, nil
}

// UpdateAdminUsdtConfig 更新配置（nil 不变；TronAPIKey 空字符串保留原值）。保存后 Reload。
func (s *UsdtOrderService) UpdateAdminUsdtConfig(ctx context.Context, req *AdminUsdtConfigUpdate) error {
	if req == nil {
		return nil
	}
	set := func(key, value string) error {
		return s.settingRepo.Set(ctx, key, value)
	}

	if req.ReceivingAddress != nil {
		addr := *req.ReceivingAddress
		if addr != "" && !payment.ValidateTronAddress(addr) {
			return fmt.Errorf("invalid TRON address")
		}
		if err := set(payment.SettingKeyUsdtReceivingAddress, addr); err != nil {
			return err
		}
	}
	if req.Enabled != nil {
		if err := set(payment.SettingKeyUsdtEnabled, strconv.FormatBool(*req.Enabled)); err != nil {
			return err
		}
	}
	if req.TronAPIBaseURL != nil {
		if err := set(payment.SettingKeyUsdtTronAPIBaseURL, *req.TronAPIBaseURL); err != nil {
			return err
		}
	}
	if req.TronAPIKey != nil && *req.TronAPIKey != "" {
		if err := set(payment.SettingKeyUsdtTronAPIKey, *req.TronAPIKey); err != nil {
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

	if s.usdt != nil {
		s.usdt.Reload(ctx)
	}
	return nil
}
