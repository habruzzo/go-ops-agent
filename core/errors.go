package core

import (
	"fmt"
	"runtime"
)

// ErrorType represents the type of error
type ErrorType string

const (
	ErrorTypeConfiguration ErrorType = "configuration"
	ErrorTypePlugin        ErrorType = "plugin"
	ErrorTypeNetwork       ErrorType = "network"
	ErrorTypeValidation    ErrorType = "validation"
	ErrorTypeTimeout       ErrorType = "timeout"
	ErrorTypeInternal      ErrorType = "internal"
)

// FrameworkError represents a framework-specific error with context
type FrameworkError struct {
	Type      ErrorType              `json:"type"`
	Message   string                 `json:"message"`
	Component string                 `json:"component"`
	Operation string                 `json:"operation"`
	Cause     error                  `json:"cause,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
	File      string                 `json:"file,omitempty"`
	Line      int                    `json:"line,omitempty"`
}

// Error implements the error interface
func (e *FrameworkError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %s (caused by: %v)", e.Type, e.Component, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s: %s", e.Type, e.Component, e.Message)
}

// Unwrap returns the underlying error
func (e *FrameworkError) Unwrap() error {
	return e.Cause
}

// WithContext adds context to the error
func (e *FrameworkError) WithContext(key string, value interface{}) *FrameworkError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// NewFrameworkError creates a new framework error
func NewFrameworkError(errorType ErrorType, component, operation, message string) *FrameworkError {
	_, file, line, _ := runtime.Caller(1)
	return &FrameworkError{
		Type:      errorType,
		Component: component,
		Operation: operation,
		Message:   message,
		File:      file,
		Line:      line,
	}
}

// WrapError wraps an existing error with framework context
func WrapError(err error, errorType ErrorType, component, operation, message string) *FrameworkError {
	_, file, line, _ := runtime.Caller(1)
	return &FrameworkError{
		Type:      errorType,
		Component: component,
		Operation: operation,
		Message:   message,
		Cause:     err,
		File:      file,
		Line:      line,
	}
}

// Error constructors for common error types
func NewConfigurationError(component, operation, message string) *FrameworkError {
	return NewFrameworkError(ErrorTypeConfiguration, component, operation, message)
}

func NewPluginError(component, operation, message string) *FrameworkError {
	return NewFrameworkError(ErrorTypePlugin, component, operation, message)
}

func NewNetworkError(component, operation, message string) *FrameworkError {
	return NewFrameworkError(ErrorTypeNetwork, component, operation, message)
}

func NewValidationError(component, operation, message string) *FrameworkError {
	return NewFrameworkError(ErrorTypeValidation, component, operation, message)
}

func NewTimeoutError(component, operation, message string) *FrameworkError {
	return NewFrameworkError(ErrorTypeTimeout, component, operation, message)
}

func NewInternalError(component, operation, message string) *FrameworkError {
	return NewFrameworkError(ErrorTypeInternal, component, operation, message)
}

// IsFrameworkError checks if an error is a FrameworkError
func IsFrameworkError(err error) bool {
	_, ok := err.(*FrameworkError)
	return ok
}

// GetErrorType returns the error type if it's a FrameworkError
func GetErrorType(err error) ErrorType {
	if ferr, ok := err.(*FrameworkError); ok {
		return ferr.Type
	}
	return ErrorTypeInternal
}
