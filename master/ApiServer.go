package master

import (
	"encoding/json"
	"fmt"
	"github.com/go-crontab/common"
	"net"
	"net/http"
	"strconv"
	"time"
)

// 任务的 HTTP 接口
type ApiServer struct {
	httpServer *http.Server
}

var (
	// 单例对象
	G_apiServer *ApiServer
)

// 保存任务接口
// POST job = {"name": "job1", "command": "echo hello", "cronExpr": "* * * * *"}
func handleJobSave(resp http.ResponseWriter, req *http.Request) {
	fmt.Println("handleJobSave")
	// 任务保存到 ETCD 中
	// 1. 解析 HTTP 表单
	var (
		err error
		postForm string
		job common.Job
	)
	if err = req.ParseForm(); err != nil {
		fmt.Println("err: ", err)
	}
	// 2. 取表单的 job 字段
	postForm = req.PostForm.Get("job")
	// 3. 反序列化 job
	if err = json.Unmarshal([]byte(postForm), &job); err != nil {
	
	}
}

// 初始化服务
func InitApiServer() (err error) {
	var (
		mux        *http.ServeMux
		listener   net.Listener
		httpServer *http.Server
	)
	// 配置路由
	mux = http.NewServeMux()
	mux.HandleFunc("/job/save", handleJobSave)
	fmt.Println("G_config: ", G_config)
	// 启动 TCP 监听
	if listener, err = net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(G_config.ApiPort)); err != nil {
		return err
	}
	// 创建 HTTP 服务
	httpServer = &http.Server{
		ReadHeaderTimeout: time.Duration(G_config.ApiReadTimeout) * time.Millisecond,
		WriteTimeout:      time.Duration(G_config.ApiWriteTimeout) * time.Millisecond,
		Handler:           mux,
	}
	
	// 赋值单例
	G_apiServer = &ApiServer{
		httpServer: httpServer,
	}
	
	// 启动服务端
	go httpServer.Serve(listener)
	
	return nil
}
