package export

import (
	"os"
	"strings"
	"testing"
)

func TestExportJSON_StructRoundTrip(t *testing.T) {
	dir := t.TempDir()
	e := NewExporter(dir, "rt")
	type D struct {
		Name  string  `json:"name"`
		Value float64 `json:"value"`
	}
	r := e.ExportJSON(D{Name: "Pr", Value: 25.0})
	if r.Error != nil {
		t.Fatal(r.Error)
	}
	if r.BytesWritten == 0 {
		t.Fatal("zero bytes")
	}
	raw, _ := os.ReadFile(r.FilePath)
	if !strings.Contains(string(raw), `"name": "Pr"`) {
		t.Fatalf("missing field in JSON: %s", raw)
	}
}

func TestExportCSV_EmptyRows(t *testing.T) {
	dir := t.TempDir()
	e := NewExporter(dir, "empty")
	r := e.ExportCSV([]string{"H"}, nil)
	if r.Error != nil {
		t.Fatal(r.Error)
	}
	data, _ := os.ReadFile(r.FilePath)
	// Should have header only
	if !strings.Contains(string(data), "H") {
		t.Fatal("missing header")
	}
}

func TestExportHTMLTable_EscapesHTML(t *testing.T) {
	dir := t.TempDir()
	e := NewExporter(dir, "esc")
	r := e.ExportHTMLTable("<script>", []string{"A&B"}, [][]string{{"<b>"}})
	if r.Error != nil {
		t.Fatal(r.Error)
	}
	data, _ := os.ReadFile(r.FilePath)
	s := string(data)
	if strings.Contains(s, "<script>") && !strings.Contains(s, "&lt;script&gt;") {
		t.Fatal("HTML not escaped")
	}
}

func TestExportCSVFromFloats_UnevenColumns(t *testing.T) {
	dir := t.TempDir()
	e := NewExporter(dir, "uneven")
	r := e.ExportCSVFromFloats([]string{"A", "B"}, []float64{1.0, 2.0, 3.0}, []float64{10.0})
	if r.Error != nil {
		t.Fatal(r.Error)
	}
	data, _ := os.ReadFile(r.FilePath)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	// header + 3 data rows
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines, got %d", len(lines))
	}
}

func TestCSVData_MultipleRowTypes(t *testing.T) {
	c := NewCSVData("I", "F", "S")
	c.AddRowFromInts(1, 2, 3)
	c.AddRowFromFloats(1.1, 2.2, 3.3)
	c.AddRow("a", "b", "c")
	if len(c.Rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(c.Rows))
	}
}

func TestExportResult_Format(t *testing.T) {
	r := &ExportResult{Format: FormatCSV}
	if r.Format != "csv" {
		t.Fatalf("unexpected format: %s", r.Format)
	}
}

func TestQuickExport_CSV_BadType(t *testing.T) {
	dir := t.TempDir()
	r := QuickExport(dir, "bad", FormatCSV, 42)
	if r.Error == nil {
		t.Fatal("expected type error")
	}
	if !strings.Contains(r.Error.Error(), "CSVData") {
		t.Fatalf("unexpected error: %v", r.Error)
	}
}

func TestQuickExport_HTML_BadType(t *testing.T) {
	dir := t.TempDir()
	r := QuickExport(dir, "bad", FormatHTML, "string")
	if r.Error == nil {
		t.Fatal("expected type error")
	}
}

func TestQuickExport_PNG_BadType(t *testing.T) {
	dir := t.TempDir()
	r := QuickExport(dir, "bad", FormatPNG, 3.14)
	if r.Error == nil {
		t.Fatal("expected type error")
	}
}
