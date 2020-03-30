/**
 * @Author: mrfox
 * @Description:
 * @File:  BlockLockChan
 * @Version: 1.0.0
 * @Date: 2020/3/14 9:50 下午
 */
//通用chan--用于分布式锁等待场景的等待通知操作
package distributedLock

import (
	"go.etcd.io/etcd/clientv3"
	"sync"
)


type BLockChanFather interface {
	close(c chan<- clientv3.WatchResponse)
}

//互斥锁
var mux sync.Mutex
type BLockChan struct {
	//通道包
	lockChan chan clientv3.WatchResponse
}

var instance *BLockChan

//获取实例
func  GetInstance() *BLockChan{
	if instance == nil{
		mux.Lock()
		if instance == nil{
			instance = &BLockChan{lockChan:make(chan clientv3.WatchResponse)}
		}
		mux.Unlock()
	}
	return instance
}


