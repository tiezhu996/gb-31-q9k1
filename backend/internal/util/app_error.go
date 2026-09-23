package util

import "fmt"

// AppError 业务错误：handler 与错误中间件统一转换为 JSON 响应。
type AppError struct {
	Code    int
	Message string
	HTTP    int
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("code=%d message=%s err=%v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("code=%d message=%s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error { return e.Err }

// NewAppError 构造业务错误。
func NewAppError(code int, message string, http int, err error) *AppError {
	return &AppError{Code: code, Message: message, HTTP: http, Err: err}
}

// BadRequest 400 业务错误。
func BadRequest(message string, err error) *AppError {
	return NewAppError(40000, message, 400, err)
}

// Validation 422 校验错误。
func Validation(message string, err error) *AppError {
	return NewAppError(42200, message, 422, err)
}

// Unauthorized 401 业务错误。
func Unauthorized(message string, err error) *AppError {
	return NewAppError(40100, message, 401, err)
}

// Forbidden 403 业务错误。
func Forbidden(message string, err error) *AppError {
	return NewAppError(40300, message, 403, err)
}

// NotFound 404 业务错误。
func NotFound(message string, err error) *AppError {
	return NewAppError(40400, message, 404, err)
}

// Conflict 409 业务错误。
func Conflict(message string, err error) *AppError {
	return NewAppError(40900, message, 409, err)
}

// ConflictWithCode 409 业务错误，携带细分错误码（如约伴撞车 40905）。
func ConflictWithCode(code int, message string, err error) *AppError {
	return NewAppError(code, message, 409, err)
}

// Internal 500 业务错误。
func Internal(message string, err error) *AppError {
	return NewAppError(50000, message, 500, err)
}

// ContentRejected 内容审核不通过错误。
func ContentRejected(message string, err error) *AppError {
	return NewAppError(42201, message, 422, err)
}
