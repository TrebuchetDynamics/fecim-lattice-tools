package mnist

import (
	"encoding/json"
	"fmt"

	"fecim-lattice-tools/shared/neural"
	"fecim-lattice-tools/workbench/experiment"
	"fecim-lattice-tools/workbench/moduleapi"
	"fecim-lattice-tools/workbench/presentation"
)

func cell(v float64, unit string) moduleapi.NumericCell {
	return moduleapi.NumericCell{Value: &v, Unit: unit, Evidence: Warning, Presentation: presentation.FormatMetric(experiment.Metric{Value: v, Unit: unit}, "")}
}
func field(id, label, value string) moduleapi.Field {
	return moduleapi.Field{ID: id, Label: label, Input: value, Presentation: presentation.Value{Display: value, Speakable: value, Availability: "available"}}
}
func jsonText(v any) string { b, _ := json.Marshal(v); return string(b) }
func weightMatrix(n *neural.DualModeNetwork, layer int) ([][]float64, [][]float64) {
	switch layer {
	case 2:
		return n.FPWeights2, n.QuantWeights2
	case 3:
		return n.SingleLayerWeights, n.QuantSingleLayerWeights
	default:
		return n.FPWeights1, n.QuantWeights1
	}
}
func (p *Port) snapshot() moduleapi.Snapshot {
	s := moduleapi.Snapshot{Revision: p.revision, Descriptor: Descriptor(), Status: "ready: select or create an input", Evidence: []string{Warning, "CPU only; models and data are never fetched automatically. Dataset samples are not drawing-preprocessed. QAT metadata describes the file; PTQ changes quantization, not training."}, Fields: []moduleapi.Field{field("config", "Current configuration", jsonText(p.cfg)), field("result", "Retained complete result", jsonText(p.result)), field("job_id", "Job identifier", p.job.ID), field("job_state", "Job state", p.job.State), field("source", "Current input source", p.source)}, Tables: []moduleapi.TablePage{}, Plots: []moduleapi.Plot{}}
	if p.model == nil {
		s.Status = "unavailable: select a model"
	} else {
		s.Fields = append(s.Fields, field("model_sha256", "Selected model SHA-256", p.model.SHA256), field("hidden", "Hidden neurons", fmt.Sprint(p.model.Hidden())))
	}
	if p.dataset != nil {
		s.Fields = append(s.Fields, field("dataset_sha256", "Selected corpus SHA-256", p.dataset.SHA256), field("corpus_count", "Actual selected corpus images", fmt.Sprint(len(p.dataset.Images))))
	}
	if p.result != nil {
		if p.resultEpoch != p.epoch {
			s.Evidence = append(s.Evidence, "STALE: retained results describe earlier model, data, input or configuration; rerun explicitly.")
		}
		s.Fields = append(s.Fields, field("result_fresh", "Result matches current inputs", fmt.Sprint(p.resultEpoch == p.epoch)))
	}
	if p.active() {
		s.Status = p.job.State
	} else if p.job.State != "idle" {
		s.Status = p.job.State + ": " + p.job.Message
	}
	for _, a := range []struct {
		id        string
		available bool
		reason    string
	}{{"load_model", true, ""}, {"load_dataset", true, ""}, {"configure", true, ""}, {"input", true, ""}, {"clear", true, ""}, {"sample", p.dataset != nil, "Select an actual labelled corpus"}, {"infer", p.model != nil && len(p.input) == 784, "Select model and input; unlabelled inputs have no accuracy"}, {"evaluate", p.model != nil && p.dataset != nil, "Select model and labelled corpus"}, {"sweep", p.model != nil && p.dataset != nil, "Explicit subset of at most 500 images and 64 cases required"}} {
		s.Actions = append(s.Actions, moduleapi.Action{ID: a.id, Label: a.id, Available: a.available && !p.active(), Reason: a.reason})
	}
	if p.inspection != nil {
		w, q := weightMatrix(p.inspection, p.layer)
		total := (len(w)*len(w[0]) + 31) / 32
		t := moduleapi.TablePage{ID: "weights", Caption: fmt.Sprintf("Current layer %d FP and CIM weights — page %d of %d", p.layer, p.page, total), Page: p.page, Total: total, Columns: []string{"Neuron", "Input", "FP weight", "CIM weight"}, Rows: [][]moduleapi.NumericCell{}}
		for k := (p.page - 1) * 32; k < min(p.page*32, len(w)*len(w[0])); k++ {
			i, j := k/len(w[0]), k%len(w[0])
			t.Rows = append(t.Rows, []moduleapi.NumericCell{cell(float64(i), "count"), cell(float64(j), "count"), cell(w[i][j], "ratio"), cell(q[i][j], "ratio")})
		}
		s.Tables = append(s.Tables, t)
		a, b := p.inspection.GetQuantizationStats()
		stats := moduleapi.TablePage{ID: "quantization", Caption: "Current layer quantization statistics", Page: 1, Total: 1, Columns: []string{"Layer", "Mean squared error", "Maximum error"}, Rows: [][]moduleapi.NumericCell{}}
		for i, v := range []neural.QuantizationStats{a, b} {
			stats.Rows = append(stats.Rows, []moduleapi.NumericCell{cell(float64(i+1), "count"), cell(v.MSE, "ratio"), cell(v.MaxError, "ratio")})
		}
		s.Tables = append(s.Tables, stats)
	}
	if p.result == nil {
		s.Evidence = append(s.Evidence, "Accuracy unavailable: no complete labelled evaluation. No fixed or synthetic accuracy baseline.")
		return s
	}
	r := p.result
	if r.Inference != nil {
		i := r.Inference
		prediction := moduleapi.TablePage{ID: "prediction", Caption: "Retained paired predictions (not dataset accuracy)", Page: 1, Total: 1, Columns: []string{"FP prediction", "CIM prediction", "FP confidence", "CIM confidence"}, Rows: [][]moduleapi.NumericCell{{cell(float64(i.FPPrediction), "count"), cell(float64(i.CIMPrediction), "count"), cell(i.FPConfidence, "ratio"), cell(i.CIMConfidence, "ratio")}}}
		s.Tables = append(s.Tables, prediction)
		prob := moduleapi.TablePage{ID: "probabilities", Caption: "Retained paired probabilities and logits", Page: 1, Total: 1, Columns: []string{"Digit", "FP probability", "CIM probability", "FP logit", "CIM logit"}, Rows: [][]moduleapi.NumericCell{}}
		for k := 0; k < 10; k++ {
			prob.Rows = append(prob.Rows, []moduleapi.NumericCell{cell(float64(k), "count"), cell(i.FPProbabilities[k], "ratio"), cell(i.CIMProbabilities[k], "ratio"), cell(i.FPLogits[k], "ratio"), cell(i.CIMLogits[k], "ratio")})
		}
		s.Tables = append(s.Tables, prob)
		layers := moduleapi.TablePage{ID: "layers", Caption: "Retained hidden-layer activations (empty for calibration architecture)", Page: 1, Total: 1, Columns: []string{"Neuron", "FP activation", "CIM activation"}, Rows: [][]moduleapi.NumericCell{}}
		if r.Config.Architecture != "784-10" {
			for k, v := range i.FPHidden {
				layers.Rows = append(layers.Rows, []moduleapi.NumericCell{cell(float64(k), "count"), cell(v, "ratio"), cell(i.CIMHidden[k], "ratio")})
			}
		}
		s.Tables = append(s.Tables, layers)
	}
	if len(r.Cases) == 0 {
		s.Evidence = append(s.Evidence, "Accuracy unavailable: this input is not a labelled evaluation.")
		return s
	}
	scores := moduleapi.TablePage{ID: "evaluation", Caption: "Retained actual labelled subset evaluation; counts are not silently reduced", Page: 1, Total: 1, Columns: []string{"Case", "Value", "Samples", "FP correct", "CIM correct", "FP accuracy", "CIM accuracy"}, Rows: [][]moduleapi.NumericCell{}}
	for k, v := range r.Cases {
		scores.Rows = append(scores.Rows, []moduleapi.NumericCell{cell(float64(k), "count"), cell(v.Value, "ratio"), cell(float64(v.Total), "samples"), cell(float64(v.FPCorrect), "count"), cell(float64(v.CIMCorrect), "count"), cell(float64(v.FPCorrect)/float64(v.Total), "ratio"), cell(float64(v.CIMCorrect)/float64(v.Total), "ratio")})
	}
	s.Tables = append(s.Tables, scores)
	c := r.Cases[p.caseIndex]
	conf := moduleapi.TablePage{ID: "confusion", Caption: fmt.Sprintf("Retained case %d confusion — actual label, predicted label, FP and CIM counts", p.caseIndex), Page: 1, Total: 1, Columns: []string{"Actual", "Predicted", "FP count", "CIM count"}, Rows: [][]moduleapi.NumericCell{}}
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			conf.Rows = append(conf.Rows, []moduleapi.NumericCell{cell(float64(i), "count"), cell(float64(j), "count"), cell(float64(c.FPConfusion[i][j]), "count"), cell(float64(c.CIMConfusion[i][j]), "count")})
		}
	}
	s.Tables = append(s.Tables, conf)
	return s
}
