//go:build !ci
// +build !ci

package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	autoXvfbEnvVar       = "FECIM_AUTO_XVFB"
	autoXvfbScreen       = "1920x1080x24"
	autoXvfbDisplayStart = 99
	autoXvfbDisplayEnd   = 120
)

var (
	autoXvfbOnce sync.Once
	autoXvfbErr  error
	autoXvfbSess *testXvfbSession
)

type testXvfbSession struct {
	display string
	cmd     *exec.Cmd
	waitCh  chan error
	stderr  bytes.Buffer
}

func (s *testXvfbSession) stop() {
	if s == nil || s.cmd == nil || s.cmd.Process == nil {
		return
	}
	_ = s.cmd.Process.Kill()
	if s.waitCh == nil {
		return
	}
	select {
	case <-s.waitCh:
	case <-time.After(2 * time.Second):
	}
}

func ensureDisplayForGraphicalTests(t *testing.T) {
	t.Helper()

	if hasGraphicalSessionForTests() {
		return
	}
	if !autoXvfbEnabledForTests() {
		t.Skipf("DISPLAY is not set (set %s=1 or run under xvfb-run -a)", autoXvfbEnvVar)
	}

	autoXvfbOnce.Do(func() {
		autoXvfbSess, autoXvfbErr = startAutoXvfbForTests(autoXvfbScreen, autoXvfbDisplayStart, autoXvfbDisplayEnd)
		if autoXvfbErr != nil {
			return
		}
		display := autoXvfbSess.display
		if err := os.Setenv("DISPLAY", display); err != nil {
			autoXvfbSess.stop()
			autoXvfbSess = nil
			autoXvfbErr = fmt.Errorf("failed to set DISPLAY=%s: %w", display, err)
			return
		}
		_ = os.Unsetenv("WAYLAND_DISPLAY")
	})

	if autoXvfbErr != nil {
		t.Skipf("DISPLAY is not set and auto-Xvfb failed: %v", autoXvfbErr)
		return
	}
	if display := strings.TrimSpace(os.Getenv("DISPLAY")); display != "" {
		t.Logf("Auto-Xvfb display active: %s", display)
	}
}

func shutdownAutoXvfbForTests() {
	if autoXvfbSess == nil {
		return
	}
	autoXvfbSess.stop()
	autoXvfbSess = nil
}

func hasGraphicalSessionForTests() bool {
	return strings.TrimSpace(os.Getenv("DISPLAY")) != "" || strings.TrimSpace(os.Getenv("WAYLAND_DISPLAY")) != ""
}

func autoXvfbEnabledForTests() bool {
	raw := strings.TrimSpace(os.Getenv(autoXvfbEnvVar))
	if raw == "" {
		return true
	}
	switch strings.ToLower(raw) {
	case "0", "false", "no", "off":
		return false
	default:
		return true
	}
}

func startAutoXvfbForTests(screen string, displayStart, displayEnd int) (*testXvfbSession, error) {
	xvfbBin, err := exec.LookPath("Xvfb")
	if err != nil {
		return nil, fmt.Errorf("Xvfb binary not found in PATH: %w", err)
	}
	if displayStart > displayEnd {
		displayStart, displayEnd = displayEnd, displayStart
	}
	if displayStart <= 0 {
		displayStart = autoXvfbDisplayStart
	}
	if displayEnd < displayStart {
		displayEnd = displayStart
	}

	var launchErrors []string
	for n := displayStart; n <= displayEnd; n++ {
		display := fmt.Sprintf(":%d", n)
		sess, err := launchAutoXvfbForTests(xvfbBin, display, screen)
		if err != nil {
			launchErrors = append(launchErrors, fmt.Sprintf("%s: %v", display, err))
			continue
		}
		if waitForDisplayForTests(display, 4*time.Second) {
			return sess, nil
		}
		stderr := strings.TrimSpace(sess.stderr.String())
		if stderr == "" {
			stderr = "display did not become ready"
		}
		launchErrors = append(launchErrors, fmt.Sprintf("%s: %s", display, stderr))
		sess.stop()
	}

	return nil, fmt.Errorf("unable to start Xvfb on displays %d-%d (%s)", displayStart, displayEnd, strings.Join(launchErrors, "; "))
}

func launchAutoXvfbForTests(xvfbBin, display, screen string) (*testXvfbSession, error) {
	sess := &testXvfbSession{display: display, waitCh: make(chan error, 1)}
	sess.cmd = exec.Command(xvfbBin, display, "-screen", "0", screen, "-nolisten", "tcp", "+extension", "GLX", "+render", "-ac")
	sess.cmd.Stdout = io.Discard
	sess.cmd.Stderr = &sess.stderr
	if err := sess.cmd.Start(); err != nil {
		return nil, err
	}
	go func() {
		sess.waitCh <- sess.cmd.Wait()
	}()
	return sess, nil
}

func waitForDisplayForTests(display string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if displayReadyForTests(display) {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return displayReadyForTests(display)
}

func displayReadyForTests(display string) bool {
	if xdpyinfoBin, err := exec.LookPath("xdpyinfo"); err == nil {
		cmd := exec.Command(xdpyinfoBin, "-display", display)
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
		return cmd.Run() == nil
	}

	displayNum, ok := parseDisplayNumberForTests(display)
	if !ok {
		return false
	}
	socket := filepath.Join("/tmp/.X11-unix", "X"+strconv.Itoa(displayNum))
	if _, err := os.Stat(socket); err != nil {
		return false
	}
	return true
}

func parseDisplayNumberForTests(display string) (int, bool) {
	trimmed := strings.TrimSpace(display)
	if !strings.HasPrefix(trimmed, ":") {
		return 0, false
	}
	num, err := strconv.Atoi(strings.TrimPrefix(trimmed, ":"))
	if err != nil {
		return 0, false
	}
	return num, true
}
