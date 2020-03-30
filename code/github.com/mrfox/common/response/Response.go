/**
 * @Author: mrfox
 * @Description:
 * @File:  Response
 * @Version: 1.0.0
 * @Date: 2020/3/21 10:25 下午
 */
package response

type Response struct {
	//状态码
	Code int  `json:"code"`
	//返回信息
	Message string `json:"message"`
	//错误
	Error string `json:"error"`
	//数据
	Data interface{} `json:"data"`
}


//带参数的响应结果
func NewSuccess(data interface{}) *Response{
	return &Response{Code:200,Message:"success",Data:data}
}

//默认失败结果
func NewFail(message string,err string) Response{
	return Response{Code: 500, Message: message, Error:err,Data: nil}
}


//默认失败结果
func NewFailByCode(code int,message string) *Response{
	return &Response{Code:code,Message:message,Data:nil}
}


