package arraysim

import (
	"math"
	"testing"

	"fecim-lattice-tools/shared/peripherals"
)

func TestModule4ArraySimE2EPeripheralPVTNoiseAndPumpChain(t *testing.T) {
	dacs := []*peripherals.DAC{peripherals.DefaultDAC(), peripherals.ThermometerDAC(5)}
	adcs := []*peripherals.ADC{peripherals.DefaultADC(), peripherals.FlashADC(4), peripherals.RampADC(6, 4), peripherals.ComparatorADC()}
	tias := []*peripherals.TIA{
		peripherals.DefaultTIA(),
		{Gain: 25e3, Bandwidth: 50e6, InputNoiseRMS: 2e-12, OutputOffset: 0.01, MaxInputCurrent: 80e-6, MaxOutputVoltage: 1.2},
	}
	temps := []float64{77, 300, 425}
	corners := []peripherals.ProcessCorner{peripherals.CornerFast, peripherals.CornerTypical, peripherals.CornerSlow}

	for _, dac := range dacs {
		for _, adc := range adcs {
			for _, tia := range tias {
				for _, temp := range temps {
					for _, corner := range corners {
						code := dac.Levels() / 2
						vwrite := dac.ConvertWithCondition(code, temp, corner)
						if math.IsNaN(vwrite) || math.IsInf(vwrite, 0) || vwrite < dac.VrefLow-1 || vwrite > dac.VrefHigh+1 {
							t.Fatalf("DAC condition output invalid dac=%+v temp=%g corner=%s v=%g", dac, temp, corner, vwrite)
						}
						current := math.Abs(vwrite) * 4e-6
						vout := tia.ConvertWithNoise(current)
						if vout < 0 || vout > tia.MaxOutputVoltage || math.IsNaN(vout) {
							t.Fatalf("TIA noisy output invalid: tia=%+v vout=%g", tia, vout)
						}
						adcCode := adc.ConvertWithCondition(vout, temp, corner)
						if adcCode < 0 || adcCode >= adc.Levels() {
							t.Fatalf("ADC code invalid adc=%s code=%d levels=%d", adc.TypeString(), adcCode, adc.Levels())
						}
						pvt := peripherals.AnalyzeINLDNLAtCondition(dac, adc, temp, corner)
						if pvt == nil || pvt.DAC == nil || pvt.ADC == nil || pvt.DAC.Levels <= 0 || pvt.ADC.Levels <= 0 || math.IsNaN(pvt.INLScale) || math.IsNaN(pvt.DNLScale) {
							t.Fatalf("PVT INL/DNL invalid: %+v", pvt)
						}
					}
				}
			}
		}
	}

	cornerSummary := peripherals.AnalyzeProcessCorners(peripherals.DefaultDAC(), peripherals.DefaultADC(), 400)
	if cornerSummary.Fast == nil || cornerSummary.Typical == nil || cornerSummary.Slow == nil || cornerSummary.Slow.DAC.MaxINL < cornerSummary.Fast.DAC.MaxINL {
		t.Fatalf("process corner summary invalid: %+v", cornerSummary)
	}

	adc := peripherals.DefaultADC()
	adc.EnableSARNoise()
	adc.SetTemperature(360)
	low, high := adc.GetEffectiveVref()
	if low < 0 || high <= low || adc.GetThermalNoiseVoltage() <= 0 {
		t.Fatalf("SAR noise effective references invalid: low=%g high=%g report=%+v", low, high, adc.GetSARNoiseReport())
	}
	nearThreshold := adc.GetMetastabilityErrorRate(0.5, 0.5)
	farThreshold := adc.GetMetastabilityErrorRate(0.1, 0.9)
	if nearThreshold < farThreshold || nearThreshold <= 0 {
		t.Fatalf("metastability rates invalid: near=%g far=%g", nearThreshold, farThreshold)
	}
	noisyA := adc.ConvertWithSARNoise(0.5, 123)
	noisyB := adc.ConvertWithSARNoise(0.5, 123)
	if noisyA != noisyB || noisyA < 0 || noisyA >= adc.Levels() {
		t.Fatalf("SAR noise deterministic conversion invalid: %d/%d", noisyA, noisyB)
	}
	adc.DisableSARNoise()

	pumpCases := []*peripherals.ChargePump{
		peripherals.DefaultChargePump(),
		peripherals.FeCAPChargePump(1.0),
		peripherals.FeFETChargePump(1.0, 3.0),
		peripherals.FeFETChargePump(0.8, 5.0),
	}
	for _, pump := range pumpCases {
		if pump.Stages <= 0 || pump.ActualOutputVoltage() < 0 || pump.EnergyPerCycle() <= 0 || pump.PowerInput() < pump.PowerOutput() || pump.EnergyPerOperation(50e-9) <= 0 || pump.Area() <= 0 {
			t.Fatalf("charge pump contract invalid: %+v", pump)
		}
		if pump.MaxCurrentCapability() <= 0 || pump.ChargeTransferEfficiency() <= 0 || pump.ChargeTransferEfficiency() > 1.5 {
			t.Fatalf("charge pump capability invalid: %+v", pump)
		}
	}
	if got := peripherals.StagesRequired(3.0, 0.2, 0.3); got != 0 {
		t.Fatalf("StagesRequired no-headroom = %d, want 0", got)
	}
}

func TestModule4ArraySimE2ESampleHoldRegulatorSenseChainWorkflow(t *testing.T) {
	sh := peripherals.DefaultSampleAndHold()
	if sh.SettledFraction(0) != 0 || sh.SettledFraction(100e-9) <= sh.SettledFraction(1e-9) || sh.HoldDroop(0) != 1 || sh.HoldDroop(1e-3) >= sh.HoldDroop(1e-6) {
		t.Fatalf("sample-and-hold dynamics invalid")
	}
	reg := peripherals.DefaultVoltageRegulator()
	loads := []float64{0, 1e-6, 20e-6, 2e-3}
	prev := math.Inf(1)
	for _, load := range loads {
		vout := reg.Regulate(1.5, load)
		if vout < 0 || vout > reg.NominalVoltage || vout > prev+1e-12 {
			t.Fatalf("regulator load response invalid load=%g vout=%g prev=%g", load, vout, prev)
		}
		prev = vout
	}
	if reg.Regulate(0.05, 0) != 0 || reg.SupplyNoiseTransfer(0.1) >= 0.1 {
		t.Fatalf("regulator dropout/noise behavior invalid")
	}

	sense := SenseChain{TIA: TIAConfig{Rf: 15e3, Vref: 0.02, Vmin: 0, Vmax: 1.0}, ADC: ADCConfig{Bits: 6, Vmin: 0, Vmax: 1.0}}
	currents := []float64{-10e-6, 0, 1e-6, 10e-6, 100e-6}
	results := sense.ConvertCurrents(currents)
	if len(results) != len(currents) {
		t.Fatalf("sense result count=%d", len(results))
	}
	for i, res := range results {
		if res.Code < 0 || res.Code >= 64 || res.Vout < 0 || res.Vout > 1.0 {
			t.Fatalf("sense result[%d] invalid: %+v", i, res)
		}
	}
	if results[len(results)-1].Code < results[2].Code || !results[len(results)-1].TIASaturated {
		t.Fatalf("large current should increase/saturate sense result: %+v", results)
	}
	badSense := SenseChain{TIA: TIAConfig{Rf: 0, Vref: 0, Vmin: 0, Vmax: 1}, ADC: ADCConfig{Bits: 0, Vmin: 0, Vmax: 1}}
	if lo, hi := badSense.CurrentRange(); lo != 0 || hi != 0 || badSense.CurrentLSB() != 0 {
		t.Fatalf("bad sense range invalid: lo=%g hi=%g lsb=%g", lo, hi, badSense.CurrentLSB())
	}

	thermal := peripherals.ThermalNoiseRMS(300, 10e3, 100e6)
	shot := peripherals.ShotNoiseCurrentRMS(10e-6, 100e6)
	qvar := peripherals.QuantizationNoiseVariance(1.0, 6)
	if thermal <= 0 || shot <= 0 || qvar <= 0 || peripherals.FlickerNoisePower(1, 10) <= peripherals.FlickerNoisePower(1, 100) || peripherals.TotalNoiseVariance(thermal, shot) <= thermal*thermal {
		t.Fatalf("noise helper contracts invalid")
	}
	if !math.IsInf(peripherals.SNRDB(1, 0), 1) || !math.IsInf(peripherals.SNRDB(0, 1), -1) || !math.IsNaN(peripherals.SNRDB(0, 0)) {
		t.Fatalf("SNR boundary contracts invalid")
	}
}
