package export

import (
	"context"
	"errors"
	"image"
	"image/color"
	"os"
	"strings"
	"testing"
)

func TestProgressExporter_CSVAndPNGAndCancel(t *testing.T) {
	dir := t.TempDir()
	pe := NewProgressExporter(dir, "progress", "test export", 2)

	csvRes := pe.ExportCSVWithProgress([]string{"a", "b"}, [][]string{{"1", "2"}, {"3", "4"}})
	if csvRes.Error != nil {
		t.Fatalf("ExportCSVWithProgress error: %v", csvRes.Error)
	}
	if _, err := os.Stat(csvRes.FilePath); err != nil {
		t.Fatalf("csv file missing: %v", err)
	}

	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	img.Set(1, 1, color.White)
	pngRes := pe.ExportPNGWithProgress(img)
	if pngRes.Error != nil {
		t.Fatalf("ExportPNGWithProgress error: %v", pngRes.Error)
	}
	if pngRes.BytesWritten <= 0 {
		t.Fatalf("expected png bytes written > 0, got %d", pngRes.BytesWritten)
	}

	pe2 := NewProgressExporter(dir, "cancel-me", "cancel export", 10)
	pe2.Progress.Start()
	pe2.Cancel()
	if !pe2.IsCancelled() {
		t.Fatal("expected exporter to be cancelled")
	}
}

func TestProgressExporter_FloatsAndCancelledContext(t *testing.T) {
	dir := t.TempDir()
	pe := NewProgressExporter(dir, "floats", "float export", 3)
	res := pe.ExportCSVFromFloatsWithProgress([]string{"x", "y"}, []float64{1, 2, 3}, []float64{4, 5})
	if res.Error != nil {
		t.Fatalf("ExportCSVFromFloatsWithProgress error: %v", res.Error)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	pe2 := NewProgressExporterWithContext(ctx, dir, "cancelled", "cancelled export", 5)
	res2 := pe2.ExportCSVWithProgress([]string{"x"}, [][]string{{"1"}, {"2"}})
	if !errors.Is(res2.Error, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", res2.Error)
	}
}

func TestBatchExporter_SummaryAndCallback(t *testing.T) {
	dir := t.TempDir()
	be := NewBatchExporter(dir, "batch", 2)
	be.Start()

	calls := 0
	be.OnProgress = func(item, total int, result *ExportResult) {
		calls++
	}

	ok := be.ExportItem("good", FormatJSON, map[string]int{"n": 1})
	if ok.Error != nil {
		t.Fatalf("expected first export success, got %v", ok.Error)
	}
	bad := be.ExportItem("bad", ExportFormat("unknown"), nil)
	if bad.Error == nil {
		t.Fatal("expected second export to fail")
	}
	be.Complete()

	s, f := be.Summary()
	if s != 1 || f != 1 {
		t.Fatalf("summary mismatch: got success=%d fail=%d", s, f)
	}
	if calls != 2 {
		t.Fatalf("expected 2 progress callbacks, got %d", calls)
	}

	be2 := NewBatchExporter(dir, "batch-cancel", 1)
	be2.Start()
	be2.Cancel()
	if !be2.Progress.IsCancelled() {
		t.Fatal("expected batch progress to be cancelled")
	}
}

func TestSimulationExporter_MetadataIncludesParameters(t *testing.T) {
	dir := t.TempDir()
	se := NewSimulationExporter(dir, "hysteresis", 2)
	se.SetParameter("temperature_K", 300)

	dataRes, metaRes := se.ExportWithMetadata([]string{"E", "P"}, [][]string{{"1", "0.1"}, {"2", "0.2"}})
	if dataRes == nil || dataRes.Error != nil {
		t.Fatalf("data export failed: %+v", dataRes)
	}
	if metaRes == nil || metaRes.Error != nil {
		t.Fatalf("metadata export failed: %+v", metaRes)
	}

	metaBytes, err := os.ReadFile(metaRes.FilePath)
	if err != nil {
		t.Fatalf("read metadata: %v", err)
	}
	meta := string(metaBytes)
	for _, token := range []string{"hysteresis simulation results", "temperature_K", "data_points", "data_file"} {
		if !strings.Contains(meta, token) {
			t.Fatalf("metadata missing token %q", token)
		}
	}
}
