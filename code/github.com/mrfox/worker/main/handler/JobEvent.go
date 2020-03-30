/**
 * @Author: mrfox
 * @Description:
 * @File:  JobEvent
 * @Version: 1.0.0
 * @Date: 2020/3/25 10:17 下午
 */
package handler

import (
	"mrfox-cron-facade/entity"
	"strconv"
)

const (
	SaveJobStatus = 1
	UpdateJobStatus = 2
	KillJobStatus = 3
	DeleteJobStatus = 4
)

type Event struct {
	EventType int //1-Save,2-Update,3-Delete
	Job *entity.Job
}

//构建任务事件
func BuildJobEvent(eventType int, job *entity.Job) (jobEvent *Event) {
	return &Event{
		EventType:eventType,
		Job:job,
	}
}

//构建基础任务事件
func BuildBaseJobEvent( job *entity.Job) (jobEvent *Event) {
	return &Event{
		EventType:0,
		Job:job,
	}
}

func (event *Event) String() string{
	return "{eventType:"+strconv.Itoa(event.EventType)+",job:"+ event.Job.String()+"}"
}