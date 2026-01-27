package recording

import (
	"errors"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// DetectFFmpeg detects FFmpeg installation and returns information about it.
func DetectFFmpeg() (*FFmpegInfo, error) {
	path, err := DetectFFmpegPath()
	if err != nil {
		return nil, err
	}

	// Get version
	cmd := exec.Command(path, "-version")
	output, err := cmd.Output()
	if err != nil {
		return nil, errors.New("ffmpeg found but could not get version")
	}

	version, err := ParseFFmpegVersion(string(output))
	if err != nil {
		// Still return info even if version parsing fails
		return &FFmpegInfo{
			Path:    path,
			Version: FFmpegVersion{},
		}, nil
	}

	info := &FFmpegInfo{
		Path:    path,
		Version: version,
	}

	// Get encoders
	info.Encoders = GetSupportedEncoders()

	// Get formats
	info.Formats = GetSupportedFormats()

	return info, nil
}

// DetectFFmpegPath finds the FFmpeg binary.
func DetectFFmpegPath() (string, error) {
	// First try PATH
	path, err := exec.LookPath("ffmpeg")
	if err == nil {
		return path, nil
	}

	// Try default paths
	for _, p := range GetDefaultFFmpegPaths() {
		if _, err := exec.LookPath(p); err == nil {
			return p, nil
		}
		// Try running it directly
		cmd := exec.Command(p, "-version")
		if err := cmd.Run(); err == nil {
			return p, nil
		}
	}

	return "", errors.New("ffmpeg not found or not installed")
}

// IsFFmpegAvailable returns true if FFmpeg is installed and accessible.
func IsFFmpegAvailable() bool {
	_, err := DetectFFmpeg()
	return err == nil
}

// ParseFFmpegVersion parses FFmpeg version output.
func ParseFFmpegVersion(output string) (FFmpegVersion, error) {
	if output == "" {
		return FFmpegVersion{}, errors.New("empty version output")
	}

	// Match patterns like "ffmpeg version 6.1.1" or "ffmpeg version n5.1.2-3-g..."
	// Also handle "4.4.2-0ubuntu..." formats
	re := regexp.MustCompile(`ffmpeg version [n]?(\d+)\.(\d+)(?:\.(\d+))?`)
	matches := re.FindStringSubmatch(output)

	if len(matches) < 3 {
		return FFmpegVersion{}, errors.New("could not parse version")
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch := 0
	if len(matches) > 3 && matches[3] != "" {
		patch, _ = strconv.Atoi(matches[3])
	}

	return FFmpegVersion{
		Major: major,
		Minor: minor,
		Patch: patch,
	}, nil
}

// CheckCodecSupport checks if a specific codec is supported.
func CheckCodecSupport(codec string) bool {
	path, err := DetectFFmpegPath()
	if err != nil {
		return false
	}

	cmd := exec.Command(path, "-encoders")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.Contains(string(output), codec)
}

// GetSupportedEncoders returns a list of supported video encoders.
func GetSupportedEncoders() []string {
	path, err := DetectFFmpegPath()
	if err != nil {
		return nil
	}

	cmd := exec.Command(path, "-encoders")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var encoders []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		// Video encoders start with "V" in the capability flags
		if strings.HasPrefix(strings.TrimSpace(line), "V") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				encoders = append(encoders, fields[1])
			}
		}
	}

	return encoders
}

// GetSupportedFormats returns a list of supported output formats.
func GetSupportedFormats() []string {
	path, err := DetectFFmpegPath()
	if err != nil {
		return nil
	}

	cmd := exec.Command(path, "-formats")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var formats []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		// Muxing formats have "E" (encode) capability
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "E") {
			fields := strings.Fields(trimmed)
			if len(fields) >= 2 {
				formats = append(formats, fields[1])
			}
		}
	}

	return formats
}

// GetDefaultFFmpegPaths returns platform-specific default FFmpeg paths.
func GetDefaultFFmpegPaths() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{
			"ffmpeg.exe",
			"C:\\ffmpeg\\bin\\ffmpeg.exe",
			"C:\\Program Files\\ffmpeg\\bin\\ffmpeg.exe",
			"C:\\Program Files (x86)\\ffmpeg\\bin\\ffmpeg.exe",
		}
	case "darwin":
		return []string{
			"/opt/homebrew/bin/ffmpeg",
			"/usr/local/bin/ffmpeg",
			"/usr/bin/ffmpeg",
		}
	default: // Linux and others
		return []string{
			"/usr/bin/ffmpeg",
			"/usr/local/bin/ffmpeg",
			"/snap/bin/ffmpeg",
		}
	}
}

// =============================================================================
// FFmpeg Command Builder
// =============================================================================

// FFmpegCommandBuilder builds FFmpeg command arguments.
type FFmpegCommandBuilder struct {
	args     []string
	settings *Settings
}

// NewFFmpegCommandBuilder creates a new command builder.
func NewFFmpegCommandBuilder() *FFmpegCommandBuilder {
	return &FFmpegCommandBuilder{
		args: make([]string, 0),
	}
}

// WithSettings sets the recording settings and applies them to the command.
func (b *FFmpegCommandBuilder) WithSettings(settings Settings) *FFmpegCommandBuilder {
	b.settings = &settings
	// Apply settings to command args
	b.Framerate(settings.FPS)
	b.CRF(settings.CRF)
	b.Preset(settings.Preset)
	return b
}

// Overwrite adds the -y flag to overwrite output files.
func (b *FFmpegCommandBuilder) Overwrite(yes bool) *FFmpegCommandBuilder {
	if yes {
		b.args = append(b.args, "-y")
	}
	return b
}

// InputFormat sets the input format.
func (b *FFmpegCommandBuilder) InputFormat(format string) *FFmpegCommandBuilder {
	b.args = append(b.args, "-f", format)
	return b
}

// InputPixelFormat sets the input pixel format.
func (b *FFmpegCommandBuilder) InputPixelFormat(format string) *FFmpegCommandBuilder {
	b.args = append(b.args, "-pixel_format", format)
	return b
}

// VideoSize sets the video dimensions.
func (b *FFmpegCommandBuilder) VideoSize(width, height int) *FFmpegCommandBuilder {
	b.args = append(b.args, "-video_size", strconv.Itoa(width)+"x"+strconv.Itoa(height))
	return b
}

// Framerate sets the input framerate.
func (b *FFmpegCommandBuilder) Framerate(fps int) *FFmpegCommandBuilder {
	b.args = append(b.args, "-framerate", strconv.Itoa(fps))
	return b
}

// Input sets the input source.
func (b *FFmpegCommandBuilder) Input(source string) *FFmpegCommandBuilder {
	b.args = append(b.args, "-i", source)
	return b
}

// VideoCodec sets the video codec.
func (b *FFmpegCommandBuilder) VideoCodec(codec VideoCodec) *FFmpegCommandBuilder {
	b.args = append(b.args, "-c:v", string(codec))
	return b
}

// Preset sets the encoding preset.
func (b *FFmpegCommandBuilder) Preset(preset EncodingPreset) *FFmpegCommandBuilder {
	b.args = append(b.args, "-preset", string(preset))
	return b
}

// CRF sets the constant rate factor.
func (b *FFmpegCommandBuilder) CRF(crf int) *FFmpegCommandBuilder {
	b.args = append(b.args, "-crf", strconv.Itoa(crf))
	return b
}

// OutputPixelFormat sets the output pixel format.
func (b *FFmpegCommandBuilder) OutputPixelFormat(format string) *FFmpegCommandBuilder {
	b.args = append(b.args, "-pix_fmt", format)
	return b
}

// Output sets the output file.
func (b *FFmpegCommandBuilder) Output(path string) *FFmpegCommandBuilder {
	b.args = append(b.args, path)
	return b
}

// AudioInput adds an audio input source (PulseAudio).
func (b *FFmpegCommandBuilder) AudioInput(deviceName string) *FFmpegCommandBuilder {
	b.args = append(b.args, "-f", "pulse", "-i", deviceName)
	return b
}

// AudioCodec sets the audio codec.
func (b *FFmpegCommandBuilder) AudioCodec(codec string) *FFmpegCommandBuilder {
	b.args = append(b.args, "-c:a", codec)
	return b
}

// AudioBitrate sets the audio bitrate in kbps.
func (b *FFmpegCommandBuilder) AudioBitrate(kbps int) *FFmpegCommandBuilder {
	b.args = append(b.args, "-b:a", strconv.Itoa(kbps)+"k")
	return b
}

// AudioSampleRate sets the audio sample rate.
func (b *FFmpegCommandBuilder) AudioSampleRate(rate int) *FFmpegCommandBuilder {
	b.args = append(b.args, "-ar", strconv.Itoa(rate))
	return b
}

// AudioChannels sets the number of audio channels.
func (b *FFmpegCommandBuilder) AudioChannels(channels int) *FFmpegCommandBuilder {
	b.args = append(b.args, "-ac", strconv.Itoa(channels))
	return b
}

// NoAudio disables audio.
func (b *FFmpegCommandBuilder) NoAudio() *FFmpegCommandBuilder {
	b.args = append(b.args, "-an")
	return b
}

// ForRecording configures the builder for standard recording.
func (b *FFmpegCommandBuilder) ForRecording(width, height int, settings Settings) *FFmpegCommandBuilder {
	b.Overwrite(true)
	b.InputFormat("rawvideo")
	b.InputPixelFormat("rgb24")
	b.VideoSize(width, height)
	b.Framerate(settings.FPS)
	b.Input("-")

	// Set codec based on format
	switch settings.Format {
	case FormatWebM:
		b.VideoCodec(CodecVP8)
	case FormatGIF:
		// GIF has different handling
		b.args = append(b.args, "-f", "gif")
	default:
		b.VideoCodec(CodecH264)
		b.Preset(settings.Preset)
		b.CRF(settings.CRF)
		b.OutputPixelFormat("yuv420p")
	}

	return b
}

// ForRecordingWithAudio configures the builder for recording with audio.
func (b *FFmpegCommandBuilder) ForRecordingWithAudio(width, height int, settings Settings) *FFmpegCommandBuilder {
	b.Overwrite(true)

	// Video input (from pipe)
	b.InputFormat("rawvideo")
	b.InputPixelFormat("rgb24")
	b.VideoSize(width, height)
	b.Framerate(settings.FPS)
	b.Input("-")

	// Audio input (from PulseAudio)
	if settings.Audio.Enabled && settings.Format.SupportsAudio() {
		deviceName := settings.Audio.DeviceName
		if deviceName == "" {
			deviceName = "default"
		}
		b.AudioInput(deviceName)
	}

	// Video codec settings
	switch settings.Format {
	case FormatWebM:
		b.VideoCodec(CodecVP8)
		if settings.Audio.Enabled {
			b.AudioCodec("libopus")
		}
	case FormatGIF:
		// GIF doesn't support audio
		b.args = append(b.args, "-f", "gif")
	default:
		b.VideoCodec(CodecH264)
		b.Preset(settings.Preset)
		b.CRF(settings.CRF)
		b.OutputPixelFormat("yuv420p")

		// Audio codec for MP4
		if settings.Audio.Enabled {
			b.AudioCodec(settings.Audio.Codec)
			b.AudioBitrate(settings.Audio.Bitrate)
			b.AudioSampleRate(settings.Audio.SampleRate)
			b.AudioChannels(settings.Audio.Channels)
		}
	}

	// Sync audio/video
	if settings.Audio.Enabled && settings.Format.SupportsAudio() {
		// Use shortest to stop when video ends
		b.args = append(b.args, "-shortest")
	}

	return b
}

// Build returns the command arguments.
func (b *FFmpegCommandBuilder) Build() []string {
	return b.args
}
