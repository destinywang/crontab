package common


const (
	JobSaveDir = "/cron/jobs/"
	JobKillDir = "/cron/killer/"
	
	// 保存任务事件
	JOB_EVENT_SAVE = 1
	
	// 删除任务事件
	JOB_EVENT_DEL = 2
)