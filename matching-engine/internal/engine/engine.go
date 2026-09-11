// Package engine 实现撮合引擎：消费事件流，维护订单簿，
// 按价格/时间优先执行撮合，全部成交写入 doneList，并把成交与最终盘口输出到 io.Writer。
package engine

import (
	"fmt"
	"io"

	"codernull.github.io/matching-engine/internal/book"
	"codernull.github.io/matching-engine/internal/order"
)

// Engine 是撮合引擎。所有状态都在 Engine 内，便于后续并发改造
// （如将 Handle 收口到单一 goroutine，IO 读取与撮合彻底分离）。
type Engine struct {
	book    *book.OrderBook
	seq     uint64 // 订单到达顺序号，用于同价档位时间优先
	done    []order.Trade
	tradeNo uint64
	out     io.Writer
}

// New 创建撮合引擎，成交与最终盘口写入 out。
func New(out io.Writer) *Engine {
	return &Engine{book: book.New(), out: out}
}

// Handle 处理一条输入消息：'A' 新增并尝试撮合，'X' 撤单。
func (e *Engine) Handle(ev order.Event) {
	switch ev.Kind {
	case 'A':
		e.seq++
		e.book.Add(&order.Order{ID: ev.ID, Side: ev.Side, Qty: ev.Qty, Price: ev.Price, Seq: e.seq})
		e.match(ev.Side)
	case 'X':
		e.book.Cancel(ev.ID)
	}
}

// match 撮合循环：最优买价 ≥ 最优卖价即可成交，直至无法撮合。
//
// active 是本次触发撮合的主动方向；成交价取被动方（先挂单方）价格：
//   - 主动买单吃卖盘 → 成交价 = 卖盘价（ask）
//   - 主动卖单吃买盘 → 成交价 = 买盘价（bid）
//
// 部分成交：撮合数量 = min(双方剩余量)，未成交完的订单留在队首继续参与下一轮。
func (e *Engine) match(active order.Side) {
	for {
		bid, okB := e.book.BestBid()
		ask, okA := e.book.BestAsk()
		if !okB || !okA || bid < ask {
			return
		}
		b := e.book.FrontBid()
		a := e.book.FrontAsk()

		qty := b.Qty
		if a.Qty < qty {
			qty = a.Qty
		}
		price := ask
		if active == order.Sell {
			price = bid
		}

		e.tradeNo++
		e.done = append(e.done, order.Trade{
			ID: fmt.Sprintf("T%d", e.tradeNo), BuyID: b.ID, SellID: a.ID,
			Qty: qty, Price: price, Seq: e.seq,
		})
		fmt.Fprintf(e.out, "TRADE %s %d@%d %s\n", b.ID, qty, price, a.ID)

		b.Qty -= qty
		a.Qty -= qty
		if b.Qty == 0 {
			e.book.PopBid()
		}
		if a.Qty == 0 {
			e.book.PopAsk()
		}
	}
}

// Trades 返回全部成交记录（doneList 快照），供审计与撤单返还核对。
func (e *Engine) Trades() []order.Trade { return e.done }

// Dump 输出订单簿最终状态：卖盘按价格升序、买盘按价格降序。
func (e *Engine) Dump(w io.Writer) {
	fmt.Fprintln(w, "=================")
	fmt.Fprintln(w, "ASK")
	e.book.AscendAsks(func(price int64) bool {
		fmt.Fprintf(w, "%d: %v\n", price, e.book.LevelSummary(order.Sell, price))
		return true
	})
	fmt.Fprintln(w, "------------")
	fmt.Fprintln(w, "BID")
	var bids []int64
	e.book.AscendBids(func(price int64) bool {
		bids = append(bids, price)
		return true
	})
	for i := len(bids) - 1; i >= 0; i-- {
		fmt.Fprintf(w, "%d: %v\n", bids[i], e.book.LevelSummary(order.Buy, bids[i]))
	}
	fmt.Fprintln(w, "=================")
}
