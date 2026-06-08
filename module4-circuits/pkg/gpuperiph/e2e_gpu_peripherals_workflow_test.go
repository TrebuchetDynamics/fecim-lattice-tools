package gpuperiph

import (
	"strings"
	"testing"
	"unsafe"
)

func TestModule4GPUPeriphE2EAvailabilityLayoutAndBatchWorkflow(t *testing.T) {
	if err := validateGPUPeripheralStructLayout(); err != nil {
		t.Fatalf("GPU peripheral struct layout invalid: %v", err)
	}
	for name, payload := range map[string]any{
		"dac": DefaultDACParams(4),
		"adc": DefaultADCParams(4),
		"tia": DefaultTIAParams(4),
	} {
		t.Run("encode-"+name, func(t *testing.T) {
			var raw []byte
			var err error
			switch v := payload.(type) {
			case DACParams:
				raw, err = structToBytes(&v)
			case ADCParams:
				raw, err = structToBytes(&v)
			case TIAParams:
				raw, err = structToBytes(&v)
			}
			if err != nil || len(raw) != 32 {
				t.Fatalf("structToBytes(%s) len/err = %d/%v", name, len(raw), err)
			}
		})
	}
	if _, err := structToBytes(struct{}{}); err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("unsupported structToBytes should fail with context, err=%v", err)
	}
	if unsafe.Sizeof(DACParams{}) != 32 || unsafe.Sizeof(ADCParams{}) != 32 || unsafe.Sizeof(TIAParams{}) != 32 {
		t.Fatalf("unexpected GPU param struct sizes dac=%d adc=%d tia=%d", unsafe.Sizeof(DACParams{}), unsafe.Sizeof(ADCParams{}), unsafe.Sizeof(TIAParams{}))
	}

	g, err := NewGPUPeripherals()
	if err != nil {
		t.Fatalf("NewGPUPeripherals: %v", err)
	}
	defer g.Destroy()

	codes := []int32{0, 7, 15, 31}
	voltages := []float32{0, 0.25, 0.5, 1.0}
	currents := []float32{-1e-6, 0, 10e-6, 200e-6}
	if !g.IsAvailable() {
		if out, err := g.BatchDAC(codes, DefaultDACParams(len(codes))); err == nil || out != nil || !strings.Contains(err.Error(), "not available") {
			t.Fatalf("BatchDAC unavailable contract changed: out=%v err=%v", out, err)
		}
		if codeOut, quantOut, err := g.BatchADC(voltages, DefaultADCParams(len(voltages))); err == nil || codeOut != nil || quantOut != nil || !strings.Contains(err.Error(), "not available") {
			t.Fatalf("BatchADC unavailable contract changed: codes=%v quant=%v err=%v", codeOut, quantOut, err)
		}
		if out, err := g.BatchTIA(currents, DefaultTIAParams(len(currents))); err == nil || out != nil || !strings.Contains(err.Error(), "not available") {
			t.Fatalf("BatchTIA unavailable contract changed: out=%v err=%v", out, err)
		}
		return
	}

	dacOut, err := g.BatchDAC(codes, DefaultDACParams(len(codes)))
	if err != nil || len(dacOut) != len(codes) {
		t.Fatalf("BatchDAC available len/err = %d/%v", len(dacOut), err)
	}
	adcCodes, adcQuant, err := g.BatchADC(voltages, DefaultADCParams(len(voltages)))
	if err != nil || len(adcCodes) != len(voltages) || len(adcQuant) != len(voltages) {
		t.Fatalf("BatchADC available lens/err = %d/%d/%v", len(adcCodes), len(adcQuant), err)
	}
	tiaOut, err := g.BatchTIA(currents, DefaultTIAParams(len(currents)))
	if err != nil || len(tiaOut) != len(currents) {
		t.Fatalf("BatchTIA available len/err = %d/%v", len(tiaOut), err)
	}
	if empty, err := g.BatchDAC(nil, DefaultDACParams(0)); err != nil || len(empty) != 0 {
		t.Fatalf("BatchDAC empty contract = %v/%v", empty, err)
	}
	if c, q, err := g.BatchADC(nil, DefaultADCParams(0)); err != nil || len(c) != 0 || len(q) != 0 {
		t.Fatalf("BatchADC empty contract = %v/%v/%v", c, q, err)
	}
	if empty, err := g.BatchTIA(nil, DefaultTIAParams(0)); err != nil || len(empty) != 0 {
		t.Fatalf("BatchTIA empty contract = %v/%v", empty, err)
	}
}
