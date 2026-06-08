package simulation

import (
	"math"
	"testing"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
)

func TestModule1SimulationE2EWideEngineMaterialWaveformMatrix(t *testing.T) {
	tests := []struct {
		name      string
		material  *ferroelectric.HZOMaterial
		frequency float64
		amplitude float64
	}{
		{name: "default-hzo", material: ferroelectric.DefaultHZO(), frequency: 1e6, amplitude: 2.4},
		{name: "fecim-hzo", material: ferroelectric.FeCIMMaterial(), frequency: 2.5e6, amplitude: 2.0},
		{name: "superlattice", material: ferroelectric.LiteratureSuperlattice(), frequency: 5e5, amplitude: 1.7},
		{name: "cryogenic", material: ferroelectric.CryogenicHZO(), frequency: 1.25e6, amplitude: 3.0},
	}
	waveforms := []struct {
		name string
		kind WaveformType
	}{
		{name: "sine", kind: WaveformSine},
		{name: "triangle", kind: WaveformTriangle},
		{name: "square", kind: WaveformSquare},
		{name: "manual", kind: WaveformManual},
	}

	for _, tc := range tests {
		for _, wf := range waveforms {
			t.Run(tc.name+"/"+wf.name, func(t *testing.T) {
				engine := NewEngine(tc.material)
				engine.SetFrequency(tc.frequency)
				engine.SetAmplitude(tc.amplitude)
				engine.SetWaveform(wf.kind)
				if wf.kind == WaveformManual {
					engine.SetVoltage(tc.amplitude / 2)
				}
				engine.Start()
				for i := 0; i < 240; i++ {
					engine.Step()
				}
				state := engine.State()
				if !engine.IsRunning() || engine.IsPaused() {
					t.Fatalf("engine state running=%v paused=%v, want running and unpaused", engine.IsRunning(), engine.IsPaused())
				}
				if state.Time <= 0 || len(state.VoltageHistory) != 240 || len(state.PolHistory) != 240 {
					t.Fatalf("state time/history = %.3e/%d/%d, want 240 stepped samples", state.Time, len(state.VoltageHistory), len(state.PolHistory))
				}
				if !isFiniteE2E(state.Voltage) || !isFiniteE2E(state.ElectricField) || !isFiniteE2E(state.Polarization) || !isFiniteE2E(state.NormPol) {
					t.Fatalf("state contains non-finite values: %+v", state)
				}
				if math.Abs(state.NormPol) > 1.25 {
					t.Fatalf("normalized polarization = %.6f, want bounded around physical range", state.NormPol)
				}
				if wf.kind == WaveformManual && math.Abs(state.Voltage-tc.amplitude/2) > 1e-12 {
					t.Fatalf("manual voltage = %.12f, want %.12f", state.Voltage, tc.amplitude/2)
				}
				fields, pol := engine.GetHysteresisData()
				if len(fields) != 401 || len(pol) != 401 {
					t.Fatalf("hysteresis data lengths = %d/%d, want 401/401", len(fields), len(pol))
				}
				if fields[0] >= 0 || fields[len(fields)-1] >= 0 {
					t.Fatalf("hysteresis field endpoints = %.3e %.3e, want closed loop ending negative", fields[0], fields[len(fields)-1])
				}
				engine.Pause()
				paused := engine.State()
				engine.Step()
				if after := engine.State(); after.Time != paused.Time || len(after.VoltageHistory) != len(paused.VoltageHistory) {
					t.Fatalf("paused step mutated state: before %.3e/%d after %.3e/%d", paused.Time, len(paused.VoltageHistory), after.Time, len(after.VoltageHistory))
				}
				engine.Stop()
				if engine.IsRunning() {
					t.Fatal("engine still running after Stop")
				}
			})
		}
	}
}

func TestModule1SimulationE2EHistoryResetAndInvalidInputIsolation(t *testing.T) {
	engine := NewEngine(ferroelectric.DefaultHZO())
	engine.SetWaveform(WaveformSquare)
	engine.SetFrequency(2e6)
	engine.SetAmplitude(2.0)
	engine.Start()
	for i := 0; i < 1250; i++ {
		engine.Step()
	}
	state := engine.State()
	if len(state.VoltageHistory) != state.MaxHistory || len(state.PolHistory) != state.MaxHistory || state.MaxHistory != defaultMaxHistory {
		t.Fatalf("history length/max = %d/%d/%d, want capped at default max %d", len(state.VoltageHistory), len(state.PolHistory), state.MaxHistory, defaultMaxHistory)
	}
	beforeInvalid := engine.State()
	engine.SetFrequency(math.NaN())
	engine.SetAmplitude(math.Inf(1))
	engine.SetWaveform(WaveformType(99))
	engine.SetVoltage(math.NaN())
	engine.Step()
	afterInvalid := engine.State()
	if !isFiniteE2E(afterInvalid.Voltage) || !isFiniteE2E(afterInvalid.Polarization) || afterInvalid.Time <= beforeInvalid.Time {
		t.Fatalf("invalid inputs corrupted or halted engine: before=%+v after=%+v", beforeInvalid, afterInvalid)
	}
	engine.Reset()
	reset := engine.State()
	if reset.Time != 0 || reset.Voltage != 0 || reset.Polarization != 0 || len(reset.VoltageHistory) != 0 || len(reset.PolHistory) != 0 || reset.MaxHistory != defaultMaxHistory {
		t.Fatalf("reset state = %+v, want empty default history and zero state", reset)
	}
}

func TestModule1SimulationE2EInertEngineForInvalidMaterialIsSafe(t *testing.T) {
	bad := ferroelectric.DefaultHZO()
	bad.Thickness = 0
	engine := NewEngine(bad)
	engine.Start()
	for i := 0; i < 10; i++ {
		engine.Step()
	}
	state := engine.State()
	if state.Time != 0 || state.ElectricField != 0 || state.Polarization != 0 || len(state.VoltageHistory) != 0 || len(state.PolHistory) != 0 {
		t.Fatalf("inert invalid-material engine state = %+v, want no simulation advance", state)
	}
	fields, pol := engine.GetHysteresisData()
	if fields != nil || pol != nil {
		t.Fatalf("invalid-material hysteresis data = %v/%v, want nil", fields, pol)
	}
}

func isFiniteE2E(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
