/**
 * 任务类
 * @Author: mrfox
 * @Description:
 * @File:  Job
 * @Version: 1.0.0
 * @Date: 2020/3/19 10:16 下午
 */
package router

import (
	"github.com/labstack/echo"
	"mrfox-cron-common/response"
	"mrfox-cron-common/utils/validator"
	"mrfox-cron-facade/entity"
	"mrfox-cron-master/main/handler"
	"net/http"
	"strconv"
)

//新增任务
func SaveJob(context echo.Context) error {
	var (
		err error
		job *entity.Job
	)

	job = new(entity.Job)

	if err = context.Bind(job);err!=nil{
		return context.JSON(http.StatusOK,
			response.NewFail("参数传入不正确",err.Error()))
	}else if err = context.Validate(job);err!=nil{
		return context.JSON(http.StatusOK,
			response.NewFail("参数校验错误", validator.Translation2zh(err)))
	}

	context.Logger().Infof("[cron]SaveJob:%v",job)

	//保存到etcd中
	if oldJob, errStr := handler.SingJobMgr.SaveJob(job);errStr != ""{
		return context.JSON(http.StatusOK,
			response.NewFail("新增任务失败",errStr))
	}else {
		return context.JSON(http.StatusOK,response.NewSuccess(oldJob))
	}

}

//删除任务
func DeleteJob(context echo.Context) error {
	var (
		name string
	)

	 name = context.FormValue("name")

	 if name =="" || len(name) == 0 {
	 	return context.JSON(http.StatusInternalServerError,
			response.NewFail("任务名不能为空","任务名不能为空"))
	 }

	context.Logger().Infof("[cron]DeleteJob:%v",name)

	//保存到etcd中
	if oldJob, errStr := handler.SingJobMgr.DeleteJob(name);errStr != ""{
		return context.JSON(http.StatusInternalServerError,
			response.NewFail("删除任务失败",errStr))
	}else {
		return context.JSON(http.StatusOK,response.NewSuccess(oldJob))
	}

}


//查询所有的job
func ListJob(context echo.Context)error{
	context.Logger().Info("[cron]ListJob:%v")
	var (
		jobList []*entity.Job
		errStr string
	)
	//获取任务列表
	if jobList, errStr = handler.SingJobMgr.ListJob();len(errStr) > 0 {
		return context.JSON(http.StatusOK,
			response.NewFail("获取任务失败",errStr))
	}

	return context.JSON(http.StatusOK,response.NewSuccess(jobList))
}

//杀死指定名称的job
func KillJob(context echo.Context)error{
	var (
		name string
		errStr string
	)
	name = context.FormValue("name")

	if name =="" || len(name) == 0 {
		return context.JSON(http.StatusInternalServerError,
			response.NewFail("任务名不能为空","任务名不能为空"))
	}

	context.Logger().Infof("[cron]KillJob:%v",name)

	if errStr = handler.SingJobMgr.KllJob(name);len(errStr)>0{
		return context.JSON(http.StatusOK,response.NewFail("强杀"+name+"任务失败","强杀任务失败"))
	}

	return context.JSON(http.StatusOK,response.NewSuccess(nil))
}

//查询日志
func HandleJobLog(context echo.Context) error {
	var (
		name string
		//从第多少条开始
		skipParam int
		//返回多少条
		limitParam int
		err error
		logArr []*entity.JobLog
	)
	name = context.FormValue("name")
	if skipParam, err = strconv.Atoi(context.FormValue("skip"));err != nil{
		skipParam = 0
	}
	if limitParam, err = strconv.Atoi(context.FormValue("limit"));err != nil{
		limitParam = 20
	}
	if name =="" || len(name) == 0 {
		return context.JSON(http.StatusInternalServerError,
			response.NewFail("任务名不能为空","任务名不能为空"))
	}

	 if logArr, err = handler.SingLogMsg.ListMsg(name,"",int64(skipParam),int64(limitParam));err != nil{
	 	return context.JSON(http.StatusOK,response.NewFail("获取"+name+"任务日志失败","获取任务日志失败"))
	 }

	 return context.JSON(http.StatusOK,response.NewSuccess(logArr))
}
