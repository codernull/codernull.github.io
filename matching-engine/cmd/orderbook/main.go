// Command orderbook 读取行情数据文件，运行撮合引擎，打印成交与最终订单簿。
//
// 用法：
//
//	go run ./cmd/orderbook -file testdata/sample.txt
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"codernull.github.io/matching-engine/internal/engine"
	"codernull.github.io/matching-engine/internal/feed"
)

func main() {
	file := flag.String("file", "", "行情数据文件路径")
	flag.Parse()
	if *file == "" {
		log.Fatal("用法: orderbook -file <market-data-file>")
	}

	// IO 读取与撮合解耦：reader goroutine 逐行投递事件。
	events, errs := feed.ReadFile(*file)
	eng := engine.New(os.Stdout)

	for ev := range events {
		eng.Handle(ev)
	}
	select {
	case err := <-errs:
		if err != nil {
			log.Fatalf("读取失败: %v", err)
		}
	default:
	}

	fmt.Println("\n=== 最终订单簿 ===")
	eng.Dump(os.Stdout)
}
