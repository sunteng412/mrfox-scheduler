/**
 * @Author: mrfox
 * @Description:
 * @File:  JobLock
 * @Version: 1.0.0
 * @Date: 2020/3/28 11:59 下午
 */
package handler

import (
	"context"
	"go.etcd.io/etcd/clientv3"
	"log"
	"mrfox-cron-common/utils"
)

//分布式锁
type DistributedLock struct {
	//etcd连接
	etcdClient *EtcdClient
	//锁的资源名
	LockKey string
	//用于终止自动续租
	cancelFunc context.CancelFunc
	//租约id
	leaseId  clientv3.LeaseID
	//是否上锁成功
	isLocked bool
}

//定义方法
type Lock interface {

	//尝试获取锁-获取不到就返回false
	TryLock(leaseTime int)(getLock bool)

	//释放锁
	UnLock()

	//获取锁-获取不到最多等待waitSecond秒,抢不到返回false
	Lock(lockKey string,waitSecond int)(getLock bool)
}

//尝试上锁
func (distributedLock *DistributedLock)TryLock(leaseTime int64)(err error) {
	var (
		//租约响应
		leaseGrantResp *clientv3.LeaseGrantResponse
		//自动续租上下文
		cancelCtx context.Context
		//取消自动续租
		cancelFunc context.CancelFunc
		//租约id
		leaseId clientv3.LeaseID
		//返回leaseChan
		keepRespChan <- chan *clientv3.LeaseKeepAliveResponse
		//创建抢锁事务
		txn clientv3.Txn
		//事务响应
		txnResp *clientv3.TxnResponse
	)
	//1.创建租约
	if leaseGrantResp,err = distributedLock.etcdClient.lease.Grant(context.TODO(),leaseTime);err!=nil{
		return utils.ERR_CREATE_LEASE
	}

	//2.自动续租
	//2.1创建context用于取消自动续租
	cancelCtx,cancelFunc = context.WithCancel(context.TODO())
	//2.2拿到租约id
	leaseId = leaseGrantResp.ID
	//2.3自动续租
	if keepRespChan,err = distributedLock.etcdClient.lease.KeepAlive(cancelCtx,leaseId);err != nil{
		goto FAIL
	}

	//3.处理续租应答的协程
	go func() {
		var (
			//续租应答
			keepResp *clientv3.LeaseKeepAliveResponse
		)
		for  {
			select {
			//接收续租应答
			case keepResp = <- keepRespChan://自动续租应答
				if keepResp == nil{//说明自动续租被取消或者异常被cancel
					goto END
				}
				log.Printf("[distributedLock][%v]自动续租成功",distributedLock.LockKey)
			}
		}
	END:
	}()

	//3.创建事务
	txn = distributedLock.etcdClient.kv.Txn(context.TODO())

	//4.事务抢锁
	txn.If(clientv3.Compare(clientv3.CreateRevision(distributedLock.LockKey),"=",0)).
		Then(clientv3.OpPut(distributedLock.LockKey,utils.GetGoroutineId(),clientv3.WithLease(leaseId))).
		Else(clientv3.OpGet(distributedLock.LockKey))
	//4.1提交事务
	if txnResp,err = txn.Commit();err != nil{
		err = utils.ERR_EXEC_TXN
		goto FAIL
	}

	//5.成功返回,失败释放租约
	if !txnResp.Succeeded{//说明锁被占用
		err = utils.ERR_LOCK_ALREADY_REQUIRED
		goto FAIL
	}
	distributedLock.leaseId = leaseId
	distributedLock.cancelFunc = cancelFunc
	distributedLock.isLocked = true
	return

	FAIL:
		//失败取消自动续租
		cancelFunc()
		//释放租约
		distributedLock.etcdClient.lease.Revoke(context.TODO(),leaseId)
 	return
}


//释放锁
func (distributedLock *DistributedLock)Unlock(){
	if distributedLock.isLocked{
		//取消自动续租
		distributedLock.cancelFunc()
		distributedLock.etcdClient.lease.Revoke(context.TODO(),distributedLock.leaseId)
	}

}

//初始化etcd资源
func GetDistributedLock(lockKey string) (distributedLock *DistributedLock) {
	distributedLock =  &DistributedLock{
		etcdClient:SingEtcdClient,
		LockKey:lockKey,
	}
	return
}
