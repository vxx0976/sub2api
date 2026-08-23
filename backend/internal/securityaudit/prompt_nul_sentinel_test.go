package securityaudit

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

// 客户端文本里的 NUL 必须在归一化时被剥掉。两个内部哨兵都是 NUL 包裹的，
// 没有 NUL 就伪造不出任何一个，这一条同时堵住下面两个测试覆盖的攻击面。
func TestPromptSegmentsStripClientNUL(t *testing.T) {
	if got := stripNUL("a\x00b\x00c"); got != "abc" {
		t.Errorf("stripNUL = %q, want %q", got, "abc")
	}
	if got := stripNUL("no nul here"); got != "no nul here" {
		t.Errorf("无 NUL 时应原样返回，got %q", got)
	}
	segs := normalizedPromptSegments([]promptSegment{
		{text: "hello" + promptAuditPrioritySeparator + "world", user: true, role: "user"},
		{text: "x" + promptAuditArchiveSeparator + "y", user: false, role: "assistant"},
	})
	for _, s := range segs {
		if strings.ContainsRune(s.text, 0) {
			t.Errorf("归一化后仍含 NUL: %q", s.text)
		}
		if strings.Contains(s.text, promptAuditPrioritySeparator) ||
			strings.Contains(s.text, promptAuditArchiveSeparator) {
			t.Errorf("归一化后仍含哨兵: %q", s.text)
		}
	}
}

// 放大式 DoS：客户端在正文里塞 N 个优先级分隔符，扫描器按它切块、每块一次 guard 调用，
// 一次请求就能放大成 N 次 guard 调用。guard 本就是产能瓶颈，这是现成的打死手段。
func TestClientCannotAmplifyGuardChunkCount(t *testing.T) {
	const forged = 2000
	body, err := json.Marshal(map[string]any{
		"model": "gpt-5.4",
		"messages": []map[string]any{
			{"role": "user", "content": strings.Repeat("hi"+promptAuditPrioritySeparator, forged)},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	snap, err := ExtractPromptSnapshot(Request{
		Body: body, Protocol: "openai_chat_completions", RequestID: "amp", Stage: "http",
	}, false)
	if err != nil {
		t.Fatalf("提取失败: %v", err)
	}
	// 扫描器按优先级分隔符切块；正文里伪造的那些必须已经不存在。
	chunks := strings.Count(snap.ScanText, promptAuditPrioritySeparator) + 1
	if chunks > 2 {
		t.Errorf("客户端把 guard 调用放大到 %d 块（伪造了 %d 个分隔符）", chunks, forged)
	}
}

// 二次方 CPU 炸弹：把哨兵拆成前缀 A 与后缀 B，正文写 A^k + B^k，
// 原实现「循环 ReplaceAll 直到稳定」每趟只消掉最内层一个，逼出 k 趟全串扫描。
// 实测过 1MB 请求体单函数 17.9 秒，且开关关着也会跑。
func TestArchiveSeparatorStripIsLinear(t *testing.T) {
	sep := promptAuditArchiveSeparator
	a, b := sep[:len(sep)/2], sep[len(sep)/2:]
	const k = 200_000
	bomb := strings.Repeat(a, k) + strings.Repeat(b, k)

	done := make(chan string, 1)
	go func() { done <- stripArchiveSeparator(bomb) }()
	select {
	case out := <-done:
		// 线性单趟：内层拼出的那些哨兵不保证全部消掉，这没关系——
		// stripNUL 已在上游让它们根本进不来，这里只是纵深防御。
		if len(out) > len(bomb) {
			t.Errorf("输出反而变长了: %d > %d", len(out), len(bomb))
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("stripArchiveSeparator 在 %d 层嵌套哨兵下超过 2 秒——二次方行为回归了", k)
	}
}

// 端到端:从真实入站 JSON 触发，确认整条提取链路对嵌套哨兵是线性的。
func TestExtractSnapshotLinearAgainstNestedSentinels(t *testing.T) {
	sep := promptAuditArchiveSeparator
	a, b := sep[:len(sep)/2], sep[len(sep)/2:]
	const k = 100_000
	body, err := json.Marshal(map[string]any{
		"model": "gpt-5.4",
		"messages": []map[string]any{
			{"role": "user", "content": strings.Repeat(a, k) + strings.Repeat(b, k)},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("入站 body %d KB", len(body)/1024)

	done := make(chan error, 1)
	go func() {
		snap, err := ExtractPromptSnapshot(Request{
			Body: body, Protocol: "openai_chat_completions", RequestID: "bomb", Stage: "http",
		}, true)
		if err == nil && strings.ContainsRune(snap.PayloadText(), 0) {
			err = fmt.Errorf("payload 仍含 NUL")
		}
		done <- err
	}()
	select {
	case err := <-done:
		if err != nil && err != ErrNoPromptText {
			t.Fatalf("提取失败: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("整条提取链路在嵌套哨兵下超过 5 秒——远程可达的 CPU 炸弹")
	}
}
