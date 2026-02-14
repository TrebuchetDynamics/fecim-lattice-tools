package peripherals

import (
	"math"
	"testing"
)

func TestPeripheralsINLDNLRegression_DACAndADC(t *testing.T) {
	for _, bits := range []int{4, 5, 8} {
		t.Run("DAC_"+itoa(bits)+"bit", func(t *testing.T) {
			dac := DefaultDAC()
			dac.Bits = bits
			maxINL, maxAbsDNL := sweepDACINLDNL(dac)
			if maxINL >= 1.0 {
				t.Fatalf("DAC %d-bit INL too high: %.4f LSB", bits, maxINL)
			}
			if maxAbsDNL >= 0.5 {
				t.Fatalf("DAC %d-bit DNL too high: %.4f LSB", bits, maxAbsDNL)
			}
		})

		t.Run("ADC_"+itoa(bits)+"bit", func(t *testing.T) {
			adc := DefaultADC()
			adc.Bits = bits
			maxINL, maxAbsDNL := sweepADCINLDNL(adc)
			if maxINL >= 1.0 {
				t.Fatalf("ADC %d-bit INL too high: %.4f LSB", bits, maxINL)
			}
			if maxAbsDNL >= 0.5 {
				t.Fatalf("ADC %d-bit DNL too high: %.4f LSB", bits, maxAbsDNL)
			}
		})
	}
}

func sweepDACINLDNL(dac *DAC) (maxAbsINL, maxAbsDNL float64) {
	levels := dac.Levels()
	lsb := dac.Resolution()
	prev := dac.ConvertWithNonlinearity(0)
	for code := 0; code < levels; code++ {
		ideal := dac.Convert(code)
		actual := dac.ConvertWithNonlinearity(code)
		inl := (actual - ideal) / lsb
		if math.Abs(inl) > maxAbsINL {
			maxAbsINL = math.Abs(inl)
		}
		if code > 0 {
			step := actual - prev
			dnl := step/lsb - 1.0
			if math.Abs(dnl) > maxAbsDNL {
				maxAbsDNL = math.Abs(dnl)
			}
		}
		prev = actual
	}
	return
}

func sweepADCINLDNL(adc *ADC) (maxAbsINL, maxAbsDNL float64) {
	levels := adc.Levels()
	lsb := adc.Resolution()
	inlEff, _ := EffectiveINLDNL(adc.INL, adc.DNL, referenceTemperatureK, CornerTypical)
	for code := 0; code < levels; code++ {
		inl := inlEff * math.Sin(math.Pi*float64(code)/float64(levels-1))
		if math.Abs(inl) > maxAbsINL {
			maxAbsINL = math.Abs(inl)
		}
	}
	for code := 1; code < levels; code++ {
		width := findCodeWidth(adc, code, lsb)
		dnl := (width - lsb) / lsb
		if math.Abs(dnl) > maxAbsDNL {
			maxAbsDNL = math.Abs(dnl)
		}
	}
	return
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	buf := [20]byte{}
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + (v % 10))
		v /= 10
	}
	return string(buf[i:])
}
