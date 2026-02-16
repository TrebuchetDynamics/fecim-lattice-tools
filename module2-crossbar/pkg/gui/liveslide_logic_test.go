package gui

import (
	"strings"
	"testing"

	fyneTest "fyne.io/fyne/v2/test"
)

func TestEducationalPanelAndLogLogic(t *testing.T) {
	a := fyneTest.NewApp()
	defer a.Quit()

	ep := NewEducationalPanel(EducationalPanelConfig{})
	ep.SetMVMExplanation(0)
	title, content := ep.GetContent()
	if title != "Compute-in-Memory" || !strings.Contains(content, "Matrix-Vector") {
		t.Fatalf("unexpected MVM content: title=%q content=%q", title, content)
	}
	ep.SetMVMExplanation(1)
	ep.SetMVMExplanation(2)
	ep.SetIRDropExplanation()
	title, _ = ep.GetContent()
	if !strings.Contains(title, "IR Drop") {
		t.Fatalf("unexpected IR drop title: %q", title)
	}
	ep.SetSneakPathExplanation()
	ep.SetIdleExplanation()
	_, content = ep.GetContent()
	if !strings.Contains(content, "CROSSBAR") {
		t.Fatalf("unexpected idle content: %q", content)
	}

	log := NewOperationLog(OperationLogConfig{})
	for i := 0; i < 10; i++ {
		log.Add("op")
	}
	if len(log.GetEntries()) != log.GetMaxEntries() {
		t.Fatalf("expected entries capped to %d, got %d", log.GetMaxEntries(), len(log.GetEntries()))
	}
	log.AddWithResult("mvm", "ok", true)
	log.AddWithResult("mvm", "err", false)
	log.Clear()
	if len(log.GetEntries()) != 0 {
		t.Fatal("expected clear to remove entries")
	}
}

func TestInputOutputDisplayFormatting(t *testing.T) {
	a := fyneTest.NewApp()
	defer a.Quit()

	d := NewInputOutputDisplay()
	short := d.formatVector([]float64{0.1, 0.2}, "V")
	if !strings.Contains(short, "V") {
		t.Fatalf("expected prefix in short format: %q", short)
	}
	long := d.formatVector([]float64{1, 2, 3, 4, 5, 6, 7}, "I")
	if !strings.Contains(long, "...") {
		t.Fatalf("expected ellipsis for long vector: %q", long)
	}

	d.SetInput([]float64{0.1, 0.2, 0.3})
	d.SetOutput([]float64{0.4, 0.5, 0.6})
}

func TestDemoModeStringsAndModeIndicator(t *testing.T) {
	a := fyneTest.NewApp()
	defer a.Quit()

	cases := map[DemoMode]string{
		DemoModeIdle:      "IDLE",
		DemoModeCompute:   "COMPUTE",
		DemoModeWrite:     "WRITE",
		DemoModeRead:      "READ",
		DemoModeIRDrop:    "IR DROP",
		DemoModeSneakPath: "SNEAK",
	}
	for m, want := range cases {
		if got := m.String(); got != want {
			t.Fatalf("mode %v string=%q want=%q", m, got, want)
		}
	}
	if got := DemoMode(99).String(); got != "UNKNOWN" {
		t.Fatalf("unexpected unknown mode string: %q", got)
	}

	mi := NewModeIndicatorBox()
	mi.SetMode(DemoModeRead)
	if mi.GetMode() != DemoModeRead {
		t.Fatalf("expected mode indicator to store mode, got %v", mi.GetMode())
	}
}
