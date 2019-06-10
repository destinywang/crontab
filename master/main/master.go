package main

import (
	"flag"
	"fmt"
	"github.com/go-crontab/master"
	"runtime"
	"time"
)

var (
	confPath string // 文件路径
)

// 初始化
func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

// 解析命令行参数
func initArgs() {
	// 支持 `master -config ./master.json` 方式传入参数
	flag.StringVar(&confPath, "config", "./master.json", "master 端系统配置文件路径")
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
	if err = master.InitConfig(confPath); err != nil {
		fmt.Println("err: ", err)
	}
	// 启动任务管理器
	if err = master.InitJobManager(); err != nil {
		fmt.Println("err: ", err)
	}
	// 启动 ApiHTTP 服务
	if err = master.InitApiServer(); err != nil {
		fmt.Println(err)
	}
	for {
		fmt.Println("ok")
		time.Sleep(1 * time.Minute)
	}
	return
}
