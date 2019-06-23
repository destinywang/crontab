package common

import (
	"encoding/json"
	"strings"
)

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

// 变化事件
type JobEvent struct {
	EventType int
	job *Job
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

func UnpackJob(value []byte) (ret *Job, err error) {
	var (
		job *Job
	)
	job = &Job{}
	if err = json.Unmarshal(value, job); err != nil {
		return
	}
	ret = job
	return
}

// 提取任务名
func ExtractJobName(jobKey string) (string) {
	return strings.TrimPrefix(jobKey, JobSaveDir)
}

// 任务变化事件: 更新和删除
func BuildJobEvent(eventType int, job *Job) (jobEvent *JobEvent) {
	return &JobEvent{
		EventType:eventType,
		job:job,
	}
}