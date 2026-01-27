// Package recording provides screen recording functionality for FeCIM Lattice Tools.
// It wraps FFmpeg for video encoding with support for various quality presets,
// formats, and buffer pooling for efficient memory usage.
package recording

import (
	"errors"
	"time"
)

// =============================================================================
// Recording State
// =============================================================================

// State represents the current state of the recording manager.
type State int

const (
	StateIdle State = iota
	StateRecording
	StatePaused
	StateStopped
)

// String returns a human-readable representation of the state.
func (s State) String() string {
	switch s {
	case StateIdle:
		return "Idle"
	case StateRecording:
		return "Recording"
	case StatePaused:
		return "Paused"
	case StateStopped:
		return "Stopped"
	default:
		return "Unknown"
	}
}

// IsActive returns true if recording is in progress (including paused).
func (s State) IsActive() bool {
	return s == StateRecording || s == StatePaused
}

// =============================================================================
// Quality Presets
// =============================================================================

// QualityPreset represents predefined quality settings.
type QualityPreset int

const (
	QualityLow QualityPreset = iota
	QualityMedium
	QualityHigh
	QualityCustom
)

// String returns a human-readable representation of the quality preset.
func (q QualityPreset) String() string {
	switch q {
	case QualityLow:
		return "Low"
	case QualityMedium:
		return "Medium"
	case QualityHigh:
		return "High"
	case QualityCustom:
		return "Custom"
	default:
		return "Unknown"
	}
}

// =============================================================================
// Video Format
// =============================================================================

// VideoFormat represents the output video format.
type VideoFormat string

const (
	FormatMP4  VideoFormat = "mp4"
	FormatWebM VideoFormat = "webm"
	FormatGIF  VideoFormat = "gif"
)

// IsValid returns true if the format is a valid, known video format.
// This prevents command injection via arbitrary format values.
func (f VideoFormat) IsValid() bool {
	switch f {
	case FormatMP4, FormatWebM, FormatGIF:
		return true
	default:
		return false
	}
}

// Extension returns the file extension for the format (with dot).
// Returns ".mp4" for invalid formats as a safe default.
func (f VideoFormat) Extension() string {
	if !f.IsValid() {
		return ".mp4" // Safe default for invalid formats
	}
	return "." + string(f)
}

// DefaultCodec returns the default codec for the format.
func (f VideoFormat) DefaultCodec() VideoCodec {
	switch f {
	case FormatMP4:
		return CodecH264
	case FormatWebM:
		return CodecVP9
	case FormatGIF:
		return CodecGIF
	default:
		return CodecH264
	}
}

// MIMEType returns the MIME type for the format.
func (f VideoFormat) MIMEType() string {
	switch f {
	case FormatMP4:
		return "video/mp4"
	case FormatWebM:
		return "video/webm"
	case FormatGIF:
		return "image/gif"
	default:
		return "video/mp4"
	}
}

// SupportsAudio returns true if the format supports audio.
func (f VideoFormat) SupportsAudio() bool {
	return f == FormatMP4 || f == FormatWebM
}

// SupportsAlpha returns true if the format supports alpha transparency.
func (f VideoFormat) SupportsAlpha() bool {
	return f == FormatWebM
}

// =============================================================================
// Video Codec
// =============================================================================

// VideoCodec represents a video codec.
type VideoCodec string

const (
	CodecH264 VideoCodec = "libx264"
	CodecH265 VideoCodec = "libx265"
	CodecVP8  VideoCodec = "libvpx"
	CodecVP9  VideoCodec = "libvpx-vp9"
	CodecGIF  VideoCodec = "gif"
)

// =============================================================================
// Encoding Preset
// =============================================================================

// EncodingPreset represents FFmpeg encoding speed presets.
type EncodingPreset string

const (
	PresetUltrafast EncodingPreset = "ultrafast"
	PresetSuperfast EncodingPreset = "superfast"
	PresetVeryfast  EncodingPreset = "veryfast"
	PresetFaster    EncodingPreset = "faster"
	PresetFast      EncodingPreset = "fast"
	PresetMedium    EncodingPreset = "medium"
	PresetSlow      EncodingPreset = "slow"
	PresetSlower    EncodingPreset = "slower"
	PresetVeryslow  EncodingPreset = "veryslow"
)

// String returns the preset as a string.
func (p EncodingPreset) String() string {
	return string(p)
}

// IsValid returns true if the preset is a valid FFmpeg preset.
func (p EncodingPreset) IsValid() bool {
	switch p {
	case PresetUltrafast, PresetSuperfast, PresetVeryfast, PresetFaster,
		PresetFast, PresetMedium, PresetSlow, PresetSlower, PresetVeryslow:
		return true
	default:
		return false
	}
}

// =============================================================================
// FFmpeg Version
// =============================================================================

// FFmpegVersion represents a parsed FFmpeg version.
type FFmpegVersion struct {
	Major int
	Minor int
	Patch int
}

// String returns the version as a string (e.g., "6.1.1").
func (v FFmpegVersion) String() string {
	if v.Patch == 0 {
		return formatVersion(v.Major, v.Minor)
	}
	return formatVersionFull(v.Major, v.Minor, v.Patch)
}

// Compare compares two versions. Returns -1 if v < other, 0 if equal, 1 if v > other.
func (v FFmpegVersion) Compare(other FFmpegVersion) int {
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}
	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}
	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}
	return 0
}

// AtLeast returns true if the version is at least major.minor.
func (v FFmpegVersion) AtLeast(major, minor int) bool {
	if v.Major > major {
		return true
	}
	if v.Major < major {
		return false
	}
	return v.Minor >= minor
}

func formatVersion(major, minor int) string {
	return itoa(major) + "." + itoa(minor)
}

func formatVersionFull(major, minor, patch int) string {
	return itoa(major) + "." + itoa(minor) + "." + itoa(patch)
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	if i < 0 {
		return "-" + itoa(-i)
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	return string(buf[pos:])
}

// =============================================================================
// FFmpeg Info
// =============================================================================

// FFmpegInfo contains information about the installed FFmpeg.
type FFmpegInfo struct {
	Path     string
	Version  FFmpegVersion
	Encoders []string
	Formats  []string
}

// HasEncoder returns true if the encoder is available.
func (i *FFmpegInfo) HasEncoder(name string) bool {
	for _, enc := range i.Encoders {
		if enc == name {
			return true
		}
	}
	return false
}

// HasFormat returns true if the format is available.
func (i *FFmpegInfo) HasFormat(name string) bool {
	for _, fmt := range i.Formats {
		if fmt == name {
			return true
		}
	}
	return false
}

// CanRecord returns true if FFmpeg has the required capabilities for recording.
func (i *FFmpegInfo) CanRecord() bool {
	// Basic check - if we have FFmpeg path, we can attempt recording
	return i.Path != ""
}

// =============================================================================
// Buffer Pool Stats
// =============================================================================

// BufferPoolStats contains statistics about buffer pool usage.
type BufferPoolStats struct {
	BufferSize int
	Gets       int64
	Puts       int64
	Hits       int64
	Misses     int64
}

// =============================================================================
// Validation Error
// =============================================================================

// ValidationError represents a validation error for a specific field.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// AsValidationError attempts to convert an error to a ValidationError.
func AsValidationError(err error, target **ValidationError) bool {
	return errors.As(err, target)
}

// =============================================================================
// Capture Source Interface
// =============================================================================

// CaptureSource is the interface for something that can provide frames.
type CaptureSource interface {
	Capture() ([]byte, error)
	Width() int
	Height() int
}

// =============================================================================
// Manager Interface
// =============================================================================

// Manager is the interface for the recording manager.
type Manager interface {
	Start(source CaptureSource) error
	Stop() (outputFile string, err error)
	Pause() error
	Resume() error
	State() State
	IsRecording() bool
	IsPaused() bool
	ElapsedTime() time.Duration
	EstimatedFileSize() int64
	FramesCaptured() int
	FramesDropped() int
	SetSettings(settings Settings)
	GetSettings() Settings
	OnStateChange(callback func(State))
	OnFrameCaptured(callback func(frameNum int))
	OnError(callback func(error))
}
