package peloop

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"strconv"
)

// Sample is a physical P-E point in acquisition order. PointIndex is zero-based
// and consecutive within each ascending/descending branch. Fields must be
// distinct and follow the labelled direction within each branch.
type Sample struct {
	Branch            string
	PointIndex        int
	FieldMVcm         float64
	PolarizationUCcm2 float64
}

// ErrorMetrics uses uC/cm2 for RMSE and MAE. NormalizedRMSE is a fraction, not a
// percentage, and is nil for a zero observed normalizer. Each branch uses its
// own max(abs(reference P)); the pooled metric uses all reference points.
type ErrorMetrics struct {
	Count                         int
	RMSEUCcm2                     float64
	MAEUCcm2                      float64
	Normalizer                    string
	ObservedMaxAbsReferencePUCcm2 float64
	NormalizedRMSE                *float64
}

// Metrics reports per-branch and point-weighted pooled errors, without fitting.
type Metrics struct {
	PerBranch map[string]ErrorMetrics
	Pooled    ErrorMetrics
}

const (
	physicalSampleRowLimit  = 10000
	physicalSampleByteLimit = 2 * 1024 * 1024
)

// ReadSamples reads the exact physical CSV header
// branch,point_index,field_MV_cm,polarization_uC_cm2. It preserves values and
// order, accepts at most 10000 rows/2 MiB, and returns no partial data on error.
func ReadSamples(r io.Reader) ([]Sample, error) {
	if r == nil {
		return nil, fmt.Errorf("samples: nil reader")
	}
	limited := &io.LimitedReader{R: r, N: physicalSampleByteLimit + 1}
	reader := csv.NewReader(limited)
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("samples header: %w", err)
	}
	want := []string{"branch", "point_index", "field_MV_cm", "polarization_uC_cm2"}
	if len(header) != len(want) {
		return nil, fmt.Errorf("samples: expected four physical columns")
	}
	for i := range want {
		if header[i] != want[i] {
			return nil, fmt.Errorf("samples: column %d must be %q", i, want[i])
		}
	}
	var samples []Sample
	for {
		row, err := reader.Read()
		if limited.N == 0 {
			return nil, fmt.Errorf("samples: CSV exceeds 2 MiB")
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("sample row %d: %w", len(samples)+1, err)
		}
		if len(samples) == physicalSampleRowLimit {
			return nil, fmt.Errorf("samples: too many rows")
		}
		index, err := strconv.Atoi(row[1])
		if err != nil {
			return nil, fmt.Errorf("sample index: %w", err)
		}
		field, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			return nil, fmt.Errorf("sample field: %w", err)
		}
		polarization, err := strconv.ParseFloat(row[3], 64)
		if err != nil {
			return nil, fmt.Errorf("sample polarization: %w", err)
		}
		samples = append(samples, Sample{row[0], index, field, polarization})
	}
	if err := validatePhysicalSamples(samples); err != nil {
		return nil, err
	}
	return samples, nil
}

func validatePhysicalSamples(samples []Sample) error {
	if len(samples) == 0 || len(samples) > physicalSampleRowLimit {
		return fmt.Errorf("samples: require 1..%d points", physicalSampleRowLimit)
	}
	counts := make(map[string]int)
	lastFields := make(map[string]float64)
	for i, s := range samples {
		if s.Branch != "ascending" && s.Branch != "descending" {
			return fmt.Errorf("sample %d: unsupported branch %q", i, s.Branch)
		}
		if s.PointIndex != counts[s.Branch] || !finite(s.FieldMVcm) || !finite(s.PolarizationUCcm2) {
			return fmt.Errorf("sample %d: invalid index or nonfinite value", i)
		}
		if counts[s.Branch] > 0 {
			previous := lastFields[s.Branch]
			if (s.Branch == "ascending" && s.FieldMVcm <= previous) || (s.Branch == "descending" && s.FieldMVcm >= previous) {
				return fmt.Errorf("sample %d: field repeats or reverses branch direction", i)
			}
		}
		counts[s.Branch]++
		lastFields[s.Branch] = s.FieldMVcm
	}
	return nil
}

// Compare requires equal-length inputs with exactly matching branch, index and
// field at every position. It never interpolates, reorders or matches branches.
// Reference P supplies only residuals and observed normalizers, never a fit.
func Compare(reference, simulated []Sample) (Metrics, error) {
	if err := validatePhysicalSamples(reference); err != nil {
		return Metrics{}, fmt.Errorf("reference: %w", err)
	}
	if err := validatePhysicalSamples(simulated); err != nil {
		return Metrics{}, fmt.Errorf("simulated: %w", err)
	}
	if len(reference) != len(simulated) {
		return Metrics{}, fmt.Errorf("samples: reference/simulation counts differ")
	}
	result := Metrics{PerBranch: make(map[string]ErrorMetrics)}
	for i, ref := range reference {
		sim := simulated[i]
		if ref.Branch != sim.Branch || ref.PointIndex != sim.PointIndex || ref.FieldMVcm != sim.FieldMVcm {
			return Metrics{}, fmt.Errorf("sample %d: branch/index/field mismatch", i)
		}
		residual := sim.PolarizationUCcm2 - ref.PolarizationUCcm2
		if !finite(residual) {
			return Metrics{}, fmt.Errorf("sample %d: residual overflows", i)
		}
		branch := result.PerBranch[ref.Branch]
		branch.accumulate(residual, ref.PolarizationUCcm2)
		result.PerBranch[ref.Branch] = branch
		result.Pooled.accumulate(residual, ref.PolarizationUCcm2)
	}
	for name, metric := range result.PerBranch {
		if err := metric.normalizeObserved(); err != nil {
			return Metrics{}, err
		}
		result.PerBranch[name] = metric
	}
	if err := result.Pooled.normalizeObserved(); err != nil {
		return Metrics{}, err
	}
	return result, nil
}

func (m *ErrorMetrics) accumulate(residual, referenceP float64) {
	m.Count++
	n := float64(m.Count)
	// Running means and Hypot avoid overflow from squaring large residuals.
	m.RMSEUCcm2 = math.Hypot(m.RMSEUCcm2*math.Sqrt((n-1)/n), residual/math.Sqrt(n))
	m.MAEUCcm2 = m.MAEUCcm2*((n-1)/n) + math.Abs(residual)/n
	m.ObservedMaxAbsReferencePUCcm2 = math.Max(m.ObservedMaxAbsReferencePUCcm2, math.Abs(referenceP))
}

func (m *ErrorMetrics) normalizeObserved() error {
	if !finite(m.RMSEUCcm2) || !finite(m.MAEUCcm2) {
		return fmt.Errorf("samples: error metrics overflow")
	}
	m.Normalizer = "observed max(abs(reference P))"
	if m.ObservedMaxAbsReferencePUCcm2 != 0 {
		value := m.RMSEUCcm2 / m.ObservedMaxAbsReferencePUCcm2
		if !finite(value) {
			return fmt.Errorf("samples: normalized RMSE overflows")
		}
		m.NormalizedRMSE = &value
	}
	return nil
}
