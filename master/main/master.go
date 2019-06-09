package main

import (
	"fmt"
	"github.com/go-crontab/master"
	"runtime"
)

func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	var (
		err error
	)
	// 初始化线程
	initEnv()
	// 启动 ApiHTTP 服务
	if err = master.InitApiServer(); err != nil {
		fmt.Println(err)
	}
	fmt.Println("ok")
	return
}
