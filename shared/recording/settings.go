package recording

import (
	"encoding/json"
	"os"
	"strings"
)

// Settings contains all recording configuration options.
type Settings struct {
	Quality      QualityPreset  `json:"quality"`
	FPS          int            `json:"fps"`
	Format       VideoFormat    `json:"format"`
	CRF          int            `json:"crf"`
	Preset       EncodingPreset `json:"preset"`
	AudioEnabled bool           `json:"audio_enabled"` // Deprecated: use Audio.Enabled
	Audio        AudioSettings  `json:"audio"`         // Audio recording settings
}

// DefaultSettings returns the default recording settings.
func DefaultSettings() Settings {
	return Settings{
		Quality:      QualityMedium,
		FPS:          20,
		Format:       FormatMP4,
		CRF:          23,
		Preset:       PresetUltrafast,
		AudioEnabled: false,
		Audio:        DefaultAudioSettings(),
	}
}

// DefaultSettingsWithFormat returns default settings for a specific format.
func DefaultSettingsWithFormat(format VideoFormat) Settings {
	s := DefaultSettings()
	s.Format = format
	return s
}

// NewSettingsFromPreset creates settings based on a quality preset.
func NewSettingsFromPreset(preset QualityPreset) Settings {
	s := DefaultSettings()
	s.Quality = preset

	switch preset {
	case QualityLow:
		s.FPS = 10
		s.CRF = 28
		s.Preset = PresetUltrafast
	case QualityMedium:
		s.FPS = 20
		s.CRF = 23
		s.Preset = PresetFast
	case QualityHigh:
		s.FPS = 30
		s.CRF = 18
		s.Preset = PresetMedium
	}

	return s
}

// Validate checks if the settings are valid.
// This includes security validation to prevent command injection.
func (s Settings) Validate() error {
	// Validate Format (CRITICAL: prevents command injection)
	if !s.Format.IsValid() {
		return &ValidationError{
			Field:   "Format",
			Message: "Format must be one of: mp4, webm, gif",
		}
	}

	// Validate Preset (prevents invalid FFmpeg arguments)
	if !s.Preset.IsValid() {
		return &ValidationError{
			Field:   "Preset",
			Message: "Preset must be a valid FFmpeg encoding preset",
		}
	}

	// Validate CRF (1-51 for H.264)
	if s.CRF < 1 || s.CRF > 51 {
		return &ValidationError{
			Field:   "CRF",
			Message: "CRF must be between 1 and 51",
		}
	}

	// Validate FPS (1-60)
	if s.FPS < 1 || s.FPS > 60 {
		return &ValidationError{
			Field:   "FPS",
			Message: "FPS must be between 1 and 60",
		}
	}

	// Validate audio settings if enabled
	if s.Audio.Enabled || s.AudioEnabled {
		if err := s.Audio.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Copy returns a copy of the settings.
func (s Settings) Copy() Settings {
	return Settings{
		Quality:      s.Quality,
		FPS:          s.FPS,
		Format:       s.Format,
		CRF:          s.CRF,
		Preset:       s.Preset,
		AudioEnabled: s.AudioEnabled,
		Audio:        s.Audio,
	}
}

// ApplyPreset applies a quality preset to the settings.
func (s *Settings) ApplyPreset(preset QualityPreset) {
	newSettings := NewSettingsFromPreset(preset)
	s.Quality = newSettings.Quality
	s.FPS = newSettings.FPS
	s.CRF = newSettings.CRF
	s.Preset = newSettings.Preset
}

// SaveToFile saves settings to a JSON file.
func (s Settings) SaveToFile(path string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// LoadSettingsFromFile loads settings from a JSON file.
func LoadSettingsFromFile(path string) (Settings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Settings{}, err
	}

	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return Settings{}, err
	}

	return s, nil
}

// EstimatedBitrate returns the estimated bitrate in bits per second.
func (s Settings) EstimatedBitrate(width, height int) int64 {
	// Rough estimation based on CRF, resolution, and FPS
	// Lower CRF = higher quality = higher bitrate
	pixels := int64(width * height)
	fps := int64(s.FPS)

	// Base bitrate calculation for 1080p at CRF 23 is roughly 4-8 Mbps
	// We use a reference of 5 Mbps for 1920x1080 at 30fps and CRF 23
	referencePixels := int64(1920 * 1080)
	referenceFPS := int64(30)
	referenceBitrate := int64(5_000_000) // 5 Mbps

	// Scale by resolution
	resolutionFactor := float64(pixels) / float64(referencePixels)

	// Scale by FPS
	fpsFactor := float64(fps) / float64(referenceFPS)

	// Adjust for CRF (each CRF point roughly changes bitrate by ~6%)
	// CRF 18 = ~2x bitrate, CRF 28 = ~0.5x bitrate relative to CRF 23
	crfDiff := 23 - s.CRF                        // Positive means higher quality
	crfFactor := 1.0 + (float64(crfDiff) * 0.12) // ~12% per CRF point
	if crfFactor < 0.1 {
		crfFactor = 0.1
	}

	return int64(float64(referenceBitrate) * resolutionFactor * fpsFactor * crfFactor)
}

// EstimatedFileSize returns the estimated file size in bytes for a given duration.
func (s Settings) EstimatedFileSize(width, height int, durationSeconds int) int64 {
	bitrate := s.EstimatedBitrate(width, height)
	// Convert bits per second to bytes for duration
	return (bitrate * int64(durationSeconds)) / 8
}

// QualityPresetFromString parses a quality preset from a string.
func QualityPresetFromString(s string) QualityPreset {
	switch strings.ToLower(s) {
	case "low":
		return QualityLow
	case "medium":
		return QualityMedium
	case "high":
		return QualityHigh
	case "custom":
		return QualityCustom
	default:
		return QualityMedium
	}
}

// VideoFormatFromString parses a video format from a string.
func VideoFormatFromString(s string) VideoFormat {
	s = strings.TrimPrefix(strings.ToLower(s), ".")
	switch s {
	case "mp4":
		return FormatMP4
	case "webm":
		return FormatWebM
	case "gif":
		return FormatGIF
	default:
		return FormatMP4
	}
}

// EncodingPresetFromString parses an encoding preset from a string.
func EncodingPresetFromString(s string) EncodingPreset {
	preset := EncodingPreset(strings.ToLower(s))
	if preset.IsValid() {
		return preset
	}
	return PresetMedium
}
