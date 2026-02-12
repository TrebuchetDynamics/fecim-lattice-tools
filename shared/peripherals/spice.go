package peripherals

import "fmt"

// BuildBehavioralSpiceSubcircuits returns minimal behavioral peripheral subcircuits.
func BuildBehavioralSpiceSubcircuits(dac *DAC, adc *ADC, tia *TIA, sh *SampleAndHold, vr *VoltageRegulator) string {
	if dac == nil {
		dac = DefaultDAC()
	}
	if adc == nil {
		adc = DefaultADC()
	}
	if tia == nil {
		tia = DefaultTIA()
	}
	if sh == nil {
		sh = DefaultSampleAndHold()
	}
	if vr == nil {
		vr = DefaultVoltageRegulator()
	}

	adcScale := float64((int64(1)<<uint(adc.Bits))-1) / (adc.VrefHigh - adc.VrefLow)

	return fmt.Sprintf(`
* ===== Peripheral behavioral subcircuits =====
.subckt DAC5 vin vout vss
Rdac vout vin 1k
Edac vout vss vin vss 1
.ends DAC5

.subckt SAMPLE_HOLD vin vout vclk vss
Rsw vin n_hold %.6g
Chold n_hold vss %.6g
Rleak n_hold vss %.6g
Ebuf vout vss n_hold vss 1
.ends SAMPLE_HOLD

.subckt TIA_BASIC iin vout vss
Rtia vout iin %.6g
Voff vout n_off %.6g
Rclamp n_off vss 1e12
.ends TIA_BASIC

.subckt ADC5 vin vcode vss
Eadc vcode vss vin vss %.6g
Radc vcode vss 1e9
.ends ADC5

.subckt VREG_BASIC vin vout vss
Ereg nreg vss vin vss 1
Rdrop nreg vout %.6g
Rpsrr vout vss %.6g
.ends VREG_BASIC
* =============================================
`,
		sh.SwitchResistance,
		sh.HoldCapacitance,
		sh.LeakageResistance,
		tia.Gain,
		tia.OutputOffset,
		adcScale,
		vr.OutputResistance,
		1.0)
}
