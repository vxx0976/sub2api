package service

import "time"

// ResolveProxyFallbackTarget 计算一个过期代理 start 应把账号改投到哪里。
// 返回 (targetID, change)：
//   - change=false：不改动账号（mode=none，或链路成环/无解的兜底）
//   - change=true, targetID=nil：改投为直连
//   - change=true, targetID!=nil：改投到该备用代理 id
//
// byID 是「全部代理」的快照（id -> Proxy），now 为判定基准时间。
func ResolveProxyFallbackTarget(start Proxy, byID map[int64]Proxy, now time.Time) (*int64, bool) {
	switch start.FallbackMode {
	case FallbackModeDirect:
		return nil, true
	case FallbackModeProxy:
		visited := map[int64]struct{}{start.ID: {}}
		curID := start.BackupProxyID
		for {
			if curID == nil {
				return nil, false
			}
			if _, seen := visited[*curID]; seen {
				return nil, false
			}
			p, ok := byID[*curID]
			if !ok {
				return nil, false
			}
			if !(&p).IsExpired(now) && p.Status != StatusExpired {
				id := p.ID
				return &id, true
			}
			visited[*curID] = struct{}{}
			switch p.FallbackMode {
			case FallbackModeDirect:
				return nil, true
			case FallbackModeProxy:
				curID = p.BackupProxyID
			default:
				return nil, false
			}
		}
	default:
		return nil, false
	}
}

// ResolveProxyFailureFallbackTarget 计算故障代理 start 应把账号改投到哪里，返回值语义同
// ResolveProxyFallbackTarget。备用代理必须 active、未过期且未被判定故障；不满足时沿该备用代理
// 自身的故障回退配置继续查找，成环或无解则不改动账号。
func ResolveProxyFailureFallbackTarget(start Proxy, byID map[int64]Proxy, now time.Time) (*int64, bool) {
	visited := map[int64]struct{}{start.ID: {}}
	mode := start.FailureFallbackMode
	curID := start.FailureBackupProxyID
	for {
		switch mode {
		case FallbackModeDirect:
			return nil, true
		case FallbackModeProxy:
			if curID == nil {
				return nil, false
			}
			if _, seen := visited[*curID]; seen {
				return nil, false
			}
			p, ok := byID[*curID]
			if !ok {
				return nil, false
			}
			visited[p.ID] = struct{}{}
			if p.Status == StatusActive && !(&p).IsExpired(now) && !p.IsDegraded() {
				id := p.ID
				return &id, true
			}
			mode = p.FailureFallbackMode
			curID = p.FailureBackupProxyID
		default:
			return nil, false
		}
	}
}
