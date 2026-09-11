// Package feed 提供行情数据文件的并发读取：读取 goroutine 逐行解析，
// 通过 channel 将事件投递给撮合引擎，实现 IO 读取与撮合逻辑的解耦。
package feed

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"codernull.github.io/matching-engine/internal/order"
)

// ReadFile 启动一个读取 goroutine，将文件逐行解析为 Event 投递到返回的 channel。
//
// 错误通过 errs channel 传递（缓冲 1，避免 goroutine 泄漏）；正常结束时 errs 无值。
// 调用方 range events 结束后，用非阻塞读 errs 判断是否有解析/读取错误。
func ReadFile(path string) (<-chan order.Event, <-chan error) {
	events := make(chan order.Event)
	errs := make(chan error, 1)
	go func() {
		defer close(events)
		f, err := os.Open(path)
		if err != nil {
			errs <- err
			return
		}
		defer f.Close()

		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" {
				continue
			}
			ev, err := ParseLine(line)
			if err != nil {
				errs <- err
				return
			}
			events <- ev
		}
		if err := sc.Err(); err != nil {
			errs <- err
		}
	}()
	return events, errs
}

// ParseLine 解析一行行情数据，格式：A|X,order_id,side,quantity,price。
//
// 例如 "A,16113575,B,18,585" 表示新增买单 18 股 @585；"X" 行携带同样的五段信息，
// 撤单时仅使用 order_id。
func ParseLine(line string) (order.Event, error) {
	parts := strings.Split(line, ",")
	if len(parts) != 5 {
		return order.Event{}, fmt.Errorf("feed: 非法行（应为 5 段）: %q", line)
	}
	kind := parts[0]
	if len(kind) != 1 || (kind[0] != 'A' && kind[0] != 'X') {
		return order.Event{}, fmt.Errorf("feed: 非法消息类型: %q", line)
	}
	side := order.Buy
	if parts[2] == "S" {
		side = order.Sell
	}
	qty, err := strconv.ParseInt(parts[3], 10, 64)
	if err != nil {
		return order.Event{}, fmt.Errorf("feed: 非法数量 %q: %w", line, err)
	}
	price, err := strconv.ParseInt(parts[4], 10, 64)
	if err != nil {
		return order.Event{}, fmt.Errorf("feed: 非法价格 %q: %w", line, err)
	}
	return order.Event{Kind: kind[0], ID: parts[1], Side: side, Qty: qty, Price: price}, nil
}
