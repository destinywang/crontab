package master

import (
	"context"
	"encoding/json"
	"github.com/go-crontab/common"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
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
	jobKey = common.JOB_SAVE_DIR + job.Name
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
		jobKey string
		delResp *clientv3.DeleteResponse
		oldObj common.Job
	)
	// etcd 中保存任务的 key
	jobKey = common.JOB_SAVE_DIR + name
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
		dirKey string
		getResp *clientv3.GetResponse
		kv *mvccpb.KeyValue
		job *common.Job
	)
	dirKey = common.JOB_SAVE_DIR
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