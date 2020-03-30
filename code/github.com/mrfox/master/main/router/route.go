/**
 * @Author: mrfox
 * @Description:
 * @File:  route
 * @Version: 1.0.0
 * @Date: 2020/3/21 1:37 下午
 */
package router

import (
	"github.com/labstack/echo"
)

//配置路由
func RegisterRouter(e *echo.Echo) {
	//保存任务
	e.POST("/job/save", SaveJob)
	//删除任务
	e.POST("/job/delete",DeleteJob)
	//查询所有任务
	e.GET("/job/list",ListJob)
	//杀死指定任务
	e.POST("/job/kill",KillJob)
	//查询任务日志
	e.GET("/job/log",HandleJobLog)

	//查询机器
	e.GET("/worker/list",ListWorker)
	//查询执行日志
	e.GET("/worker/log",WorkerLog)
}


