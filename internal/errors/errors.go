// Package errors provides application-level error types and sentinel errors.
package errors

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound           = errors.New("resource not found")
	ErrAlreadyExists      = errors.New("resource already exists")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrBadRequest         = errors.New("bad request")
	ErrInternal           = errors.New("internal server error")
	ErrOptimisticLock     = errors.New("optimistic lock conflict")
	ErrAccountBlocked     = errors.New("account is blocked")
	ErrSameAccount        = errors.New("cannot transfer to the same account")
	ErrAmountExceeded     = errors.New("amount exceeds max transfer limit")
	ErrDailyLimit         = errors.New("daily transfer limit exceeded")
	ErrNoExchangeRate     = errors.New("no exchange rate available")
	ErrOTPExpired         = errors.New("OTP code expired")
	ErrOTPInvalid         = errors.New("invalid OTP code")
	ErrAccountNotVerified = errors.New("account not verified")
)

type AppError struct {
	Err     error
	Message string
	Code    int
}

func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Err.Error()
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(err error, message string, code int) *AppError {
	return &AppError{Err: err, Message: message, Code: code}
}

func NotFound(message string) *AppError {
	return &AppError{Err: ErrNotFound, Message: message, Code: 404}
}

func BadRequest(message string) *AppError {
	return &AppError{Err: ErrBadRequest, Message: message, Code: 400}
}

func Conflict(message string) *AppError {
	return &AppError{Err: ErrAlreadyExists, Message: message, Code: 409}
}

func Unauthorized(message string) *AppError {
	return &AppError{Err: ErrUnauthorized, Message: message, Code: 401}
}

func Forbidden(message string) *AppError {
	return &AppError{Err: ErrForbidden, Message: message, Code: 403}
}

func Internal(message string) *AppError {
	return &AppError{Err: ErrInternal, Message: message, Code: 500}
}

func Wrap(err error, message string) *AppError {
	return &AppError{Err: err, Message: fmt.Sprintf("%s: %v", message, err), Code: 500}
}
