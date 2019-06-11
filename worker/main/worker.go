package main

import (
	"flag"
	"fmt"
	"github.com/DestinyWang/go-crontab/worker"
	"runtime"
)

var (
	confPath string
)

// 初始化
func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

// 解析命令行参数
func initArgs() {
	// 支持 `master -config ./master.json` 方式传入参数
	flag.StringVar(&confPath, "config", "./worker.json", "worker 端系统配置文件路径")
	flag.Parse()
}

func main() {
	var (
		err error
	)
	// 初始化命令行参数
	initArgs()
	// 初始化线程
	initEnv()
	// 加载配置
	if err = worker.InitConfig(confPath); err != nil {
		fmt.Println("err: ", err)
	}
	// 任务管理器
	if err = worker.InitJobManager(); err != nil {
		fmt.Println("err: ", err)
	}
	
	return
}
