package worker

import (
	"context"
	"encoding/json"
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

func (jobManager *JobManager) watchJobs() (err error) {
	var (
		getResp            *clientv3.GetResponse
		kv                 *mvccpb.KeyValue
		job                *common.Job
		watchStartRevision int64
		watchChan clientv3.WatchChan
		watchResp clientv3.WatchResponse
		watchEvent *clientv3.Event
		jobName string
	)
	// get /cron/jobs/ 目录下所有任务, 获知当前集群的 revision
	if getResp, err = jobManager.kv.Get(context.TODO(), common.JobSaveDir, clientv3.WithPrefix()); err != nil {
		return
	}
	// 遍历当前任务
	for _, kv = range getResp.Kvs {
		if job, err = common.UnpackJob(kv.Value); err != nil {
			// TODO 把 job 同步给 scheduler(调度协程)
		}
	}
	// 从该 revision 向后监听变化事件
	go func() {
		// 监听协程
		watchStartRevision = getResp.Header.Revision + 1
		// 监听目录的后续变化
		watchChan = jobManager.watcher.Watch(context.TODO(), common.JobSaveDir, clientv3.WithRev(watchStartRevision))
		for watchResp = range watchChan {
			for watchEvent = range watchResp.Events {
				switch watchEvent.Type {
				case mvccpb.PUT:
					// 保存任务
					// TODO 反序列化 job 推送给 scheduler
					if job, err = common.UnpackJob(watchEvent.Kv.Value); err != nil {
						continue
					}
					// 构造一个更新 Event
				case mvccpb.DELETE:
					// 删除任务
					// TODO 推送删除事件给 scheduler
					jobName = common.ExtractJobName(string(watchEvent.Kv.Key))
					// 构造一个删除 Event
				}
			}
		}
	}()
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
	return
}

func (jobManager *JobManager) SaveJob(job *common.Job) (oldJob *common.Job, err error) {
	// 把任务保存到 `/cron/jobs/` 任务名 -> json
	var (
		jobKey      string
		jobValue    []byte
		putResponse *clientv3.PutResponse
		oldJobObj   common.Job
	)
	// etcd 路径
	jobKey = common.JobSaveDir + job.Name
	// 任务信息 JSON
	if jobValue, err = json.Marshal(job); err != nil {
		return
	}
	// 保存到 ETCD
	if putResponse, err = G_jobManager.kv.Put(context.TODO(), jobKey, string(jobValue), clientv3.WithPrevKV()); err != nil {
		return
	}
	// 如果是更新返回旧值
	if putResponse.PrevKv != nil {
		// 对旧值做反序列化
		if err = json.Unmarshal(putResponse.PrevKv.Value, &oldJobObj); err != nil {
			err = nil
			return
		}
		oldJob = &oldJobObj
	}
	return
}

func (jobManager *JobManager) DeleteJob(name string) (oldJob *common.Job, err error) {
	var (
		jobKey  string
		delResp *clientv3.DeleteResponse
		oldObj  common.Job
	)
	// etcd 中保存任务的 key
	jobKey = common.JobSaveDir + name
	// 从 etcd 中删除
	if delResp, err = jobManager.kv.Delete(context.TODO(), jobKey, clientv3.WithPrevKV()); err != nil {
		return
	}
	// 返回被删除的任务信息
	if len(delResp.PrevKvs) != 0 {
		if err = json.Unmarshal(delResp.PrevKvs[0].Value, &oldObj); err != nil {
			err = nil
			return
		}
		oldJob = &oldObj
	}
	return
}

func (jobManager *JobManager) ListJobs() (jobList []*common.Job, err error) {
	var (
		dirKey  string
		getResp *clientv3.GetResponse
		kv      *mvccpb.KeyValue
		job     *common.Job
	)
	dirKey = common.JobSaveDir
	// 获取指定前缀(目录)的所有任务信息
	if getResp, err = jobManager.kv.Get(context.TODO(), dirKey, clientv3.WithPrefix()); err != nil {
		return
	}
	// 初始化数组
	jobList = make([]*common.Job, 0)
	// 遍历所有任务进行反序列化
	for _, kv = range getResp.Kvs {
		job = &common.Job{}
		if err = json.Unmarshal(kv.Value, job); err != nil {
			err = nil
			continue
		}
		jobList = append(jobList, job)
	}
	return
}

// 杀死任务
func (jobManager *JobManager) KillJob(name string) (err error) {
	// 向 /cron/killer/ 目录更新任务名
	var (
		killerKey      string
		leaseGrantResp *clientv3.LeaseGrantResponse
		leaseId        clientv3.LeaseID
	)
	killerKey = common.JobKillDir + name
	// 让 worker 监听到一次 put 操作, 创建一个租约让其自动过期即可
	if leaseGrantResp, err = jobManager.lease.Grant(context.TODO(), 1); err != nil {
		return
	}
	leaseId = leaseGrantResp.ID
	// 设置 killer 标记
	if _, err = jobManager.kv.Put(context.TODO(), killerKey, "", clientv3.WithLease(leaseId)); err != nil {
		return
	}
	
	return
}
