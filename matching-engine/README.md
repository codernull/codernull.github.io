# matching-engine

Go 实现的价格/时间优先订单簿撮合引擎。

源于一次交易系统笔试题（order book 实现）的完整重写：把笔试时的单文件 C++ 草稿，
重构成可测试、IO 与撮合解耦的工程化 Go 项目。

## 功能

- 读取行情数据文件（`A` 新增 / `X` 撤单），构建订单簿
- 撮合规则：**价格优先 → 同价时间优先（FIFO）**
- 支持部分成交（partial fill），未成交数量留在队首继续撮合
- **成交价 = 被动方（先挂单方）价格**：主动买单按卖价成交，主动卖单按买价成交
- 撤单按 `order_id` O(1) 定位；撤单后重新挂单进队尾
- 全部成交写入 doneList（审计 / 撤单返还核对）
- 输出成交明细与最终订单簿状态到 stdout

## 架构

```text
matching-engine/
├── cmd/orderbook/          # CLI 入口
├── internal/
│   ├── order/              # 数据模型：Order / Event / Trade
│   ├── book/               # 订单簿：Skiplist 价格档位 + 每档 FIFO + order_id 索引
│   ├── engine/             # 撮合引擎：事件处理、撮合循环、doneList
│   └── feed/               # 行情文件并发读取（reader goroutine → channel）
└── testdata/sample.txt     # 题目示例输入
```

设计要点（与原 C++ 复杂版设计的对应）：

| 原设计 | 本实现 |
|---|---|
| 价格为 key 的有序结构 | `Skiplist`（升序存储，买盘取 Max / 卖盘取 Min） |
| queue 时间队列 | `container/list` 每档 FIFO |
| `unordered_map<id, 位置>` 撤单索引 | `map[string]*slot`（方向 + 价格 + 链表节点） |
| 多线程 IO 读取分离 | `feed.ReadFile` goroutine + channel 投递 |
| doneList 已成交记录 | `Engine.done []Trade` |
| 可测试接口 | `book` / `engine` 独立包 + 单元测试 |

## 运行

```bash
# 编译
go build ./...

# 测试
go test ./...

# 跑题目示例
go run ./cmd/orderbook -file testdata/sample.txt
```

## 示例输出

```text
TRADE 100008 2@1025 100005
TRADE 100008 1@1025 100007

=== 最终订单簿 ===
=================
ASK
1025: [4]
1050: [10]
1075: [1]
------------
BID
1000: [9 1]
975: [30]
=================
```

## License

MIT
