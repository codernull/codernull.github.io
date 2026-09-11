package engine

import (
	"bytes"
	"fmt"
	"testing"

	"codernull.github.io/matching-engine/internal/order"
)

// BenchmarkEngineMatch 撮合全流程：100k 交替买单/卖单（买 1050、卖 1000，必成交）。
// 每轮 100k 单 → 50k 笔成交，输出写入 bytes.Buffer。
func BenchmarkEngineMatch(b *testing.B) {
	const n = 100000
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		eng := New(&buf)
		for j := 0; j < n; j++ {
			if j%2 == 0 {
				eng.Handle(order.Event{Kind: 'A', ID: fmt.Sprintf("B%d", j), Side: order.Buy, Qty: 1, Price: 1050})
			} else {
				eng.Handle(order.Event{Kind: 'A', ID: fmt.Sprintf("S%d", j), Side: order.Sell, Qty: 1, Price: 1000})
			}
		}
	}
}
