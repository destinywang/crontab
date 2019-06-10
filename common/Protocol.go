package common

import "encoding/json"

// 定时任务 Job
type Job struct {
	Name     string `json:"name"`     // 任务名
	Command  string `json:"command"`  // shell 命令
	CronExpr string `json:"cronExpr"` // cron 表达式
}

type Response struct {
	Errno int `json:"errno"`
	Msg string `json:"msg"`
	Data interface{} `json:"data"`
}

func BuildResp(errno int, msg string, data interface{}) (resp []byte, err error) {
	// 1. 定义一个 response
	var (
		response Response
	)
	response.Errno = errno
	response.Msg = msg
	response.Data = data
	// 2. 系列化 JSON
	if resp, err = json.Marshal(response); err != nil {
	
	}
	return
}

func BuildErrResp(newErr error) (resp []byte, err error) {
	var (
		response Response
	)
	response.Errno = -1
	response.Msg = newErr.Error()
	response.Data = nil
	// 2. 系列化 JSON
	if resp, err = json.Marshal(response); err != nil {
	
	}
	return
}