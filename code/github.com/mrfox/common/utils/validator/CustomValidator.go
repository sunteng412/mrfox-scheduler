/**
 * @Author: mrfox
 * @Description:
 * @File:  CustomValidator
 * @Version: 1.0.0
 * @Date: 2020/3/22 10:46 下午
 */
package validator

import (
	"github.com/gorhill/cronexpr"
	"gopkg.in/go-playground/validator.v9"
)

//cron表达式校验
func CronExp(fl validator.FieldLevel) bool {
	if _, err := cronexpr.Parse(fl.Field().String()); err != nil {
		return false
	}
	return true
}
