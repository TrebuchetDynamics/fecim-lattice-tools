package hysteresis

import (
	"errors"
	"fmt"
	"math"

	"fecim-lattice-tools/shared/constants"
	"fecim-lattice-tools/workbench/experiment"
	"fecim-lattice-tools/workbench/presentation"
)

// Quantity retains exact canonical SI independently of display and speech text.
type Quantity struct {
	SI           float64
	Unit         string
	Presentation presentation.Value
}

func scalar(si float64, unit string) Quantity {
	return Quantity{si, unit, presentation.FormatMetric(experiment.Metric{Value: si, Unit: unit}, "")}
}
func configValues(c Config, prefix string) map[string]Quantity {
	return map[string]Quantity{
		prefix + "Frequency": scalar(c.Frequency, "Hz"), prefix + "TimeScale": scalar(c.TimeScale, "ratio"),
		prefix + "Temperature": scalar(c.Temperature, "K"), prefix + "Stress": scalar(c.Stress, "Pa"),
		prefix + "RangeFraction": scalar(c.RangeFraction, "ratio"),
	}
}

// materialValues presents the original preset in explicit SI units, without
// changing the producer parameters (Ea is stored in eV, StressGPa in GPa).
func materialValues(m Material, prefix string) map[string]Quantity {
	p := m.Parameters
	return map[string]Quantity{
		prefix + "Pr": polarization(p.Pr), prefix + "Ps": polarization(p.Ps), prefix + "Ec": field(p.Ec),
		prefix + "Epsilon": scalar(p.Epsilon, "ratio"), prefix + "EpsilonLF": scalar(p.EpsilonLF, "ratio"), prefix + "LossAngle": scalar(p.LossAngle, "ratio"),
		prefix + "Thickness": scalar(p.Thickness, "m"), prefix + "Area": scalar(p.Area, "m²"),
		prefix + "Tau": scalar(p.Tau, "s"), prefix + "Tau0": scalar(p.Tau0, "s"),
		prefix + "Ea": scalar(p.Ea*constants.ElectronChargeC, "J"), prefix + "Alpha": scalar(p.Alpha, "ratio"),
		prefix + "K_dep":      scalar(p.K_dep, "V·m/C"),
		prefix + "BetaLandau": scalar(p.BetaLandau, "J·m⁵/C⁴"), prefix + "GammaLandau": scalar(p.GammaLandau, "J·m⁹/C⁶"),
		prefix + "RhoViscosity": scalar(p.RhoViscosity, "Ω·m"), prefix + "SeriesResistanceOhm": scalar(p.SeriesResistanceOhm, "Ω"),
		prefix + "CurieTemp": scalar(p.CurieTemp, "K"), prefix + "CurieConst": scalar(p.CurieConst, "K"),
		prefix + "TempCoeffEc": scalar(p.TempCoeffEc, "V/(m·K)"), prefix + "TempCoeffPr": scalar(p.TempCoeffPr, "C/(m²·K)"),
		prefix + "StressGPa": scalar(p.StressGPa*1e9, "Pa"), prefix + "EnduranceCycles": scalar(p.EnduranceCycles, "cycles"),
		prefix + "RetentionTime": scalar(p.RetentionTime, "s"), prefix + "ImrintField": field(p.ImrintField),
		prefix + "TargetRangeFrac": scalar(p.TargetRangeFrac, "ratio"), prefix + "Tau0NLS": scalar(p.Tau0NLS, "s"),
		prefix + "EaNLS": field(p.EaNLS), prefix + "NLSSigma": scalar(p.NLSSigma, "ratio"),
		prefix + "Gmin": scalar(p.Gmin, "S"), prefix + "Gmax": scalar(p.Gmax, "S"),
		prefix + "KvT": scalar(p.KvT, "V"), prefix + "VGSReadV": scalar(p.VGSReadV, "V"), prefix + "VT0V": scalar(p.VT0V, "V"),
		prefix + "C2CSigmaBase": scalar(p.C2CSigmaBase, "ratio"), prefix + "C2CExponent": scalar(p.C2CExponent, "ratio"),
		prefix + "Q11": scalar(p.Q11, "m⁴/C²"), prefix + "Q12": scalar(p.Q12, "m⁴/C²"),
	}
}

func field(si float64) Quantity { return Quantity{si, "V/m", presentation.FormatElectricField(si, "")} }
func polarization(si float64) Quantity {
	return Quantity{si, "C/m²", presentation.FormatPolarization(si, "")}
}

type Row struct {
	TimeValue Quantity
	Index     int
	Time      float64
	Branch    string
	E, P      Quantity
}
type PlotPoint struct {
	Index      int
	Time, E, P float64
}

// Control describes supported SI bounds and a suggested native slider interval.
// Step is an interaction increment, not scientific quantization.
type Control struct {
	Enabled              bool
	Reason, Unit         string
	Min, Max, Step       float64
	MinExclusive         bool
	SliderMin, SliderMax float64
}
type Controls struct{ ManualField, Frequency, TimeScale, Temperature, Stress, Levels, RangeFraction, History Control }

// Snapshot is bounded: use Page for all retained rows. ConfigRevision marks
// settings invalidation for consumers such as calibration/export eligibility;
// Revision binds every state/view mutation and each completed numerical step.
type Snapshot struct {
	ScientificValues                       map[string]Quantity
	Revision, ConfigRevision               uint64
	Config                                 Config
	State                                  string
	View                                   View
	Material                               Material
	Catalogue                              []Material
	Models                                 []Model
	Modes                                  []Mode
	Evidence                               string
	Assumptions                            []string
	Controls                               Controls
	E, P                                   Quantity
	Time                                   float64
	Level                                  int
	Count, Dropped, SolverSteps, ViewCount int
	Plot                                   []PlotPoint
}

func (s *Session) snapshot() Snapshot {
	mf := s.maxField()
	manual := Control{Enabled: s.config.Mode == Manual, Unit: "V/m", Min: -mf, Max: mf, Step: mf / 250}
	if !manual.Enabled {
		manual.Reason = "Automatic waveform owns the field; select Manual for direct field control"
	}
	if s.owner != nil {
		manual.Enabled = false
		manual.Reason = "Programming owns the drive; cancel or reset through Programmer"
	}
	temp := Control{Enabled: s.config.Model == Landau, Unit: "K", Min: 200, Max: 700, Step: 25}
	stress := Control{Enabled: s.config.Model == Landau, Unit: "Pa", Min: 0, Max: 5e9, Step: 1e8}
	if !temp.Enabled {
		temp.Reason = "Preisach uses fixed room temperature (300 K)"
		stress.Reason = "Preisach uses fixed default stress (1 GPa)"
	}
	start, end := s.viewBounds()
	v := Snapshot{Revision: s.revision, ConfigRevision: s.configRevision, Config: s.config, State: s.state, View: s.view, Material: s.material, Catalogue: append([]Material(nil), s.catalogue...), Models: []Model{Preisach, Landau}, Modes: []Mode{Manual, Sine, Triangle}, Evidence: "simulation_only", Assumptions: []string{
		"Material presets are educational simulation parameters, not validated device measurements",
		"Interactive L-K disables NLS and noise; programming can explicitly select seeded polydomain dynamics",
		"Discrete levels are configurable educational bins, not demonstrated device states",
	}, Controls: Controls{ManualField: manual, Frequency: Control{Enabled: true, Unit: "Hz", Min: 0, MinExclusive: true, Max: math.MaxFloat64 / (2 * math.Pi * maxTime), SliderMin: 0.01, SliderMax: 1}, TimeScale: Control{Enabled: true, Unit: "ratio", Min: 1e-9, Max: 1}, Temperature: temp, Stress: stress, Levels: Control{Enabled: true, Unit: "levels", Min: 2, Max: 64, Step: 1}, RangeFraction: Control{Enabled: true, Unit: "ratio", Min: 0.5, Max: 1}, History: Control{Enabled: true, Unit: "samples", Min: 1000, Max: MaxSamples, Step: 1}}, E: field(s.e), P: polarization(s.p), Time: s.time, Level: s.level, Count: s.count, Dropped: s.dropped, SolverSteps: s.steps, ViewCount: end - start}
	v.ScientificValues = configValues(s.config, "Config / ")
	v.ScientificValues["Time"] = scalar(s.time, "s")
	for key, value := range materialValues(s.material, "Material / Parameters / ") {
		v.ScientificValues[key] = value
	}
	// Include both ends using bounded index selection; all omitted plot points
	// remain accessible through the revision-bound table.
	n := min(MaxPlotPoints, end-start)
	v.Plot = make([]PlotPoint, 0, n)
	for i := 0; i < n; i++ {
		j := start
		if n > 1 {
			j += i * (end - start - 1) / (n - 1)
		}
		p := s.at(j)
		v.Plot = append(v.Plot, PlotPoint{p.Index, p.Time, p.E, p.P})
	}
	return v
}

// viewBounds mirrors native lastCompleteCycle without full-history copies.
// The selected view never reconstructs the solver or changes its memory.
func (s *Session) viewBounds() (int, int) {
	if s.view != LastCycle || s.count < 10 {
		return 0, s.count
	}
	first, last := -1, -1
	for i := 1; i < s.count-1; i++ {
		if s.at(i).E >= s.at(i-1).E && s.at(i).E >= s.at(i+1).E {
			first, last = last, i
		}
	}
	if first >= 0 && last-first >= 5 {
		return first, last
	}
	first, last = -1, -1
	for i := 1; i < s.count; i++ {
		if s.at(i-1).E <= 0 && s.at(i).E > 0 {
			first, last = last, i
		}
	}
	if first >= 0 && last-first >= 5 {
		return first, last
	}
	return 0, s.count
}

// PageQuery binds a 32-row window in the selected view to an exact revision.
type PageQuery struct {
	Revision uint64
	Offset   int
}
type Page struct {
	Revision      uint64
	Offset, Total int
	Position      string
	Rows          []Row
}

func (s *Session) Page(q PageQuery) (Page, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if q.Revision != s.revision {
		return Page{}, ErrStaleRevision
	}
	start, end := s.viewBounds()
	total := end - start
	if q.Offset < 0 || q.Offset > total {
		return Page{}, errors.New("page offset outside selected history")
	}
	n := min(PageSize, total-q.Offset)
	p := Page{Revision: s.revision, Offset: q.Offset, Total: total, Rows: make([]Row, 0, n)}
	for i := 0; i < n; i++ {
		r := s.at(start + q.Offset + i)
		p.Rows = append(p.Rows, Row{TimeValue: scalar(r.Time, "s"), Index: r.Index, Time: r.Time, Branch: r.Branch, E: field(r.E), P: polarization(r.P)})
	}
	if n == 0 {
		p.Position = fmt.Sprintf("No rows; %d retained samples in this view", total)
	} else {
		p.Position = fmt.Sprintf("Rows %d–%d of %d", q.Offset+1, q.Offset+n, total)
	}
	return p, nil
}
