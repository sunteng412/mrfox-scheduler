/**
 * @Author: mrfox
 * @Description:
 * @File:  fileUtil
 * @Version: 1.0.0
 * @Date: 2020/3/21 4:42 下午
 */
package utils

import (
	"os"
	"strings"
)

//获取当前项目路径
func GetCurrentPath() string {
	dir, _ := os.Getwd()
	return strings.Replace(dir, "\\", "/", -1)
}
