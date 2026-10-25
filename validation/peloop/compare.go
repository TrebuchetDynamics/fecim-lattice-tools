// Package peloop compares branch-labelled physical P-E samples without fitting
// model parameters, interpolating fields, or matching across branches.
package peloop

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"strconv"
)

const (
	maxSampleRows      = 10000
	maxSampleBytes     = 2 * 1024 * 1024
	observedNormalizer = "observed max(abs(reference P))"
)

// Sample is one physical P-E point. Indices are zero-based and consecutive within
// each ascending/descending branch, in acquisition order. Units are MV/cm and uC/cm2.
type Sample struct {
	Branch            string
	PointIndex        int
	FieldMVcm         float64
	PolarizationUCcm2 float64
}

// ErrorMetrics reports absolute errors in uC/cm2. NormalizedRMSE is a fraction,
// not a percentage, and is nil when all reference polarizations are zero.
// Each branch uses its own observed normalizer; Pooled uses all reference points.
type ErrorMetrics struct {
	Count                         int
	RMSEUCcm2                     float64
	MAEUCcm2                      float64
	Normalizer                    string
	ObservedMaxAbsReferencePUCcm2 float64
	NormalizedRMSE                *float64
}

// Metrics contains per-branch and point-weighted pooled errors.
type Metrics struct {
	PerBranch map[string]ErrorMetrics
	Pooled    ErrorMetrics
}

// ReadSamples reads exactly branch,point_index,field_MV_cm,polarization_uC_cm2.
// It preserves input order and values and rejects empty, malformed, nonfinite,
// duplicate-field or invalid-index inputs. Input is bounded to 2 MiB/10000 rows.
func ReadSamples(r io.Reader) ([]Sample, error) {
	if r == nil {
		return nil, fmt.Errorf("nil sample reader")
	}
	data, err := io.ReadAll(io.LimitReader(r, maxSampleBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read samples: %w", err)
	}
	if len(data) > maxSampleBytes {
		return nil, fmt.Errorf("sample CSV exceeds %d bytes", maxSampleBytes)
	}
	reader := csv.NewReader(bytes.NewReader(data))
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("sample CSV header: %w", err)
	}
	want := []string{"branch", "point_index", "field_MV_cm", "polarization_uC_cm2"}
	if len(header) != len(want) {
		return nil, fmt.Errorf("expected four physical sample columns")
	}
	for i := range want {
		if header[i] != want[i] {
			return nil, fmt.Errorf("sample column %d must be %q", i, want[i])
		}
	}
	var samples []Sample
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("sample row %d: %w", len(samples)+1, err)
		}
		if len(samples) == maxSampleRows {
			return nil, fmt.Errorf("sample CSV exceeds %d rows", maxSampleRows)
		}
		index, err := strconv.Atoi(row[1])
		if err != nil {
			return nil, fmt.Errorf("sample row %d index: %w", len(samples)+1, err)
		}
		field, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			return nil, fmt.Errorf("sample row %d field: %w", len(samples)+1, err)
		}
		polarization, err := strconv.ParseFloat(row[3], 64)
		if err != nil {
			return nil, fmt.Errorf("sample row %d polarization: %w", len(samples)+1, err)
		}
		samples = append(samples, Sample{row[0], index, field, polarization})
	}
	if err := validateSamples(samples); err != nil {
		return nil, err
	}
	return samples, nil
}

func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

func validateSamples(samples []Sample) error {
	if len(samples) == 0 || len(samples) > maxSampleRows {
		return fmt.Errorf("expected 1..%d samples", maxSampleRows)
	}
	next := make(map[string]int)
	fields := make(map[string]map[float64]bool)
	for i, s := range samples {
		if s.Branch != "ascending" && s.Branch != "descending" {
			return fmt.Errorf("sample %d: unknown branch %q", i, s.Branch)
		}
		if s.PointIndex != next[s.Branch] {
			return fmt.Errorf("sample %d: expected %s index %d", i, s.Branch, next[s.Branch])
		}
		if !finite(s.FieldMVcm) || !finite(s.PolarizationUCcm2) {
			return fmt.Errorf("sample %d: nonfinite physical value", i)
		}
		if fields[s.Branch] == nil {
			fields[s.Branch] = make(map[float64]bool)
		}
		if fields[s.Branch][s.FieldMVcm] {
			return fmt.Errorf("sample %d: duplicate field within %s", i, s.Branch)
		}
		fields[s.Branch][s.FieldMVcm] = true
		next[s.Branch]++
	}
	return nil
}

// Compare requires exact branch/index/field equality at every position. It never
// reorders or interpolates samples. Reference P is used only to compute residuals
// and the explicitly labelled observed normalizer, never to fit a model.
func Compare(reference, simulated []Sample) (Metrics, error) {
	if err := validateSamples(reference); err != nil {
		return Metrics{}, fmt.Errorf("reference: %w", err)
	}
	if err := validateSamples(simulated); err != nil {
		return Metrics{}, fmt.Errorf("simulated: %w", err)
	}
	if len(reference) != len(simulated) {
		return Metrics{}, fmt.Errorf("sample counts differ")
	}
	result := Metrics{PerBranch: make(map[string]ErrorMetrics)}
	for i, ref := range reference {
		sim := simulated[i]
		if ref.Branch != sim.Branch || ref.PointIndex != sim.PointIndex || ref.FieldMVcm != sim.FieldMVcm {
			return Metrics{}, fmt.Errorf("sample %d: branch/index/field mismatch", i)
		}
		errorP := sim.PolarizationUCcm2 - ref.PolarizationUCcm2
		if !finite(errorP) {
			return Metrics{}, fmt.Errorf("sample %d: unrepresentable residual", i)
		}
		branch := result.PerBranch[ref.Branch]
		branch.add(errorP, ref.PolarizationUCcm2)
		result.PerBranch[ref.Branch] = branch
		result.Pooled.add(errorP, ref.PolarizationUCcm2)
	}
	for branch, m := range result.PerBranch {
		if err := m.normalize(); err != nil {
			return Metrics{}, err
		}
		result.PerBranch[branch] = m
	}
	if err := result.Pooled.normalize(); err != nil {
		return Metrics{}, err
	}
	return result, nil
}

func (m *ErrorMetrics) add(residual, referenceP float64) {
	m.Count++
	n := float64(m.Count)
	// Weighted running means avoid squaring large finite residuals directly.
	m.RMSEUCcm2 = math.Hypot(m.RMSEUCcm2*math.Sqrt((n-1)/n), residual/math.Sqrt(n))
	m.MAEUCcm2 = m.MAEUCcm2*((n-1)/n) + math.Abs(residual)/n
	m.ObservedMaxAbsReferencePUCcm2 = math.Max(m.ObservedMaxAbsReferencePUCcm2, math.Abs(referenceP))
}

func (m *ErrorMetrics) normalize() error {
	if !finite(m.RMSEUCcm2) || !finite(m.MAEUCcm2) {
		return fmt.Errorf("unrepresentable accumulated errors")
	}
	m.Normalizer = observedNormalizer
	if m.ObservedMaxAbsReferencePUCcm2 > 0 {
		value := m.RMSEUCcm2 / m.ObservedMaxAbsReferencePUCcm2
		if !finite(value) {
			return fmt.Errorf("unrepresentable normalized RMSE")
		}
		m.NormalizedRMSE = &value
	}
	return nil
}
