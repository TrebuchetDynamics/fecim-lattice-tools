package recording

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"fecim-lattice-tools/shared/logging"
)

// Package-level logger for recording
var log = logging.NewLogger("recording")

// AudioDevice represents an available audio input device.
type AudioDevice struct {
	ID          string
	Name        string
	Description string
	IsDefault   bool
}

// AudioMonitor provides real-time audio level monitoring.
type AudioMonitor struct {
	mu sync.RWMutex

	// State
	isRunning   atomic.Bool
	stopChan    chan struct{}
	currentLevel atomic.Int32 // 0-100 percentage
	peakLevel    atomic.Int32 // Peak level for visualization

	// Device
	device     AudioDevice
	sampleRate int

	// FFmpeg process for monitoring
	cmd    *exec.Cmd
	stderr io.ReadCloser

	// Callbacks
	onLevelChange func(level, peak int)
	onError       func(error)
}

// NewAudioMonitor creates a new audio level monitor.
func NewAudioMonitor() *AudioMonitor {
	return &AudioMonitor{
		sampleRate: 44100,
	}
}

// DetectAudioDevices returns a list of available audio input devices.
func DetectAudioDevices() ([]AudioDevice, error) {
	log.Debug("Detecting audio devices...")
	devices := []AudioDevice{}

	// Try pactl for PulseAudio/PipeWire (most common on modern Linux)
	cmd := exec.Command("pactl", "list", "sources", "short")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				name := fields[1]
				// Filter out monitor sources (these are output monitors, not inputs)
				// Real microphones have "input" in the name or don't have "monitor"
				isMonitor := strings.Contains(strings.ToLower(name), "monitor")
				isInput := strings.Contains(strings.ToLower(name), "input")

				// Include if it's an input OR if it doesn't have "monitor"
				if isInput || !isMonitor {
					// Build description from remaining fields if available
					description := name
					if len(fields) >= 3 {
						description = strings.Join(fields[2:], " ")
					}

					device := AudioDevice{
						ID:          fields[0],
						Name:        name,
						Description: description,
						IsDefault:   isInput, // Prefer input devices as default
					}
					devices = append(devices, device)
				}
			}
		}
	}

	// Try FFmpeg's pulse source detection as alternative
	if len(devices) == 0 {
		cmd = exec.Command("ffmpeg", "-sources", "pulse")
		output, err = cmd.CombinedOutput()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				// Look for lines starting with "  " (device entries) or "*" (default)
				if strings.HasPrefix(line, "  ") || strings.HasPrefix(line, "*") {
					// Format: "  device_name [description] (state)" or "* device_name..."
					isDefault := strings.HasPrefix(line, "*")
					line = strings.TrimPrefix(line, "*")
					line = strings.TrimPrefix(line, " ")
					line = strings.TrimSpace(line)

					// Skip monitor devices
					if strings.Contains(strings.ToLower(line), "monitor") {
						continue
					}

					// Extract device name (first word before [)
					parts := strings.SplitN(line, " [", 2)
					if len(parts) >= 1 {
						deviceName := strings.TrimSpace(parts[0])
						description := deviceName
						if len(parts) >= 2 {
							description = strings.TrimSuffix(parts[1], "]")
							// Remove trailing (state) if present
							if idx := strings.LastIndex(description, "] ("); idx > 0 {
								description = description[:idx]
							}
						}

						if deviceName != "" && strings.Contains(strings.ToLower(deviceName), "input") {
							devices = append(devices, AudioDevice{
								ID:          deviceName,
								Name:        deviceName,
								Description: description,
								IsDefault:   isDefault,
							})
						}
					}
				}
			}
		}
	}

	// Try ALSA as fallback
	if len(devices) == 0 {
		cmd = exec.Command("arecord", "-l")
		output, err = cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "card ") {
					// Parse: "card 0: Device [Device Name], device 0: ..."
					parts := strings.SplitN(line, ":", 2)
					if len(parts) >= 2 {
						cardNum := strings.TrimPrefix(parts[0], "card ")
						cardNum = strings.TrimSpace(cardNum)
						device := AudioDevice{
							ID:          "hw:" + cardNum,
							Name:        "hw:" + cardNum,
							Description: strings.TrimSpace(parts[1]),
							IsDefault:   cardNum == "0" || cardNum == "1",
						}
						devices = append(devices, device)
					}
				}
			}
		}
	}

	// Always add "default" as a fallback option
	hasDefault := false
	for _, d := range devices {
		if d.Name == "default" {
			hasDefault = true
			break
		}
	}
	if !hasDefault && len(devices) > 0 {
		// Add default as first option
		devices = append([]AudioDevice{{
			ID:          "default",
			Name:        "default",
			Description: "System default audio input",
			IsDefault:   true,
		}}, devices...)
	}

	if len(devices) == 0 {
		log.Debug("No audio input devices found")
		return nil, errors.New("no audio input devices found")
	}

	log.Debug("Found %d audio device(s)", len(devices))
	for i, d := range devices {
		log.Debug("  Device %d: ID=%s Name=%s Default=%v", i+1, d.ID, d.Name, d.IsDefault)
	}

	return devices, nil
}

// IsAudioAvailable checks if audio recording is available.
func IsAudioAvailable() bool {
	devices, err := DetectAudioDevices()
	return err == nil && len(devices) > 0
}

// GetDefaultAudioDevice returns the default audio input device.
// Prefers actual input devices over the "default" placeholder.
func GetDefaultAudioDevice() (AudioDevice, error) {
	devices, err := DetectAudioDevices()
	if err != nil {
		return AudioDevice{}, err
	}

	// First, look for an actual input device (has "input" in name, not "default" placeholder)
	for _, d := range devices {
		if strings.Contains(strings.ToLower(d.Name), "input") && d.Name != "default" {
			log.Debug("GetDefaultAudioDevice: using actual input device: %s", d.Name)
			return d, nil
		}
	}

	// Second, look for any non-default device
	for _, d := range devices {
		if d.Name != "default" && d.ID != "default" {
			log.Debug("GetDefaultAudioDevice: using non-default device: %s", d.Name)
			return d, nil
		}
	}

	// Fall back to default if it's the only option
	for _, d := range devices {
		if d.IsDefault {
			log.Debug("GetDefaultAudioDevice: falling back to default device: %s", d.Name)
			return d, nil
		}
	}

	// Return first device if no default marked
	if len(devices) > 0 {
		log.Debug("GetDefaultAudioDevice: using first available device: %s", devices[0].Name)
		return devices[0], nil
	}

	return AudioDevice{}, errors.New("no audio device found")
}

// SetDevice sets the audio device to monitor.
func (am *AudioMonitor) SetDevice(device AudioDevice) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.device = device
}

// Start begins monitoring audio levels.
func (am *AudioMonitor) Start() error {
	if am.isRunning.Load() {
		log.Debug("AudioMonitor.Start: already running")
		return errors.New("already monitoring")
	}

	am.mu.Lock()
	device := am.device
	am.mu.Unlock()

	if device.ID == "" || device.Name == "" {
		log.Debug("AudioMonitor.Start: no device configured, getting default")
		// Try to get default device
		defaultDevice, err := GetDefaultAudioDevice()
		if err != nil {
			log.Debug("AudioMonitor.Start: failed to get default device: %v", err)
			return fmt.Errorf("no audio device configured and no default found: %w", err)
		}
		am.mu.Lock()
		am.device = defaultDevice
		device = defaultDevice
		am.mu.Unlock()
	}

	log.Info("AudioMonitor.Start: using device '%s' (ID=%s)", device.Name, device.ID)

	// Use FFmpeg to capture audio and output volume levels to stderr
	// The ebur128 filter outputs loudness levels to stderr
	am.cmd = exec.Command("ffmpeg",
		"-f", "pulse",
		"-i", device.Name,
		"-af", "ebur128=peak=true",
		"-f", "null",
		"-",
	)

	log.Debug("AudioMonitor.Start: FFmpeg command: ffmpeg -f pulse -i %s -af ebur128=peak=true -f null -", device.Name)

	var err error
	am.stderr, err = am.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	if err := am.cmd.Start(); err != nil {
		log.Debug("AudioMonitor.Start: failed to start FFmpeg: %v", err)
		return fmt.Errorf("failed to start audio monitor: %w", err)
	}

	am.mu.Lock()
	am.stopChan = make(chan struct{})
	am.mu.Unlock()

	am.isRunning.Store(true)
	log.Info("AudioMonitor.Start: FFmpeg started, beginning level parsing")

	// Start level parsing goroutine
	go am.parseAudioLevels()

	return nil
}

// Stop stops monitoring audio levels.
func (am *AudioMonitor) Stop() {
	if !am.isRunning.Swap(false) {
		return // Already stopped or was never running
	}

	am.mu.Lock()
	stopChan := am.stopChan
	am.stopChan = nil
	cmd := am.cmd
	am.cmd = nil
	am.mu.Unlock()

	if stopChan != nil {
		close(stopChan)
	}

	if cmd != nil && cmd.Process != nil {
		cmd.Process.Kill()
		cmd.Wait()
	}

	am.currentLevel.Store(0)
	am.peakLevel.Store(0)
}

// parseAudioLevels reads FFmpeg output and extracts audio levels.
func (am *AudioMonitor) parseAudioLevels() {
	defer am.Stop()

	// Get stopChan under lock
	am.mu.RLock()
	stopChan := am.stopChan
	stderr := am.stderr
	am.mu.RUnlock()

	if stopChan == nil || stderr == nil {
		return
	}

	scanner := bufio.NewScanner(stderr)

	// Start peak decay goroutine
	go func() {
		peakDecay := time.NewTicker(100 * time.Millisecond)
		defer peakDecay.Stop()
		for {
			select {
			case <-stopChan:
				return
			case <-peakDecay.C:
				// Slowly decay peak level
				current := am.peakLevel.Load()
				if current > 0 {
					newPeak := current - 3
					if newPeak < 0 {
						newPeak = 0
					}
					am.peakLevel.Store(newPeak)
				}
			}
		}
	}()

	for {
		select {
		case <-stopChan:
			return
		default:
			if !scanner.Scan() {
				// Scanner error or EOF
				if err := scanner.Err(); err != nil {
					am.mu.RLock()
					errCallback := am.onError
					am.mu.RUnlock()
					if errCallback != nil {
						errCallback(err)
					}
				}
				return
			}

			line := scanner.Text()

			// Parse ebur128 output which looks like:
			// [Parsed_ebur128_0 @ ...] M: -20.5 S: -21.3 I: -23.0 LUFS LRA: 5.2 LU FTPK: -10.2 -9.8 dBFS TPK: -8.3 -7.9 dBFS
			// We want the FTPK (true peak) or M (momentary loudness)
			var db float64
			found := false

			// Try to parse momentary loudness (M:)
			if idx := strings.Index(line, "M:"); idx >= 0 {
				rest := line[idx+2:]
				rest = strings.TrimSpace(rest)
				fields := strings.Fields(rest)
				if len(fields) >= 1 {
					val := parseDbValue(fields[0])
					if val > -200 { // Valid value
						db = val
						found = true
					}
				}
			}

			// Also try parsing FTPK (faster true peak) for more responsive display
			if idx := strings.Index(line, "FTPK:"); idx >= 0 {
				rest := line[idx+5:]
				rest = strings.TrimSpace(rest)
				fields := strings.Fields(rest)
				if len(fields) >= 1 {
					val := parseDbValue(fields[0])
					if val > -200 { // Valid value
						// Use peak if it's higher
						if val > db {
							db = val
						}
						found = true
					}
				}
			}

			if found {
				// Clamp very low values (silence) to -70
				if db < -70 {
					db = -70
				}
				// Convert dB/LUFS to percentage (0-100)
				// LUFS range is typically -70 to 0, but we use a more practical range
				level := am.dbToPercent(db)

				// Log significant level changes (only when level > 5 to avoid spam)
				oldLevel := int(am.currentLevel.Load())
				if level > 5 && (level-oldLevel > 10 || oldLevel-level > 10) {
					log.Debug("Audio level: %.1f dB -> %d%% (was %d%%)", db, level, oldLevel)
				}

				am.currentLevel.Store(int32(level))

				// Update peak if higher
				if int32(level) > am.peakLevel.Load() {
					am.peakLevel.Store(int32(level))
				}

				// Notify callback
				am.mu.RLock()
				callback := am.onLevelChange
				am.mu.RUnlock()

				if callback != nil {
					callback(level, int(am.peakLevel.Load()))
				}
			}
		}
	}
}

// parseDbValue parses a dB value string, handling "-inf" and other special cases.
func parseDbValue(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "-inf" || s == "inf" || s == "-infinity" {
		return -200 // Treat as very low value (silence)
	}
	if val, err := strconv.ParseFloat(s, 64); err == nil {
		return val
	}
	return -200 // Return very low value on parse error
}

// dbToPercent converts decibels to a 0-100 percentage.
// Uses a wider range (-80 to 0 dB) to capture quiet sounds from typical microphones.
func (am *AudioMonitor) dbToPercent(db float64) int {
	// Use a wider range for microphone input
	// Typical microphone levels:
	//   -80 dB = very quiet / noise floor
	//   -60 dB = quiet speech
	//   -40 dB = normal speech
	//   -20 dB = loud speech
	//   -10 dB = very loud / clipping danger
	//     0 dB = maximum / clipping

	minDb := -80.0 // Noise floor
	maxDb := -10.0 // Loud (leave headroom before 0)

	if db < minDb {
		return 0
	}
	if db > maxDb {
		return 100
	}

	// Linear mapping from minDb..maxDb to 0..100
	percent := ((db - minDb) / (maxDb - minDb)) * 100
	if percent < 0 {
		return 0
	}
	if percent > 100 {
		return 100
	}
	return int(percent)
}

// Level returns the current audio level (0-100).
func (am *AudioMonitor) Level() int {
	return int(am.currentLevel.Load())
}

// PeakLevel returns the peak audio level (0-100).
func (am *AudioMonitor) PeakLevel() int {
	return int(am.peakLevel.Load())
}

// IsRunning returns true if the monitor is active.
func (am *AudioMonitor) IsRunning() bool {
	return am.isRunning.Load()
}

// OnLevelChange sets a callback for level changes.
func (am *AudioMonitor) OnLevelChange(callback func(level, peak int)) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.onLevelChange = callback
}

// OnError sets a callback for errors.
func (am *AudioMonitor) OnError(callback func(error)) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.onError = callback
}

// =============================================================================
// Audio Settings for Recording
// =============================================================================

// AudioSettings contains audio recording configuration.
type AudioSettings struct {
	Enabled    bool   `json:"enabled"`
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
	SampleRate int    `json:"sample_rate"` // 44100, 48000
	Channels   int    `json:"channels"`    // 1 (mono), 2 (stereo)
	Codec      string `json:"codec"`       // aac, mp3, opus
	Bitrate    int    `json:"bitrate"`     // kbps
}

// DefaultAudioSettings returns default audio recording settings.
func DefaultAudioSettings() AudioSettings {
	return AudioSettings{
		Enabled:    false,
		SampleRate: 44100,
		Channels:   2,
		Codec:      "aac",
		Bitrate:    128,
	}
}

// Validate checks if the audio settings are valid.
func (as AudioSettings) Validate() error {
	if !as.Enabled {
		return nil // Disabled settings don't need validation
	}

	if as.SampleRate != 44100 && as.SampleRate != 48000 && as.SampleRate != 96000 {
		return &ValidationError{
			Field:   "SampleRate",
			Message: "SampleRate must be 44100, 48000, or 96000",
		}
	}

	if as.Channels < 1 || as.Channels > 2 {
		return &ValidationError{
			Field:   "Channels",
			Message: "Channels must be 1 (mono) or 2 (stereo)",
		}
	}

	validCodecs := map[string]bool{"aac": true, "mp3": true, "opus": true, "flac": true}
	if !validCodecs[as.Codec] {
		return &ValidationError{
			Field:   "Codec",
			Message: "Codec must be one of: aac, mp3, opus, flac",
		}
	}

	if as.Bitrate < 64 || as.Bitrate > 320 {
		return &ValidationError{
			Field:   "Bitrate",
			Message: "Bitrate must be between 64 and 320 kbps",
		}
	}

	return nil
}
