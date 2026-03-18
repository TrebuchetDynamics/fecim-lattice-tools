package peripherals

import "math"

const (
	boltzmannConstantJPerK = 1.380649e-23
	electronChargeC        = 1.602176634e-19
)

// ThermalNoiseRMS computes Johnson-Nyquist thermal noise voltage RMS: sqrt(4*k*T*R*BW).
func ThermalNoiseRMS(tempK, resistanceOhm, bandwidthHz float64) float64 {
	if tempK <= 0 || resistanceOhm <= 0 || bandwidthHz <= 0 {
		return 0
	}
	return math.Sqrt(4 * boltzmannConstantJPerK * tempK * resistanceOhm * bandwidthHz)
}

// FlickerNoisePower computes 1/f (flicker) noise power using K/f.
func FlickerNoisePower(k, frequencyHz float64) float64 {
	if k <= 0 || frequencyHz <= 0 {
		return 0
	}
	return k / frequencyHz
}

// ShotNoiseCurrentRMS computes shot noise current RMS: sqrt(2*q*I*BW).
func ShotNoiseCurrentRMS(currentA, bandwidthHz float64) float64 {
	if currentA < 0 {
		currentA = -currentA
	}
	if currentA == 0 || bandwidthHz <= 0 {
		return 0
	}
	return math.Sqrt(2 * electronChargeC * currentA * bandwidthHz)
}

// QuantizationNoiseVariance computes quantization noise variance for a uniform quantizer.
// Uses lsb = vRefSpan / (2^N - 1) to match ADC Resolution() which maps
// the reference span across (2^N - 1) intervals between 2^N levels.
func QuantizationNoiseVariance(vRefSpan float64, bits int) float64 {
	if bits <= 0 || vRefSpan <= 0 {
		return 0
	}
	levels := math.Pow(2, float64(bits)) - 1
	if levels <= 0 {
		levels = 1
	}
	lsb := vRefSpan / levels
	return (lsb * lsb) / 12.0
}

// TotalNoiseVariance sums independent RMS contributors using variance additivity.
func TotalNoiseVariance(sigmas ...float64) float64 {
	total := 0.0
	for _, s := range sigmas {
		total += s * s
	}
	return total
}

// SNRDB computes signal-to-noise ratio in dB from RMS values.
func SNRDB(signalRMS, noiseRMS float64) float64 {
	if noiseRMS == 0 {
		if signalRMS == 0 {
			return math.NaN()
		}
		return math.Inf(1)
	}
	if signalRMS == 0 {
		return math.Inf(-1)
	}
	return 20 * math.Log10(signalRMS/noiseRMS)
}
