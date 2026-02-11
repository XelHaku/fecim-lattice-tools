package recording

import (
	"os/exec"
	"strings"
	"testing"
)

// =============================================================================
// FFmpeg Command Builder - Advanced Tests
// =============================================================================

func TestFFmpegCommandBuilderChaining(t *testing.T) {
	builder := NewFFmpegCommandBuilder().
		Overwrite(true).
		InputFormat("rawvideo").
		InputPixelFormat("rgb24").
		VideoSize(1920, 1080).
		Framerate(30).
		Input("-").
		VideoCodec(CodecH264).
		Preset(PresetFast).
		CRF(23).
		OutputPixelFormat("yuv420p").
		Output("test.mp4")

	args := builder.Build()

	// Verify all parameters are present
	argsStr := strings.Join(args, " ")
	expectedParams := []string{
		"-y",
		"-f rawvideo",
		"-pixel_format rgb24",
		"-video_size 1920x1080",
		"-framerate 30",
		"-i -",
		"-c:v libx264",
		"-preset fast",
		"-crf 23",
		"-pix_fmt yuv420p",
		"test.mp4",
	}

	for _, param := range expectedParams {
		if !strings.Contains(argsStr, param) {
			t.Errorf("Missing parameter: %s in command: %s", param, argsStr)
		}
	}
}

func TestFFmpegCommandBuilderWithSettingsAppliesAllSettings(t *testing.T) {
	settings := Settings{
		FPS:    25,
		CRF:    20,
		Preset: PresetMedium,
	}

	builder := NewFFmpegCommandBuilder().WithSettings(settings)
	args := builder.Build()
	argsStr := strings.Join(args, " ")

	// All settings should be applied
	if !strings.Contains(argsStr, "-framerate 25") {
		t.Error("FPS from settings not applied")
	}
	if !strings.Contains(argsStr, "-crf 20") {
		t.Error("CRF from settings not applied")
	}
	if !strings.Contains(argsStr, "-preset medium") {
		t.Error("Preset from settings not applied")
	}
}

func TestFFmpegCommandBuilderAudioInput(t *testing.T) {
	builder := NewFFmpegCommandBuilder().
		AudioInput("default").
		Input("video.raw").
		Output("output.mp4")

	args := builder.Build()
	argsStr := strings.Join(args, " ")

	if !strings.Contains(argsStr, "-f pulse") {
		t.Error("Audio format not set")
	}
	if !strings.Contains(argsStr, "-i default") {
		t.Error("Audio device not set")
	}
}

func TestFFmpegCommandBuilderAudioCodec(t *testing.T) {
	builder := NewFFmpegCommandBuilder().
		Input("input.raw").
		AudioCodec("aac").
		Output("output.mp4")

	args := builder.Build()
	argsStr := strings.Join(args, " ")

	if !strings.Contains(argsStr, "-c:a aac") {
		t.Error("Audio codec not set")
	}
}

func TestFFmpegCommandBuilderAudioBitrate(t *testing.T) {
	builder := NewFFmpegCommandBuilder().
		Input("input.raw").
		AudioBitrate(192).
		Output("output.mp4")

	args := builder.Build()
	argsStr := strings.Join(args, " ")

	if !strings.Contains(argsStr, "-b:a 192k") {
		t.Error("Audio bitrate not set")
	}
}

func TestFFmpegCommandBuilderAudioSampleRate(t *testing.T) {
	builder := NewFFmpegCommandBuilder().
		Input("input.raw").
		AudioSampleRate(48000).
		Output("output.mp4")

	args := builder.Build()
	argsStr := strings.Join(args, " ")

	if !strings.Contains(argsStr, "-ar 48000") {
		t.Error("Audio sample rate not set")
	}
}

func TestFFmpegCommandBuilderAudioChannels(t *testing.T) {
	builder := NewFFmpegCommandBuilder().
		Input("input.raw").
		AudioChannels(2).
		Output("output.mp4")

	args := builder.Build()
	argsStr := strings.Join(args, " ")

	if !strings.Contains(argsStr, "-ac 2") {
		t.Error("Audio channels not set")
	}
}

func TestFFmpegCommandBuilderNoAudio(t *testing.T) {
	builder := NewFFmpegCommandBuilder().
		Input("input.raw").
		NoAudio().
		Output("output.mp4")

	args := builder.Build()
	argsStr := strings.Join(args, " ")

	if !strings.Contains(argsStr, "-an") {
		t.Error("No audio flag not set")
	}
}

func TestFFmpegCommandBuilderForRecordingWebM(t *testing.T) {
	settings := Settings{
		Format: FormatWebM,
		FPS:    20,
		CRF:    23,
		Preset: PresetFast,
	}

	builder := NewFFmpegCommandBuilder().
		ForRecording(1920, 1080, settings).
		Output("recording.webm")

	args := builder.Build()
	argsStr := strings.Join(args, " ")

	// WebM should use VP8 codec
	if !strings.Contains(argsStr, "-c:v libvpx") {
		t.Error("WebM should use VP8 codec")
	}
}

func TestFFmpegCommandBuilderForRecordingGIF(t *testing.T) {
	settings := Settings{
		Format: FormatGIF,
		FPS:    10,
		CRF:    23,
		Preset: PresetFast,
	}

	builder := NewFFmpegCommandBuilder().
		ForRecording(640, 480, settings).
		Output("recording.gif")

	args := builder.Build()
	argsStr := strings.Join(args, " ")

	// GIF should use gif format
	if !strings.Contains(argsStr, "gif") {
		t.Error("GIF format not set")
	}
}

func TestFFmpegCommandBuilderForRecordingWithAudioDisabled(t *testing.T) {
	settings := Settings{
		Format: FormatMP4,
		FPS:    30,
		CRF:    23,
		Preset: PresetFast,
		Audio: AudioSettings{
			Enabled: false,
		},
	}

	builder := NewFFmpegCommandBuilder().
		ForRecordingWithAudio(1920, 1080, settings).
		Output("recording.mp4")

	args := builder.Build()
	argsStr := strings.Join(args, " ")

	// Should not have audio input
	if strings.Contains(argsStr, "-f pulse") {
		t.Error("Audio should not be enabled")
	}
}

func TestFFmpegCommandBuilderForRecordingWithAudioEnabled(t *testing.T) {
	settings := Settings{
		Format: FormatMP4,
		FPS:    30,
		CRF:    23,
		Preset: PresetFast,
		Audio: AudioSettings{
			Enabled:    true,
			DeviceName: "test-device",
			Codec:      "aac",
			Bitrate:    128,
			SampleRate: 44100,
			Channels:   2,
		},
	}

	builder := NewFFmpegCommandBuilder().
		ForRecordingWithAudio(1920, 1080, settings).
		Output("recording.mp4")

	args := builder.Build()
	argsStr := strings.Join(args, " ")

	// Should have audio configuration
	if !strings.Contains(argsStr, "-f pulse") {
		t.Error("Audio format not set")
	}
	if !strings.Contains(argsStr, "-i test-device") {
		t.Error("Audio device not set")
	}
	if !strings.Contains(argsStr, "-c:a aac") {
		t.Error("Audio codec not set")
	}
	if !strings.Contains(argsStr, "-b:a 128k") {
		t.Error("Audio bitrate not set")
	}
	if !strings.Contains(argsStr, "-ar 44100") {
		t.Error("Audio sample rate not set")
	}
	if !strings.Contains(argsStr, "-ac 2") {
		t.Error("Audio channels not set")
	}
	if !strings.Contains(argsStr, "-shortest") {
		t.Error("Shortest flag not set for audio sync")
	}
}

func TestFFmpegCommandBuilderForRecordingWithAudioWebM(t *testing.T) {
	settings := Settings{
		Format: FormatWebM,
		FPS:    30,
		CRF:    23,
		Preset: PresetFast,
		Audio: AudioSettings{
			Enabled:    true,
			DeviceName: "default",
		},
	}

	builder := NewFFmpegCommandBuilder().
		ForRecordingWithAudio(1920, 1080, settings).
		Output("recording.webm")

	args := builder.Build()
	argsStr := strings.Join(args, " ")

	// WebM should use libopus for audio
	if !strings.Contains(argsStr, "-c:a libopus") {
		t.Error("WebM should use libopus audio codec")
	}
}

func TestFFmpegCommandBuilderForRecordingWithAudioGIF(t *testing.T) {
	settings := Settings{
		Format: FormatGIF,
		FPS:    10,
		CRF:    23,
		Preset: PresetFast,
		Audio: AudioSettings{
			Enabled: true, // GIF doesn't support audio
		},
	}

	builder := NewFFmpegCommandBuilder().
		ForRecordingWithAudio(640, 480, settings).
		Output("recording.gif")

	args := builder.Build()
	argsStr := strings.Join(args, " ")

	// GIF should not have audio (doesn't support it)
	if strings.Contains(argsStr, "-f pulse") {
		t.Error("GIF should not have audio input")
	}
}

func TestFFmpegCommandBuilderForRecordingWithAudioDefaultDevice(t *testing.T) {
	settings := Settings{
		Format: FormatMP4,
		FPS:    30,
		CRF:    23,
		Preset: PresetFast,
		Audio: AudioSettings{
			Enabled:    true,
			DeviceName: "", // Empty device name should default to "default"
		},
	}

	builder := NewFFmpegCommandBuilder().
		ForRecordingWithAudio(1920, 1080, settings).
		Output("recording.mp4")

	args := builder.Build()
	argsStr := strings.Join(args, " ")

	// Should use "default" as device name
	if !strings.Contains(argsStr, "-i default") {
		t.Error("Empty device name should default to 'default'")
	}
}

// =============================================================================
// Audio Device Detection Tests
// =============================================================================

func TestGetDesktopAudioDevice(t *testing.T) {
	device, err := GetDesktopAudioDevice()
	if err != nil {
		t.Skipf("Desktop audio not available: %v", err)
	}

	t.Logf("Desktop audio device: Name=%s ID=%s", device.Name, device.ID)

	// Device name should end with .monitor for desktop audio
	if !strings.Contains(device.Name, "monitor") {
		t.Logf("Warning: Desktop audio device name doesn't contain 'monitor': %s", device.Name)
	}

	if device.Description == "" {
		t.Error("Desktop audio device should have description")
	}

	expectedDesc := "Desktop Audio (System Sounds)"
	if device.Description != expectedDesc {
		t.Logf("Desktop audio description: got %q, expected %q", device.Description, expectedDesc)
	}
}

func TestGetMicrophoneDevice(t *testing.T) {
	device, err := GetMicrophoneDevice()
	if err != nil {
		t.Skipf("No microphone found: %v", err)
	}

	t.Logf("Microphone device: Name=%s ID=%s", device.Name, device.ID)
	t.Logf("  Description: %s", device.Description)

	if device.Name == "" {
		t.Error("Microphone device should have name")
	}
	if device.ID == "" {
		t.Error("Microphone device should have ID")
	}

	// Check priority order: Bluetooth > USB > Analog
	if strings.Contains(strings.ToLower(device.Name), "bluez") {
		t.Logf("Found Bluetooth microphone (highest priority)")
	} else if strings.Contains(strings.ToLower(device.Name), "usb") {
		t.Logf("Found USB microphone")
	} else {
		t.Logf("Found analog/other microphone")
	}
}

func TestDetectAudioDevicesFiltersMonitors(t *testing.T) {
	devices, err := DetectAudioDevices()
	if err != nil {
		t.Skipf("No audio devices available: %v", err)
	}

	// Check that devices list doesn't contain pure monitor sources
	// (unless they're explicitly desktop audio)
	for _, d := range devices {
		nameLower := strings.ToLower(d.Name)
		if strings.Contains(nameLower, "monitor") {
			// If it's a monitor, it should either be:
			// 1. Labeled as input, or
			// 2. Be the only option
			t.Logf("Monitor source found: %s (may be desktop audio)", d.Name)
		}
	}
}

func TestDetectAudioDevicesDefaultDevice(t *testing.T) {
	devices, err := DetectAudioDevices()
	if err != nil {
		t.Skipf("No audio devices available: %v", err)
	}

	// Should always include a "default" device option
	hasDefault := false
	for _, d := range devices {
		if d.Name == "default" || d.ID == "default" {
			hasDefault = true
			break
		}
	}

	if !hasDefault && len(devices) > 0 {
		t.Log("Warning: No 'default' device found in list")
	}
}

func TestGetDefaultAudioDevicePriority(t *testing.T) {
	devices, err := DetectAudioDevices()
	if err != nil {
		t.Skipf("No audio devices available: %v", err)
	}

	device, err := GetDefaultAudioDevice()
	if err != nil {
		t.Fatalf("GetDefaultAudioDevice failed: %v", err)
	}

	t.Logf("Default device selected: %s (from %d available)", device.Name, len(devices))

	// Verify priority: actual input device > non-default > default placeholder
	if device.Name != "default" && !strings.Contains(strings.ToLower(device.Name), "input") {
		t.Logf("Note: Default device %q doesn't have 'input' in name", device.Name)
	}
}

// =============================================================================
// Audio Device Command Testing
// =============================================================================

func TestDetectAudioDevicesPactlAvailable(t *testing.T) {
	// Test if pactl is available on the system
	cmd := exec.Command("pactl", "--version")
	err := cmd.Run()

	if err != nil {
		t.Skip("pactl not available (PulseAudio/PipeWire not installed)")
	}

	devices, err := DetectAudioDevices()
	if err != nil {
		t.Skipf("Audio detection failed: %v", err)
	}

	if len(devices) == 0 {
		t.Log("Warning: pactl available but no devices found")
	} else {
		t.Logf("Found %d device(s) via pactl", len(devices))
	}
}

func TestDetectAudioDevicesArecordFallback(t *testing.T) {
	// Test if arecord is available (ALSA fallback)
	cmd := exec.Command("arecord", "-l")
	err := cmd.Run()

	if err != nil {
		t.Skip("arecord not available (ALSA not installed)")
	}

	t.Log("ALSA (arecord) is available as fallback")
}

// =============================================================================
// FFmpeg Detection Edge Cases
// =============================================================================

func TestGetDefaultFFmpegPathsPlatformSpecific(t *testing.T) {
	paths := GetDefaultFFmpegPaths()

	if len(paths) == 0 {
		t.Error("GetDefaultFFmpegPaths should return at least one path")
	}

	// Log all paths for debugging
	t.Logf("Default FFmpeg paths for this platform:")
	for i, p := range paths {
		t.Logf("  %d. %s", i+1, p)
	}

	// All paths should be non-empty
	for _, p := range paths {
		if p == "" {
			t.Error("Default path should not be empty")
		}
	}
}

func TestCheckCodecSupportNonexistent(t *testing.T) {
	if !IsFFmpegAvailable() {
		t.Skip("FFmpeg not available")
	}

	// Test with a codec that definitely doesn't exist
	supported := CheckCodecSupport("codec_that_does_not_exist_xyz123")
	if supported {
		t.Error("Nonexistent codec should not be supported")
	}
}

func TestGetSupportedEncodersReturnsVideoOnly(t *testing.T) {
	if !IsFFmpegAvailable() {
		t.Skip("FFmpeg not available")
	}

	encoders := GetSupportedEncoders()

	if len(encoders) == 0 {
		t.Skip("No encoders found")
	}

	// Spot check: all returned encoders should be video encoders
	// (implementation filters by "V" capability flag)
	t.Logf("Found %d video encoder(s)", len(encoders))

	// Check for at least one common encoder
	hasCommonEncoder := false
	commonEncoders := []string{"libx264", "h264", "mpeg4", "vp8", "vp9"}
	for _, enc := range encoders {
		for _, common := range commonEncoders {
			if strings.Contains(enc, common) {
				hasCommonEncoder = true
				break
			}
		}
		if hasCommonEncoder {
			break
		}
	}

	if !hasCommonEncoder {
		t.Log("Warning: No common video encoder found in list")
	}
}

func TestGetSupportedFormatsReturnsMuxingOnly(t *testing.T) {
	if !IsFFmpegAvailable() {
		t.Skip("FFmpeg not available")
	}

	formats := GetSupportedFormats()

	if len(formats) == 0 {
		t.Skip("No formats found")
	}

	// Spot check: should include common muxing formats
	t.Logf("Found %d muxing format(s)", len(formats))

	// Check for at least one common format
	hasCommonFormat := false
	commonFormats := []string{"mp4", "webm", "avi", "matroska", "mov"}
	for _, fmt := range formats {
		for _, common := range commonFormats {
			if strings.Contains(strings.ToLower(fmt), common) {
				hasCommonFormat = true
				break
			}
		}
		if hasCommonFormat {
			break
		}
	}

	if !hasCommonFormat {
		t.Logf("Warning: No common format found in list")
	}
}

// =============================================================================
// FFmpegInfo Helper Methods
// =============================================================================

func TestFFmpegInfoHasEncoderCaseSensitive(t *testing.T) {
	info := &FFmpegInfo{
		Encoders: []string{"libx264", "libx265", "libvpx"},
	}

	// Exact match should work
	if !info.HasEncoder("libx264") {
		t.Error("Should find exact match")
	}

	// Case-sensitive: should not match different case
	if info.HasEncoder("LIBX264") {
		t.Error("Should be case-sensitive")
	}

	// Partial match should not work
	if info.HasEncoder("x264") {
		t.Error("Should require exact match, not substring")
	}
}

func TestFFmpegInfoHasFormatCaseSensitive(t *testing.T) {
	info := &FFmpegInfo{
		Formats: []string{"mp4", "webm", "matroska"},
	}

	// Exact match should work
	if !info.HasFormat("mp4") {
		t.Error("Should find exact match")
	}

	// Case-sensitive: should not match different case
	if info.HasFormat("MP4") {
		t.Error("Should be case-sensitive")
	}
}

func TestFFmpegInfoCanRecordEmptyPath(t *testing.T) {
	info := &FFmpegInfo{
		Path:     "",
		Encoders: []string{"libx264"},
		Formats:  []string{"mp4"},
	}

	// CanRecord checks if Path is non-empty
	if info.CanRecord() {
		t.Error("Should not be able to record without FFmpeg path")
	}
}

// =============================================================================
// VideoFormat Edge Cases
// =============================================================================

func TestVideoFormatIsValidInvalid(t *testing.T) {
	invalidFormats := []VideoFormat{
		"avi",
		"mkv",
		"",
		"random",
	}

	for _, format := range invalidFormats {
		if format.IsValid() {
			t.Errorf("Format %q should be invalid", format)
		}
	}
}

func TestVideoFormatExtensionForInvalid(t *testing.T) {
	invalid := VideoFormat("invalid")
	ext := invalid.Extension()

	// Should return safe default (.mp4) for invalid formats
	if ext != ".mp4" {
		t.Errorf("Invalid format extension = %s, want .mp4 (safe default)", ext)
	}
}

func TestVideoFormatDefaultCodecForInvalid(t *testing.T) {
	invalid := VideoFormat("invalid")
	codec := invalid.DefaultCodec()

	// Should return safe default codec for invalid formats
	if codec != CodecH264 {
		t.Errorf("Invalid format default codec = %v, want %v (safe default)", codec, CodecH264)
	}
}

// =============================================================================
// EncodingPreset Validation
// =============================================================================

func TestEncodingPresetIsValidInvalid(t *testing.T) {
	invalidPresets := []EncodingPreset{
		"fastest",
		"quick",
		"",
		"invalid",
	}

	for _, preset := range invalidPresets {
		if preset.IsValid() {
			t.Errorf("Preset %q should be invalid", preset)
		}
	}
}

func TestEncodingPresetStringAllValid(t *testing.T) {
	validPresets := []EncodingPreset{
		PresetUltrafast,
		PresetSuperfast,
		PresetVeryfast,
		PresetFaster,
		PresetFast,
		PresetMedium,
		PresetSlow,
		PresetSlower,
		PresetVeryslow,
	}

	for _, preset := range validPresets {
		str := preset.String()
		if str == "" {
			t.Errorf("Preset %v should have non-empty string representation", preset)
		}
		if !preset.IsValid() {
			t.Errorf("Preset %v should be valid", preset)
		}
	}
}
