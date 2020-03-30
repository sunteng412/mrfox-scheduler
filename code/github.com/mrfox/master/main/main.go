/**
 * @Author: mrfox
 * @Description:
 * @File:  ApiServer
 * @Version: 1.0.0
 * @Date: 2020/3/16 11:45 下午
 */
package main

import (
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	_log "mrfox-cron-common/log"
	"mrfox-cron-common/utils/validator"
	_yaml "mrfox-cron-common/yaml"
	"mrfox-cron-master/main/handler"
	"mrfox-cron-master/main/router"
)


var (
    //错误信息
	errStr string
	//echo
	e *echo.Echo
)


//-propertiesPath=/Users/mrfox/golandProject/mrfox-cron/github.com/mrfox/master/Application.yml
func main() {

	//解析配置
	_yaml.ParseProperties()

	//初始化一个echo
	e = echo.New()

	//初始化日志
	if errStr = _log.InitLog(e,_yaml.SingYmlConfig);len(errStr) > 0 {
		goto ERR
	}

	//初始化mongo
	if err :=handler.InitLogSink();err!=nil{
		errStr = err.Error()
		goto ERR
	}

	//初始化etcd配置
	if errStr = handler.InitJobMgr();len(errStr) > 0 {
		goto ERR
	}

	//绑定校验器
	e.Validator = validator.GetInstance()

	//注册路由
	router.RegisterRouter(e)

	//注册静态文件
	//设置Static中间件
	e.Static("/", "main/webroot")

	log.Fatal(e.Start(_yaml.SingYmlConfig.Cron.Port))

	ERR:
		fmt.Printf("启动错误,原因为:%v\n",errStr)
}




