package logging

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	hysteresisRetentionMarker = ".hysteresis-retention-baseline"
	maxHysteresisLogFiles     = 100
	maxHysteresisLogBytes     = int64(2 << 30)
	maxHysteresisLogAge       = 180 * 24 * time.Hour
)

var hysteresisLogMu sync.Mutex
var activeHysteresisLogs = make(map[string]struct{})
var lastHysteresisPrune = make(map[string]time.Time)

// RegisterHysteresisLog protects the new path and prunes only logs created
// after retention was enabled. Existing user logs are deliberately preserved.
func RegisterHysteresisLog(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if err := ensureHysteresisRetentionMarker(filepath.Dir(absPath)); err != nil {
		return err
	}
	hysteresisLogMu.Lock()
	defer hysteresisLogMu.Unlock()
	activeHysteresisLogs[absPath] = struct{}{}
	dir := filepath.Dir(absPath)
	if last := lastHysteresisPrune[dir]; time.Since(last) < 24*time.Hour {
		return nil
	}
	if err := pruneHysteresisLogs(dir); err != nil {
		delete(activeHysteresisLogs, absPath)
		return err
	}
	lastHysteresisPrune[dir] = time.Now()
	return nil
}

func UnregisterHysteresisLog(path string) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return
	}
	hysteresisLogMu.Lock()
	delete(activeHysteresisLogs, absPath)
	hysteresisLogMu.Unlock()
}

func ensureHysteresisRetentionMarker(dir string) error {
	path := filepath.Join(dir, hysteresisRetentionMarker)
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err == nil {
		return file.Close()
	}
	if errors.Is(err, os.ErrExist) {
		return nil
	}
	return err
}

type hysteresisLogFile struct {
	path    string
	modTime time.Time
	size    int64
}

func pruneHysteresisLogs(dir string) error {
	marker, err := os.Stat(filepath.Join(dir, hysteresisRetentionMarker))
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	managed := make([]hysteresisLogFile, 0)
	eligible := make([]hysteresisLogFile, 0)
	var totalBytes int64
	cutoff := marker.ModTime()
	now := time.Now()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "hysteresis-") || !strings.HasSuffix(entry.Name(), ".csv") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.ModTime().After(cutoff) {
			continue
		}
		file := hysteresisLogFile{path: filepath.Join(dir, entry.Name()), modTime: info.ModTime(), size: info.Size()}
		managed = append(managed, file)
		totalBytes += info.Size()
		_, active := activeHysteresisLogs[file.path]
		// ponytail: wait one day before pruning another process's file; use lock/heartbeat files if multi-day concurrent runs become supported.
		if !active && now.Sub(file.modTime) >= 24*time.Hour {
			eligible = append(eligible, file)
		}
	}
	sort.Slice(eligible, func(i, j int) bool { return eligible[i].modTime.Before(eligible[j].modTime) })
	managedCount := len(managed)
	for _, file := range eligible {
		expired := now.Sub(file.modTime) > maxHysteresisLogAge
		if !expired && managedCount <= maxHysteresisLogFiles && totalBytes <= maxHysteresisLogBytes {
			continue
		}
		if err := os.Remove(file.path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		managedCount--
		totalBytes -= file.size
	}
	return nil
}
