package neural

import (
	"bytes"
	"encoding/json"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestMNISTWebAdmission(t *testing.T) {
	// Synthetic structural models exercise admission, never accuracy claims.
	fixture := func(h int) WeightsFile {
		w := WeightsFile{Layer1Weights: make([][]float64, h), Layer2Weights: make([][]float64, 10)}
		for i := range w.Layer1Weights {
			w.Layer1Weights[i] = make([]float64, 784)
		}
		for i := range w.Layer2Weights {
			w.Layer2Weights[i] = make([]float64, h)
		}
		return w
	}
	for _, h := range []int{0, 1, 128, 1024, 1025} {
		w := fixture(h)
		b, _ := json.Marshal(w)
		_, e := DecodeMNISTModel(bytes.NewReader(b))
		if (e == nil) != (h >= 1 && h <= 1024) {
			t.Fatalf("hidden %d: %v", h, e)
		}
	}
	for _, change := range []func(*WeightsFile){func(w *WeightsFile) { w.Layer1Weights[0] = make([]float64, 783) }, func(w *WeightsFile) { w.Layer2Weights = make([][]float64, 11) }, func(w *WeightsFile) { w.Biases1 = []float64{1, 2} }, func(w *WeightsFile) { w.SingleLayerWeights = [][]float64{{1}} }, func(w *WeightsFile) { w.L1Scale = -1 }, func(w *WeightsFile) { w.QuantLevels = MaxDemoLevels + 1 }} {
		w := fixture(1)
		change(&w)
		b, _ := json.Marshal(w)
		if _, e := DecodeMNISTModel(bytes.NewReader(b)); e == nil {
			t.Fatal("malformed model accepted")
		}
	}
	for _, raw := range []string{`{"layer1_weights":[[1e999]]}`, `{"layer1_weights":null}`, `{"layer1_weights":[],"layer1_weights":[]}`, strings.Repeat("[", 20)} {
		if _, e := DecodeMNISTModel(strings.NewReader(raw)); e == nil {
			t.Fatal("invalid tokens accepted")
		}
	}
	b, _ := json.Marshal(fixture(1))
	for _, extra := range []int{0, 1} {
		r := io.MultiReader(bytes.NewReader(b), io.LimitReader(zeroSpace{}, int64((32<<20)-len(b)+extra)))
		_, e := DecodeMNISTModel(r)
		if (e == nil) != (extra == 0) {
			t.Fatalf("model bytes +%d: %v", extra, e)
		}
	}
	paths, e := filepath.Glob("../../module3-mnist/data/pretrained_weights*.json")
	if e != nil || len(paths) != 7 {
		t.Fatalf("required committed models missing: %v %v", paths, e)
	}
	for _, path := range paths {
		b, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
		}
		m, e := DecodeMNISTModel(bytes.NewReader(b))
		if e != nil {
			t.Fatalf("%s: %v", path, e)
		}
		if m.Hidden() != 128 {
			t.Fatal("committed minimum shape changed")
		}
		net := m.NewNetwork(42)
		direct := NewDualModeNetwork(784, 128, 10)
		if e := direct.LoadWeights(path); e != nil {
			t.Fatal(e)
		}
		a, second, c, d := net.GetFPWeights()
		x, y, z, q := direct.GetFPWeights()
		if !reflect.DeepEqual(a, x) || !reflect.DeepEqual(second, y) || !reflect.DeepEqual(c, z) || !reflect.DeepEqual(d, q) {
			t.Fatalf("model reconstruction changed: %s", path)
		}
	}
	w := fixture(1)
	w.L1Scale = math.MaxFloat64
	w.Layer1Weights[0][0] = 2
	b, _ = json.Marshal(w)
	if _, e := DecodeMNISTModel(bytes.NewReader(b)); e == nil {
		t.Fatal("reconstructed infinity admitted")
	}
}

type zeroSpace struct{}

func (zeroSpace) Read(b []byte) (int, error) {
	for i := range b {
		b[i] = ' '
	}
	return len(b), nil
}
