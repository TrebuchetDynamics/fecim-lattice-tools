package mnist

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"fecim-lattice-tools/shared/neural"
	"fecim-lattice-tools/workbench/moduleapi"
	"fecim-lattice-tools/workbench/moduleapi/native"
)

// Files are committed native fixtures, not downloaded or synthetic accuracy evidence.
type fixtureRoot struct{ root *os.Root }

func (f fixtureRoot) Open(ctx context.Context, ref moduleapi.ResourceReference) (io.ReadCloser, native.ResourceInfo, error) {
	r, err := f.root.Open(ref.ID)
	if err != nil {
		return nil, native.ResourceInfo{}, err
	}
	st, err := r.Stat()
	if err != nil {
		r.Close()
		return nil, native.ResourceInfo{}, err
	}
	return r, native.ResourceInfo{ID: ref.ID, Kind: ref.Kind, Size: st.Size()}, ctx.Err()
}
func fixtureContext(t *testing.T) context.Context {
	t.Helper()
	root, err := os.OpenRoot("../../../module3-mnist/data")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { root.Close() })
	return native.WithResources(context.Background(), fixtureRoot{root})
}
func apply(t *testing.T, p *Port, ctx context.Context, action string, input map[string]string, refs ...moduleapi.ResourceReference) moduleapi.Snapshot {
	t.Helper()
	s, e := p.Snapshot(ctx)
	if e != nil {
		t.Fatal(e)
	}
	s, e = p.Apply(ctx, moduleapi.Command{Revision: s.Revision, Action: action, Input: input, Resources: refs})
	if e != nil {
		t.Fatal(e)
	}
	return s
}
func finish(t *testing.T, p *Port, ctx context.Context, action string, input map[string]string) (moduleapi.Snapshot, moduleapi.Job) {
	t.Helper()
	s, _ := p.Snapshot(ctx)
	j, e := p.Start(ctx, moduleapi.Command{Revision: s.Revision, Action: action, Input: input})
	if e != nil {
		t.Fatal(e)
	}
	deadline := time.After(3 * time.Minute)
	for j.State == "running" || j.State == "cancelling" {
		select {
		case <-deadline:
			t.Fatal("job did not terminate")
		case <-time.After(time.Millisecond):
		}
		j, e = p.Job(ctx, j.ID)
		if e != nil {
			t.Fatal(e)
		}
	}
	s, e = p.Snapshot(ctx)
	if e != nil {
		t.Fatal(e)
	}
	return s, j
}
func result(t *testing.T, s moduleapi.Snapshot) Result {
	t.Helper()
	for _, f := range s.Fields {
		if f.ID == "result" {
			var r Result
			if e := json.Unmarshal([]byte(f.Input), &r); e != nil {
				t.Fatal(e)
			}
			return r
		}
	}
	t.Fatal("missing result")
	return Result{}
}
func TestMNISTWebWorkflow(t *testing.T) {
	ctx := fixtureContext(t)
	p := New()
	defer p.Close()
	s, _ := p.Snapshot(ctx)
	if s.Status != "unavailable: select a model" {
		t.Fatalf("missing model presented as ready: %s", s.Status)
	}
	if _, e := p.Start(ctx, moduleapi.Command{Action: "infer"}); e == nil {
		t.Fatal("missing model admitted")
	}
	apply(t, p, ctx, "load_model", nil, moduleapi.ResourceReference{ID: "pretrained_weights.json", Kind: "model"})
	input := make([]float64, 784)
	for y := 2; y < 24; y++ {
		input[y*28+2] = 0.7
	}
	raw, _ := json.Marshal(input)
	apply(t, p, ctx, "configure", map[string]string{"noise": "0", "levels": "30"})
	apply(t, p, ctx, "input", map[string]string{"pixels": string(raw), "source": "drawing"})
	s, j := finish(t, p, ctx, "infer", nil)
	if j.State != "succeeded" {
		t.Fatal(j)
	}
	r := result(t, s)
	direct := neural.NewDualModeNetwork(784, 128, 10)
	if e := direct.LoadWeights(filepath.Join("../../../module3-mnist/data", "pretrained_weights.json")); e != nil {
		t.Fatal(e)
	}
	direct.SetNoiseLevel(0)
	want := direct.Infer(neural.PreprocessDigit(input))
	if !reflect.DeepEqual(r.Inference, want) {
		t.Fatal("paired inference differs from direct shared/neural execution")
	}
	if r.ModelSHA == "" || r.InputSHA == "" || r.Source != "drawing" || r.Preprocessing != "drawing/1" || len(r.Cases) != 0 {
		t.Fatalf("unlabelled identity/accuracy: %+v", r)
	}
	s, j = finish(t, p, ctx, "infer", nil)
	if !reflect.DeepEqual(result(t, s).Inference, want) {
		t.Fatal("drawing preprocessed twice")
	}
	apply(t, p, ctx, "input", map[string]string{"pixels": string(raw), "source": "input"})
	s, _ = finish(t, p, ctx, "infer", nil)
	if !reflect.DeepEqual(result(t, s).Inference, direct.Infer(input)) {
		t.Fatal("non-drawing input was preprocessed")
	}
	// Malformed pixels must not be coerced to valid zeros or change a prior result.
	before, _ := json.Marshal(s)
	short, _ := json.Marshal(input[:783])
	long, _ := json.Marshal(append(append([]float64{}, input...), 0))
	for _, bad := range []string{string(short), string(long), strings.Replace(string(raw), "[0,", "[null,", 1), strings.Replace(string(raw), "[0,", "[-0.1,", 1), strings.Replace(string(raw), "[0,", "[1.1,", 1)} {
		if _, err := p.Apply(ctx, moduleapi.Command{Revision: s.Revision, Action: "input", Input: map[string]string{"source": "input", "pixels": bad}}); err == nil {
			t.Fatal("malformed input admitted")
		}
		after, _ := p.Snapshot(ctx)
		b, _ := json.Marshal(after)
		if string(b) != string(before) {
			t.Fatal("rejected input changed retained state")
		}
	}
	apply(t, p, ctx, "load_model", nil, moduleapi.ResourceReference{ID: "pretrained_weights_8.json", Kind: "model"})
	apply(t, p, ctx, "configure", map[string]string{"quantization": "model-qat"})
	s, _ = finish(t, p, ctx, "infer", nil)
	if e := direct.LoadWeights("../../../module3-mnist/data/pretrained_weights_8.json"); e != nil {
		t.Fatal(e)
	}
	direct.SetPerLayerQuant(true)
	q := result(t, s)
	if !reflect.DeepEqual(q.Inference, direct.Infer(input)) || !reflect.DeepEqual(q.ResolvedConfig, *direct.Config) {
		t.Fatal("declared model quantization identity differs from direct execution")
	}
}
