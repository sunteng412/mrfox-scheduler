/**
 * @Author: mrfox
 * @Description:
 * @File:  cron_test
 * @Version: 1.0.0
 * @Date: 2020/3/8 10:25 下午
 */
package test

import (
	"fmt"
	"github.com/gorhill/cronexpr"
	"testing"
	"time"
)

//测试cron
func Test_cron(t *testing.T) {

	var (
		expr *cronexpr.Expression
		err error
	)
	//那一分钟(0-59) 哪小时(0-23) 哪天(1-31) 哪月(1-12) 星期几(0-6)
	if expr, err = cronexpr.Parse("* * * * * *");err != nil{
		fmt.Println(err)
		return
	}else {
		fmt.Println(expr.NextN(time.Now(),2))
	}
}


//测试cron之后执行
func Test_afterCronExec(t *testing.T) {
	var (
		expr *cronexpr.Expression
		err error
		//当前时间
		now time.Time
		//下一次时间
		nextTime time.Time
	)

	//那一分钟(0-59) 哪小时(0-23) 哪天(1-31) 哪月(1-12) 星期几(0-6)
	if expr, err = cronexpr.Parse("*/5 * * * * * *");err != nil {
		fmt.Println(err)
		return
	}

	//当前时间
	now = time.Now()
	//下次调度时间
	nextTime = expr.Next(now)

	//等待定时器超时
	time.AfterFunc(nextTime.Sub(now), func() {
		fmt.Println("嘿嘿",nextTime)
	})

	time.Sleep(time.Second * 5)
}
