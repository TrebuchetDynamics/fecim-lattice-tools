package mnist

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeGZ(t *testing.T, path string, payload []byte) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	if _, err := gz.Write(payload); err != nil {
		t.Fatalf("write gzip payload: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("close gzip: %v", err)
	}
}

func makeImageHeader(magic, count, rows, cols int32) []byte {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.BigEndian, magic)
	_ = binary.Write(buf, binary.BigEndian, count)
	_ = binary.Write(buf, binary.BigEndian, rows)
	_ = binary.Write(buf, binary.BigEndian, cols)
	return buf.Bytes()
}

func makeLabelHeader(magic, count int32) []byte {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.BigEndian, magic)
	_ = binary.Write(buf, binary.BigEndian, count)
	return buf.Bytes()
}

func TestLoadImages_BoundsCheckCount(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "train-images-idx3-ubyte.gz")
	payload := makeImageHeader(2051, maxMNISTImages+1, 28, 28)
	writeGZ(t, path, payload)

	_, err := loadImages(path)
	if err == nil || !strings.Contains(err.Error(), "invalid image count") {
		t.Fatalf("expected invalid image count error, got %v", err)
	}
}

func TestLoadLabels_BoundsCheckCount(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "train-labels-idx1-ubyte.gz")
	payload := makeLabelHeader(2049, maxMNISTLabels+1)
	writeGZ(t, path, payload)

	_, err := loadLabels(path)
	if err == nil || !strings.Contains(err.Error(), "invalid label count") {
		t.Fatalf("expected invalid label count error, got %v", err)
	}
}

func TestLoadMNIST_MismatchedCounts(t *testing.T) {
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "train-images-idx3-ubyte.gz")
	lblPath := filepath.Join(dir, "train-labels-idx1-ubyte.gz")

	img := append(makeImageHeader(2051, 2, 28, 28), make([]byte, 2*28*28)...)
	lbl := append(makeLabelHeader(2049, 1), []byte{7}...)
	writeGZ(t, imgPath, img)
	writeGZ(t, lblPath, lbl)

	_, _, err := LoadMNIST(dir, true)
	if err == nil || !strings.Contains(err.Error(), "image count") {
		t.Fatalf("expected image/label mismatch error, got %v", err)
	}
}
