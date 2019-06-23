package worker

import (
	"context"
	"fmt"
	"github.com/DestinyWang/go-crontab/common"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"time"
)

// 任务管理器
type JobManager struct {
	client  *clientv3.Client
	kv      clientv3.KV
	lease   clientv3.Lease
	watcher clientv3.Watcher
}

var (
	// 单例
	G_jobManager *JobManager
)

// 监听任务变化, 监听 ETCD
func (jobManager *JobManager) watchJobs() (err error) {
	var (
		getResp            *clientv3.GetResponse
		kv                 *mvccpb.KeyValue
		job                *common.Job
		watchStartRevision int64
		watchChan          clientv3.WatchChan
		watchResp          clientv3.WatchResponse
		watchEvent         *clientv3.Event
		jobName            string
		jobEvent           *common.JobEvent
	)
	// 1. get /cron/jobs/ 目录下所有任务, 获知当前集群的 revision
	if getResp, err = jobManager.kv.Get(context.TODO(), common.JobSaveDir, clientv3.WithPrefix()); err != nil {
		return
	}
	// 遍历当前任务
	for _, kv = range getResp.Kvs {
		if job, err = common.UnpackJob(kv.Value); err == nil {
			// TODO 把 job 同步给 scheduler(调度协程)
			jobEvent = common.BuildJobEvent(common.JOB_EVENT_SAVE, job)
			fmt.Printf("exists: %v", *jobEvent)
		} else {
			continue
		}
	}
	// 2. 从该 revision 向后监听变化事件
	go func() {
		// 监听协程
		watchStartRevision = getResp.Header.Revision + 1
		// 监听目录的后续变化
		watchChan = jobManager.watcher.Watch(context.TODO(), common.JobSaveDir, clientv3.WithRev(watchStartRevision), clientv3.WithPrefix())
		for watchResp = range watchChan {
			for _, watchEvent = range watchResp.Events {
				switch watchEvent.Type {
				case mvccpb.PUT:
					// 保存任务
					if job, err = common.UnpackJob(watchEvent.Kv.Value); err != nil {
						continue
					}
					// 构造一个更新 Event
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_SAVE, job)
				case mvccpb.DELETE:
					// 删除任务
					jobName = common.ExtractJobName(string(watchEvent.Kv.Key))
					// 构造一个删除 Event
					job = &common.Job{
						Name:jobName,
					}
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_DEL, job)
				}
				// 推给 scheduler
				// G_Scheduler.PushJobEvent(jobEvent)
				fmt.Printf("new: %v", *jobEvent)
			}
		}
	}()
	return
}

// 初始化管理器
func InitJobManager() (err error) {
	var (
		config  clientv3.Config
		client  *clientv3.Client
		kv      clientv3.KV
		lease   clientv3.Lease
		watcher clientv3.Watcher
	)
	// 初始化配置
	config = clientv3.Config{
		Endpoints:   G_config.EtcdEndpoints,
		DialTimeout: time.Duration(G_config.EtcdDialTimeout) * time.Millisecond,
	}
	// 建立连接
	if client, err = clientv3.New(config); err != nil {
		return err
	}
	// 得到 kv 和 Lease 的子集
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)
	watcher = clientv3.NewWatcher(client)
	
	// 赋值单例
	G_jobManager = &JobManager{
		client:  client,
		kv:      kv,
		lease:   lease,
		watcher: watcher,
	}
	
	// 启动任务监听
	G_jobManager.watchJobs()
	return
}
