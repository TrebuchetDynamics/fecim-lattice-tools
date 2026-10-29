package circuits

import (
	"context"
	"testing"

	"fecim-lattice-tools/workbench/moduleapi"
)

func TestCircuitsWebWorkflow(t *testing.T) {
	provider, ok := any(New()).(moduleapi.Provider)
	if !ok {
		t.Fatal("circuits has no real persistent web port; synthetic compute is not device parity")
	}
	p := provider.WebPort()
	defer p.Close()
	ctx := context.Background()
	s, err := p.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	apply := func(action string, input map[string]string) {
		t.Helper()
		s, err = p.Apply(ctx, moduleapi.Command{Action: action, Input: input, Revision: s.Revision})
		if err != nil {
			t.Fatalf("%s: %v", action, err)
		}
	}
	apply("configure", map[string]string{"rows": "2", "cols": "2", "architecture": "1T1R", "coupling": "Ideal", "input": "[255,-128]"})
	apply("select", map[string]string{"row": "0", "col": "0"})
	apply("write", map[string]string{"target": "20"})
	apply("read", nil)
	apply("compute", nil)
	var found bool
	for _, table := range s.Tables {
		if table.ID == "cells" && len(table.Rows) > 0 {
			found = true
			if table.Rows[0][2].Value == nil || *table.Rows[0][2].Value == 0 {
				t.Fatal("WRITE did not persist a programmed cell")
			}
		}
	}
	if !found {
		t.Fatal("missing real cell values")
	}
	apply("undo", nil)
	apply("reset", nil)
}
