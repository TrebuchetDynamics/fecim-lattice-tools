package server

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"io/fs"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"fecim-lattice-tools/module1-hysteresis/pkg/calibration"
	"fecim-lattice-tools/shared/presets"
	"fecim-lattice-tools/workbench/hysteresis"
)

const provenance = "browser simulation data; supplied snapshot is unverified, not a verified persisted study"

var fileKinds = []string{"preset", "calibration", "pe-json", "pe-csv", "pe-svg", "analysis-json", "pund-csv", "forc-sweep-csv", "forc-density-csv", "log-csv"}

type LoopRow struct {
	Index      int
	Time, E, P float64
}
type BrowserSnapshot struct {
	Config      hysteresis.Config
	Rows        []LoopRow
	Calibration *hysteresis.CalibrationSnapshot
	Analysis    *hysteresis.AnalysisResult
}
type FileRequest struct {
	Root      string              `json:"root"`
	Path      string              `json:"path"`
	Kind      string              `json:"kind"`
	Operation string              `json:"operation"`
	UI        string              `json:"UI"`
	Identity  hysteresis.Identity `json:"identity"`
	Snapshot  BrowserSnapshot     `json:"snapshot"`
}

func safeRelative(name string) bool {
	if !fs.ValidPath(name) || name == "." || len(name) > 240 || strings.ContainsAny(name, "\\:%\x00") {
		return false
	}
	for _, part := range strings.Split(name, "/") {
		if strings.HasSuffix(part, ".") || strings.HasSuffix(part, " ") || strings.HasPrefix(part, ".") {
			return false
		}
		base := strings.ToUpper(strings.Split(part, ".")[0])
		if base == "CON" || base == "PRN" || base == "AUX" || base == "NUL" || (len(base) == 4 && (strings.HasPrefix(base, "COM") || strings.HasPrefix(base, "LPT")) && base[3] >= '0' && base[3] <= '9') {
			return false
		}
		for _, ch := range part {
			if ch < 32 || ch == 127 {
				return false
			}
		}
	}
	return true
}
func (s *Server) files(w http.ResponseWriter, r *http.Request) {
	var q FileRequest
	if !readJSON(w, r, &q) {
		return
	}
	releaseSession, err := s.beginSession(r.Context())
	if err != nil {
		problem(w, 401, "Session ended or replaced")
		return
	}
	defer releaseSession()
	if q.Identity != s.manifest.Identity || q.UI != s.manifest.UI {
		problem(w, 409, "asset fingerprint mismatch; rebuild and reload")
		return
	}
	root := s.roots[q.Root]
	if root == nil || !safeRelative(q.Path) {
		problem(w, 400, "invalid granted root or relative path")
		return
	}
	supported := false
	for _, kind := range fileKinds {
		if q.Kind == kind {
			supported = true
		}
	}
	if !supported || (q.Operation != "save" && q.Operation != "load") || (q.Operation == "load" && q.Kind != "preset" && q.Kind != "calibration") {
		problem(w, 400, "unsupported file operation or artifact kind")
		return
	}
	if q.Operation == "load" {
		// Lstat rejects final symlinks and non-regular files before Open. os.Root
		// independently prevents ancestor-link escapes even across rename races.
		info, err := root.Lstat(q.Path)
		if err != nil || !info.Mode().IsRegular() {
			problem(w, 400, "file unavailable or not regular")
			return
		}
		if info.Size() > MaxBodyBytes {
			problem(w, 413, "file exceeds 8 MiB")
			return
		}
		f, err := root.Open(q.Path)
		if err != nil {
			problem(w, 400, "file unavailable within grant")
			return
		}
		defer f.Close()
		info, err = f.Stat()
		if err != nil || !info.Mode().IsRegular() {
			problem(w, 400, "file is not regular")
			return
		}
		if info.Size() > MaxBodyBytes {
			problem(w, 413, "file exceeds 8 MiB")
			return
		}
		data, err := io.ReadAll(io.LimitReader(f, MaxBodyBytes+1))
		if err != nil {
			problem(w, 400, "cannot read file")
			return
		}
		if len(data) > MaxBodyBytes {
			problem(w, 413, "file exceeds 8 MiB")
			return
		}
		value, err := loadFormat(q.Kind, data)
		if err != nil {
			problem(w, 400, "File data rejected; check the selected format and simulation state")
			return
		}
		jsonReply(w, map[string]any{"status": "Loaded " + q.Path + "; imported/unverified, not stored-run verification", "provenance": provenance, "value": value})
		return
	}
	data, err := render(q.Kind, q.Snapshot, path.Base(q.Path))
	if err != nil {
		problem(w, 400, "File data rejected; check the selected format and simulation state")
		return
	}
	if len(data) > MaxBodyBytes {
		problem(w, 413, "artifact exceeds 8 MiB")
		return
	}
	// Publish a complete file without overwriting: write a private sibling,
	// fsync, then root-scoped hard-link with exclusive destination semantics.
	// No rename-based check/overwrite race with external editors is claimed.
	dir := path.Dir(q.Path)
	parent, err := root.OpenRoot(dir)
	if err != nil {
		problem(w, 400, "destination directory unavailable within grant")
		return
	}
	defer parent.Close()
	temp := ".fecim-" + randomSecret()
	f, err := parent.OpenFile(temp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		problem(w, 400, "cannot stage file")
		return
	}
	defer parent.Remove(temp)
	_, err = f.Write(data)
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		problem(w, 500, "cannot stage complete artifact")
		return
	}
	err = parent.Link(temp, path.Base(q.Path))
	if errors.Is(err, os.ErrExist) {
		problem(w, 409, "Conflict: destination exists; choose a new path. Nothing overwritten.")
		return
	}
	if err != nil {
		problem(w, 400, "cannot publish within granted root")
		return
	}
	jsonReply(w, map[string]string{"status": "Saved " + q.Path + "; " + provenance, "provenance": provenance})
}
func canonical(value any) ([]byte, error) {
	b, err := json.MarshalIndent(value, "", "  ")
	return append(b, '\n'), err
}
func validConfig(c hysteresis.Config) error {
	if c == (hysteresis.Config{}) {
		return errors.New("explicit complete browser configuration required")
	}
	_, err := hysteresis.NewSession(c)
	return err
}
func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func render(kind string, s BrowserSnapshot, name string) ([]byte, error) {
	if err := validConfig(s.Config); err != nil {
		return nil, err
	}
	switch kind {
	case "preset":
		material := ""
		for _, m := range hysteresis.Catalogue() {
			if m.ID == s.Config.MaterialID {
				material = m.Parameters.Name
			}
		}
		waveform := map[hysteresis.Mode]string{hysteresis.Manual: "Manual", hysteresis.Sine: "Sine Wave", hysteresis.Triangle: "Triangle Wave"}[s.Config.Mode]
		engine := "preisach"
		if s.Config.Model == hysteresis.Landau {
			engine = "lk"
		}
		p := presets.NewPreset(name, provenance, presets.ModuleHysteresis, presets.CategoryCustom, map[string]any{"material": material, "waveform": waveform, "physics_engine": engine, "frequency": s.Config.Frequency, "time_scale": s.Config.TimeScale, "num_levels": s.Config.Levels, "temperature": s.Config.Temperature, "target_range": s.Config.RangeFraction, "stress_pa": s.Config.Stress, "history_limit": s.Config.HistoryLimit, "seed": strconv.FormatUint(s.Config.Seed, 10)})
		return canonical(p)
	case "calibration":
		c := s.Calibration
		if c == nil || c.Status != hysteresis.CalibrationFresh || !c.CanExport || c.Config != s.Config || len(c.Ascending) != s.Config.Levels || len(c.Descending) != s.Config.Levels {
			return nil, errors.New("fresh matching complete simulation calibration required")
		}
		name := ""
		for _, m := range hysteresis.Catalogue() {
			if m.ID == s.Config.MaterialID {
				name = m.Parameters.Name
			}
		}
		temp := &calibration.TempCalibration{Temperature: s.Config.Temperature}
		for i := range c.Ascending {
			if !finite(c.Ascending[i].SI) || !finite(c.Descending[i].SI) {
				return nil, errors.New("nonfinite calibration")
			}
			temp.CalibrationUp = append(temp.CalibrationUp, c.Ascending[i].SI)
			temp.CalibrationDown = append(temp.CalibrationDown, c.Descending[i].SI)
		}
		return canonical(struct {
			calibration.CalibrationData
			Provenance             string `json:"provenance"`
			ExperimentalValidation bool   `json:"experimental_validation"`
		}{calibration.CalibrationData{Version: 4, MaterialName: name, NumLevels: s.Config.Levels, TargetRangeFrac: s.Config.RangeFraction, Calibrations: map[int]*calibration.TempCalibration{int(math.Round(s.Config.Temperature)): temp}, SavedAt: time.Now().UTC().Format(time.RFC3339)}, provenance, false})
	case "analysis-json", "pund-csv", "forc-sweep-csv", "forc-density-csv":
		if s.Analysis == nil || s.Analysis.Context.Config != s.Config {
			return nil, errors.New("matching complete diagnostic required")
		}
		return hysteresis.BrowserAnalysisArtifact(*s.Analysis, hysteresis.AnalysisArtifact(kind))
	}
	if len(s.Rows) == 0 || len(s.Rows) > hysteresis.MaxSamples {
		return nil, errors.New("loop export requires 1–100000 retained rows")
	}
	for i, row := range s.Rows {
		if !finite(row.Time) || !finite(row.E) || !finite(row.P) || row.Time < 0 || (i > 0 && (row.Index <= s.Rows[i-1].Index || row.Time < s.Rows[i-1].Time)) {
			return nil, errors.New("invalid ordered finite loop rows")
		}
	}
	switch kind {
	case "pe-json":
		return canonical(struct {
			Provenance             string `json:"provenance"`
			ExperimentalValidation bool   `json:"experimental_validation"`
			Config                 hysteresis.Config
			Rows                   []LoopRow
		}{provenance, false, s.Config, s.Rows})
	case "pe-csv", "log-csv":
		var b bytes.Buffer
		w := csv.NewWriter(&b)
		w.Write([]string{"E_V_m", "P_C_m2", "time_s", "index", "evidence", "experimental_validation"})
		for _, row := range s.Rows {
			w.Write([]string{strconv.FormatFloat(row.E, 'g', -1, 64), strconv.FormatFloat(row.P, 'g', -1, 64), strconv.FormatFloat(row.Time, 'g', -1, 64), strconv.Itoa(row.Index), provenance, "false"})
		}
		w.Flush()
		return b.Bytes(), w.Error()
	case "pe-svg":
		// Only Go-generated numeric geometry and escaped fixed labels, never client SVG/HTML.
		var b strings.Builder
		b.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 800 600" role="img"><title>P-E browser simulation data, unverified</title><desc>` + html.EscapeString(provenance) + `</desc><polyline fill="none" stroke="black" points="`)
		maxE, maxP := 1.0, 1.0
		for _, row := range s.Rows {
			maxE = math.Max(maxE, math.Abs(row.E))
			maxP = math.Max(maxP, math.Abs(row.P))
		}
		for _, row := range s.Rows {
			fmt.Fprintf(&b, "%.6f,%.6f ", 400+row.E/maxE*350, 300-row.P/maxP*250)
		}
		b.WriteString(`"/></svg>`)
		return []byte(b.String()), nil
	default:
		return nil, errors.New("unsupported artifact kind")
	}
}
func loadFormat(kind string, data []byte) (any, error) {
	if kind == "preset" {
		var p presets.Preset
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, errors.New("invalid preset JSON")
		}
		if err := p.Validate(); err != nil {
			return nil, err
		}
		if p.Metadata.Module != presets.ModuleHysteresis {
			return nil, errors.New("not a Hysteresis preset")
		}
		c := hysteresis.DefaultConfig()
		if value, ok := p.GetString("material"); ok {
			found := false
			for _, m := range hysteresis.Catalogue() {
				if m.Parameters.Name == value {
					c.MaterialID = m.ID
					found = true
				}
			}
			if !found {
				return nil, errors.New("unsupported preset material")
			}
		}
		if value, ok := p.GetString("physics_engine"); ok {
			switch value {
			case "preisach":
				c.Model = hysteresis.Preisach
			case "lk":
				c.Model = hysteresis.Landau
			default:
				return nil, errors.New("unsupported preset engine")
			}
		}
		if value, ok := p.GetString("waveform"); ok {
			switch value {
			case "Manual":
				c.Mode = hysteresis.Manual
			case "Sine Wave":
				c.Mode = hysteresis.Sine
			case "Triangle Wave":
				c.Mode = hysteresis.Triangle
			default:
				return nil, errors.New("preset waveform not supported by loop")
			}
		}
		for key, dst := range map[string]*float64{"frequency": &c.Frequency, "time_scale": &c.TimeScale, "temperature": &c.Temperature, "target_range": &c.RangeFraction, "stress_pa": &c.Stress} {
			if value, ok := p.GetFloat(key); ok {
				*dst = value
			}
		}
		if value, ok := p.GetFloat("num_levels"); ok {
			if value != math.Trunc(value) {
				return nil, errors.New("noninteger preset levels")
			}
			c.Levels = int(value)
		}
		if value, ok := p.GetFloat("history_limit"); ok {
			if value != math.Trunc(value) || value < 1000 || value > hysteresis.MaxSamples {
				return nil, errors.New("invalid preset history limit")
			}
			c.HistoryLimit = int(value)
		}
		if value, ok := p.GetString("seed"); ok {
			seed, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return nil, errors.New("invalid preset seed")
			}
			c.Seed = seed
		}
		if err := validConfig(c); err != nil {
			return nil, err
		}
		return struct {
			Preset presets.Preset
			Config hysteresis.Config
		}{p, c}, nil
	}
	var c calibration.CalibrationData
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, errors.New("invalid calibration JSON")
	}
	if c.Version == 1 {
		// Legacy files describe a single 300 K table. Keep their original
		// fields/version and expose a typed read-only view, never install it.
		c.Calibrations = map[int]*calibration.TempCalibration{300: {Temperature: 300, CalibrationUp: c.CalibrationUp, CalibrationDown: c.CalibrationDown}}
	}
	if (c.Version != 4 && c.Version != 1) || c.NumLevels < 2 || c.NumLevels > 64 || c.TargetRangeFrac < 0.5 || c.TargetRangeFrac > 1 || len(c.Calibrations) == 0 || len(c.Calibrations) > 501 {
		return nil, errors.New("unsupported or invalid calibration format (version 1 or 4 required)")
	}
	found := false
	for _, m := range hysteresis.Catalogue() {
		if m.Parameters.Name == c.MaterialName {
			found = true
		}
	}
	if !found {
		return nil, errors.New("unsupported calibration material")
	}
	for _, v := range c.Calibrations {
		if v == nil || v.Temperature < 200 || v.Temperature > 700 || len(v.CalibrationUp) != c.NumLevels || len(v.CalibrationDown) != c.NumLevels {
			return nil, errors.New("invalid calibration dimensions")
		}
	}
	return c, nil
}
