package units

import "testing"

func TestElectricFieldConversions(t *testing.T) {
	if VPerMPerMVPerCm != 1e8 {
		t.Fatalf("VPerMPerMVPerCm = %g, want 1e8", VPerMPerMVPerCm)
	}
	if got := VPerMToMVPerCm(2.5e8); got != 2.5 {
		t.Fatalf("VPerMToMVPerCm = %g, want 2.5", got)
	}
	if got := MVPerCmToVPerM(1.5); got != 1.5e8 {
		t.Fatalf("MVPerCmToVPerM = %g, want 1.5e8", got)
	}
}

func TestFormatSIHelpers(t *testing.T) {
	checks := map[string]string{
		FormatEnergy(1.5e-15):      "1.50 fJ",
		FormatConductance(50e-6):   "50.00 µS",
		FormatCurrent(50e-9):       "50.00 nA",
		FormatVoltage(1e-3):        "1.00 mV",
		FormatTime(1e-6):           "1.00 µs",
		FormatFrequency(1e6):       "1.00 MHz",
		FormatResistance(4700):     "4.70 kΩ",
		FormatCapacitance(47e-12):  "47.00 pF",
		FormatPower(1500):          "1.50 kW",
		FormatCharge(50e-12):       "50.00 pC",
		FormatPolarization(0.20):   "20.0 µC/cm²",
		FormatElectricField(1.5e8): "1.50 MV/cm",
	}
	for got, want := range checks {
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	}
}

func TestEnergyConvenienceWrappers(t *testing.T) {
	if got := FormatEnergyMJ(0.001); got != "1.00 µJ" {
		t.Fatalf("FormatEnergyMJ = %q, want 1.00 µJ", got)
	}
	if got := FormatEnergyUJ(1000); got != "1.00 mJ" {
		t.Fatalf("FormatEnergyUJ = %q, want 1.00 mJ", got)
	}
}
