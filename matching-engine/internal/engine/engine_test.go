package engine

import (
	"bytes"
	"strings"
	"testing"

	"codernull.github.io/matching-engine/internal/feed"
	"codernull.github.io/matching-engine/internal/order"
)

// runInput 逐行解析并送入引擎，返回 stdout（成交 + 最终盘口）与成交记录。
func runInput(t *testing.T, input string) (string, []order.Trade) {
	t.Helper()
	var buf bytes.Buffer
	eng := New(&buf)
	for _, line := range strings.Split(strings.TrimSpace(input), "\n") {
		ev, err := feed.ParseLine(line)
		if err != nil {
			t.Fatalf("解析失败 %q: %v", line, err)
		}
		eng.Handle(ev)
	}
	eng.Dump(&buf)
	return buf.String(), eng.Trades()
}

// 题目示例输入：含 3 条撤单（其中 100008 与 100005 已成交，撤单应被忽略）。
// 成交语义：每撮合一对订单记一笔成交，部分成交的剩余量继续与下一单撮合。
func TestSampleInputFromProblem(t *testing.T) {
	input := `A,100000,S,1,1075
A,100001,B,9,1000
A,100002,B,30,975
A,100003,S,10,1050
A,100004,B,10,950
A,100005,S,2,1025
A,100006,B,1,1000
X,100004,B,10,950
A,100007,S,5,1025
A,100008,B,3,1050
X,100008,B,3,1050
X,100005,S,2,1025`
	out, trades := runInput(t, input)

	// 预期两笔成交：买单 100008(3@1050) 吃 1025 档队首 100005(2) 得 2@1025，
	// 剩余 1 股再吃 100007(1) 得 1@1025。
	if len(trades) != 2 {
		t.Fatalf("成交笔数 = %d, want 2；输出:\n%s", len(trades), out)
	}
	if trades[0].BuyID != "100008" || trades[0].SellID != "100005" || trades[0].Qty != 2 || trades[0].Price != 1025 {
		t.Fatalf("首笔成交 = %+v, want BuyID=100008 SellID=100005 Qty=2 Price=1025", trades[0])
	}
	if trades[1].BuyID != "100008" || trades[1].SellID != "100007" || trades[1].Qty != 1 || trades[1].Price != 1025 {
		t.Fatalf("次笔成交 = %+v, want BuyID=100008 SellID=100007 Qty=1 Price=1025", trades[1])
	}

	// 最终盘口：ASK 1025:[4]（100007 剩 5-1） 1050:[10] 1075:[1]；BID 1000:[9 1] 975:[30]
	for _, want := range []string{"1025: [4]", "1050: [10]", "1075: [1]", "1000: [9 1]", "975: [30]"} {
		if !strings.Contains(out, want) {
			t.Fatalf("最终盘口缺少 %q；输出:\n%s", want, out)
		}
	}
}

// 主动买单吃卖盘：成交价取被动方（卖盘）价格。
func TestAggressiveBuyEatsAsk(t *testing.T) {
	input := `A,10,S,23,80
A,11,B,4,100
A,12,B,6,100
A,13,B,10,90
A,14,B,2,90
A,15,B,3,90`
	_, trades := runInput(t, input)
	if len(trades) != 5 {
		t.Fatalf("成交笔数 = %d, want 5", len(trades))
	}
	for _, tr := range trades {
		if tr.Price != 80 {
			t.Fatalf("成交价 = %d, want 80（被动卖价）: %+v", tr.Price, tr)
		}
	}
}

// 主动卖单吃买盘：成交价取被动方（买盘）价格；部分成交剩余留在队首。
func TestAggressiveSellEatsBid(t *testing.T) {
	input := `A,20,B,4,100
A,21,B,6,100
A,22,S,23,80`
	out, trades := runInput(t, input)
	if len(trades) != 2 {
		t.Fatalf("成交笔数 = %d, want 2；输出:\n%s", len(trades), out)
	}
	if trades[0].Price != 100 || trades[0].Qty != 4 {
		t.Fatalf("首笔成交 = %+v, want 4@100", trades[0])
	}
	if trades[1].Price != 100 || trades[1].Qty != 6 {
		t.Fatalf("次笔成交 = %+v, want 6@100", trades[1])
	}
	// 卖单 23 - 4 - 6 = 13 剩余成为被动单，最终盘口 ASK 80:[13]
	if !strings.Contains(out, "80: [13]") {
		t.Fatalf("最终盘口缺少 80:[13]；输出:\n%s", out)
	}
}

// 撤单后重新挂单进队尾：时间优先按新到达顺序重新计算。
func TestCancelThenReAddGoesToBack(t *testing.T) {
	input := `A,1,B,5,100
A,2,B,5,100
X,1,B,5,100
A,3,B,5,100
A,4,S,100,100`
	out, trades := runInput(t, input)
	// 撮合顺序应为：2（最早在册）→ 3（撤单后重挂进队尾）→ 1 已撤不参与
	if len(trades) != 2 {
		t.Fatalf("成交笔数 = %d, want 2；输出:\n%s", len(trades), out)
	}
	if trades[0].BuyID != "2" || trades[1].BuyID != "3" {
		t.Fatalf("成交买单顺序 = [%s %s], want [2 3]", trades[0].BuyID, trades[1].BuyID)
	}
}

// 同价档位时间优先：先挂单先成交。
func TestTimePrioritySamePrice(t *testing.T) {
	input := `A,1,B,5,100
A,2,B,5,100
A,3,S,100,100`
	out, trades := runInput(t, input)
	if len(trades) != 2 {
		t.Fatalf("成交笔数 = %d, want 2；输出:\n%s", len(trades), out)
	}
	if trades[0].BuyID != "1" || trades[1].BuyID != "2" {
		t.Fatalf("成交买单顺序 = [%s %s], want [1 2]", trades[0].BuyID, trades[1].BuyID)
	}
}
