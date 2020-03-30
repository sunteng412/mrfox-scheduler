/**
 * @Author: mrfox
 * @Description:
 * @File:  HelloController
 * @Version: 1.0.0
 * @Date: 2020/3/15 10:01 下午
 */
package main

import (
	"github.com/labstack/echo"
	"net/http"
)

// 业务处理
func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}
