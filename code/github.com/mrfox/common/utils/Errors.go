/**
 * @Author: mrfox
 * @Description:
 * @File:  Errors
 * @Version: 1.0.0
 * @Date: 2020/3/29 3:11 下午
 */
package utils

import "errors"

//自定义错误
var (
	ERR_LOCK_ALREADY_REQUIRED = errors.New("锁被占用!")
	ERR_CREATE_LEASE = errors.New("创建租约失败!")
	ERR_EXEC_TXN = errors.New("执行事务失败!")
)
