/**
 * @Author: mrfox
 * @Description:
 * @File:  Yaml
 * @Version: 1.0.0
 * @Date: 2020/3/20 5:54 下午
 */
package yaml

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"mrfox-cron-common/utils"
	"os"
)


//配置实体
type Yaml struct {
	//项目相关
	Cron struct{
		//端口
		Port string `yaml:"port"`
	}

	//日志相关
	Log struct{
		Path string `yaml:"path"`
		//日志输出类型
		LogType string `yaml:"logType"`
	}

	//etcd相关配置
	Etcd struct{
		//地址
		Endpoints string  `yaml:"endpoints"`
		//连接超时时间-单位:秒
		DialTimeout int `yaml:"dialTimeout"`
	}

	//worker相关配置
	Worker struct{
		//地址
		Group string  `yaml:"group"`
		//是否是容器应用-1:是,0-否
		IsContainer int `yaml:"isContainer"`
		//心跳间隔时间-单位:秒
		HeartbeatTime int64 `yaml:"heartbeatTime"`
		//日志批次上传阈值
		JobLogBatchSize int `yaml:"jobLogBatchSize"`
		//如果日志未超过阈值,但是间隔时间超过N秒之后,也会自动上传执行日志
		WriteJobLogCommitTimeout int64 `yaml:"writeJobLogCommitTimeout"`
	}

	//mongo连接配置
	Mongo struct{
		//uri
		Uri string `yaml:"uri"`
		//连接超时时间-单位:秒
		ConnectTimeOut int `yaml:"connectTimeOut"`
		//用户名
		UserName string `yaml:"userName"`
		//密码
		Password string `yaml:"password"`
		//database
		Database string `yaml:"database"`
	}

}

var (
	//单例
	SingYmlConfig *Yaml
)



//解析KV
func ParseProperties() {
	//配置路径--如果未配置则取当前项目同级目录
	var propertiesPath string
	SingYmlConfig = new(Yaml)

	var err error

	//解析yml文件路径参数
	flag.StringVar(&propertiesPath,"propertiesPath","","配置文件路径")
	//解析命令参数
	flag.Parse()

	//未配置则取当前项目同级目录
	if propertiesPath == ""{
		propertiesPath = utils.GetCurrentPath() + "/application.yml"
	}

	//读取
	file, err := ioutil.ReadFile(propertiesPath)
	if err != nil {
		fmt.Printf("[cron]配置路径错误,错误原因:%v\n",err.Error())
		os.Exit(1)
	}

	//解析
	err = yaml.Unmarshal(file, SingYmlConfig)
	if err != nil {
		fmt.Printf("[cron]解析yml失败,错误原因:%v\n",err.Error())
		os.Exit(1)
	}
}
