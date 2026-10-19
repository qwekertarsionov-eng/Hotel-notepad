package domain

import "fmt"

// FieldError describes a single invalid field for a validation error.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// DomainError is a typed domain error with a code and a message.
type DomainError struct {
	Code    string
	Message string
	Fields  []FieldError
	Err     error
}

// Error implements the error interface.
func (d DomainError) Error() string {
	if d.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", d.Code, d.Message, d.Err)
	}
	return fmt.Sprintf("%s: %s", d.Code, d.Message)
}

// Unwrap returns the wrapped error.
func (d DomainError) Unwrap() error {
	return d.Err
}

// NewDomainError creates a new DomainError.
func NewDomainError(code, message string) DomainError {
	return DomainError{Code: code, Message: message}
}

// NewValidationError creates a VALIDATION_FAILED error with field details.
func NewValidationError(message string, fields ...FieldError) DomainError {
	return DomainError{
		Code:    "VALIDATION_FAILED",
		Message: message,
		Fields:  fields,
	}
}

// ErrNotImplemented is returned by stub operations that are not yet implemented.
var ErrNotImplemented = DomainError{Code: "NOT_IMPLEMENTED", Message: "not implemented"}
