package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ===== BNB Smart Chain (BEP20) 链适配器 =====

const (
	// UsdtBEP20Contract 是 BSC 上的 Binance-Peg USDT(BSC-USD) 合约地址。
	UsdtBEP20Contract = "0x55d398326f99059fF775485246999027B3197955"
	// UsdtBEP20Decimals BEP20 USDT 精度为 18。
	UsdtBEP20Decimals = 18
	// DefaultEtherscanV2BaseURL Etherscan V2 统一网关（2025 起 BscScan 并入，chainid=56 即 BSC）。
	DefaultEtherscanV2BaseURL = "https://api.etherscan.io/v2/api"
	// bscChainID BSC 在 Etherscan V2 的 chainid。
	bscChainID = "56"
)

type bscAdapter struct{}

func (a *bscAdapter) Chain() string { return ChainBEP20 }

func (a *bscAdapter) Decimals() int { return UsdtBEP20Decimals }

// ValidateAddress 校验 EVM 地址：0x + 40 位十六进制（大小写不敏感，不强制 EIP-55）。
func (a *bscAdapter) ValidateAddress(addr string) bool {
	addr = strings.TrimSpace(addr)
	if len(addr) != 42 || !strings.HasPrefix(addr, "0x") {
		return false
	}
	for _, c := range addr[2:] {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// QueryIncoming 通过 Etherscan V2 (chainid=56) 的 tokentx 拉取流入收款地址的 USDT(BEP20) 转账。
func (a *bscAdapter) QueryIncoming(ctx context.Context, address, apiKey, baseURL string, minTimestampMs int64) ([]IncomingTransfer, error) {
	if baseURL == "" {
		baseURL = DefaultEtherscanV2BaseURL
	}
	q := url.Values{}
	q.Set("chainid", bscChainID)
	q.Set("module", "account")
	q.Set("action", "tokentx")
	q.Set("contractaddress", UsdtBEP20Contract)
	q.Set("address", address)
	q.Set("page", "1")
	q.Set("offset", "100")
	q.Set("sort", "desc")
	if apiKey != "" {
		q.Set("apikey", apiKey)
	}
	endpoint := baseURL + "?" + q.Encode()

	body, status, err := adapterGet(ctx, endpoint, nil)
	if err != nil {
		return nil, err
	}
	if status != 200 {
		return nil, fmt.Errorf("etherscan status %d: %s", status, truncate(string(body), 300))
	}

	var parsed struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			Hash            string `json:"hash"`
			From            string `json:"from"`
			To              string `json:"to"`
			Value           string `json:"value"`
			TokenDecimal    string `json:"tokenDecimal"`
			ContractAddress string `json:"contractAddress"`
			TimeStamp       string `json:"timeStamp"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("parse etherscan response: %w (body: %s)", err, truncate(string(body), 300))
	}
	// status="0" 且无结果是正常的"无交易"；带错误消息才报错。
	if parsed.Status != "1" && len(parsed.Result) == 0 {
		if parsed.Message != "" && !strings.EqualFold(parsed.Message, "No transactions found") && !strings.EqualFold(parsed.Message, "OK") {
			return nil, fmt.Errorf("etherscan error: %s", parsed.Message)
		}
		return nil, nil
	}

	addrLower := strings.ToLower(address)
	out := make([]IncomingTransfer, 0, len(parsed.Result))
	for _, r := range parsed.Result {
		if strings.ToLower(r.To) != addrLower {
			continue
		}
		if r.ContractAddress != "" && !strings.EqualFold(r.ContractAddress, UsdtBEP20Contract) {
			continue
		}
		decimals := UsdtBEP20Decimals
		if r.TokenDecimal != "" {
			if d, e := strconv.Atoi(r.TokenDecimal); e == nil && d > 0 {
				decimals = d
			}
		}
		human, ok := humanFromAtomic(r.Value, decimals)
		if !ok || human <= 0 {
			continue
		}
		var blockMs int64
		if ts, e := strconv.ParseInt(r.TimeStamp, 10, 64); e == nil {
			blockMs = ts * 1000
		}
		if minTimestampMs > 0 && blockMs > 0 && blockMs < minTimestampMs {
			continue
		}
		out = append(out, IncomingTransfer{
			TxID:        r.Hash,
			From:        r.From,
			To:          r.To,
			AmountHuman: human,
			BlockTimeMs: blockMs,
		})
	}
	return out, nil
}
