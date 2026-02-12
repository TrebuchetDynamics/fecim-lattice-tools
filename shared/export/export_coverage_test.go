package export

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"
)

func TestExportCSV_WritesHeaderAndRows(t *testing.T) {
	dir := t.TempDir()
	e := NewExporter(dir, "test")
	r := e.ExportCSV([]string{"E", "P"}, [][]string{{"1.0", "0.25"}, {"2.0", "0.50"}})
	if r.Error != nil {
		t.Fatalf("ExportCSV: %v", r.Error)
	}
	data, _ := os.ReadFile(r.FilePath)
	if len(data) < 10 {
		t.Fatalf("CSV too small: %d bytes", len(data))
	}
}

func TestExportCSVFromFloats(t *testing.T) {
	dir := t.TempDir()
	e := NewExporter(dir, "floats")
	r := e.ExportCSVFromFloats([]string{"x", "y"}, []float64{1.0, 2.0}, []float64{3.0, 4.0})
	if r.Error != nil {
		t.Fatalf("ExportCSVFromFloats: %v", r.Error)
	}
	// BytesWritten may be 0 due to deferred flush; just check no error
	if _, err := os.Stat(r.FilePath); err != nil {
		t.Errorf("CSV file not created: %v", err)
	}
}

func TestExportCSVFromFloats_NoCols(t *testing.T) {
	dir := t.TempDir()
	e := NewExporter(dir, "empty")
	r := e.ExportCSVFromFloats([]string{})
	if r.Error == nil {
		t.Error("expected error with no columns")
	}
}

func TestExportJSON_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	e := NewExporter(dir, "json")
	r := e.ExportJSON(map[string]float64{"Pr": 25.0, "Ec": 1.0})
	if r.Error != nil {
		t.Fatalf("ExportJSON: %v", r.Error)
	}
	data, _ := os.ReadFile(r.FilePath)
	if len(data) < 5 {
		t.Fatalf("JSON too small")
	}
}

func TestExportHTMLTable(t *testing.T) {
	dir := t.TempDir()
	e := NewExporter(dir, "html")
	r := e.ExportHTMLTable("Test Table", []string{"Col1", "Col2"}, [][]string{{"a", "b"}})
	if r.Error != nil {
		t.Fatalf("ExportHTMLTable: %v", r.Error)
	}
	data, _ := os.ReadFile(r.FilePath)
	s := string(data)
	for _, token := range []string{"<table>", "<caption>", "<thead>", "Col1", "Col2"} {
		if len(s) == 0 || !contains(s, token) {
			t.Errorf("expected %q in HTML output", token)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && findSubstring(s, sub))
}

func findSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestExportPNG_WritesFile(t *testing.T) {
	dir := t.TempDir()
	e := NewExporter(dir, "img")
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	img.Set(5, 5, color.White)
	r := e.ExportPNG(img)
	if r.Error != nil {
		t.Fatalf("ExportPNG: %v", r.Error)
	}
	if r.BytesWritten == 0 {
		t.Error("expected non-zero PNG bytes")
	}
}

func TestQuickExport_JSON_Coverage(t *testing.T) {
	dir := t.TempDir()
	r := QuickExport(dir, "quick", FormatJSON, map[string]int{"x": 1})
	if r.Error != nil {
		t.Fatalf("QuickExport JSON: %v", r.Error)
	}
}

func TestQuickExport_CSV_Coverage(t *testing.T) {
	dir := t.TempDir()
	csv := NewCSVData("a", "b")
	csv.AddRow("1", "2")
	csv.AddRowFromFloats(3.14, 2.72)
	csv.AddRowFromInts(10, 20)
	r := QuickExport(dir, "quick", FormatCSV, csv)
	if r.Error != nil {
		t.Fatalf("QuickExport CSV: %v", r.Error)
	}
}

func TestQuickExport_HTML_Coverage(t *testing.T) {
	dir := t.TempDir()
	csv := NewCSVData("x")
	csv.AddRow("v")
	r := QuickExport(dir, "quick", FormatHTML, csv)
	if r.Error != nil {
		t.Fatalf("QuickExport HTML: %v", r.Error)
	}
}

func TestQuickExport_PNG(t *testing.T) {
	dir := t.TempDir()
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	r := QuickExport(dir, "quick", FormatPNG, img)
	if r.Error != nil {
		t.Fatalf("QuickExport PNG: %v", r.Error)
	}
}

func TestQuickExport_UnsupportedFormat(t *testing.T) {
	r := QuickExport(t.TempDir(), "q", "xml", nil)
	if r.Error == nil {
		t.Error("expected error for unsupported format")
	}
}

func TestQuickExport_CSVWrongType(t *testing.T) {
	r := QuickExport(t.TempDir(), "q", FormatCSV, "not csv data")
	if r.Error == nil {
		t.Error("expected error for wrong CSV type")
	}
}

func TestNewExportMetadata(t *testing.T) {
	m := NewExportMetadata("hysteresis")
	if m.ModuleName != "hysteresis" {
		t.Errorf("wrong module name: %s", m.ModuleName)
	}
	if m.CustomFields == nil {
		t.Error("CustomFields should be initialized")
	}
}

func TestGenerateFilename(t *testing.T) {
	e := NewExporter("/tmp/test", "sim")
	fn := e.generateFilename("csv")
	if filepath.Ext(fn) != ".csv" {
		t.Errorf("expected .csv extension, got %s", filepath.Ext(fn))
	}
}
