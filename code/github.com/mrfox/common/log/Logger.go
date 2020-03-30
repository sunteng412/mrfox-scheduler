/**
 * 自定义日志
 * @Author: mrfox
 * @Description:
 * @File:  Logger
 * @Version: 1.0.0
 * @Date: 2020/3/19 11:18 下午
 */
package log

import (
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"github.com/rs/zerolog"
	"io"
	"path/filepath"
	"strconv"
)

var logDefaultHeader = map[string]string{
	"time":   "${time_rfc3339_nano}",
	"level":  "${level}",
	"prefix": "${prefix}",
	"file":   "${file}",
	"line":   "${line}",
}

func init() {
	zerolog.CallerMarshalFunc = func(file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}
}
var l echo.Logger = &CronLogger{}

type CronLogger struct {
	*log.Logger
	ZeroLog zerolog.Logger
}



func  New(writer io.Writer) *CronLogger {
	l := &CronLogger{
		Logger:  log.New("-"),
		ZeroLog: zerolog.New(writer).With().Caller().Timestamp().Logger(),
	}

	// log 默认是 ERROR，将 Level 默认都改为 INFO
	l.SetLevel(log.INFO)

	l.Logger.SetOutput(writer)

	return l
}


//func New(writer io.Writer) *Logger {
//	l := &Logger{
//		Logger:  log.New("-"),
//		ZeroLog: zerolog.New(writer).With().Caller().Timestamp().Logger(),
//	}
//
//	// log 默认是 ERROR，将 Level 默认都改为 INFO
//	l.SetLevel(log.INFO)
//
//	l.Logger.SetOutput(writer)
//
//	return l
//}

func (l *CronLogger) SetOutput(writer io.Writer) {
	l.Logger.SetOutput(writer)
	l.ZeroLog.Output(writer)
}

func (l *CronLogger) SetLevel(level log.Lvl) {
	l.Logger.SetLevel(level)
	if level == log.OFF {
		l.ZeroLog = l.ZeroLog.Level(zerolog.Disabled)
	} else {
		zeroLevel := int8(level) - 1
		l.ZeroLog = l.ZeroLog.Level(zerolog.Level(zeroLevel))
	}
}




