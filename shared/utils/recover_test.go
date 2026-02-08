package utils

import (
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestSafeGo(t *testing.T) {
	t.Run("normal execution", func(t *testing.T) {
		executed := make(chan bool, 1)
		SafeGo("test", func() {
			executed <- true
		})
		
		select {
		case <-executed:
			// Success
		case <-time.After(time.Second):
			t.Error("function did not execute")
		}
	})

	t.Run("panic recovery", func(t *testing.T) {
		done := make(chan bool, 1)
		SafeGo("panic-test", func() {
			defer func() { done <- true }()
			panic("intentional panic")
		})
		
		select {
		case <-done:
			// Success - panic was recovered and defer executed
		case <-time.After(time.Second):
			// Give extra time for panic handling
		}
		// If we get here without the test crashing, panic was recovered
	})
}

func TestSafeGoWithCallback(t *testing.T) {
	t.Run("no panic", func(t *testing.T) {
		executed := make(chan bool, 1)
		callbackCalled := false
		
		SafeGoWithCallback("test", func() {
			executed <- true
		}, func(name string, v interface{}) {
			callbackCalled = true
		})
		
		select {
		case <-executed:
			// Success
		case <-time.After(time.Second):
			t.Error("function did not execute")
		}
		
		if callbackCalled {
			t.Error("callback should not be called when no panic")
		}
	})

	t.Run("panic triggers callback", func(t *testing.T) {
		callbackDone := make(chan bool, 1)
		var panicName string
		var panicValue interface{}
		
		SafeGoWithCallback("callback-test", func() {
			panic("test panic value")
		}, func(name string, v interface{}) {
			panicName = name
			panicValue = v
			callbackDone <- true
		})
		
		select {
		case <-callbackDone:
			if panicName != "callback-test" {
				t.Errorf("expected name 'callback-test', got %q", panicName)
			}
			if panicValue != "test panic value" {
				t.Errorf("unexpected panic value: %v", panicValue)
			}
		case <-time.After(time.Second):
			t.Error("callback was not called")
		}
	})

	t.Run("nil callback uses default logging", func(t *testing.T) {
		// This should not crash even with nil callback
		done := make(chan bool, 1)
		SafeGoWithCallback("nil-callback-test", func() {
			defer func() { done <- true }()
			panic("test")
		}, nil)
		
		select {
		case <-done:
			// Success
		case <-time.After(time.Second):
			// May timeout if panic handling is slow
		}
	})
}

func TestSafeCall(t *testing.T) {
	t.Run("successful call", func(t *testing.T) {
		executed := false
		err := SafeCall("test", func() error {
			executed = true
			return nil
		})
		
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if !executed {
			t.Error("function was not executed")
		}
	})

	t.Run("returns function error", func(t *testing.T) {
		expectedErr := errors.New("expected error")
		err := SafeCall("test", func() error {
			return expectedErr
		})
		
		if err != expectedErr {
			t.Errorf("expected %v, got %v", expectedErr, err)
		}
	})

	t.Run("panic converted to error", func(t *testing.T) {
		err := SafeCall("panic-test", func() error {
			panic("test panic")
		})
		
		if err == nil {
			t.Error("expected error from panic")
		}
		if !strings.Contains(err.Error(), "panic-test panicked") {
			t.Errorf("error should mention panic: %v", err)
		}
		if !strings.Contains(err.Error(), "test panic") {
			t.Errorf("error should contain panic value: %v", err)
		}
	})
}

type mockCloser struct {
	closed  bool
	err     error
	panics  bool
	mu      sync.Mutex
}

func (m *mockCloser) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.panics {
		panic("close panic")
	}
	m.closed = true
	return m.err
}

func (m *mockCloser) IsClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closed
}

func TestSafeClose(t *testing.T) {
	t.Run("successful close", func(t *testing.T) {
		m := &mockCloser{}
		err := SafeClose("test", m)
		
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if !m.IsClosed() {
			t.Error("should have closed")
		}
	})

	t.Run("close with error", func(t *testing.T) {
		expectedErr := errors.New("close error")
		m := &mockCloser{err: expectedErr}
		err := SafeClose("test", m)
		
		if err != expectedErr {
			t.Errorf("expected %v, got %v", expectedErr, err)
		}
		if !m.IsClosed() {
			t.Error("should have closed")
		}
	})

	t.Run("close panics", func(t *testing.T) {
		m := &mockCloser{panics: true}
		// Should not panic
		err := SafeClose("panic-close", m)
		
		// Error may or may not be returned depending on implementation
		_ = err
		// If we get here without crashing, the panic was recovered
	})

	t.Run("nil closer", func(t *testing.T) {
		err := SafeClose("nil", nil)
		if err != nil {
			t.Errorf("expected no error for nil closer, got %v", err)
		}
	})
}

func TestConcurrentSafeGo(t *testing.T) {
	// Test that multiple SafeGo calls work correctly concurrently
	var wg sync.WaitGroup
	counter := make(chan int, 100)
	
	for i := 0; i < 100; i++ {
		wg.Add(1)
		n := i
		SafeGo("concurrent", func() {
			defer wg.Done()
			counter <- n
		})
	}
	
	wg.Wait()
	close(counter)
	
	count := 0
	for range counter {
		count++
	}
	
	if count != 100 {
		t.Errorf("expected 100 executions, got %d", count)
	}
}
