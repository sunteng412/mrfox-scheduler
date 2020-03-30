/**
 * @Author: mrfox
 * @Description:
 * @File:  ValidatorUtil
 * @Version: 1.0.0
 * @Date: 2020/3/21 9:44 下午
 */
package validator

import (
	zh2 "github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"gopkg.in/go-playground/validator.v9"
	"gopkg.in/go-playground/validator.v9/translations/zh"
	"strings"
	"sync"
)

type cronValidator struct {
	validator *validator.Validate
}

//校验
func (cv *cronValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}


var (
	//锁对象
	syncLock sync.Mutex

	//单例对象
	SingCronValidator *cronValidator

	//绑定中文翻译
	trans ut.Translator
)

//初始化
func initValidator()  {
	SingCronValidator = &cronValidator{validator:validator.New()}

	//绑定中文翻译
	bindZNTranslation()

	//绑定自定义校验
	bindCustomValidator()
}

//绑定自定义校验
func bindCustomValidator() {
	SingCronValidator.validator.RegisterValidation("cron",CronExp)
	RegisterTagTranslation("cron",
		map[string]string{
		"zh": "{0}错误的cron格式",
	})
}


// 自定义翻译
func RegisterTagTranslation(tag string, messages map[string]string) {
	for _, message := range messages {
		_ = SingCronValidator.validator.RegisterTranslation(tag, trans, func(ut ut.Translator) error {
			return ut.Add(tag, message, false)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, err := ut.T(fe.Tag(), fe.Field())
			if err != nil {
				return fe.(error).Error()
			}
			return t
		})
	}
}

//绑定中文翻译器
func bindZNTranslation(){
	//中文翻译器
	zhCh := zh2.New()
	uni := ut.New(zhCh)
	trans, _  = uni.GetTranslator("zh")
	//验证器注册翻译器
	zh.RegisterDefaultTranslations(SingCronValidator.validator, trans)
}

//返回实例
func GetInstance() *cronValidator{
	if SingCronValidator == nil{
		syncLock.Lock()
		if SingCronValidator == nil{
			//初始化
			initValidator()
			syncLock.Unlock()
		}
	}
	return SingCronValidator
}

//返回中文提示
func Translation2zh(err error) string{
	var build strings.Builder
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			build.WriteString(err.Translate(trans))
			build.WriteString(";")
		}
		return build.String()
	}
	return ""
}