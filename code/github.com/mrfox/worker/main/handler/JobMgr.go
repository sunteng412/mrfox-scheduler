/**
 * @Author: mrfox
 * @Description:
 * @File:  etcdClient
 * @Version: 1.0.0
 * @Date: 2020/3/21 4:58 下午
 */
package handler

import (
	"context"
	"encoding/json"
	"log"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"

	"mrfox-cron-common/utils"
	_yaml "mrfox-cron-common/yaml"
	_const "mrfox-cron-facade/const"
	"mrfox-cron-facade/entity"
	"strings"
	"time"
)

//任务管理
type EtcdClient struct {
	//客户端
	client *clientv3.Client

	//kv对象
	kv clientv3.KV

	//租约
	lease clientv3.Lease

	watcher clientv3.Watcher
}

var (
	//单例
	SingEtcdClient *EtcdClient
	//机器ip
	WorkerIp string
)


//注册到etcd上-上报机器
func RegisterMachine()(cancelFunc context.CancelFunc,err error) {
	var (
		//申请租约
		leaseGrantRes *clientv3.LeaseGrantResponse
		//租约id
		leaseId  clientv3.LeaseID
		//租约通道
		keepRespChan <- chan *clientv3.LeaseKeepAliveResponse
		//续租响应
		keepResp *clientv3.LeaseKeepAliveResponse
		//上下文
		ctx context.Context
		//put结果
		putResp *clientv3.PutResponse
		//worker信息
		workerInfo *entity.WorkerInfo
		//转换的json
		byteJson []byte
	)
	//上报机器
	if WorkerIp, err = utils.GetLocalIP();err != nil{
		return nil,err
	}

	//设置路径
	var machineKey = _const.MACHINE_PREFIX + "/" + WorkerIp

	//申请租约
	//申请一个租约
	if leaseGrantRes,err = SingEtcdClient.lease.Grant(context.TODO(),_yaml.SingYmlConfig.Worker.HeartbeatTime * 3);err != nil{
		log.Printf("[etcd lease]租约创建失败")
	}

	//拿到租约ID
	leaseId = leaseGrantRes.ID

	//准备一个自动续租的context,带有取消函数
	ctx, cancelFunc = context.WithCancel(context.TODO())

	if keepRespChan,err = SingEtcdClient.lease.KeepAlive(ctx,leaseId);err != nil{
		log.Printf("[etcd lease]创建续租失败:%v",err)
		return
	}


	//创建worker信息
	workerInfo = &entity.WorkerInfo{
		Group:         _yaml.SingYmlConfig.Worker.Group,
		IsContainer:   _yaml.SingYmlConfig.Worker.IsContainer,
		HeartbeatTime: _yaml.SingYmlConfig.Worker.HeartbeatTime,
		Ip:            WorkerIp,
	}

	byteJson, err = json.Marshal(workerInfo)

	if putResp, err = SingEtcdClient.kv.Put(context.TODO(), machineKey, string(byteJson), clientv3.WithLease(leaseId));
	err!=nil{
		log.Printf("[etcd put]上报失败,e:%v\n",err.Error())
	}else {
		log.Printf("[etcd put]上报成功,ip:%v,revision:%v", WorkerIp,putResp.Header.Revision)
	}


	//处理续租应答的协程
	go func() {
		for{
			select {
			case keepResp = <- keepRespChan:
				if keepRespChan == nil{
					//租约失效
					log.Printf("[etcd lease] 租约已过期")
					goto END
				}else {
					//每3秒续租一次,每次会受到一次应答
					log.Printf("[etcd lease]上送心跳成功,keepResp.ID:%v",keepResp.ID)
				}
			}
		}
	END:
	}()


	return cancelFunc,err
}

//监听etcd任务节点的变化
func (etcdClient *EtcdClient) WatchJobs()(err error)  {
	var (
		getResp *clientv3.GetResponse
		job *entity.Job
		kvpair *mvccpb.KeyValue
		watchStartRevision int64
		watchChan clientv3.WatchChan
		watchResp clientv3.WatchResponse
		watchEvent *clientv3.Event
		//任务事件
		jobEvent *Event
	)

	//get一下/cron/jobs/目录下所有任务,并且获知当前集群的revision
	if getResp,err = etcdClient.kv.Get(context.TODO(),_const.JOB_PREFIX,clientv3.WithPrefix());err != nil{
		return err
	}

	//遍历当前有哪些任务
	for _,kvpair = range getResp.Kvs {
		//反序列化
		if job,err = entity.UnpackJob(kvpair.Value);err != nil{
			log.Printf("[etcd]解析任务错误,%v",kvpair.Key)
		}else {
			jobEvent = BuildJobEvent(SaveJobStatus,job)
			log.Printf("[etcd]启动时获取任务列表,jobEvent:%v",jobEvent)
			//把这个job同步给scheduler
			SingScheduler.PushEvent(jobEvent)
		}
	}

	//从该revision中向后监听变化事件
	go func() {
		//从GET时刻的后续版本开始监听变化
		watchStartRevision = getResp.Header.Revision +1

		//启动监听/cron/jobs的后续变化
		watchChan = SingEtcdClient.watcher.Watch(context.TODO(),_const.JOB_PREFIX,
			clientv3.WithRev(watchStartRevision),clientv3.WithPrevKV(),clientv3.WithPrefix())

		//处理监听事件
		for watchResp = range watchChan{

			//遍历事件
			for _,watchEvent = range watchResp.Events {
				//事件类型
				var eventType int
				//反序列化Job,推送给调度协程
				switch watchEvent.Type {
					case mvccpb.PUT://保存/更新事件
						 if job,err = entity.UnpackJob(watchEvent.Kv.Value);err != nil{
						 	continue
						 }
						jobEvent = BuildBaseJobEvent(job)

						 //构造一个event事件
						 if watchEvent.PrevKv == nil{
						 	//保存
						 	eventType = SaveJobStatus
						 }else {
						 	//判断是强杀还是更新
						 	if jobEvent.Job.Operator == KillJobStatus{
						 		eventType = KillJobStatus
							}else {
								eventType = UpdateJobStatus
							}
						 }

						 jobEvent.EventType = eventType

						log.Printf("[etcd]eventType:%v,jobEvent:%v",eventType,jobEvent)
						 //推送事件
						 SingScheduler.PushEvent(jobEvent)
					case mvccpb.DELETE://强杀事件
						//提取到某个job的名称
						jobName := strings.TrimPrefix(string(watchEvent.Kv.Key),_const.JOB_PREFIX)

						//删除
						eventType = DeleteJobStatus
						//构造一个删除Event推送给调度协程
						job = &entity.Job{Name:jobName}

						jobEvent = BuildJobEvent(eventType,job)
						log.Printf("[etcd]eventType:%v,jobEvent:%v",eventType,jobEvent)
						//推送事件
						SingScheduler.PushEvent(jobEvent)
				}
				//推送调度协程

			}
		}
	}()
	return
}

//初始化管理器
func InitEtcdClient() (err error) {
	var (
		//etcd的连接配置
		config clientv3.Config
		//连接
		client *clientv3.Client
		//kv
		kv clientv3.KV
		//租约
		lease clientv3.Lease

		//watcher
		watcher clientv3.Watcher
	)

	//初始化配置
	config = clientv3.Config{
		Endpoints:   strings.Split(_yaml.SingYmlConfig.Etcd.Endpoints, ","),
		DialTimeout: time.Second * time.Duration(_yaml.SingYmlConfig.Etcd.DialTimeout),
	}

	//建立连接
	if client, err = clientv3.New(config); err != nil {
		return err
	}

	//得到KV和Lease的子集
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)
	watcher = clientv3.NewWatcher(client)


	//赋值单例
	SingEtcdClient = &EtcdClient{
		client: client,
		kv:     kv,
		lease:  lease,
		watcher:watcher,
	}
	log.Printf("[cron]etcd连接成功,address:%v", _yaml.SingYmlConfig.Etcd.Endpoints)
	return
}

