package peripherals

import "testing"

func BenchmarkTIAADCChainConvert(b *testing.B) {
	tia := DefaultTIA()
	adc := DefaultADC()

	// Representative read current from crossbar column (A).
	currents := []float64{2e-6, 6e-6, 12e-6, 18e-6, 24e-6}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, current := range currents {
			v := tia.Convert(current)
			code := adc.Convert(v)
			if code < 0 || code >= adc.Levels() {
				b.Fatalf("ADC code out of range: %d", code)
			}
		}
	}
}
