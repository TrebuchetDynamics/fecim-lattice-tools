package mnisttrainsingle

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"fecim-lattice-tools/module3-mnist/pkg/training"
)

func TestAppendSingleLayerToMainWeights(t *testing.T) {
	net, err := training.NewSingleLayerNetwork()
	if err != nil {
		t.Fatalf("new network: %v", err)
	}
	d := t.TempDir()
	p := filepath.Join(d, "weights.json")
	if err := os.WriteFile(p, []byte(`{"existing":1}`), 0o644); err != nil {
		t.Fatalf("write seed: %v", err)
	}
	if err := appendSingleLayerToMainWeights(p, net); err != nil {
		t.Fatalf("appendSingleLayerToMainWeights: %v", err)
	}
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read out: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal out: %v", err)
	}
	if _, ok := m["single_layer_weights"]; !ok {
		t.Fatal("missing single_layer_weights")
	}
	if _, ok := m["single_layer_bias"]; !ok {
		t.Fatal("missing single_layer_bias")
	}
}
