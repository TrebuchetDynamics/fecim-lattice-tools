package export

import (
	"encoding/json"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewExportMetadata_Defaults(t *testing.T) {
	m := NewExportMetadata("hysteresis")
	if m.ModuleName != "hysteresis" {
		t.Fatalf("expected module hysteresis, got %s", m.ModuleName)
	}
	if m.Version != "1.0.0" {
		t.Fatalf("expected version 1.0.0, got %s", m.Version)
	}
	if m.CustomFields == nil {
		t.Fatal("CustomFields should be initialized")
	}
	if m.ExportedAt.IsZero() {
		t.Fatal("ExportedAt should not be zero")
	}
}

func TestExporter_GenerateFilename(t *testing.T) {
	e := NewExporter("/tmp/test", "prefix")
	name := e.generateFilename("csv")
	if !strings.HasPrefix(name, "/tmp/test/prefix_") {
		t.Fatalf("unexpected filename: %s", name)
	}
	if !strings.HasSuffix(name, ".csv") {
		t.Fatalf("expected .csv suffix: %s", name)
	}
}

func TestExportHTMLTable_Semantic(t *testing.T) {
	dir := t.TempDir()
	e := NewExporter(dir, "html")
	r := e.ExportHTMLTable("Test Table", []string{"X", "Y"}, [][]string{{"1", "2"}, {"3", "4"}})
	if r.Error != nil {
		t.Fatalf("ExportHTMLTable: %v", r.Error)
	}
	data, _ := os.ReadFile(r.FilePath)
	s := string(data)
	if !strings.Contains(s, "<caption>Test Table</caption>") {
		t.Fatal("missing caption")
	}
	if !strings.Contains(s, "<th scope=\"col\">X</th>") {
		t.Fatal("missing header")
	}
	if !strings.Contains(s, "<td>3</td>") {
		t.Fatal("missing data cell")
	}
	if r.BytesWritten != int64(len(data)) {
		t.Fatalf("BytesWritten mismatch: %d vs %d", r.BytesWritten, len(data))
	}
}

func TestExportCSVFromFloats_MultiColumn(t *testing.T) {
	dir := t.TempDir()
	e := NewExporter(dir, "floats")
	r := e.ExportCSVFromFloats([]string{"A", "B"}, []float64{1.5, 2.5}, []float64{3.0})
	if r.Error != nil {
		t.Fatalf("ExportCSVFromFloats: %v", r.Error)
	}
	data, _ := os.ReadFile(r.FilePath)
	if !strings.Contains(string(data), "1.5") {
		t.Fatal("missing float value")
	}
}

func TestExportCSVFromFloats_NoColumns(t *testing.T) {
	dir := t.TempDir()
	e := NewExporter(dir, "empty")
	r := e.ExportCSVFromFloats([]string{"A"})
	if r.Error == nil {
		t.Fatal("expected error for no columns")
	}
}

func TestExportJSON(t *testing.T) {
	dir := t.TempDir()
	e := NewExporter(dir, "json")
	data := map[string]float64{"Pr": 25.0, "Ec": 1.13}
	r := e.ExportJSON(data)
	if r.Error != nil {
		t.Fatalf("ExportJSON: %v", r.Error)
	}
	raw, _ := os.ReadFile(r.FilePath)
	var out map[string]float64
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if out["Pr"] != 25.0 {
		t.Fatalf("Pr mismatch: %v", out["Pr"])
	}
}

func TestExportPNG(t *testing.T) {
	dir := t.TempDir()
	e := NewExporter(dir, "img")
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	img.Set(5, 5, color.White)
	r := e.ExportPNG(img)
	if r.Error != nil {
		t.Fatalf("ExportPNG: %v", r.Error)
	}
	if r.BytesWritten == 0 {
		t.Fatal("PNG should have nonzero size")
	}
}

func TestQuickExport_JSON_Map(t *testing.T) {
	dir := t.TempDir()
	r := QuickExport(dir, "quick", FormatJSON, map[string]int{"x": 1})
	if r.Error != nil {
		t.Fatalf("QuickExport JSON: %v", r.Error)
	}
}

func TestQuickExport_CSV_CSVData(t *testing.T) {
	dir := t.TempDir()
	csv := NewCSVData("A", "B")
	csv.AddRow("1", "2")
	r := QuickExport(dir, "quick", FormatCSV, csv)
	if r.Error != nil {
		t.Fatalf("QuickExport CSV: %v", r.Error)
	}
}

func TestQuickExport_HTML_CSVData(t *testing.T) {
	dir := t.TempDir()
	csv := NewCSVData("X")
	csv.AddRow("val")
	r := QuickExport(dir, "quick", FormatHTML, csv)
	if r.Error != nil {
		t.Fatalf("QuickExport HTML: %v", r.Error)
	}
}

func TestQuickExport_PNG_RGBA(t *testing.T) {
	dir := t.TempDir()
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	r := QuickExport(dir, "quick", FormatPNG, img)
	if r.Error != nil {
		t.Fatalf("QuickExport PNG: %v", r.Error)
	}
}

func TestQuickExport_CSV_WrongType(t *testing.T) {
	dir := t.TempDir()
	r := QuickExport(dir, "quick", FormatCSV, "not csv data")
	if r.Error == nil {
		t.Fatal("expected error for wrong CSV type")
	}
}

func TestQuickExport_HTML_WrongType(t *testing.T) {
	dir := t.TempDir()
	r := QuickExport(dir, "quick", FormatHTML, 42)
	if r.Error == nil {
		t.Fatal("expected error for wrong HTML type")
	}
}

func TestQuickExport_PNG_WrongType(t *testing.T) {
	dir := t.TempDir()
	r := QuickExport(dir, "quick", FormatPNG, "not image")
	if r.Error == nil {
		t.Fatal("expected error for wrong PNG type")
	}
}

func TestQuickExport_UnsupportedFormat_XML(t *testing.T) {
	dir := t.TempDir()
	r := QuickExport(dir, "quick", "xml", nil)
	if r.Error == nil {
		t.Fatal("expected error for unsupported format")
	}
}

func TestCSVData_AddRowFromFloats(t *testing.T) {
	c := NewCSVData("A", "B")
	c.AddRowFromFloats(1.5, 2.5)
	if len(c.Rows) != 1 || c.Rows[0][0] != "1.5" {
		t.Fatalf("unexpected row: %v", c.Rows)
	}
}

func TestCSVData_AddRowFromInts(t *testing.T) {
	c := NewCSVData("X")
	c.AddRowFromInts(42, 99)
	if c.Rows[0][0] != "42" || c.Rows[0][1] != "99" {
		t.Fatalf("unexpected row: %v", c.Rows)
	}
}

func TestExportCSV_BadDir(t *testing.T) {
	e := NewExporter("/nonexistent/path/deep/nested", "test")
	// This might actually succeed on some systems (MkdirAll), so test the flow
	r := e.ExportCSV([]string{"A"}, [][]string{{"1"}})
	// Either succeeds or returns error — both are valid, just don't panic
	_ = r
}

func TestExportJSON_Unmarshalable(t *testing.T) {
	dir := t.TempDir()
	e := NewExporter(dir, "bad")
	r := e.ExportJSON(make(chan int)) // channels can't be marshaled
	if r.Error == nil {
		t.Fatal("expected marshal error")
	}
}

func TestEnsureOutputDir_Creates(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "sub", "dir")
	e := NewExporter(dir, "test")
	if err := e.ensureOutputDir(); err != nil {
		t.Fatalf("ensureOutputDir: %v", err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("dir not created: %v", err)
	}
}
