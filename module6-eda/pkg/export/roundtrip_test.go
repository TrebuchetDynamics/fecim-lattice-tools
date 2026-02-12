package export

import (
	"encoding/csv"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
)

const floatTol = 1e-3

func TestEDAExportRoundTripFidelity(t *testing.T) {
	t.Parallel()

	// (1) Generate lattice config programmatically
	weights := [][]float64{
		{0.125, -0.250, 0.375, -0.500},
		{-0.625, 0.750, -0.875, 1.000},
		{0.333, -0.666, 0.999, 0.000},
	}
	cfg := compiler.NewComputeConfig(8, 8)
	cfg.Name = "roundtrip_test_array"
	cfg.With1T1R()
	cfg.WithWeights(weights)

	design, err := compiler.GenerateDesign(cfg)
	if err != nil {
		t.Fatalf("GenerateDesign failed: %v", err)
	}

	activeCells := filterActiveCells(design)
	if len(activeCells) == 0 {
		t.Fatal("expected non-empty active cells")
	}

	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "roundtrip.json")
	csvPath := filepath.Join(dir, "roundtrip.csv")
	defPath := filepath.Join(dir, "roundtrip.def")

	// (2) Export to all supported formats (JSON, CSV, DEF-like)
	if err := ExportJSON(design, jsonPath); err != nil {
		t.Fatalf("ExportJSON failed: %v", err)
	}
	if err := ExportCSV(design, csvPath); err != nil {
		t.Fatalf("ExportCSV failed: %v", err)
	}
	if err := ExportDEF(design, defPath); err != nil {
		t.Fatalf("ExportDEF failed: %v", err)
	}

	// (3) Re-import exported data
	jsonDesign, err := importJSONDesign(jsonPath)
	if err != nil {
		t.Fatalf("JSON re-import failed: %v", err)
	}
	csvCells, err := importCSVCells(csvPath)
	if err != nil {
		t.Fatalf("CSV re-import failed: %v", err)
	}
	defPlacements, err := importDEFPlacements(defPath)
	if err != nil {
		t.Fatalf("DEF re-import failed: %v", err)
	}

	// (4) Verify round-trip fidelity
	assertJSONRoundTrip(t, activeCells, filterActiveCells(&jsonDesign))
	assertCSVRoundTrip(t, activeCells, csvCells)
	assertDEFRoundTrip(t, design, activeCells, defPlacements)

	// (5) Verify export file sizes are reasonable
	assertFileSizeReasonable(t, jsonPath, 200, 1_000_000)
	assertFileSizeReasonable(t, csvPath, 100, 500_000)
	assertFileSizeReasonable(t, defPath, 200, 2_000_000)
}

func importJSONDesign(path string) (compiler.ArrayDesign, error) {
	var design compiler.ArrayDesign
	b, err := os.ReadFile(path)
	if err != nil {
		return design, err
	}
	err = json.Unmarshal(b, &design)
	return design, err
}

func importCSVCells(path string) ([]compiler.CellAssignment, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) < 2 {
		return nil, nil
	}

	header := strings.Join(records[0], ",")
	hasWeight := strings.Contains(header, "weight")

	out := make([]compiler.CellAssignment, 0, len(records)-1)
	for _, rec := range records[1:] {
		if len(rec) < 6 {
			continue
		}
		row, _ := strconv.Atoi(rec[0])
		col, _ := strconv.Atoi(rec[1])

		idx := 2
		cell := compiler.CellAssignment{Row: row, Col: col}
		if hasWeight {
			w, _ := strconv.ParseFloat(rec[idx], 64)
			cell.InitialWeight = w
			idx++
		}
		level, _ := strconv.Atoi(rec[idx])
		g, _ := strconv.ParseFloat(rec[idx+1], 64)
		r, _ := strconv.ParseFloat(rec[idx+2], 64)
		v, _ := strconv.ParseFloat(rec[idx+3], 64)

		cell.Level = level
		cell.Conductance = g
		cell.Resistance = r
		cell.ProgramV = v
		out = append(out, cell)
	}
	return out, nil
}

type defPlacement struct {
	row int
	col int
	xDBU int
	yDBU int
}

func importDEFPlacements(path string) (map[string]defPlacement, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(b), "\n")

	// Example line:
	// - R_1_2 fecim_1t1r + FIXED ( 11920 13400 ) N ;
	re := regexp.MustCompile(`-\s+R_(\d+)_(\d+)\s+\S+\s+\+\s+FIXED\s+\(\s*(\d+)\s+(\d+)\s*\)\s+N\s*;`)
	out := make(map[string]defPlacement)

	for _, ln := range lines {
		m := re.FindStringSubmatch(ln)
		if len(m) != 5 {
			continue
		}
		row, _ := strconv.Atoi(m[1])
		col, _ := strconv.Atoi(m[2])
		x, _ := strconv.Atoi(m[3])
		y, _ := strconv.Atoi(m[4])
		k := key(row, col)
		out[k] = defPlacement{row: row, col: col, xDBU: x, yDBU: y}
	}
	return out, nil
}

func assertJSONRoundTrip(t *testing.T, expected, actual []compiler.CellAssignment) {
	t.Helper()
	if len(actual) != len(expected) {
		t.Fatalf("JSON round-trip count mismatch: got %d want %d", len(actual), len(expected))
	}
	actualByKey := toMap(actual)
	for _, e := range expected {
		a, ok := actualByKey[key(e.Row, e.Col)]
		if !ok {
			t.Fatalf("JSON missing cell (%d,%d)", e.Row, e.Col)
		}
		assertCellApproxEqual(t, "JSON", e, a)
	}
}

func assertCSVRoundTrip(t *testing.T, expected, actual []compiler.CellAssignment) {
	t.Helper()
	if len(actual) != len(expected) {
		t.Fatalf("CSV round-trip count mismatch: got %d want %d", len(actual), len(expected))
	}
	actualByKey := toMap(actual)
	for _, e := range expected {
		a, ok := actualByKey[key(e.Row, e.Col)]
		if !ok {
			t.Fatalf("CSV missing cell (%d,%d)", e.Row, e.Col)
		}
		assertCellApproxEqual(t, "CSV", e, a)
	}
}

func assertDEFRoundTrip(t *testing.T, design *compiler.ArrayDesign, expected []compiler.CellAssignment, placements map[string]defPlacement) {
	t.Helper()
	if len(placements) != len(expected) {
		t.Fatalf("DEF round-trip count mismatch: got %d want %d", len(placements), len(expected))
	}

	cfg := DEFConfigFrom(design)
	dbu := float64(cfg.DatabaseUnit)
	for _, e := range expected {
		p, ok := placements[key(e.Row, e.Col)]
		if !ok {
			t.Fatalf("DEF missing placement for cell (%d,%d)", e.Row, e.Col)
		}
		expectedX := int((cfg.OriginX + float64(e.Col)*cfg.CellWidth) * dbu)
		expectedY := int((cfg.OriginY + float64(e.Row)*cfg.CellHeight) * dbu)
		if p.xDBU != expectedX || p.yDBU != expectedY {
			t.Fatalf("DEF coordinate mismatch for (%d,%d): got (%d,%d) want (%d,%d)",
				e.Row, e.Col, p.xDBU, p.yDBU, expectedX, expectedY)
		}
	}
}

func assertFileSizeReasonable(t *testing.T, path string, minBytes, maxBytes int64) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s failed: %v", path, err)
	}
	if info.Size() < minBytes {
		t.Fatalf("file %s unexpectedly small: %d < %d bytes", filepath.Base(path), info.Size(), minBytes)
	}
	if info.Size() > maxBytes {
		t.Fatalf("file %s unexpectedly large: %d > %d bytes", filepath.Base(path), info.Size(), maxBytes)
	}
}

func toMap(cells []compiler.CellAssignment) map[string]compiler.CellAssignment {
	m := make(map[string]compiler.CellAssignment, len(cells))
	for _, c := range cells {
		m[key(c.Row, c.Col)] = c
	}
	return m
}

func assertCellApproxEqual(t *testing.T, format string, expected, actual compiler.CellAssignment) {
	t.Helper()
	if expected.Row != actual.Row || expected.Col != actual.Col {
		t.Fatalf("%s row/col mismatch: got (%d,%d) want (%d,%d)", format, actual.Row, actual.Col, expected.Row, expected.Col)
	}
	if expected.Level != actual.Level {
		t.Fatalf("%s level mismatch at (%d,%d): got %d want %d", format, expected.Row, expected.Col, actual.Level, expected.Level)
	}
	assertClose(t, format+" conductance", expected.Conductance, actual.Conductance, 1e-3)
	assertClose(t, format+" resistance", expected.Resistance, actual.Resistance, 1e-1)
	assertClose(t, format+" programV", expected.ProgramV, actual.ProgramV, 1e-3)
	assertClose(t, format+" weight", expected.InitialWeight, actual.InitialWeight, 1e-3)
}

func assertClose(t *testing.T, field string, want, got, tol float64) {
	t.Helper()
	if math.Abs(want-got) > tol {
		t.Fatalf("%s mismatch: got %.8f want %.8f (tol %.2e)", field, got, want, tol)
	}
}

func key(row, col int) string {
	return strconv.Itoa(row) + ":" + strconv.Itoa(col)
}
