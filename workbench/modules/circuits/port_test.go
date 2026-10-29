package circuits

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"fecim-lattice-tools/workbench/moduleapi"
)

func TestCircuitsTransactions(t *testing.T) {
	t.Run("native dense admission", func(t *testing.T) {
		c := DefaultConfig()
		c.Rows, c.Cols = 128, 128
		if err := c.Validate(128); err == nil {
			t.Fatal("dense TierA must not admit a 16 GiB network")
		}
		c.Coupling = "TierB"
		if err := c.Validate(128); err != nil {
			t.Fatal(err)
		}
	})
	ctx := context.Background()
	p := New(8)
	defer p.Close()
	before, _ := p.Snapshot(ctx)
	for _, input := range []map[string]string{{"rows": "16"}, {"rows": "3"}, {"cols": "0"}, {"dac_bits": "3"}, {"adc_bits": "9"}, {"node": "22"}, {"gain": "NaN"}, {"sense_min": "1", "sense_max": "0"}, {"max_iterations": "4001"}, {"max_pulses": "41"}, {"material": "missing"}, {"input": "[1]"}, {"active_rows": "[true]"}, {"active_rows": "[false,false,false,false,false,false,false,false]"}, {"read_mask": "[[true]]"}, {"rows": "2", "cols": "2", "gain": "-1"}} {
		if _, err := p.Apply(ctx, moduleapi.Command{Action: "configure", Revision: before.Revision, Input: input}); err == nil {
			t.Fatalf("invalid configuration admitted: %v", input)
		}
		after, _ := p.Snapshot(ctx)
		if !reflect.DeepEqual(before, after) {
			t.Fatalf("rejected configuration mutated state: %v", input)
		}
	}
	job, err := p.Start(ctx, moduleapi.Command{Action: "write", Revision: before.Revision, Input: map[string]string{"target": "29"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.Cancel(ctx, job.ID); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(3 * time.Second)
	for {
		job, err = p.Job(ctx, job.ID)
		if err != nil {
			t.Fatal(err)
		}
		if job.State != "running" && job.State != "cancelling" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("cancellation did not drain")
		}
		time.Sleep(time.Millisecond)
	}
	if job.State != "cancelled" {
		t.Fatalf("expected cancelled transaction: %+v", job)
	}
	after, _ := p.Snapshot(ctx)
	for i, table := range before.Tables {
		if table.ID == "cells" && !reflect.DeepEqual(table, after.Tables[i]) {
			t.Fatal("cancel committed partial device state")
		}
	}
	if p.Canonical() != "" {
		t.Fatal("cancel published successful result bytes")
	}
}

func TestCircuitsPeripheralControls(t *testing.T) {
	p := New(8)
	defer p.Close()
	ctx := context.Background()
	s, _ := p.Snapshot(ctx)
	s, err := p.Apply(ctx, moduleapi.Command{Action: "configure", Revision: s.Revision, Input: map[string]string{"rows": "2", "cols": "2", "architecture": "1T1R", "sense_preset": "High Sensitivity", "voltages": "[0.1,-0.2]"}})
	if err != nil {
		t.Fatal(err)
	}
	s, err = p.Apply(ctx, moduleapi.Command{Action: "compute", Revision: s.Revision})
	if err != nil {
		t.Fatal(err)
	}
	var result Result
	if err = json.Unmarshal([]byte(p.Canonical()), &result); err != nil {
		t.Fatal(err)
	}
	if result.Config.Gain != 50000 || result.Device.CellVoltages[0][0] < 0.09 || result.Device.CellVoltages[0][1] >= 0 {
		t.Fatalf("manual signed DAC/sense preset did not reach the native device path: %+v", result)
	}
	found := false
	for _, table := range s.Tables {
		if table.ID == "columns" {
			found = true
			if len(table.Rows) != 2 {
				t.Fatal("missing full column readouts")
			}
		}
	}
	if !found {
		t.Fatal("missing actual column current/stress table")
	}
	snr := func(s moduleapi.Snapshot) float64 {
		for _, table := range s.Tables {
			if table.ID == "sense" && table.Rows[0][5].Value != nil {
				return *table.Rows[0][5].Value
			}
		}
		t.Fatal("missing finite current-referred SNR")
		return 0
	}
	beforeSNR := snr(s)
	s, err = p.Apply(ctx, moduleapi.Command{Action: "configure", Revision: s.Revision, Input: map[string]string{"adc_bits": "8", "sense_min": "0.2", "sense_max": "0.8"}})
	if err != nil {
		t.Fatal(err)
	}
	if snr(s) != beforeSNR {
		t.Fatal("retained SNR silently recomputed before explicit operation")
	}
	s, err = p.Apply(ctx, moduleapi.Command{Action: "compute", Revision: s.Revision})
	if err != nil {
		t.Fatal(err)
	}
	if snr(s) == beforeSNR {
		t.Fatal("ADC resolution and reference changes did not reach SNR composition")
	}
	before := p.Canonical()
	if _, err = p.Apply(ctx, moduleapi.Command{Action: "configure", Revision: s.Revision, Input: map[string]string{"voltages": "[99,0]"}}); err == nil {
		t.Fatal("unsafe DAC voltage admitted")
	}
	if p.Canonical() != before {
		t.Fatal("invalid peripheral settings lost result")
	}
}

func TestCircuitsRetainedPages(t *testing.T) {
	p := New(8)
	defer p.Close()
	ctx := context.Background()
	snap, _ := p.Snapshot(ctx)
	apply := func(action string, input map[string]string) {
		t.Helper()
		var err error
		snap, err = p.Apply(ctx, moduleapi.Command{Action: action, Input: input, Revision: snap.Revision})
		if err != nil {
			t.Fatal(err)
		}
	}
	apply("compute", nil)
	apply("configure", map[string]string{"rows": "1", "cols": "1"})
	apply("view", map[string]string{"scope": "result", "page": "2"})
	for _, table := range snap.Tables {
		if table.ID == "cell_results" {
			if table.Page != 2 || len(table.Rows) != 32 {
				t.Fatalf("retained result page unreachable: %+v", table)
			}
			return
		}
	}
	t.Fatal("retained result missing")
}

func TestCircuitsMaskedSignedState(t *testing.T) {
	ctx := context.Background()
	c := DefaultConfig()
	c.Rows, c.Cols = 2, 2
	c.Architecture = "2T1R"
	c.Coupling = "Ideal"
	s := Initial(c, false)
	s.Input = []int{255, -128}
	s.Weights = [][]int{{20, 10}, {5, 15}}
	s.ActiveRows = []bool{true, false}
	s.ReadMask = [][]bool{{true, false}, {true, true}}
	result, err := Execute(ctx, c, s, "compute", 0, 8, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Device.CellCurrents[0][1] != 0 || result.Device.RowCurrents[1] != 0 {
		t.Fatal("masked cells or inactive rows conduct")
	}
	if result.Device.CellCurrents[0][0] <= 0 {
		t.Fatal("unmasked cell did not conduct")
	}
	s.ReadMask = nil
	s.ActiveRows[1] = true
	result, err = Execute(ctx, c, s, "compute", 0, 8, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Device.CellCurrents[0][1] >= 0 {
		t.Fatal("signed input lost negative cell current")
	}
}
