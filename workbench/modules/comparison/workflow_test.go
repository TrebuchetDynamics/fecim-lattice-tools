package comparison

import (
	"context"

	"fecim-lattice-tools/workbench/moduleapi"
	"strings"
	"testing"
)

func TestComparisonWebWorkflow(t *testing.T) {
	ctx := context.Background()
	p := New()
	defer p.Close()
	s, err := p.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, workload := range []string{"MNIST", "ResNet-50", "BERT-Base", "GPT-2", "LLM-70B"} {
		for _, profile := range []string{"conservative", "baseline", "optimistic"} {
			s, err = p.Apply(ctx, moduleapi.Command{Action: "calculate", Revision: s.Revision, Input: map[string]string{"workload": workload, "profile": profile, "rate": "2", "mode": "Technical Review"}})
			if err != nil {
				t.Fatal(err)
			}
			co2 := false
			for _, table := range s.Tables {
				for _, row := range table.Rows {
					if len(row) > 1 && row[0].Presentation.Display == "CO2 reduction" {
						co2 = row[1].Value == nil && row[1].Presentation.Availability == "unavailable" && strings.Contains(row[1].Presentation.Reason, "physical energy")
					}
				}
			}
			if !co2 {
				t.Fatal("CO2 must be unavailable with a physical-energy reason")
			}
		}
	}
	for _, action := range []string{"diff", "sensitivity", "presentation", "reset"} {
		s, err = p.Apply(ctx, moduleapi.Command{Action: action, Revision: s.Revision})
		if err != nil {
			t.Fatalf("%s: %v", action, err)
		}
		if action == "diff" || action == "sensitivity" {
			found := false
			for _, table := range s.Tables {
				if table.ID == action && len(table.Rows) > 0 {
					found = true
				}
			}
			if !found {
				t.Fatalf("%s has no actual results", action)
			}
		}
		if action == "presentation" {
			for _, table := range s.Tables {
				if table.ID == "sources" {
					t.Error("presentation still shows raw assumptions table")
				}
			}
		}
	}
	for _, input := range []map[string]string{{"rate": "NaN"}, {"rate": "0"}, {"workload": "missing"}, {"profile": "unknown"}} {
		if _, err = p.Apply(ctx, moduleapi.Command{Action: "calculate", Revision: s.Revision, Input: input}); err == nil {
			t.Fatalf("accepted invalid input %v", input)
		}
	}
	if _, err = p.Apply(ctx, moduleapi.Command{Action: "calculate", Revision: 0}); err == nil {
		t.Fatal("accepted stale revision")
	}
	if _, err = p.Apply(ctx, moduleapi.Command{Action: "export", Revision: s.Revision, Input: map[string]string{"format": "json"}}); err == nil {
		t.Fatal("hosted export acquired filesystem capability")
	}
}
