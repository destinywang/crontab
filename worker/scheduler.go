package worker

import "github.com/DestinyWang/go-crontab/common"

type Scheduler struct {
	jobEventChan chan *common.JobEvent             // etcd 任务事件队列
	jobPlanTable map[string]*common.JobScedulePlan // 任务调度计划表
}

var (
	G_scheduler *Scheduler
)

// 处理任务事件
func (scheduler *Scheduler) handleJobEvent(jobEvent *common.JobEvent) {
	var (
		jobSchedulePlan *common.JobScedulePlan
		err             error
		jobExisted bool
	)
	switch jobEvent.EventType {
	case common.JOB_EVENT_SAVE: // 保存任务事件
		if jobSchedulePlan, err = common.BuildJobSchedulePlan(jobEvent.Job); err != nil {
			return
		}
		scheduler.jobPlanTable[jobEvent.Job.Name] = jobSchedulePlan
	case common.JOB_EVENT_DEL: // 删除任务事件
		if jobSchedulePlan, jobExisted = scheduler.jobPlanTable[jobEvent.Job.Name]; jobExisted {
			delete(scheduler.jobPlanTable, jobEvent.Job.Name)
		}
	}
}

// 调度协程
func (scheduler *Scheduler) scheduleLoop() {
	var (
		jobEvent *common.JobEvent
	)
	// 定时任务
	for {
		select {
		case jobEvent = <-scheduler.jobEventChan:
			// 监听任务变化事件
			// 对内存中维护的任务列表做增删改查
			scheduler.handleJobEvent(jobEvent)
		}
	}
}

// 接受推送
func (scheduler *Scheduler) PushJobEvent(jobEvent *common.JobEvent) {

}

// 初始化调度器
func InitScheduler() {
	G_scheduler = &Scheduler{
		jobEventChan: make(chan *common.JobEvent, 1000),
		jobPlanTable:make(map[string]*common.JobScedulePlan),
	}
	go G_scheduler.scheduleLoop()
	return
}
