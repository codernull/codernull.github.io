package engine

import (
	"bytes"
	"strings"
	"testing"

	"codernull.github.io/matching-engine/internal/feed"
)

// TestReadFileFlow 验证 feed.ReadFile 并发读取路径与 ParseLine 路径结果一致。
func TestReadFileFlow(t *testing.T) {
	var buf bytes.Buffer
	eng := New(&buf)
	events, errs := feed.ReadFile("../../testdata/sample.txt")
	for ev := range events {
		eng.Handle(ev)
	}
	select {
	case err := <-errs:
		if err != nil {
			t.Fatalf("read: %v", err)
		}
	default:
	}
	eng.Dump(&buf)

	trades := eng.Trades()
	if len(trades) != 2 {
		t.Fatalf("成交笔数 = %d, want 2；OUT:\n%s", len(trades), buf.String())
	}
	for _, want := range []string{
		"TRADE 100008 2@1025 100005",
		"TRADE 100008 1@1025 100007",
		"1025: [4]",
		"1050: [10]",
		"1075: [1]",
		"1000: [9 1]",
		"975: [30]",
	} {
		if !strings.Contains(buf.String(), want) {
			t.Fatalf("输出缺少 %q；OUT:\n%s", want, buf.String())
		}
	}
}
