// Package exception 定义了项目的自定义异常体系。
//
// 该包属于 DDD 架构的基础设施层，提供统一的错误类型定义：
//   - BusinessError：通用业务异常（带可选错误码）
//   - EntityNotFoundError：实体未找到异常
//   - ParamValidationError：参数校验异常
//   - ApiKeyError：API Key 相关异常
package exception

import "fmt"

// BusinessError 通用业务异常。
//
// 包含可选的错误码（ErrorCode）和错误消息。
// 当 ErrorCode 不为空时，Error() 输出格式为 "[CODE] message"。
type BusinessError struct {
	ErrorCode string // 错误码（如 NO_AVAILABLE_INSTANCE、FALLBACK_EXHAUSTED）
	Message   string // 错误消息
}

func (e *BusinessError) Error() string {
	if e.ErrorCode != "" {
		return fmt.Sprintf("[%s] %s", e.ErrorCode, e.Message)
	}
	return e.Message
}

// NewBusinessError 创建业务异常
func NewBusinessError(message string) *BusinessError {
	return &BusinessError{Message: message}
}

// NewBusinessErrorWithCode 创建带错误码的业务异常
func NewBusinessErrorWithCode(errorCode, message string) *BusinessError {
	return &BusinessError{ErrorCode: errorCode, Message: message}
}

// EntityNotFoundError 实体未找到异常。
//
// 当通过 ID 或业务键查询实体不存在时抛出。
// 在 DDD 规范中，get 方法找不到必须返回此错误。
type EntityNotFoundError struct {
	Message string
}

func (e *EntityNotFoundError) Error() string {
	return e.Message
}

// NewEntityNotFoundError 创建实体未找到异常
func NewEntityNotFoundError(message string) *EntityNotFoundError {
	return &EntityNotFoundError{Message: message}
}

// ParamValidationError 参数校验异常。
//
// 当请求参数不符合要求时抛出，错误码固定为 PARAM_VALIDATION_ERROR。
type ParamValidationError struct {
	BusinessError
}

// NewParamValidationError 创建参数校验异常
func NewParamValidationError(message string) *ParamValidationError {
	return &ParamValidationError{
		BusinessError: BusinessError{
			ErrorCode: "PARAM_VALIDATION_ERROR",
			Message:   message,
		},
	}
}

// NewParamValidationErrorWithName 创建带参数名的校验异常
func NewParamValidationErrorWithName(paramName, message string) *ParamValidationError {
	return &ParamValidationError{
		BusinessError: BusinessError{
			ErrorCode: "PARAM_VALIDATION_ERROR",
			Message:   fmt.Sprintf("参数[%s]无效: %s", paramName, message),
		},
	}
}

// ApiKeyError API Key 相关异常。
//
// 包含多种预定义的错误类型：
//   - INVALID_API_KEY：Key 无效
//   - API_KEY_EXPIRED：Key 已过期
//   - API_KEY_REVOKED：Key 已撤销
//   - API_KEY_GENERATION_FAILED：Key 生成失败
type ApiKeyError struct {
	BusinessError
}

// API Key 错误码常量
const (
	InvalidApiKey          = "INVALID_API_KEY"          // Key 无效
	ApiKeyExpired          = "API_KEY_EXPIRED"           // Key 已过期
	ApiKeyRevoked          = "API_KEY_REVOKED"           // Key 已撤销
	ApiKeyGenerationFailed = "API_KEY_GENERATION_FAILED" // Key 生成失败
)

// NewApiKeyError 创建 API Key 异常
func NewApiKeyError(errorCode, message string) *ApiKeyError {
	return &ApiKeyError{
		BusinessError: BusinessError{
			ErrorCode: errorCode,
			Message:   message,
		},
	}
}

// ApiKeyInvalid 无效的 API Key
func ApiKeyInvalid() *ApiKeyError {
	return NewApiKeyError(InvalidApiKey, "API Key 无效")
}

// ApiKeyExpiredError 已过期的 API Key
func ApiKeyExpiredError() *ApiKeyError {
	return NewApiKeyError(ApiKeyExpired, "API Key 已过期")
}

// ApiKeyRevokedError 已撤销的 API Key
func ApiKeyRevokedError() *ApiKeyError {
	return NewApiKeyError(ApiKeyRevoked, "API Key 已被撤销")
}

// ApiKeyGenerationFailedError API Key 生成失败
func ApiKeyGenerationFailedError(reason string) *ApiKeyError {
	return NewApiKeyError(ApiKeyGenerationFailed, "API Key 生成失败: "+reason)
}

// IsBusinessError 判断是否为业务异常
func IsBusinessError(err error) bool {
	_, ok := err.(*BusinessError)
	return ok
}

// IsEntityNotFoundError 判断是否为实体未找到异常
func IsEntityNotFoundError(err error) bool {
	_, ok := err.(*EntityNotFoundError)
	return ok
}
