package mnist

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
)

const MaxLocalCorpusImages = 10000
const MaxLocalCompressedBytes = 32 << 20
const MaxLocalDecodedBytes = 16 << 20

type LocalDataset struct {
	Images [][]float64
	Labels []int
	SHA256 string // exact selected image and label file bytes, framed by length
}

// DecodeLocalDataset reads explicitly supplied rooted readers, never paths or
// the network. Both IDX headers, exact lengths and labels are validated before
// allocating float64 image arrays. No truncation, fallback or downsampling.
func DecodeLocalDataset(images, labels io.Reader) (*LocalDataset, error) {
	a, e := io.ReadAll(io.LimitReader(images, MaxLocalCompressedBytes+1))
	if e != nil {
		return nil, e
	}
	if len(a) > MaxLocalCompressedBytes {
		return nil, errors.New("corpus exceeds 32 MiB compressed bytes")
	}
	b, e := io.ReadAll(io.LimitReader(labels, int64(MaxLocalCompressedBytes-len(a)+1)))
	if e != nil {
		return nil, e
	}
	if len(a)+len(b) > MaxLocalCompressedBytes {
		return nil, errors.New("corpus exceeds 32 MiB compressed bytes")
	}
	h := sha256.New()
	binary.Write(h, binary.BigEndian, uint64(len(a)))
	h.Write(a)
	h.Write(b)
	fingerprint := hex.EncodeToString(h.Sum(nil))
	decode := func(raw []byte, limit int) ([]byte, error) {
		var r io.Reader = bytes.NewReader(raw)
		var gz *gzip.Reader
		if len(raw) >= 2 && raw[0] == 0x1f && raw[1] == 0x8b {
			var err error
			gz, err = gzip.NewReader(r)
			if err != nil {
				return nil, err
			}
			defer gz.Close()
			r = gz
		}
		out, err := io.ReadAll(io.LimitReader(r, int64(limit+1)))
		if err != nil {
			return nil, err
		}
		if len(out) > limit {
			return nil, errors.New("corpus exceeds 16 MiB decoded IDX")
		}
		return out, nil
	}
	a, e = decode(a, MaxLocalDecodedBytes)
	if e != nil {
		return nil, e
	}
	b, e = decode(b, MaxLocalDecodedBytes-len(a))
	if e != nil {
		return nil, e
	}
	if len(a) < 16 || len(b) < 8 {
		return nil, errors.New("truncated IDX header")
	}
	u := func(b []byte, i int) uint32 { return binary.BigEndian.Uint32(b[i : i+4]) }
	n := u(a, 4)
	if u(a, 0) != 2051 || u(b, 0) != 2049 || u(a, 8) != 28 || u(a, 12) != 28 || n < 1 || n > MaxLocalCorpusImages || u(b, 4) != n {
		return nil, errors.New("IDX requires matching 1..10000 labelled 28 by 28 images")
	}
	if len(a) != 16+int(n)*784 || len(b) != 8+int(n) {
		return nil, errors.New("IDX length mismatch or trailing bytes")
	}
	for _, v := range b[8:] {
		if v > 9 {
			return nil, errors.New("IDX label outside 0..9")
		}
	}
	d := &LocalDataset{Images: make([][]float64, n), Labels: make([]int, n), SHA256: fingerprint}
	for i := range d.Images {
		d.Images[i] = make([]float64, 784)
		for j := range d.Images[i] {
			d.Images[i][j] = float64(a[16+i*784+j]) / 255
		}
		d.Labels[i] = int(b[8+i])
	}
	return d, nil
}
