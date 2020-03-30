# # mrfox-cron
go实现分布式shell任务调度
   &nbsp;&nbsp; 设计这套的初始想法是考虑到传统项目及docker化容器内执行shell的繁琐流程，想一想当我们需要在凌晨一点钟登录某台机器去执行命令或者需要一天收集一次系统内存
及磁盘使用情况时，内心万马奔腾，如果有这么一套程序可以提前定义好shell，到时间自己去执行，执行结果可以通过界面直接展示，那么心里岂不是美滋滋？

&nbsp;为应对一下场景:  

&emsp; 1.配置任务时需要手动登录服务器执行  

&emsp; 2.服务器宕机时，任务将终止调度，需要人工进行迁移  

&emsp;  3.排查问题低效，无妨方便的查看任务执行的准确时间和最后结果  

&nbsp;我们可以考虑实现一套基于cron的任务调度器，该项目基于golang进行实现，通过此代码可以学习到多任务调度实现逻辑、使用etcd实现分布式锁、任务分发、日志的异步处理事件广播及机器的服务注册发现。  

整体架构如下：

&emsp;应用角色分为master和worker,master负责作为B端负责任务管理、服务发现及日志管理，而worker负责任务调度嵌在服务器中，其中模块包含任务调度、本地任务同步、日志转储及注册应用。下面是worker和master的架构流程图以及master上的流程示例



&emsp;这里只是初步的实现，目前正在开发java客户端所用的starter,实现分布式任务管理。由界面化动态控制task的执行时间及状态。并支持热更。

憧憬:  

&emsp;1.实现通过ssh与服务器进行实时通信及互动  

&emsp;2.可作为调度系统保证应用的CI/CD,监听gitlab变化实现动态部署  

&emsp;3.机器的资源告罄钉钉预警及JVM监控  

&emsp;4.test/dev环境的资源(es7、mongo、TiDB、Mysql、Canal、Sentinel、Nacos)等的快速部署  



如何使用?
拉下代码后在master和worker的main目录下执行打包命令:
举例如下:

```
Mac 下编译 Linux 和 Windows 64位可执行程序
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build main.go

Linux 下编译 Mac 和 Windows 64位可执行程序
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build main.go
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build main.go

Windows 下编译 Mac 和 Linux 64位可执行程序
SET CGO_ENABLED=0
SET GOOS=darwin
SET GOARCH=amd64
go build main.go
 
SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64
go build main.go
```
