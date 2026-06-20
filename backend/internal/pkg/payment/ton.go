package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// ===== TON 链适配器（USDT 为 Jetton）=====

const (
	// UsdtTONJettonMaster 是 TON 上 Tether USDT 的 Jetton master 地址（user-friendly）。
	UsdtTONJettonMaster = "EQCxE6mUtQJKFnGfaROTKOt1lZbDiiX1kCixRv7Nw2Id_sDs"
	// UsdtTONDecimals TON 上 USDT 精度为 6。
	UsdtTONDecimals = 6
	// DefaultTONAPIBaseURL toncenter v3 网关。
	DefaultTONAPIBaseURL = "https://toncenter.com/api/v3"
)

type tonAdapter struct{}

func (a *tonAdapter) Chain() string { return ChainTON }

func (a *tonAdapter) Decimals() int { return UsdtTONDecimals }

var (
	tonRawRe      = regexp.MustCompile(`^-?[0-9]+:[0-9a-fA-F]{64}$`)
	tonFriendlyRe = regexp.MustCompile(`^[A-Za-z0-9_-]{48}$`)
)

// ValidateAddress 校验 TON 地址：raw(0:64hex) 或 user-friendly(48 位 base64url，EQ/UQ/kQ/0Q 开头)。
func (a *tonAdapter) ValidateAddress(addr string) bool {
	addr = strings.TrimSpace(addr)
	if tonRawRe.MatchString(addr) {
		return true
	}
	if tonFriendlyRe.MatchString(addr) {
		switch addr[:2] {
		case "EQ", "UQ", "kQ", "0Q", "Ef", "Uf":
			return true
		}
	}
	return false
}

// QueryIncoming 通过 toncenter v3 /jetton/transfers 拉取流入收款地址的 USDT(Jetton) 转账。
// owner_address + direction=in 已由 API 过滤出"收到"的转账，故无需在本地比对地址（规避 TON 地址多形态）。
func (a *tonAdapter) QueryIncoming(ctx context.Context, address, apiKey, baseURL string, minTimestampMs int64) ([]IncomingTransfer, error) {
	if baseURL == "" {
		baseURL = DefaultTONAPIBaseURL
	}
	q := url.Values{}
	q.Set("owner_address", address)
	q.Set("jetton_master", UsdtTONJettonMaster)
	q.Set("direction", "in")
	q.Set("limit", "100")
	q.Set("sort", "desc")
	if minTimestampMs > 0 {
		q.Set("start_utime", fmt.Sprintf("%d", minTimestampMs/1000))
	}
	endpoint := strings.TrimRight(baseURL, "/") + "/jetton/transfers?" + q.Encode()

	headers := map[string]string{}
	if apiKey != "" {
		headers["X-API-Key"] = apiKey
	}
	body, status, err := adapterGet(ctx, endpoint, headers)
	if err != nil {
		return nil, err
	}
	if status != 200 {
		return nil, fmt.Errorf("toncenter status %d: %s", status, truncate(string(body), 300))
	}

	var parsed struct {
		JettonTransfers []struct {
			TransactionHash    string `json:"transaction_hash"`
			Source             string `json:"source"`
			Destination        string `json:"destination"`
			Amount             string `json:"amount"`
			TransactionNow     int64  `json:"transaction_now"`
			TransactionAborted bool   `json:"transaction_aborted"`
		} `json:"jetton_transfers"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("parse toncenter response: %w (body: %s)", err, truncate(string(body), 300))
	}

	// 不在本地按 jetton_master 复核：API 返回的是 raw(0:HEX) 形态，与我们的 EQ 常量永不相等，
	// 会误删所有转账；服务端已用 jetton_master 查询参数过滤，直接信任。
	out := make([]IncomingTransfer, 0, len(parsed.JettonTransfers))
	for _, tr := range parsed.JettonTransfers {
		if tr.TransactionAborted {
			// 失败/弹回的转账不入账。
			continue
		}
		human, ok := humanFromAtomic(tr.Amount, UsdtTONDecimals)
		if !ok || human <= 0 {
			continue
		}
		out = append(out, IncomingTransfer{
			TxID:        tr.TransactionHash,
			From:        tr.Source,
			To:          tr.Destination,
			AmountHuman: human,
			BlockTimeMs: tr.TransactionNow * 1000,
		})
	}
	return out, nil
}
