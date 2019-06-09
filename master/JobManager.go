package master

import (
	"context"
	"encoding/json"
	"github.com/go-crontab/common"
	"go.etcd.io/etcd/clientv3"
	"time"
)

// 任务管理器
type JobManager struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
}

var (
	// 单例
	G_jobManager *JobManager
)

// 初始化管理器
func InitJobManager() (err error) {
	var (
		config clientv3.Config
		client *clientv3.Client
		kv     clientv3.KV
		lease  clientv3.Lease
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
	
	// 赋值单例
	G_jobManager = &JobManager{
		client: client,
		kv:     kv,
		lease:  lease,
	}
	return
}

func (jobManager *JobManager) SaveJob(job *common.Job) (oldJob *common.Job, err error) {
	// 把任务保存到 `/cron/jobs/` 任务名 -> json
	var (
		jobKey string
		jobValue []byte
		putResponse *clientv3.PutResponse
		oldJobObj common.Job
	)
	// etcd 路径
	jobKey = "/cron/jobs/" + job.Name
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