package main

import (
	"fmt"
	"time"
)

// コレクターを起動
func collector(n int) {
	fmt.Printf("Log Collector no-%d: Started\n", n)
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
	fmt.Println("コンフィグマップから取得する部分に相当")
	fmt.Println("Read iDRAC server list")
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

	// パラメータ取得
	// コレクターの起動数取得
	wn := 3

	// メインとコレクター（ワーカー）の間を繋ぐキューを作成
	queue = make([]Machine, 0)

	// コレクター（ワーカー）を起動
	for i := 0; i < wn; i++ {
		go collector(i)
	}

	// メインループ
	for {

		// CSVを読んで、構造体へセットする
		list, err := MachineListReader("testdata/bmc-list.csv")
		fmt.Println("list=", list, "err=", err)

		// ターゲット読込(テスト用）
		machineList := targetReader()

		// キューへ積む
		queue = machineList.machine

		// 待機
		fmt.Println("スリープ")
		time.Sleep(120 * time.Second)
	}
}
