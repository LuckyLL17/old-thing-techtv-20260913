package errors

import "fmt"

type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

func Wrap(code int, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

const (
	CodeBadRequest     = 40000
	CodeUnauthorized   = 40100
	CodeForbidden      = 40300
	CodeNotFound       = 40400
	CodeConflict       = 40900
	CodeValidation     = 42200
	CodeInternal       = 50000
	CodeDB             = 50001
	CodeToken          = 50002
	CodeFile           = 50003
	CodeRateLimit      = 42900
)

var (
	ErrBadRequest     = New(CodeBadRequest, "请求参数错误")
	ErrUnauthorized   = New(CodeUnauthorized, "未授权访问")
	ErrForbidden      = New(CodeForbidden, "权限不足")
	ErrNotFound       = New(CodeNotFound, "资源不存在")
	ErrConflict       = New(CodeConflict, "资源冲突")
	ErrValidation     = New(CodeValidation, "数据验证失败")
	ErrInternal       = New(CodeInternal, "服务器内部错误")
	ErrDB             = New(CodeDB, "数据库操作失败")
	ErrToken          = New(CodeToken, "令牌处理失败")
	ErrFile           = New(CodeFile, "文件处理失败")
	ErrRateLimit      = New(CodeRateLimit, "请求过于频繁")
	ErrUserExists     = New(CodeConflict, "用户已存在")
	ErrUserNotFound   = New(CodeNotFound, "用户不存在")
	ErrWrongPassword  = New(CodeUnauthorized, "密码错误")
	ErrTutorialNotFound = New(CodeNotFound, "教程不存在")
	ErrProjectNotFound  = New(CodeNotFound, "作品不存在")
)
