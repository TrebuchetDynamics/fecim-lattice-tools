//go:build legacy_fyne

// Package computelog contains compute-log storage and persistence helpers for module 4.
package computelog

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	sharedio "fecim-lattice-tools/shared/io"
	"fecim-lattice-tools/shared/logging"
)

// Entry represents a single MVM compute operation.
type Entry struct {
	Timestamp    string      `json:"timestamp"`
	ArraySize    string      `json:"array_size"`
	Material     string      `json:"material"`
	QuantLevels  int         `json:"quant_levels"`
	InputVector  []float64   `json:"input_vector_volts"`
	Weights      [][]int     `json:"weight_matrix"`
	Conductances [][]float64 `json:"conductance_matrix_uS"`
	RowResults   []RowResult `json:"row_results"`
}

// RowResult holds the compute result for a single row.
type RowResult struct {
	Row        int       `json:"row"`
	Active     bool      `json:"active"`
	CurrentUA  float64   `json:"current_uA"`
	TIAVoltage float64   `json:"tia_voltage_V"`
	ADCLevel   int       `json:"adc_level"`
	Saturated  bool      `json:"saturated"`
	CellDetail []CellMVM `json:"cell_details"`
}

// CellMVM holds per-cell MVM calculation details.
type CellMVM struct {
	Col           int     `json:"col"`
	Weight        int     `json:"weight"`
	ConductanceUS float64 `json:"conductance_uS"`
	VoltageV      float64 `json:"voltage_V"`
	CurrentUA     float64 `json:"current_uA"`
}

// Log manages compute log entries and output path.
type Log struct {
	mu       sync.Mutex
	entries  []Entry
	filePath string
	enabled  bool
}

// New returns a compute log with the default file path and environment-controlled enablement.
func New() *Log {
	return &Log{
		entries:  make([]Entry, 0),
		filePath: filepath.Join(logging.LogsDir(), "compute_log.json"),
		enabled:  os.Getenv("FECIM_CIRCUITS_COMPUTE_LOG") != "",
	}
}

// Enable enables or disables compute logging.
func (l *Log) Enable(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.enabled = enabled
}

// Enabled reports whether compute logging is currently enabled.
func (l *Log) Enabled() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.enabled
}

// SetPath sets the path for the compute log file.
func (l *Log) SetPath(path string) error {
	cleanPath, err := sharedio.ValidatePath(path)
	if err != nil {
		return fmt.Errorf("invalid compute log path: %w", err)
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.filePath = cleanPath
	return nil
}

// Clear removes all logged entries.
func (l *Log) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = make([]Entry, 0)
}

// Entries returns a copy of all logged entries.
func (l *Log) Entries() []Entry {
	l.mu.Lock()
	defer l.mu.Unlock()
	result := make([]Entry, len(l.entries))
	copy(result, l.entries)
	return result
}

// Append stores one log entry and keeps only the newest limit entries when limit is positive.
func (l *Log) Append(entry Entry, limit int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, entry)
	if limit > 0 && len(l.entries) > limit {
		l.entries = l.entries[len(l.entries)-limit:]
	}
}

// Save saves all logged entries to the configured JSON file.
func (l *Log) Save() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return saveEntries(l.filePath, l.entries)
}

// SaveTo saves all logged entries to a specific file.
func (l *Log) SaveTo(path string) error {
	entries := l.Entries()
	return saveEntries(path, entries)
}

func saveEntries(path string, entries []Entry) error {
	if len(entries) == 0 {
		return nil
	}
	if err := sharedio.SaveJSON(path, entries); err != nil {
		return fmt.Errorf("failed to write compute log: %w", err)
	}
	return nil
}
