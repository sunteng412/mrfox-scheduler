/**
 * @Author: mrfox
 * @Description:
 * @File:  JobInfo
 * @Version: 1.0.0
 * @Date: 2020/3/21 4:51 下午
 */
package entity

import (
	"encoding/json"
	"strconv"
)

//任务
type Job struct {
	//任务名
	Name string `json:"name" form:"name" query:"name" validate:"required"`

	//任务执行命令
	Command string `json:"command" form:"command" validate:"required"`

	//cron表达式
	CronExpr string `json:"cronExpr" form:"cronExpr" validate:"required,cron"`

	//最后操作
	Operator int `json:"operator" form:"operator"`

}

func (job *Job) String()string{
	return "{Name:"+job.Name + ",Command:"+job.Command+",CronExpr:"+job.CronExpr+",Operator:"+strconv.Itoa(job.Operator)+"}"
}

//最后强杀状态
const KILLED_OPERATOR = 3

//反序列化
func UnpackJob(value []byte)(ret *Job,err error)  {
	var (
		job *Job
	)

	job = &Job{}
	if err = json.Unmarshal(value,job);err != nil{
		return
	}
	ret = job
	return
}


