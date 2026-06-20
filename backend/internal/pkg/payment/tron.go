package payment

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"net/url"
	"strconv"
	"strings"
)

// TRON / TRC20-USDT 常量与工具。

const (
	// UsdtTRC20Contract 是 TRON 主网 USDT(TRC20) 合约地址（base58）。
	UsdtTRC20Contract = "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"
	// UsdtDecimals 是 USDT(TRC20) 的精度（6 位小数）。
	UsdtDecimals = 6
	// usdtAtomicPerUnit = 10^6，1 USDT 的最小单位数。
	usdtAtomicPerUnit = 1_000_000

	// DefaultTronAPIBaseURL 默认 TronGrid 网关。
	DefaultTronAPIBaseURL = "https://api.trongrid.io"
)

const b58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// base58Decode 解码 base58 字符串，返回原始字节。
func base58Decode(s string) ([]byte, bool) {
	if s == "" {
		return nil, false
	}
	result := big.NewInt(0)
	radix := big.NewInt(58)
	for _, r := range s {
		idx := strings.IndexRune(b58Alphabet, r)
		if idx < 0 {
			return nil, false
		}
		result.Mul(result, radix)
		result.Add(result, big.NewInt(int64(idx)))
	}
	decoded := result.Bytes()
	// 前导 '1' 对应前导 0 字节
	leading := 0
	for _, r := range s {
		if r == '1' {
			leading++
		} else {
			break
		}
	}
	out := make([]byte, leading+len(decoded))
	copy(out[leading:], decoded)
	return out, true
}

// ValidateTronAddress 校验是否是合法的 TRON base58check 地址（T 开头、0x41 前缀、双 SHA256 校验和）。
func ValidateTronAddress(addr string) bool {
	addr = strings.TrimSpace(addr)
	if len(addr) != 34 || !strings.HasPrefix(addr, "T") {
		return false
	}
	decoded, ok := base58Decode(addr)
	if !ok || len(decoded) != 25 {
		return false
	}
	if decoded[0] != 0x41 {
		return false
	}
	payload := decoded[:21]
	checksum := decoded[21:]
	h1 := sha256.Sum256(payload)
	h2 := sha256.Sum256(h1[:])
	return bytes.Equal(h2[:4], checksum)
}

// UsdtToAtomic 把人类可读的 USDT 金额换算成链上最小单位（6 位小数）。
func UsdtToAtomic(amount float64) int64 {
	return int64(math.Round(amount * usdtAtomicPerUnit))
}

// AtomicToUsdt 把链上最小单位字符串（来自 TronGrid 的 value 字段）换算成 USDT 金额。
func AtomicToUsdt(atomic string) (float64, bool) {
	atomic = strings.TrimSpace(atomic)
	if atomic == "" {
		return 0, false
	}
	v, ok := new(big.Int).SetString(atomic, 10)
	if !ok {
		return 0, false
	}
	// 用 big.Rat 保证精度
	r := new(big.Rat).SetFrac(v, big.NewInt(usdtAtomicPerUnit))
	f, _ := r.Float64()
	return f, true
}

// FormatUsdt 把 USDT 金额格式化为固定 6 位小数字符串（用于展示精确应付金额）。
func FormatUsdt(amount float64) string {
	return strconv.FormatFloat(amount, 'f', UsdtDecimals, 64)
}

// ===== TRON(TRC20) 链适配器 =====

type tronAdapter struct{}

func (a *tronAdapter) Chain() string { return ChainTRC20 }

func (a *tronAdapter) Decimals() int { return UsdtDecimals }

func (a *tronAdapter) ValidateAddress(addr string) bool { return ValidateTronAddress(addr) }

// QueryIncoming 通过 TronGrid /v1/accounts/{addr}/transactions/trc20 拉取流入的 USDT 转账。
func (a *tronAdapter) QueryIncoming(ctx context.Context, address, apiKey, baseURL string, minTimestampMs int64) ([]IncomingTransfer, error) {
	if baseURL == "" {
		baseURL = DefaultTronAPIBaseURL
	}
	q := url.Values{}
	q.Set("only_to", "true")
	q.Set("contract_address", UsdtTRC20Contract)
	q.Set("limit", "200")
	q.Set("order_by", "block_timestamp,desc")
	if minTimestampMs > 0 {
		q.Set("min_timestamp", strconv.FormatInt(minTimestampMs, 10))
	}
	endpoint := fmt.Sprintf("%s/v1/accounts/%s/transactions/trc20?%s", baseURL, url.PathEscape(address), q.Encode())

	headers := map[string]string{}
	if apiKey != "" {
		headers["TRON-PRO-API-KEY"] = apiKey
	}
	body, status, err := adapterGet(ctx, endpoint, headers)
	if err != nil {
		return nil, err
	}
	if status != 200 {
		return nil, fmt.Errorf("trongrid status %d: %s", status, truncate(string(body), 300))
	}

	var parsed struct {
		Data []struct {
			TransactionID  string `json:"transaction_id"`
			From           string `json:"from"`
			To             string `json:"to"`
			Value          string `json:"value"`
			Type           string `json:"type"`
			BlockTimestamp int64  `json:"block_timestamp"`
			TokenInfo      struct {
				Address string `json:"address"`
			} `json:"token_info"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("parse trongrid response: %w (body: %s)", err, truncate(string(body), 300))
	}

	out := make([]IncomingTransfer, 0, len(parsed.Data))
	for _, d := range parsed.Data {
		if d.Type != "" && d.Type != "Transfer" {
			continue
		}
		if d.To != address {
			continue
		}
		if d.TokenInfo.Address != "" && d.TokenInfo.Address != UsdtTRC20Contract {
			continue
		}
		human, ok := humanFromAtomic(d.Value, UsdtDecimals)
		if !ok || human <= 0 {
			continue
		}
		out = append(out, IncomingTransfer{
			TxID:        d.TransactionID,
			From:        d.From,
			To:          d.To,
			AmountHuman: human,
			BlockTimeMs: d.BlockTimestamp,
		})
	}
	return out, nil
}
