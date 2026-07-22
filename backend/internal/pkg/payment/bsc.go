package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"
)

// ===== BNB Smart Chain (BEP20) 链适配器 — 免费公共 RPC eth_getLogs =====
//
// 不依赖已收费的 Etherscan V2 tokentx。改为直接打免费公共 BSC JSON-RPC：
//   eth_blockNumber       拿链头高度
//   eth_getLogs           按 USDT 合约 + Transfer 事件 + 收款地址(indexed to) 过滤拉日志
//   eth_getBlockByNumber  仅为命中的区块补真实时间戳
// 收款匹配/确认逻辑全在 watcher 里按金额+时间做，本适配器只负责"查到流入并带准时间戳"。

const (
	// UsdtBEP20Contract 是 BSC 上的 Binance-Peg USDT(BSC-USD) 合约地址。
	UsdtBEP20Contract = "0x55d398326f99059fF775485246999027B3197955"
	// UsdtBEP20Decimals BEP20 USDT 精度为 18。
	UsdtBEP20Decimals = 18

	// DefaultBSCRPCURL 默认免费公共 BSC RPC（keyless，支持 eth_getLogs）。
	// 注意：官方 bsc-dataseed*.binance.org 关闭了 eth_getLogs，绝不能用作默认。
	DefaultBSCRPCURL = "https://bsc-rpc.publicnode.com"

	// erc20TransferTopic = keccak256("Transfer(address,address,uint256)")，所有 ERC20/BEP20 通用。
	erc20TransferTopic = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"

	// bscBlockSeconds 是 BSC 出块间隔（秒），用于把回看时间窗换算成区块数。
	bscBlockSeconds = 3
	// bscGetLogsMaxRange 是单次 eth_getLogs 的区块跨度（公共免费节点普遍限制 getLogs 跨度 <=50 块，
	// 故取 50 分页查询：跨度=50、首尾差 49，安全低于"0-50"上限，杜绝 -32602）。
	bscGetLogsMaxRange = 50
	// bscLookbackCapBlocks 是【时间窗口路径】fromBlock 回看的区块上限。
	// ⚠️ BSC 2025 Maxwell 硬分叉后出块降到亚秒级(~0.45s)，150 块只覆盖 ~1 分钟、小于确认窗(60s)，
	// 会漏收到账。生产走下面的【游标扫描路径】(QueryIncomingCursor)，此常量仅在无 settings 写能力时兜底。
	bscLookbackCapBlocks = 150
)

// 游标扫描（cursoredAdapter）相关常量：按实测出块速度换算确认滞后 + 持久化游标续扫，抗出块提速与重启。
const (
	// bscConfirmSafetyBlocks 在"确认滞后块数"上额外多留的安全余量(块)，确保游标绝不越过未确认块。
	bscConfirmSafetyBlocks = 40
	// bscFallbackBlockSeconds 测量出块时间失败时的保守回退(秒/块)。
	// 取偏小值(块更快)→确认滞后块数偏大→游标更靠后→只会多回扫、不会漏。
	bscFallbackBlockSeconds = 0.4
	// bscMaxScanPagesPerCycle 单轮最多扫的页数(每页 bscGetLogsMaxRange 块)，防免费节点限流风暴；
	// 追赶超出则本轮扫一段、游标推进、下轮继续(不跳块)。
	bscMaxScanPagesPerCycle = 24
	// bscBlockTimeSampleBlocks 测量出块时间的采样跨度(块)。
	bscBlockTimeSampleBlocks = 1000
)

// 冷启动/追赶回补的「覆盖时长」——实际取 max(可匹配窗口 + 余量, bscMinCatchupCover)，
// 随 OrderTimeout 配置自适应，再按实测出块速度换算成块数（时间基准、抗未来出块再提速）。
// 保证绝不把"仍可入账"的历史块跳过/钳掉：被钳/跳的块必然已超过订单可匹配窗口、无单可对。
const (
	// bscCatchupSafetyMargin 在"可匹配窗口"上额外多留的余量(应对确认滞后/出块波动)。
	bscCatchupSafetyMargin = 15 * time.Minute
	// bscMinCatchupCover 回补覆盖的时长下限(即便可匹配窗口配得很小也至少覆盖这么久)。
	bscMinCatchupCover = 45 * time.Minute
)

// bscFallbackRPCs 是按顺序尝试的免费 keyless 公共 BSC RPC（均支持 eth_getLogs）。
// 配置的 baseURL 永远排在最前面优先尝试。
var bscFallbackRPCs = []string{
	DefaultBSCRPCURL,
	"https://bsc.drpc.org",
	"https://binance.llamarpc.com",
	"https://bsc.meowrpc.com",
	"https://1rpc.io/bnb",
}

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
		isHexDigit := (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
		if !isHexDigit {
			return false
		}
	}
	return true
}

// QueryIncoming 通过免费公共 BSC RPC 的 eth_getLogs 拉取流入收款地址的 USDT(BEP20) 转账。
//
//	apiKey  当前未用（公共节点无需 Key；保留以兼容未来付费节点）。
//	baseURL 可选，空时用 DefaultBSCRPCURL；并按 bscFallbackRPCs 自动故障转移。
//	minTimestampMs 回看窗口下界，换算成 fromBlock。
func (a *bscAdapter) QueryIncoming(ctx context.Context, address, apiKey, baseURL string, minTimestampMs int64) ([]IncomingTransfer, error) {
	_ = apiKey
	endpoints := bscEndpoints(baseURL)

	head, used, err := a.blockNumber(ctx, endpoints)
	if err != nil {
		return nil, fmt.Errorf("bsc eth_blockNumber: %w", err)
	}
	// 同一轮 poll 把 head 所在节点 pin 到最前，避免 getLogs/getBlock 落到滞后的另一后端导致漏扫。
	endpoints = pinEndpointFirst(endpoints, used)

	fromBlock := bscFromBlock(head, minTimestampMs)

	padded := padTopicAddress(address) // 0x + 24 个 0 + 40 hex
	addrLower := strings.ToLower(strings.TrimSpace(address))

	tsCache := make(map[string]int64) // blockNumber(hex) -> blockTimeMs，去重补时间戳
	out := make([]IncomingTransfer, 0, 8)

	for lo := fromBlock; lo <= head; lo += bscGetLogsMaxRange {
		hi := lo + bscGetLogsMaxRange - 1
		if hi > head {
			hi = head
		}
		logs, err := a.getLogs(ctx, endpoints, lo, hi, padded)
		if err != nil {
			return nil, fmt.Errorf("bsc eth_getLogs [%d-%d]: %w", lo, hi, err)
		}
		for _, lg := range logs {
			// 防御：标准 Transfer 有 3 个 topic，每个 32 字节(0x+64hex)。畸形日志直接跳过，避免切片越界 panic。
			if len(lg.Topics) < 3 || len(lg.Topics[1]) < 40 || len(lg.Topics[2]) < 40 {
				continue
			}
			to := "0x" + strings.ToLower(lg.Topics[2][len(lg.Topics[2])-40:])
			if to != addrLower {
				continue // server 端已按 to 过滤；双保险。
			}
			from := "0x" + strings.ToLower(lg.Topics[1][len(lg.Topics[1])-40:])
			human, ok := humanFromAtomic(hexToDecString(lg.Data), UsdtBEP20Decimals)
			if !ok || human <= 0 {
				continue
			}
			// 为命中区块补真实时间戳（watcher 的确认/有效期判定全靠它，绝不能为 0）。
			blockMs, err := a.blockTimeMs(ctx, endpoints, lg.BlockNumber, tsCache)
			if err != nil {
				continue // 拿不到时间戳本轮跳过，下个周期再试；返回 0 会绕过 finality 保护。
			}
			if minTimestampMs > 0 && blockMs > 0 && blockMs < minTimestampMs {
				continue
			}
			out = append(out, IncomingTransfer{
				TxID:        lg.TxHash,
				From:        from,
				To:          to,
				AmountHuman: human,
				BlockTimeMs: blockMs,
			})
		}
	}
	return out, nil
}

// bscScanRange 计算本轮游标扫描应覆盖的区块区间 [from, scanTo]（含端点，纯函数、便于单测）。
//
//	safeHead：已用墙钟校正过的「已确认链头」(见 bscConfirmedSafeHead)，游标绝不越过它。
//	有真实游标(fromCursor>0)：一律 from=游标+1 连续续扫，绝不跳块；仅当旧到超 maxStaleBlocks 才有界钳制
//	  (被钳掉的块都已远超订单可匹配窗口、无单可对，跳过无害)。
//	冷启动(fromCursor==0)：从 safeHead-coldStartBlocks 起有界回补(覆盖订单最长可匹配窗口)。
//	scanTo=safeHead，但单轮跨度不超 bscMaxScanPagesPerCycle*bscGetLogsMaxRange(下轮续扫、不跳块)。
//	hasWork=false 表示当前无新确认块可扫。
func bscScanRange(safeHead, fromCursor, coldStartBlocks, maxStaleBlocks uint64) (from, scanTo uint64, hasWork bool) {
	if safeHead == 0 {
		return 0, 0, false
	}
	if fromCursor == 0 {
		// 冷启动：有界回补，避免扫全链。
		if safeHead > coldStartBlocks {
			from = safeHead - coldStartBlocks
		} else {
			from = 0
		}
	} else {
		// 有真实游标：从游标+1 连续续扫(绝不跳块)；仅极旧时按 maxStale 有界钳制。
		from = fromCursor + 1
		var floor uint64
		if safeHead > maxStaleBlocks {
			floor = safeHead - maxStaleBlocks
		}
		if from < floor {
			from = floor
		}
	}
	if from > safeHead {
		return 0, 0, false
	}
	scanTo = safeHead
	if maxSpan := uint64(bscMaxScanPagesPerCycle) * uint64(bscGetLogsMaxRange); scanTo-from+1 > maxSpan {
		scanTo = from + maxSpan - 1
	}
	return from, scanTo, true
}

// blocksFor 按出块速度把时长换算成块数(至少 1)。
func blocksFor(d time.Duration, blockSec float64) uint64 {
	if blockSec <= 0 {
		blockSec = bscFallbackBlockSeconds
	}
	n := uint64(math.Ceil(d.Seconds() / blockSec))
	if n == 0 {
		n = 1
	}
	return n
}

// measureBlockTime 用 head 与 head-sample 的时间戳估算当前出块间隔(秒/块)，并返回 head 的时间戳(ms)。
func (a *bscAdapter) measureBlockTime(ctx context.Context, endpoints []string, head uint64) (blockSec float64, headTsMs int64, err error) {
	if head <= bscBlockTimeSampleBlocks {
		return 0, 0, fmt.Errorf("chain too short to sample block time")
	}
	c := map[string]int64{}
	hiMs, err := a.blockTimeMs(ctx, endpoints, fmt.Sprintf("0x%x", head), c)
	if err != nil {
		return 0, 0, err
	}
	loMs, err := a.blockTimeMs(ctx, endpoints, fmt.Sprintf("0x%x", head-bscBlockTimeSampleBlocks), c)
	if err != nil {
		return 0, hiMs, err
	}
	sec := float64(hiMs-loMs) / 1000.0
	if sec <= 0 {
		return 0, hiMs, fmt.Errorf("non-positive block time span")
	}
	return sec / float64(bscBlockTimeSampleBlocks), hiMs, nil
}

// bscConfirmedSafeHead 返回「墙钟已确认」(时间戳 <= headTs - confirmDur)的最高块。
// 先用测得出块速度估个候选，再用候选真实时间戳按 head↔候选的实际(近期)速率校正一次——
// 应对「1000 块平均出块时间」高于当前 tip 速率导致按块数算的 safeHead 偏新、被 watcher 判太新而丢的情形。
// 全程用链上时钟(headTs)算年龄，避免服务器-链时钟偏差。返回 0 表示当前无已确认块。
func (a *bscAdapter) bscConfirmedSafeHead(ctx context.Context, endpoints []string, head uint64, headTsMs int64, blockSec float64, confirmDur time.Duration) (uint64, error) {
	estLag := uint64(math.Ceil(confirmDur.Seconds()/blockSec)) + bscConfirmSafetyBlocks
	if head <= estLag {
		return 0, nil
	}
	cand := head - estLag
	candTsMs, err := a.blockTimeMs(ctx, endpoints, fmt.Sprintf("0x%x", cand), map[string]int64{})
	if err != nil {
		return 0, err
	}
	ageSec := float64(headTsMs-candTsMs) / 1000.0
	if ageSec < confirmDur.Seconds() {
		// 候选太新：用 head↔cand 的真实近期速率重估，多退到够老。
		realRate := (float64(headTsMs-candTsMs) / 1000.0) / float64(head-cand)
		if realRate <= 0 {
			realRate = bscFallbackBlockSeconds
		}
		extra := uint64(math.Ceil((confirmDur.Seconds()-ageSec)/realRate)) + bscConfirmSafetyBlocks
		if cand <= extra {
			return 0, nil
		}
		cand -= extra
	}
	return cand, nil
}

// QueryIncomingCursor 用持久化区块游标扫描 BSC 上流入 address 的 USDT 转账（实现 cursoredAdapter）。
//
// 只扫「已确认」块 [from, head-确认滞后块]：返回的转账都够老，watcher 不会因太新而丢弃；
// 游标只推进到已成功扫完的最高块。任一分页/取时间戳出错都不推进游标(返回原 fromCursor)、丢弃本轮结果，
// 下轮整体重扫（幂等由 usdt_orders.trade_no 唯一索引兜底）。绝不漏块、抗 BSC 出块提速与进程重启。
func (a *bscAdapter) QueryIncomingCursor(ctx context.Context, address, apiKey, baseURL string, fromCursor uint64, confirmDur, matchWindow time.Duration) ([]IncomingTransfer, uint64, error) {
	_ = apiKey
	endpoints := bscEndpoints(baseURL)

	head, used, err := a.blockNumber(ctx, endpoints)
	if err != nil {
		return nil, fromCursor, fmt.Errorf("bsc eth_blockNumber: %w", err)
	}
	endpoints = pinEndpointFirst(endpoints, used)

	blockSec, headTsMs, err := a.measureBlockTime(ctx, endpoints, head)
	if err != nil || blockSec <= 0 {
		blockSec = bscFallbackBlockSeconds
	}
	if headTsMs <= 0 {
		// 拿不到 head 时间戳：无法做墙钟确认校正，本轮跳过(不冒进推进游标)。
		return nil, fromCursor, nil
	}
	// safeHead 用墙钟(head 链上时钟)校正，确保返回的块都够老、watcher 不会判太新而丢。
	safeHead, err := a.bscConfirmedSafeHead(ctx, endpoints, head, headTsMs, blockSec, confirmDur)
	if err != nil {
		return nil, fromCursor, fmt.Errorf("bsc confirmed head: %w", err)
	}
	// 回补/钳制下限 = max(可匹配窗口 + 余量, 最小覆盖)，据实测出块速度换算成块。
	// 冷启动与"极旧游标钳制"都用它：随 OrderTimeout 自适应，绝不跳过/钳掉仍可入账的历史块。
	cover := matchWindow + bscCatchupSafetyMargin
	if cover < bscMinCatchupCover {
		cover = bscMinCatchupCover
	}
	coverBlocks := blocksFor(cover, blockSec)
	from, scanTo, hasWork := bscScanRange(safeHead, fromCursor, coverBlocks, coverBlocks)
	if !hasWork {
		return nil, fromCursor, nil
	}

	padded := padTopicAddress(address)
	addrLower := strings.ToLower(strings.TrimSpace(address))
	tsCache := make(map[string]int64)
	out := make([]IncomingTransfer, 0, 8)

	for lo := from; lo <= scanTo; lo += bscGetLogsMaxRange {
		hi := lo + bscGetLogsMaxRange - 1
		if hi > scanTo {
			hi = scanTo
		}
		logs, err := a.getLogs(ctx, endpoints, lo, hi, padded)
		if err != nil {
			return nil, fromCursor, fmt.Errorf("bsc eth_getLogs [%d-%d]: %w", lo, hi, err)
		}
		for _, lg := range logs {
			if len(lg.Topics) < 3 || len(lg.Topics[1]) < 40 || len(lg.Topics[2]) < 40 {
				continue
			}
			to := "0x" + strings.ToLower(lg.Topics[2][len(lg.Topics[2])-40:])
			if to != addrLower {
				continue
			}
			fromAddr := "0x" + strings.ToLower(lg.Topics[1][len(lg.Topics[1])-40:])
			human, ok := humanFromAtomic(hexToDecString(lg.Data), UsdtBEP20Decimals)
			if !ok || human <= 0 {
				continue
			}
			blockMs, err := a.blockTimeMs(ctx, endpoints, lg.BlockNumber, tsCache)
			if err != nil {
				// 拿不到时间戳：不推进游标，本轮丢弃、下轮重扫（避免 BlockTimeMs=0 绕过 finality 保护）。
				return nil, fromCursor, fmt.Errorf("bsc block time %s: %w", lg.BlockNumber, err)
			}
			out = append(out, IncomingTransfer{
				TxID:        lg.TxHash,
				From:        fromAddr,
				To:          to,
				AmountHuman: human,
				BlockTimeMs: blockMs,
			})
		}
	}
	return out, scanTo, nil
}

// ----- JSON-RPC -----

type bscLog struct {
	Address     string   `json:"address"`
	Topics      []string `json:"topics"`
	Data        string   `json:"data"`
	BlockNumber string   `json:"blockNumber"`
	TxHash      string   `json:"transactionHash"`
}

// bscRPCCall 对 endpoints 顺序故障转移地发一次 JSON-RPC，把 result 反序列化进 out。
// 返回实际成功的 endpoint（供调用方把同一节点 pin 在后续调用最前，避免 head 与 getLogs 落到不同后端导致的滞后不一致）。
func bscRPCCall(ctx context.Context, endpoints []string, method string, params []any, out any) (string, error) {
	reqBody, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0", "id": 1, "method": method, "params": params,
	})
	if err != nil {
		return "", err
	}
	// 显式 UA：部分公共节点会拦截空/陌生 User-Agent（默认 publicnode 不拦，但 fallback 节点更挑）。
	headers := map[string]string{"User-Agent": "sub2api-usdt/1.0"}
	var lastErr error
	for _, ep := range endpoints {
		body, status, err := adapterPostJSON(ctx, ep, reqBody, headers)
		if err != nil {
			lastErr = err
			continue
		}
		if status != 200 {
			lastErr = fmt.Errorf("rpc status %d: %s", status, truncate(string(body), 200))
			continue
		}
		var resp struct {
			Result json.RawMessage `json:"result"`
			Error  *struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			lastErr = fmt.Errorf("parse rpc response: %w (body: %s)", err, truncate(string(body), 200))
			continue
		}
		if resp.Error != nil {
			lastErr = fmt.Errorf("rpc %s error %d: %s", method, resp.Error.Code, resp.Error.Message)
			continue // 换下一个节点（范围超限/限流/节点滞后都可能）。
		}
		if out == nil {
			return ep, nil
		}
		if err := json.Unmarshal(resp.Result, out); err != nil {
			return ep, fmt.Errorf("decode rpc result: %w", err)
		}
		return ep, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no rpc endpoint available")
	}
	return "", lastErr
}

func (a *bscAdapter) blockNumber(ctx context.Context, endpoints []string) (uint64, string, error) {
	var hexHead string
	used, err := bscRPCCall(ctx, endpoints, "eth_blockNumber", []any{}, &hexHead)
	if err != nil {
		return 0, "", err
	}
	return hexToUint64(hexHead), used, nil
}

func (a *bscAdapter) getLogs(ctx context.Context, endpoints []string, from, to uint64, paddedTo string) ([]bscLog, error) {
	filter := map[string]any{
		"fromBlock": fmt.Sprintf("0x%x", from),
		"toBlock":   fmt.Sprintf("0x%x", to),
		"address":   UsdtBEP20Contract,
		// topics: [Transfer, anyFrom, to=收款地址]，节点端按 to 过滤。
		"topics": []any{erc20TransferTopic, nil, paddedTo},
	}
	var logs []bscLog
	if _, err := bscRPCCall(ctx, endpoints, "eth_getLogs", []any{filter}, &logs); err != nil {
		return nil, err
	}
	return logs, nil
}

func (a *bscAdapter) blockTimeMs(ctx context.Context, endpoints []string, blockHex string, cache map[string]int64) (int64, error) {
	if ms, ok := cache[blockHex]; ok {
		return ms, nil
	}
	var blk struct {
		Timestamp string `json:"timestamp"`
	}
	if _, err := bscRPCCall(ctx, endpoints, "eth_getBlockByNumber", []any{blockHex, false}, &blk); err != nil {
		return 0, err
	}
	if blk.Timestamp == "" {
		return 0, fmt.Errorf("empty block timestamp for %s", blockHex)
	}
	ms := int64(hexToUint64(blk.Timestamp)) * 1000
	cache[blockHex] = ms
	return ms, nil
}

// pinEndpointFirst 把 first 排到最前（同一轮 poll 复用同一节点，保证 head 与 getLogs 一致），保留其余作 fallback。
func pinEndpointFirst(endpoints []string, first string) []string {
	if first == "" {
		return endpoints
	}
	out := make([]string, 0, len(endpoints))
	out = append(out, first)
	for _, e := range endpoints {
		if e != first {
			out = append(out, e)
		}
	}
	return out
}

// ----- 辅助 -----

// bscEndpoints 构造尝试顺序：配置的 baseURL 优先，然后是不重复的内置 fallback。
func bscEndpoints(baseURL string) []string {
	eps := make([]string, 0, len(bscFallbackRPCs)+1)
	seen := make(map[string]bool)
	add := func(u string) {
		u = strings.TrimRight(strings.TrimSpace(u), "/")
		if u == "" || seen[u] {
			return
		}
		seen[u] = true
		eps = append(eps, u)
	}
	add(baseURL)
	for _, u := range bscFallbackRPCs {
		add(u)
	}
	return eps
}

// bscFromBlock 把回看时间窗换算成 fromBlock：head - ceil(lookbackSeconds/3)，带安全上限。
func bscFromBlock(head uint64, minTimestampMs int64) uint64 {
	if minTimestampMs <= 0 {
		if head > bscLookbackCapBlocks {
			return head - bscLookbackCapBlocks
		}
		return 0
	}
	lookbackSec := (time.Now().UnixMilli() - minTimestampMs) / 1000
	if lookbackSec < 0 {
		lookbackSec = 0
	}
	blocks := uint64((lookbackSec+bscBlockSeconds-1)/bscBlockSeconds) + 5
	if blocks > bscLookbackCapBlocks {
		blocks = bscLookbackCapBlocks
	}
	if head > blocks {
		return head - blocks
	}
	return 0
}

// padTopicAddress 把 EVM 地址左补零成 32 字节 topic：0x + 24 个 0 + 40 hex（小写）。
func padTopicAddress(addr string) string {
	a := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(addr), "0x"))
	if len(a) > 64 {
		a = a[len(a)-64:]
	}
	return "0x" + strings.Repeat("0", 64-len(a)) + a
}

// hexToUint64 解析 0x 十六进制为 uint64。
func hexToUint64(h string) uint64 {
	n := new(big.Int)
	n.SetString(strings.TrimPrefix(strings.TrimSpace(h), "0x"), 16)
	return n.Uint64()
}

// hexToDecString 把 0x 十六进制（uint256）转成十进制字符串，喂给 humanFromAtomic。
func hexToDecString(h string) string {
	n := new(big.Int)
	if _, ok := n.SetString(strings.TrimPrefix(strings.TrimSpace(h), "0x"), 16); !ok {
		return "0"
	}
	return n.String()
}
