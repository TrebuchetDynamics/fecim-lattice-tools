package arraysim

import (
	"fecim-lattice-tools/shared/peripherals"
	"math"
)

// ComposedSenseSNR is the native GUI's input-referred noise diagnostic. The
// signal is current, never the TIA's DC output offset. Nonfinite results remain
// explicitly unavailable in browser numeric cells, not invented finite values.
func ComposedSenseSNR(currentA float64, sense SenseChain, tia *peripherals.TIA) float64 {
	if tia == nil {
		return math.NaN()
	}
	signalRMS := math.Abs(currentA)
	bandwidth := tia.Bandwidth
	if bandwidth <= 0 {
		bandwidth = 1
	}
	tiaNoiseRMS := math.Abs(tia.InputNoiseRMS) * math.Sqrt(bandwidth)
	shotNoiseRMS := peripherals.ShotNoiseCurrentRMS(math.Abs(currentA), bandwidth)
	quantNoiseRMS := 0.
	if lsb := sense.CurrentLSB(); lsb > 0 {
		quantNoiseRMS = lsb / math.Sqrt(12)
	}
	totalNoiseRMS := math.Sqrt(peripherals.TotalNoiseVariance(tiaNoiseRMS, shotNoiseRMS, quantNoiseRMS))
	return peripherals.SNRDB(signalRMS, totalNoiseRMS)
}
func DeviceSenseSNR(s DeviceSettings, currentA float64) float64 {
	sense := SenseChain{TIA: TIAConfig{Rf: s.TIA.Gain, Vref: s.TIA.OutputOffset, Vmin: 0, Vmax: s.TIA.MaxOutputVoltage}, ADC: ADCConfig{Bits: s.ADC.Bits, Vmin: s.ADC.VrefLow, Vmax: s.ADC.VrefHigh}}
	return ComposedSenseSNR(currentA, sense, s.TIA)
}
