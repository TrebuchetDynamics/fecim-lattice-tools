//go:build js && wasm

// A validation-only entry point, not a public application API. Input is a bundle
// already resolved by the real native Go loader in the parity test. There is no
// duplicate browser YAML/schema loader, persistence adapter or TypeScript model.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"syscall/js"

	"fecim-lattice-tools/workbench/experiment"
	"fecim-lattice-tools/workbench/fecim"
	"fecim-lattice-tools/workbench/project"
	"fecim-lattice-tools/workbench/report"
)

var runCallback js.Func

func main() {
	runCallback = js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) != 1 || args[0].Type() != js.TypeString {
			return response("", fmt.Errorf("one resolved input JSON string is required"))
		}
		results, err := run([]byte(args[0].String()))
		return response(results, err)
	})
	js.Global().Set("fecimParityRun", runCallback)
	select {}
}

func response(results string, err error) string {
	out := struct {
		ResultsJSON string `json:"resultsJSON"`
		Error       string `json:"error,omitempty"`
	}{ResultsJSON: results}
	if err != nil {
		out.Error = err.Error()
	}
	data, _ := json.Marshal(out)
	return string(data)
}

func run(input []byte) (string, error) {
	var request struct {
		Bundle   project.Bundle `json:"bundle"`
		Revision string         `json:"revision"`
	}
	if err := json.Unmarshal(input, &request); err != nil {
		return "", err
	}
	points, err := experiment.Expand(request.Bundle)
	if err != nil {
		return "", err
	}
	// This entry point is restricted to the bounded validation example, not an
	// unbounded hosted solver. Production admission/cancellation is separate work.
	if len(points) > 256 {
		return "", fmt.Errorf("parity fixture exceeds 256 points")
	}
	runs := make([]experiment.RunRecord, 0, len(points))
	for _, point := range points {
		id, err := experiment.ID(point, fecim.EvaluatorVersion, request.Bundle.Inputs)
		if err != nil {
			return "", err
		}
		result, err := fecim.Evaluate(context.Background(), point.Design, point.Seed)
		if err != nil {
			return "", err
		}
		runs = append(runs, experiment.RunRecord{
			Manifest: experiment.RunManifest{
				SchemaVersion: 2, RunID: id, PointIndex: point.Index,
				Seed: point.Seed, ModelVersion: point.ModelVersion,
				EvaluatorVersion:   fecim.EvaluatorVersion,
				RepositoryRevision: request.Revision, Inputs: request.Bundle.Inputs,
				Status: result.Status,
			},
			Design: point.Design, Result: result,
		})
	}
	analysis := experiment.Analyze(request.Bundle, runs)
	var output bytes.Buffer
	if err := report.WriteJSON(&output, request.Bundle, analysis); err != nil {
		return "", err
	}
	return output.String(), nil
}
