package service

import (
	"errors"
	"fmt"
)

type ErrorCode string

const (
	CodeBadUserInput ErrorCode = "BAD_USER_INPUT"
	CodeNotFound     ErrorCode = "NOT_FOUND"
	CodeConflict     ErrorCode = "CONFLICT"
	CodeInternal     ErrorCode = "INTERNAL"
)

type AppError struct {
	Code    ErrorCode
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	}
	return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewBadInput(msg string) *AppError {
	return &AppError{Code: CodeBadUserInput, Message: msg}
}

func NewNotFound(msg string) *AppError {
	return &AppError{Code: CodeNotFound, Message: msg}
}

func NewConflict(msg string, err error) *AppError {
	return &AppError{Code: CodeConflict, Message: msg, Err: err}
}

func NewInternal(msg string, err error) *AppError {
	return &AppError{Code: CodeInternal, Message: msg, Err: err}
}

func IsAppErrorCode(err error, code ErrorCode) bool {
	var appErr *AppError
	if !errors.As(err, &appErr) {
		return false
	}
	return appErr.Code == code
}
