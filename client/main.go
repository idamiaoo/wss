package main

import (
	"fmt"
	"time"
)

func main() {
	cli, err := Init(true, "cmkj.lunarhalo.cn:8787")
	if err != nil {
		fmt.Println(err)
		return
	}
	go cli.Start()
	go func() {
		i := 0
		for {
			i++
			cli.SendMsg([]byte(fmt.Sprintf("你好%d", i)))
			<-time.After(1 * time.Second)
		}
	}()
	recv := cli.ReceiveChan()
	for {
		select {
		case msg := <-recv:
			fmt.Println(string(msg))
		}
	}
}
