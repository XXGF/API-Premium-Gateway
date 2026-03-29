// Package response 定义了统一的 API 响应结构。
//
// 该包属于 DDD 架构的接口层，提供统一的 Result 响应格式，
// 确保所有 API 接口返回一致的 JSON 结构。
package response

import (
	"time"
)

// Result 统一 API 响应结果。
//
// 所有接口都返回此结构，包含状态码、消息、数据和时间戳。
type Result struct {
	Code      int         `json:"code"`              // 状态码：200=成功，400=参数错误，401=未授权，404=未找到，500=服务器错误
	Message   string      `json:"message"`           // 操作结果描述
	Data      interface{} `json:"data,omitempty"`     // 响应数据，失败时为 null
	Timestamp int64       `json:"timestamp"`         // 响应时间戳（毫秒）
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