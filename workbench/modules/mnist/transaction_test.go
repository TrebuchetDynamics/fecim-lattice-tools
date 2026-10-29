package mnist

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fecim-lattice-tools/shared/neural"
	"fecim-lattice-tools/workbench/moduleapi"
	"fecim-lattice-tools/workbench/moduleapi/native"
)

func TestMNISTWebModelTransaction(t *testing.T) {
	ctx := fixtureContext(t)
	p := New()
	defer p.Close()
	good := moduleapi.ResourceReference{ID: "pretrained_weights.json", Kind: "model"}
	apply(t, p, ctx, "load_model", nil, good)
	// Synthetic structural fixtures are not accuracy evidence.
	wf := neural.WeightsFile{Layer1Weights: [][]float64{make([]float64, 784)}, Layer2Weights: make([][]float64, 10), SingleLayerWeights: make([][]float64, 10)}
	for i := range wf.Layer2Weights {
		wf.Layer2Weights[i] = []float64{0}
		wf.SingleLayerWeights[i] = make([]float64, 784)
	}
	dir := t.TempDir()
	b, _ := json.Marshal(wf)
	if e := os.WriteFile(filepath.Join(dir, "calibration.json"), b, 0600); e != nil {
		t.Fatal(e)
	}
	wf.Layer1Weights[0][0] = 1e308
	b, _ = json.Marshal(wf)
	if e := os.WriteFile(filepath.Join(dir, "overflow.json"), b, 0600); e != nil {
		t.Fatal(e)
	}
	root, e := os.OpenRoot(dir)
	if e != nil {
		t.Fatal(e)
	}
	defer root.Close()
	synthetic := native.WithResources(ctx, fixtureRoot{root})
	before, _ := p.Snapshot(ctx)
	want, _ := json.Marshal(before)
	if _, e = p.Apply(synthetic, moduleapi.Command{Action: "load_model", Revision: before.Revision, Resources: []moduleapi.ResourceReference{{ID: "overflow.json", Kind: "model"}}}); e == nil {
		t.Fatal("finite model causing non-finite quantization committed")
	}
	after, _ := p.Snapshot(ctx)
	got, e := json.Marshal(after)
	if e != nil || string(want) != string(got) {
		t.Fatal("rejected model corrupted JSON transport/state")
	}
	apply(t, p, synthetic, "load_model", nil, moduleapi.ResourceReference{ID: "calibration.json", Kind: "model"})
	apply(t, p, ctx, "view", map[string]string{"layer": "3"})
	s := apply(t, p, ctx, "load_model", nil, good)
	for _, table := range s.Tables {
		if table.ID == "weights" && !strings.Contains(table.Caption, "layer 1 ") {
			t.Fatal("replacement exposed unavailable calibration weights")
		}
	}
}
