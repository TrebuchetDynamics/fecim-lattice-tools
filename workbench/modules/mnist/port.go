package mnist

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync"

	data "fecim-lattice-tools/module3-mnist/pkg/mnist"
	"fecim-lattice-tools/shared/neural"
	"fecim-lattice-tools/workbench/moduleapi"
	"fecim-lattice-tools/workbench/moduleapi/native"
)

func Descriptor() moduleapi.ModuleDescriptor {
	return moduleapi.ModuleDescriptor{ID: "mnist", Title: "MNIST", Available: true, SchemaVersion: moduleapi.SchemaVersion, ModelVersion: "mnist-native/1", Features: []string{"load_model", "load_dataset", "input", "sample", "configure", "infer", "evaluate", "sweep", "view"}, Capabilities: []moduleapi.Capability{{Profile: "local", Available: true, Reason: "Root-confined CPU inference and labelled evaluation"}, {Profile: "hosted", Available: true, Reason: "Help and unverified read-only result import only; inference requires local mode"}}}
}

type Port struct {
	mu                            sync.Mutex
	closed                        bool
	revision, epoch, resultEpoch  uint64
	cfg                           Config
	model                         *neural.MNISTModel
	inspection                    *neural.DualModeNetwork
	dataset                       *data.LocalDataset
	input                         []float64
	rawSHA, source, preprocessing string
	sample                        int
	result                        *Result
	page, layer, caseIndex        int
	job                           moduleapi.Job
	cancel                        context.CancelFunc
	done                          chan struct{}
}

func New() *Port {
	return &Port{cfg: Defaults(), sample: -1, page: 1, layer: 1, job: moduleapi.Job{State: "idle", ResultsState: "unavailable", Counts: map[string]int{}}}
}
func (p *Port) Descriptor() moduleapi.ModuleDescriptor { return Descriptor() }
func (p *Port) active() bool                           { return p.job.State == "running" || p.job.State == "cancelling" }
func (p *Port) Snapshot(ctx context.Context) (moduleapi.Snapshot, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if e := ctx.Err(); e != nil {
		return moduleapi.Snapshot{}, e
	}
	if p.closed {
		return moduleapi.Snapshot{}, errors.New("MNIST session closed")
	}
	return p.snapshot(), nil
}
func (p *Port) read(ctx context.Context, ref moduleapi.ResourceReference, kind string) (io.ReadCloser, error) {
	r, ok := native.Resources(ctx)
	if !ok || ref.Kind != kind {
		return nil, errors.New("explicit rooted local resource capability required")
	}
	reader, info, e := r.Open(ctx, ref)
	if e != nil {
		return nil, e
	}
	if info.Size < 0 || info.Size > 32<<20 {
		reader.Close()
		return nil, errors.New("resource exceeds MNIST 32 MiB input limit")
	}
	return reader, nil
}
func (p *Port) Apply(ctx context.Context, c moduleapi.Command) (moduleapi.Snapshot, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	fail := func(e error) (moduleapi.Snapshot, error) { return moduleapi.Snapshot{}, e }
	if e := ctx.Err(); e != nil {
		return fail(e)
	}
	if p.closed || p.active() || c.Revision != p.revision {
		return fail(errors.New("closed, busy or stale MNIST revision"))
	}
	changed := true
	switch c.Action {
	case "load_model":
		if len(c.Resources) != 1 {
			return fail(errors.New("one model reference required"))
		}
		reader, e := p.read(ctx, c.Resources[0], "model")
		if e != nil {
			return fail(e)
		}
		m, e := neural.DecodeMNISTModel(reader)
		reader.Close()
		if e != nil {
			return fail(e)
		}
		if e = p.cfg.validate(m); e != nil {
			return fail(e)
		}
		if e = AdmitWork(m.Hidden(), 1, 1, false); e != nil {
			return fail(e)
		}
		if e = ctx.Err(); e != nil {
			return fail(e)
		}
		n, e := prepareInspection(m, p.cfg)
		if e != nil {
			return fail(e)
		}
		p.model, p.inspection = m, n
		p.page, p.layer = 1, 1
	case "load_dataset":
		// Ordered references: IDX images first, IDX labels second. Roles never guessed.
		if len(c.Resources) != 2 {
			return fail(errors.New("ordered image and label dataset references required"))
		}
		a, e := p.read(ctx, c.Resources[0], "dataset")
		if e != nil {
			return fail(e)
		}
		defer a.Close()
		b, e := p.read(ctx, c.Resources[1], "dataset")
		if e != nil {
			return fail(e)
		}
		defer b.Close()
		d, e := data.DecodeLocalDataset(a, b)
		if e != nil {
			return fail(e)
		}
		if e = ctx.Err(); e != nil {
			return fail(e)
		}
		p.dataset = d
		if p.source == "dataset" {
			p.input = nil
			p.source = ""
			p.sample = -1
		}
	case "configure":
		cfg, e := configure(p.cfg, c.Input)
		if e != nil {
			return fail(e)
		}
		if e = cfg.validate(p.model); e != nil {
			return fail(e)
		}
		var n *neural.DualModeNetwork
		if p.model != nil {
			n, e = prepareInspection(p.model, cfg)
			if e != nil {
				return fail(e)
			}
		}
		p.cfg, p.inspection = cfg, n
	case "input":
		if len(c.Input["pixels"]) > 64<<10 {
			return fail(errors.New("pixel payload exceeds 64 KiB"))
		}
		var encoded []*float64
		if e := json.Unmarshal([]byte(c.Input["pixels"]), &encoded); e != nil {
			return fail(e)
		}
		if len(encoded) != 784 {
			return fail(errors.New("exactly 784 pixels required"))
		}
		pixels := make([]float64, 784)
		for i, v := range encoded {
			if v == nil || !finite(*v) || *v < 0 || *v > 1 {
				return fail(errors.New("pixels must be finite numbers in [0,1]"))
			}
			pixels[i] = *v
		}
		source := c.Input["source"]
		if source != "drawing" && source != "input" {
			return fail(errors.New("source must be drawing or unlabelled input"))
		}
		raw := digest(pixels)
		pre := "none"
		if source == "drawing" {
			pixels = neural.PreprocessDigit(pixels)
			pre = "drawing/1"
		}
		p.input, p.rawSHA, p.source, p.preprocessing, p.sample = pixels, raw, source, pre, -1
	case "clear":
		p.input = nil
		p.source = ""
		p.rawSHA = ""
		p.preprocessing = ""
		p.sample = -1
	case "sample":
		index, e := strconv.Atoi(c.Input["index"])
		if e != nil || p.dataset == nil || index < 0 || index >= len(p.dataset.Images) {
			return fail(errors.New("sample outside selected labelled corpus"))
		}
		p.input = p.dataset.Images[index]
		p.rawSHA = digest(p.input)
		p.source = "dataset"
		p.preprocessing = "none"
		p.sample = index
	case "view":
		page, layer, index := p.page, p.layer, p.caseIndex
		if _, ok := c.Input["layer"]; ok {
			page = 1
		}
		var e error
		for k, v := range c.Input {
			switch k {
			case "page":
				page, e = strconv.Atoi(v)
			case "layer":
				layer, e = strconv.Atoi(v)
			case "case":
				index, e = strconv.Atoi(v)
			default:
				e = errors.New("unknown inspection setting")
			}
			if e != nil {
				return fail(e)
			}
		}
		if layer < 1 || layer > 3 || p.inspection == nil || (layer == 3 && !p.model.HasCalibration()) {
			return fail(errors.New("weight layer unavailable"))
		}
		w, _ := weightMatrix(p.inspection, layer)
		if page < 1 || page > (len(w)*len(w[0])+31)/32 || index < 0 || (p.result != nil && len(p.result.Cases) > 0 && index >= len(p.result.Cases)) {
			return fail(errors.New("inspection page/case outside result"))
		}
		p.page, p.layer, p.caseIndex = page, layer, index
		changed = false
	default:
		return fail(errors.New("unsupported MNIST action"))
	}
	if changed {
		p.epoch++
	}
	p.revision++
	return p.snapshot(), nil
}
func selection(c moduleapi.Command, n int, sweep bool) (int, int, []float64, error) {
	start, count := 0, n
	var e error
	if v, ok := c.Input["start"]; ok {
		start, e = strconv.Atoi(v)
		if e != nil {
			return 0, 0, nil, e
		}
	}
	if v, ok := c.Input["count"]; ok {
		count, e = strconv.Atoi(v)
		if e != nil {
			return 0, 0, nil, e
		}
	}
	if start < 0 || count < 1 || start > n || count > n-start {
		return 0, 0, nil, errors.New("explicit subset exceeds selected corpus; no sample reduction")
	}
	values := []float64{0}
	if sweep {
		if len(c.Input["values"]) > 4096 {
			return 0, 0, nil, errors.New("sweep payload too large")
		}
		if e = json.Unmarshal([]byte(c.Input["values"]), &values); e != nil {
			return 0, 0, nil, e
		}
		for _, v := range values {
			if !finite(v) {
				return 0, 0, nil, errors.New("non-finite sweep")
			}
			switch c.Input["kind"] {
			case "levels":
				if v != float64(int(v)) || v < 2 || v > neural.MaxDemoLevels {
					return 0, 0, nil, errors.New("invalid sweep levels")
				}
			case "adc":
				if v != float64(int(v)) || v < 3 || v > 16 {
					return 0, 0, nil, errors.New("invalid ADC sweep")
				}
			case "noise":
				if v < 0 || v > .2 {
					return 0, 0, nil, errors.New("invalid noise sweep")
				}
			default:
				return 0, 0, nil, errors.New("unknown sweep kind")
			}
		}
	}
	return start, count, values, nil
}
func (p *Port) Start(ctx context.Context, c moduleapi.Command) (moduleapi.Job, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	bad := func(e error) (moduleapi.Job, error) { return moduleapi.Job{}, e }
	if e := ctx.Err(); e != nil {
		return bad(e)
	}
	if p.closed || p.active() || c.Revision != p.revision {
		return bad(errors.New("closed, busy or stale MNIST job"))
	}
	if p.model == nil {
		return bad(errors.New("unavailable: select a model"))
	}
	if c.Action != "infer" && c.Action != "evaluate" && c.Action != "sweep" {
		return bad(errors.New("unsupported MNIST job"))
	}
	start, count, values := p.sample, 1, []float64{0}
	var e error
	if c.Action == "infer" {
		if len(p.input) != 784 {
			return bad(errors.New("unavailable: select or create an input"))
		}
	} else {
		if p.dataset == nil {
			return bad(errors.New("unavailable: select a labelled dataset"))
		}
		start, count, values, e = selection(c, len(p.dataset.Images), c.Action == "sweep")
		if e != nil {
			return bad(e)
		}
	}
	if e = AdmitWork(p.model.Hidden(), count, len(values), c.Action == "sweep"); e != nil {
		return bad(e)
	}
	cfg, model, dataset, input, epoch := p.cfg, p.model, p.dataset, p.input, p.epoch
	result := Result{Version: "mnist-native/1", ModelSHA: model.SHA256, Config: cfg, Source: p.source, Preprocessing: p.preprocessing, Start: start, Count: count, InputSHA: digest(input), RawInputSHA: p.rawSHA}
	a, b, d := model.TrainingLevels()
	result.TrainingLevels = [3]int{a, b, d}
	if dataset != nil {
		result.CorpusCount = len(dataset.Images)
	}
	if c.Action != "infer" || p.source == "dataset" {
		result.DatasetSHA = dataset.SHA256
	}
	if c.Action != "infer" {
		result.Source = "dataset"
		result.Preprocessing = "none"
		result.InputSHA = ""
		result.RawInputSHA = ""
		if c.Action == "sweep" {
			result.Sweep = c.Input["kind"]
		}
	}
	run, cancel := context.WithTimeout(ctx, JobTimeout)
	p.cancel = cancel
	p.done = make(chan struct{})
	done := p.done
	p.revision++
	p.job = moduleapi.Job{ID: fmt.Sprintf("mnist-%d", p.revision), State: "running", InputRevision: strconv.FormatUint(p.revision, 10), ResultsState: "unavailable", Progress: moduleapi.Progress{Total: count * len(values)}, Counts: map[string]int{}, Message: "Running CPU MNIST job; no accuracy claimed until all selected samples finish"}
	go func() {
		defer close(done)
		defer cancel()
		n := model.NewNetwork(cfg.Seed)
		cfg.configure(n, model)
		result.ResolvedConfig = *n.Config
		var err error
		for _, value := range values {
			if err = run.Err(); err != nil {
				break
			}
			cc := cfg
			switch result.Sweep {
			case "levels":
				cc.Levels = int(value)
				cc.Quantization = "uniform-ptq"
			case "adc":
				cc.ADC = int(value)
			case "noise":
				cc.Noise = value
			}
			cc.configure(n, model)
			metrics := Case{Value: value, Config: cc, ResolvedConfig: *n.Config}
			for i := 0; i < count; i++ {
				if err = run.Err(); err != nil {
					break
				}
				pixels := input
				if c.Action != "infer" {
					pixels = dataset.Images[start+i]
				}
				inferred := n.Infer(pixels)
				if inferred == nil || !validInference(inferred) {
					err = errors.New("non-finite or unavailable numerical inference")
					break
				}
				if c.Action == "infer" {
					result.Inference = inferred
				} else {
					label := dataset.Labels[start+i]
					metrics.Total++
					metrics.FPConfusion[label][inferred.FPPrediction]++
					metrics.CIMConfusion[label][inferred.CIMPrediction]++
					if label == inferred.FPPrediction {
						metrics.FPCorrect++
					}
					if label == inferred.CIMPrediction {
						metrics.CIMCorrect++
					}
				}
				p.mu.Lock()
				p.job.Progress.Completed++
				p.job.Counts["fp_correct"] += boolInt(c.Action != "infer" && dataset.Labels[start+i] == inferred.FPPrediction)
				p.job.Counts["cim_correct"] += boolInt(c.Action != "infer" && dataset.Labels[start+i] == inferred.CIMPrediction)
				p.mu.Unlock()
			}
			if err != nil {
				break
			}
			if c.Action != "infer" {
				result.Cases = append(result.Cases, metrics)
			}
		}
		p.mu.Lock()
		defer p.mu.Unlock()
		p.revision++
		if err == nil {
			err = run.Err()
		}
		if err != nil {
			if run.Err() != nil {
				p.job.State = "cancelled"
				p.job.Progress.Cancelled = p.job.Progress.Total - p.job.Progress.Completed
				p.job.Message = "Cancelled or timed out; partial counts only, previous complete result retained"
			} else {
				p.job.State = "failed"
				p.job.Progress.Failed = p.job.Progress.Total - p.job.Progress.Completed
				p.job.Message = err.Error()
			}
			return
		}
		p.result = &result
		p.resultEpoch = epoch
		p.caseIndex = 0
		p.job.State = "succeeded"
		p.job.ResultsState = "complete"
		p.job.ResultsRevision = strconv.FormatUint(p.revision, 10)
		p.job.AnalysisAvailable = true
		p.job.Message = "Complete: actual selected samples; inspect model/data fingerprints and configuration"
	}()
	return cloneJob(p.job), nil
}
func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
func validInference(r *neural.InferenceResult) bool {
	if r.FPPrediction < 0 || r.FPPrediction > 9 || r.CIMPrediction < 0 || r.CIMPrediction > 9 {
		return false
	}
	for _, a := range [][]float64{r.FPProbabilities, r.CIMProbabilities, r.FPLogits, r.CIMLogits, r.FPHidden, r.CIMHidden, {r.FPConfidence, r.CIMConfidence, r.Disagreement, r.EnergyUsed, r.ReadLatencyUS}} {
		for _, v := range a {
			if !finite(v) {
				return false
			}
		}
	}
	return len(r.FPProbabilities) == 10 && len(r.CIMProbabilities) == 10
}
func cloneJob(j moduleapi.Job) moduleapi.Job {
	m := map[string]int{}
	for k, v := range j.Counts {
		m[k] = v
	}
	j.Counts = m
	return j
}
func (p *Port) Job(ctx context.Context, id string) (moduleapi.Job, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if e := ctx.Err(); e != nil {
		return moduleapi.Job{}, e
	}
	if id != p.job.ID {
		return moduleapi.Job{}, errors.New("unknown MNIST job")
	}
	return cloneJob(p.job), nil
}
func (p *Port) Cancel(ctx context.Context, id string) (moduleapi.Job, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if e := ctx.Err(); e != nil {
		return moduleapi.Job{}, e
	}
	if id != p.job.ID {
		return moduleapi.Job{}, errors.New("unknown MNIST job")
	}
	if p.active() {
		p.cancel()
		p.job.State = "cancelling"
		p.job.CancellationRequested = true
	}
	return cloneJob(p.job), nil
}
func (p *Port) Close() {
	p.mu.Lock()
	p.closed = true
	if p.cancel != nil {
		p.cancel()
	}
	done := p.done
	p.mu.Unlock()
	if done != nil {
		<-done
	}
}
