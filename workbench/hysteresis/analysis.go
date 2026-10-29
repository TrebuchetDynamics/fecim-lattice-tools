package hysteresis

import (
	"context"
	"errors"
	"math"
	"sync"

	"fecim-lattice-tools/shared/physics"
)

type AnalysisKind string

const (
	AnalysisPUND      AnalysisKind = "pund"
	AnalysisFORC      AnalysisKind = "forc"
	MaxFORCReversals               = 64
	MaxFORCResolution              = 256
)

var ErrAnalysisIncomplete = errors.New("analysis is not complete")

type FORCParameters struct {
	EmaxVm                float64
	Reversals, Resolution int
}

func DefaultFORCParameters() FORCParameters { return FORCParameters{1, 21, 21} }

// Input is an owned session snapshot. Only its revisions and validated Config
// are consumed; material parameters are resolved from the Go catalogue, never
// trusted from caller-edited snapshot fields. Neither job nor session is linked.
type AnalysisRequest struct {
	Input      Snapshot
	Kind       AnalysisKind
	FORC       FORCParameters
	References []ReferenceInput
}
type AnalysisContext struct {
	Revision, ConfigRevision uint64
	Config                   Config
}
type AppliedDiagnosticParameters struct {
	EcVm, PsCm2, DeltaVm, AmplitudeVm, PulseWidthS, SampleIntervalS, AreaM2 float64
	EmaxVm                                                                  float64
	Reversals, Resolution                                                   int
}
type DiagnosticIdentity struct {
	Profile    string
	Version    int
	Model      Model
	MaterialID string
	Applied    AppliedDiagnosticParameters
}
type PUNDDiagnostic struct {
	Charges physics.PUNDResult
	Traces  [4][]physics.PulseSample
}

// FORCMatrix preserves GUI nearest-index resampling: rows are reversal fields,
// columns are selected applied-field grid positions, not a new physical sweep.
type FORCMatrix struct {
	EaAxis  []float64   `json:"ea_axis_vm"`
	EbAxis  []float64   `json:"eb_axis_vm"`
	Density [][]float64 `json:"density"`
	Rows    int         `json:"rows"`
	Cols    int         `json:"cols"`
}
type FORCDiagnostic struct {
	Sweep  physics.FORCResult
	Matrix FORCMatrix
}
type AnalysisResult struct {
	ScientificValues       map[string]Quantity
	AppliedParameters      map[string]any
	Revision               uint64
	Kind                   AnalysisKind
	Context                AnalysisContext
	Diagnostic             DiagnosticIdentity
	Evidence               string
	ExperimentalValidation bool
	Notices                []string
	ReferenceContext       ReferenceContext
	References             []Reference
	PUND                   *PUNDDiagnostic
	FORC                   *FORCDiagnostic
}

// Steps counts bounded protocol work units, not physical time: the fixed PUND
// profile (105 updates), one FORC curve (<=65 updates), density (<=64² cells), or
// one resampled matrix row (<=256 cells). Context is checked inside shared loops.
// These are safety bounds, not browser responsiveness measurements.
type AnalysisBudget struct {
	Steps   int
	Context context.Context
}
type AnalysisProgress struct {
	Revision         uint64
	State            string
	Completed, Total int
}
type AnalysisJob struct {
	mu               sync.Mutex
	result           AnalysisResult
	state            string
	completed, total int
	stack            *physics.PreisachStack
}

func NewAnalysis(req AnalysisRequest) (*AnalysisJob, error) {
	m, err := validateConfig(req.Input.Config, Catalogue())
	if err != nil {
		return nil, err
	}
	if req.Input.Revision == 0 || req.Input.ConfigRevision == 0 {
		return nil, errors.New("analysis requires session input revisions")
	}
	if req.Kind != AnalysisPUND && req.Kind != AnalysisFORC {
		return nil, errors.New("unsupported analysis kind")
	}
	if req.Kind != AnalysisFORC && req.FORC != (FORCParameters{}) {
		return nil, errors.New("unexpected FORC parameters")
	}
	if len(req.References) > 32 {
		return nil, errors.New("at most 32 reference branches")
	}
	size := 0
	for _, ref := range req.References {
		if len(ref.Data) > MaxReferenceBytes || len(ref.Source) > 1024 || len(ref.Material) > 256 {
			return nil, errors.New("reference input too large")
		}
		size += len(ref.Data) + len(ref.Source) + len(ref.Material) + len(ref.Format)
		if size > MaxReferenceBytes {
			return nil, errors.New("reference envelope exceeds 1 MiB")
		}
	}
	refs := make([]Reference, 0, len(req.References))
	remaining := MaxReferencePoints
	for _, in := range req.References {
		ref, err := parseReference(in, remaining)
		if err != nil {
			return nil, err
		}
		refs = append(refs, ref)
		remaining -= len(ref.Points)
	}
	j := &AnalysisJob{state: "running", total: 1, result: AnalysisResult{Revision: 1, Kind: req.Kind, Context: AnalysisContext{req.Input.Revision, req.Input.ConfigRevision, req.Input.Config}, Evidence: "simulation_only", References: refs, ReferenceContext: materialReferenceContext(m), Notices: []string{
		"Independent Preisach diagnostic; selected loop model, temperature and stress are context only, not applied",
		"Simulation only; no experimental validation, fitting or stored-run verification",
	}}}
	if req.Kind == AnalysisPUND {
		a := AppliedDiagnosticParameters{EcVm: 3e7, PsCm2: 0.25, DeltaVm: 3e7 * 0.3, AmplitudeVm: 5 * 3e7, PulseWidthS: 100e-9, SampleIntervalS: 5e-9, AreaM2: 1e-12}
		j.result.Diagnostic = DiagnosticIdentity{"legacy-pund-demo", 1, Preisach, "hzo-demo", a}
		j.stack = physics.NewPreisachStack(a.EcVm, &physics.TanhEverett{Ec: a.EcVm, Ps: a.PsCm2, Delta: a.DeltaVm})
	} else {
		f := req.FORC
		if f == (FORCParameters{}) {
			f = DefaultFORCParameters()
		}
		if !finite(f.EmaxVm) || f.EmaxVm <= 0 || f.Reversals < 3 || f.Reversals > MaxFORCReversals || f.Resolution < 3 || f.Resolution > MaxFORCResolution {
			return nil, errors.New("FORC requires finite positive Emax, 3–64 reversals and resolution 3–256")
		}
		dE := 2 * f.EmaxVm / float64(f.Reversals-1)
		denom := 4 * dE * dE
		ps := m.Parameters.Ps
		if !finite(2*f.EmaxVm) || !finite(denom) || denom <= 0 || !finite(4*ps/denom) || !finite(ps) || ps <= 0 {
			return nil, errors.New("FORC field/density scale is not representable")
		}
		grid := make([]float64, f.Reversals)
		for i := range grid {
			grid[i] = -f.EmaxVm + float64(i)*dE
			if i > 0 && grid[i] <= grid[i-1] {
				return nil, errors.New("FORC grid collapsed")
			}
		}
		j.result.Diagnostic = DiagnosticIdentity{"gui-forc", 1, Preisach, m.ID, AppliedDiagnosticParameters{PsCm2: ps, EmaxVm: f.EmaxVm, Reversals: f.Reversals, Resolution: f.Resolution}}
		j.result.FORC = &FORCDiagnostic{Sweep: physics.FORCResult{Emax_Vm: f.EmaxVm, ReversalFields_Vm: grid}, Matrix: FORCMatrix{Rows: f.Reversals, Cols: f.Resolution}}
		j.stack = physics.NewPreisachStack(f.EmaxVm, physics.GUIFORCEverett{SaturationE: f.EmaxVm, Ps: ps})
		j.total = 2*f.Reversals + 1
	}
	return j, nil
}
func (j *AnalysisJob) progress() AnalysisProgress {
	return AnalysisProgress{j.result.Revision, j.state, j.completed, j.total}
}

func (j *AnalysisJob) Advance(b AnalysisBudget) (AnalysisProgress, error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if b.Steps < 1 || b.Steps > MaxSteps {
		return j.progress(), errors.New("analysis budget requires 1–256 work units")
	}
	ctx := b.Context
	if ctx == nil {
		ctx = context.Background()
	}
	for i := 0; i < b.Steps && j.state == "running"; i++ {
		err := ctx.Err()
		if err == nil {
			err = j.step(ctx)
		}
		if err != nil {
			j.state = "failed"
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				j.state = "cancelled"
			}
			j.result.Revision++
			return j.progress(), err
		}
		j.completed++
		j.result.Revision++
		if j.completed == j.total {
			j.state = "complete"
		}
	}
	return j.progress(), nil
}
func (j *AnalysisJob) step(ctx context.Context) error {
	a := j.result.Diagnostic.Applied
	if j.result.Kind == AnalysisPUND {
		charges, traces, err := physics.RunPUNDSimulationContext(ctx, j.stack, a.AmplitudeVm, a.PulseWidthS, a.SampleIntervalS, a.AreaM2)
		if err != nil {
			return err
		}
		j.result.PUND = &PUNDDiagnostic{charges, traces}
		return nil
	}
	f := j.result.FORC
	if j.completed < a.Reversals {
		curve, err := physics.RunFORCCurveContext(ctx, j.stack, f.Sweep.ReversalFields_Vm, j.completed)
		if err != nil {
			return err
		}
		f.Sweep.Curves = append(f.Sweep.Curves, curve)
		return nil
	}
	if j.completed == a.Reversals {
		density, pairs, err := physics.ComputeFORCDensityContext(ctx, f.Sweep)
		if err != nil {
			return err
		}
		f.Sweep.PreisachDensity = density
		f.Sweep.ReversalPairs = pairs
		f.Matrix.EaAxis = make([]float64, a.Resolution)
		for c := range f.Matrix.EaAxis {
			src := int(math.Round(float64(c) * float64(a.Reversals-1) / float64(a.Resolution-1)))
			f.Matrix.EaAxis[c] = f.Sweep.ReversalFields_Vm[src]
		}
		f.Matrix.EbAxis = append([]float64(nil), f.Sweep.ReversalFields_Vm...)
		return nil
	}
	row := j.completed - a.Reversals - 1
	values := make([]float64, a.Resolution)
	for c := range values {
		src := int(math.Round(float64(c) * float64(a.Reversals-1) / float64(a.Resolution-1)))
		values[c] = f.Sweep.PreisachDensity[row][src]
	}
	f.Matrix.Density = append(f.Matrix.Density, values)
	return nil
}

// Result only exposes complete diagnostics. Returned arrays are owned copies.
func (j *AnalysisJob) Result() (AnalysisResult, error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.state != "complete" {
		return AnalysisResult{}, ErrAnalysisIncomplete
	}
	r := j.result
	r.ScientificValues = configValues(r.Context.Config, "Context / Config / ")
	a := r.Diagnostic.Applied
	if r.Kind == AnalysisPUND {
		r.AppliedParameters = map[string]any{"EcVm": a.EcVm, "PsCm2": a.PsCm2, "DeltaVm": a.DeltaVm, "AmplitudeVm": a.AmplitudeVm, "PulseWidthS": a.PulseWidthS, "SampleIntervalS": a.SampleIntervalS, "AreaM2": a.AreaM2}
		for key, value := range map[string]Quantity{
			"EcVm": field(a.EcVm), "PsCm2": polarization(a.PsCm2), "DeltaVm": field(a.DeltaVm), "AmplitudeVm": field(a.AmplitudeVm),
			"PulseWidthS": scalar(a.PulseWidthS, "s"), "SampleIntervalS": scalar(a.SampleIntervalS, "s"), "AreaM2": scalar(a.AreaM2, "m²"),
		} {
			r.ScientificValues["Diagnostic / Applied / "+key] = value
		}
		c := r.PUND.Charges
		for key, value := range map[string]float64{"QP_C": c.QP_C, "QU_C": c.QU_C, "QN_C": c.QN_C, "QD_C": c.QD_C, "SwitchingPositive_C": c.SwitchingPositive_C, "SwitchingNegative_C": c.SwitchingNegative_C} {
			r.ScientificValues["PUNDCharges / "+key] = scalar(value, "C")
		}
	} else {
		r.AppliedParameters = map[string]any{"PsCm2": a.PsCm2, "EmaxVm": a.EmaxVm, "Reversals": a.Reversals, "Resolution": a.Resolution}
		r.ScientificValues["Diagnostic / Applied / PsCm2"] = polarization(a.PsCm2)
		r.ScientificValues["Diagnostic / Applied / EmaxVm"] = field(a.EmaxVm)
	}
	r.Notices = append([]string(nil), r.Notices...)
	r.ReferenceContext.Citations = append([]Citation(nil), r.ReferenceContext.Citations...)
	r.References = append([]Reference(nil), r.References...)
	for i := range r.References {
		r.References[i].Points = append([]ReferencePoint(nil), r.References[i].Points...)
	}
	if r.PUND != nil {
		p := *r.PUND
		r.PUND = &p
		for i := range p.Traces {
			p.Traces[i] = append([]physics.PulseSample(nil), p.Traces[i]...)
		}
	}
	if r.FORC != nil {
		f := *r.FORC
		r.FORC = &f
		f.Sweep.ReversalFields_Vm = append([]float64(nil), f.Sweep.ReversalFields_Vm...)
		f.Sweep.ReversalPairs = append([][2]float64(nil), f.Sweep.ReversalPairs...)
		f.Sweep.Curves = append([]physics.FORCCurve(nil), f.Sweep.Curves...)
		for i := range f.Sweep.Curves {
			c := &f.Sweep.Curves[i]
			c.AppliedField_Vm = append([]float64(nil), c.AppliedField_Vm...)
			c.Polarization_Cm2 = append([]float64(nil), c.Polarization_Cm2...)
		}
		f.Sweep.PreisachDensity = cloneMatrix(f.Sweep.PreisachDensity)
		f.Matrix.EaAxis = append([]float64(nil), f.Matrix.EaAxis...)
		f.Matrix.EbAxis = append([]float64(nil), f.Matrix.EbAxis...)
		f.Matrix.Density = cloneMatrix(f.Matrix.Density)
	}
	return r, nil
}
func cloneMatrix(in [][]float64) [][]float64 {
	out := make([][]float64, len(in))
	for i := range in {
		out[i] = append([]float64(nil), in[i]...)
	}
	return out
}

type AnalysisTable string

const (
	PUNDTraces  AnalysisTable = "pund-traces"
	FORCSweep   AnalysisTable = "forc-sweep"
	FORCDensity AnalysisTable = "forc-density"
)

type AnalysisPageQuery struct {
	Revision uint64
	Table    AnalysisTable
	Offset   int
}
type AnalysisRow struct {
	TimeValue, CurrentValue, DensityValue Quantity
	Index                                 int
	Pulse, Branch                         string
	TimeS, CurrentA, Density              float64
	E, P, Ea, Eb, Reversal                Quantity
}
type AnalysisPage struct {
	Context       AnalysisContext
	Diagnostic    DiagnosticIdentity
	Evidence      string
	Revision      uint64
	State         string
	Table         AnalysisTable
	Offset, Total int
	Rows          []AnalysisRow
}

func (j *AnalysisJob) Page(q AnalysisPageQuery) (AnalysisPage, error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if q.Revision != j.result.Revision {
		return AnalysisPage{}, ErrStaleRevision
	}
	total, err := analysisRowCount(j.result, q.Table)
	if err != nil {
		return AnalysisPage{}, err
	}
	if q.Offset < 0 || q.Offset > total {
		return AnalysisPage{}, errors.New("analysis page offset outside retained rows")
	}
	p := AnalysisPage{Context: j.result.Context, Diagnostic: j.result.Diagnostic, Evidence: j.result.Evidence, Revision: j.result.Revision, State: j.state, Table: q.Table, Offset: q.Offset, Total: total, Rows: make([]AnalysisRow, 0, min(PageSize, total-q.Offset))}
	for i := q.Offset; i < min(q.Offset+PageSize, total); i++ {
		p.Rows = append(p.Rows, analysisRow(j.result, q.Table, i))
	}
	return p, nil
}
func analysisRowCount(r AnalysisResult, table AnalysisTable) (int, error) {
	switch table {
	case PUNDTraces:
		if r.Kind != AnalysisPUND {
			break
		}
		n := 0
		if r.PUND != nil {
			for _, trace := range r.PUND.Traces {
				n += len(trace)
			}
		}
		return n, nil
	case FORCSweep:
		if r.Kind != AnalysisFORC {
			break
		}
		n := 0
		for _, c := range r.FORC.Sweep.Curves {
			n += len(c.AppliedField_Vm)
		}
		return n, nil
	case FORCDensity:
		if r.Kind != AnalysisFORC {
			break
		}
		return len(r.FORC.Matrix.Density) * r.FORC.Matrix.Cols, nil
	}
	return 0, errors.New("unsupported analysis table for diagnostic")
}
func analysisRow(r AnalysisResult, table AnalysisTable, index int) AnalysisRow {
	row := AnalysisRow{Index: index}
	switch table {
	case PUNDTraces:
		for i, trace := range r.PUND.Traces {
			if index < len(trace) {
				s := trace[index]
				row.Pulse = []string{"P", "U", "N", "D"}[i]
				row.TimeS = s.TimeS
				row.CurrentA = s.CurrentA
				row.TimeValue = scalar(s.TimeS, "s")
				row.CurrentValue = scalar(s.CurrentA, "A")
				break
			}
			index -= len(trace)
		}
	case FORCSweep:
		for _, c := range r.FORC.Sweep.Curves {
			if index < len(c.AppliedField_Vm) {
				row.Branch = "descending"
				row.Reversal = field(c.ReversalField_Vm)
				row.E = field(c.AppliedField_Vm[index])
				row.P = polarization(c.Polarization_Cm2[index])
				break
			}
			index -= len(c.AppliedField_Vm)
		}
	case FORCDensity:
		m := r.FORC.Matrix
		i, c := index/m.Cols, index%m.Cols
		row.Ea = field(m.EaAxis[c])
		row.Eb = field(m.EbAxis[i])
		row.Density = m.Density[i][c]
		row.DensityValue = scalar(row.Density, "C/V²")
	}
	return row
}
