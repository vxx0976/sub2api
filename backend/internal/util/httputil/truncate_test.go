package httputil

import (
	"strings"
	"testing"
)

// truncateAtRuneBoundary 的契约：UTF-8 内容回退到 rune 边界（最多 3 字节）；
// 非 UTF-8（GBK/二进制）内容按原字节截断——不能把整个前缀剥成空串丢光日志内容。
func TestTruncateAtRuneBoundary(t *testing.T) {
	t.Run("ASCII 原样截断", func(t *testing.T) {
		got := truncateAtRuneBoundary("hello world", 5)
		if got != "hello" {
			t.Fatalf("got %q, want %q", got, "hello")
		}
	})

	t.Run("中文截在多字节中间回退到边界", func(t *testing.T) {
		s := "ab中文" // '中' 占 3 字节，从 index 2 开始
		got := truncateAtRuneBoundary(s, 4)
		if got != "ab" {
			t.Fatalf("got %q, want %q", got, "ab")
		}
	})

	t.Run("四字节 emoji 切剩 3 字节也能回退", func(t *testing.T) {
		s := "ab😀x" // 😀 占 4 字节，从 index 2 开始
		got := truncateAtRuneBoundary(s, 5)
		if got != "ab" {
			t.Fatalf("got %q, want %q", got, "ab")
		}
	})

	t.Run("纯非 UTF-8 内容按字节截断而非清空", func(t *testing.T) {
		// 模拟 GBK 中文/二进制：全部字节均为非法 UTF-8 尾字节
		s := strings.Repeat("\xd6\xd0", 100) // GBK "中" ×100
		got := truncateAtRuneBoundary(s, 20)
		if got == "" {
			t.Fatalf("non-UTF-8 prefix was stripped to empty — diagnostic content lost")
		}
		if len(got) != 20 {
			t.Fatalf("len = %d, want 20 (raw byte cut)", len(got))
		}
	})

	t.Run("TruncateBody 非 UTF-8 保留前缀", func(t *testing.T) {
		body := []byte(strings.Repeat("\xb4\xed\xce\xf3", 200)) // GBK "错误"
		got := TruncateBody(body, 64)
		if got == "...(truncated)" {
			t.Fatalf("body preview lost entirely: %q", got)
		}
		if !strings.HasSuffix(got, "...(truncated)") {
			t.Fatalf("missing truncation marker: %q", got)
		}
	})
}
