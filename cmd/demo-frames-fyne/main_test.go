package main

import (
	"path/filepath"
	"testing"
)

func TestDemoFramesOutputPath(t *testing.T) {
	out := "/tmp/demo"
	got := frameOutputPath(out)
	if filepath.Base(got) != "frame_007_docs.png" {
		t.Fatalf("unexpected frame filename: %s", got)
	}
	if filepath.Dir(got) != out {
		t.Fatalf("unexpected output directory: %s", filepath.Dir(got))
	}
}
