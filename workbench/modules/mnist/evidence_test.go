package mnist

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	data "fecim-lattice-tools/module3-mnist/pkg/mnist"
	"fecim-lattice-tools/shared/neural"
	"fecim-lattice-tools/workbench/moduleapi"
)

func TestMNISTWebEvidence(t *testing.T) {
	ctx := fixtureContext(t)
	p := New()
	defer p.Close()
	apply(t, p, ctx, "load_model", nil, moduleapi.ResourceReference{ID: "pretrained_weights.json", Kind: "model"})
	apply(t, p, ctx, "load_dataset", nil, moduleapi.ResourceReference{ID: "t10k-images-idx3-ubyte.gz", Kind: "dataset"}, moduleapi.ResourceReference{ID: "t10k-labels-idx1-ubyte.gz", Kind: "dataset"})
	apply(t, p, ctx, "configure", map[string]string{"noise": "0"})
	apply(t, p, ctx, "sample", map[string]string{"index": "17"})
	s, _ := finish(t, p, ctx, "infer", nil)
	r := result(t, s)
	if r.Source != "dataset" || r.Preprocessing != "none" || r.DatasetSHA == "" {
		t.Fatalf("sample provenance: %+v", r)
	}
	images, labels, e := data.LoadMNIST("../../../module3-mnist/data", false)
	if e != nil {
		t.Fatal(e)
	}
	direct := neural.NewDualModeNetwork(784, 128, 10)
	if e = direct.LoadWeights("../../../module3-mnist/data/pretrained_weights.json"); e != nil {
		t.Fatal(e)
	}
	direct.SetNoiseLevel(0)
	if !reflect.DeepEqual(r.Inference, direct.Infer(images[17])) {
		t.Fatal("dataset sample was transformed")
	}
	s, j := finish(t, p, ctx, "evaluate", map[string]string{"start": "17", "count": "3"})
	r = result(t, s)
	want := neural.EvaluateDualModeDataset(direct, images[17:20], labels[17:20])
	if j.State != "succeeded" || len(r.Cases) != 1 || r.Cases[0].Total != 3 || float64(r.Cases[0].CIMCorrect)/3 != want.CIM.Accuracy {
		t.Fatalf("actual sample accounting: %+v %+v", r, j)
	}
	var cm [10][10]int
	for i := 17; i < 20; i++ {
		cm[labels[i]][direct.Infer(images[i]).CIMPrediction]++
	}
	if r.Cases[0].CIMConfusion != cm {
		t.Fatal("confusion differs from direct execution")
	}
	s, _ = finish(t, p, ctx, "sweep", map[string]string{"start": "17", "count": "3", "kind": "levels", "values": "[2,30]"})
	r = result(t, s)
	oracle, e := neural.SweepQuantizationLevels(direct, images[17:20], labels[17:20], []int{2, 30})
	if e != nil {
		t.Fatal(e)
	}
	for i, c := range r.Cases {
		if c.CIMCorrect != oracle[i].Correct || c.Total != oracle[i].Total {
			t.Fatal("sweep differs from native oracle")
		}
	}
	// The complete committed test corpus is admitted and evaluated, not downsampled.
	s, j = finish(t, p, ctx, "evaluate", map[string]string{"start": "0", "count": "10000"})
	full := result(t, s)
	if j.State != "succeeded" || full.CorpusCount != 10000 || full.Cases[0].Total != 10000 {
		t.Fatal("full evaluation silently reduced the committed test corpus")
	}
	before, _ := json.Marshal(s)
	_, e = p.Apply(ctx, moduleapi.Command{Revision: s.Revision, Action: "load_model", Resources: []moduleapi.ResourceReference{{ID: "../../../../etc/passwd", Kind: "model"}}})
	if e == nil {
		t.Fatal("root escape accepted")
	}
	after, _ := p.Snapshot(ctx)
	b, _ := json.Marshal(after)
	if string(before) != string(b) {
		t.Fatal("failed load mutated state")
	}
	s, j = finish(t, p, ctx, "infer", nil)
	if !reflect.DeepEqual(result(t, s).Inference, direct.Infer(images[17])) {
		t.Fatal("sweep changed session configuration")
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, e = p.Start(cancelled, moduleapi.Command{Revision: s.Revision, Action: "evaluate", Input: map[string]string{"count": "10000"}}); e == nil {
		t.Fatal("cancelled start admitted")
	}
	j, e = p.Start(ctx, moduleapi.Command{Revision: s.Revision, Action: "evaluate", Input: map[string]string{"count": "10000"}})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = p.Cancel(ctx, j.ID); e != nil {
		t.Fatal(e)
	}
	p.Close()
	j, e = p.Job(ctx, j.ID)
	if e != nil || j.State != "cancelled" || j.Progress.Completed >= 10000 {
		t.Fatalf("cancel accounting: %+v %v", j, e)
	}
}

func TestMNISTWebWorkAdmission(t *testing.T) {
	for _, c := range []struct {
		h, n, k   int
		sweep, ok bool
	}{{128, 10000, 1, false, true}, {1024, 10000, 1, false, true}, {128, 500, 64, true, true}, {1024, 500, 64, true, true}, {0, 1, 1, false, false}, {1025, 1, 1, false, false}, {128, 10001, 1, false, false}, {128, 501, 1, true, false}, {128, 500, 65, true, false}, {128, 0, 1, false, false}, {128, 1, 0, true, false}} {
		if e := AdmitWork(c.h, c.n, c.k, c.sweep); (e == nil) != c.ok {
			t.Fatalf("%+v: %v", c, e)
		}
	}
}
