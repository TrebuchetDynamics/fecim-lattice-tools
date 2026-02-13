package peripherals

import (
	"math"
	"testing"
)

// M4-INV-04 refine: confirm useful ENOB ceiling near prior ~12.75 result.
func TestM4INV04_ThermalNoiseVsADCRefine(t *testing.T) {
	const (
		vRange = 1.8
		tempK  = 300.0
		rTIA   = 10e3
		bwHz   = 10e6
	)
	thermalVar := math.Pow(ThermalNoiseRMS(tempK, rTIA, bwHz), 2)
	signalRMS := vRange / (2 * math.Sqrt2)

	bestENOB := 0.0
	bestBits := 0
	for bits := 6; bits <= 16; bits++ {
		totalVar := thermalVar + QuantizationNoiseVariance(vRange, bits)
		enob := (SNRDB(signalRMS, math.Sqrt(totalVar)) - 1.76) / 6.02
		if enob > bestENOB {
			bestENOB = enob
			bestBits = bits
		}
		t.Logf("bits=%d enob=%.3f", bits, enob)
	}
	t.Logf("best ENOB=%.3f at %d bits", bestENOB, bestBits)
	if bestENOB < 12.0 {
		t.Fatalf("unexpectedly low ENOB %.3f", bestENOB)
	}
}

// M4-INV-05: Dickson efficiency for 3V output from 1.8V SKY130 supply.
func TestM4INV05_ChargePumpDicksonEfficiencyAt3V(t *testing.T) {
	cp := &ChargePump{
		InputVoltage:   1.8,
		OutputVoltage:  3.0,
		Stages:         2,
		DiodeDrop:      0.25,
		ClockFrequency: 100e6,
		LoadCurrent:    50e-6,
		FlyCapacitance: 200e-12,
		Efficiency:     0.72,
	}
	actualV := cp.ActualOutputVoltage()
	stageEff := cp.ChargeTransferEfficiency()
	pout := cp.PowerOutput()
	pin := cp.PowerInput()
	syseff := 0.0
	if pin > 0 {
		syseff = pout / pin
	}

	t.Logf("Vout_actual=%.3fV stage_eff=%.3f system_eff=%.3f", actualV, stageEff, syseff)
	if actualV < 2.7 {
		t.Fatalf("insufficient boosted voltage: %.3fV", actualV)
	}
	if stageEff < 0.55 {
		t.Fatalf("stage efficiency too low: %.3f", stageEff)
	}
}
