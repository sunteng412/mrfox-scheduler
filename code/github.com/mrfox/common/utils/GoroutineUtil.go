/**
 * @Author: mrfox
 * @Description:
 * @File:  GoroutineUtil
 * @Version: 1.0.0
 * @Date: 2020/3/29 3:01 下午
 */
package utils

import (
	"bytes"
	"runtime"
)

//获取goroutine的ID
func GetGoroutineId() string {
	b := make([]byte,64)
	runtime.Stack(b,false)
	b = bytes.TrimPrefix(b,[]byte("goroutine "))
	b = b[:bytes.IndexByte(b,' ')]
	return string(b)
}
