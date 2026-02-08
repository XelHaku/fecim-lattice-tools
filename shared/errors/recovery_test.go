package errors

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
)

func TestRecoverFunc(t *testing.T) {
	t.Run("successful function", func(t *testing.T) {
		err := RecoverFunc(func() error {
			return nil
		})
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	t.Run("function returns error", func(t *testing.T) {
		expectedErr := errors.New("expected error")
		err := RecoverFunc(func() error {
			return expectedErr
		})
		if err != expectedErr {
			t.Errorf("expected %v, got %v", expectedErr, err)
		}
	})

	t.Run("function panics", func(t *testing.T) {
		err := RecoverFunc(func() error {
			panic("test panic")
		})
		if err == nil {
			t.Fatal("expected error from panic")
		}
		if !strings.Contains(err.Error(), "panic recovered") {
			t.Errorf("error should mention panic, got %q", err.Error())
		}
	})

	t.Run("function panics with error", func(t *testing.T) {
		err := RecoverFunc(func() error {
			panic(errors.New("panic error"))
		})
		if err == nil {
			t.Fatal("expected error from panic")
		}
		if !strings.Contains(err.Error(), "panic recovered") {
			t.Errorf("error should mention panic, got %q", err.Error())
		}
	})
}

func TestRecoverFuncValue(t *testing.T) {
	t.Run("successful function", func(t *testing.T) {
		result, err := RecoverFuncValue(func() (int, error) {
			return 42, nil
		})
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
		if result != 42 {
			t.Errorf("expected 42, got %d", result)
		}
	})

	t.Run("function panics", func(t *testing.T) {
		result, err := RecoverFuncValue(func() (int, error) {
			panic("test panic")
		})
		if err == nil {
			t.Fatal("expected error from panic")
		}
		if result != 0 {
			t.Errorf("expected zero value, got %d", result)
		}
	})
}

func TestRecoverAction(t *testing.T) {
	t.Run("successful action", func(t *testing.T) {
		executed := false
		err := RecoverAction(func() {
			executed = true
		})
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
		if !executed {
			t.Error("action should have executed")
		}
	})

	t.Run("panicking action", func(t *testing.T) {
		err := RecoverAction(func() {
			panic("test panic")
		})
		if err == nil {
			t.Fatal("expected error from panic")
		}
	})
}

func TestMustRecover(t *testing.T) {
	t.Run("successful action", func(t *testing.T) {
		err := MustRecover("test", func() {
			// do nothing
		})
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	t.Run("panicking action", func(t *testing.T) {
		// This should not panic - it should be recovered
		err := MustRecover("test-panic", func() {
			panic("intentional panic")
		})
		// Note: MustRecover currently doesn't return the error, just logs it
		// This test verifies it doesn't propagate the panic
		_ = err
	})
}

type mockCloser struct {
	closed  bool
	err     error
	panics  bool
}

func (m *mockCloser) Close() error {
	if m.panics {
		panic("close panic")
	}
	m.closed = true
	return m.err
}

func TestSafeClose(t *testing.T) {
	t.Run("successful close", func(t *testing.T) {
		m := &mockCloser{}
		SafeClose(m)
		if !m.closed {
			t.Error("should have closed")
		}
	})

	t.Run("close with error", func(t *testing.T) {
		m := &mockCloser{err: errors.New("close error")}
		SafeClose(m) // Should not panic
		if !m.closed {
			t.Error("should have closed")
		}
	})

	t.Run("close panics", func(t *testing.T) {
		m := &mockCloser{panics: true}
		SafeClose(m) // Should not panic
	})

	t.Run("nil closer", func(t *testing.T) {
		SafeClose(nil) // Should not panic
	})
}

func TestSafeCloseAll(t *testing.T) {
	t.Run("all successful", func(t *testing.T) {
		m1 := &mockCloser{}
		m2 := &mockCloser{}
		errs := SafeCloseAll(m1, m2)
		if len(errs) != 0 {
			t.Errorf("expected no errors, got %v", errs)
		}
		if !m1.closed || !m2.closed {
			t.Error("both should be closed")
		}
	})

	t.Run("some errors", func(t *testing.T) {
		m1 := &mockCloser{}
		m2 := &mockCloser{err: errors.New("close error")}
		errs := SafeCloseAll(m1, m2)
		if len(errs) != 1 {
			t.Errorf("expected 1 error, got %d", len(errs))
		}
	})

	t.Run("some panics", func(t *testing.T) {
		m1 := &mockCloser{}
		m2 := &mockCloser{panics: true}
		m3 := &mockCloser{}
		errs := SafeCloseAll(m1, m2, m3)
		if len(errs) != 1 {
			t.Errorf("expected 1 error from panic, got %d", len(errs))
		}
		if !m1.closed || !m3.closed {
			t.Error("non-panicking closers should still close")
		}
	})

	t.Run("with nil", func(t *testing.T) {
		m1 := &mockCloser{}
		errs := SafeCloseAll(m1, nil)
		if len(errs) != 0 {
			t.Errorf("expected no errors, got %v", errs)
		}
	})
}

func TestErrorCollector(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		c := NewErrorCollector()
		if c.HasErrors() {
			t.Error("should not have errors")
		}
		if c.Count() != 0 {
			t.Errorf("expected count 0, got %d", c.Count())
		}
		if c.Combined() != nil {
			t.Error("Combined should return nil")
		}
	})

	t.Run("add nil", func(t *testing.T) {
		c := NewErrorCollector()
		c.Add(nil)
		if c.HasErrors() {
			t.Error("nil should not be added")
		}
	})

	t.Run("add errors", func(t *testing.T) {
		c := NewErrorCollector()
		c.Add(errors.New("error 1"))
		c.Add(errors.New("error 2"))
		
		if !c.HasErrors() {
			t.Error("should have errors")
		}
		if c.Count() != 2 {
			t.Errorf("expected count 2, got %d", c.Count())
		}
		errs := c.Errors()
		if len(errs) != 2 {
			t.Errorf("expected 2 errors, got %d", len(errs))
		}
	})

	t.Run("combined single error", func(t *testing.T) {
		c := NewErrorCollector()
		orig := errors.New("single error")
		c.Add(orig)
		
		combined := c.Combined()
		if combined != orig {
			t.Error("single error should return original")
		}
	})

	t.Run("combined multiple errors", func(t *testing.T) {
		c := NewErrorCollector()
		c.Add(errors.New("error 1"))
		c.Add(errors.New("error 2"))
		c.Add(errors.New("error 3"))
		
		combined := c.Combined()
		if combined == nil {
			t.Fatal("Combined should not return nil")
		}
		msg := combined.Error()
		if !strings.Contains(msg, "3 errors") {
			t.Errorf("should mention error count, got %q", msg)
		}
	})

	t.Run("concurrent adds", func(t *testing.T) {
		c := NewErrorCollector()
		var wg sync.WaitGroup
		
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				c.Add(fmt.Errorf("error %d", n))
			}(i)
		}
		
		wg.Wait()
		if c.Count() != 100 {
			t.Errorf("expected 100 errors, got %d", c.Count())
		}
	})

	t.Run("AddFromRecover", func(t *testing.T) {
		c := NewErrorCollector()
		c.AddFromRecover("test", func() {
			panic("test panic")
		})
		
		if !c.HasErrors() {
			t.Error("should have captured panic")
		}
		errs := c.Errors()
		if len(errs) != 1 {
			t.Fatalf("expected 1 error, got %d", len(errs))
		}
		if !strings.Contains(errs[0].Error(), "test panicked") {
			t.Errorf("error should mention panic, got %q", errs[0].Error())
		}
	})
}

func TestGracefulDegradation(t *testing.T) {
	t.Run("primary succeeds", func(t *testing.T) {
		g := &GracefulDegradation[int]{
			Primary: func() (int, error) {
				return 42, nil
			},
			Default: -1,
		}
		
		result, err := g.Execute()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != 42 {
			t.Errorf("expected 42, got %d", result)
		}
	})

	t.Run("fallback to first alternative", func(t *testing.T) {
		g := &GracefulDegradation[int]{
			Primary: func() (int, error) {
				return 0, errors.New("primary failed")
			},
			Fallbacks: []func() (int, error){
				func() (int, error) {
					return 100, nil
				},
			},
			Default: -1,
		}
		
		result, err := g.Execute()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != 100 {
			t.Errorf("expected 100, got %d", result)
		}
	})

	t.Run("fallback to second alternative", func(t *testing.T) {
		g := &GracefulDegradation[int]{
			Primary: func() (int, error) {
				return 0, errors.New("primary failed")
			},
			Fallbacks: []func() (int, error){
				func() (int, error) {
					return 0, errors.New("first fallback failed")
				},
				func() (int, error) {
					return 200, nil
				},
			},
			Default: -1,
		}
		
		result, err := g.Execute()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != 200 {
			t.Errorf("expected 200, got %d", result)
		}
	})

	t.Run("all fail returns default", func(t *testing.T) {
		g := &GracefulDegradation[int]{
			Primary: func() (int, error) {
				return 0, errors.New("primary failed")
			},
			Fallbacks: []func() (int, error){
				func() (int, error) {
					return 0, errors.New("fallback failed")
				},
			},
			Default: -1,
		}
		
		result, err := g.Execute()
		if err == nil {
			t.Error("expected error when all fail")
		}
		if result != -1 {
			t.Errorf("expected default -1, got %d", result)
		}
	})

	t.Run("primary panics triggers fallback", func(t *testing.T) {
		g := &GracefulDegradation[int]{
			Primary: func() (int, error) {
				panic("primary panic")
			},
			Fallbacks: []func() (int, error){
				func() (int, error) {
					return 300, nil
				},
			},
			Default: -1,
		}
		
		result, err := g.Execute()
		if err != nil {
			t.Errorf("expected no error after fallback, got %v", err)
		}
		if result != 300 {
			t.Errorf("expected 300, got %d", result)
		}
	})

	t.Run("onFallback callback", func(t *testing.T) {
		var callbackCalled bool
		var callbackAttempt int
		
		g := &GracefulDegradation[int]{
			Primary: func() (int, error) {
				return 0, errors.New("failed")
			},
			Fallbacks: []func() (int, error){
				func() (int, error) {
					return 42, nil
				},
			},
			OnFallback: func(attempt int, err error) {
				callbackCalled = true
				callbackAttempt = attempt
			},
		}
		
		g.Execute()
		if !callbackCalled {
			t.Error("OnFallback should be called")
		}
		if callbackAttempt != 1 {
			t.Errorf("expected attempt 1, got %d", callbackAttempt)
		}
	})
}

func TestSetRecoveryHandler(t *testing.T) {
	t.Run("custom handler", func(t *testing.T) {
		var called bool
		SetRecoveryHandler(func(panicValue interface{}, stack []byte) {
			called = true
		})
		defer SetRecoveryHandler(nil) // Reset to default
		
		_ = RecoverFunc(func() error {
			panic("test")
		})
		
		if !called {
			t.Error("custom handler should be called")
		}
	})

	t.Run("nil handler resets to default", func(t *testing.T) {
		SetRecoveryHandler(nil)
		// Should not panic
		_ = RecoverFunc(func() error {
			panic("test")
		})
	})
}

func TestDefaultRetryConfig(t *testing.T) {
	cfg := DefaultRetryConfig()
	if cfg.MaxAttempts != 3 {
		t.Errorf("expected 3 attempts, got %d", cfg.MaxAttempts)
	}
	if cfg.InitialWait != 100 {
		t.Errorf("expected 100ms initial wait, got %d", cfg.InitialWait)
	}
	if cfg.Multiplier != 2.0 {
		t.Errorf("expected 2.0 multiplier, got %f", cfg.Multiplier)
	}
}

func TestRetry(t *testing.T) {
	t.Run("succeeds first try", func(t *testing.T) {
		cfg := DefaultRetryConfig()
		attempts := 0
		
		err := cfg.Retry(func() error {
			attempts++
			return nil
		})
		
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if attempts != 1 {
			t.Errorf("expected 1 attempt, got %d", attempts)
		}
	})

	t.Run("retries on recoverable error", func(t *testing.T) {
		cfg := DefaultRetryConfig()
		cfg.RetryOn = func(err error) bool { return true }
		attempts := 0
		
		err := cfg.Retry(func() error {
			attempts++
			if attempts < 3 {
				return errors.New("temporary error")
			}
			return nil
		})
		
		if err != nil {
			t.Errorf("expected eventual success, got %v", err)
		}
		if attempts != 3 {
			t.Errorf("expected 3 attempts, got %d", attempts)
		}
	})

	t.Run("fails after max attempts", func(t *testing.T) {
		cfg := DefaultRetryConfig()
		cfg.MaxAttempts = 2
		cfg.RetryOn = func(err error) bool { return true }
		attempts := 0
		
		err := cfg.Retry(func() error {
			attempts++
			return errors.New("persistent error")
		})
		
		if err == nil {
			t.Error("expected error after max attempts")
		}
		if attempts != 2 {
			t.Errorf("expected 2 attempts, got %d", attempts)
		}
		if !strings.Contains(err.Error(), "2 attempts") {
			t.Errorf("error should mention attempts, got %q", err.Error())
		}
	})

	t.Run("no retry on non-retryable error", func(t *testing.T) {
		cfg := DefaultRetryConfig()
		cfg.RetryOn = func(err error) bool { return false }
		attempts := 0
		
		err := cfg.Retry(func() error {
			attempts++
			return errors.New("fatal error")
		})
		
		if err == nil {
			t.Error("expected error")
		}
		if attempts != 1 {
			t.Errorf("expected 1 attempt (no retry), got %d", attempts)
		}
	})
}
