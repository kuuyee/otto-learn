package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
)

// 设置中断信号的监听，这样做是因为需要为我们关心的子命令设置监听。
// 如果不这样做，那么我们执行的像Terraform或Vagrant命令如果收不到ctrl-C
// 命令就不会优雅的被清理
func initSignalHandlers() {
	fmt.Printf("[KuuYee_DEBUG]====> %s\n", "main.go/initSignalHandlers/14")
	signalCh := make(chan os.Signal, 2)
	signal.Notify(signalCh, os.Interrupt)
	fmt.Printf("[KuuYee_DEBUG]====> %s\n", "main.go/initSignalHandlers/17")
	go func() {
		for {
			<-signalCh
			log.Printf("[DEBUG] main: 接收到信号中断. ignoring since command should also listen")
		}
	}()
}
