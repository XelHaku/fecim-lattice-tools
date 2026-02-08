// Package errors provides recovery mechanisms for graceful degradation.
package errors

import (
	"fmt"
	"log"
	"runtime/debug"
	"sync"
)

// RecoveryHandler is called when a panic is recovered.
type RecoveryHandler func(panicValue interface{}, stack []byte)

// DefaultRecoveryHandler logs the panic with stack trace.
var DefaultRecoveryHandler RecoveryHandler = func(panicValue interface{}, stack []byte) {
	log.Printf("[PANIC RECOVERED] %v\n%s", panicValue, stack)
}

var (
	globalHandler   RecoveryHandler = DefaultRecoveryHandler
	globalHandlerMu sync.RWMutex
)

// SetRecoveryHandler sets the global panic recovery handler.
func SetRecoveryHandler(h RecoveryHandler) {
	globalHandlerMu.Lock()
	defer globalHandlerMu.Unlock()
	if h == nil {
		globalHandler = DefaultRecoveryHandler
	} else {
		globalHandler = h
	}
}

// getRecoveryHandler returns the current global handler.
func getRecoveryHandler() RecoveryHandler {
	globalHandlerMu.RLock()
	defer globalHandlerMu.RUnlock()
	return globalHandler
}

// RecoverFunc runs a function with panic recovery, returning any error.
// If the function panics, the panic is caught and returned as an error.
func RecoverFunc(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			getRecoveryHandler()(r, stack)
			err = InternalError(fmt.Sprintf("panic recovered: %v", r)).
				WithDetails("Stack trace available in logs")
		}
	}()
	return fn()
}

// RecoverFuncValue runs a function that returns a value with panic recovery.
func RecoverFuncValue[T any](fn func() (T, error)) (result T, err error) {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			getRecoveryHandler()(r, stack)
			err = InternalError(fmt.Sprintf("panic recovered: %v", r)).
				WithDetails("Stack trace available in logs")
		}
	}()
	return fn()
}

// RecoverAction runs an action that might panic and converts panics to errors.
// Unlike RecoverFunc, this is for actions that don't return errors normally.
func RecoverAction(fn func()) error {
	return RecoverFunc(func() error {
		fn()
		return nil
	})
}

// MustRecover is similar to RecoverAction but takes a name for logging.
func MustRecover(name string, fn func()) error {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			log.Printf("[PANIC in %s] %v\n%s", name, r, stack)
		}
	}()
	fn()
	return nil
}

// SafeClose closes a resource, ignoring any panic.
// Useful for cleanup in defer statements.
func SafeClose(closer interface{ Close() error }) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[WARN] Panic during close: %v", r)
		}
	}()
	if closer != nil {
		_ = closer.Close()
	}
}

// SafeCloseAll closes multiple resources, collecting all errors.
func SafeCloseAll(closers ...interface{ Close() error }) []error {
	var errs []error
	for _, c := range closers {
		if c == nil {
			continue
		}
		func() {
			defer func() {
				if r := recover(); r != nil {
					errs = append(errs, fmt.Errorf("panic during close: %v", r))
				}
			}()
			if err := c.Close(); err != nil {
				errs = append(errs, err)
			}
		}()
	}
	return errs
}

// Retry executes a function with retries and exponential backoff.
type RetryConfig struct {
	MaxAttempts int           // Maximum number of attempts (default: 3)
	InitialWait int           // Initial wait in milliseconds (default: 100)
	MaxWait     int           // Maximum wait in milliseconds (default: 5000)
	Multiplier  float64       // Backoff multiplier (default: 2.0)
	RetryOn     func(error) bool // Function to determine if error is retryable
}

// DefaultRetryConfig returns a sensible default retry configuration.
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts: 3,
		InitialWait: 100,
		MaxWait:     5000,
		Multiplier:  2.0,
		RetryOn: func(err error) bool {
			// By default, only retry recoverable errors
			return IsRecoverable(err)
		},
	}
}

// Retry executes a function with the configured retry policy.
// Note: For actual use, import time package and implement proper sleep.
// This is a simplified version for demonstration.
func (c *RetryConfig) Retry(fn func() error) error {
	var lastErr error
	
	for attempt := 1; attempt <= c.MaxAttempts; attempt++ {
		err := RecoverFunc(fn)
		if err == nil {
			return nil
		}
		lastErr = err
		
		// Check if we should retry
		if c.RetryOn != nil && !c.RetryOn(err) {
			return err
		}
		
		// Don't sleep on last attempt
		if attempt < c.MaxAttempts {
			// In real implementation, would use time.Sleep here
			// For now, this is a template
		}
	}
	
	return Wrap(lastErr, fmt.Sprintf("failed after %d attempts", c.MaxAttempts))
}

// GracefulDegradation provides fallback behavior when operations fail.
type GracefulDegradation[T any] struct {
	// Primary is the main operation to try.
	Primary func() (T, error)
	// Fallbacks are alternative operations if primary fails.
	Fallbacks []func() (T, error)
	// Default is returned if all operations fail.
	Default T
	// OnFallback is called when falling back to an alternative.
	OnFallback func(attempt int, err error)
}

// Execute runs the primary operation, falling back to alternatives on failure.
func (g *GracefulDegradation[T]) Execute() (T, error) {
	// Try primary
	result, err := RecoverFuncValue(g.Primary)
	if err == nil {
		return result, nil
	}
	
	// Try fallbacks
	for i, fb := range g.Fallbacks {
		if g.OnFallback != nil {
			g.OnFallback(i+1, err)
		}
		
		result, err = RecoverFuncValue(fb)
		if err == nil {
			return result, nil
		}
	}
	
	// Return default
	return g.Default, err
}

// ErrorCollector collects multiple errors during batch operations.
type ErrorCollector struct {
	errors []error
	mu     sync.Mutex
}

// NewErrorCollector creates a new error collector.
func NewErrorCollector() *ErrorCollector {
	return &ErrorCollector{}
}

// Add adds an error to the collection (ignores nil errors).
func (c *ErrorCollector) Add(err error) {
	if err == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.errors = append(c.errors, err)
}

// AddFromRecover adds a panic recovery as an error.
func (c *ErrorCollector) AddFromRecover(name string, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			c.Add(InternalError(fmt.Sprintf("%s panicked: %v", name, r)))
		}
	}()
	fn()
}

// HasErrors returns true if any errors were collected.
func (c *ErrorCollector) HasErrors() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.errors) > 0
}

// Count returns the number of errors collected.
func (c *ErrorCollector) Count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.errors)
}

// Errors returns all collected errors.
func (c *ErrorCollector) Errors() []error {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]error, len(c.errors))
	copy(result, c.errors)
	return result
}

// Combined returns a single error summarizing all collected errors.
func (c *ErrorCollector) Combined() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if len(c.errors) == 0 {
		return nil
	}
	if len(c.errors) == 1 {
		return c.errors[0]
	}
	
	var sb fmt.Stringer = &errorSummary{errors: c.errors}
	return fmt.Errorf("%d errors occurred: %s", len(c.errors), sb)
}

type errorSummary struct {
	errors []error
}

func (s *errorSummary) String() string {
	if len(s.errors) == 0 {
		return ""
	}
	if len(s.errors) == 1 {
		return s.errors[0].Error()
	}
	result := s.errors[0].Error()
	for i := 1; i < len(s.errors) && i < 3; i++ {
		result += "; " + s.errors[i].Error()
	}
	if len(s.errors) > 3 {
		result += fmt.Sprintf(" (and %d more)", len(s.errors)-3)
	}
	return result
}
