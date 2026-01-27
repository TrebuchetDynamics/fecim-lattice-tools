package recording

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestNewAudioMonitor(t *testing.T) {
	am := NewAudioMonitor()
	if am == nil {
		t.Fatal("NewAudioMonitor returned nil")
	}
	if am.sampleRate != 44100 {
		t.Errorf("Expected default sample rate 44100, got %d", am.sampleRate)
	}
	if am.IsRunning() {
		t.Error("New monitor should not be running")
	}
}

func TestAudioMonitorLevel(t *testing.T) {
	am := NewAudioMonitor()

	// Initial level should be 0
	if am.Level() != 0 {
		t.Errorf("Initial level should be 0, got %d", am.Level())
	}
	if am.PeakLevel() != 0 {
		t.Errorf("Initial peak should be 0, got %d", am.PeakLevel())
	}
}

func TestAudioMonitorDbToPercent(t *testing.T) {
	am := NewAudioMonitor()

	// dbToPercent uses -80 dB (noise floor) to -10 dB (loud) range
	// Linear mapping: percent = (db - (-80)) / ((-10) - (-80)) * 100
	tests := []struct {
		db       float64
		expected int
	}{
		{-80, 0},   // Noise floor (min)
		{-10, 100}, // Loud/max
		{-45, 50},  // Mid-range: (-45+80)/70*100 = 50
		{-60, 28},  // Quiet speech: (-60+80)/70*100 ≈ 28
		{-30, 71},  // Loud speech: (-30+80)/70*100 ≈ 71
		{0, 100},   // Above max (clamped)
		{-90, 0},   // Below min (clamped)
		{10, 100},  // Way above max (clamped)
	}

	for _, test := range tests {
		result := am.dbToPercent(test.db)
		if result != test.expected {
			t.Errorf("dbToPercent(%f) = %d, expected %d", test.db, result, test.expected)
		}
	}
}

func TestDefaultAudioSettings(t *testing.T) {
	s := DefaultAudioSettings()

	if s.Enabled {
		t.Error("Default audio should be disabled")
	}
	if s.SampleRate != 44100 {
		t.Errorf("Default sample rate should be 44100, got %d", s.SampleRate)
	}
	if s.Channels != 2 {
		t.Errorf("Default channels should be 2 (stereo), got %d", s.Channels)
	}
	if s.Codec != "aac" {
		t.Errorf("Default codec should be aac, got %s", s.Codec)
	}
	if s.Bitrate != 128 {
		t.Errorf("Default bitrate should be 128, got %d", s.Bitrate)
	}
}

func TestAudioSettingsValidateDisabled(t *testing.T) {
	s := DefaultAudioSettings()
	s.Enabled = false

	// Disabled settings should always be valid
	if err := s.Validate(); err != nil {
		t.Errorf("Disabled settings should be valid: %v", err)
	}

	// Even with invalid values, disabled settings pass
	s.SampleRate = 12345
	s.Channels = 99
	if err := s.Validate(); err != nil {
		t.Errorf("Disabled settings should bypass validation: %v", err)
	}
}

func TestAudioSettingsValidateSampleRate(t *testing.T) {
	s := DefaultAudioSettings()
	s.Enabled = true

	validRates := []int{44100, 48000, 96000}
	for _, rate := range validRates {
		s.SampleRate = rate
		if err := s.Validate(); err != nil {
			t.Errorf("Sample rate %d should be valid: %v", rate, err)
		}
	}

	invalidRates := []int{22050, 32000, 11025, 0}
	for _, rate := range invalidRates {
		s.SampleRate = rate
		err := s.Validate()
		if err == nil {
			t.Errorf("Sample rate %d should be invalid", rate)
		}
		var ve *ValidationError
		if !AsValidationError(err, &ve) || ve.Field != "SampleRate" {
			t.Errorf("Expected SampleRate validation error for rate %d", rate)
		}
	}
}

func TestAudioSettingsValidateChannels(t *testing.T) {
	s := DefaultAudioSettings()
	s.Enabled = true

	// Valid channels
	for _, ch := range []int{1, 2} {
		s.Channels = ch
		if err := s.Validate(); err != nil {
			t.Errorf("Channels %d should be valid: %v", ch, err)
		}
	}

	// Invalid channels
	for _, ch := range []int{0, 3, 8} {
		s.Channels = ch
		err := s.Validate()
		if err == nil {
			t.Errorf("Channels %d should be invalid", ch)
		}
	}
}

func TestAudioSettingsValidateCodec(t *testing.T) {
	s := DefaultAudioSettings()
	s.Enabled = true

	validCodecs := []string{"aac", "mp3", "opus", "flac"}
	for _, codec := range validCodecs {
		s.Codec = codec
		if err := s.Validate(); err != nil {
			t.Errorf("Codec %s should be valid: %v", codec, err)
		}
	}

	invalidCodecs := []string{"wav", "ogg", "wma", ""}
	for _, codec := range invalidCodecs {
		s.Codec = codec
		err := s.Validate()
		if err == nil {
			t.Errorf("Codec %s should be invalid", codec)
		}
	}
}

func TestAudioSettingsValidateBitrate(t *testing.T) {
	s := DefaultAudioSettings()
	s.Enabled = true

	// Valid bitrates
	validBitrates := []int{64, 128, 192, 256, 320}
	for _, br := range validBitrates {
		s.Bitrate = br
		if err := s.Validate(); err != nil {
			t.Errorf("Bitrate %d should be valid: %v", br, err)
		}
	}

	// Invalid bitrates
	invalidBitrates := []int{32, 400, 0, -128}
	for _, br := range invalidBitrates {
		s.Bitrate = br
		err := s.Validate()
		if err == nil {
			t.Errorf("Bitrate %d should be invalid", br)
		}
	}
}

func TestAudioDevice(t *testing.T) {
	device := AudioDevice{
		ID:          "hw:0",
		Name:        "Test Device",
		Description: "A test audio device",
		IsDefault:   true,
	}

	if device.ID != "hw:0" {
		t.Errorf("Device ID mismatch")
	}
	if device.Name != "Test Device" {
		t.Errorf("Device Name mismatch")
	}
	if !device.IsDefault {
		t.Errorf("Device should be default")
	}
}

func TestAudioMonitorSetDevice(t *testing.T) {
	am := NewAudioMonitor()
	device := AudioDevice{
		ID:   "test",
		Name: "Test Device",
	}

	am.SetDevice(device)

	am.mu.RLock()
	savedDevice := am.device
	am.mu.RUnlock()

	if savedDevice.ID != device.ID {
		t.Error("Device not set correctly")
	}
}

func TestAudioMonitorCallbacks(t *testing.T) {
	am := NewAudioMonitor()

	levelCalled := false
	errorCalled := false

	am.OnLevelChange(func(level, peak int) {
		levelCalled = true
	})

	am.OnError(func(err error) {
		errorCalled = true
	})

	// Verify callbacks are set
	am.mu.RLock()
	hasLevelCallback := am.onLevelChange != nil
	hasErrorCallback := am.onError != nil
	am.mu.RUnlock()

	if !hasLevelCallback {
		t.Error("Level callback not set")
	}
	if !hasErrorCallback {
		t.Error("Error callback not set")
	}

	// Note: We don't test if callbacks are actually called since that
	// requires starting the monitor which needs audio hardware
	_ = levelCalled
	_ = errorCalled
}

func TestSettingsWithAudio(t *testing.T) {
	s := DefaultSettings()

	// Audio should be disabled by default
	if s.Audio.Enabled {
		t.Error("Audio should be disabled by default")
	}

	// Enable audio
	s.Audio.Enabled = true
	s.Audio.SampleRate = 48000
	s.Audio.Channels = 1
	s.Audio.Codec = "opus"
	s.Audio.Bitrate = 192

	// Should validate successfully
	if err := s.Validate(); err != nil {
		t.Errorf("Settings with valid audio should validate: %v", err)
	}
}

func TestSettingsValidationWithInvalidAudio(t *testing.T) {
	s := DefaultSettings()
	s.Audio.Enabled = true
	s.Audio.SampleRate = 12345 // Invalid

	err := s.Validate()
	if err == nil {
		t.Error("Settings with invalid audio should fail validation")
	}

	var ve *ValidationError
	if !AsValidationError(err, &ve) {
		t.Error("Should return ValidationError")
	}
}

func TestSettingsCopyIncludesAudio(t *testing.T) {
	s := DefaultSettings()
	s.Audio.Enabled = true
	s.Audio.DeviceID = "test-device"
	s.Audio.Bitrate = 256

	copy := s.Copy()

	if !copy.Audio.Enabled {
		t.Error("Copy should include Audio.Enabled")
	}
	if copy.Audio.DeviceID != "test-device" {
		t.Error("Copy should include Audio.DeviceID")
	}
	if copy.Audio.Bitrate != 256 {
		t.Error("Copy should include Audio.Bitrate")
	}
}

// TestDetectAudioDevicesIntegration tests actual device detection on the system.
// This test may be skipped if no audio devices are available.
func TestDetectAudioDevicesIntegration(t *testing.T) {
	devices, err := DetectAudioDevices()
	if err != nil {
		t.Skipf("No audio devices available: %v", err)
	}

	t.Logf("Found %d audio device(s):", len(devices))
	for i, d := range devices {
		t.Logf("  %d. ID=%s Name=%s Default=%v", i+1, d.ID, d.Name, d.IsDefault)
		t.Logf("     Description: %s", d.Description)
	}

	if len(devices) == 0 {
		t.Skip("No audio devices found")
	}

	// Verify at least one device has a name
	hasName := false
	for _, d := range devices {
		if d.Name != "" {
			hasName = true
			break
		}
	}
	if !hasName {
		t.Error("All devices have empty names")
	}
}

// TestIsAudioAvailableIntegration tests if audio is available on the system.
func TestIsAudioAvailableIntegration(t *testing.T) {
	available := IsAudioAvailable()
	t.Logf("Audio available: %v", available)
}

// TestGetDefaultAudioDeviceIntegration tests getting the default audio device.
func TestGetDefaultAudioDeviceIntegration(t *testing.T) {
	device, err := GetDefaultAudioDevice()
	if err != nil {
		t.Skipf("No default audio device: %v", err)
	}

	t.Logf("Default device: Name=%s ID=%s", device.Name, device.ID)
	t.Logf("  Description: %s", device.Description)

	if device.Name == "" {
		t.Error("Default device has empty name")
	}
}

// TestAudioMonitorStartStopIntegration tests starting and stopping the audio monitor.
func TestAudioMonitorStartStopIntegration(t *testing.T) {
	if !IsAudioAvailable() {
		t.Skip("No audio devices available")
	}

	am := NewAudioMonitor()

	// Set up callback to track level updates (use atomic for thread safety)
	var levelUpdates atomic.Int32
	am.OnLevelChange(func(level, peak int) {
		count := levelUpdates.Add(1)
		if count <= 3 {
			t.Logf("Level update %d: level=%d peak=%d", count, level, peak)
		}
	})

	// Start monitoring
	err := am.Start()
	if err != nil {
		t.Skipf("Failed to start audio monitor (may need audio permissions): %v", err)
	}

	if !am.IsRunning() {
		t.Error("Monitor should be running after Start()")
	}

	// Wait a bit for level updates
	time.Sleep(500 * time.Millisecond)

	finalCount := levelUpdates.Load()
	t.Logf("Received %d level updates in 500ms", finalCount)

	// Stop monitoring
	am.Stop()

	if am.IsRunning() {
		t.Error("Monitor should not be running after Stop()")
	}

	// Should have received some updates (ebur128 outputs ~10 times per second)
	if finalCount == 0 {
		t.Log("Warning: No level updates received (microphone may be muted or silent)")
	}
}
