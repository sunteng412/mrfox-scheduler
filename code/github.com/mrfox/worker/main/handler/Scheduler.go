/**
 * @Author: mrfox
 * @Description:
 * @File:  Scheduler
 * @Version: 1.0.0
 * @Date: 2020/3/27 9:20 下午
 */
package handler

import (
	"log"
	"mrfox-cron-common/utils"
	"mrfox-cron-facade/entity"
	"time"
)

//任务调度
type Scheduler struct {
	//etcd任务事件列表
	jobEventChan chan *Event
	//任务调度计划数据表
	jobPlanTable map[string]*JobSchedulePlan
	//当前正在执行的任务
	jobExecutingTable map[string]*JobExecuteInfo
	//定义chan,用于接收结果的任务结果队列
	jobResultChan chan *JobExecuteResult
}

var (
	SingScheduler *Scheduler
)

//初始化任务调度器
func InitScheduler() (err error) {
	//初始化
	SingScheduler = &Scheduler{
		jobEventChan:      make(chan *Event, 1000),
		jobPlanTable:      make(map[string]*JobSchedulePlan),
		jobExecutingTable: make(map[string]*JobExecuteInfo),
		jobResultChan:     make(chan *JobExecuteResult, 1000),
	}

	//启动调度协程
	go SingScheduler.SchedulerLoop()
	log.Print("[scheduler]初始化任务调度器成功...")
	return
}

//处理任务事件
func (scheduler *Scheduler) handleJobEvent(jobEvent *Event) {
	var (
		jobSchedulerPlan *JobSchedulePlan
		err              error
		//jobSchedulerPlan是否存在
		jobExisted bool
	)
	//判断一下事件类型
	switch jobEvent.EventType {
	case SaveJobStatus: //保存任务事件
		if jobSchedulerPlan, err = BuildJobSchedulePlan(jobEvent.Job); err != nil {
			//解析失败
			log.Printf("[scheduler]解析事件失败,e:%v", jobEvent)
			return
		}
		log.Printf("[scheduler]加入任务:%v", jobEvent.Job.Name)
		//插入任务调度计划数据表
		scheduler.jobPlanTable[jobEvent.Job.Name] = jobSchedulerPlan

	case UpdateJobStatus: //更新任务事件
		if jobSchedulerPlan, err = BuildJobSchedulePlan(jobEvent.Job); err != nil {
			//解析失败
			log.Printf("[scheduler]解析事件失败,e:%v", jobEvent)
			return
		}
		log.Printf("[scheduler]更新任务:%v", jobEvent.Job.Name)
		//插入任务调度计划数据表
		scheduler.jobPlanTable[jobEvent.Job.Name] = jobSchedulerPlan
	case DeleteJobStatus, KillJobStatus: //删除/强杀任务事件
		if jobSchedulerPlan, jobExisted = scheduler.jobPlanTable[jobEvent.Job.Name]; jobExisted {
			log.Printf("[scheduler]删除任务:%v", jobEvent.Job.Name)
			//如果存在就删除
			delete(scheduler.jobPlanTable, jobEvent.Job.Name)
		} else {
			//不存在
			log.Printf("[scheduler]事件不存在,无法删除,e:%v", jobEvent)
		}
	}
}

//调度协程-主循环,查看任务是否过期
func (scheduler *Scheduler) SchedulerLoop() {
	var (
		jobEvent       *Event
		schedulerAfter time.Duration
		schedulerTimer *time.Timer
		jobResult      *JobExecuteResult
	)

	//初始化一次(最开始是休眠一秒)
	schedulerAfter = scheduler.TryScheduler()

	//调度的延时定时器
	schedulerTimer = time.NewTimer(schedulerAfter)

	//定时任务
	for {
		select {
		//监听任务变化事件
		case jobEvent = <-scheduler.jobEventChan:
			//处理事件-对内存中维护的任务列表做增删改查
			scheduler.handleJobEvent(jobEvent)
		case <-schedulerTimer.C: //最近的任务到期
		case jobResult = <-scheduler.jobResultChan: //监听任务执行结果
			scheduler.HandleJobResult(jobResult)

		}
		//调度一次任务
		schedulerAfter = scheduler.TryScheduler()
		//重置一下调度间隔
		schedulerTimer.Reset(schedulerAfter)
	}
}

//推送变化事件
func (scheduler *Scheduler) PushEvent(jobEvent *Event) {
	scheduler.jobEventChan <- jobEvent
}

//尝试执行任务
func (scheduler *Scheduler) TryStartJob(jobSchedulePlan *JobSchedulePlan) {
	log.Printf("[scheduler]TryStartJob,[%v]", jobSchedulePlan.Job.Name)

	//执行的任务与可能会很久,比如一个任务需要1分钟执行60次,但是每次需要执行60秒,
	//所以一个机器同一时间内只能相同任务不会执行多次,所以需要防止并发,如果任务正在执行,调过本次调度
	var (
		jobExecuteInfo *JobExecuteInfo
		//任务是否在执行
		jobExecuting bool
	)

	//如果任务正在执行,则跳过本次调度
	if jobExecuteInfo, jobExecuting = scheduler.jobExecutingTable[jobSchedulePlan.Job.Name]; jobExecuting {
		log.Printf("[scheduler][%v]当前任务正在执行,取消本次调度", jobSchedulePlan.Job.Name)
		return
	}

	//构建执行状态信息
	jobExecuteInfo = BuildJobExecuteInfo(jobSchedulePlan)

	//保存执行状态
	scheduler.jobExecutingTable[jobSchedulePlan.Job.Name] = jobExecuteInfo

	//执行任务
	SingExecutor.ExecuteJob(jobExecuteInfo)
}

//重新计算任务调度状态
func (scheduler *Scheduler) TryScheduler() (scheduleAfter time.Duration) {
	var (
		jobSchedulePlan *JobSchedulePlan
		nowTime         time.Time
		//后面最近一次执行时间
		nearTime *time.Time
	)

	//如果任务表为空的话,就睡眠一秒
	if len(scheduler.jobPlanTable) == 0 {
		scheduleAfter = time.Second * 1
	}

	//获取当前时间
	nowTime = time.Now()

	//1.遍历所有任务
	for _, jobSchedulePlan = range scheduler.jobPlanTable {
		if jobSchedulePlan.NextExecTime.Before(nowTime) || jobSchedulePlan.NextExecTime.Equal(nowTime) {

			//尝试执行任务
			scheduler.TryStartJob(jobSchedulePlan)

			jobSchedulePlan.NextExecTime = jobSchedulePlan.Expr.Next(nowTime) //更新下次执行时间
		}

		//统计最近一个要过期的任务执行时间
		if nearTime == nil || jobSchedulePlan.NextExecTime.Before(*nearTime) {
			//当前任务的下次执行时间就是最近的
			nearTime = &jobSchedulePlan.NextExecTime
		}
		//统计最近的要过期的任务的时间(N秒之后过期 == scheduleAfter)
		scheduleAfter = (*nearTime).Sub(nowTime)
	}

	return scheduleAfter
}

//回传任务执行结果
func (scheduler *Scheduler) PushJobResult(jobResult *JobExecuteResult) {
	scheduler.jobResultChan <- jobResult
}

//处理执行结果
func (scheduler *Scheduler) HandleJobResult(result *JobExecuteResult) {
	var (
		jobLog *entity.JobLog
	)

	//删除正在执行表的数据,让下次能够正常执行
	delete(scheduler.jobExecutingTable, result.ExecuteInfo.Job.Name)
	log.Printf("[schedule][%v]执行结束,执行结果:[%v]",
		result.ExecuteInfo.Job.Name, string(result.Output))

	//插入mongo
	if result.Err != utils.ERR_LOCK_ALREADY_REQUIRED { //排除锁抢占失败
		jobLog = &entity.JobLog{
			JobName:      result.ExecuteInfo.Job.Name,
			WorkerIp:     WorkerIp,
			Command:      result.ExecuteInfo.Job.Command,
			OutPut:       string(result.Output),
			PlanTime:     result.ExecuteInfo.PlanTime.UnixNano() / 1000 / 1000, //精确到毫秒
			ScheduleTime: result.ExecuteInfo.RealTime.UnixNano() / 1000 / 1000,
			StartTime:    result.StartTime.UnixNano() / 1000 / 1000,
			EndTime:      result.EndTime.UnixNano() / 1000 / 1000,
		}
		if result.Err != nil {
			jobLog.Err = result.Err.Error()
		} else {
			jobLog.Err = ""
		}
	}

	SingLogSink.Append(jobLog)
}
