package logging

import (
	"sync"
	"time"
)

// LogLevel represents the severity of a log entry.
type LogLevel int

const (
	LevelTrace LogLevel = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
)

// Entry represents a single log entry stored in the global buffer.
type Entry struct {
	Timestamp time.Time
	Level     LogLevel
	Source    string
	Message   string
	Category  string
	Error     error
	Fields    map[string]interface{}
}

// NewEntry creates a new log entry.
func NewEntry(level LogLevel, source, message string) *Entry {
	return &Entry{
		Timestamp: time.Now(),
		Level:     level,
		Source:    source,
		Message:   message,
	}
}

// WithCategory sets the category on the entry and returns it for chaining.
func (e *Entry) WithCategory(cat string) *Entry {
	e.Category = cat
	return e
}

// WithError attaches an error to the entry and returns it for chaining.
func (e *Entry) WithError(err error) *Entry {
	e.Error = err
	return e
}

// WithField adds a single key-value pair and returns the entry for chaining.
func (e *Entry) WithField(key string, value interface{}) *Entry {
	if e.Fields == nil {
		e.Fields = make(map[string]interface{})
	}
	e.Fields[key] = value
	return e
}

// WithFields merges multiple key-value pairs and returns the entry for chaining.
func (e *Entry) WithFields(fields map[string]interface{}) *Entry {
	if e.Fields == nil {
		e.Fields = make(map[string]interface{}, len(fields))
	}
	for k, v := range fields {
		e.Fields[k] = v
	}
	return e
}

// Ring buffer for log entries accessible by the LogViewer.
const defaultBufferSize = 1000

var (
	bufferMu      sync.RWMutex
	buffer        []*Entry
	bufferSize    = defaultBufferSize
	bufferWriteAt int
	bufferCount   int
)

func init() {
	buffer = make([]*Entry, bufferSize)
}

// AddToBuffer appends an entry to the global ring buffer.
func AddToBuffer(entry *Entry) {
	if entry == nil {
		return
	}
	bufferMu.Lock()
	buffer[bufferWriteAt] = entry
	bufferWriteAt = (bufferWriteAt + 1) % bufferSize
	if bufferCount < bufferSize {
		bufferCount++
	}
	bufferMu.Unlock()
}

// ReadBuffer returns up to the last n entries from the buffer (oldest first).
func ReadBuffer(n int) []*Entry {
	bufferMu.RLock()
	defer bufferMu.RUnlock()

	if n <= 0 || bufferCount == 0 {
		return nil
	}
	if n > bufferCount {
		n = bufferCount
	}

	result := make([]*Entry, n)
	start := (bufferWriteAt - n + bufferSize) % bufferSize
	for i := 0; i < n; i++ {
		result[i] = buffer[(start+i)%bufferSize]
	}
	return result
}

// ClearBuffer resets the global log buffer.
func ClearBuffer() {
	bufferMu.Lock()
	buffer = make([]*Entry, bufferSize)
	bufferWriteAt = 0
	bufferCount = 0
	bufferMu.Unlock()
}
