package payment

import (
	"bytes"
	"context"
	"io"
	"math/big"
	"net/http"
	"time"
)

// adapterHTTPClient 所有链适配器共用的 HTTP 客户端。
var adapterHTTPClient = &http.Client{Timeout: 15 * time.Second}

// adapterGet 发起 GET 并返回 body + 状态码。
func adapterGet(ctx context.Context, endpoint string, headers map[string]string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, 0, err
	}
	for k, v := range headers {
		if v != "" {
			req.Header.Set(k, v)
		}
	}
	resp, err := adapterHTTPClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}

// adapterPostJSON 发起 POST JSON 请求并返回 body + 状态码（用于 EVM JSON-RPC）。
func adapterPostJSON(ctx context.Context, endpoint string, payload []byte, headers map[string]string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		if v != "" {
			req.Header.Set(k, v)
		}
	}
	resp, err := adapterHTTPClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}

// truncate 截断字符串用于错误日志。
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// 多链 USDT 收款的链适配器抽象。
// 每条链(TRC20/BEP20/TON)实现一个 ChainAdapter：地址校验 + 扫链查到账。
// 核心的唯一金额匹配/订单/入账逻辑与链无关，全部复用。

// IncomingTransfer 是某条链上一笔流入收款地址的 USDT 转账（金额已按该链精度换算成人类可读 USDT）。
type IncomingTransfer struct {
	TxID        string
	From        string
	To          string
	AmountHuman float64 // 已换算成 USDT（如 13.8912），匹配统一在 6 位定点(micro-USDT)上做
	BlockTimeMs int64
}

// ChainAdapter 抽象一条链的 USDT 收款能力。
type ChainAdapter interface {
	// Chain 返回链标识：trc20 / bep20 / ton。
	Chain() string
	// Decimals 返回该链 USDT 的链上精度（trc20=6, bep20=18, ton=6）。
	Decimals() int
	// ValidateAddress 校验收款地址格式是否属于本链。
	ValidateAddress(addr string) bool
	// QueryIncoming 拉取 address 在 minTimestampMs 之后收到的 USDT 转账。
	QueryIncoming(ctx context.Context, address, apiKey, baseURL string, minTimestampMs int64) ([]IncomingTransfer, error)
}

// 支持的链标识常量。
const (
	ChainTRC20 = "trc20"
	ChainBEP20 = "bep20"
	ChainTON   = "ton"
)

// SupportedChains 是当前支持的链顺序（用于前端选择器展示顺序与校验）。
// 按手续费从低到高排：BEP20/TON 便宜在前，TRC20 手续费最贵排最后。
var SupportedChains = []string{ChainBEP20, ChainTON, ChainTRC20}

// IsSupportedChain 判断链标识是否受支持。
func IsSupportedChain(chain string) bool {
	for _, c := range SupportedChains {
		if c == chain {
			return true
		}
	}
	return false
}

// newAdapters 构造所有链适配器（无状态，单例）。
func newAdapters() map[string]ChainAdapter {
	return map[string]ChainAdapter{
		ChainTRC20: &tronAdapter{},
		ChainBEP20: &bscAdapter{},
		ChainTON:   &tonAdapter{},
	}
}

// humanFromAtomic 把某精度的链上最小单位字符串精确换算成人类可读 USDT（用 big.Rat 防丢精度）。
func humanFromAtomic(atomic string, decimals int) (float64, bool) {
	if atomic == "" {
		return 0, false
	}
	v, ok := new(big.Int).SetString(atomic, 10)
	if !ok {
		return 0, false
	}
	denom := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	f, _ := new(big.Rat).SetFrac(v, denom).Float64()
	return f, true
}
