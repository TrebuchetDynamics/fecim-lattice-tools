package peripherals

import (
	"math"
	"strconv"
	"testing"
)

func testDAC(bits int) *DAC {
	return &DAC{
		Bits:     bits,
		VrefHigh: 1.0,
		VrefLow:  0.0,
		INL:      0.0,
		DNL:      0.25,
	}
}

func testADC(bits int) *ADC {
	return &ADC{
		Bits:     bits,
		VrefHigh: 1.0,
		VrefLow:  0.0,
		INL:      0.0,
		DNL:      0.0,
	}
}

func TestLinearity_DACMonotonicNoMissingCodes(t *testing.T) {
	for _, bits := range []int{4, 6, 8} {
		t.Run("bits="+strconv.Itoa(bits), func(t *testing.T) {
			dac := testDAC(bits)
			prev := dac.ConvertWithNonlinearity(0)
			for code := 1; code < dac.Levels(); code++ {
				v := dac.ConvertWithNonlinearity(code)
				if !(v > prev) {
					t.Fatalf("DAC not monotonic at bits=%d code=%d: v[%d]=%.9f <= v[%d]=%.9f", bits, code, code, v, code-1, prev)
				}
				prev = v
			}
		})
	}
}

func TestLinearity_ADCMonotonicSweep(t *testing.T) {
	for _, bits := range []int{4, 6, 8} {
		t.Run("bits="+strconv.Itoa(bits), func(t *testing.T) {
			adc := testADC(bits)
			steps := adc.Levels() * 16
			prev := adc.Convert(adc.VrefLow)
			for i := 1; i <= steps; i++ {
				v := adc.VrefLow + (adc.VrefHigh-adc.VrefLow)*float64(i)/float64(steps)
				code := adc.Convert(v)
				if code < prev {
					t.Fatalf("ADC not monotonic at bits=%d step=%d: code=%d < prev=%d", bits, i, code, prev)
				}
				prev = code
			}
		})
	}
}

func TestLinearity_DACDNLWithinOneLSB(t *testing.T) {
	for _, bits := range []int{4, 6, 8} {
		t.Run("bits="+strconv.Itoa(bits), func(t *testing.T) {
			dac := testDAC(bits)
			lsb := dac.Resolution()

			for code := 1; code < dac.Levels(); code++ {
				step := dac.ConvertWithNonlinearity(code) - dac.ConvertWithNonlinearity(code-1)
				dnl := step/lsb - 1.0
				if math.Abs(dnl) > 1.0 {
					t.Fatalf("DAC DNL out of bounds at bits=%d code=%d: DNL=%.6f LSB", bits, code, dnl)
				}
			}
		})
	}
}

func TestLinearity_ADCINLWithinOneLSBIdeal(t *testing.T) {
	for _, bits := range []int{4, 6, 8} {
		t.Run("bits="+strconv.Itoa(bits), func(t *testing.T) {
			adc := testADC(bits)
			lsb := adc.Resolution()

			for code := 0; code < adc.Levels(); code++ {
				vin := adc.VrefLow + float64(code)*lsb
				actual := adc.Convert(vin)
				inl := float64(actual - code)
				if math.Abs(inl) > 1.0 {
					t.Fatalf("ADC INL out of bounds at bits=%d code=%d: INL=%.6f LSB", bits, code, inl)
				}
			}
		})
	}
}

func TestLinearity_DACADCRoundTripWithinOneLSB(t *testing.T) {
	for _, bits := range []int{4, 6, 8} {
		t.Run("bits="+strconv.Itoa(bits), func(t *testing.T) {
			dac := testDAC(bits)
			adc := testADC(bits)

			for code := 0; code < dac.Levels(); code++ {
				vout := dac.Convert(code)
				read := adc.Convert(vout)
				err := int(math.Abs(float64(read - code)))
				if err > 1 {
					t.Fatalf("roundtrip error exceeds 1 LSB at bits=%d code=%d: adc=%d err=%d", bits, code, read, err)
				}
			}
		})
	}
}
