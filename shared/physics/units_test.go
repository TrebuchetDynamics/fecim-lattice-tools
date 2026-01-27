package physics

import "testing"

func TestFormatEnergy(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0, "0 J"},
		{-1, "0 J"},
		{1e-15, "1.00 fJ"},
		{1.5e-15, "1.50 fJ"},
		{1e-12, "1.00 pJ"},
		{2.3e-12, "2.30 pJ"},
		{1e-9, "1.00 nJ"},
		{4.5e-9, "4.50 nJ"},
		{1e-6, "1.00 µJ"},
		{6.7e-6, "6.70 µJ"},
		{1e-3, "1.00 mJ"},
		{8.9e-3, "8.90 mJ"},
		{1.0, "1.00 J"},
		{1.2, "1.20 J"},
		{1000, "1000.00 J"},
	}

	for _, tt := range tests {
		result := FormatEnergy(tt.input)
		if result != tt.expected {
			t.Errorf("FormatEnergy(%e) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatEnergyMJ(t *testing.T) {
	// 1 mJ = 1e-3 J
	result := FormatEnergyMJ(1.0)
	expected := "1.00 mJ"
	if result != expected {
		t.Errorf("FormatEnergyMJ(1.0) = %q, want %q", result, expected)
	}

	// 0.001 mJ = 1 µJ
	result = FormatEnergyMJ(0.001)
	expected = "1.00 µJ"
	if result != expected {
		t.Errorf("FormatEnergyMJ(0.001) = %q, want %q", result, expected)
	}
}

func TestFormatEnergyUJ(t *testing.T) {
	// 1 µJ = 1e-6 J
	result := FormatEnergyUJ(1.0)
	expected := "1.00 µJ"
	if result != expected {
		t.Errorf("FormatEnergyUJ(1.0) = %q, want %q", result, expected)
	}

	// 1000 µJ = 1 mJ
	result = FormatEnergyUJ(1000)
	expected = "1.00 mJ"
	if result != expected {
		t.Errorf("FormatEnergyUJ(1000) = %q, want %q", result, expected)
	}
}

func TestFormatConductance(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0, "0 S"},
		{1e-9, "1.00 nS"},
		{50e-6, "50.00 µS"},
		{1e-3, "1.00 mS"},
		{1.0, "1.00 S"},
	}

	for _, tt := range tests {
		result := FormatConductance(tt.input)
		if result != tt.expected {
			t.Errorf("FormatConductance(%e) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatCurrent(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0, "0 A"},
		{1e-12, "1.00 pA"},
		{50e-9, "50.00 nA"},
		{1e-6, "1.00 µA"},
		{1e-3, "1.00 mA"},
		{1.0, "1.00 A"},
	}

	for _, tt := range tests {
		result := FormatCurrent(tt.input)
		if result != tt.expected {
			t.Errorf("FormatCurrent(%e) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatVoltage(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0, "0 V"},
		{1e-6, "1.00 µV"},
		{1e-3, "1.00 mV"},
		{1.5, "1.50 V"},
	}

	for _, tt := range tests {
		result := FormatVoltage(tt.input)
		if result != tt.expected {
			t.Errorf("FormatVoltage(%e) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatTime(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0, "0 s"},
		{1e-12, "1.00 ps"},
		{1e-9, "1.00 ns"},
		{1e-6, "1.00 µs"},
		{1e-3, "1.00 ms"},
		{1.0, "1.00 s"},
	}

	for _, tt := range tests {
		result := FormatTime(tt.input)
		if result != tt.expected {
			t.Errorf("FormatTime(%e) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatFrequency(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0, "0 Hz"},
		{100, "100.00 Hz"},
		{1e3, "1.00 kHz"},
		{1e6, "1.00 MHz"},
		{1e9, "1.00 GHz"},
	}

	for _, tt := range tests {
		result := FormatFrequency(tt.input)
		if result != tt.expected {
			t.Errorf("FormatFrequency(%e) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
