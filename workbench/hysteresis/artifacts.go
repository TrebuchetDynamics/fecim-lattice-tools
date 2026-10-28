package hysteresis

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"strconv"
)

const MaxAnalysisArtifactBytes = 8 << 20

var ErrAnalysisArtifactLimit = errors.New("analysis artifact exceeds 8 MiB")

// The retained-point cap bounds JSON encoder scratch work; this buffer also
// enforces an explicit output limit before accumulating serializer bytes.
type artifactBuffer struct{ data bytes.Buffer }

func (b *artifactBuffer) Write(p []byte) (int, error) {
	if len(p) > MaxAnalysisArtifactBytes-b.data.Len() {
		return 0, ErrAnalysisArtifactLimit
	}
	return b.data.Write(p)
}
func (b *artifactBuffer) Bytes() []byte { return b.data.Bytes() }

type AnalysisArtifact string

const (
	AnalysisJSON   AnalysisArtifact = "analysis-json"
	PUNDCSV        AnalysisArtifact = "pund-csv"
	FORCSweepCSV   AnalysisArtifact = "forc-sweep-csv"
	FORCDensityCSV AnalysisArtifact = "forc-density-csv"
)

// Artifact returns bytes in memory, never reads/writes files or verifies a
// stored run. Persistence is a separate explicitly authorized local operation.
// JSON retains reference provenance; diagnostic CSVs carry applied parameters
// and explicitly contextual loop inputs. No client-provided evidence is trusted.
func (j *AnalysisJob) Artifact(kind AnalysisArtifact) ([]byte, error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.state != "complete" {
		return nil, ErrAnalysisIncomplete
	}
	// Serialize under the job lock rather than cloning every reference point,
	// especially for diagnostic CSVs which do not include references at all.
	r := j.result
	if kind == AnalysisJSON {
		var buf artifactBuffer
		encoder := json.NewEncoder(&buf)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(r); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
	var table AnalysisTable
	var columns []string
	switch kind {
	case PUNDCSV:
		table = PUNDTraces
		columns = []string{"pulse", "time_s", "current_a"}
	case FORCSweepCSV:
		table = FORCSweep
		columns = []string{"branch", "reversal_field_v_m", "E_V_m", "P_C_m2"}
	case FORCDensityCSV:
		table = FORCDensity
		columns = []string{"Ea_V_m", "Eb_V_m", "density_C_V2"}
	default:
		return nil, errors.New("unsupported analysis artifact")
	}
	count, err := analysisRowCount(r, table)
	if err != nil {
		return nil, err
	}
	params, err := json.Marshal(r.Diagnostic.Applied)
	if err != nil {
		return nil, err
	}
	header := []string{"diagnostic_profile", "diagnostic_version", "diagnostic_model", "diagnostic_material", "evidence", "experimental_validation", "loop_revision", "loop_config_revision", "loop_model_context_only", "loop_material_context_only", "loop_temperature_context_only_K", "loop_stress_context_only_Pa", "applied_parameters_json"}
	meta := []string{r.Diagnostic.Profile, strconv.Itoa(r.Diagnostic.Version), string(r.Diagnostic.Model), r.Diagnostic.MaterialID, r.Evidence, "false", strconv.FormatUint(r.Context.Revision, 10), strconv.FormatUint(r.Context.ConfigRevision, 10), string(r.Context.Config.Model), r.Context.Config.MaterialID, number(r.Context.Config.Temperature), number(r.Context.Config.Stress), string(params)}
	var buf artifactBuffer
	w := csv.NewWriter(&buf)
	if err := w.Write(append(header, columns...)); err != nil {
		return nil, err
	}
	for i := 0; i < count; i++ {
		row := analysisRow(r, table, i)
		var values []string
		switch table {
		case PUNDTraces:
			values = []string{row.Pulse, number(row.TimeS), number(row.CurrentA)}
		case FORCSweep:
			values = []string{row.Branch, number(row.Reversal.SI), number(row.E.SI), number(row.P.SI)}
		case FORCDensity:
			values = []string{number(row.Ea.SI), number(row.Eb.SI), number(row.Density)}
		}
		record := append(append([]string(nil), meta...), values...)
		if err := w.Write(record); err != nil {
			return nil, err
		}
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}
func number(v float64) string { return strconv.FormatFloat(v, 'g', -1, 64) }
