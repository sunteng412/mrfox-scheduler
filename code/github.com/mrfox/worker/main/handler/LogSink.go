/**
 * @Author: mrfox
 * @Description:
 * @File:  LogSink
 * @Version: 1.0.0
 * @Date: 2020/3/29 8:05 下午
 */
package handler

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"mrfox-cron-common/yaml"
	"mrfox-cron-facade/entity"
	"time"
)

//mongo存储日志
type LogSink struct {
	client *mongo.Client
	//集合
	logCollection *mongo.Collection
	//log通道
	logChan chan *entity.JobLog
	//自动提交chan
	autoCommitChan chan *entity.LogBatch
}

var (
	//单例
	SingLogSink *LogSink
)

//批量写入日志
func (logSink *LogSink) saveLogs(batch *entity.LogBatch) {
	logSink.logCollection.InsertMany(context.TODO(), batch.Logs)
}

//日志存储协程
func (logSink *LogSink) writeLoop() {
	var (
		jobLog *entity.JobLog
		//正常日志批次
		logBatch *entity.LogBatch
		//超时日志批次
		timeoutBatch *entity.LogBatch
		commitTimer *time.Timer
	)

	for {
		select {
		case jobLog = <-logSink.logChan: //取出数据
			if logBatch == nil {
				logBatch = &entity.LogBatch{}
				commitTimer = time.AfterFunc(
					time.Second * time.Duration(yaml.SingYmlConfig.Worker.WriteJobLogCommitTimeout),
					//立即调用函数,生成一个回调函数,这个batch就是一个闭包上下文了,与外部的参数无关了,也就是生成了一个新的参数副本
					func(batch *entity.LogBatch) func(){
						//该函数就是定时器的函数,到期后会执行这个
					return func() {
						logSink.autoCommitChan <- batch
					}
					}(logBatch))
			}
			//把这条log写到mongo中-分批次写
			logBatch.Logs = append(logBatch.Logs, jobLog)
			//上传条件:
			//1.当日志池大小等于或者超过阈值
			//2.距离上次发送事件超过N秒之后并且日志池里有日志则上传
			if len(logBatch.Logs) >= yaml.SingYmlConfig.Worker.JobLogBatchSize{
				//如果到达阈值,则直接上传
				SingLogSink.saveLogs(logBatch)
				logBatch = nil
				//WriteJobLogCommitTimeout假如是5秒,加入上送日志阈值是100,在到达5秒的时候logBatch里的容量是99,所以会触发
				//下面的定时上送,但是在6秒的时候logBatch到达了100,这个时候这里也会上送一次,所以本该上送一次是100条的日志,
				//这个时候确上送了199条,所以我们首先需要在这里先将定时器给stop
				commitTimer.Stop()
			}
		case timeoutBatch = <-logSink.autoCommitChan://超过时间阈值的批次
			//WriteJobLogCommitTimeout假如是5秒,加入上送日志阈值是100,在到达5秒的时候logBatch里的容量是99,所以会触发
			//下面的定时上送,但是在6秒的时候logBatch到达了100,这个时候这里也会上送一次,所以本该上送一次是100条的日志,
			//这个时候确上送了199条,所以我们首先需要在这里先将定时器给stop,然后再判断下这个过期批次是否仍旧是当前的批次(俩里面的容量是否相等)
			if timeoutBatch != logBatch{
				//这里原因可能是logBatch已经上送了,被nil赋值了或者是生成了一个新的nil,所以我们需要跳过已经被提交的批次
				continue
			}
			//写入到mongo
			logSink.saveLogs(timeoutBatch)
			logBatch = nil

		}
	}
}

//发送日志
func (logSink *LogSink)Append (jobLog *entity.JobLog) {
	select {
	case logSink.logChan <- jobLog:
	default:
		//队列满了就丢弃
	}
}

//初始化
func InitLogSink() (err error) {
	var (
		client *mongo.Client
	)

	//建立mongo连接
	if client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(yaml.SingYmlConfig.Mongo.Uri).
		//连接超时时间
		SetConnectTimeout(time.Second*time.Duration(yaml.SingYmlConfig.Mongo.ConnectTimeOut)).
		SetAuth(options.Credential{Username: yaml.SingYmlConfig.Mongo.UserName, Password: yaml.SingYmlConfig.Mongo.Password})); err != nil {
		return
	}

	//选择db和collection
	SingLogSink = &LogSink{
		client:        client,
		logCollection: client.Database(yaml.SingYmlConfig.Mongo.Database).Collection("log"),
		logChan:       make(chan *entity.JobLog, 1000),
		autoCommitChan: make(chan  *entity.LogBatch,1000),
	}

	//启动一个mongoDB处理协程
	go SingLogSink.writeLoop()

	return

}
