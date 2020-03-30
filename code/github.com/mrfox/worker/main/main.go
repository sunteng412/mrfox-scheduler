package main

import (
	"context"
	"fmt"
	_yaml "mrfox-cron-common/yaml"
	"mrfox-cron-worker/main/handler"
	"runtime"
	"time"
)

// 初始化线程数量
func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

var (
	//错误信息
	err error
	//取消续租
	cancelFun context.CancelFunc
)

func main() {
	//解析配置
	_yaml.ParseProperties()

	//配置goroutine数量
	initEnv()
	//启动日志处理协程
	if err = handler.InitLogSink();err != nil{
		goto ERR
	}

	//启动任务调度器
	if err = handler.InitScheduler();err != nil{
		goto ERR
	}

	//启动任务执行器
	if err = handler.InitExecutor();err != nil{
		goto ERR
	}

	//初始化etcd配置
	if err = handler.InitEtcdClient();err != nil{
		goto ERR
	}


	//上报机器
	if cancelFun,err = handler.RegisterMachine();err != nil{
		goto ERR
	}

	//监听任务变化
	if err = handler.SingEtcdClient.WatchJobs();err != nil{
		goto ERR
	}



	select {} // 阻塞

	//正常退出
	for  {
		time.Sleep(time.Second *1)
		cancelFun()
	}

	ERR:
		fmt.Println(err.Error())
}