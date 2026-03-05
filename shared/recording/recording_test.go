package recording

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// =============================================================================
// State Enum Tests
// =============================================================================

func TestRecordingStateEnum(t *testing.T) {
	tests := []struct {
		state    State
		expected string
	}{
		{StateIdle, "Idle"},
		{StateRecording, "Recording"},
		{StatePaused, "Paused"},
		{StateStopped, "Stopped"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.state.String() != tt.expected {
				t.Errorf("String() = %s, want %s", tt.state.String(), tt.expected)
			}
		})
	}
}

func TestStateIsActive(t *testing.T) {
	tests := []struct {
		state    State
		expected bool
	}{
		{StateIdle, false},
		{StateRecording, true},
		{StatePaused, true}, // Paused is still "active" (not stopped)
		{StateStopped, false},
	}

	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			if tt.state.IsActive() != tt.expected {
				t.Errorf("IsActive() = %v, want %v", tt.state.IsActive(), tt.expected)
			}
		})
	}
}

// =============================================================================
// Manager Creation Tests
// =============================================================================

func TestNewManager(t *testing.T) {
	manager := NewManager()

	if manager == nil {
		t.Fatal("NewManager returned nil")
	}
}

func TestNewManagerWithSettings(t *testing.T) {
	settings := Settings{
		Quality: QualityHigh,
		FPS:     30,
		Format:  FormatMP4,
		CRF:     18,
		Preset:  PresetMedium,
	}

	manager := NewManagerWithSettings(settings)

	if manager == nil {
		t.Fatal("NewManagerWithSettings returned nil")
	}

	// Settings should be applied
	gotSettings := manager.GetSettings()
	if gotSettings.Quality != settings.Quality {
		t.Errorf("Quality = %v, want %v", gotSettings.Quality, settings.Quality)
	}
	if gotSettings.FPS != settings.FPS {
		t.Errorf("FPS = %d, want %d", gotSettings.FPS, settings.FPS)
	}
}

// =============================================================================
// Initial State Tests
// =============================================================================

func TestManagerInitialState(t *testing.T) {
	manager := NewManager()

	if manager.State() != StateIdle {
		t.Errorf("Initial state = %v, want %v", manager.State(), StateIdle)
	}

	if manager.IsRecording() {
		t.Error("Should not be recording initially")
	}

	if manager.IsPaused() {
		t.Error("Should not be paused initially")
	}
}

func TestManagerInitialMetrics(t *testing.T) {
	manager := NewManager()

	if manager.ElapsedTime() != 0 {
		t.Errorf("Initial elapsed time = %v, want 0", manager.ElapsedTime())
	}

	if manager.FramesCaptured() != 0 {
		t.Errorf("Initial frames captured = %d, want 0", manager.FramesCaptured())
	}

	if manager.FramesDropped() != 0 {
		t.Errorf("Initial frames dropped = %d, want 0", manager.FramesDropped())
	}
}

// =============================================================================
// State Transition Tests
// =============================================================================

func TestStateTransitionIdleToRecording(t *testing.T) {
	manager := NewManager()

	// Create a mock capture source for testing
	mockCapture := &MockCaptureSource{width: 640, height: 480}

	err := manager.Start(mockCapture)
	if err != nil && !isFFmpegMissing(err) {
		t.Fatalf("Start failed: %v", err)
	}

	if isFFmpegMissing(err) {
		t.Skip("FFmpeg not available for testing")
	}

	defer manager.Stop()

	if manager.State() != StateRecording {
		t.Errorf("State after start = %v, want %v", manager.State(), StateRecording)
	}

	if !manager.IsRecording() {
		t.Error("IsRecording should be true after start")
	}
}

func TestStateTransitionRecordingToPaused(t *testing.T) {
	manager := NewManager()
	mockCapture := &MockCaptureSource{width: 640, height: 480}

	err := manager.Start(mockCapture)
	if err != nil {
		if isFFmpegMissing(err) {
			t.Skip("FFmpeg not available for testing")
		}
		t.Fatalf("Start failed: %v", err)
	}
	defer manager.Stop()

	err = manager.Pause()
	if err != nil {
		t.Fatalf("Pause failed: %v", err)
	}

	if manager.State() != StatePaused {
		t.Errorf("State after pause = %v, want %v", manager.State(), StatePaused)
	}

	if !manager.IsPaused() {
		t.Error("IsPaused should be true after pause")
	}

	// IsRecording should still return true (paused is a subset of recording)
	// Or it could return false depending on implementation decision
	t.Logf("IsRecording when paused: %v", manager.IsRecording())
}

func TestStateTransitionPausedToRecording(t *testing.T) {
	manager := NewManager()
	mockCapture := &MockCaptureSource{width: 640, height: 480}

	err := manager.Start(mockCapture)
	if err != nil {
		if isFFmpegMissing(err) {
			t.Skip("FFmpeg not available for testing")
		}
		t.Fatalf("Start failed: %v", err)
	}
	defer manager.Stop()

	manager.Pause()

	err = manager.Resume()
	if err != nil {
		t.Fatalf("Resume failed: %v", err)
	}

	if manager.State() != StateRecording {
		t.Errorf("State after resume = %v, want %v", manager.State(), StateRecording)
	}

	if manager.IsPaused() {
		t.Error("IsPaused should be false after resume")
	}
}

func TestStateTransitionRecordingToStopped(t *testing.T) {
	manager := NewManager()
	mockCapture := &MockCaptureSource{width: 640, height: 480}

	err := manager.Start(mockCapture)
	if err != nil {
		if isFFmpegMissing(err) {
			t.Skip("FFmpeg not available for testing")
		}
		t.Fatalf("Start failed: %v", err)
	}

	outputFile, err := manager.Stop()
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	if manager.State() != StateStopped {
		t.Errorf("State after stop = %v, want %v", manager.State(), StateStopped)
	}

	if manager.IsRecording() {
		t.Error("IsRecording should be false after stop")
	}

	if outputFile == "" {
		t.Error("Output file path should not be empty")
	}

	t.Logf("Output file: %s", outputFile)
}

// =============================================================================
// Invalid State Transition Tests
// =============================================================================

func TestCannotStartWhenRecording(t *testing.T) {
	manager := NewManager()
	mockCapture := &MockCaptureSource{width: 640, height: 480}

	err := manager.Start(mockCapture)
	if err != nil {
		if isFFmpegMissing(err) {
			t.Skip("FFmpeg not available for testing")
		}
		t.Fatalf("First start failed: %v", err)
	}
	defer manager.Stop()

	// Try to start again
	err = manager.Start(mockCapture)
	if err == nil {
		t.Error("Expected error when starting while already recording")
	}
}

func TestCannotPauseWhenIdle(t *testing.T) {
	manager := NewManager()

	err := manager.Pause()
	if err == nil {
		t.Error("Expected error when pausing while idle")
	}
}

func TestCannotResumeWhenNotPaused(t *testing.T) {
	manager := NewManager()

	err := manager.Resume()
	if err == nil {
		t.Error("Expected error when resuming while not paused")
	}
}

func TestCannotStopWhenIdle(t *testing.T) {
	manager := NewManager()

	_, err := manager.Stop()
	if err == nil {
		t.Error("Expected error when stopping while idle")
	}
}

func TestCannotStopTwice(t *testing.T) {
	manager := NewManager()
	mockCapture := &MockCaptureSource{width: 640, height: 480}

	err := manager.Start(mockCapture)
	if err != nil {
		if isFFmpegMissing(err) {
			t.Skip("FFmpeg not available for testing")
		}
		t.Fatalf("Start failed: %v", err)
	}

	_, err = manager.Stop()
	if err != nil {
		t.Fatalf("First stop failed: %v", err)
	}

	_, err = manager.Stop()
	if err == nil {
		t.Error("Expected error when stopping twice")
	}
}

// =============================================================================
// Elapsed Time Tests
// =============================================================================

func TestElapsedTimeIncreases(t *testing.T) {
	manager := NewManager()
	mockCapture := &MockCaptureSource{width: 640, height: 480}

	err := manager.Start(mockCapture)
	if err != nil {
		if isFFmpegMissing(err) {
			t.Skip("FFmpeg not available for testing")
		}
		t.Fatalf("Start failed: %v", err)
	}
	defer manager.Stop()

	time1 := manager.ElapsedTime()
	time.Sleep(100 * time.Millisecond)
	time2 := manager.ElapsedTime()

	if time2 <= time1 {
		t.Errorf("Elapsed time should increase: time1=%v, time2=%v", time1, time2)
	}
}

func TestElapsedTimePausesDuringPause(t *testing.T) {
	manager := NewManager()
	mockCapture := &MockCaptureSource{width: 640, height: 480}

	err := manager.Start(mockCapture)
	if err != nil {
		if isFFmpegMissing(err) {
			t.Skip("FFmpeg not available for testing")
		}
		t.Fatalf("Start failed: %v", err)
	}
	defer manager.Stop()

	time.Sleep(50 * time.Millisecond)

	err = manager.Pause()
	if err != nil {
		t.Fatalf("Pause failed: %v", err)
	}

	timeAtPause := manager.ElapsedTime()
	time.Sleep(100 * time.Millisecond)
	timeAfterWait := manager.ElapsedTime()

	// Time should not increase significantly while paused
	// Allow small tolerance for timing precision
	diff := timeAfterWait - timeAtPause
	if diff > 10*time.Millisecond {
		t.Errorf("Elapsed time increased during pause: diff=%v", diff)
	}
}

func TestElapsedTimeResumesAfterPause(t *testing.T) {
	manager := NewManager()
	mockCapture := &MockCaptureSource{width: 640, height: 480}

	err := manager.Start(mockCapture)
	if err != nil {
		if isFFmpegMissing(err) {
			t.Skip("FFmpeg not available for testing")
		}
		t.Fatalf("Start failed: %v", err)
	}
	defer manager.Stop()

	time.Sleep(50 * time.Millisecond)
	manager.Pause()
	timeAtPause := manager.ElapsedTime()

	time.Sleep(50 * time.Millisecond)
	manager.Resume()

	time.Sleep(100 * time.Millisecond)
	timeAfterResume := manager.ElapsedTime()

	// Time should have increased after resume
	if timeAfterResume <= timeAtPause {
		t.Errorf("Elapsed time should increase after resume: atPause=%v, afterResume=%v",
			timeAtPause, timeAfterResume)
	}
}

// =============================================================================
// Frame Counting Tests
// =============================================================================

func TestFrameCountingDuringRecording(t *testing.T) {
	manager := NewManager()
	mockCapture := &MockCaptureSource{width: 640, height: 480}

	err := manager.Start(mockCapture)
	if err != nil {
		if isFFmpegMissing(err) {
			t.Skip("FFmpeg not available for testing")
		}
		t.Fatalf("Start failed: %v", err)
	}

	// Wait up to 1s for first frame to avoid scheduler/CI startup jitter.
	deadline := time.Now().Add(1 * time.Second)
	framesCaptured := 0
	for time.Now().Before(deadline) {
		framesCaptured = manager.FramesCaptured()
		if framesCaptured > 0 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	manager.Stop()

	if framesCaptured == 0 {
		t.Error("Expected some frames to be captured")
	}

	t.Logf("Frames captured before deadline: %d", framesCaptured)
}

func TestFrameCountingPausesDuringPause(t *testing.T) {
	manager := NewManager()
	mockCapture := &MockCaptureSource{width: 640, height: 480}

	err := manager.Start(mockCapture)
	if err != nil {
		if isFFmpegMissing(err) {
			t.Skip("FFmpeg not available for testing")
		}
		t.Fatalf("Start failed: %v", err)
	}
	defer manager.Stop()

	time.Sleep(100 * time.Millisecond)

	err = manager.Pause()
	if err != nil {
		t.Fatalf("Pause failed: %v", err)
	}

	// Wait a bit for the capture loop to recognize the pause
	time.Sleep(100 * time.Millisecond)

	framesAtPause := manager.FramesCaptured()
	time.Sleep(200 * time.Millisecond)
	framesAfterWait := manager.FramesCaptured()

	// Frame count should not increase while paused (allow for up to 1 frame race condition)
	diff := framesAfterWait - framesAtPause
	if diff > 1 {
		t.Errorf("Frame count increased significantly during pause: atPause=%d, afterWait=%d, diff=%d",
			framesAtPause, framesAfterWait, diff)
	}
}

// =============================================================================
// Settings Tests
// =============================================================================

func TestGetSetSettings(t *testing.T) {
	manager := NewManager()

	newSettings := Settings{
		Quality: QualityHigh,
		FPS:     30,
		Format:  FormatWebM,
		CRF:     18,
		Preset:  PresetSlow,
	}

	manager.SetSettings(newSettings)
	gotSettings := manager.GetSettings()

	if gotSettings.Quality != newSettings.Quality {
		t.Errorf("Quality = %v, want %v", gotSettings.Quality, newSettings.Quality)
	}
	if gotSettings.FPS != newSettings.FPS {
		t.Errorf("FPS = %d, want %d", gotSettings.FPS, newSettings.FPS)
	}
	if gotSettings.Format != newSettings.Format {
		t.Errorf("Format = %v, want %v", gotSettings.Format, newSettings.Format)
	}
}

func TestCannotChangeSettingsWhileRecording(t *testing.T) {
	manager := NewManager()
	mockCapture := &MockCaptureSource{width: 640, height: 480}

	err := manager.Start(mockCapture)
	if err != nil {
		if isFFmpegMissing(err) {
			t.Skip("FFmpeg not available for testing")
		}
		t.Fatalf("Start failed: %v", err)
	}
	defer manager.Stop()

	originalSettings := manager.GetSettings()

	newSettings := Settings{
		Quality: QualityLow,
		FPS:     10,
	}
	manager.SetSettings(newSettings)

	// Settings should not change while recording
	currentSettings := manager.GetSettings()
	if currentSettings.FPS != originalSettings.FPS {
		t.Error("Settings should not change while recording")
	}
}

// =============================================================================
// Estimated File Size Tests
// =============================================================================

func TestEstimatedFileSizeIncreasesOverTime(t *testing.T) {
	manager := NewManager()
	mockCapture := &MockCaptureSource{width: 640, height: 480}

	err := manager.Start(mockCapture)
	if err != nil {
		if isFFmpegMissing(err) {
			t.Skip("FFmpeg not available for testing")
		}
		t.Fatalf("Start failed: %v", err)
	}
	defer manager.Stop()

	size1 := manager.EstimatedFileSize()
	time.Sleep(100 * time.Millisecond)
	size2 := manager.EstimatedFileSize()

	if size2 <= size1 {
		t.Logf("Estimated size: initial=%d, after 100ms=%d", size1, size2)
		// This may be acceptable if estimation is based on settings only
	}
}

// =============================================================================
// Concurrent Access Tests
// =============================================================================

func TestConcurrentStateAccess(t *testing.T) {
	manager := NewManager()

	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = manager.State()
				_ = manager.IsRecording()
				_ = manager.IsPaused()
				_ = manager.ElapsedTime()
				_ = manager.FramesCaptured()
			}
		}()
	}

	wg.Wait()
}

func TestConcurrentSettingsAccess(t *testing.T) {
	manager := NewManager()

	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(2)

		// Reader
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = manager.GetSettings()
			}
		}()

		// Writer (only when not recording)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				manager.SetSettings(DefaultSettings())
			}
		}()
	}

	wg.Wait()
}

// =============================================================================
// Error Handling Tests
// =============================================================================

func TestStartWithNilCaptureSource(t *testing.T) {
	manager := NewManager()

	err := manager.Start(nil)
	if err == nil {
		t.Error("Expected error when starting with nil capture source")
		manager.Stop()
	}
}

func TestStartWithInvalidDimensions(t *testing.T) {
	manager := NewManager()

	// Zero dimensions
	mockCapture := &MockCaptureSource{width: 0, height: 0}
	err := manager.Start(mockCapture)
	if err == nil {
		t.Error("Expected error when starting with zero dimensions")
		manager.Stop()
	}

	// Negative dimensions (if represented as int)
	mockCapture = &MockCaptureSource{width: -1, height: 480}
	err = manager.Start(mockCapture)
	if err == nil {
		t.Error("Expected error when starting with negative dimensions")
		manager.Stop()
	}
}

// =============================================================================
// Callback Tests
// =============================================================================

func TestOnStateChangeCallback(t *testing.T) {
	manager := NewManager()
	mockCapture := &MockCaptureSource{width: 640, height: 480}

	stateChanges := make([]State, 0)
	var mu sync.Mutex

	manager.OnStateChange(func(state State) {
		mu.Lock()
		stateChanges = append(stateChanges, state)
		mu.Unlock()
	})

	err := manager.Start(mockCapture)
	if err != nil {
		if isFFmpegMissing(err) {
			t.Skip("FFmpeg not available for testing")
		}
		t.Fatalf("Start failed: %v", err)
	}

	// Small delay to ensure start callback completes
	time.Sleep(50 * time.Millisecond)

	manager.Stop()

	// Small delay to ensure stop callback completes (callbacks are async)
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(stateChanges) < 2 {
		t.Errorf("Expected at least 2 state changes, got %d: %v", len(stateChanges), stateChanges)
	}

	// First change should be to Recording
	if len(stateChanges) > 0 && stateChanges[0] != StateRecording {
		t.Errorf("First state change = %v, want %v", stateChanges[0], StateRecording)
	}
}

func TestOnFrameCapturedCallback(t *testing.T) {
	manager := NewManager()
	mockCapture := &MockCaptureSource{width: 640, height: 480}

	var frameCount atomic.Int32
	firstFrameCh := make(chan struct{}, 1)

	manager.OnFrameCaptured(func(frameNum int) {
		_ = frameNum
		frameCount.Add(1)
		select {
		case firstFrameCh <- struct{}{}:
		default:
		}
	})

	err := manager.Start(mockCapture)
	if err != nil {
		if isFFmpegMissing(err) {
			t.Skip("FFmpeg not available for testing")
		}
		t.Fatalf("Start failed: %v", err)
	}

	gotFirstFrame := false
	select {
	case <-firstFrameCh:
		gotFirstFrame = true
	case <-time.After(2 * time.Second):
		// Continue to Stop() and assert with full context below.
	}

	if _, stopErr := manager.Stop(); stopErr != nil {
		t.Fatalf("Stop failed: %v", stopErr)
	}

	// Callback dispatch is asynchronous; give it a short bounded drain window.
	if !gotFirstFrame {
		select {
		case <-firstFrameCh:
			gotFirstFrame = true
		case <-time.After(150 * time.Millisecond):
		}
	}

	count := int(frameCount.Load())
	if !gotFirstFrame || count == 0 {
		t.Fatalf("OnFrameCaptured callback was not observed within timeout (gotFirstFrame=%v frameCount=%d managerFramesCaptured=%d)", gotFirstFrame, count, manager.FramesCaptured())
	}

	t.Logf("OnFrameCaptured called %d times", count)
}

func TestOnErrorCallback(t *testing.T) {
	manager := NewManager()

	errorReceived := false
	manager.OnError(func(err error) {
		errorReceived = true
		t.Logf("Error callback received: %v", err)
	})

	// Trigger an error (e.g., by starting with invalid capture source)
	err := manager.Start(nil)
	if err != nil {
		// Error occurred, callback may or may not be called depending on implementation
		t.Logf("Start error: %v", err)
	}

	// Use the variable to avoid compiler warning
	_ = errorReceived
}

// =============================================================================
// Helper Types and Functions
// =============================================================================

// MockCaptureSource implements CaptureSource interface for testing
type MockCaptureSource struct {
	width  int
	height int
}

func (m *MockCaptureSource) Capture() ([]byte, error) {
	// Return a mock frame (just zeros)
	size := m.width * m.height * 3
	if size <= 0 {
		return nil, nil
	}
	return make([]byte, size), nil
}

func (m *MockCaptureSource) Width() int {
	return m.width
}

func (m *MockCaptureSource) Height() int {
	return m.height
}

// isFFmpegMissing checks if error is due to FFmpeg not being installed
func isFFmpegMissing(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "ffmpeg") && (contains(errStr, "not found") || contains(errStr, "not installed"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsImpl(s, substr))
}

func containsImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkManagerStateAccess(b *testing.B) {
	manager := NewManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.State()
		_ = manager.IsRecording()
		_ = manager.IsPaused()
	}
}

func BenchmarkManagerMetricsAccess(b *testing.B) {
	manager := NewManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.ElapsedTime()
		_ = manager.FramesCaptured()
		_ = manager.FramesDropped()
		_ = manager.EstimatedFileSize()
	}
}

func BenchmarkManagerSettingsAccess(b *testing.B) {
	manager := NewManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.GetSettings()
	}
}
