package experiment

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

var runIDRE = regexp.MustCompile(`^[0-9a-f]{64}$`)

type store struct {
	root string
}

func (s store) load(id string) (RunRecord, bool, error) {
	if !runIDRE.MatchString(id) {
		return RunRecord{}, false, fmt.Errorf("invalid run ID %q", id)
	}
	dir := filepath.Join(s.root, "runs", id)
	info, err := os.Stat(dir)
	if errors.Is(err, os.ErrNotExist) {
		return RunRecord{}, false, nil
	}
	if err != nil {
		return RunRecord{}, false, err
	}
	if !info.IsDir() {
		return RunRecord{}, false, fmt.Errorf("run path %s is not a directory", dir)
	}
	var record RunRecord
	if err := decodeJSON(filepath.Join(dir, "manifest.json"), &record.Manifest); err != nil {
		return RunRecord{}, false, fmt.Errorf("load run %s manifest: %w", id, err)
	}
	if err := decodeJSON(filepath.Join(dir, "resolved-design.json"), &record.Design); err != nil {
		return RunRecord{}, false, fmt.Errorf("load run %s design: %w", id, err)
	}
	if err := decodeJSON(filepath.Join(dir, "result.json"), &record.Result); err != nil {
		return RunRecord{}, false, fmt.Errorf("load run %s result: %w", id, err)
	}
	if record.Manifest.RunID != id || record.Manifest.Status != record.Result.Status {
		return RunRecord{}, false, fmt.Errorf("run %s has inconsistent identity or status", id)
	}
	return record, true, nil
}

func (s store) commit(record RunRecord) (RunRecord, error) {
	id := record.Manifest.RunID
	if !runIDRE.MatchString(id) {
		return RunRecord{}, fmt.Errorf("invalid run ID %q", id)
	}
	runsDir := filepath.Join(s.root, "runs")
	if err := os.MkdirAll(runsDir, 0o755); err != nil {
		return RunRecord{}, err
	}
	if existing, ok, err := s.load(id); err != nil {
		return RunRecord{}, err
	} else if ok {
		existing.Reused = true
		return existing, nil
	}

	tempDir, err := os.MkdirTemp(runsDir, ".tmp-"+id+"-")
	if err != nil {
		return RunRecord{}, err
	}
	defer os.RemoveAll(tempDir)
	for name, value := range map[string]any{
		"manifest.json":        record.Manifest,
		"resolved-design.json": record.Design,
		"result.json":          record.Result,
	} {
		if err := writeJSON(filepath.Join(tempDir, name), value); err != nil {
			return RunRecord{}, err
		}
	}
	finalDir := filepath.Join(runsDir, id)
	if err := os.Rename(tempDir, finalDir); err != nil {
		if existing, ok, loadErr := s.load(id); loadErr == nil && ok {
			existing.Reused = true
			return existing, nil
		}
		return RunRecord{}, err
	}
	return record, nil
}

func writeJSON(path string, value any) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}

func decodeJSON(path string, target any) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("multiple JSON values are not allowed")
	}
	return nil
}
