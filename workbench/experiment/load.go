package experiment

import (
	"errors"
	"fmt"
	"os"
)

// LoadRunRecord verifies an exact committed run, including failed design points.
// A missing run returns false without an error; corrupt records return an error.
func LoadRunRecord(path, runID string) (RunRecord, bool, error) {
	if !runIDRE.MatchString(runID) {
		return RunRecord{}, false, fmt.Errorf("invalid run ID %q", runID)
	}
	root, err := os.OpenRoot(path)
	if errors.Is(err, os.ErrNotExist) {
		return RunRecord{}, false, nil
	}
	if err != nil {
		return RunRecord{}, false, err
	}
	defer root.Close()
	return LoadRunRecordRoot(root, runID)
}

// LoadRunRecordRoot verifies an exact committed run, including failed points.
// Missing runs return false without error. The caller retains ownership of root
// and must keep it open until this operation returns.
func LoadRunRecordRoot(root *os.Root, runID string) (RunRecord, bool, error) {
	storage, err := openStore(root)
	if err != nil {
		return RunRecord{}, false, err
	}
	defer storage.close()
	return storage.load(runID)
}

// LoadRun returns a successful run only after verifying its persisted hashes and identity.
func LoadRun(root, runID string) (RunRecord, error) {
	record, ok, err := LoadRunRecord(root, runID)
	return successfulRun(record, ok, runID, err)
}

// LoadRunRoot returns only a verified successful run. The caller owns root and
// must keep it open until this operation returns.
func LoadRunRoot(root *os.Root, runID string) (RunRecord, error) {
	record, ok, err := LoadRunRecordRoot(root, runID)
	return successfulRun(record, ok, runID, err)
}

func successfulRun(record RunRecord, ok bool, runID string, err error) (RunRecord, error) {
	if err != nil {
		return RunRecord{}, err
	}
	if !ok {
		return RunRecord{}, fmt.Errorf("run %s: %w", runID, os.ErrNotExist)
	}
	if record.Result.Status != StatusSuccess {
		return RunRecord{}, fmt.Errorf("run %s has status %q; only successful runs can be exported", runID, record.Result.Status)
	}
	return record, nil
}
