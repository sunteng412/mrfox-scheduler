/**
 * @Author: mrfox
 * @Description:
 * @File:  LogUtils
 * @Version: 1.0.0
 * @Date: 2020/3/20 5:53 下午
 */
package log

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"io"
	yaml1 "mrfox-cron-common/yaml"
	"os"
)

//配置日志
func InitLog(e *echo.Echo, config *yaml1.Yaml) string {

	//输出到控制台
	if config.Log.LogType == "console"{
		e.Use(middleware.Logger())
	}else if config.Log.LogType == "file"{
		logFile, s, done := getLogFileConfig(config)
		if done {
			return s
		}
		//设置输出路径
		e.Logger.SetOutput(logFile)
	} else {
		logFile, s, done := getLogFileConfig(config)
		if done {
			return s
		}
		//设置输出到控制台和文件
		e.Logger.SetOutput(io.MultiWriter(os.Stdout, logFile))
	}

	//自定义日志实现
	//logFile, s, done := getLogFileConfig(config)
	//if done {
	//	return s
	//}
	//logger := New(logFile)
	//
	//e.Logger = logger

	//设置日志级别
	e.Logger.SetLevel(log.INFO)
	return ""
}

//配置日志变量
func getLogFileConfig(config *yaml1.Yaml) (*os.File, string, bool) {
	var logPath = config.Log.Path

	if len(logPath) == 0 {
		return nil, "日志路径错误,请重新设置", true
	}

	//如果log文件不存在，创建一个新的文件os.O_CREATE
	//打开文件的读写os.O_RDWR
	//log文件的权限位0666（即所有用户可读写）
	logFile, logErr := os.OpenFile(
		logPath,
		os.O_CREATE|os.O_RDWR|os.O_APPEND,
		0666)

	if logErr != nil {
		return nil, "日志路径:[" + logPath + "]错误,原因:" + logErr.Error(), true
	}
	return logFile, "", false
}
