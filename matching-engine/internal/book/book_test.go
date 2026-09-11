package book

import (
	"testing"

	"codernull.github.io/matching-engine/internal/order"
)

func TestSkiplistInsertMinMax(t *testing.T) {
	s := NewSkiplist()
	if _, ok := s.Min(); ok {
		t.Fatal("空表不应有 Min")
	}
	for _, k := range []int64{110, 90, 105, 100} {
		if !s.Insert(k) {
			t.Fatalf("Insert(%d) 应返回新 key", k)
		}
	}
	if s.Insert(100) {
		t.Fatal("重复 Insert(100) 应返回 false")
	}
	if v, ok := s.Min(); !ok || v != 90 {
		t.Fatalf("Min = %d, want 90", v)
	}
	if v, ok := s.Max(); !ok || v != 110 {
		t.Fatalf("Max = %d, want 110", v)
	}
}

func TestSkiplistDelete(t *testing.T) {
	s := NewSkiplist()
	for _, k := range []int64{110, 90, 105, 100} {
		s.Insert(k)
	}
	if !s.Delete(100) {
		t.Fatal("Delete(100) 应返回 true")
	}
	if s.Delete(100) {
		t.Fatal("重复 Delete(100) 应返回 false")
	}
	if v, ok := s.Min(); !ok || v != 90 {
		t.Fatalf("删除后 Min = %d, want 90", v)
	}
	if v, ok := s.Max(); !ok || v != 110 {
		t.Fatalf("删除后 Max = %d, want 110", v)
	}
	// 删到空
	for _, k := range []int64{110, 105, 90} {
		s.Delete(k)
	}
	if _, ok := s.Min(); ok {
		t.Fatal("删空后不应有 Min")
	}
}

func TestOrderBookAddCancel(t *testing.T) {
	b := New()
	b.Add(&order.Order{ID: "a", Side: order.Buy, Qty: 5, Price: 100})
	b.Add(&order.Order{ID: "b", Side: order.Buy, Qty: 5, Price: 100})
	if !b.Contains("a") || !b.Contains("b") {
		t.Fatal("新增后应都在盘口内")
	}
	if v, _ := b.BestBid(); v != 100 {
		t.Fatalf("BestBid = %d, want 100", v)
	}
	// 撤掉队首 a，队首应变为 b
	b.Cancel("a")
	if b.Contains("a") {
		t.Fatal("撤单后 a 不应在盘口内")
	}
	if f := b.FrontBid(); f == nil || f.ID != "b" {
		t.Fatalf("撤单后队首应为 b, got %v", f)
	}
	// 撤不存在的单应静默
	b.Cancel("nope")
	// 撤空档位后 BestBid 消失
	b.Cancel("b")
	if _, ok := b.BestBid(); ok {
		t.Fatal("撤空后不应有 BestBid")
	}
}

func TestOrderBookPopCleansLevel(t *testing.T) {
	b := New()
	b.Add(&order.Order{ID: "a", Side: order.Sell, Qty: 5, Price: 110})
	b.Add(&order.Order{ID: "b", Side: order.Sell, Qty: 7, Price: 110})
	if v, _ := b.BestAsk(); v != 110 {
		t.Fatalf("BestAsk = %d, want 110", v)
	}
	b.PopAsk()
	if !b.Contains("b") || b.Contains("a") {
		t.Fatal("PopAsk 后应只剩 b")
	}
	if f := b.FrontAsk(); f.ID != "b" {
		t.Fatalf("PopAsk 后队首应为 b")
	}
	b.PopAsk()
	if _, ok := b.BestAsk(); ok {
		t.Fatal("弹出全部后不应有 BestAsk")
	}
}
