// Package moduleapi defines storage-independent browser/Worker contracts. Native
// routing and filesystem authority deliberately live outside this package.
package moduleapi

import (
	"context"
	"encoding/json"

	"fecim-lattice-tools/workbench/hysteresis"
	"fecim-lattice-tools/workbench/presentation"
)

const SchemaVersion = "fecim-module/1"

type Capability struct {
	Profile   string `json:"profile"`
	Available bool   `json:"available"`
	Reason    string `json:"reason,omitempty"`
}
type ModuleDescriptor struct {
	ID            string       `json:"id"`
	Title         string       `json:"title"`
	Available     bool         `json:"available"`
	Reason        string       `json:"reason,omitempty"`
	Features      []string     `json:"features"`
	Capabilities  []Capability `json:"capabilities"`
	SchemaVersion string       `json:"schema_version"`
	ModelVersion  string       `json:"model_version"`
}
type ResourceReference struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
}
type Command struct {
	Action    string              `json:"action"`
	Input     map[string]string   `json:"input,omitempty"`
	Revision  uint64              `json:"revision"`
	Resources []ResourceReference `json:"resources,omitempty"`
	// Payload retains the existing Go Hysteresis bridge protocol without parsing
	// numeric display strings. Other ports may use typed input strings instead.
	Payload json.RawMessage `json:"payload,omitempty"`
}
type Field struct {
	ID           string             `json:"id"`
	Label        string             `json:"label"`
	Input        string             `json:"input"`
	Unit         string             `json:"unit,omitempty"`
	Presentation presentation.Value `json:"presentation"`
}
type Action struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Available bool   `json:"available"`
	Reason    string `json:"reason,omitempty"`
}
type NumericCell struct {
	Value        *float64           `json:"value,omitempty"`
	Unit         string             `json:"unit"`
	Presentation presentation.Value `json:"presentation"`
	Evidence     string             `json:"evidence"`
}
type TablePage struct {
	ID      string          `json:"id"`
	Caption string          `json:"caption"`
	Page    int             `json:"page"`
	Total   int             `json:"total"`
	Columns []string        `json:"columns"`
	Rows    [][]NumericCell `json:"rows"`
}
type Plot struct {
	ID     string       `json:"id"`
	Title  string       `json:"title"`
	XUnit  string       `json:"x_unit"`
	YUnit  string       `json:"y_unit"`
	Points [][2]float64 `json:"points"`
}
type Snapshot struct {
	Session    string           `json:"session"`
	Revision   uint64           `json:"revision"`
	Descriptor ModuleDescriptor `json:"descriptor"`
	Status     string           `json:"status"`
	Evidence   []string         `json:"evidence"`
	Fields     []Field          `json:"fields"`
	Actions    []Action         `json:"actions"`
	Tables     []TablePage      `json:"tables"`
	Plots      []Plot           `json:"plots"`
	// Bridge is the original typed Go scientific result, not legacy Metric.Value.
	Bridge *hysteresis.Reply `json:"bridge,omitempty"`
}

// Port is a memory-only session. The native owner serializes calls and bounds
// session/job lifetimes; ports must honor context cancellation for long work.
type Port interface {
	Descriptor() ModuleDescriptor
	Snapshot(context.Context) (Snapshot, error)
	Apply(context.Context, Command) (Snapshot, error)
	Close()
}

// Provider is an optional self-supplying capability on an existing constructor.
// Implementing legacy viewmodel.ModulePort does not imply web readiness.
type Provider interface{ WebPort() Port }

type Progress struct {
	Total     int `json:"total"`
	Completed int `json:"completed"`
	Reused    int `json:"reused"`
	Failed    int `json:"failed"`
	Cancelled int `json:"cancelled"`
}
type Job struct {
	ID                    string         `json:"id"`
	Project               string         `json:"project,omitempty"`
	Session               string         `json:"session,omitempty"`
	State                 string         `json:"state"`
	Progress              Progress       `json:"progress"`
	CancellationRequested bool           `json:"cancellation_requested"`
	ResultsState          string         `json:"results_state"`
	InputRevision         string         `json:"input_revision"`
	AnalysisAvailable     bool           `json:"analysis_available"`
	Counts                map[string]int `json:"counts"`
	ResultsRevision       string         `json:"results_revision"`
	Message               string         `json:"message,omitempty"`
}

// JobPort is optional asynchronous work, with the same counts/terminal
// vocabulary in native and Worker transports. Disconnect never means cancel.
type JobPort interface {
	Port
	Start(context.Context, Command) (Job, error)
	Job(context.Context, string) (Job, error)
	Cancel(context.Context, string) (Job, error)
}
