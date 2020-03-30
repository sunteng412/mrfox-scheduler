/**
 * @Author: mrfox
 * @Description:
 * @File:  JobLog
 * @Version: 1.0.0
 * @Date: 2020/3/29 7:44 下午
 */
package entity

//任务执行日志
type JobLog struct {
	JobName string `json:"jobName" bson:"jobName"`//任务名
	WorkerIp string `json:"workerIp" bson:"workerIp"`//机器Ip
	Command string `json:"command" bson:"command"`//命令
	Err string `json:"err" bson:"err"`//错误原因
	OutPut string `json:"output" bson:"output"`//脚本结果输出
	PlanTime int64 `json:"planTime" bson:"planTime"`//计划开始时间
	ScheduleTime int64 `json:"scheduleTime" bson:"scheduleTime"`//实际调度时间
	StartTime int64 `json:"startTime" bson:"startTime"`//任务执行开始时间
	EndTime int64 `json:"endTime"bson:"endTime"`//任务执行结束时间
}

//日志批处理
type LogBatch struct {
	Logs []interface{}//多条日志
}