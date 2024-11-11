package errors

import "fmt"

// Error is the error message
type Error struct {
	HTTPCode *int   `json:"-"`
	Code     int    `json:"code"`
	Message  string `json:"message"`
}

// Error returns the error message
func (e *Error) Error() string {
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}

// GetHTTPCode returns the HTTP code
func (e *Error) GetHTTPCode() int {
	if e.HTTPCode == nil {
		return 400
	}
	return *e.HTTPCode
}

// New creates a new error
func New(code int, message string) error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// NewWithHTTPCode creates a new error with HTTP code
func NewWithHTTPCode(httpCode, code int, message string) error {
	return &Error{
		HTTPCode: &httpCode,
		Code:     code,
		Message:  message,
	}
}
