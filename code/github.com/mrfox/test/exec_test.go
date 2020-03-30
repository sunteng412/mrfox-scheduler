/**
 * @Author: mrfox
 * @Description:
 * @File:  main_test.go
 * @Version: 1.0.0
 * @Date: 2020/2/23 11:47 下午
 */
package test

import (
	"context"
	"fmt"
	"os/exec"
	"testing"
)

//测试执行shell命令
func Test_exec(t *testing.T) {

	var  (
		cmd *exec.Cmd
		output []byte
		err error
	)
	cmd = exec.Command("/bin/bash","-c","echo hello;ls -l")

	//创建子进程并执行
	//err = cmd.Run()

	fmt.Println(err)
	//创建子进程执行命令并且捕获pipe的输出
	if  output,err = cmd.CombinedOutput();err != nil{
		fmt.Println(err)
	}else {
		fmt.Println(string(output))
	}
}

type result struct {
	result []byte
	err error
}

//测试强杀命令
func Test_killCmd(t *testing.T) {
	var (
		//context
		ctx context.Context

		//取消函数
		cancelFunc context.CancelFunc

		//结果通道
		resultChan chan *result
		//结果
		res *result
	)

	//创建一个容量为1000的结果通道
	resultChan = make(chan *result,1000)

	 ctx, cancelFunc = context.WithCancel(context.TODO())
	go func() {
		var (
			cmd *exec.Cmd
		)
		//执行任务捕获输出
		cmd = exec.CommandContext(ctx,"/bin/bash","-c","sleep 2;echo hello...")

		//把任务结果给main协程
		output, err := cmd.CombinedOutput()
		resultChan <- &result{
			err:err,
			result:output,
		}
	}()

	 //在main协程里等待子协程的退出并打印任务执行结果
	 res = <-resultChan

	fmt.Println(string(res.result),res.err)

	//取消/强杀执行
	cancelFunc()
}