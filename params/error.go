package params

import (
	"fmt"
)

// ErrorCode holds the class of an error in machine-readable format.
// It is also an error in its own right.
type ErrorCode string

func (code ErrorCode) Error() string {
	return string(code)
}

func (code ErrorCode) ErrorCode() ErrorCode {
	return code
}

const (
	ErrNotFound         ErrorCode = "not found"
	ErrMetadataNotFound ErrorCode = "metadata not found"
	ErrForbidden        ErrorCode = "forbidden"
	ErrBadRequest       ErrorCode = "bad request"
	ErrDuplicateUpload  ErrorCode = "duplicate upload"
	ErrMultipleErrors   ErrorCode = "multiple errors"
	ErrUnauthorized     ErrorCode = "unauthorized"
	ErrMethodNotAllowed ErrorCode = "method not allowed"
)

// Error represents an error - it is returned for any response
// that fails.
// See http://tinyurl.com/knr3csp .
type Error struct {
	Message string
	Code    ErrorCode
	Info    map[string]*Error `json:",omitempty"`
}

// NewError returns a new *Error with the given error code
// and message.
func NewError(code ErrorCode, f string, a ...interface{}) error {
	return &Error{
		Message: fmt.Sprintf(f, a...),
		Code:    code,
	}
}

// Error implements error.Error.
func (e *Error) Error() string {
	return e.Message
}

// ErrorCode holds the class of the error in
// machine readable format.
func (e *Error) ErrorCode() string {
	return e.Code.Error()
}

func (e *Error) ErrorInfo() map[string]*Error {
	return e.Info
}

// Cause implements errgo.Causer.Cause.
func (e *Error) Cause() error {
	if e.Code != "" {
		return e.Code
	}
	return nil
}
