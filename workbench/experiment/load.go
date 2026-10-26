package experiment

import (
	"fmt"
	"os"
)

// LoadRunRecord verifies an exact committed run, including failed design points.
// A missing run returns false without an error; corrupt records return an error.
func LoadRunRecord(root, runID string) (RunRecord, bool, error) {
	return (store{root: root}).load(runID)
}

// LoadRun returns a successful run only after verifying its persisted hashes and identity.
func LoadRun(root, runID string) (RunRecord, error) {
	record, ok, err := (store{root: root}).load(runID)
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
