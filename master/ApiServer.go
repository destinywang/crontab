package master

import (
	"encoding/json"
	"fmt"
	"github.com/DestinyWang/go-crontab/common"
	"log"
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
		err      error
		postForm string
		job      common.Job
		oldJob   *common.Job
		bytes    []byte
	)
	if err = req.ParseForm(); err != nil {
		//resp.Write(common.BuildErrResp(err))
		fmt.Println("err: ", err)
	}
	// 2. 取表单的 job 字段
	postForm = req.PostForm.Get("job")
	// 3. 反序列化 job
	if err = json.Unmarshal([]byte(postForm), &job); err != nil {
		fmt.Println("err: ", err)
	}
	// 4. 保存到 etcd
	if oldJob, err = G_jobManager.SaveJob(&job); err != nil {
		fmt.Println("err: ", err)
	}
	// 5. 返回正常应答
	if bytes, err = common.BuildResp(0, "success", oldJob); err == nil {
		resp.Write(bytes)
	}
	return
}

// 删除任务接口
// POST /job/delete name=job1
func handleJobDelete(resp http.ResponseWriter, req *http.Request) {
	var (
		err    error
		name   string
		oldJob *common.Job
		bytes  []byte
	)
	// POST: name=job1&code=1
	if err = req.ParseForm(); err != nil {
		log.Fatal("err: ", err)
	}
	// 删除的任务名
	name = req.PostForm.Get("name")
	// 删除任务
	if oldJob, err = G_jobManager.DeleteJob(name); err != nil {
		log.Fatal("err: ", err)
	}
	// 正常应答
	if bytes, err = common.BuildResp(0, "success", oldJob); err == nil {
		resp.Write(bytes)
	}
}

// 查询任务列表
func handleJobList(resp http.ResponseWriter, req *http.Request) {
	fmt.Println("handleJobList")
	var (
		jobList []*common.Job
		err     error
		bytes   []byte
	)
	if jobList, err = G_jobManager.ListJobs(); err != nil {
		log.Fatal("err: ", err)
	}
	// 返回正常应答
	if bytes, err = common.BuildResp(0, "success", jobList); err == nil {
		resp.Write(bytes)
	}
}

// 强制杀死某个任务
func handleJobKill(resp http.ResponseWriter, req *http.Request) {
	var (
		err   error
		name  string
		bytes [] byte
	)
	if err = req.ParseForm(); err != nil {
		log.Fatal("err: ", err)
	}
	name = req.PostForm.Get("name")
	if err = G_jobManager.KillJob(name); err != nil {
		log.Fatal("err: ", err)
	}
	// 返回正常应答
	if bytes, err = common.BuildResp(0, "success", nil); err == nil {
		resp.Write(bytes)
	}
}

// 初始化服务
func InitApiServer() (err error) {
	var (
		mux           *http.ServeMux
		listener      net.Listener
		httpServer    *http.Server
		staticDir     http.Dir          // 静态文件根目录
		staticHandler http.Handler      // 静态文件的 HTTP 回调
	)
	// 配置路由
	mux = http.NewServeMux()
	mux.HandleFunc("/job/save", handleJobSave)
	mux.HandleFunc("/job/delete", handleJobDelete)
	mux.HandleFunc("/job/list", handleJobList)
	mux.HandleFunc("/job/kill", handleJobKill)
	//fmt.Println("G_config: ", G_config)
	log.Print("G_config: ", G_config)
	// 静态文件目录
	staticDir = http.Dir(G_config.WebRoot)
	staticHandler = http.FileServer(staticDir)
	// /index.html -> index.html -> ./webroot/index.html
	mux.Handle("/", http.StripPrefix("/", staticHandler))
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
