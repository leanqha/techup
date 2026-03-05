package apperrors

import (
	"errors"
	"net/http"
)

type Code string

const (
	CodeInvalidArgument Code = "invalid_argument"
	CodeUnauthorized    Code = "unauthorized"
	CodeForbidden       Code = "forbidden"
	CodeNotFound        Code = "not_found"
	CodeConflict        Code = "conflict"
	CodeInternal        Code = "internal"
)

type Error struct {
	Code    Code
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return string(e.Code)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func New(code Code, message string) *Error {
	return &Error{Code: code, Message: message}
}

func Wrap(code Code, message string, err error) *Error {
	return &Error{Code: code, Message: message, Err: err}
}

func InvalidArgument(message string) *Error {
	return New(CodeInvalidArgument, message)
}

func Unauthorized(message string) *Error {
	return New(CodeUnauthorized, message)
}

func Forbidden(message string) *Error {
	return New(CodeForbidden, message)
}

func NotFound(message string) *Error {
	return New(CodeNotFound, message)
}

func Conflict(message string) *Error {
	return New(CodeConflict, message)
}

func Internal(message string, err error) *Error {
	return Wrap(CodeInternal, message, err)
}

func CodeOf(err error) Code {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return ""
}

func IsCode(err error, code Code) bool {
	return CodeOf(err) == code
}

func StatusCode(err error) int {
	switch CodeOf(err) {
	case CodeInvalidArgument:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func Message(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
