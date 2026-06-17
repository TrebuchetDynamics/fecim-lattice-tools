package gui

import (
	"math"
	"testing"
)

// TestLKSubStepSize_FrequencyAdaptiveCap verifies that lkSubStepSize caps the
// step to T/50 for periodic waveforms at high frequencies, preventing aliasing.
//
// Bug: with dtNominal=1e-4 and f=1MHz (T=1µs), sin(2π·f·n·dt) = sin(200πn) = 0
// at every integer step n → E-field appears frozen. Fix: cap step to T/50=20ns.
func TestLKSubStepSize_FrequencyAdaptiveCap(t *testing.T) {
	tests := []struct {
		name       string
		freq       float64
		waveform   WaveformType
		wantStep   float64
		desc       string
	}{
		{
			name:     "1MHz_sine_capped_to_dtMin",
			freq:     1e6,
			waveform: WaveformSine,
			wantStep: dtMin, // T/50=20ns < dtMin=1µs → floor at dtMin
			desc:     "1MHz: T/50=20ns < dtMin; cap floors at dtMin=1µs (still avoids full dtNominal aliasing)",
		},
		{
			name:     "100kHz_sine_capped_to_dtMin",
			freq:     1e5,
			waveform: WaveformSine,
			wantStep: dtMin, // T/50=200ns < dtMin=1µs → floor at dtMin
			desc:     "100kHz: T/50=200ns < dtMin; cap floors at dtMin=1µs",
		},
		{
			name:     "1kHz_sine_capped_to_T_over_50",
			freq:     1e3,
			waveform: WaveformSine,
			wantStep: 1.0 / (50.0 * 1e3), // 20µs — within [dtMin, dtNominal]
			desc:     "1kHz: T/50=20µs is between dtMin and dtNominal; cap applies exactly",
		},
		{
			name:     "200Hz_boundary_no_cap",
			freq:     200.0,
			waveform: WaveformSine,
			wantStep: dtNominal, // T/50 = 100µs = dtNominal → no cap
			desc:     "At 200Hz T/50=100µs equals dtNominal; cap is not binding",
		},
		{
			name:     "low_freq_no_cap",
			freq:     0.5,
			waveform: WaveformSine,
			wantStep: dtNominal,
			desc:     "Default 0.5Hz: dtNominal is fine, no cap needed",
		},
		{
			name:     "1MHz_triangle_capped_to_dtMin",
			freq:     1e6,
			waveform: WaveformTriangle,
			wantStep: dtMin, // T/50=20ns < dtMin → floor at dtMin
			desc:     "Triangle waveform at 1MHz also caps to dtMin",
		},
		{
			name:     "1MHz_wrd_no_cap",
			freq:     1e6,
			waveform: WaveformWriteReadDemo,
			wantStep: dtNominal,
			desc:     "WRD waveform uses state-machine timing; cap not applied",
		},
		{
			name:     "1MHz_manual_no_cap",
			freq:     1e6,
			waveform: WaveformManual,
			wantStep: dtNominal,
			desc:     "Manual waveform is not periodic; cap not applied",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := lkSubStepSize(tc.freq, tc.waveform, 0 /*matEc=0 → no Ec proximity*/, 0 /*E=0*/)
			if math.Abs(got-tc.wantStep)/tc.wantStep > 1e-9 {
				t.Errorf("%s\n  got step=%.3e, want %.3e", tc.desc, got, tc.wantStep)
			}
		})
	}
}

// TestLKSubStepSize_EcProximityOverridesFreqCap verifies that near-Ec switching
// gets dtMin regardless of frequency. This preserves capture of sharp switching.
func TestLKSubStepSize_EcProximityOverridesFreqCap(t *testing.T) {
	const matEc = 1e8 // 1 MV/cm
	const freqHz = 1e6
	// Place E-field just inside the ±10% Ec proximity band
	nearEc := matEc * (1 - ecProximityFrac/2)

	got := lkSubStepSize(freqHz, WaveformSine, matEc, nearEc)
	if got != dtMin {
		t.Errorf("near Ec: got step=%.3e, want dtMin=%.3e (Ec proximity must win over freq cap)", got, dtMin)
	}
}

// TestLKSubStepSize_ZeroFrequencyNoDiv verifies no divide-by-zero for f=0.
func TestLKSubStepSize_ZeroFrequencyNoDiv(t *testing.T) {
	got := lkSubStepSize(0, WaveformSine, 0, 0)
	if got != dtNominal {
		t.Errorf("freq=0: got step=%.3e, want dtNominal=%.3e", got, dtNominal)
	}
}
