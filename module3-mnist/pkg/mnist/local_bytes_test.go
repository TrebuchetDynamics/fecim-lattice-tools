package mnist

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"io"
	"testing"
)

func TestLocalDatasetAdmission(t *testing.T) {
	fixture := func(n int) ([]byte, []byte) {
		im := new(bytes.Buffer)
		la := new(bytes.Buffer)
		binary.Write(im, binary.BigEndian, [4]uint32{2051, uint32(n), 28, 28})
		binary.Write(la, binary.BigEndian, [2]uint32{2049, uint32(n)})
		im.Write(make([]byte, n*784))
		la.Write(make([]byte, n))
		return im.Bytes(), la.Bytes()
	}
	for _, n := range []int{0, 1, 10000, 10001} {
		im, la := fixture(n)
		d, e := DecodeLocalDataset(bytes.NewReader(im), bytes.NewReader(la))
		if (e == nil) != (n > 0 && n <= 10000) {
			t.Fatalf("count %d: %v", n, e)
		}
		if e == nil && len(d.Images) != n {
			t.Fatal("silently reduced corpus")
		}
	}
	im, la := fixture(1)
	la[8] = 10
	if _, e := DecodeLocalDataset(bytes.NewReader(im), bytes.NewReader(la)); e == nil {
		t.Fatal("invalid label admitted")
	}
	la[8] = 0
	for _, bad := range [][]byte{im[:len(im)-1], append(append([]byte{}, im...), 0), func() []byte { b := append([]byte{}, im...); b[15] = 29; return b }()} {
		if _, e := DecodeLocalDataset(bytes.NewReader(bad), bytes.NewReader(la)); e == nil {
			t.Fatal("malformed IDX admitted")
		}
	}
	zipped := new(bytes.Buffer)
	gz := gzip.NewWriter(zipped)
	gz.Write(im)
	gz.Close()
	d, e := DecodeLocalDataset(bytes.NewReader(zipped.Bytes()), bytes.NewReader(la))
	if e != nil || d.Images[0][0] != 0 || d.SHA256 == "" {
		t.Fatalf("gzip: %v", e)
	}
	bomb := new(bytes.Buffer)
	gz = gzip.NewWriter(bomb)
	io.CopyN(gz, zeroBytes{}, 16<<20+1)
	gz.Close()
	if _, e := DecodeLocalDataset(bytes.NewReader(bomb.Bytes()), bytes.NewReader(la)); e == nil {
		t.Fatal("decoded byte overflow accepted")
	}
	if _, e := DecodeLocalDataset(io.LimitReader(zeroBytes{}, 32<<20+1), bytes.NewReader(la)); e == nil {
		t.Fatal("compressed byte overflow accepted")
	}
}

type zeroBytes struct{}

func (zeroBytes) Read(b []byte) (int, error) { clear(b); return len(b), nil }
