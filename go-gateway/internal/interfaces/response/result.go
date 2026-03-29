package response

import (
	"time"
)

// Result 通用API响应结果
type Result struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// Success 成功响应
func Success(data interface{}) *Result {
	return &Result{
		Code:      200,
		Message:   "操作成功",
		Data:      data,
		Timestamp: time.Now().UnixMilli(),
	}
}

// SuccessWithMessage 成功响应（自定义消息和数据）
func SuccessWithMessage(message string, data interface{}) *Result {
	return &Result{
		Code:      200,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().UnixMilli(),
	}
}

// Fail 失败响应
func Fail(message string) *Result {
	return &Result{
		Code:      500,
		Message:   message,
		Timestamp: time.Now().UnixMilli(),
	}
}

// FailWithCode 带错误码的失败响应
func FailWithCode(code int, message string) *Result {
	return &Result{
		Code:      code,
		Message:   message,
		Timestamp: time.Now().UnixMilli(),
	}
}

// FailWithError 根据 error 类型生成失败响应
func FailWithError(err error) *Result {
	return &Result{
		Code:      500,
		Message:   err.Error(),
		Timestamp: time.Now().UnixMilli(),
	}
}

// BadRequest 参数错误
func BadRequest(message string) *Result {
	return FailWithCode(400, message)
}

// Unauthorized 未授权
func Unauthorized(message string) *Result {
	return FailWithCode(401, message)
}

// NotFound 资源不存在
func NotFound(message string) *Result {
	return FailWithCode(404, message)
}