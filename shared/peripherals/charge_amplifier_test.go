package peripherals

// Tests for the ChargeAmplifier (LIT-P2-03).

import (
	"math"
	"testing"
)

func TestChargeAmplifier_Defaults(t *testing.T) {
	ca := DefaultChargeAmplifier()
	if ca.Cfb <= 0 {
		t.Errorf("Cfb=%e <= 0", ca.Cfb)
	}
	if ca.Bandwidth <= 0 {
		t.Errorf("Bandwidth=%e <= 0", ca.Bandwidth)
	}
	if ca.MaxOutputVoltage <= 0 {
		t.Errorf("MaxOutputVoltage=%e <= 0", ca.MaxOutputVoltage)
	}
}

func TestChargeAmplifier_DefaultIntegratedKTCSampleNoise(t *testing.T) {
	ca := DefaultChargeAmplifier()

	const (
		minExpectedNoise = 20e-18
		maxExpectedNoise = 26e-18
		fourBitLevels    = 1 << 4
	)
	if ca.InputChargeNoiseRMS < minExpectedNoise || ca.InputChargeNoiseRMS > maxExpectedNoise {
		t.Fatalf("default integrated charge noise=%e C, want near 23 aC", ca.InputChargeNoiseRMS)
	}
	if got := ca.MinDetectableCharge(); got != ca.InputChargeNoiseRMS {
		t.Fatalf("MinDetectableCharge=%e C, want integrated RMS %e C", got, ca.InputChargeNoiseRMS)
	}
	if got := ca.SNR(ca.MaxInputCharge); got <= 10*fourBitLevels {
		t.Fatalf("full-scale SNR=%e, want comfortably above %d levels", got, fourBitLevels)
	}

	floor := ca.MinDetectableCharge()
	ca.Bandwidth *= 10
	if got := ca.MinDetectableCharge(); got != floor {
		t.Fatalf("integrated kT/C floor changed with bandwidth: got %e C want %e C", got, floor)
	}
}

func TestChargeAmplifier_Sense(t *testing.T) {
	ca := DefaultChargeAmplifier()

	// V_out = Q / Cfb
	q := 64e-15 // 64 fC
	vOut := ca.Sense(q)
	want := q / ca.Cfb
	if math.Abs(vOut-want) > 1e-12 {
		t.Errorf("Sense(%e): got %e want %e", q, vOut, want)
	}

	// Clipping at MaxOutputVoltage.
	vClipped := ca.Sense(1000e-15) // huge charge
	if vClipped != ca.MaxOutputVoltage {
		t.Errorf("clipped output %e != MaxOutputVoltage %e", vClipped, ca.MaxOutputVoltage)
	}
	vNeg := ca.Sense(-1000e-15)
	if vNeg != -ca.MaxOutputVoltage {
		t.Errorf("negative clipped output %e != -%e", vNeg, ca.MaxOutputVoltage)
	}

	// Zero charge → zero output.
	if ca.Sense(0) != 0 {
		t.Errorf("Sense(0) = %e != 0", ca.Sense(0))
	}
}

func TestChargeAmplifier_SenseWithNoiseUsesSignedGaussianSample(t *testing.T) {
	ca := DefaultChargeAmplifier()
	sigmaV := ca.InputChargeNoiseRMS / ca.Cfb

	positive := ca.senseWithGaussian(0, func() float64 { return 1 })
	negative := ca.senseWithGaussian(0, func() float64 { return -1 })

	if math.Abs(positive-sigmaV) > 1e-15 {
		t.Fatalf("positive sample=%e want %e", positive, sigmaV)
	}
	if math.Abs(negative+sigmaV) > 1e-15 {
		t.Fatalf("negative sample=%e want -%e", negative, sigmaV)
	}
	if math.Abs(positive+negative) > 1e-15 {
		t.Fatalf("injected symmetric samples are biased: +%e -%e", positive, negative)
	}
}

func TestChargeAmplifier_SenseWithNoisePublicPathSamplesBothSigns(t *testing.T) {
	ca := DefaultChargeAmplifier()
	var sawPositive, sawNegative bool
	for range 256 {
		sample := ca.SenseWithNoise(0)
		sawPositive = sawPositive || sample > 0
		sawNegative = sawNegative || sample < 0
		if sawPositive && sawNegative {
			return
		}
	}
	t.Fatalf("public Gaussian source did not produce both signs: positive=%t negative=%t", sawPositive, sawNegative)
}

func TestChargeAmplifier_SenseWithNoiseClipsOnceAfterNoise(t *testing.T) {
	ca := DefaultChargeAmplifier()
	sigmaV := ca.InputChargeNoiseRMS / ca.Cfb

	gotHigh := ca.senseWithGaussian(0.99*ca.MaxInputCharge, func() float64 { return 1000 })
	gotLow := ca.senseWithGaussian(-0.99*ca.MaxInputCharge, func() float64 { return -1000 })
	if gotHigh != ca.MaxOutputVoltage {
		t.Fatalf("high noisy output=%e want rail %e", gotHigh, ca.MaxOutputVoltage)
	}
	if gotLow != -ca.MaxOutputVoltage {
		t.Fatalf("low noisy output=%e want rail -%e", gotLow, ca.MaxOutputVoltage)
	}

	inwardSample := -100.0
	gotInward := ca.senseWithGaussian(1.01*ca.MaxInputCharge, func() float64 { return inwardSample })
	wantInward := 1.01*ca.MaxInputCharge/ca.Cfb + inwardSample*sigmaV
	if math.Abs(gotInward-wantInward) > 1e-15 {
		t.Fatalf("inward noisy output=%e want unclipped post-noise value %e", gotInward, wantInward)
	}
}

func TestChargeAmplifier_SNR(t *testing.T) {
	ca := DefaultChargeAmplifier()

	// Larger charge → higher SNR.
	snr1 := ca.SNR(10e-15)
	snr2 := ca.SNR(100e-15)
	if snr2 <= snr1 {
		t.Errorf("SNR should increase with charge: SNR(10fC)=%e SNR(100fC)=%e", snr1, snr2)
	}
}

func TestChargeAmplifier_SettlingTime(t *testing.T) {
	ca := DefaultChargeAmplifier()
	ts := ca.SettlingTime()
	// For 500 MHz BW: ts = 7/(2π*500e6) ≈ 2.2 ns
	if ts <= 0 || ts > 100e-9 {
		t.Errorf("SettlingTime=%e out of range (0, 100 ns)", ts)
	}
}

func TestChargeAmplifier_MinDetectableCharge(t *testing.T) {
	ca := DefaultChargeAmplifier()
	qMin := ca.MinDetectableCharge()
	if qMin <= 0 {
		t.Errorf("MinDetectableCharge=%e <= 0", qMin)
	}
	// At SNR=1 the signal equals the noise floor.
	if math.Abs(ca.SNR(qMin)-1) > 0.01 {
		t.Errorf("SNR at MinDetectable=%e != 1.0", ca.SNR(qMin))
	}
}
