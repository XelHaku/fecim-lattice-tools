package render

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestConfigValidate(t *testing.T) {
	cfg := DefaultConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("default config should validate: %v", err)
	}

	cfg.Width = 0
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected invalid width error")
	}

	cfg = DefaultConfig()
	cfg.TargetFPS = 0
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected invalid fps error")
	}
}

func TestRendererInitializeValidateAndState(t *testing.T) {
	r := NewRenderer(DefaultConfig())
	if err := r.Initialize(); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}
	if r.IsRunning() {
		t.Fatalf("renderer should not be running after init")
	}

	bad := NewRenderer(&Config{Width: -1, Height: 720, TargetFPS: 60})
	if err := bad.Initialize(); err == nil {
		t.Fatalf("expected init to fail for invalid config")
	}
}

func TestRendererRunLifecycle(t *testing.T) {
	r := NewRenderer(DefaultConfig())
	if err := r.Initialize(); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}

	var calls atomic.Int32
	r.SetUpdateCallback(func() {
		if calls.Add(1) >= 2 {
			r.Stop()
		}
	})

	errCh := make(chan error, 1)
	go func() {
		errCh <- r.Run()
	}()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("run failed: %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("renderer did not stop in time")
	}

	if calls.Load() < 2 {
		t.Fatalf("expected callback to run at least twice, got %d", calls.Load())
	}
	if r.IsRunning() {
		t.Fatalf("renderer should be stopped")
	}
}

func TestRendererRunRequiresInitialize(t *testing.T) {
	r := NewRenderer(DefaultConfig())
	err := r.Run()
	if !errors.Is(err, ErrRendererNotInitialized) {
		t.Fatalf("expected ErrRendererNotInitialized, got %v", err)
	}
}
