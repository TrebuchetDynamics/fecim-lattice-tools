# Research and Design Workbench Core Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first working research-design slice: load a file-based project, expand a deterministic device/array/circuit sweep, execute existing Go cost models, persist immutable runs, calculate feasibility/Pareto results, and expose reproducible reports through the existing CLI.

**Architecture:** New top-level `workbench/` packages own project loading, experiments, the FeCIM evaluator, and reports. They call existing validated `shared/system` and `shared/peripherals` models through one evaluator function; they do not move or duplicate model code. The existing `cmd/fecim-lattice-tools` command dispatch becomes the first client.

**Tech Stack:** Go 1.25, Go standard library, installed `gopkg.in/yaml.v3`, existing `shared/system` and `shared/peripherals` packages, standard `testing` package.

## Global Constraints

- Follow test-driven development: record RED before production changes, then GREEN, then refactor.
- Keep Go as the headless-core language; add no Rust component.
- Add no database, service, distributed runner, optimizer framework, or dependency.
- Use plain YAML inputs and JSON/CSV/Markdown outputs.
- Use canonical SI units internally; configuration field names carry explicit units.
- Keep completed successful and failed-design-point runs immutable.
- Treat non-finite numbers, unknown YAML fields, unsafe paths, and hash mismatches as trust-boundary errors.
- Keep the current seven GUI modules unchanged in this plan.
- Preserve all existing model and validation tolerances.

---

## Scope Decomposition

The approved design covers three independently reviewable clients/integrations. Implement them as three plans:

1. **This plan:** headless project/experiment/evaluator/report core, CLI, and example project.
2. **Follow-up plan:** Fyne Define/Simulate/Analyze/Validate workflow and CLI/Fyne run-ID parity.
3. **Follow-up plan:** selected-run SPICE/Verilog/Liberty export integration.

This plan is independently useful: it ends with a reproducible command-line parameter sweep and trade-off report. It intentionally does not claim the full design specification is complete.

## File Structure

Create:

- `workbench/project/types.go` — versioned project, design, sweep, objective, constraint, and input types.
- `workbench/project/load.go` — strict YAML loading, path/hash/citation validation, and bundle resolution.
- `workbench/project/load_test.go` — loader trust-boundary and schema tests.
- `workbench/experiment/types.go` — design points, metrics, failures, run manifests, and summaries.
- `workbench/experiment/sweep.go` — deterministic range expansion and design-path assignment.
- `workbench/experiment/sweep_test.go` — order, range, limit, and invalid-path tests.
- `workbench/experiment/identity.go` — canonical run identity.
- `workbench/experiment/identity_test.go` — stable identity and behavior-affecting input tests.
- `workbench/fecim/evaluator.go` — thin evaluator over existing system/peripheral models.
- `workbench/fecim/evaluator_test.go` — golden metric, units, warning, and domain-failure tests.
- `workbench/experiment/store.go` — immutable atomic run persistence.
- `workbench/experiment/runner.go` — reuse, evaluation, bounded workers, cancellation, and resume.
- `workbench/experiment/runner_test.go` — persistence, continuation, systemic halt, worker determinism, and cancellation tests.
- `workbench/experiment/analyze.go` — feasibility, constraints, and generic Pareto membership.
- `workbench/experiment/analyze_test.go` — classification and objective-direction tests.
- `workbench/report/report.go` — deterministic JSON, CSV, and Markdown reports.
- `workbench/report/report_test.go` — golden report semantics and provenance tests.
- `cmd/fecim-lattice-tools/workbench_subcommand.go` — `project`, `experiment`, and `report` commands.
- `cmd/fecim-lattice-tools/workbench_subcommand_test.go` — command routing and end-to-end CLI tests.
- `examples/research-design-workbench/project.yaml` — example hypothesis, objectives, and constraints.
- `examples/research-design-workbench/design.yaml` — baseline HZO/array/peripheral design.
- `examples/research-design-workbench/sweep.yaml` — finite device/array/circuit grid.
- `docs/guides/research-design-workbench.md` — first-slice user guide and trust boundary.

Modify:

- `cmd/fecim-lattice-tools/subcommands.go` — dispatch and advertise new commands.
- `README.md` — position the research-design workflow and link the guide.
- `TODO.md` — record the completed core slice and remaining Fyne/export follow-ups.

## Shared Interfaces

These signatures are fixed for this plan so tasks compose without redesign:

```go
// workbench/project
func Load(root string, opts LoadOptions) (Bundle, error)
func ValidateDesign(d Design) error

// workbench/experiment
type Evaluator func(context.Context, project.Design, int64) (Result, error)
func Expand(bundle project.Bundle) ([]DesignPoint, error)
func ID(point DesignPoint, evaluatorVersion string, inputs []project.ResolvedInput) (string, error)
func Run(ctx context.Context, bundle project.Bundle, opts RunOptions) (Summary, error)
func Analyze(bundle project.Bundle, runs []RunRecord) Analysis

// workbench/fecim
const EvaluatorVersion = "fecim-system-v1"
func Evaluate(context.Context, project.Design, int64) (experiment.Result, error)

// workbench/report
func WriteJSON(io.Writer, experiment.Analysis) error
func WriteCSV(io.Writer, experiment.Analysis) error
func WriteMarkdown(io.Writer, project.Bundle, experiment.Analysis) error
```

---

### Task 1: Strict Project Bundle Loading

**Files:**

- Create: `workbench/project/types.go`
- Create: `workbench/project/load.go`
- Test: `workbench/project/load_test.go`

**Interfaces:**

- Consumes: `gopkg.in/yaml.v3`; filesystem paths supplied by callers.
- Produces: `Bundle`, `Design`, `ResolvedInput`, `LoadOptions`, and `Load(root, opts)` for all later tasks.

- [ ] **Step 1: Write failing loader tests**

Create `workbench/project/load_test.go` with table-driven fixtures written under `t.TempDir()`:

```go
package project

import (
    "crypto/sha256"
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "testing"
)

func writeBundle(t *testing.T, root, projectYAML, designYAML, sweepYAML string) {
    t.Helper()
    for name, body := range map[string]string{
        "project.yaml": projectYAML,
        "design.yaml":  designYAML,
        "sweep.yaml":   sweepYAML,
    } {
        if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0o644); err != nil {
            t.Fatal(err)
        }
    }
}

const validProjectYAML = `schema_version: 1
id: hzo-study
name: HZO study
hypothesis: Increasing ADC resolution trades area for energy fidelity.
model_version: fecim-system-v1
objectives:
  - metric: energy_pj
    direction: minimize
  - metric: latency_ns
    direction: minimize
constraints:
  - metric: area_um2
    operator: <=
    value: 5000
    unit: um2
citations: [park2015_advmat_hzo]
`

const validDesignYAML = `schema_version: 1
device:
  material: HZO
  conductance_levels: 30
  g_min_s: 0.000001
  g_max_s: 0.00003
array:
  rows: 32
  cols: 32
  read_voltage_v: 0.2
circuit:
  adc_bits: 6
  dac_bits: 4
  tia_gain_ohm: 10000
  tech_node: 65nm
`

const validSweepYAML = `schema_version: 1
seed: 17
max_points: 32
parameters:
  - path: device.conductance_levels
    values: [16, 30]
  - path: array.rows
    range: {start: 16, stop: 32, count: 2}
  - path: circuit.adc_bits
    values: [4, 6]
`

func TestLoadStrictBundle(t *testing.T) {
    root := t.TempDir()
    citations := filepath.Join(root, "citations")
    if err := os.Mkdir(citations, 0o755); err != nil { t.Fatal(err) }
    if err := os.WriteFile(filepath.Join(citations, "park2015_advmat_hzo.md"), []byte("# Park"), 0o644); err != nil { t.Fatal(err) }
    writeBundle(t, root, validProjectYAML, validDesignYAML, validSweepYAML)

    got, err := Load(root, LoadOptions{CitationDir: citations})
    if err != nil { t.Fatalf("Load: %v", err) }
    if got.Project.ID != "hzo-study" || got.Design.Array.Rows != 32 || got.Sweep.Seed != 17 {
        t.Fatalf("unexpected bundle: %+v", got)
    }
}

func TestLoadRejectsUnknownYAMLField(t *testing.T) {
    root := t.TempDir()
    writeBundle(t, root, validProjectYAML+"mystery: true\n", validDesignYAML, validSweepYAML)
    _, err := Load(root, LoadOptions{})
    if err == nil || !strings.Contains(err.Error(), "field mystery not found") {
        t.Fatalf("err=%v, want unknown-field error", err)
    }
}

func TestLoadValidatesInputHashAndContainment(t *testing.T) {
    root := t.TempDir()
    data := []byte("measured values\n")
    if err := os.Mkdir(filepath.Join(root, "inputs"), 0o755); err != nil { t.Fatal(err) }
    if err := os.WriteFile(filepath.Join(root, "inputs", "sample.csv"), data, 0o644); err != nil { t.Fatal(err) }
    digest := fmt.Sprintf("%x", sha256.Sum256(data))
    p := validProjectYAML + fmt.Sprintf("inputs:\n  - path: inputs/sample.csv\n    sha256: %s\n    citation: park2015_advmat_hzo\n    evidence: experiment-calibrated\n", digest)
    writeBundle(t, root, p, validDesignYAML, validSweepYAML)

    got, err := Load(root, LoadOptions{})
    if err != nil { t.Fatalf("Load: %v", err) }
    if len(got.Inputs) != 1 || got.Inputs[0].SHA256 != digest {
        t.Fatalf("inputs=%+v", got.Inputs)
    }

    writeBundle(t, root, strings.Replace(p, digest, strings.Repeat("0", 64), 1), validDesignYAML, validSweepYAML)
    if _, err := Load(root, LoadOptions{}); err == nil || !strings.Contains(err.Error(), "sha256") {
        t.Fatalf("err=%v, want digest mismatch", err)
    }
}

func TestLoadRejectsInvalidDesignTrustBoundary(t *testing.T) {
    root := t.TempDir()
    bad := strings.Replace(validDesignYAML, "g_max_s: 0.00003", "g_max_s: .nan", 1)
    writeBundle(t, root, validProjectYAML, bad, validSweepYAML)
    if _, err := Load(root, LoadOptions{}); err == nil {
        t.Fatal("Load succeeded with non-finite conductance")
    }
}
```

- [ ] **Step 2: Run the tests and record RED**

Run:

```bash
go test ./workbench/project -count=1
```

Expected: FAIL because `workbench/project` does not exist.

- [ ] **Step 3: Add the versioned project types**

Create `workbench/project/types.go`:

```go
package project

type Direction string
const (
    Minimize Direction = "minimize"
    Maximize Direction = "maximize"
)

type Objective struct {
    Metric string    `yaml:"metric" json:"metric"`
    Direction Direction `yaml:"direction" json:"direction"`
}

type Constraint struct {
    Metric string  `yaml:"metric" json:"metric"`
    Operator string `yaml:"operator" json:"operator"`
    Value float64  `yaml:"value" json:"value"`
    Unit string    `yaml:"unit" json:"unit"`
}

type InputRef struct {
    Path string `yaml:"path" json:"path"`
    SHA256 string `yaml:"sha256" json:"sha256"`
    Citation string `yaml:"citation" json:"citation"`
    Evidence string `yaml:"evidence" json:"evidence"`
}

type Project struct {
    SchemaVersion int `yaml:"schema_version" json:"schema_version"`
    ID string `yaml:"id" json:"id"`
    Name string `yaml:"name" json:"name"`
    Hypothesis string `yaml:"hypothesis" json:"hypothesis"`
    ModelVersion string `yaml:"model_version" json:"model_version"`
    Objectives []Objective `yaml:"objectives" json:"objectives"`
    Constraints []Constraint `yaml:"constraints" json:"constraints"`
    Citations []string `yaml:"citations" json:"citations"`
    Inputs []InputRef `yaml:"inputs" json:"inputs"`
}

type Device struct {
    Material string `yaml:"material" json:"material"`
    ConductanceLevels int `yaml:"conductance_levels" json:"conductance_levels"`
    GMinS float64 `yaml:"g_min_s" json:"g_min_s"`
    GMaxS float64 `yaml:"g_max_s" json:"g_max_s"`
}

type Array struct {
    Rows int `yaml:"rows" json:"rows"`
    Cols int `yaml:"cols" json:"cols"`
    ReadVoltageV float64 `yaml:"read_voltage_v" json:"read_voltage_v"`
}

type Circuit struct {
    ADCBits int `yaml:"adc_bits" json:"adc_bits"`
    DACBits int `yaml:"dac_bits" json:"dac_bits"`
    TIAGainOhm float64 `yaml:"tia_gain_ohm" json:"tia_gain_ohm"`
    TechNode string `yaml:"tech_node" json:"tech_node"`
}

type Design struct {
    SchemaVersion int `yaml:"schema_version" json:"schema_version"`
    Device Device `yaml:"device" json:"device"`
    Array Array `yaml:"array" json:"array"`
    Circuit Circuit `yaml:"circuit" json:"circuit"`
}

type LinearRange struct {
    Start float64 `yaml:"start" json:"start"`
    Stop float64 `yaml:"stop" json:"stop"`
    Count int `yaml:"count" json:"count"`
}

type Parameter struct {
    Path string `yaml:"path" json:"path"`
    Values []float64 `yaml:"values" json:"values"`
    Range *LinearRange `yaml:"range" json:"range,omitempty"`
}

type Sweep struct {
    SchemaVersion int `yaml:"schema_version" json:"schema_version"`
    Seed int64 `yaml:"seed" json:"seed"`
    MaxPoints int `yaml:"max_points" json:"max_points"`
    Parameters []Parameter `yaml:"parameters" json:"parameters"`
}

type ResolvedInput struct {
    Path string `json:"path"`
    SHA256 string `json:"sha256"`
    Citation string `json:"citation"`
    Evidence string `json:"evidence"`
}

type Bundle struct {
    Root string
    Project Project
    Design Design
    Sweep Sweep
    Inputs []ResolvedInput
}

type LoadOptions struct { CitationDir string }
```

- [ ] **Step 4: Implement strict loading and validation**

Create `workbench/project/load.go`. Implement these exact helpers and rules:

```go
package project

import (
    "crypto/sha256"
    "encoding/hex"
    "errors"
    "fmt"
    "io"
    "math"
    "os"
    "path/filepath"
    "regexp"
    "strings"

    "gopkg.in/yaml.v3"
)

const maxProjectFileBytes int64 = 1 << 20
var idRE = regexp.MustCompile(`^[a-z][a-z0-9-]{1,62}$`)
var citationRE = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]*$`)

func Load(root string, opts LoadOptions) (Bundle, error) {
    abs, err := filepath.Abs(root)
    if err != nil { return Bundle{}, err }
    var b Bundle
    b.Root = abs
    if err := decodeStrict(filepath.Join(abs, "project.yaml"), &b.Project); err != nil { return Bundle{}, fmt.Errorf("project.yaml: %w", err) }
    if err := decodeStrict(filepath.Join(abs, "design.yaml"), &b.Design); err != nil { return Bundle{}, fmt.Errorf("design.yaml: %w", err) }
    if err := decodeStrict(filepath.Join(abs, "sweep.yaml"), &b.Sweep); err != nil { return Bundle{}, fmt.Errorf("sweep.yaml: %w", err) }
    if b.Project.SchemaVersion != 1 || b.Design.SchemaVersion != 1 || b.Sweep.SchemaVersion != 1 {
        return Bundle{}, errors.New("all schema_version values must be 1")
    }
    if !idRE.MatchString(b.Project.ID) || strings.TrimSpace(b.Project.Hypothesis) == "" || len(b.Project.Objectives) == 0 {
        return Bundle{}, errors.New("project requires a valid id, hypothesis, and objective")
    }
    if b.Project.ModelVersion == "" { return Bundle{}, errors.New("project model_version is required") }
    for _, o := range b.Project.Objectives {
        if o.Metric == "" || (o.Direction != Minimize && o.Direction != Maximize) { return Bundle{}, fmt.Errorf("invalid objective %+v", o) }
    }
    for _, c := range b.Project.Constraints {
        if c.Metric == "" || c.Unit == "" || !finite(c.Value) { return Bundle{}, fmt.Errorf("invalid constraint %+v", c) }
        switch c.Operator { case "<", "<=", ">", ">=", "==": default: return Bundle{}, fmt.Errorf("unsupported constraint operator %q", c.Operator) }
    }
    if err := ValidateDesign(b.Design); err != nil { return Bundle{}, err }
    if b.Sweep.MaxPoints <= 0 { b.Sweep.MaxPoints = 10_000 }
    if b.Sweep.MaxPoints > 100_000 { return Bundle{}, errors.New("max_points must not exceed 100000") }
    if len(b.Sweep.Parameters) == 0 { return Bundle{}, errors.New("sweep requires at least one parameter") }
    for _, key := range append(append([]string{}, b.Project.Citations...), inputCitations(b.Project.Inputs)... ) {
        if !citationRE.MatchString(key) { return Bundle{}, fmt.Errorf("invalid citation key %q", key) }
        if opts.CitationDir != "" {
            if _, err := os.Stat(filepath.Join(opts.CitationDir, key+".md")); err != nil { return Bundle{}, fmt.Errorf("citation %s: %w", key, err) }
        }
    }
    inputs, err := resolveInputs(abs, b.Project.Inputs)
    if err != nil { return Bundle{}, err }
    b.Inputs = inputs
    return b, nil
}

func decodeStrict(path string, dst any) error {
    st, err := os.Stat(path)
    if err != nil { return err }
    if st.Size() > maxProjectFileBytes { return fmt.Errorf("file exceeds %d bytes", maxProjectFileBytes) }
    f, err := os.Open(path)
    if err != nil { return err }
    defer f.Close()
    dec := yaml.NewDecoder(io.LimitReader(f, maxProjectFileBytes+1))
    dec.KnownFields(true)
    if err := dec.Decode(dst); err != nil { return err }
    var extra any
    if err := dec.Decode(&extra); err != io.EOF { return errors.New("multiple YAML documents are not allowed") }
    return nil
}

func ValidateDesign(d Design) error {
    if d.Device.Material == "" || d.Device.ConductanceLevels < 2 || !finite(d.Device.GMinS) || !finite(d.Device.GMaxS) || d.Device.GMinS <= 0 || d.Device.GMaxS <= d.Device.GMinS {
        return errors.New("invalid device configuration")
    }
    if d.Array.Rows <= 0 || d.Array.Cols <= 0 || !finite(d.Array.ReadVoltageV) || d.Array.ReadVoltageV <= 0 {
        return errors.New("invalid array configuration")
    }
    if d.Circuit.ADCBits < 1 || d.Circuit.ADCBits > 16 || d.Circuit.DACBits < 1 || d.Circuit.DACBits > 16 || !finite(d.Circuit.TIAGainOhm) || d.Circuit.TIAGainOhm <= 0 {
        return errors.New("invalid circuit configuration")
    }
    switch d.Circuit.TechNode { case "130nm", "65nm", "28nm", "22nm", "14nm": default: return fmt.Errorf("unsupported tech_node %q", d.Circuit.TechNode) }
    return nil
}

func resolveInputs(root string, refs []InputRef) ([]ResolvedInput, error) {
    out := make([]ResolvedInput, 0, len(refs))
    for _, ref := range refs {
        clean := filepath.Clean(ref.Path)
        if filepath.IsAbs(clean) || clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) { return nil, fmt.Errorf("unsafe input path %q", ref.Path) }
        path := filepath.Join(root, clean)
        resolved, err := filepath.EvalSymlinks(path)
        if err != nil { return nil, err }
        rel, err := filepath.Rel(root, resolved)
        if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) { return nil, fmt.Errorf("input escapes project root: %q", ref.Path) }
        data, err := os.ReadFile(resolved)
        if err != nil { return nil, err }
        sum := sha256.Sum256(data)
        actual := hex.EncodeToString(sum[:])
        if actual != strings.ToLower(ref.SHA256) { return nil, fmt.Errorf("input %s sha256 mismatch: got %s", ref.Path, actual) }
        switch ref.Evidence { case "literature-backed", "experiment-calibrated", "simulation-default": default: return nil, fmt.Errorf("input %s has invalid evidence %q", ref.Path, ref.Evidence) }
        out = append(out, ResolvedInput{Path: filepath.ToSlash(clean), SHA256: actual, Citation: ref.Citation, Evidence: ref.Evidence})
    }
    return out, nil
}

func inputCitations(in []InputRef) []string { out := make([]string, 0, len(in)); for _, v := range in { if v.Citation != "" { out = append(out, v.Citation) } }; return out }
func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
```

- [ ] **Step 5: Run GREEN and commit**

```bash
gofmt -w workbench/project/*.go
go test ./workbench/project -count=1
git add workbench/project
git commit -m "feat(workbench): load strict project bundles" -m "RED: workbench/project was absent. GREEN: strict schema, input hash, citation, and unit-boundary tests pass."
```

Expected: PASS.

---

### Task 2: Deterministic Sweep Expansion and Run Identity

**Files:**

- Create: `workbench/experiment/types.go`
- Create: `workbench/experiment/sweep.go`
- Create: `workbench/experiment/sweep_test.go`
- Create: `workbench/experiment/identity.go`
- Create: `workbench/experiment/identity_test.go`

**Interfaces:**

- Consumes: `project.Bundle`, `project.Design`, `project.ResolvedInput`.
- Produces: `DesignPoint`, `Metric`, `Result`, `RunRecord`, `Expand`, and `ID`.

- [ ] **Step 1: Write RED tests for deterministic expansion**

Create tests that assert the exact path order, Cartesian order, seeds `17..24`, linear range values `[16,32]`, rejection of duplicate paths, non-integral integer targets, unknown paths, non-finite values, and `max_points` overflow. Core assertion:

```go
func fixtureBundle() project.Bundle {
    return project.Bundle{
        Root: "/tmp/project",
        Project: project.Project{
            SchemaVersion: 1,
            ID: "hzo-study",
            Hypothesis: "ADC resolution trades area for energy fidelity.",
            ModelVersion: "fecim-system-v1",
            Objectives: []project.Objective{
                {Metric: "energy_pj", Direction: project.Minimize},
                {Metric: "latency_ns", Direction: project.Minimize},
            },
        },
        Design: project.Design{
            SchemaVersion: 1,
            Device: project.Device{Material: "HZO", ConductanceLevels: 30, GMinS: 1e-6, GMaxS: 30e-6},
            Array: project.Array{Rows: 32, Cols: 32, ReadVoltageV: 0.2},
            Circuit: project.Circuit{ADCBits: 6, DACBits: 4, TIAGainOhm: 10_000, TechNode: "65nm"},
        },
        Sweep: project.Sweep{
            SchemaVersion: 1,
            Seed: 17,
            MaxPoints: 32,
            Parameters: []project.Parameter{
                {Path: "device.conductance_levels", Values: []float64{16, 30}},
                {Path: "array.rows", Range: &project.LinearRange{Start: 16, Stop: 32, Count: 2}},
                {Path: "circuit.adc_bits", Values: []float64{4, 6}},
            },
        },
    }
}

func TestExpandDeterministicCartesianOrder(t *testing.T) {
    b := fixtureBundle()
    points, err := Expand(b)
    if err != nil { t.Fatal(err) }
    if len(points) != 8 { t.Fatalf("points=%d want 8", len(points)) }
    got := []string{}
    for _, p := range points {
        got = append(got, fmt.Sprintf("%d/%d/%d/%d", p.Design.Device.ConductanceLevels, p.Design.Array.Rows, p.Design.Circuit.ADCBits, p.Seed))
    }
    want := []string{"16/16/4/17", "16/16/6/18", "16/32/4/19", "16/32/6/20", "30/16/4/21", "30/16/6/22", "30/32/4/23", "30/32/6/24"}
    if !reflect.DeepEqual(got, want) { t.Fatalf("got=%v want=%v", got, want) }
}
```

- [ ] **Step 2: Write RED tests for stable identity**

```go
func TestIDStableAndBehaviorSensitive(t *testing.T) {
    point := DesignPoint{Index: 0, Design: fixtureBundle().Design, Seed: 17}
    inputs := []project.ResolvedInput{{Path: "inputs/a.csv", SHA256: strings.Repeat("a", 64), Citation: "park2015_advmat_hzo", Evidence: "literature-backed"}}
    first, err := ID(point, "fecim-system-v1", inputs)
    if err != nil { t.Fatal(err) }
    second, _ := ID(point, "fecim-system-v1", append([]project.ResolvedInput(nil), inputs...))
    if first != second || len(first) != 64 { t.Fatalf("ids=%q/%q", first, second) }
    point.Design.Circuit.ADCBits++
    changed, _ := ID(point, "fecim-system-v1", inputs)
    if changed == first { t.Fatal("ADC change did not change identity") }
}
```

- [ ] **Step 3: Run RED**

```bash
go test ./workbench/experiment -run 'TestExpand|TestID' -count=1
```

Expected: FAIL because the package is absent.

- [ ] **Step 4: Add experiment types**

Create `types.go` with these exact public contracts:

```go
package experiment

import (
    "context"
    "time"
    "fecim-lattice-tools/workbench/project"
)

type Status string
const ( StatusSuccess Status = "success"; StatusFailed Status = "failed" )
type Evidence string
const ( EvidenceLiterature Evidence = "literature-backed"; EvidenceExperiment Evidence = "experiment-calibrated"; EvidenceDefault Evidence = "simulation-default"; EvidenceDerived Evidence = "derived" )

type Metric struct { Name string `json:"name"`; Value float64 `json:"value"`; Unit string `json:"unit"`; Model string `json:"model"`; Assumptions []string `json:"assumptions"`; Evidence Evidence `json:"evidence"` }
type Failure struct { Kind string `json:"kind"`; Message string `json:"message"` }
type Result struct { Status Status `json:"status"`; Metrics []Metric `json:"metrics"`; Warnings []string `json:"warnings,omitempty"`; Failure *Failure `json:"failure,omitempty"` }
type DesignPoint struct { Index int `json:"index"`; Design project.Design `json:"design"`; Seed int64 `json:"seed"`; RunID string `json:"run_id,omitempty"` }
type RunManifest struct { SchemaVersion int `json:"schema_version"`; RunID string `json:"run_id"`; PointIndex int `json:"point_index"`; Seed int64 `json:"seed"`; EvaluatorVersion string `json:"evaluator_version"`; RepositoryRevision string `json:"repository_revision,omitempty"`; Inputs []project.ResolvedInput `json:"inputs"`; StartedAt time.Time `json:"started_at"`; CompletedAt time.Time `json:"completed_at"`; Status Status `json:"status"` }
type RunRecord struct { Manifest RunManifest `json:"manifest"`; Design project.Design `json:"design"`; Result Result `json:"result"`; Reused bool `json:"-"` }
type Summary struct { Runs []RunRecord `json:"runs"` }
type Evaluator func(context.Context, project.Design, int64) (Result, error)
```

- [ ] **Step 5: Implement sweep expansion**

In `sweep.go`, implement `Expand(bundle project.Bundle) ([]DesignPoint, error)`, `parameterValues`, and `apply`. Use the parameter declaration order as the outer-to-inner loop order. `range.count == 1` yields only `start`; otherwise use `start + i*(stop-start)/(count-1)`. Require exactly one of `values` or `range`. Reject duplicate paths and any product above `MaxPoints` before allocating. Supported paths are exactly:

```text
device.conductance_levels
device.g_min_s
device.g_max_s
array.rows
array.cols
array.read_voltage_v
circuit.adc_bits
circuit.dac_bits
circuit.tia_gain_ohm
```

Integer fields require `math.Trunc(value) == value`. Set `DesignPoint.Index` to output order and `Seed` to `bundle.Sweep.Seed + int64(index)`.

- [ ] **Step 6: Implement canonical identity**

Create `identity.go`:

```go
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
        SchemaVersion int `json:"schema_version"`
        Design project.Design `json:"design"`
        Seed int64 `json:"seed"`
        EvaluatorVersion string `json:"evaluator_version"`
        Inputs []project.ResolvedInput `json:"inputs"`
    }{1, point.Design, point.Seed, evaluatorVersion, ordered}
    data, err := json.Marshal(payload)
    if err != nil { return "", err }
    sum := sha256.Sum256(data)
    return hex.EncodeToString(sum[:]), nil
}
```

The identity payload contains structs and sorted slices only; do not add maps.

- [ ] **Step 7: Run GREEN and commit**

```bash
gofmt -w workbench/experiment/*.go
go test ./workbench/experiment -run 'TestExpand|TestID' -count=1
git add workbench/experiment
git commit -m "feat(workbench): expand deterministic design sweeps" -m "RED: sweep and identity contracts were missing. GREEN: Cartesian order, limits, field validation, and stable hashes pass."
```

---

### Task 3: FeCIM Device/Array/Circuit Evaluator

**Files:**

- Create: `workbench/fecim/evaluator.go`
- Test: `workbench/fecim/evaluator_test.go`

**Interfaces:**

- Consumes: `project.Design`, `experiment.Result`, `shared/system`, `shared/peripherals`.
- Produces: `EvaluatorVersion` and `Evaluate` for the runner.

- [ ] **Step 1: Write evaluator RED tests**

Cover one HZO golden point, metric ordering/names/units/evidence, context cancellation, unsupported material as a failed design point with nil Go error, and unsupported technology node as a failed design point. The golden test must assert values with a documented tolerance rather than updateable snapshots:

```go
func fixtureDesign() project.Design {
    return project.Design{
        SchemaVersion: 1,
        Device: project.Device{Material: "HZO", ConductanceLevels: 30, GMinS: 1e-6, GMaxS: 30e-6},
        Array: project.Array{Rows: 32, Cols: 32, ReadVoltageV: 0.2},
        Circuit: project.Circuit{ADCBits: 6, DACBits: 4, TIAGainOhm: 10_000, TechNode: "65nm"},
    }
}

func TestEvaluateHZOProducesTraceableMetrics(t *testing.T) {
    got, err := Evaluate(context.Background(), fixtureDesign(), 17)
    if err != nil { t.Fatal(err) }
    if got.Status != experiment.StatusSuccess { t.Fatalf("result=%+v", got) }
    names := []string{}
    for _, m := range got.Metrics {
        names = append(names, m.Name+":"+m.Unit)
        if m.Model == "" || len(m.Assumptions) == 0 || m.Evidence == "" { t.Fatalf("metric lacks provenance: %+v", m) }
    }
    want := []string{"area_um2:um2", "energy_pj:pJ", "latency_ns:ns", "tia_snr_db:dB"}
    if !reflect.DeepEqual(names, want) { t.Fatalf("got=%v want=%v", names, want) }
}
```

- [ ] **Step 2: Run RED**

```bash
go test ./workbench/fecim -count=1
```

Expected: FAIL because the evaluator is absent.

- [ ] **Step 3: Implement the thin evaluator**

Create `evaluator.go`. Use existing models exactly as follows:

```go
package fecim

import (
    "context"
    "fmt"
    "math"

    "fecim-lattice-tools/shared/peripherals"
    "fecim-lattice-tools/shared/system"
    "fecim-lattice-tools/workbench/experiment"
    "fecim-lattice-tools/workbench/project"
)

const EvaluatorVersion = "fecim-system-v1"

func Evaluate(ctx context.Context, d project.Design, seed int64) (experiment.Result, error) {
    if err := ctx.Err(); err != nil { return experiment.Result{}, err }
    if d.Device.Material != "HZO" {
        return failed("unsupported-material", fmt.Sprintf("material %q is not supported by %s", d.Device.Material, EvaluatorVersion)), nil
    }
    node, ok := technologyNode(d.Circuit.TechNode)
    if !ok { return failed("unsupported-tech-node", fmt.Sprintf("technology node %q is unsupported", d.Circuit.TechNode)), nil }
    if err := project.ValidateDesign(d); err != nil { return failed("model-domain", err.Error()), nil }
    _ = seed // retained in run identity; current evaluator is deterministic.

    dac := peripherals.DefaultDAC()
    dac.Bits = d.Circuit.DACBits
    adc := peripherals.DefaultADC()
    adc.Bits = d.Circuit.ADCBits
    tia := peripherals.DefaultTIA()
    tia.Gain = d.Circuit.TIAGainOhm

    latencyModel := system.NewLatencyModel(d.Array.Rows, d.Array.Cols, node)
    latencyNS := dac.SettleTime + latencyModel.CrossbarSettlingNS() + latencyModel.ADCLatencyNS(adc.Bits) + tia.SettlingTime()*1e9

    energy := system.EstimateMLPEnergyJ(system.MLPEnergyConfig{
        LayerSizes: []int{d.Array.Rows, d.Array.Cols},
        LevelsPerLayer: []int{d.Device.ConductanceLevels},
        EnergyPerDACJ: dac.EnergyPerConversion(),
        EnergyPerADCJ: adc.EnergyPerConversion(),
    })
    tiaEnergyJ := float64(d.Array.Cols) * tia.PowerConsumption() * tia.SettlingTime()
    energyPJ := (energy.TotalJ + tiaEnergyJ) * 1e12

    area := system.NewCrossbarAreaModel(d.Array.Rows, d.Array.Cols, node, system.CellFeFET).TotalAreaUM2(adc.Bits)
    meanConductance := (d.Device.GMinS + d.Device.GMaxS) / 2
    columnCurrent := float64(d.Array.Rows) * d.Array.ReadVoltageV * meanConductance
    snrDB := tia.SNR(columnCurrent)
    if math.IsNaN(snrDB) || math.IsInf(snrDB, 0) { return failed("model-domain", "TIA SNR is non-finite"), nil }

    assumptions := []string{"literature-calibrated pre-silicon model", "not measured silicon"}
    return experiment.Result{
        Status: experiment.StatusSuccess,
        Metrics: []experiment.Metric{
            {Name: "area_um2", Value: area, Unit: "um2", Model: "shared/system.CrossbarAreaModel", Assumptions: assumptions, Evidence: experiment.EvidenceDefault},
            {Name: "energy_pj", Value: energyPJ, Unit: "pJ", Model: "shared/system.EstimateMLPEnergyJ+shared/peripherals.TIA", Assumptions: assumptions, Evidence: experiment.EvidenceDerived},
            {Name: "latency_ns", Value: latencyNS, Unit: "ns", Model: "shared/system.LatencyModel+shared/peripherals", Assumptions: assumptions, Evidence: experiment.EvidenceDerived},
            {Name: "tia_snr_db", Value: snrDB, Unit: "dB", Model: "shared/peripherals.TIA", Assumptions: assumptions, Evidence: experiment.EvidenceDefault},
        },
        Warnings: []string{"simulation estimates only; calibrate before quantitative device claims"},
    }, nil
}

func failed(kind, message string) experiment.Result { return experiment.Result{Status: experiment.StatusFailed, Failure: &experiment.Failure{Kind: kind, Message: message}} }
func technologyNode(raw string) (system.TechnologyNode, bool) {
    switch raw {
    case "130nm": return system.Node130nm, true
    case "65nm": return system.Node65nm, true
    case "28nm": return system.Node28nm, true
    case "22nm": return system.Node22nm, true
    case "14nm": return system.Node14nm, true
    default: return "", false
    }
}
```

- [ ] **Step 4: Record golden values and run GREEN**

Run once with `go test -v` and record the exact expected area, energy, latency, and SNR values in the test with relative tolerance `1e-9`. Do not create an update flag.

```bash
gofmt -w workbench/fecim/*.go
go test ./workbench/fecim -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add workbench/fecim
git commit -m "feat(workbench): evaluate FeCIM design points" -m "RED: no headless device-array-circuit evaluator existed. GREEN: golden metrics, provenance, cancellation, and domain failures pass."
```

---

### Task 4: Immutable Atomic Store and Sequential Runner

**Files:**

- Create: `workbench/experiment/store.go`
- Create: `workbench/experiment/runner.go`
- Test: `workbench/experiment/runner_test.go`

**Interfaces:**

- Consumes: `Expand`, `ID`, `Evaluator`, and `project.Bundle`.
- Produces: `RunOptions` and `Run` with committed-run reuse.

- [ ] **Step 1: Write RED tests**

Add tests for:

- A successful run writes exactly `manifest.json`, `resolved-design.json`, and `result.json` under `runs/<64-hex-id>/`.
- A second run reuses records without calling the evaluator.
- A design-point failure is committed and later points continue.
- An evaluator Go error halts and does not commit that point.
- A pre-existing non-directory run path is reported as corruption.
- No `.tmp-` directory remains after success or evaluator failure.

Use an evaluator closure with an atomic call counter. Inject `Now` and `RepositoryRevision` so assertions are stable.

- [ ] **Step 2: Run RED**

```bash
go test ./workbench/experiment -run 'TestRun|TestStore' -count=1
```

Expected: FAIL because store and runner are absent.

- [ ] **Step 3: Implement atomic storage**

Create `store.go` with an unexported `store` type. Validate IDs with `^[0-9a-f]{64}$`. `load(id)` returns `(RunRecord, true, nil)` only after decoding all three files and confirming matching run IDs. `commit(record)` must:

1. `MkdirAll(<project>/runs, 0755)`.
2. Return the existing decoded record if the final directory exists.
3. `MkdirTemp(runs, ".tmp-"+id+"-")`.
4. Write indented JSON files with mode `0644`.
5. Rename the complete temporary directory to the final ID directory.
6. Remove the temporary directory on every error path.
7. If rename loses a race to an existing final directory, decode and return the existing record.

Use `json.Decoder.DisallowUnknownFields()` while loading committed runs.

- [ ] **Step 4: Implement the sequential runner first**

Create `runner.go`:

```go
package experiment

import (
    "context"
    "errors"
    "fmt"
    "time"
    "fecim-lattice-tools/workbench/project"
)

type RunOptions struct {
    Evaluator Evaluator
    EvaluatorVersion string
    Workers int
    RepositoryRevision string
    Now func() time.Time
}

func Run(ctx context.Context, bundle project.Bundle, opts RunOptions) (Summary, error) {
    if opts.Evaluator == nil || opts.EvaluatorVersion == "" { return Summary{}, errors.New("evaluator and evaluator version are required") }
    if opts.Now == nil { opts.Now = time.Now }
    points, err := Expand(bundle)
    if err != nil { return Summary{}, err }
    s := store{root: bundle.Root}
    summary := Summary{Runs: make([]RunRecord, 0, len(points))}
    for _, point := range points {
        if err := ctx.Err(); err != nil { return summary, err }
        id, err := ID(point, opts.EvaluatorVersion, bundle.Inputs)
        if err != nil { return summary, err }
        point.RunID = id
        if existing, ok, err := s.load(id); err != nil { return summary, err } else if ok {
            existing.Reused = true
            summary.Runs = append(summary.Runs, existing)
            continue
        }
        started := opts.Now().UTC()
        result, err := opts.Evaluator(ctx, point.Design, point.Seed)
        if err != nil { return summary, fmt.Errorf("run %s: %w", id, err) }
        if result.Status != StatusSuccess && result.Status != StatusFailed { return summary, fmt.Errorf("run %s returned invalid status %q", id, result.Status) }
        record := RunRecord{
            Manifest: RunManifest{SchemaVersion: 1, RunID: id, PointIndex: point.Index, Seed: point.Seed, EvaluatorVersion: opts.EvaluatorVersion, RepositoryRevision: opts.RepositoryRevision, Inputs: bundle.Inputs, StartedAt: started, CompletedAt: opts.Now().UTC(), Status: result.Status},
            Design: point.Design,
            Result: result,
        }
        committed, err := s.commit(record)
        if err != nil { return summary, err }
        summary.Runs = append(summary.Runs, committed)
    }
    return summary, nil
}
```

- [ ] **Step 5: Run GREEN and commit**

```bash
gofmt -w workbench/experiment/*.go
go test ./workbench/experiment -run 'TestRun|TestStore' -count=1
git add workbench/experiment/store.go workbench/experiment/runner.go workbench/experiment/runner_test.go
git commit -m "feat(workbench): persist immutable experiment runs" -m "RED: run reuse and atomic persistence were absent. GREEN: success, failed-point continuation, corruption, reuse, and cleanup tests pass."
```

---

### Task 5: Bounded Concurrency, Cancellation, and Resume

**Files:**

- Modify: `workbench/experiment/runner.go`
- Modify: `workbench/experiment/runner_test.go`

**Interfaces:**

- Consumes: Task 4 runner semantics.
- Produces: worker-count-independent ordering and resumable cancellation.

- [ ] **Step 1: Add failing concurrency tests**

Add tests proving:

- `Workers: 1` and `Workers: 4` produce the same ordered run IDs and result payloads.
- Maximum simultaneous evaluator calls never exceeds `Workers`.
- Cancelling after two commits returns `context.Canceled`; a subsequent call reuses those records and completes the rest.
- A systemic evaluator error cancels scheduling and returns the first error while preserving already committed records.

Use channels rather than sleeps to coordinate cancellation.

- [ ] **Step 2: Run RED**

```bash
go test ./workbench/experiment -run 'TestRunBounded|TestRunCancellation|TestRunWorkerDeterminism' -count=1
```

Expected: FAIL because Task 4 runs sequentially and ignores `Workers`.

- [ ] **Step 3: Replace the loop with a bounded worker pool**

Implement a private `runJob{index int; point DesignPoint; id string}` and `runOutcome{index int; record RunRecord; err error}`. Required behavior:

- Default `Workers <= 0` to `1`.
- Precompute all run IDs in deterministic point order.
- Load reusable records before scheduling.
- Workers call the evaluator and atomic store only for missing records.
- The coordinator stores outcomes by original index and returns `Summary.Runs` in that order.
- The first systemic error cancels an internal child context.
- Never overwrite the caller's cancellation error with a later worker error.
- Close channels from the producer/coordinator only; workers never close shared channels.

Use `sync.WaitGroup`, `context.WithCancel`, and fixed-size channels from the standard library. Do not add `errgroup` because this plan needs indexed partial outcomes.

- [ ] **Step 4: Run GREEN, race test, and commit**

```bash
gofmt -w workbench/experiment/*.go
go test ./workbench/experiment -run 'TestRun' -count=1
go test -race ./workbench/experiment -run 'TestRun' -count=1
git add workbench/experiment/runner.go workbench/experiment/runner_test.go
git commit -m "feat(workbench): run bounded resumable sweeps" -m "RED: worker limits, cancellation, and deterministic ordering failed. GREEN: focused and race tests pass across worker counts."
```

---

### Task 6: Feasibility and Generic Pareto Analysis

**Files:**

- Create: `workbench/experiment/analyze.go`
- Test: `workbench/experiment/analyze_test.go`

**Interfaces:**

- Consumes: `project.Objective`, `project.Constraint`, and ordered `RunRecord` values.
- Produces: `Analysis` and `AnalyzedRun` for reports and later Fyne views.

- [ ] **Step 1: Write RED classification tests**

Cover:

- Failed runs remain failed and not feasible.
- Missing objective/constraint metrics make a successful run unusable with a reason.
- Unit mismatches make a run unusable.
- Constraint operators `<`, `<=`, `>`, `>=`, and `==` work.
- Infeasible successful runs remain inspectable.
- Mixed minimize/maximize objectives produce the expected Pareto front.
- Equal points are both Pareto-optimal; neither strictly dominates the other.

- [ ] **Step 2: Run RED**

```bash
go test ./workbench/experiment -run 'TestAnalyze|TestPareto' -count=1
```

Expected: FAIL because analysis types do not exist.

- [ ] **Step 3: Implement analysis**

Create these types and functions:

```go
type ConstraintOutcome struct { Constraint project.Constraint `json:"constraint"`; Passed bool `json:"passed"`; Actual float64 `json:"actual"` }
type AnalyzedRun struct { Run RunRecord `json:"run"`; Feasible bool `json:"feasible"`; Pareto bool `json:"pareto"`; UnusableReason string `json:"unusable_reason,omitempty"`; Constraints []ConstraintOutcome `json:"constraints"` }
type Analysis struct { Runs []AnalyzedRun `json:"runs"`; Counts map[string]int `json:"counts"` }
func Analyze(bundle project.Bundle, runs []RunRecord) Analysis
```

Build a metric lookup by name per run and reject duplicates. A constraint's unit must exactly match the metric unit. Objective direction comes from `project.Project.Objectives`. `dominates(a,b)` requires `a` no worse in every objective and strictly better in at least one. Run Pareto comparison only across feasible runs. Populate counts for `success`, `failed`, `feasible`, `infeasible`, `unusable`, and `pareto`.

- [ ] **Step 4: Run GREEN and commit**

```bash
gofmt -w workbench/experiment/*.go
go test ./workbench/experiment -run 'TestAnalyze|TestPareto' -count=1
git add workbench/experiment/analyze.go workbench/experiment/analyze_test.go
git commit -m "feat(workbench): classify feasible Pareto designs" -m "RED: generic constraints and objective directions were absent. GREEN: feasibility, units, missing metrics, and Pareto tests pass."
```

---

### Task 7: Deterministic JSON, CSV, and Markdown Reports

**Files:**

- Create: `workbench/report/report.go`
- Test: `workbench/report/report_test.go`

**Interfaces:**

- Consumes: `project.Bundle` and `experiment.Analysis`.
- Produces: `WriteJSON`, `WriteCSV`, `WriteMarkdown`, and `Generate(root, bundle, analysis)`.

- [ ] **Step 1: Write RED report tests**

Use two analyzed runs and assert:

- JSON excludes `StartedAt`, `CompletedAt`, and `Reused` but includes run ID, design, result, provenance, feasibility, Pareto, and constraint outcomes.
- CSV has stable columns: `run_id,status,feasible,pareto,unusable_reason,<objective metrics>,<constraint metrics>,warnings`.
- Markdown includes the hypothesis, trust-boundary banner, counts, objective directions, selected Pareto rows, warnings, and evidence levels.
- Two calls with the same analysis are byte-identical.
- `Generate` writes all three files via temporary files and rename.

- [ ] **Step 2: Run RED**

```bash
go test ./workbench/report -count=1
```

Expected: FAIL because the package is absent.

- [ ] **Step 3: Implement deterministic reporting**

Create `report.go` with:

```go
func WriteJSON(w io.Writer, a experiment.Analysis) error
func WriteCSV(w io.Writer, a experiment.Analysis) error
func WriteMarkdown(w io.Writer, b project.Bundle, a experiment.Analysis) error
func Generate(root string, b project.Bundle, a experiment.Analysis) error
```

Before encoding, project each run into a report DTO that omits timestamps and `Reused`. Preserve analysis order. Sort metric columns by the order in `project.yaml`: objectives first, constraints second, deduplicated. Format floats with `strconv.FormatFloat(value, 'g', -1, 64)`. Join warnings with `" | "`. Markdown must begin:

```markdown
> Simulation-only, literature-calibrated pre-silicon estimates. Not measured silicon or foundry sign-off.
```

`Generate` writes `reports/results.json`, `reports/results.csv`, and `reports/report.md`; write each sibling temporary file, close it, then rename it over the destination. Remove temporary files on errors.

- [ ] **Step 4: Run GREEN and commit**

```bash
gofmt -w workbench/report/*.go
go test ./workbench/report -count=1
git add workbench/report
git commit -m "feat(workbench): generate deterministic trade-off reports" -m "RED: no stable report contract existed. GREEN: JSON, CSV, Markdown, provenance, and atomic generation tests pass."
```

---

### Task 8: CLI Commands and Checked-In Example Project

**Files:**

- Create: `cmd/fecim-lattice-tools/workbench_subcommand.go`
- Create: `cmd/fecim-lattice-tools/workbench_subcommand_test.go`
- Create: `examples/research-design-workbench/project.yaml`
- Create: `examples/research-design-workbench/design.yaml`
- Create: `examples/research-design-workbench/sweep.yaml`
- Modify: `cmd/fecim-lattice-tools/subcommands.go`

**Interfaces:**

- Consumes: all headless-core packages.
- Produces: `project init`, `project validate`, `experiment run`, and `report generate` command surfaces.

- [ ] **Step 1: Write command-routing RED tests**

Test `dispatchSubcommandWithWriters` with injected stdout/stderr for:

- `project init <tempdir>` creates only the three YAML files and refuses a non-empty destination.
- `project validate <example-copy> --citation-dir <repo>/citations/papers` prints `valid: <id>`.
- `experiment run <example-copy> --workers 2` creates expected run count and reports reused count on the second call.
- `report generate <example-copy>` writes all report files.
- Unknown nested commands and malformed flags write no scientific result to stdout.
- All errors return through `runMain` without `os.Exit` inside helpers.

- [ ] **Step 2: Run RED**

```bash
go test ./cmd/fecim-lattice-tools -run 'TestWorkbench|TestProjectCommand|TestExperimentCommand|TestReportCommand' -count=1
```

Expected: FAIL because dispatch has no workbench commands.

- [ ] **Step 3: Add the example project**

Use the exact valid schemas from Task 1. The sweep must vary all three layers while staying below 32 points:

```yaml
parameters:
  - path: device.conductance_levels
    values: [16, 30]
  - path: array.rows
    values: [32, 64]
  - path: circuit.adc_bits
    values: [4, 6]
```

Set objectives to minimize `energy_pj` and `latency_ns`; constrain `area_um2`; cite `park2015_advmat_hzo`. Do not check in generated `runs/`, `reports/`, or `exports/`.

- [ ] **Step 4: Implement command handlers**

Create `workbench_subcommand.go` with writer-injected helpers:

```go
func runProjectSubcommand(args []string, stdout, stderr io.Writer) error
func runExperimentSubcommand(args []string, stdout, stderr io.Writer) error
func runReportSubcommand(args []string, stdout, stderr io.Writer) error
```

Use one `flag.FlagSet` per nested command with `ContinueOnError`. Exact syntax:

```text
fecim-lattice-tools project init DIRECTORY
fecim-lattice-tools project validate DIRECTORY [-citation-dir PATH]
fecim-lattice-tools experiment run DIRECTORY [-workers N] [-citation-dir PATH]
fecim-lattice-tools report generate DIRECTORY [-citation-dir PATH]
```

`project init` writes embedded constant YAML templates with `os.OpenFile(..., O_CREATE|O_EXCL, 0644)` and removes files created in the current invocation if a later template write fails. It refuses an existing non-empty directory.

`experiment run` calls:

```go
bundle, err := project.Load(dir, project.LoadOptions{CitationDir: citationDir})
summary, err := experiment.Run(context.Background(), bundle, experiment.RunOptions{
    Evaluator: fecim.Evaluate,
    EvaluatorVersion: fecim.EvaluatorVersion,
    Workers: workers,
    RepositoryRevision: buildRevision(),
})
analysis := experiment.Analyze(bundle, summary.Runs)
```

Then call `report.Generate`. Print one concise line containing total, reused, failed, feasible, and Pareto counts. `buildRevision()` should use `debug.ReadBuildInfo()` and return the VCS revision setting when present; an empty revision is allowed and must not alter run identity.

`report generate` loads every committed run directory through a new exported `experiment.LoadRuns(root) ([]RunRecord, error)` that wraps the strict store loader and sorts by `Manifest.PointIndex`, then analyzes and generates reports without invoking the evaluator.

- [ ] **Step 5: Wire root dispatch and help**

In `subcommands.go`, add cases for `project`, `experiment`, and `report`. Add one-line help entries and one example command. Keep existing module and research commands unchanged.

- [ ] **Step 6: Run GREEN and end-to-end reproduction**

```bash
gofmt -w cmd/fecim-lattice-tools/workbench_subcommand*.go cmd/fecim-lattice-tools/subcommands.go
go test ./cmd/fecim-lattice-tools -run 'TestWorkbench|TestProjectCommand|TestExperimentCommand|TestReportCommand' -count=1
rm -rf /tmp/fecim-workbench-example
cp -R examples/research-design-workbench /tmp/fecim-workbench-example
go run ./cmd/fecim-lattice-tools experiment run /tmp/fecim-workbench-example -workers 2 -citation-dir citations/papers
go run ./cmd/fecim-lattice-tools report generate /tmp/fecim-workbench-example -citation-dir citations/papers
```

Expected: 8 runs, no command failure, and three report files. Do not commit `/tmp` outputs.

- [ ] **Step 7: Commit**

```bash
git add cmd/fecim-lattice-tools/subcommands.go cmd/fecim-lattice-tools/workbench_subcommand.go cmd/fecim-lattice-tools/workbench_subcommand_test.go examples/research-design-workbench
git commit -m "feat(cli): run reproducible FeCIM design studies" -m "RED: project, experiment, and report commands were unknown. GREEN: init, validate, run, reuse, and report E2E tests pass."
```

---

### Task 9: Documentation and Final Gates

**Files:**

- Create: `docs/guides/research-design-workbench.md`
- Modify: `README.md`
- Modify: `TODO.md`

**Interfaces:**

- Consumes: completed CLI behavior and example project.
- Produces: accurate user documentation and implementation receipts.

- [ ] **Step 1: Add a failing documentation contract**

Create or extend `cmd/fecim-lattice-tools/command_surface_test.go` to require the README and guide to contain:

```text
research and design workbench
project validate
experiment run
report generate
literature-calibrated pre-silicon
```

Run:

```bash
go test ./cmd/fecim-lattice-tools -run TestReleasedCommandSurface -count=1
```

Expected: FAIL because the guide and README do not yet document the commands.

- [ ] **Step 2: Write the user guide**

Document:

- Product purpose and simulation-only boundary
- Project directory schema
- Exact example-copy/run/report commands
- Objective, constraint, unit, evidence, and run-ID semantics
- Immutable run behavior and how resume works
- Meaning of failed, infeasible, unusable, and Pareto states
- Why output is not foundry sign-off or experimental validation
- Deferred Fyne workflow and selected-run EDA export

- [ ] **Step 3: Update README and TODO**

Make research/design the lead purpose without removing educational value. Link the checked-in example and guide. In `TODO.md`, mark the headless core/CLI slice complete and add two explicit pending items: Fyne workflow parity and selected-run export integration. Include RED/GREEN/final verification receipts.

- [ ] **Step 4: Run focused and full validation**

```bash
gofmt -w workbench cmd/fecim-lattice-tools
go test ./workbench/... -count=1
go test -race ./workbench/... -count=1
go test ./cmd/fecim-lattice-tools -count=1
git diff --check
go test ./...
```

Expected: all commands PASS; no formatting or whitespace errors.

- [ ] **Step 5: Verify deterministic reproduction twice**

```bash
rm -rf /tmp/fecim-workbench-final
cp -R examples/research-design-workbench /tmp/fecim-workbench-final
go run ./cmd/fecim-lattice-tools experiment run /tmp/fecim-workbench-final -workers 1 -citation-dir citations/papers
sha256sum /tmp/fecim-workbench-final/reports/results.json /tmp/fecim-workbench-final/reports/results.csv /tmp/fecim-workbench-final/reports/report.md > /tmp/fecim-report-hashes-1
go run ./cmd/fecim-lattice-tools experiment run /tmp/fecim-workbench-final -workers 4 -citation-dir citations/papers
sha256sum /tmp/fecim-workbench-final/reports/results.json /tmp/fecim-workbench-final/reports/results.csv /tmp/fecim-workbench-final/reports/report.md > /tmp/fecim-report-hashes-2
diff -u /tmp/fecim-report-hashes-1 /tmp/fecim-report-hashes-2
```

Expected: second run reports all runs reused and `diff` has no output.

- [ ] **Step 6: Commit documentation**

```bash
git add docs/guides/research-design-workbench.md README.md TODO.md cmd/fecim-lattice-tools/command_surface_test.go
git commit -m "docs: introduce the research design workflow" -m "TDD: documentation contract failed before README/guide updates and passes afterward. Full Go and workbench race suites pass."
```

---

## Completion Gate

The plan is complete only when:

- The example produces eight immutable run directories.
- Running with worker counts 1 and 4 yields identical ordered scientific results.
- A repeated run reuses all eight records.
- JSON, CSV, and Markdown report hashes remain unchanged after reuse.
- Failure, infeasibility, unusable results, and Pareto membership are distinguishable.
- Every metric carries units, model source, assumptions, evidence, and run ID.
- `go test -race ./workbench/...` passes.
- `go test ./...` passes.
- No generated runs or reports are committed from `/tmp` or the example directory.

## Follow-Up Planning Triggers

After this plan is green:

1. Write a Fyne workflow plan using `experiment.Run`, `experiment.Analyze`, and persisted report DTOs; it must include UI-thread safety, cancellation, and CLI/Fyne run-ID parity.
2. Write a selected-run export plan that maps one immutable successful run into existing Module 6 compiler/export types without rerunning models.
