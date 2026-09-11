// Package order 定义撮合引擎的核心数据模型：订单、输入事件与成交记录。
package order

// Side 表示订单方向。
type Side byte

const (
	// Buy 买单（BID 侧）。
	Buy Side = 'B'
	// Sell 卖单（ASK 侧）。
	Sell Side = 'S'
)

// Order 是一笔挂单。
//
// Seq 为订单到达撮合引擎的顺序号：同价格档位内按 Seq 升序成交（时间优先）。
type Order struct {
	ID    string
	Side  Side
	Qty   int64
	Price int64
	Seq   uint64
}

// Event 是输入流中的一条消息：新增订单（A）或撤单（X）。
type Event struct {
	Kind  byte // 'A' 新增，'X' 撤单
	ID    string
	Side  Side
	Qty   int64
	Price int64
}

// Trade 是一笔成交记录，写入 doneList 供审计与撤单返还核对。
type Trade struct {
	ID     string // 成交编号，自增
	BuyID  string // 买单订单号
	SellID string // 卖单订单号
	Qty    int64  // 成交数量
	Price  int64  // 成交价（被动方价格）
	Seq    uint64 // 触发该成交的订单到达序号
}
