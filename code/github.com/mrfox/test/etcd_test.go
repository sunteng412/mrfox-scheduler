/**
 * @Author: mrfox
 * @Description:
 * @File:  etcd_test
 * @Version: 1.0.0
 * @Date: 2020/3/10 10:56 下午
 */
package test

import (
	"context"
	"fmt"
	"github.com/prometheus/common/log"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"testing"
	"time"
)

func Test_etcd(t *testing.T) {
	var (
		config clientv3.Config
		client *clientv3.Client
		err    error
	)

	//配置客户端
	config = clientv3.Config{
		Endpoints:   []string{"10.211.55.7:2379"},
		DialTimeout: time.Second * 5,
	}

	//建立连接
	if client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
		return
	}

	//用于读写ETCD的键值对
	kv := clientv3.NewKV(client)
	if put, err := kv.Put(context.TODO(), "/cron/jobs/job1", "hello"); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Revision", put.Header.Revision)
	}

	//重新赋值并拿到以前的
	if put, err := kv.Put(context.TODO(), "/cron/jobs/job1", "1", clientv3.WithPrevKV()); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Revision", put.Header.Revision)
		if put.PrevKv != nil {
			fmt.Println("PrevValue", string(put.PrevKv.Value))
		}
	}
}

//测试etcd读取
func Test_etcdRead(t *testing.T) {
	var (
		config clientv3.Config
		client *clientv3.Client
		err    error
	)
	config = clientv3.Config{
		Endpoints:        []string{"10.211.55.7:2379"},
		AutoSyncInterval: 0,
		//超时时间
		DialTimeout:          time.Second * 5,
		DialKeepAliveTime:    0,
		DialKeepAliveTimeout: 0,
		MaxCallSendMsgSize:   0,
		MaxCallRecvMsgSize:   0,
		TLS:                  nil,
		Username:             "",
		Password:             "",
		RejectOldCluster:     false,
		DialOptions:          nil,
		Context:              nil,
		LogConfig:            nil,
		PermitWithoutStream:  false,
	}

	//连接
	if client, err = clientv3.New(config); err != nil {
		log.Infof("[etcd]连接失败,%v", err)
		return
	} else {
		log.Infof("[etcd]连接成功,%v", config)
	}

	log.Info("------------------查询单个-----------------------------")
	//获取单个数据
	kv := clientv3.NewKV(client)

	kv.Put(context.TODO(),"/cron/jobs/job2","2")

	if get, err := kv.Get(context.TODO(), "/cron/jobs/job1"); err != nil {
		log.Errorf("[etcd] get is err : %v", err)
	} else {
		log.Infof("[etcd] get is : %v", get.Kvs)
		for i := 0; i < len(get.Kvs); i++ {
			log.Infof("[i : %v,value : %v]\n", i, get.Kvs[i])
		}

	}

	log.Info("------------------查询前缀-----------------------------")

	err = nil
	var prefixKvRes *clientv3.GetResponse
	//获取已xxx为前缀的数据
	if prefixKvRes,err = kv.Get(context.TODO(), "/cron/jobs",clientv3.WithPrefix());err !=nil{
		log.Errorf("[etcd]get kv is error,err:%v",err)
	}else {
		for i:= 0; i< len(prefixKvRes.Kvs) ; i++ {
			log.Infof("[etcd]index is:%d,value is [%v]",i,prefixKvRes.Kvs[i])
		}
	}

	log.Info("------------------删除-----------------------------")
	var delRes *clientv3.DeleteResponse
	err = nil
	if delRes,err = kv.Delete(context.TODO(),"/cron/jobs/job1");err != nil{
		log.Errorf("[etcd]del kv is error,err:%v",err)
	}
	log.Infof("[etcd] before del,the value is:%v",delRes.Deleted)

	log.Info("------------------删除之后返回之前的kv-----------------------------")
	var (
		idx int
		kvpair *mvccpb.KeyValue
	)
	var batchDelRes *clientv3.DeleteResponse
	err = nil
	if batchDelRes,err = kv.Delete(context.TODO(),"/cron/jobs/",clientv3.WithPrevKV());err != nil{
		log.Errorf("[etcd]batch del kv is error,err:%v",err)
	}

	for idx,kvpair = range batchDelRes.PrevKvs {
		log.Infof("etcd] batch del kv,index:%v,value:%v",idx,kvpair)
	}

}

func Test_lease(t *testing.T) {
	var (
		config clientv3.Config
		client *clientv3.Client
		err    error
		//租约
		lease clientv3.Lease
		//申请租约
		leaseGrantRes *clientv3.LeaseGrantResponse
		//租约id
		leaseId  clientv3.LeaseID
		//kv对象
		kv clientv3.KV
		//put结果
		putRes *clientv3.PutResponse
		//续租请求响应
		keepResp *clientv3.LeaseKeepAliveResponse
		//只读chan
		keepRespChan <- chan *clientv3.LeaseKeepAliveResponse
	)

	//配置客户端
	config = clientv3.Config{
		Endpoints:   []string{"10.211.55.7:2379"},
		DialTimeout: time.Second * 5,
	}

	//建立连接
	if client, err = clientv3.New(config); err != nil {
		log.Infof("[etcd lease]创建 client成功....")
		return
	}

	//申请一个lease(租约)
	lease = clientv3.NewLease(client)

	//申请一个10秒的租约
	if leaseGrantRes,err = lease.Grant(context.TODO(),10);err != nil{
		log.Errorf("[etcd  lease] lease is error,cause:%v",err)
		return
	}

	//获取租约Id
	 leaseId = leaseGrantRes.ID

	 //定义一个5秒后停止的context
	timeout, _ := context.WithTimeout(context.TODO(), time.Second*5)

	//自动续租,5秒后停止续租
	if keepRespChan,err  = lease.KeepAlive(timeout,leaseId);err != nil{
		log.Infof("[etcd keepalive]keepalive is err,err:%v",err)
	}

	//自动续租
	//if keepRespChan,err  = lease.KeepAlive(context.TODO(),leaseId);err != nil{
	//	log.Infof("[etcd keepalive]keepalive is err,err:%v",err)
	//}



	//启动一个协程去通道中读取心跳应答请求
	go func() {
		select {
		case keepResp = <- keepRespChan:
			if keepRespChan == nil{
				//续租失败
				log.Infof("[etcd keepalive] lease is failure")
				goto END
			}else {//大概每秒会续租一次,所以就会受到一次应答
				//续租成功
				log.Infof("[etcd keepalive]lease is success,keepResp:%v",keepResp)
			}
		}
		END:
	}()

	 //获取kv对象 API子集
	kv = clientv3.NewKV(client)

	//创建一个kv,并与租约关联,实现10秒过期
	if putRes,err = kv.Put(context.TODO(),"/cron/lock/job1","",clientv3.WithLease(leaseId));
	err !=nil{
		log.Errorf("[etcd lease]put is error,cause:%v",err)
		return
	}else {
		log.Infof("[etcd lease] put is success,putRes:%v",putRes)
	}

	var getResp *clientv3.GetResponse
	//定义for循环查看key是否过期
	for  {
		if getResp,err = kv.Get(context.TODO(),"/cron/lock/job1");err != nil{
			log.Errorf("[etcd lease]get is error, err:%v",err)
			return
		}

		//过期就跳出循环
		if getResp.Count == 0{
			log.Info("[etcd lease]lease is expired....")
			break
		}else {
			log.Infof("[etcd lease]lease not expired....getResp:%v",getResp)
		}
		time.Sleep(time.Second)
	}
}

//测试watch
func Test_watch(t *testing.T) {
	var (
		config clientv3.Config
		client *clientv3.Client
		err    error
		//kv对象
		kv clientv3.KV
		//get响应
		getResp *clientv3.GetResponse
		//监听的事务id
		watchStartRevision int64
		//监听对象
		watcher clientv3.Watcher
		//监听chan
		watchRespChan <- chan clientv3.WatchResponse
		//watch结果
		watchResp clientv3.WatchResponse
		//事件
		event *clientv3.Event

	)

	//配置客户端
	config = clientv3.Config{
		Endpoints:   []string{"10.211.55.7:2379"},
		DialTimeout: time.Second * 5,
	}

	//建立连接
	if client, err = clientv3.New(config); err != nil {
		log.Infof("[etcd lease]创建 client成功....")
		return
	}
	kv = clientv3.NewKV(client)

	//模拟etcd中KV的变化
	go func() {
		for  {
			kv.Put(context.TODO(),"/cron/jobs/job7","窝窝头")

			kv.Delete(context.TODO(),"/cron/jobs/job7")

			time.Sleep(time.Second *1)
		}
	}()

	//先拿到当前的值,从当前年代开始监听
	 if getResp,err = kv.Get(context.TODO(),"/cron/jobs/job7");err != nil{
	 	log.Errorf("[etcd watch]get is error,cause:%v",err)
		 return
	 }

	 //如果当前值是存在的
	 if len(getResp.Kvs) >0 {
	 	log.Infof("[etcd watch]kv is exist,value:%v",getResp.Kvs[0].Value)
	 }

	 //当前etcd集群事务id,从当前事务id开始监听变化
	watchStartRevision  = getResp.Header.Revision + 1

	//创建一个监听器
	watcher = clientv3.NewWatcher(client)
	log.Infof("[etcd watch]from [txId:%v]watch",watchStartRevision)

	ctx, cancelFunc := context.WithCancel(context.TODO())
	time.AfterFunc(time.Second * 5, func() {
		cancelFunc()
	})

	//创建监听--5秒后取消
	watchRespChan = watcher.Watch(ctx,"/cron/jobs/job7",clientv3.WithRev(watchStartRevision))

	//创建监听
	//watchRespChan = watcher.Watch(context.TODO(),"/cron/jobs/job7",clientv3.WithRev(watchStartRevision))

	//处理kv变化时间
	for watchResp = range watchRespChan{
		for _,event =range watchResp.Events {
			switch event.Type {
			case mvccpb.PUT:
				log.Infof("[etcd watch] put -> %v,Revision:%v,final Revision:%v",
					string(event.Kv.Value),event.Kv.CreateRevision,event.Kv.ModRevision)
			case mvccpb.DELETE:
				log.Infof("[etcd watch] del,Revision:%v",event.Kv.ModRevision)
			}
		}
	}
}



//使用op操作代替get/put/delete
func Test_op(t *testing.T) {
	var (
		config clientv3.Config
		client *clientv3.Client
		err    error
		kv clientv3.KV
		//op对象-put
		putOp clientv3.Op
		//op对象-get
		getOp clientv3.Op
		//操作结果
		opResp clientv3.OpResponse
		delOp  clientv3.Op
	)

	//配置客户端
	config = clientv3.Config{
		Endpoints:   []string{"10.211.55.7:2379"},
		DialTimeout: time.Second * 5,
	}

	//建立连接
	if client, err = clientv3.New(config); err != nil {
		log.Infof("[etcd op]创建 client成功....")
		return
	}

	//获取kv对象
	kv = clientv3.NewKV(client)
	//创建op:operation-put
	delOp = clientv3.OpDelete("/cron/jobs/job8")
	kv.Do(context.TODO(),delOp)

	//创建op:operation-put
	putOp = clientv3.OpPut("/cron/jobs/job8","")
	//执行op
	if opResp, err = kv.Do(context.TODO(), putOp);err != nil{
		log.Errorf("[etcd op] do() is error,error:%v",opResp)
		return
	}
	//写入Revision
	log.Infof("[etcd op]write Revision:%v",opResp.Put().Header.Revision)



	//创建op:operation-get
	 getOp= clientv3.OpGet("/cron/jobs/job8")

	 //执行op
	 if opResp,err = kv.Do(context.TODO(),getOp);err != nil{
		log.Errorf("[etcd op]get is error,getOp:%v",getOp)
		 return
	 }
	 //打印
	 log.Infof("[etcd op]get Revision:[%v],Value:[%v]",opResp.Get().Kvs[0].ModRevision,string(opResp.Get().Kvs[0].Value))
}

//使用etcd创建分布式锁
/**
lease实现锁自动过期
op操作
txn事务:if else then

1.上锁(创建租约,自动续租,拿着租约去抢占一个key)
2.处理业务
3.释放锁(取消自动续租,释放租约)
*/
func Test_lock(t *testing.T) {
	var (
		config clientv3.Config
		client *clientv3.Client
		err    error
		//租约
		lease clientv3.Lease
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
		//取消函数
		cancelFunc context.CancelFunc
		//kv
		kv clientv3.KV
		//txn事务操作
		txn clientv3.Txn
		//txn结果
		txnResp *clientv3.TxnResponse
	)

	//配置客户端
	config = clientv3.Config{
		Endpoints:   []string{"10.211.55.7:2379"},
		DialTimeout: time.Second * 5,
	}

	//建立连接
	if client, err = clientv3.New(config); err != nil {
		log.Infof("[etcd lease]创建 client成功....")
		return
	}

	//申请一个lease(租约)
	lease = clientv3.NewLease(client)

	//申请一个10秒的租约
	if leaseGrantRes,err = lease.Grant(context.TODO(),10);err != nil{
		log.Errorf("[etcd  lease] lease is error,cause:%v",err)
		return
	}

	//拿到租约ID
	leaseId = leaseGrantRes.ID

	//准备一个自动续租的context,带有取消函数
	ctx, cancelFunc = context.WithCancel(context.TODO())

	//确保函数最后退出调用,自动续租停止
	defer cancelFunc()
	//立即释放锁
	defer lease.Revoke(context.TODO(),leaseId)

	if keepRespChan,err =lease.KeepAlive(ctx,leaseId);err != nil{
		log.Errorf("[etcd lease]lease is error,error:%v",err)
		return
	}

	//处理续租应答的协程
	go func() {
		for{
			select {
			case keepResp = <- keepRespChan:
				if keepRespChan == nil{
					//租约失效
					log.Errorf("[etcd lease] lease is expired")
					goto END
				}else {
					//每3秒续租一次,每次会受到一次应答
					log.Infof("[etcd lease]lease is success,keepResp.ID:%v",keepResp.ID)
				}
			}
		}
		END:
	}()


	kv = clientv3.NewKV(client)

	//创建事务
	txn = kv.Txn(context.TODO())

	//定义事务 if不存在key,then设置,else抢锁失败--事务操作
	//等于0说明key不存在
	txn.If(clientv3.Compare(clientv3.CreateRevision("/cron/lock/job9"),"=",0)).
		//设置
		Then(clientv3.OpPut("/cron/lock/job9","节点id:1",clientv3.WithLease(leaseId))).
		//否则抢锁失败
		Else(clientv3.OpGet("/cron/lock/job9"))

	//提交业务
	if txnResp,err = txn.Commit();err != nil{
		log.Errorf("[etcd lock]lock is error,error:%v",err)
		return
	}

	//判断是否抢到了锁--没有抢到值
	if !txnResp.Succeeded{
		log.Infof("[etcd lock]lock be occupied,node id:%v.",string(txnResp.Responses[0].GetResponseRange().Kvs[0].Value))
		return
	}

	//假装处理业务
	log.Info("----------")
	time.Sleep(time.Second * 5)

}