package recording

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// RecordingManager implements the Manager interface for screen recording.
type RecordingManager struct {
	mu sync.RWMutex

	// State
	state      State
	settings   Settings
	outputFile string

	// FFmpeg process
	cmd   *exec.Cmd
	stdin io.WriteCloser

	// Capture
	source     CaptureSource
	bufferPool *BufferPool
	stopChan   chan struct{}

	// Metrics
	startTime       time.Time
	pauseTime       time.Time
	totalPausedTime time.Duration
	framesCaptured  int
	framesDropped   int

	// Callbacks
	onStateChange   func(State)
	onFrameCaptured func(int)
	onError         func(error)
}

// NewManager creates a new recording manager with default settings.
func NewManager() *RecordingManager {
	return &RecordingManager{
		state:    StateIdle,
		settings: DefaultSettings(),
	}
}

// NewManagerWithSettings creates a new recording manager with custom settings.
func NewManagerWithSettings(settings Settings) *RecordingManager {
	return &RecordingManager{
		state:    StateIdle,
		settings: settings,
	}
}

// Start begins recording from the given capture source.
func (m *RecordingManager) Start(source CaptureSource) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.state == StateRecording || m.state == StatePaused {
		return errors.New("already recording")
	}

	if source == nil {
		return errors.New("capture source is nil")
	}

	width := source.Width()
	height := source.Height()

	if width <= 0 || height <= 0 {
		return fmt.Errorf("invalid dimensions: %dx%d", width, height)
	}

	// Validate settings before use (CRITICAL: prevents command injection)
	if err := m.settings.Validate(); err != nil {
		return fmt.Errorf("invalid settings: %w", err)
	}

	// Check FFmpeg availability
	ffmpegPath, err := DetectFFmpegPath()
	if err != nil {
		return fmt.Errorf("ffmpeg not found or not installed: %w", err)
	}

	// Ensure dimensions are even (required by H.264)
	if width%2 != 0 {
		width--
	}
	if height%2 != 0 {
		height--
	}

	// Create output directory
	recordingDir := "recordings"
	if err := os.MkdirAll(recordingDir, 0755); err != nil {
		return fmt.Errorf("failed to create recordings directory: %w", err)
	}

	// Generate output filename
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	m.outputFile = filepath.Join(recordingDir, fmt.Sprintf("fecim_recording_%s%s", timestamp, m.settings.Format.Extension()))

	// Build FFmpeg command (with or without audio)
	var builder *FFmpegCommandBuilder
	if m.settings.Audio.Enabled && m.settings.Format.SupportsAudio() {
		log.Info("Recording with audio enabled: device=%s codec=%s bitrate=%d",
			m.settings.Audio.DeviceName, m.settings.Audio.Codec, m.settings.Audio.Bitrate)
		builder = NewFFmpegCommandBuilder().
			ForRecordingWithAudio(width, height, m.settings).
			Output(m.outputFile)
	} else {
		log.Info("Recording without audio (Audio.Enabled=%v, Format.SupportsAudio=%v)",
			m.settings.Audio.Enabled, m.settings.Format.SupportsAudio())
		builder = NewFFmpegCommandBuilder().
			ForRecording(width, height, m.settings).
			Output(m.outputFile)
	}

	args := builder.Build()
	log.Debug("FFmpeg command: %s %v", ffmpegPath, args)

	// Create FFmpeg process
	m.cmd = exec.Command(ffmpegPath, args...)

	// Get stdin pipe
	m.stdin, err = m.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	// Start FFmpeg
	if err := m.cmd.Start(); err != nil {
		// Clean up stdin pipe on start failure to avoid resource leak
		m.stdin.Close()
		m.stdin = nil
		return fmt.Errorf("failed to start FFmpeg: %w", err)
	}

	// Initialize state
	m.source = source
	m.bufferPool = NewBufferPool(width, height)
	m.stopChan = make(chan struct{})
	m.startTime = time.Now()
	m.totalPausedTime = 0
	m.framesCaptured = 0
	m.framesDropped = 0
	m.state = StateRecording

	// Start capture goroutine
	go m.captureLoop(width, height)

	// Notify state change
	if m.onStateChange != nil {
		go m.onStateChange(StateRecording)
	}

	return nil
}

// Stop stops recording and returns the output file path.
func (m *RecordingManager) Stop() (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.state != StateRecording && m.state != StatePaused {
		return "", errors.New("not recording")
	}

	// Signal capture loop to stop
	if m.stopChan != nil {
		close(m.stopChan)
		m.stopChan = nil
	}

	// Close stdin to signal EOF to FFmpeg
	if m.stdin != nil {
		m.stdin.Close()
		m.stdin = nil
	}

	// Wait for FFmpeg to finish
	if m.cmd != nil {
		m.cmd.Wait()
		m.cmd = nil
	}

	m.state = StateStopped
	outputFile := m.outputFile

	// Notify state change
	if m.onStateChange != nil {
		go m.onStateChange(StateStopped)
	}

	return outputFile, nil
}

// Pause pauses the recording.
func (m *RecordingManager) Pause() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.state != StateRecording {
		return errors.New("not recording")
	}

	m.pauseTime = time.Now()
	m.state = StatePaused

	if m.onStateChange != nil {
		go m.onStateChange(StatePaused)
	}

	return nil
}

// Resume resumes a paused recording.
func (m *RecordingManager) Resume() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.state != StatePaused {
		return errors.New("not paused")
	}

	m.totalPausedTime += time.Since(m.pauseTime)
	m.state = StateRecording

	if m.onStateChange != nil {
		go m.onStateChange(StateRecording)
	}

	return nil
}

// State returns the current recording state.
func (m *RecordingManager) State() State {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state
}

// IsRecording returns true if actively recording (not paused).
func (m *RecordingManager) IsRecording() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state == StateRecording
}

// IsPaused returns true if recording is paused.
func (m *RecordingManager) IsPaused() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state == StatePaused
}

// ElapsedTime returns the total recording time (excluding pauses).
func (m *RecordingManager) ElapsedTime() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.state == StateIdle || m.state == StateStopped {
		return 0
	}

	if m.state == StatePaused {
		return m.pauseTime.Sub(m.startTime) - m.totalPausedTime
	}

	return time.Since(m.startTime) - m.totalPausedTime
}

// EstimatedFileSize returns the estimated current file size in bytes.
func (m *RecordingManager) EstimatedFileSize() int64 {
	// Extract all needed values under a single lock to avoid nested RLock
	m.mu.RLock()
	source := m.source
	settings := m.settings
	state := m.state
	startTime := m.startTime
	pauseTime := m.pauseTime
	totalPausedTime := m.totalPausedTime
	m.mu.RUnlock()

	if source == nil {
		return 0
	}

	// Calculate elapsed time without holding lock (avoids nested RLock)
	var elapsed time.Duration
	switch state {
	case StateIdle, StateStopped:
		elapsed = 0
	case StatePaused:
		elapsed = pauseTime.Sub(startTime) - totalPausedTime
	default:
		elapsed = time.Since(startTime) - totalPausedTime
	}

	return settings.EstimatedFileSize(source.Width(), source.Height(), int(elapsed.Seconds()))
}

// FramesCaptured returns the number of frames captured.
func (m *RecordingManager) FramesCaptured() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.framesCaptured
}

// FramesDropped returns the number of frames dropped.
func (m *RecordingManager) FramesDropped() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.framesDropped
}

// SetSettings sets the recording settings (only when not recording).
func (m *RecordingManager) SetSettings(settings Settings) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.state == StateRecording || m.state == StatePaused {
		return // Cannot change settings while recording
	}

	m.settings = settings
}

// GetSettings returns the current settings.
func (m *RecordingManager) GetSettings() Settings {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settings.Copy()
}

// OnStateChange sets a callback for state changes.
func (m *RecordingManager) OnStateChange(callback func(State)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onStateChange = callback
}

// OnFrameCaptured sets a callback for each captured frame.
func (m *RecordingManager) OnFrameCaptured(callback func(int)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onFrameCaptured = callback
}

// OnError sets a callback for errors.
func (m *RecordingManager) OnError(callback func(error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onError = callback
}

// captureLoop continuously captures frames and sends them to FFmpeg.
func (m *RecordingManager) captureLoop(width, height int) {
	interval := time.Second / time.Duration(m.settings.FPS)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Store stopChan locally to avoid race with Stop()
	m.mu.RLock()
	stopChan := m.stopChan
	m.mu.RUnlock()

	for {
		select {
		case <-stopChan:
			return
		case <-ticker.C:
			m.mu.RLock()
			state := m.state
			stdin := m.stdin
			source := m.source
			m.mu.RUnlock()

			if state == StatePaused {
				continue
			}

			if state != StateRecording || stdin == nil || source == nil {
				return
			}

			// Capture frame
			frameData, err := source.Capture()
			if err != nil {
				m.mu.Lock()
				m.framesDropped++
				m.mu.Unlock()
				continue
			}

			if frameData == nil || len(frameData) == 0 {
				m.mu.Lock()
				m.framesDropped++
				m.mu.Unlock()
				continue
			}

			// Write frame to FFmpeg
			_, err = stdin.Write(frameData)
			if err != nil {
				m.mu.Lock()
				m.framesDropped++
				if m.onError != nil {
					go m.onError(err)
				}
				m.mu.Unlock()
				return
			}

			m.mu.Lock()
			m.framesCaptured++
			frameNum := m.framesCaptured
			callback := m.onFrameCaptured
			m.mu.Unlock()

			if callback != nil {
				go callback(frameNum)
			}
		}
	}
}
