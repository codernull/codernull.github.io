package book

import (
	"strconv"
	"testing"

	"codernull.github.io/matching-engine/internal/order"
)

// BenchmarkSkiplistInsert 插入 10k 个不同价格档位（含跳表新建）。
func BenchmarkSkiplistInsert(b *testing.B) {
	const n = 10000
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sl := NewSkiplist()
		for j := 0; j < n; j++ {
			sl.Insert(int64(j))
		}
	}
}

// BenchmarkSkiplistDelete 插入 10k 个档位后全部删除。
func BenchmarkSkiplistDelete(b *testing.B) {
	const n = 10000
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sl := NewSkiplist()
		for j := 0; j < n; j++ {
			sl.Insert(int64(j))
		}
		for j := 0; j < n; j++ {
			sl.Delete(int64(j))
		}
	}
}

// BenchmarkOrderBookAddCancel 新增 10k 单后全部撤单（覆盖 Add/Cancel/索引/档位清理）。
func BenchmarkOrderBookAddCancel(b *testing.B) {
	const n = 10000
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ob := New()
		for j := 0; j < n; j++ {
			ob.Add(&order.Order{ID: strconv.Itoa(j), Side: order.Buy, Qty: 1, Price: int64(1000 + j%100)})
		}
		for j := 0; j < n; j++ {
			ob.Cancel(strconv.Itoa(j))
		}
	}
}
