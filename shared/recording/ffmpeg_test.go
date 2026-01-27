package recording

import (
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

// =============================================================================
// FFmpeg Detection Tests
// =============================================================================

func TestDetectFFmpeg(t *testing.T) {
	info, err := DetectFFmpeg()

	// FFmpeg may or may not be installed - test both scenarios
	if err != nil {
		// If not installed, error should be clear
		if !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "not installed") {
			t.Errorf("Expected 'not found' or 'not installed' error, got: %v", err)
		}
		t.Skip("FFmpeg not installed, skipping detection tests")
	}

	// If installed, info should be populated
	if info == nil {
		t.Fatal("FFmpegInfo should not be nil when FFmpeg is found")
	}

	if info.Path == "" {
		t.Error("FFmpegInfo.Path should not be empty")
	}

	if info.Version.String() == "" && info.Version.Major == 0 {
		t.Error("FFmpegInfo.Version should not be empty")
	}
}

func TestDetectFFmpegPath(t *testing.T) {
	path, err := DetectFFmpegPath()

	if err != nil {
		t.Skip("FFmpeg not installed")
	}

	if path == "" {
		t.Error("Path should not be empty when FFmpeg is found")
	}

	// Verify the path is executable
	cmd := exec.Command(path, "-version")
	if err := cmd.Run(); err != nil {
		t.Errorf("FFmpeg at path %s is not executable: %v", path, err)
	}
}

func TestIsFFmpegAvailable(t *testing.T) {
	available := IsFFmpegAvailable()

	// Verify consistency with DetectFFmpeg
	_, err := DetectFFmpeg()
	expectedAvailable := err == nil

	if available != expectedAvailable {
		t.Errorf("IsFFmpegAvailable() = %v, but DetectFFmpeg error = %v", available, err)
	}
}

// =============================================================================
// Version Parsing Tests
// =============================================================================

func TestParseFFmpegVersion(t *testing.T) {
	tests := []struct {
		name           string
		versionOutput  string
		expectedMajor  int
		expectedMinor  int
		expectedPatch  int
		expectedString string
		expectError    bool
	}{
		{
			name: "standard version format",
			versionOutput: `ffmpeg version 6.1.1 Copyright (c) 2000-2023 the FFmpeg developers
built with gcc 13.2.0`,
			expectedMajor:  6,
			expectedMinor:  1,
			expectedPatch:  1,
			expectedString: "6.1.1",
			expectError:    false,
		},
		{
			name: "version with git hash",
			versionOutput: `ffmpeg version n5.1.2-3-g12345abc Copyright (c) 2000-2023 the FFmpeg developers
built with gcc 12.1.0`,
			expectedMajor:  5,
			expectedMinor:  1,
			expectedPatch:  2,
			expectedString: "5.1.2",
			expectError:    false,
		},
		{
			name: "ubuntu/debian version format",
			versionOutput: `ffmpeg version 4.4.2-0ubuntu0.22.04.1 Copyright (c) 2000-2021 the FFmpeg developers
built with gcc 11 (Ubuntu 11.2.0-19ubuntu1)`,
			expectedMajor:  4,
			expectedMinor:  4,
			expectedPatch:  2,
			expectedString: "4.4.2",
			expectError:    false,
		},
		{
			name: "version without patch",
			versionOutput: `ffmpeg version 7.0 Copyright (c) 2000-2024 the FFmpeg developers
built with clang`,
			expectedMajor:  7,
			expectedMinor:  0,
			expectedPatch:  0,
			expectedString: "7.0",
			expectError:    false,
		},
		{
			name:          "empty output",
			versionOutput: "",
			expectError:   true,
		},
		{
			name:          "invalid format",
			versionOutput: "not a valid version string",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := ParseFFmpegVersion(tt.versionOutput)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if version.Major != tt.expectedMajor {
				t.Errorf("Major = %d, want %d", version.Major, tt.expectedMajor)
			}
			if version.Minor != tt.expectedMinor {
				t.Errorf("Minor = %d, want %d", version.Minor, tt.expectedMinor)
			}
			if version.Patch != tt.expectedPatch {
				t.Errorf("Patch = %d, want %d", version.Patch, tt.expectedPatch)
			}
			if version.String() != tt.expectedString {
				t.Errorf("String() = %s, want %s", version.String(), tt.expectedString)
			}
		})
	}
}

func TestFFmpegVersionComparison(t *testing.T) {
	tests := []struct {
		name     string
		v1       FFmpegVersion
		v2       FFmpegVersion
		expected int // -1: v1 < v2, 0: v1 == v2, 1: v1 > v2
	}{
		{
			name:     "equal versions",
			v1:       FFmpegVersion{Major: 6, Minor: 1, Patch: 1},
			v2:       FFmpegVersion{Major: 6, Minor: 1, Patch: 1},
			expected: 0,
		},
		{
			name:     "v1 major less than v2",
			v1:       FFmpegVersion{Major: 5, Minor: 1, Patch: 1},
			v2:       FFmpegVersion{Major: 6, Minor: 0, Patch: 0},
			expected: -1,
		},
		{
			name:     "v1 major greater than v2",
			v1:       FFmpegVersion{Major: 7, Minor: 0, Patch: 0},
			v2:       FFmpegVersion{Major: 6, Minor: 9, Patch: 9},
			expected: 1,
		},
		{
			name:     "v1 minor less than v2",
			v1:       FFmpegVersion{Major: 6, Minor: 0, Patch: 5},
			v2:       FFmpegVersion{Major: 6, Minor: 1, Patch: 0},
			expected: -1,
		},
		{
			name:     "v1 patch less than v2",
			v1:       FFmpegVersion{Major: 6, Minor: 1, Patch: 0},
			v2:       FFmpegVersion{Major: 6, Minor: 1, Patch: 1},
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v1.Compare(tt.v2)
			if result != tt.expected {
				t.Errorf("Compare() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestFFmpegVersionAtLeast(t *testing.T) {
	v := FFmpegVersion{Major: 6, Minor: 1, Patch: 1}

	tests := []struct {
		name     string
		major    int
		minor    int
		expected bool
	}{
		{"exact version", 6, 1, true},
		{"lower version", 5, 0, true},
		{"lower major higher minor", 5, 9, true},
		{"higher major", 7, 0, false},
		{"same major higher minor", 6, 2, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := v.AtLeast(tt.major, tt.minor)
			if result != tt.expected {
				t.Errorf("AtLeast(%d, %d) = %v, want %v", tt.major, tt.minor, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// Codec Support Tests
// =============================================================================

func TestCheckCodecSupport(t *testing.T) {
	if !IsFFmpegAvailable() {
		t.Skip("FFmpeg not installed")
	}

	tests := []struct {
		codec    string
		expected bool // Most systems should have these
	}{
		{"libx264", true},      // Standard H.264 encoder
		{"libx265", true},      // HEVC encoder (usually available)
		{"nonexistent", false}, // Should not exist
	}

	for _, tt := range tests {
		t.Run(tt.codec, func(t *testing.T) {
			supported := CheckCodecSupport(tt.codec)
			// For common codecs, we expect them to be available
			// For nonexistent, we expect false
			if tt.codec == "nonexistent" && supported {
				t.Error("nonexistent codec should not be supported")
			}
			// Log actual support status for common codecs
			t.Logf("Codec %s supported: %v", tt.codec, supported)
		})
	}
}

func TestGetSupportedEncoders(t *testing.T) {
	if !IsFFmpegAvailable() {
		t.Skip("FFmpeg not installed")
	}

	encoders := GetSupportedEncoders()

	if len(encoders) == 0 {
		t.Error("Expected at least some encoders to be available")
	}

	// Check for common encoders
	hasH264 := false
	for _, enc := range encoders {
		if strings.Contains(enc, "264") || strings.Contains(enc, "h264") {
			hasH264 = true
			break
		}
	}

	if !hasH264 {
		t.Log("Warning: No H.264 encoder found (may be expected on some systems)")
	}
}

func TestGetSupportedFormats(t *testing.T) {
	if !IsFFmpegAvailable() {
		t.Skip("FFmpeg not installed")
	}

	formats := GetSupportedFormats()

	if len(formats) == 0 {
		t.Error("Expected at least some formats to be available")
	}

	// Check for common formats
	expectedFormats := []string{"mp4", "webm", "gif"}
	for _, expected := range expectedFormats {
		found := false
		for _, format := range formats {
			if strings.Contains(strings.ToLower(format), expected) {
				found = true
				break
			}
		}
		if !found {
			t.Logf("Format %s not found in supported formats", expected)
		}
	}
}

// =============================================================================
// Command Builder Tests
// =============================================================================

func TestNewFFmpegCommandBuilder(t *testing.T) {
	builder := NewFFmpegCommandBuilder()

	if builder == nil {
		t.Fatal("NewFFmpegCommandBuilder returned nil")
	}
}

func TestFFmpegCommandBuilderBasic(t *testing.T) {
	builder := NewFFmpegCommandBuilder().
		Input("-").
		Output("output.mp4")

	args := builder.Build()

	// Should have input and output
	hasInput := false
	hasOutput := false
	for i, arg := range args {
		if arg == "-i" && i+1 < len(args) && args[i+1] == "-" {
			hasInput = true
		}
		if arg == "output.mp4" && i == len(args)-1 {
			hasOutput = true
		}
	}

	if !hasInput {
		t.Error("Command should have input '-i -'")
	}
	if !hasOutput {
		t.Error("Command should have output as last argument")
	}
}

func TestFFmpegCommandBuilderWithSettings(t *testing.T) {
	settings := Settings{
		Quality: QualityHigh,
		FPS:     30,
		Format:  FormatMP4,
		CRF:     18,
		Preset:  PresetMedium,
	}

	builder := NewFFmpegCommandBuilder().
		WithSettings(settings).
		InputFormat("rawvideo").
		InputPixelFormat("rgb24").
		VideoSize(1920, 1080).
		Input("-").
		Output("test.mp4")

	args := builder.Build()
	argsStr := strings.Join(args, " ")

	// Check that settings are applied
	if !strings.Contains(argsStr, "-r 30") && !strings.Contains(argsStr, "-framerate 30") {
		t.Errorf("Expected framerate 30, got: %s", argsStr)
	}

	if !strings.Contains(argsStr, "-crf 18") {
		t.Errorf("Expected CRF 18, got: %s", argsStr)
	}

	if !strings.Contains(argsStr, "-preset medium") {
		t.Errorf("Expected preset medium, got: %s", argsStr)
	}
}

func TestFFmpegCommandBuilderOverwrite(t *testing.T) {
	builder := NewFFmpegCommandBuilder().
		Overwrite(true).
		Input("input.mp4").
		Output("output.mp4")

	args := builder.Build()
	argsStr := strings.Join(args, " ")

	if !strings.Contains(argsStr, "-y") {
		t.Error("Overwrite flag (-y) not found in command")
	}
}

func TestFFmpegCommandBuilderVideoCodec(t *testing.T) {
	tests := []struct {
		name     string
		codec    VideoCodec
		expected string
	}{
		{"H264", CodecH264, "libx264"},
		{"H265", CodecH265, "libx265"},
		{"VP8", CodecVP8, "libvpx"},
		{"VP9", CodecVP9, "libvpx-vp9"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewFFmpegCommandBuilder().
				VideoCodec(tt.codec).
				Input("input.raw").
				Output("output.mp4")

			args := builder.Build()
			argsStr := strings.Join(args, " ")

			if !strings.Contains(argsStr, "-c:v "+tt.expected) {
				t.Errorf("Expected codec %s, got: %s", tt.expected, argsStr)
			}
		})
	}
}

func TestFFmpegCommandBuilderPixelFormat(t *testing.T) {
	builder := NewFFmpegCommandBuilder().
		OutputPixelFormat("yuv420p").
		Input("input.raw").
		Output("output.mp4")

	args := builder.Build()
	argsStr := strings.Join(args, " ")

	if !strings.Contains(argsStr, "-pix_fmt yuv420p") {
		t.Errorf("Expected pixel format yuv420p, got: %s", argsStr)
	}
}

func TestFFmpegCommandBuilderForRecording(t *testing.T) {
	settings := Settings{
		Quality: QualityMedium,
		FPS:     20,
		Format:  FormatMP4,
		CRF:     23,
		Preset:  PresetUltrafast,
	}

	builder := NewFFmpegCommandBuilder().
		ForRecording(1920, 1080, settings).
		Output("recording.mp4")

	args := builder.Build()
	argsStr := strings.Join(args, " ")

	// Check essential recording parameters
	requiredParams := []string{
		"-f rawvideo",
		"-pixel_format rgb24",
		"1920x1080",
		"-framerate 20",
		"-i -",
		"-c:v libx264",
		"-preset ultrafast",
		"-crf 23",
		"-pix_fmt yuv420p",
	}

	for _, param := range requiredParams {
		if !strings.Contains(argsStr, param) {
			t.Errorf("Missing parameter: %s in command: %s", param, argsStr)
		}
	}
}

func TestFFmpegCommandBuilderFormatSpecificSettings(t *testing.T) {
	tests := []struct {
		name     string
		format   VideoFormat
		expected []string
	}{
		{
			name:   "MP4 format",
			format: FormatMP4,
			expected: []string{
				"-c:v libx264",
				"-pix_fmt yuv420p",
			},
		},
		{
			name:   "WebM format",
			format: FormatWebM,
			expected: []string{
				"-c:v libvpx",
			},
		},
		{
			name:   "GIF format",
			format: FormatGIF,
			expected: []string{
				"gif",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := Settings{Format: tt.format, FPS: 20, CRF: 23, Preset: PresetFast}
			builder := NewFFmpegCommandBuilder().
				ForRecording(640, 480, settings).
				Output("test." + string(tt.format))

			args := builder.Build()
			argsStr := strings.Join(args, " ")

			for _, exp := range tt.expected {
				if !strings.Contains(argsStr, exp) {
					t.Errorf("Expected %s for format %s, got: %s", exp, tt.format, argsStr)
				}
			}
		})
	}
}

// =============================================================================
// FFmpeg Info Tests
// =============================================================================

func TestFFmpegInfoHasCapability(t *testing.T) {
	info := &FFmpegInfo{
		Path:    "/usr/bin/ffmpeg",
		Version: FFmpegVersion{Major: 6, Minor: 1, Patch: 0},
		Encoders: []string{
			"libx264",
			"libx265",
			"libvpx",
		},
		Formats: []string{
			"mp4",
			"webm",
			"gif",
		},
	}

	// Test encoder capability
	if !info.HasEncoder("libx264") {
		t.Error("Should have libx264 encoder")
	}

	if info.HasEncoder("nonexistent") {
		t.Error("Should not have nonexistent encoder")
	}

	// Test format capability
	if !info.HasFormat("mp4") {
		t.Error("Should have mp4 format")
	}

	if info.HasFormat("nonexistent") {
		t.Error("Should not have nonexistent format")
	}
}

func TestFFmpegInfoCanRecord(t *testing.T) {
	// FFmpeg with full capabilities
	fullInfo := &FFmpegInfo{
		Path:    "/usr/bin/ffmpeg",
		Version: FFmpegVersion{Major: 6, Minor: 1, Patch: 0},
		Encoders: []string{
			"libx264",
			"rawvideo",
		},
		Formats: []string{
			"mp4",
			"rawvideo",
		},
	}

	if !fullInfo.CanRecord() {
		t.Error("FFmpeg with full capabilities should be able to record")
	}

	// FFmpeg without rawvideo input
	limitedInfo := &FFmpegInfo{
		Path:     "/usr/bin/ffmpeg",
		Version:  FFmpegVersion{Major: 6, Minor: 1, Patch: 0},
		Encoders: []string{},
		Formats:  []string{"mp4"},
	}

	// Should still work as long as basic FFmpeg is available
	t.Logf("Limited FFmpeg CanRecord: %v", limitedInfo.CanRecord())
}

// =============================================================================
// Platform-Specific Tests
// =============================================================================

func TestFFmpegPathsByPlatform(t *testing.T) {
	paths := GetDefaultFFmpegPaths()

	if len(paths) == 0 {
		t.Error("Expected at least one default path")
	}

	// Platform-specific checks
	switch runtime.GOOS {
	case "windows":
		// Should include common Windows paths
		found := false
		for _, p := range paths {
			if strings.Contains(p, "ffmpeg.exe") {
				found = true
				break
			}
		}
		if !found {
			t.Log("No Windows-specific ffmpeg.exe path found")
		}
	case "darwin":
		// Should include Homebrew paths
		found := false
		for _, p := range paths {
			if strings.Contains(p, "/opt/homebrew") || strings.Contains(p, "/usr/local") {
				found = true
				break
			}
		}
		if !found {
			t.Log("No macOS Homebrew paths found")
		}
	case "linux":
		// Should include standard Linux paths
		found := false
		for _, p := range paths {
			if strings.Contains(p, "/usr/bin") {
				found = true
				break
			}
		}
		if !found {
			t.Log("No Linux standard paths found")
		}
	}
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkDetectFFmpeg(b *testing.B) {
	if !IsFFmpegAvailable() {
		b.Skip("FFmpeg not installed")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DetectFFmpeg()
	}
}

func BenchmarkCommandBuilder(b *testing.B) {
	settings := Settings{
		Quality: QualityMedium,
		FPS:     20,
		Format:  FormatMP4,
		CRF:     23,
		Preset:  PresetUltrafast,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewFFmpegCommandBuilder().
			ForRecording(1920, 1080, settings).
			Output("test.mp4").
			Build()
	}
}
