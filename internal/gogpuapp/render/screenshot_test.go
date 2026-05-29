//go:build !cgo

package render

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gogpu/gg"
)

func TestCaptureAndSaveWritesGeneratedScreenshotUnderDocsAssets(t *testing.T) {
	tmp := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir temp: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWD) })

	err = CaptureAndSave(8, 8, func(dc *gg.Context) {
		dc.Clear()
	}, "smoke.png")
	if err != nil {
		t.Fatalf("CaptureAndSave: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmp, "docs", "assets", "screenshots", "smoke.png")); err != nil {
		t.Fatalf("expected generated screenshot under docs/assets/screenshots: %v", err)
	}
}
