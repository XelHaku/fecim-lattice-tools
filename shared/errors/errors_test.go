package errors

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestFeCIMError(t *testing.T) {
	t.Run("basic error creation", func(t *testing.T) {
		err := ConfigError("test error")
		if err == nil {
			t.Fatal("expected non-nil error")
		}
		if err.Category != CategoryConfig {
			t.Errorf("expected CategoryConfig, got %v", err.Category)
		}
		if err.Message != "test error" {
			t.Errorf("expected 'test error', got %q", err.Message)
		}
		if !strings.Contains(err.Error(), "test error") {
			t.Errorf("Error() should contain message, got %q", err.Error())
		}
	})

	t.Run("error with details", func(t *testing.T) {
		err := DataError("invalid format").WithDetails("expected %s, got %s", "JSON", "XML")
		if err.Details != "expected JSON, got XML" {
			t.Errorf("expected formatted details, got %q", err.Details)
		}
		if !strings.Contains(err.Error(), "expected JSON") {
			t.Errorf("Error() should contain details, got %q", err.Error())
		}
	})

	t.Run("error with recovery", func(t *testing.T) {
		err := FileNotFound("/path/to/file")
		if err.Recovery == "" {
			t.Error("FileNotFound should have recovery suggestion")
		}
		msg := err.UserMessage()
		if !strings.Contains(msg, "fix") || !strings.Contains(msg, "To fix") {
			t.Errorf("UserMessage should include recovery, got %q", msg)
		}
	})

	t.Run("error with cause", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := IOError("read failed", cause).WithCause(cause)
		if err.Cause != cause {
			t.Error("cause should be set")
		}
		if !errors.Is(err, cause) {
			t.Error("errors.Is should find cause")
		}
	})

	t.Run("error wrapping", func(t *testing.T) {
		inner := errors.New("inner error")
		outer := Wrap(inner, "operation failed")
		if outer == nil {
			t.Fatal("wrapped error should not be nil")
		}
		if !errors.Is(outer, inner) {
			t.Error("wrapped error should contain inner")
		}
	})

	t.Run("nil wrap returns nil", func(t *testing.T) {
		result := Wrap(nil, "message")
		if result != nil {
			t.Error("Wrap(nil) should return nil")
		}
	})

	t.Run("full message", func(t *testing.T) {
		err := ConfigError("config missing").
			WithDetails("looked in /etc/app/config.yaml").
			WithRecovery("create the config file")
		
		msg := err.FullMessage()
		if !strings.Contains(msg, "Configuration") {
			t.Error("FullMessage should include category")
		}
		if !strings.Contains(msg, "config missing") {
			t.Error("FullMessage should include message")
		}
		if !strings.Contains(msg, "/etc/app") {
			t.Error("FullMessage should include details")
		}
		if !strings.Contains(msg, "Recovery:") {
			t.Error("FullMessage should include recovery")
		}
	})
}

func TestErrorConstructors(t *testing.T) {
	tests := []struct {
		name     string
		err      *FeCIMError
		category Category
	}{
		{"ConfigError", ConfigError("test"), CategoryConfig},
		{"ConfigNotFound", ConfigNotFound("physics"), CategoryConfig},
		{"IOError", IOError("read failed", nil), CategoryIO},
		{"FileNotFound", FileNotFound("/path"), CategoryIO},
		{"DataError", DataError("invalid"), CategoryData},
		{"InvalidData", InvalidData("JSON", "missing field"), CategoryData},
		{"ResourceError", ResourceError("GPU", "out of memory"), CategoryResource},
		{"GPUError", GPUError("shader compile failed", nil), CategoryResource},
		{"InternalError", InternalError("assertion failed"), CategoryInternal},
		{"UserInputError", UserInputError("invalid input"), CategoryUser},
		{"InvalidParameter", InvalidParameter("level", -1, "positive integer"), CategoryUser},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Category != tt.category {
				t.Errorf("expected category %v, got %v", tt.category, tt.err.Category)
			}
			if tt.err.Message == "" {
				t.Error("error should have a message")
			}
		})
	}
}

func TestRecoverability(t *testing.T) {
	t.Run("config errors are recoverable with recovery", func(t *testing.T) {
		err := ConfigError("missing config").WithRecovery("create the config")
		if !err.Recoverable() {
			t.Error("config error with recovery should be recoverable")
		}
	})

	t.Run("internal errors are not recoverable", func(t *testing.T) {
		err := InternalError("nil pointer")
		if err.Recoverable() {
			t.Error("internal errors should not be recoverable")
		}
	})

	t.Run("resource errors are not recoverable", func(t *testing.T) {
		err := ResourceError("Memory", "out of memory")
		if err.Recoverable() {
			t.Error("resource errors should not be recoverable")
		}
	})

	t.Run("IsRecoverable helper", func(t *testing.T) {
		recoverable := ConfigError("x").WithRecovery("do Y")
		if !IsRecoverable(recoverable) {
			t.Error("expected recoverable")
		}

		notRecoverable := InternalError("crash")
		if IsRecoverable(notRecoverable) {
			t.Error("expected not recoverable")
		}
	})
}

func TestErrorExtraction(t *testing.T) {
	t.Run("IsFeCIMError", func(t *testing.T) {
		fErr := ConfigError("test")
		stdErr := errors.New("standard error")

		if !IsFeCIMError(fErr) {
			t.Error("FeCIMError should be detected")
		}
		if IsFeCIMError(stdErr) {
			t.Error("standard error should not be FeCIMError")
		}
	})

	t.Run("GetFeCIMError", func(t *testing.T) {
		fErr := ConfigError("test")
		extracted := GetFeCIMError(fErr)
		if extracted != fErr {
			t.Error("should extract the same error")
		}

		stdErr := errors.New("standard")
		if GetFeCIMError(stdErr) != nil {
			t.Error("should return nil for non-FeCIMError")
		}
	})

	t.Run("GetFeCIMError with wrapped", func(t *testing.T) {
		inner := DataError("corrupt")
		outer := fmt.Errorf("wrapping: %w", inner)
		
		extracted := GetFeCIMError(outer)
		if extracted == nil {
			t.Error("should extract wrapped FeCIMError")
		}
		if extracted.Category != CategoryData {
			t.Error("should preserve category")
		}
	})
}

func TestCategoryString(t *testing.T) {
	tests := []struct {
		cat  Category
		want string
	}{
		{CategoryUnknown, "Unknown"},
		{CategoryConfig, "Configuration"},
		{CategoryIO, "I/O"},
		{CategoryData, "Data"},
		{CategoryResource, "Resource"},
		{CategoryInternal, "Internal"},
		{CategoryUser, "User Input"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.cat.String(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSentinelErrors(t *testing.T) {
	tests := []error{
		ErrNotInitialized,
		ErrAlreadyInitialized,
		ErrClosed,
		ErrTimeout,
		ErrCanceled,
		ErrNotSupported,
	}

	for _, err := range tests {
		t.Run(err.Error(), func(t *testing.T) {
			if err == nil {
				t.Error("sentinel error should not be nil")
			}
			if err.Error() == "" {
				t.Error("sentinel error should have message")
			}
		})
	}
}

func TestLocation(t *testing.T) {
	err := NewError(CategoryConfig, "test")
	if err.Location == "" {
		t.Error("location should be set")
	}
	if !strings.Contains(err.Location, "errors_test.go:") {
		t.Errorf("location should contain file:line, got %q", err.Location)
	}
}

func TestWrapWithCategory(t *testing.T) {
	inner := errors.New("file not found")
	wrapped := WrapWithCategory(inner, CategoryIO, "failed to load config")
	
	if wrapped.Category != CategoryIO {
		t.Errorf("expected CategoryIO, got %v", wrapped.Category)
	}
	if wrapped.Message != "failed to load config" {
		t.Errorf("unexpected message: %q", wrapped.Message)
	}
	if !errors.Is(wrapped, inner) {
		t.Error("should wrap inner error")
	}
}

func TestWrapPreservesCategory(t *testing.T) {
	inner := DataError("invalid JSON")
	outer := Wrap(inner, "parse failed")
	
	if outer.Category != CategoryData {
		t.Errorf("Wrap should preserve category, got %v", outer.Category)
	}
}
