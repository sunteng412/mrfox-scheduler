/**
 * @Author: mrfox
 * @Description:
 * @File:  JobSchedulePlan
 * @Version: 1.0.0
 * @Date: 2020/3/27 9:27 下午
 */
package handler

import (
	"github.com/gorhill/cronexpr"
	"mrfox-cron-facade/entity"
	"time"
)

//任务调度计划
type JobSchedulePlan struct {
	//要调度的任务信息
	Job *entity.Job
	//cron表达式
	Expr *cronexpr.Expression
	//任务的下次执行时间
	NextExecTime time.Time
}

//任务执行状态信息
type JobExecuteInfo struct {
	//任务信息
	Job *entity.Job
	//任务的计划调度时间
	PlanTime time.Time
	//任务的实际调度时间
	RealTime time.Time
}

//构造执行状态信息
func BuildJobExecuteInfo(jobSchedulePlan *JobSchedulePlan)(jobExecuteInfo *JobExecuteInfo){
	return  &JobExecuteInfo{
		Job: jobSchedulePlan.Job     ,
		//计划调度时间
		PlanTime: jobSchedulePlan.NextExecTime,
		RealTime: time.Now(),
	}
}


//构造任务执行计划
func BuildJobSchedulePlan(job *entity.Job) (jobSchedulePlan *JobSchedulePlan,err error) {
	var(
		expr *cronexpr.Expression
	)

	//解析cron表达式
	if expr, err = cronexpr.Parse(job.CronExpr);err != nil{
		return
	}

	//生成调度计划
	jobSchedulePlan = &JobSchedulePlan{
		Job:job,
		Expr:expr,
		NextExecTime:expr.Next(time.Now()),
	}
	return
}
