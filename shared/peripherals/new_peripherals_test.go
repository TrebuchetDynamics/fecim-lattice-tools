package peripherals

import (
	"strings"
	"testing"
)

func TestSampleAndHold_BasicBehavior(t *testing.T) {
	sh := DefaultSampleAndHold()
	if got := sh.SettledFraction(5e-9); got <= 0 || got >= 1 {
		t.Fatalf("settled fraction out of range: %g", got)
	}
	if got := sh.HoldDroop(10e-6); got <= 0 || got > 1 {
		t.Fatalf("hold droop ratio out of range: %g", got)
	}
}

func TestVoltageRegulator_BasicBehavior(t *testing.T) {
	vr := DefaultVoltageRegulator()
	v := vr.Regulate(1.8, 100e-6)
	if v <= 0 || v > vr.NominalVoltage {
		t.Fatalf("unexpected regulated output: %g", v)
	}
	r := vr.SupplyNoiseTransfer(10e-3)
	if r <= 0 || r >= 10e-3 {
		t.Fatalf("psrr attenuation failed: %g", r)
	}
}

func TestBuildBehavioralSpiceSubcircuits_IncludesNewPeripherals(t *testing.T) {
	deck := BuildBehavioralSpiceSubcircuits(nil, nil, nil, nil, nil)
	if deck == "" {
		t.Fatal("expected non-empty spice subcircuit deck")
	}
	for _, token := range []string{".subckt SAMPLE_HOLD", ".subckt VREG_BASIC", ".subckt DAC5", ".subckt ADC5", ".subckt TIA_BASIC"} {
		if !strings.Contains(deck, token) {
			t.Fatalf("expected token %q in deck", token)
		}
	}
}
