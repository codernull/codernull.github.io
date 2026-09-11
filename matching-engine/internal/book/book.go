// Package book 实现价格/时间优先的订单簿：
//
//   - 价格优先：买盘按价格降序取最优（最高买价），卖盘按价格升序取最优（最低卖价）；
//   - 时间优先：同一价格档位内用 FIFO 队列，先挂单先成交；
//   - 撤单定位：order_id → (方向, 价格, 链表节点) 索引，O(1) 定位删除。
//
// 结构映射（对应原 C++ 复杂版设计）：
//
//	价格索引    → Skiplist（有序价格档位集合，买/卖各一棵）
//	时间队列    → container/list（每档一个 FIFO）
//	order_id 索引 → map[string]*slot（撤单 O(1) 定位）
//
// 注意：买盘与卖盘的价格档位必须分开存储（bidLvls / askLvls），
// 否则买单与卖单价格相同时会挤进同一个队列，破坏两侧独立的 FIFO 语义。
package book

import (
	"container/list"

	"codernull.github.io/matching-engine/internal/order"
)

// level 是一个价格档位：同价订单按到达顺序排队。
type level struct {
	orders *list.List
}

// slot 记录订单在盘口中的位置，供撤单定位。
type slot struct {
	side  order.Side
	price int64
	elem  *list.Element
}

// OrderBook 维护买盘、卖盘与 order_id 索引。
type OrderBook struct {
	bids    *Skiplist // 买盘价格档位集合（升序存储，最优买价取 Max）
	asks    *Skiplist // 卖盘价格档位集合（升序存储，最优卖价取 Min）
	bidLvls map[int64]*level
	askLvls map[int64]*level
	idx     map[string]*slot
}

// New 创建空订单簿。
func New() *OrderBook {
	return &OrderBook{
		bids:    NewSkiplist(),
		asks:    NewSkiplist(),
		bidLvls: make(map[int64]*level),
		askLvls: make(map[int64]*level),
		idx:     make(map[string]*slot),
	}
}

// levelMap 返回对应方向的档位表。
func (b *OrderBook) levelMap(side order.Side) map[int64]*level {
	if side == order.Sell {
		return b.askLvls
	}
	return b.bidLvls
}

// priceSet 返回对应方向的价格档位集合（Skiplist）。
func (b *OrderBook) priceSet(side order.Side) *Skiplist {
	if side == order.Sell {
		return b.asks
	}
	return b.bids
}

// Add 将订单放入对应盘口；同价订单追加到队尾（时间优先）。
func (b *OrderBook) Add(o *order.Order) {
	sl := b.priceSet(o.Side)
	lvls := b.levelMap(o.Side)
	lv, ok := lvls[o.Price]
	if !ok {
		lv = &level{orders: list.New()}
		lvls[o.Price] = lv
		sl.Insert(o.Price)
	}
	elem := lv.orders.PushBack(o)
	b.idx[o.ID] = &slot{side: o.Side, price: o.Price, elem: elem}
}

// Cancel 按订单号撤单；订单已成交或不存在时直接忽略。
func (b *OrderBook) Cancel(id string) {
	s, ok := b.idx[id]
	if !ok {
		return
	}
	delete(b.idx, id)
	lvls := b.levelMap(s.side)
	lv := lvls[s.price]
	lv.orders.Remove(s.elem)
	if lv.orders.Len() == 0 {
		delete(lvls, s.price)
		b.priceSet(s.side).Delete(s.price)
	}
}

// Contains 返回订单是否仍在盘口内（已成交/已撤单则为 false）。
func (b *OrderBook) Contains(id string) bool {
	_, ok := b.idx[id]
	return ok
}

// BestBid 返回最优买价（最高买价）。
func (b *OrderBook) BestBid() (int64, bool) { return b.bids.Max() }

// BestAsk 返回最优卖价（最低卖价）。
func (b *OrderBook) BestAsk() (int64, bool) { return b.asks.Min() }

// FrontBid 返回最优买价档的队首订单。
func (b *OrderBook) FrontBid() *order.Order {
	price, ok := b.bids.Max()
	if !ok {
		return nil
	}
	return b.bidLvls[price].orders.Front().Value.(*order.Order)
}

// FrontAsk 返回最优卖价档的队首订单。
func (b *OrderBook) FrontAsk() *order.Order {
	price, ok := b.asks.Min()
	if !ok {
		return nil
	}
	return b.askLvls[price].orders.Front().Value.(*order.Order)
}

// PopBid 弹出最优买价档的队首订单（已全部成交），并清理空档位与索引。
func (b *OrderBook) PopBid() *order.Order { return b.pop(order.Buy) }

// PopAsk 弹出最优卖价档的队首订单（已全部成交），并清理空档位与索引。
func (b *OrderBook) PopAsk() *order.Order { return b.pop(order.Sell) }

func (b *OrderBook) pop(side order.Side) *order.Order {
	sl := b.priceSet(side)
	lvls := b.levelMap(side)
	price, ok := sl.Min()
	if side == order.Buy {
		price, ok = sl.Max()
	}
	if !ok {
		return nil
	}
	lv := lvls[price]
	front := lv.orders.Front()
	o := front.Value.(*order.Order)
	lv.orders.Remove(front)
	delete(b.idx, o.ID)
	if lv.orders.Len() == 0 {
		delete(lvls, price)
		sl.Delete(price)
	}
	return o
}

// LevelSummary 返回指定方向某价格档位的数量序列（按到达顺序）。
func (b *OrderBook) LevelSummary(side order.Side, price int64) []int64 {
	lv, ok := b.levelMap(side)[price]
	if !ok {
		return nil
	}
	out := make([]int64, 0, lv.orders.Len())
	for e := lv.orders.Front(); e != nil; e = e.Next() {
		out = append(out, e.Value.(*order.Order).Qty)
	}
	return out
}

// AscendAsks 按价格升序遍历卖盘档位（低价在前，即最优卖价在前）。
func (b *OrderBook) AscendAsks(fn func(price int64) bool) { b.asks.Ascend(fn) }

// AscendBids 按价格升序遍历买盘档位（调用方可逆序得到高价在前）。
func (b *OrderBook) AscendBids(fn func(price int64) bool) { b.bids.Ascend(fn) }
