package hysteresis

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"
)

const (
	MaxReferenceBytes  = 1 << 20
	MaxReferencePoints = 4096 // aggregate job limit as well as the per-branch limit
)

type ReferenceInput struct {
	Format, Source, Material string
	Data                     []byte
}
type ReferenceUnits struct{ EFieldUnit, PolarizationUnit string }
type ReferencePoint struct{ E, P Quantity }
type ComparisonAvailability struct {
	Available bool
	Reason    string
}
type Reference struct {
	Source, Material, Branch, Evidence string
	SourceUnits                        ReferenceUnits
	ExperimentalValidation             bool
	Points                             []ReferencePoint
	Comparison                         ComparisonAvailability
}
type Citation struct{ Title, URL, Scope string }
type ReferenceBand struct{ Min, Max Quantity }
type ReferenceContext struct {
	Evidence               string
	ExperimentalValidation bool
	PrBand, EcBand         ReferenceBand
	Citations              []Citation
	Comparison             ComparisonAvailability
	Notice                 string
}

func materialReferenceContext(m Material) ReferenceContext {
	return ReferenceContext{Evidence: "simulation_default", PrBand: ReferenceBand{polarization(m.Parameters.Pr * 0.8), polarization(m.Parameters.Pr * 1.2)}, EcBand: ReferenceBand{field(m.Parameters.Ec * 0.7), field(m.Parameters.Ec * 1.3)}, Citations: []Citation{
		{Title: "Materlik et al., J. Appl. Phys. 117, 134109 (2015)", URL: "https://doi.org/10.1063/1.4916229", Scope: "HfO2 model background; not measurement evidence for this diagnostic or every preset"},
	}, Comparison: ComparisonAvailability{Reason: "No branch-matched verified measurement comparison was supplied"}, Notice: "Pr ±20% and Ec ±30% bands reproduce illustrative GUI preset ranges, not experimental uncertainty or evidence"}
}

// referenceNumbers bounds allocation while decoding JSON arrays, rather than
// first admitting an arbitrarily large number of short numeric tokens.
type referenceNumbers []float64

func (v *referenceNumbers) UnmarshalJSON(data []byte) error {
	d := json.NewDecoder(bytes.NewReader(data))
	token, err := d.Token()
	if err != nil || token != json.Delim('[') {
		return errors.New("reference values must be numeric arrays")
	}
	for d.More() {
		if len(*v) >= MaxReferencePoints {
			return errors.New("reference point limit exceeded")
		}
		var value *float64
		if err := d.Decode(&value); err != nil {
			return err
		}
		if value == nil || !finite(*value) {
			return errors.New("reference value must be a finite number, not null")
		}
		*v = append(*v, *value)
	}
	_, err = d.Token()
	return err
}

// ParseReference consumes bytes only. Supported formats match the native overlay:
// CSV E_MV_cm,P_uC_cm2, or JSON source/material/E_V_m/P_C_m2. One strictly
// monotonic branch is preserved in input order; full loops must be split by the
// caller, never sorted/collapsed into a fabricated single-valued comparison.
func ParseReference(in ReferenceInput) (Reference, error) {
	return parseReference(in, MaxReferencePoints)
}

func parseReference(in ReferenceInput, remaining int) (Reference, error) {
	if remaining < 2 {
		return Reference{}, errors.New("aggregate reference point limit exceeded")
	}
	fail := func(reason string) (Reference, error) { return Reference{}, errors.New(reason) }
	if len(in.Data) == 0 || len(in.Data) > MaxReferenceBytes || len(in.Source) > 1024 || len(in.Material) > 256 {
		return fail("reference bytes/metadata outside supported bounds")
	}
	r := Reference{Source: in.Source, Material: in.Material, Evidence: "imported_unverified", Comparison: ComparisonAvailability{Reason: "Imported reference is unverified; no branch-matched comparison has been computed"}}
	var e, p []float64
	expectedBranch := ""
	switch in.Format {
	case "csv":
		r.SourceUnits = ReferenceUnits{"MV/cm", "uC/cm2"}
		reader := csv.NewReader(bytes.NewReader(in.Data))
		header, err := reader.Read()
		if err != nil || len(header) != 2 || strings.TrimSpace(header[0]) != "E_MV_cm" || strings.TrimSpace(header[1]) != "P_uC_cm2" {
			return fail("expected CSV header E_MV_cm,P_uC_cm2")
		}
		for {
			row, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return Reference{}, err
			}
			if len(e) >= remaining {
				return fail("reference point limit exceeded")
			}
			ev, err := strconv.ParseFloat(strings.TrimSpace(row[0]), 64)
			if err != nil {
				return Reference{}, err
			}
			pv, err := strconv.ParseFloat(strings.TrimSpace(row[1]), 64)
			if err != nil {
				return Reference{}, err
			}
			e = append(e, ev*1e8)
			p = append(p, pv*1e-2)
		}
	case "json":
		r.SourceUnits = ReferenceUnits{"V/m", "C/m2"}
		var raw struct {
			Source   string           `json:"source"`
			Material string           `json:"material"`
			Branch   string           `json:"branch"`
			E        referenceNumbers `json:"E_V_m"`
			P        referenceNumbers `json:"P_C_m2"`
		}
		d := json.NewDecoder(bytes.NewReader(in.Data))
		token, err := d.Token()
		if err != nil || token != json.Delim('{') {
			return fail("reference JSON must be an object")
		}
		seen := make(map[string]bool, 5)
		for d.More() {
			token, err := d.Token()
			if err != nil {
				return Reference{}, err
			}
			key, ok := token.(string)
			if !ok {
				return fail("invalid reference member")
			}
			key = strings.ToLower(key)
			if seen[key] {
				return fail("duplicate reference member " + key)
			}
			seen[key] = true
			var destination any
			switch key {
			case "source":
				destination = &raw.Source
			case "material":
				destination = &raw.Material
			case "branch":
				destination = &raw.Branch
			case "e_v_m":
				destination = &raw.E
			case "p_c_m2":
				destination = &raw.P
			default:
				return fail("unsupported reference member " + key)
			}
			if err := d.Decode(destination); err != nil {
				return Reference{}, err
			}
		}
		if _, err := d.Token(); err != nil {
			return Reference{}, err
		}
		if err := d.Decode(new(any)); err != io.EOF {
			return fail("unexpected trailing JSON")
		}
		if len(raw.Source) > 1024 || len(raw.Material) > 256 {
			return fail("reference metadata too large")
		}
		if raw.Source != "" {
			if r.Source != "" && r.Source != raw.Source {
				return fail("conflicting reference sources")
			}
			r.Source = raw.Source
		}
		if raw.Material != "" {
			if r.Material != "" && r.Material != raw.Material {
				return fail("conflicting reference materials")
			}
			r.Material = raw.Material
		}
		e, p = raw.E, raw.P
		expectedBranch = raw.Branch
	default:
		return fail("unsupported reference format")
	}
	if len(e) > remaining {
		return fail("aggregate reference point limit exceeded")
	}
	if len(e) < 2 || len(e) != len(p) {
		return fail("reference requires matching E/P arrays with at least two points")
	}
	r.Branch = "ascending"
	if e[1] < e[0] {
		r.Branch = "descending"
	}
	for i := range e {
		if !finite(e[i]) || !finite(p[i]) {
			return fail("reference values must be finite after SI conversion")
		}
		if i > 0 && ((r.Branch == "ascending" && e[i] <= e[i-1]) || (r.Branch == "descending" && e[i] >= e[i-1])) {
			return fail("reference E must be strictly monotonic on one branch")
		}
	}
	if expectedBranch != "" && expectedBranch != r.Branch {
		return fail("reference branch conflicts with field order")
	}
	r.Points = make([]ReferencePoint, len(e))
	for i := range e {
		r.Points[i] = ReferencePoint{field(e[i]), polarization(p[i])}
	}
	return r, nil
}
