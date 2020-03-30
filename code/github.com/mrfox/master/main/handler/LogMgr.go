/**
 * @Author: mrfox
 * @Description:
 * @File:  LogMgr
 * @Version: 1.0.0
 * @Date: 2020/3/29 10:49 下午
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

//查询结构体
type JobLogFilter struct {
	JobName string `bson:"jobName,omitempty"`//任务名
	WorkerIp string `bson:"workerIp,omitempty"`//机器Ip
}

//排序规则
type SortLogByStartTime struct {
	SortOrder int `bson:"startTime"`//按{"startTime":-1}
}

//mongo存储日志
type LogMsg struct {
	client *mongo.Client
	//集合
	logCollection *mongo.Collection
}

var (
	//单例
	SingLogMsg *LogMsg
)

//日志管理
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
	SingLogMsg = &LogMsg{
		client:        client,
		logCollection: client.Database(yaml.SingYmlConfig.Mongo.Database).Collection("log"),
	}

	return
}

//按照任务名或者机器ip查询
func (logMsg *LogMsg) ListMsg(name string,ip string,skip int64,limit int64)(logArr []*entity.JobLog,err error){
	var (
		filter *JobLogFilter
		logSort *SortLogByStartTime
		//游标
		cursor *mongo.Cursor
		jobLog *entity.JobLog
	)

	//过滤条件
	filter = &JobLogFilter{}
	if name != ""{
		filter.JobName = name
	}
	if ip != ""{
		filter.WorkerIp = ip
	}
	//初始化logArr
	logArr = make([]*entity.JobLog,0)

	//查询
	logSort = &SortLogByStartTime{SortOrder:-1}
	if cursor,err = logMsg.logCollection.Find(context.TODO(),filter,
		&options.FindOptions{Sort:logSort,Skip:&skip,Limit:&limit});err!=nil{
		return
	}

	//延迟释放游标
	defer cursor.Close(context.TODO())
	for cursor.Next(context.TODO())  {
		jobLog = &entity.JobLog{}
		if err = cursor.Decode(jobLog);err != nil{
			return //日志不合法
		}
		logArr = append(logArr,jobLog)
	}
	return
}
