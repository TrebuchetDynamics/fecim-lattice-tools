// Demo 4: Peripheral Circuits for Ferroelectric CIM
//
// This demo visualizes the peripheral circuits required for a complete
// ferroelectric compute-in-memory system: DAC, ADC, TIA, and Charge Pump.
// Shows how digital values are converted to/from analog for crossbar operations.
package main

import (
	"flag"
	"fmt"
	"strings"

	"ironlattice-vis/demo4-circuits/pkg/peripherals"
)

func main() {
	// Command-line flags
	showDAC := flag.Bool("dac", false, "Show DAC (Digital-to-Analog) details")
	showADC := flag.Bool("adc", false, "Show ADC (Analog-to-Digital) details")
	showTIA := flag.Bool("tia", false, "Show TIA (Transimpedance Amplifier) details")
	showPump := flag.Bool("pump", false, "Show Charge Pump details")
	showAll := flag.Bool("all", false, "Show all peripheral circuits")
	demoLevel := flag.Int("level", 15, "Demo level for conversion (0-29)")
	flag.Parse()

	fmt.Println("================================================")
	fmt.Println("  IronLattice Demo 4: Peripheral Circuits")
	fmt.Println("  Full System Integration for CIM")
	fmt.Println("================================================")
	fmt.Println()

	// Show system overview
	showSystemOverview()

	// Show specific circuits or all
	if *showAll || *showDAC {
		showDACDemo(*demoLevel)
	}
	if *showAll || *showADC {
		showADCDemo(*demoLevel)
	}
	if *showAll || *showTIA {
		showTIADemo()
	}
	if *showAll || *showPump {
		showChargePumpDemo()
	}

	// If no specific flag, show brief overview of all
	if !*showDAC && !*showADC && !*showTIA && !*showPump && !*showAll {
		showBriefOverview(*demoLevel)
	}

	fmt.Println()
	fmt.Println("================================================")
	fmt.Println("  Peripheral circuits enable CMOS-compatible")
	fmt.Println("  ferroelectric compute-in-memory systems")
	fmt.Println("================================================")
}

func showSystemOverview() {
	fmt.Println("System Architecture:")
	fmt.Println()
	fmt.Println("     WRITE PATH                    READ PATH")
	fmt.Println("     ──────────                    ─────────")
	fmt.Println()
	fmt.Println("  Digital Level ──┐            ┌── Digital Level")
	fmt.Println("      (0-29)      │            │      (0-29)")
	fmt.Println("                  ▼            ▲")
	fmt.Println("            ┌─────────┐  ┌─────────┐")
	fmt.Println("            │   DAC   │  │   ADC   │")
	fmt.Println("            │  5-bit  │  │  5-bit  │")
	fmt.Println("            └────┬────┘  └────┬────┘")
	fmt.Println("                 │            ▲")
	fmt.Println("                 ▼            │")
	fmt.Println("            ┌─────────┐  ┌─────────┐")
	fmt.Println("            │ Charge  │  │   TIA   │")
	fmt.Println("            │  Pump   │  │ Current │")
	fmt.Println("            │ 1V→1.5V │  │→Voltage │")
	fmt.Println("            └────┬────┘  └────┬────┘")
	fmt.Println("                 │            ▲")
	fmt.Println("                 ▼            │")
	fmt.Println("            ┌────────────────────┐")
	fmt.Println("            │                    │")
	fmt.Println("            │     CROSSBAR       │")
	fmt.Println("            │   30-Level FeFET   │")
	fmt.Println("            │                    │")
	fmt.Println("            └────────────────────┘")
	fmt.Println()
}

func showDACDemo(level int) {
	fmt.Println("┌─────────────────────────────────────────────┐")
	fmt.Println("│           DAC (Write Path)                  │")
	fmt.Println("└─────────────────────────────────────────────┘")
	fmt.Println()

	dac := peripherals.DefaultDAC()

	fmt.Printf("Configuration:\n")
	fmt.Printf("  Resolution: %d bits (%d levels)\n", dac.Bits, dac.Levels())
	fmt.Printf("  Vref Range: %.1fV to %.1fV\n", dac.VrefLow, dac.VrefHigh)
	fmt.Printf("  LSB Size: %.3f V\n", dac.Resolution())
	fmt.Printf("  INL: %.2f LSB, DNL: %.2f LSB\n", dac.INL, dac.DNL)
	fmt.Printf("  Settling Time: %.0f ns\n", dac.SettleTime)
	fmt.Printf("  Energy/Conv: %.2f fJ\n", dac.EnergyPerConversion()*1e15)
	fmt.Println()

	// Show conversion for specified level
	if level < 0 || level > 29 {
		level = 15
	}
	voltage := dac.Convert(level)
	voltageNL := dac.ConvertWithNonlinearity(level)

	fmt.Printf("Level %d Conversion:\n", level)
	fmt.Printf("  Ideal Voltage: %+.3f V\n", voltage)
	fmt.Printf("  With NL Error: %+.3f V (Δ = %.3f mV)\n", voltageNL, (voltageNL-voltage)*1000)
	fmt.Println()

	// Show voltage ladder
	fmt.Println("Voltage Ladder (30 IronLattice levels):")
	fmt.Println()
	for i := 0; i < 30; i++ {
		v := dac.Convert(i)
		bar := int((v - dac.VrefLow) / (dac.VrefHigh - dac.VrefLow) * 40)
		marker := " "
		if i == level {
			marker = "→"
		}
		fmt.Printf("  %s %2d: %+.2fV │%s│\n", marker, i, v, strings.Repeat("█", bar)+strings.Repeat("░", 40-bar))
	}
	fmt.Println()
}

func showADCDemo(level int) {
	fmt.Println("┌─────────────────────────────────────────────┐")
	fmt.Println("│           ADC (Read Path)                   │")
	fmt.Println("└─────────────────────────────────────────────┘")
	fmt.Println()

	adc := peripherals.DefaultADC()

	fmt.Printf("Configuration:\n")
	fmt.Printf("  Resolution: %d bits (%d levels)\n", adc.Bits, adc.Levels())
	fmt.Printf("  Vref Range: %.1fV to %.1fV\n", adc.VrefLow, adc.VrefHigh)
	fmt.Printf("  LSB Size: %.3f V\n", adc.Resolution())
	fmt.Printf("  Type: SAR (Successive Approximation)\n")
	fmt.Printf("  Conversion Time: %.0f ns\n", adc.ConversionTime)
	fmt.Printf("  ENOB: %.2f bits\n", adc.ENOB())
	fmt.Printf("  Theoretical SNR: %.1f dB\n", adc.TheoreticalSNR())
	fmt.Printf("  Effective SNR: %.1f dB\n", adc.EffectiveSNR())
	fmt.Printf("  Energy/Conv: %.2f fJ\n", adc.EnergyPerConversion()*1e15)
	fmt.Println()

	// Demo: Convert voltage back to level
	voltage := float64(level) / 29.0 * (adc.VrefHigh - adc.VrefLow) + adc.VrefLow
	convertedLevel := adc.Convert(voltage)
	convertedLevelNL := adc.ConvertWithNonlinearity(voltage)

	fmt.Printf("ADC Conversion (Input: %.3fV for level %d):\n", voltage, level)
	fmt.Printf("  Ideal Output: Level %d\n", convertedLevel)
	fmt.Printf("  With NL: Level %d\n", convertedLevelNL)
	fmt.Println()

	// Show quantization
	fmt.Println("Quantization Thresholds:")
	for i := 0; i < 8; i++ {
		threshold := adc.VrefLow + float64(i+1)*adc.Resolution()
		fmt.Printf("  Level %d-%d boundary: %.3fV\n", i, i+1, threshold)
	}
	fmt.Println("  ... (30 total thresholds)")
	fmt.Println()
}

func showTIADemo() {
	fmt.Println("┌─────────────────────────────────────────────┐")
	fmt.Println("│     TIA (Transimpedance Amplifier)          │")
	fmt.Println("└─────────────────────────────────────────────┘")
	fmt.Println()

	tia := peripherals.DefaultTIA()

	fmt.Printf("Configuration:\n")
	fmt.Printf("  Transimpedance Gain: %.0f kΩ\n", tia.Gain/1e3)
	fmt.Printf("  Bandwidth: %.0f MHz\n", tia.Bandwidth/1e6)
	fmt.Printf("  Input Noise: %.1f pA/√Hz\n", tia.InputNoiseRMS*1e12)
	fmt.Printf("  Output Offset: %.1f mV\n", tia.OutputOffset*1000)
	fmt.Printf("  Max Input Current: %.0f µA\n", tia.MaxInputCurrent*1e6)
	fmt.Printf("  Max Output Voltage: %.1f V\n", tia.MaxOutputVoltage)
	fmt.Println()

	fmt.Printf("Performance:\n")
	fmt.Printf("  Min Detectable Current: %.2f nA\n", tia.MinDetectableCurrent()*1e9)
	fmt.Printf("  Dynamic Range: %.1f dB\n", tia.DynamicRange())
	fmt.Printf("  Settling Time: %.1f ns\n", tia.SettlingTime()*1e9)
	fmt.Printf("  Power: %.1f µW\n", tia.PowerConsumption()*1e6)
	fmt.Println()

	// Show current-to-voltage conversion examples
	fmt.Println("Current-to-Voltage Conversion:")
	testCurrents := []float64{1e-6, 10e-6, 50e-6, 100e-6}
	for _, current := range testCurrents {
		voltage := tia.Convert(current)
		snr := tia.SNR(current)
		fmt.Printf("  %5.0f µA → %.3f V (SNR: %.1f dB)\n", current*1e6, voltage, snr)
	}
	fmt.Println()
}

func showChargePumpDemo() {
	fmt.Println("┌─────────────────────────────────────────────┐")
	fmt.Println("│          Charge Pump (Voltage Boost)        │")
	fmt.Println("└─────────────────────────────────────────────┘")
	fmt.Println()

	// Positive pump
	pumpPos := peripherals.DefaultChargePump()
	fmt.Println("Positive Charge Pump (+1.5V):")
	fmt.Printf("  Input: %.1f V (CMOS supply)\n", pumpPos.InputVoltage)
	fmt.Printf("  Target Output: %.1f V\n", pumpPos.OutputVoltage)
	fmt.Printf("  Stages: %d (Dickson topology)\n", pumpPos.Stages)
	fmt.Printf("  Ideal Output: %.2f V\n", pumpPos.IdealOutputVoltage())
	fmt.Printf("  Actual Output: %.2f V (with losses)\n", pumpPos.ActualOutputVoltage())
	fmt.Printf("  Boost Factor: %.2fx\n", pumpPos.BoostFactor())
	fmt.Printf("  Efficiency: %.0f%%\n", pumpPos.Efficiency*100)
	fmt.Printf("  Ripple: %.1f mV\n", pumpPos.OutputRipple()*1000)
	fmt.Printf("  Rise Time: %.1f µs\n", pumpPos.RiseTime()*1e6)
	fmt.Println()

	// Negative pump
	pumpNeg := peripherals.NegativePump()
	fmt.Println("Negative Charge Pump (-1.5V):")
	fmt.Printf("  Input: %.1f V\n", pumpNeg.InputVoltage)
	fmt.Printf("  Target Output: %.1f V\n", pumpNeg.OutputVoltage)
	fmt.Printf("  Stages: %d\n", pumpNeg.Stages)
	fmt.Println()

	// Energy analysis
	fmt.Println("Energy Analysis:")
	pulseDuration := 100e-9 // 100 ns write pulse
	energy := pumpPos.EnergyPerOperation(pulseDuration)
	fmt.Printf("  Write Pulse Duration: %.0f ns\n", pulseDuration*1e9)
	fmt.Printf("  Energy per Write: %.2f pJ\n", energy*1e12)
	fmt.Printf("  Power Input: %.1f µW\n", pumpPos.PowerInput()*1e6)
	fmt.Printf("  Power Loss: %.1f µW\n", pumpPos.PowerLoss()*1e6)
	fmt.Println()
}

func showBriefOverview(level int) {
	dac := peripherals.DefaultDAC()
	adc := peripherals.DefaultADC()
	tia := peripherals.DefaultTIA()
	pump := peripherals.DefaultChargePump()

	fmt.Println("Peripheral Circuit Summary:")
	fmt.Println()

	// DAC summary
	voltage := dac.Convert(level)
	fmt.Printf("  DAC: Level %d → %+.3f V (5-bit, %.0f fJ/conv)\n",
		level, voltage, dac.EnergyPerConversion()*1e15)

	// Charge pump summary
	fmt.Printf("  Charge Pump: %.1f V → %.2f V (%.0f%% efficient)\n",
		pump.InputVoltage, pump.ActualOutputVoltage(), pump.Efficiency*100)

	// TIA summary
	current := float64(level) / 29.0 * tia.MaxInputCurrent
	tiaVoltage := tia.Convert(current)
	fmt.Printf("  TIA: %.1f µA → %.3f V (%.0f kΩ gain)\n",
		current*1e6, tiaVoltage, tia.Gain/1e3)

	// ADC summary
	adcLevel := adc.Convert(tiaVoltage)
	fmt.Printf("  ADC: %.3f V → Level %d (5-bit, %.0f fJ/conv)\n",
		tiaVoltage, adcLevel, adc.EnergyPerConversion()*1e15)
	fmt.Println()

	// Total energy estimate
	totalEnergy := dac.EnergyPerConversion() + adc.EnergyPerConversion() + pump.EnergyPerOperation(100e-9)
	fmt.Printf("Estimated Energy per Operation: %.1f fJ\n", totalEnergy*1e15)
	fmt.Println()

	fmt.Println("Run with --all for detailed view of all circuits")
	fmt.Println("Or use --dac, --adc, --tia, --pump for specific circuits")
}
