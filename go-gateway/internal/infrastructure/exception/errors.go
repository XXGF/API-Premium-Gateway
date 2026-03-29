package exception

import "fmt"

// BusinessError 业务异常
type BusinessError struct {
	ErrorCode string
	Message   string
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

// EntityNotFoundError 实体未找到异常
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

// ParamValidationError 参数校验异常
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

// ApiKeyError API Key 相关异常
type ApiKeyError struct {
	BusinessError
}

const (
	InvalidApiKey          = "INVALID_API_KEY"
	ApiKeyExpired          = "API_KEY_EXPIRED"
	ApiKeyRevoked          = "API_KEY_REVOKED"
	ApiKeyGenerationFailed = "API_KEY_GENERATION_FAILED"
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
