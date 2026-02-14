package peripherals

import (
	"math"
	"testing"
)

func TestPeripheralsNoiseValidation_ADCAndTIA(t *testing.T) {
	t.Run("ADC quantization noise", func(t *testing.T) {
		adc := DefaultADC()
		adc.EnableSARNoise()
		fixedInput := 0.53
		const samples = 100

		mean := 0.0
		codes := make([]float64, samples)
		for i := 0; i < samples; i++ {
			code := adc.ConvertWithSARNoise(fixedInput, int64(i+1))
			codes[i] = float64(code)
			mean += codes[i]
		}
		mean /= samples

		var ss float64
		for _, c := range codes {
			d := c - mean
			ss += d * d
		}
		rmsLSB := math.Sqrt(ss / samples)
		if rmsLSB >= 0.5 {
			t.Fatalf("ADC RMS noise too high: %.4f LSB (limit 0.5 LSB)", rmsLSB)
		}
	})

	t.Run("TIA linearity", func(t *testing.T) {
		tia := DefaultTIA()
		for i := 1; i <= 80; i++ {
			current := float64(i) * 1e-6 // 1uA..80uA within linear range
			ideal := current*tia.Gain + tia.OutputOffset
			actual := tia.Convert(current)
			relErr := math.Abs(actual-ideal) / math.Abs(ideal)
			if relErr >= 0.02 {
				t.Fatalf("TIA linearity exceeded at %.2fuA: err=%.4f (actual=%.6f ideal=%.6f)", current*1e6, relErr, actual, ideal)
			}
		}
	})
}
