package recording

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// =============================================================================
// Quality Preset Tests
// =============================================================================

func TestQualityPresetValues(t *testing.T) {
	tests := []struct {
		name    string
		preset  QualityPreset
		wantCRF int
		wantFPS int
	}{
		{"Low quality", QualityLow, 28, 10},
		{"Medium quality", QualityMedium, 23, 20},
		{"High quality", QualityHigh, 18, 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := NewSettingsFromPreset(tt.preset)

			if settings.CRF != tt.wantCRF {
				t.Errorf("CRF = %d, want %d", settings.CRF, tt.wantCRF)
			}

			if settings.FPS != tt.wantFPS {
				t.Errorf("FPS = %d, want %d", settings.FPS, tt.wantFPS)
			}

			if settings.Quality != tt.preset {
				t.Errorf("Quality = %v, want %v", settings.Quality, tt.preset)
			}
		})
	}
}

func TestQualityPresetString(t *testing.T) {
	tests := []struct {
		preset   QualityPreset
		expected string
	}{
		{QualityLow, "Low"},
		{QualityMedium, "Medium"},
		{QualityHigh, "High"},
		{QualityCustom, "Custom"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.preset.String() != tt.expected {
				t.Errorf("String() = %s, want %s", tt.preset.String(), tt.expected)
			}
		})
	}
}

func TestQualityPresetFromString(t *testing.T) {
	tests := []struct {
		input    string
		expected QualityPreset
	}{
		{"Low", QualityLow},
		{"low", QualityLow},
		{"LOW", QualityLow},
		{"Medium", QualityMedium},
		{"High", QualityHigh},
		{"Custom", QualityCustom},
		{"invalid", QualityMedium}, // Default to medium
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := QualityPresetFromString(tt.input)
			if result != tt.expected {
				t.Errorf("QualityPresetFromString(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// Video Format Tests
// =============================================================================

func TestVideoFormatValues(t *testing.T) {
	tests := []struct {
		format        VideoFormat
		expectedExt   string
		expectedCodec VideoCodec
		expectedMIME  string
		supportsAudio bool
		supportsAlpha bool
	}{
		{FormatMP4, ".mp4", CodecH264, "video/mp4", true, false},
		{FormatWebM, ".webm", CodecVP9, "video/webm", true, true},
		{FormatGIF, ".gif", CodecGIF, "image/gif", false, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			if tt.format.Extension() != tt.expectedExt {
				t.Errorf("Extension() = %s, want %s", tt.format.Extension(), tt.expectedExt)
			}

			if tt.format.DefaultCodec() != tt.expectedCodec {
				t.Errorf("DefaultCodec() = %v, want %v", tt.format.DefaultCodec(), tt.expectedCodec)
			}

			if tt.format.MIMEType() != tt.expectedMIME {
				t.Errorf("MIMEType() = %s, want %s", tt.format.MIMEType(), tt.expectedMIME)
			}

			if tt.format.SupportsAudio() != tt.supportsAudio {
				t.Errorf("SupportsAudio() = %v, want %v", tt.format.SupportsAudio(), tt.supportsAudio)
			}

			if tt.format.SupportsAlpha() != tt.supportsAlpha {
				t.Errorf("SupportsAlpha() = %v, want %v", tt.format.SupportsAlpha(), tt.supportsAlpha)
			}
		})
	}
}

func TestVideoFormatFromString(t *testing.T) {
	tests := []struct {
		input    string
		expected VideoFormat
	}{
		{"mp4", FormatMP4},
		{"MP4", FormatMP4},
		{".mp4", FormatMP4},
		{"webm", FormatWebM},
		{"gif", FormatGIF},
		{"invalid", FormatMP4}, // Default to MP4
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := VideoFormatFromString(tt.input)
			if result != tt.expected {
				t.Errorf("VideoFormatFromString(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// Settings Validation Tests
// =============================================================================

func TestSettingsValidation(t *testing.T) {
	tests := []struct {
		name        string
		settings    Settings
		expectError bool
		errorField  string
	}{
		{
			name:        "valid settings",
			settings:    DefaultSettings(),
			expectError: false,
		},
		{
			name: "valid custom CRF",
			settings: Settings{
				Quality: QualityCustom,
				FPS:     20,
				Format:  FormatMP4,
				CRF:     20,
				Preset:  PresetMedium,
			},
			expectError: false,
		},
		{
			name: "CRF too low",
			settings: Settings{
				Quality: QualityCustom,
				FPS:     20,
				Format:  FormatMP4,
				CRF:     0,
				Preset:  PresetMedium,
			},
			expectError: true,
			errorField:  "CRF",
		},
		{
			name: "CRF too high",
			settings: Settings{
				Quality: QualityCustom,
				FPS:     20,
				Format:  FormatMP4,
				CRF:     52, // 52 is out of H.264 CRF range (0-51)
				Preset:  PresetMedium,
			},
			expectError: true,
			errorField:  "CRF",
		},
		{
			name: "FPS too low",
			settings: Settings{
				Quality: QualityMedium,
				FPS:     0,
				Format:  FormatMP4,
				CRF:     23,
				Preset:  PresetMedium,
			},
			expectError: true,
			errorField:  "FPS",
		},
		{
			name: "FPS too high",
			settings: Settings{
				Quality: QualityMedium,
				FPS:     120,
				Format:  FormatMP4,
				CRF:     23,
				Preset:  PresetMedium,
			},
			expectError: true,
			errorField:  "FPS",
		},
		{
			name: "valid FPS values",
			settings: Settings{
				Quality: QualityMedium,
				FPS:     60,
				Format:  FormatMP4,
				CRF:     23,
				Preset:  PresetMedium,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.settings.Validate()

			if tt.expectError {
				if err == nil {
					t.Error("Expected validation error but got nil")
				} else if tt.errorField != "" {
					var verr *ValidationError
					if !AsValidationError(err, &verr) {
						t.Errorf("Expected ValidationError, got %T", err)
					} else if verr.Field != tt.errorField {
						t.Errorf("Expected error on field %s, got %s", tt.errorField, verr.Field)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestSettingsValidFPSValues(t *testing.T) {
	validFPS := []int{10, 15, 20, 24, 25, 30, 50, 60}

	for _, fps := range validFPS {
		t.Run("FPS_"+string(rune(fps)), func(t *testing.T) {
			settings := DefaultSettings()
			settings.FPS = fps

			err := settings.Validate()
			if err != nil {
				t.Errorf("FPS %d should be valid, got error: %v", fps, err)
			}
		})
	}
}

func TestSettingsCRFRange(t *testing.T) {
	// Valid range is 1-51 for H.264
	validCRF := []int{1, 18, 23, 28, 51}

	for _, crf := range validCRF {
		settings := Settings{
			Quality: QualityCustom,
			FPS:     20,
			Format:  FormatMP4,
			CRF:     crf,
			Preset:  PresetMedium,
		}

		err := settings.Validate()
		if err != nil {
			t.Errorf("CRF %d should be valid, got error: %v", crf, err)
		}
	}
}

// =============================================================================
// Settings Persistence Tests
// =============================================================================

func TestSettingsSaveAndLoad(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "recording_settings.json")

	original := Settings{
		Quality:      QualityHigh,
		FPS:          30,
		Format:       FormatWebM,
		CRF:          18,
		Preset:       PresetSlow,
		AudioEnabled: true,
	}

	// Save settings
	err := original.SaveToFile(configPath)
	if err != nil {
		t.Fatalf("Failed to save settings: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Settings file was not created")
	}

	// Load settings
	loaded, err := LoadSettingsFromFile(configPath)
	if err != nil {
		t.Fatalf("Failed to load settings: %v", err)
	}

	// Compare
	if loaded.Quality != original.Quality {
		t.Errorf("Quality = %v, want %v", loaded.Quality, original.Quality)
	}
	if loaded.FPS != original.FPS {
		t.Errorf("FPS = %d, want %d", loaded.FPS, original.FPS)
	}
	if loaded.Format != original.Format {
		t.Errorf("Format = %v, want %v", loaded.Format, original.Format)
	}
	if loaded.CRF != original.CRF {
		t.Errorf("CRF = %d, want %d", loaded.CRF, original.CRF)
	}
	if loaded.Preset != original.Preset {
		t.Errorf("Preset = %v, want %v", loaded.Preset, original.Preset)
	}
	if loaded.AudioEnabled != original.AudioEnabled {
		t.Errorf("AudioEnabled = %v, want %v", loaded.AudioEnabled, original.AudioEnabled)
	}
}

func TestLoadSettingsFromNonexistentFile(t *testing.T) {
	_, err := LoadSettingsFromFile("/nonexistent/path/settings.json")

	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestLoadSettingsFromInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.json")

	// Write invalid JSON
	err := os.WriteFile(configPath, []byte("not valid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err = LoadSettingsFromFile(configPath)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestSettingsJSONMarshaling(t *testing.T) {
	original := Settings{
		Quality:      QualityHigh,
		FPS:          30,
		Format:       FormatMP4,
		CRF:          18,
		Preset:       PresetMedium,
		AudioEnabled: true,
	}

	// Marshal
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal
	var loaded Settings
	err = json.Unmarshal(data, &loaded)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Compare
	if loaded != original {
		t.Errorf("Unmarshaled settings don't match: got %+v, want %+v", loaded, original)
	}
}

// =============================================================================
// Default Settings Tests
// =============================================================================

func TestDefaultSettings(t *testing.T) {
	settings := DefaultSettings()

	// Should be valid
	if err := settings.Validate(); err != nil {
		t.Errorf("Default settings should be valid: %v", err)
	}

	// Should use medium quality by default
	if settings.Quality != QualityMedium {
		t.Errorf("Default quality = %v, want %v", settings.Quality, QualityMedium)
	}

	// Should use MP4 format by default
	if settings.Format != FormatMP4 {
		t.Errorf("Default format = %v, want %v", settings.Format, FormatMP4)
	}

	// Should have reasonable defaults
	if settings.FPS < 10 || settings.FPS > 60 {
		t.Errorf("Default FPS %d is outside reasonable range [10, 60]", settings.FPS)
	}

	if settings.CRF < 18 || settings.CRF > 28 {
		t.Errorf("Default CRF %d is outside reasonable range [18, 28]", settings.CRF)
	}
}

func TestDefaultSettingsWithFormat(t *testing.T) {
	formats := []VideoFormat{FormatMP4, FormatWebM, FormatGIF}

	for _, format := range formats {
		t.Run(string(format), func(t *testing.T) {
			settings := DefaultSettingsWithFormat(format)

			if settings.Format != format {
				t.Errorf("Format = %v, want %v", settings.Format, format)
			}

			if err := settings.Validate(); err != nil {
				t.Errorf("Default settings for %v should be valid: %v", format, err)
			}
		})
	}
}

// =============================================================================
// Preset Configuration Tests
// =============================================================================

func TestPresetEnumeration(t *testing.T) {
	presets := []EncodingPreset{
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

	for _, preset := range presets {
		t.Run(string(preset), func(t *testing.T) {
			// Should convert to valid string
			str := preset.String()
			if str == "" {
				t.Error("Preset string should not be empty")
			}

			// Should be valid for FFmpeg
			if !preset.IsValid() {
				t.Error("Preset should be valid")
			}
		})
	}
}

func TestPresetFromString(t *testing.T) {
	tests := []struct {
		input    string
		expected EncodingPreset
	}{
		{"ultrafast", PresetUltrafast},
		{"ULTRAFAST", PresetUltrafast},
		{"medium", PresetMedium},
		{"slow", PresetSlow},
		{"invalid", PresetMedium}, // Default to medium
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := EncodingPresetFromString(tt.input)
			if result != tt.expected {
				t.Errorf("EncodingPresetFromString(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// Settings Copy/Clone Tests
// =============================================================================

func TestSettingsCopy(t *testing.T) {
	original := Settings{
		Quality:      QualityHigh,
		FPS:          30,
		Format:       FormatMP4,
		CRF:          18,
		Preset:       PresetSlow,
		AudioEnabled: true,
	}

	copy := original.Copy()

	// Modify original
	original.FPS = 60
	original.Quality = QualityLow

	// Copy should be unchanged
	if copy.FPS != 30 {
		t.Error("Copy FPS was modified when original changed")
	}

	if copy.Quality != QualityHigh {
		t.Error("Copy Quality was modified when original changed")
	}
}

// =============================================================================
// Settings Application Tests
// =============================================================================

func TestSettingsApplyPreset(t *testing.T) {
	settings := DefaultSettings()
	settings.Quality = QualityCustom
	settings.FPS = 60
	settings.CRF = 10

	// Apply a preset
	settings.ApplyPreset(QualityLow)

	// Should override custom values
	if settings.Quality != QualityLow {
		t.Errorf("Quality = %v, want %v", settings.Quality, QualityLow)
	}

	// Preset values should be applied
	lowSettings := NewSettingsFromPreset(QualityLow)
	if settings.FPS != lowSettings.FPS {
		t.Errorf("FPS = %d, want %d", settings.FPS, lowSettings.FPS)
	}
	if settings.CRF != lowSettings.CRF {
		t.Errorf("CRF = %d, want %d", settings.CRF, lowSettings.CRF)
	}
}

// =============================================================================
// File Size Estimation Tests
// =============================================================================

func TestEstimatedBitrate(t *testing.T) {
	tests := []struct {
		name     string
		settings Settings
		width    int
		height   int
		minBps   int64 // Minimum expected bits per second
		maxBps   int64 // Maximum expected bits per second
	}{
		{
			name:     "720p medium quality",
			settings: NewSettingsFromPreset(QualityMedium),
			width:    1280,
			height:   720,
			minBps:   500_000,   // 500 Kbps minimum
			maxBps:   5_000_000, // 5 Mbps maximum
		},
		{
			name:     "1080p high quality",
			settings: NewSettingsFromPreset(QualityHigh),
			width:    1920,
			height:   1080,
			minBps:   1_000_000,  // 1 Mbps minimum
			maxBps:   15_000_000, // 15 Mbps maximum
		},
		{
			name:     "480p low quality",
			settings: NewSettingsFromPreset(QualityLow),
			width:    640,
			height:   480,
			minBps:   50_000,    // 50 Kbps minimum (low quality at low resolution)
			maxBps:   2_000_000, // 2 Mbps maximum
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bitrate := tt.settings.EstimatedBitrate(tt.width, tt.height)

			if bitrate < tt.minBps {
				t.Errorf("Bitrate %d bps is below minimum %d bps", bitrate, tt.minBps)
			}

			if bitrate > tt.maxBps {
				t.Errorf("Bitrate %d bps is above maximum %d bps", bitrate, tt.maxBps)
			}
		})
	}
}

func TestEstimatedFileSize(t *testing.T) {
	settings := NewSettingsFromPreset(QualityMedium)

	// Estimate for 1 minute at 720p
	sizeBytes := settings.EstimatedFileSize(1280, 720, 60) // 60 seconds

	// Should be reasonable (between 1MB and 100MB for 1 minute)
	if sizeBytes < 1_000_000 {
		t.Errorf("Estimated size %d bytes is too small for 1 minute video", sizeBytes)
	}

	if sizeBytes > 100_000_000 {
		t.Errorf("Estimated size %d bytes is too large for 1 minute video", sizeBytes)
	}
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkSettingsValidation(b *testing.B) {
	settings := DefaultSettings()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		settings.Validate()
	}
}

func BenchmarkSettingsCopy(b *testing.B) {
	settings := DefaultSettings()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = settings.Copy()
	}
}

func BenchmarkSettingsJSONMarshal(b *testing.B) {
	settings := DefaultSettings()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(settings)
	}
}
