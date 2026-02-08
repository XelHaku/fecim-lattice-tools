// Package errors provides user-friendly error types and handling utilities for FeCIM tools.
//
// This package defines common error types with:
// - User-friendly messages explaining what went wrong
// - Technical details for debugging
// - Suggested recovery actions
// - Error categorization for appropriate handling
package errors

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

// Category classifies errors for handling strategies.
type Category int

const (
	// CategoryUnknown is an unclassified error.
	CategoryUnknown Category = iota
	// CategoryConfig indicates configuration/setup errors.
	CategoryConfig
	// CategoryIO indicates file/network I/O errors.
	CategoryIO
	// CategoryData indicates invalid or corrupt data.
	CategoryData
	// CategoryResource indicates resource exhaustion (memory, GPU, etc).
	CategoryResource
	// CategoryInternal indicates internal/programming errors.
	CategoryInternal
	// CategoryUser indicates user input errors.
	CategoryUser
)

func (c Category) String() string {
	switch c {
	case CategoryConfig:
		return "Configuration"
	case CategoryIO:
		return "I/O"
	case CategoryData:
		return "Data"
	case CategoryResource:
		return "Resource"
	case CategoryInternal:
		return "Internal"
	case CategoryUser:
		return "User Input"
	default:
		return "Unknown"
	}
}

// FeCIMError is the base error type for all FeCIM errors.
// It provides user-friendly messaging with technical details for debugging.
type FeCIMError struct {
	// Category classifies the error type.
	Category Category
	// Message is the user-friendly error message.
	Message string
	// Details provides technical debugging information.
	Details string
	// Recovery suggests how to fix the error.
	Recovery string
	// Cause is the underlying error if any.
	Cause error
	// Location is where the error occurred (file:line).
	Location string
}

// Error implements the error interface.
func (e *FeCIMError) Error() string {
	var sb strings.Builder
	sb.WriteString(e.Message)
	if e.Details != "" {
		sb.WriteString(" (")
		sb.WriteString(e.Details)
		sb.WriteString(")")
	}
	return sb.String()
}

// Unwrap returns the underlying cause for errors.Is/As.
func (e *FeCIMError) Unwrap() error {
	return e.Cause
}

// UserMessage returns a clean user-facing message.
func (e *FeCIMError) UserMessage() string {
	if e.Recovery != "" {
		return fmt.Sprintf("%s\n\nTo fix this: %s", e.Message, e.Recovery)
	}
	return e.Message
}

// FullMessage returns the complete error with all details.
func (e *FeCIMError) FullMessage() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[%s Error] %s\n", e.Category, e.Message))
	if e.Details != "" {
		sb.WriteString(fmt.Sprintf("  Details: %s\n", e.Details))
	}
	if e.Recovery != "" {
		sb.WriteString(fmt.Sprintf("  Recovery: %s\n", e.Recovery))
	}
	if e.Location != "" {
		sb.WriteString(fmt.Sprintf("  Location: %s\n", e.Location))
	}
	if e.Cause != nil {
		sb.WriteString(fmt.Sprintf("  Caused by: %v\n", e.Cause))
	}
	return sb.String()
}

// Recoverable returns true if the error can potentially be recovered from.
func (e *FeCIMError) Recoverable() bool {
	switch e.Category {
	case CategoryInternal:
		return false
	case CategoryResource:
		return false // Usually can't recover from OOM etc.
	default:
		return e.Recovery != ""
	}
}

// NewError creates a new FeCIMError with location tracking.
func NewError(category Category, message string) *FeCIMError {
	return &FeCIMError{
		Category: category,
		Message:  message,
		Location: getCallerLocation(2),
	}
}

// WithDetails adds technical details to the error.
func (e *FeCIMError) WithDetails(format string, args ...interface{}) *FeCIMError {
	e.Details = fmt.Sprintf(format, args...)
	return e
}

// WithRecovery adds recovery instructions to the error.
func (e *FeCIMError) WithRecovery(recovery string) *FeCIMError {
	e.Recovery = recovery
	return e
}

// WithCause wraps an underlying error.
func (e *FeCIMError) WithCause(cause error) *FeCIMError {
	e.Cause = cause
	return e
}

// getCallerLocation returns file:line of the caller.
func getCallerLocation(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}
	// Get just the filename, not full path
	parts := strings.Split(file, "/")
	if len(parts) > 0 {
		file = parts[len(parts)-1]
	}
	return fmt.Sprintf("%s:%d", file, line)
}

// --- Common Error Constructors ---

// ConfigError creates a configuration error.
func ConfigError(message string) *FeCIMError {
	return NewError(CategoryConfig, message)
}

// ConfigNotFound creates an error for missing configuration.
func ConfigNotFound(configName string) *FeCIMError {
	return NewError(CategoryConfig, fmt.Sprintf("Configuration '%s' not found", configName)).
		WithRecovery(fmt.Sprintf("Create the configuration file or check the path. Expected locations: config/%s.yaml", configName))
}

// IOError creates an I/O error.
func IOError(message string, cause error) *FeCIMError {
	return NewError(CategoryIO, message).WithCause(cause)
}

// FileNotFound creates a file not found error with helpful recovery.
func FileNotFound(path string) *FeCIMError {
	return NewError(CategoryIO, fmt.Sprintf("File not found: %s", path)).
		WithRecovery("Check that the file exists and the path is correct")
}

// DataError creates a data validation error.
func DataError(message string) *FeCIMError {
	return NewError(CategoryData, message)
}

// InvalidData creates an error for invalid data format.
func InvalidData(dataType, problem string) *FeCIMError {
	return NewError(CategoryData, fmt.Sprintf("Invalid %s: %s", dataType, problem))
}

// ResourceError creates a resource exhaustion error.
func ResourceError(resource, message string) *FeCIMError {
	return NewError(CategoryResource, fmt.Sprintf("%s resource error: %s", resource, message))
}

// GPUError creates a GPU-related error with common recovery steps.
func GPUError(message string, cause error) *FeCIMError {
	return NewError(CategoryResource, fmt.Sprintf("GPU error: %s", message)).
		WithCause(cause).
		WithRecovery("Check GPU drivers are installed. Try running with CPU fallback if available.")
}

// InternalError creates an internal programming error.
func InternalError(message string) *FeCIMError {
	return NewError(CategoryInternal, fmt.Sprintf("Internal error: %s", message)).
		WithRecovery("This is a bug. Please report it with the error details.")
}

// UserInputError creates a user input validation error.
func UserInputError(message string) *FeCIMError {
	return NewError(CategoryUser, message)
}

// InvalidParameter creates an error for invalid parameter values.
func InvalidParameter(param string, value interface{}, expected string) *FeCIMError {
	return NewError(CategoryUser, fmt.Sprintf("Invalid value for '%s': got %v", param, value)).
		WithDetails("Expected: %s", expected).
		WithRecovery(fmt.Sprintf("Provide a valid value for '%s' (%s)", param, expected))
}

// --- Error Checking Utilities ---

// Is checks if an error matches a target error.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As attempts to extract a specific error type.
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// IsFeCIMError checks if an error is a FeCIMError.
func IsFeCIMError(err error) bool {
	var fErr *FeCIMError
	return As(err, &fErr)
}

// GetFeCIMError extracts a FeCIMError if present.
func GetFeCIMError(err error) *FeCIMError {
	var fErr *FeCIMError
	if As(err, &fErr) {
		return fErr
	}
	return nil
}

// IsRecoverable checks if an error can be recovered from.
func IsRecoverable(err error) bool {
	if fErr := GetFeCIMError(err); fErr != nil {
		return fErr.Recoverable()
	}
	return false
}

// Wrap wraps a standard error into a FeCIMError.
func Wrap(err error, message string) *FeCIMError {
	if err == nil {
		return nil
	}
	// If already a FeCIMError, just add context
	if fErr := GetFeCIMError(err); fErr != nil {
		return &FeCIMError{
			Category: fErr.Category,
			Message:  message,
			Details:  fErr.Details,
			Recovery: fErr.Recovery,
			Cause:    fErr,
			Location: getCallerLocation(2),
		}
	}
	return NewError(CategoryUnknown, message).WithCause(err)
}

// WrapWithCategory wraps an error with a specific category.
func WrapWithCategory(err error, category Category, message string) *FeCIMError {
	if err == nil {
		return nil
	}
	return NewError(category, message).WithCause(err)
}

// --- Sentinel Errors ---

var (
	// ErrNotInitialized indicates a resource was used before initialization.
	ErrNotInitialized = errors.New("not initialized")
	// ErrAlreadyInitialized indicates a resource was initialized twice.
	ErrAlreadyInitialized = errors.New("already initialized")
	// ErrClosed indicates an operation on a closed resource.
	ErrClosed = errors.New("resource is closed")
	// ErrTimeout indicates an operation timed out.
	ErrTimeout = errors.New("operation timed out")
	// ErrCanceled indicates an operation was canceled.
	ErrCanceled = errors.New("operation canceled")
	// ErrNotSupported indicates an unsupported operation.
	ErrNotSupported = errors.New("operation not supported")
)
