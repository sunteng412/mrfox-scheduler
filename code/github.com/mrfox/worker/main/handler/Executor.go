/**
 * @Author: mrfox
 * @Description:
 * @File:  Executor
 * @Version: 1.0.0
 * @Date: 2020/3/28 4:09 下午
 */
package handler

import (
	"context"
	"log"
	_const "mrfox-cron-facade/const"
	"os/exec"
	"time"
)

//任务执行器
type Executor struct {

}

//任务执行结果
type JobExecuteResult struct {
	//任务执行状态信息
	ExecuteInfo *JobExecuteInfo
	//脚本输出
	Output []byte
	//脚本执行的错误原因
	Err error
	//启动时间
	StartTime time.Time
	//结束时间
	EndTime time.Time
}

var (
	SingExecutor *Executor
)

//执行信息
func (executor *Executor) ExecuteJob(jobExecuteInfo *JobExecuteInfo) {
	log.Printf("[scheduler]执行任务[%v],理论执行时间[%v],实际执行时间[%v]",jobExecuteInfo.Job.Name,jobExecuteInfo.PlanTime,jobExecuteInfo.RealTime)
	go func() {
		var (
			cmd *exec.Cmd
			err error
			output []byte
			//执行结果
			result *JobExecuteResult
			//开始时间
			startTime time.Time
			//分布式锁
			jobLock *DistributedLock
		)


		//获取分布式锁
		 jobLock = GetDistributedLock(_const.DISTRIBUTED_LOCK_PREFIX +jobExecuteInfo.Job.Name)
		//上锁
		 err = jobLock.TryLock(5)
		 defer jobLock.Unlock()

		 if err != nil{
		 	log.Printf("[executor]上锁失败")
		 	result.Err  = err
		 }else {
			 //获取可执行的shell命令
			 cmd = exec.CommandContext(context.TODO(),"/bin/bash","-c",jobExecuteInfo.Job.Command)

			 startTime = time.Now()

			 //执行并捕获输出
			 if output,err = cmd.CombinedOutput();err != nil{
				 log.Printf("[executor][%v]任务执行错误,CMD:[%v],err:%v",jobExecuteInfo.Job.Name,jobExecuteInfo.Job.Command,err)
			 }
		 }

		//任务执行完成后返回结果
		result = &JobExecuteResult{
			ExecuteInfo: jobExecuteInfo,
			Output:      output,
			Err:         err,
			StartTime:   startTime,
			EndTime:     time.Now(),
		}

		//回传结果
		SingScheduler.PushJobResult(result)
	}()
}

//初始化执行器
func InitExecutor()(err error)  {
	SingExecutor = &Executor{}
	return
}