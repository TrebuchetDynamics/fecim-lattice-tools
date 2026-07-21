package experiment

import (
	"context"
	"time"

	"fecim-lattice-tools/workbench/project"
)

type Status string

const (
	StatusSuccess Status = "success"
	StatusFailed  Status = "failed"
)

type Evidence string

const (
	EvidenceLiterature Evidence = "literature-backed"
	EvidenceExperiment Evidence = "experiment-calibrated"
	EvidenceDefault    Evidence = "simulation-default"
	EvidenceDerived    Evidence = "derived"
)

type Metric struct {
	Name        string   `json:"name"`
	Value       float64  `json:"value"`
	Unit        string   `json:"unit"`
	Model       string   `json:"model"`
	Assumptions []string `json:"assumptions"`
	Evidence    Evidence `json:"evidence"`
}

type Failure struct {
	Kind    string `json:"kind"`
	Message string `json:"message"`
}

type Result struct {
	Status   Status   `json:"status"`
	Metrics  []Metric `json:"metrics"`
	Warnings []string `json:"warnings,omitempty"`
	Failure  *Failure `json:"failure,omitempty"`
}

type DesignPoint struct {
	Index  int            `json:"index"`
	Design project.Design `json:"design"`
	Seed   int64          `json:"seed"`
	RunID  string         `json:"run_id,omitempty"`
}

type RunManifest struct {
	SchemaVersion      int                     `json:"schema_version"`
	RunID              string                  `json:"run_id"`
	PointIndex         int                     `json:"point_index"`
	Seed               int64                   `json:"seed"`
	EvaluatorVersion   string                  `json:"evaluator_version"`
	RepositoryRevision string                  `json:"repository_revision,omitempty"`
	Inputs             []project.ResolvedInput `json:"inputs"`
	StartedAt          time.Time               `json:"started_at"`
	CompletedAt        time.Time               `json:"completed_at"`
	Status             Status                  `json:"status"`
}

type RunRecord struct {
	Manifest RunManifest    `json:"manifest"`
	Design   project.Design `json:"design"`
	Result   Result         `json:"result"`
	Reused   bool           `json:"-"`
}

type Summary struct {
	Runs []RunRecord `json:"runs"`
}

type Evaluator func(context.Context, project.Design, int64) (Result, error)
