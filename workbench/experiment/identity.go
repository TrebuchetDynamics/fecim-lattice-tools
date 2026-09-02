package experiment

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"

	"fecim-lattice-tools/workbench/project"
)

func ID(point DesignPoint, evaluatorVersion string, inputs []project.ResolvedInput) (string, error) {
	ordered := append([]project.ResolvedInput(nil), inputs...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Path < ordered[j].Path })
	payload := struct {
		SchemaVersion    int                     `json:"schema_version"`
		Design           project.Design          `json:"design"`
		Seed             int64                   `json:"seed"`
		EvaluatorVersion string                  `json:"evaluator_version"`
		Inputs           []project.ResolvedInput `json:"inputs"`
	}{
		SchemaVersion:    1,
		Design:           point.Design,
		Seed:             point.Seed,
		EvaluatorVersion: evaluatorVersion,
		Inputs:           ordered,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}
