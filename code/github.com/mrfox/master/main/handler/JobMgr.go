/**
 * @Author: mrfox
 * @Description:
 * @File:  JobMgr
 * @Version: 1.0.0
 * @Date: 2020/3/21 4:58 下午
 */
package handler

import (
	"context"
	"encoding/json"
	"github.com/labstack/gommon/log"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	_yaml "mrfox-cron-common/yaml"
	_const "mrfox-cron-facade/const"
	"mrfox-cron-facade/entity"
	"mrfox-cron-master/main/common"
	"strings"
	"time"
)

//任务管理
type JobMgr struct {
	//客户端
	client *clientv3.Client

	//kv对象
	kv clientv3.KV

	//租约
	lease clientv3.Lease
}

var (
	//单例
	SingJobMgr *JobMgr
)

//初始化管理器
func InitJobMgr() string {
	var (
		//etcd的连接配置
		config clientv3.Config
		//连接
		client *clientv3.Client
		//kv
		kv clientv3.KV
		//租约
		lease clientv3.Lease
		//错误
		err error
	)

	//初始化配置
	config = clientv3.Config{
		Endpoints:   strings.Split(_yaml.SingYmlConfig.Etcd.Endpoints, ","),
		DialTimeout: time.Second * time.Duration(_yaml.SingYmlConfig.Etcd.DialTimeout),
	}

	//建立连接
	if client, err = clientv3.New(config); err != nil {
		return err.Error()
	}

	//得到KV和Lease的子集
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)

	//赋值单例
	SingJobMgr = &JobMgr{
		client: client,
		kv:     kv,
		lease:  lease,
	}
	log.Infof("[cron]etcd连接成功,address:%v", _yaml.SingYmlConfig.Etcd.Endpoints)
	return ""
}


//保存job
func (jobMgr *JobMgr) SaveJob(job *entity.Job) (oldJob *entity.Job, errStr string) {
	//把任务保存到/mrfox/cron/jobs/任务名 -> json
	var (
		err      error
		jobKey   string
		jobValue []byte
		//保存结果
		putResp *clientv3.PutResponse
	)
	//任务信息json
	jobKey = _const.JOB_PREFIX + job.Name
	if jobValue, err = json.Marshal(job); err != nil {
		return nil, common.SerializationJobKeyErr
	}

	//保存到etcd--如果是更新则返回旧值
	if putResp, err = jobMgr.kv.Put(context.TODO(), jobKey, string(jobValue), clientv3.WithPrevKV()); err != nil {
		return nil, common.SaveJobKeyErr
	}

	//如果是更新,则返回旧值
	if putResp.PrevKv != nil {
		//对旧值做一个反序列化
		_ = json.Unmarshal(putResp.PrevKv.Value, &oldJob)
	}
	return oldJob, ""
}

//根据名称删除job
func (jobMgr *JobMgr) DeleteJob(name string) (oldJob *entity.Job, errStr string) {
	var (
		jobKey  string
		delResp *clientv3.DeleteResponse
		err     error
	)

	//得到任务名
	jobKey = _const.JOB_PREFIX + name

	//删除
	if delResp, err = jobMgr.kv.
		Delete(context.TODO(), jobKey, clientv3.WithPrevKV()); err != nil {
		return nil, err.Error()
	}

	//返回被删除的任务信息
	if len(delResp.PrevKvs) != 0 {
		if err = json.Unmarshal(delResp.PrevKvs[0].Value, &oldJob); err != nil {
			return nil, ""
		}
	}
	return oldJob, ""
}

//查询所有的job
func (jobMgr *JobMgr) ListJob() (jobList []*entity.Job, errStr string) {
	var (
		dirKey  string
		getResp *clientv3.GetResponse
		err     error
		//kv
		kvPair *mvccpb.KeyValue
		//临时job指针
		tempJob *entity.Job
	)

	dirKey = _const.JOB_PREFIX
	//获取该目录下所有kv
	if getResp, err = jobMgr.kv.Get(context.TODO(), dirKey, clientv3.WithPrefix()); err != nil {
		return nil, err.Error()
	}

	//初始化数组空间
	jobList = make([]*entity.Job, len(getResp.Kvs))

	//遍历,进行反序列化
	for _, kvPair = range getResp.Kvs {
		//初始化
		tempJob = &entity.Job{}

		//反序列化
		if err = json.Unmarshal(kvPair.Value, &tempJob); err != nil {
			return nil, err.Error()
		}
		jobList = append(jobList, tempJob)
	}


	return jobList, ""
}

//强杀某个任务--不过期job
func (jobMgr *JobMgr) KllJob(name string) (errStr string) {
	//更新一下任务名
	var (
		killerKey string

		//错误
		err error

		//请求job信息
		getResp *clientv3.GetResponse
		//job
		job  = &entity.Job{}
		//反序列化值
		jobValue []byte
		//putResponse *clientv3.PutResponse
	)

	//通知worker杀死对应任务
	killerKey = _const.JOB_PREFIX + name

	//获取该目录下所有kv
	if getResp, err = jobMgr.kv.Get(context.TODO(), killerKey); err != nil {
		return err.Error()
	}

	if err = json.Unmarshal(getResp.Kvs[0].Value, job); err != nil {
		return err.Error()
	}

	//设置操作
	job.Operator = entity.KILLED_OPERATOR

	//反序列化
	if jobValue, err = json.Marshal(job); err != nil {
		return err.Error()
	}

	//设置killer标记
	if _, err = jobMgr.kv.
		Put(context.TODO(), killerKey, string(jobValue)); err != nil {
		return err.Error()
	}
	return ""
}



//查询worker-alive列表
func (jobMgr *JobMgr) ListWorker() ([]*entity.WorkerInfo, string) {
	var (
		dirKey  string
		getResp *clientv3.GetResponse
		err     error
		//kv
		kvPair *mvccpb.KeyValue
		//临时job指针
		tempInfo *entity.WorkerInfo
		//列表
		workerList []*entity.WorkerInfo
	)

	dirKey = _const.MACHINE_PREFIX
	//获取该目录下所有kv
	if getResp, err = jobMgr.kv.Get(context.TODO(), dirKey, clientv3.WithPrefix()); err != nil {
		return nil, err.Error()
	}

	//初始化数组空间
	workerList = make([]*entity.WorkerInfo, len(getResp.Kvs))

	//遍历,进行反序列化
	for _, kvPair = range getResp.Kvs {
		//初始化
		tempInfo = &entity.WorkerInfo{}

		//反序列化
		if err = json.Unmarshal(kvPair.Value, &tempInfo); err != nil {
			return nil, err.Error()
		}
		workerList = append(workerList, tempInfo)
	}

	return workerList,""
}

//强杀某个任务--过期job
//func (jobMgr *JobMgr) KllJob(name string) (errStr string) {
//	//更新一下任务名
//	var (
//		killerKey string
//		//grant返回值
//		leaseGrantResp *clientv3.LeaseGrantResponse
//		//租约ID
//		leaseId clientv3.LeaseID
//
//		//错误
//		err error
//
//		//请求job信息
//		getResp *clientv3.GetResponse
//		//job
//		job *entity.Job = &entity.Job{}
//		//反序列化值
//		jobValue []byte
//		//putResponse *clientv3.PutResponse
//	)
//
//	//通知worker杀死对应任务
//	killerKey = JOB_PREFIX + name
//
//	//获取该目录下所有kv
//	if getResp, err = jobMgr.kv.Get(context.TODO(), killerKey); err != nil {
//		return err.Error()
//	}
//
//	if err = json.Unmarshal(getResp.Kvs[0].Value, job); err != nil {
//		return err.Error()
//	}
//
//	//让worker监听到一次put操作即可,创建一个租约让其自动过期即可
//	if leaseGrantResp,err = jobMgr.lease.Grant(context.TODO(),1);err != nil{
//		return err.Error()
//	}
//
//
//	//设置操作
//	job.Operator = entity.KILLED_OPERATOR
//
//	//反序列化
//	if jobValue, err = json.Marshal(job); err != nil {
//		return err.Error()
//	}
//	//返回一个租约ID
//	leaseId = leaseGrantResp.ID
//
//	//设置killer标记-并删除
//	if _, err = jobMgr.kv.
//		Put(context.TODO(), killerKey, string(jobValue),clientv3.WithLease(leaseId));err != nil{
//		return err.Error()
//	}
//
//	return ""
//}
