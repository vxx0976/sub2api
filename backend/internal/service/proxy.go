package service

import (
	"net"
	"net/url"
	"strconv"
	"time"
)

const (
	FallbackModeNone   = "none"
	FallbackModeProxy  = "proxy"
	FallbackModeDirect = "direct"
)

const (
	ProxyHealthHealthy  = "healthy"
	ProxyHealthDegraded = "degraded"
)

// ProxyTransportTempUnschedReasonPrefix 是代理/网络传输失败写入临时不可调度原因的前缀。
// 代理被判定故障并改投账号时，据此只清掉由代理故障造成的临时不可调度。
const ProxyTransportTempUnschedReasonPrefix = "upstream transport error"

type Proxy struct {
	ID             int64
	Name           string
	Protocol       string
	Host           string
	Port           int
	Username       string
	Password       string
	Status         string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	ExpiresAt      *time.Time
	FallbackMode   string
	BackupProxyID  *int64
	ExpiryWarnDays int

	// 故障回退：健康检查判定代理故障后的改投目标，与到期回退相互独立。
	FailureFallbackMode  string
	FailureBackupProxyID *int64
	HealthStatus         string
	HealthChangedAt      *time.Time
	HealthLastError      string
}

// IsDegraded 报告代理是否被健康检查判定为故障。
func (p *Proxy) IsDegraded() bool {
	return p.HealthStatus == ProxyHealthDegraded
}

func (p *Proxy) IsActive() bool {
	return p.Status == StatusActive
}

// IsExpired 报告代理是否已过期（基于 expires_at，与 status 无关）。
func (p *Proxy) IsExpired(now time.Time) bool {
	return p.ExpiresAt != nil && !p.ExpiresAt.After(now)
}

func (p *Proxy) URL() string {
	u := &url.URL{
		Scheme: p.Protocol,
		Host:   net.JoinHostPort(p.Host, strconv.Itoa(p.Port)),
	}
	if p.Username != "" && p.Password != "" {
		u.User = url.UserPassword(p.Username, p.Password)
	}
	return u.String()
}

type ProxyWithAccountCount struct {
	Proxy
	AccountCount   int64
	LatencyMs      *int64
	LatencyStatus  string
	LatencyMessage string
	IPAddress      string
	Country        string
	CountryCode    string
	Region         string
	City           string
	QualityStatus  string
	QualityScore   *int
	QualityGrade   string
	QualitySummary string
	QualityChecked *int64
}

type ProxyAccountSummary struct {
	ID       int64
	Name     string
	Platform string
	Type     string
	Notes    *string
}
