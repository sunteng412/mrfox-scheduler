/**
 * @Author: mrfox
 * @Description:
 * @File:  WorkerInfo
 * @Version: 1.0.0
 * @Date: 2020/3/27 12:33 下午
 */
package entity

type WorkerInfo struct {
	//组别
	Group string  `json:"group"`
	//是否是容器应用-1:是,0-否
	IsContainer int `json:"isContainer"`
	//心跳间隔时间-单位:秒
	HeartbeatTime int64 `json:"heartbeatTime"`
	//IP
	Ip string `json:"ip"`
}
