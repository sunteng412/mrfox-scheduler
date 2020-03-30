/**
 * @Author: mrfox
 * @Description:
 * @File:  blockLock_tset
 * @Version: 1.0.0
 * @Date: 2020/3/14 9:48 下午
 */
package distributedLock

import (
	"context"
	"github.com/prometheus/common/log"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"testing"
	"time"
)

/**
1.上锁(创建租约,自动续租,拿着租约去抢占一个key)
2.处理业务
3.释放锁(取消自动续租,释放租约)
*/
func Test_blockLock(t *testing.T) {
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

		//申请一个lease(租约)
		lease = clientv3.NewLease(client)

		//申请一个5秒的租约
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
						//每秒续租一次,所以每秒会受到一次应答
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
			//没有抢到锁陷入等待锁释放
			//处理kv变化时间
			for watchResp = range GetInstance().lockChan{
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

		//假装处理业务
		log.Info("----------")
		time.Sleep(time.Second * 5)


}
