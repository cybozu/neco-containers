package main

import (
	"fmt"
	"time"
)

type Machine struct {
	Hostname  string
	BmcIPadr  string
	NodeIPadr string
	SerialNo  string
}

type Machines struct {
	machine []Machine
}

// コレクターを起動
func collector(n int) {
	fmt.Printf("Collector %d  started\n", n)
	for {
		if len(queue) == 0 {
			time.Sleep(1 * time.Second)
		} else {
			v := queue[0]
			queue = queue[1:]
			time.Sleep(2 * time.Second)
			fmt.Println(n, v.SerialNo)
		}
	}
}

// ターゲットを取得定期に取得する
func targetReader() Machines {
	fmt.Println("Read target list and post queue")
	var m Machines
	for i := 1; i < 100; i++ {
		mx := Machine{
			Hostname:  fmt.Sprintf("test%d", i),
			BmcIPadr:  fmt.Sprintf("192.168.0.%d", i),
			NodeIPadr: fmt.Sprintf("172.16.0.%d", i),
			SerialNo:  fmt.Sprintf("ABCDE-%d", i),
		}
		m.machine = append(m.machine, mx)
	}
	return m
}

var queue []Machine

// メイン
func main() {

	//キューを作成
	queue = make([]Machine, 0)

	// パラメータ取得
	wn := 3

	// コレクターを起動
	for i := 0; i < wn; i++ {
		go collector(i)
	}

	// メインループ
	for {
		// ターゲット読込
		machineList := targetReader()

		// キューへ積む
		queue = machineList.machine

		// 待機
		fmt.Println("スリープ")
		time.Sleep(120 * time.Second)
	}
}
