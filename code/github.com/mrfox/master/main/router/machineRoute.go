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
	"mrfox-cron-facade/entity"
	"mrfox-cron-master/main/handler"
	"net/http"
	"strconv"
)



//查询所有的job
func ListWorker(context echo.Context)error{
	context.Logger().Info("[cron]ListMachine:%v")
	var (
		workerList []*entity.WorkerInfo
		errStr string
	)
	//获取任务列表
	if workerList, errStr = handler.SingJobMgr.ListWorker();len(errStr) > 0 {
		return context.JSON(http.StatusOK,
			response.NewFail("获取任务失败",errStr))
	}

	return context.JSON(http.StatusOK,response.NewSuccess(workerList))
}

//查询机器的日志
func WorkerLog(context echo.Context)error{
	context.Logger().Info("[cron]WorkerLog:%v")
	var (
		ip string
		//从第多少条开始
		skipParam int
		//返回多少条
		limitParam int
		err error
		logArr []*entity.JobLog
	)
	ip = context.FormValue("ip")
	if skipParam, err = strconv.Atoi(context.FormValue("skip"));err != nil{
		skipParam = 0
	}
	if limitParam, err = strconv.Atoi(context.FormValue("limit"));err != nil{
		limitParam = 20
	}
	if ip =="" || len(ip) == 0 {
		return context.JSON(http.StatusInternalServerError,
			response.NewFail("ip不能为空","ip不能为空"))
	}

	if logArr, err = handler.SingLogMsg.ListMsg("",ip,int64(skipParam),int64(limitParam));err != nil{
		return context.JSON(http.StatusOK,response.NewFail("获取"+ip+"任务日志失败","获取任务日志失败"))
	}

	return context.JSON(http.StatusOK,response.NewSuccess(logArr))
}