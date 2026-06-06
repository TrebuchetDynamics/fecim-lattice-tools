package tui

import (
	"math"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModule1TUIE2EWideInteractiveWorkflow(t *testing.T) {
	m := NewModel()
	if m.Init() == nil {
		t.Fatal("Init() returned nil command")
	}

	m = applyTUIE2EMsg(t, m, tea.WindowSizeMsg{Width: 120, Height: 45})
	if m.plotWidth != 80 || m.plotHeight != 25 {
		t.Fatalf("window sizing = %dx%d, want capped 80x25", m.plotWidth, m.plotHeight)
	}

	m.lastTick = time.Now().Add(-50 * time.Millisecond)
	m = applyTUIE2EMsg(t, m, tickMsg(time.Now()))
	assertTUIE2EStateFinite(t, m)
	if len(m.eHistory) != 1 || len(m.pHistory) != 1 || m.waveform != WaveformSine || !m.autoMode {
		t.Fatalf("initial tick state = history %d/%d waveform=%s auto=%v", len(m.eHistory), len(m.pHistory), m.waveform, m.autoMode)
	}

	sequence := []struct {
		key          string
		wantWave     WaveformType
		wantAuto     bool
		wantPaused   bool
		wantHelp     bool
		wantMaterial bool
	}{
		{key: "tab", wantWave: WaveformTriangle, wantAuto: true},
		{key: "tab", wantWave: WaveformSquare, wantAuto: true},
		{key: "tab", wantWave: WaveformManual, wantAuto: true},
		{key: "a", wantWave: WaveformManual, wantAuto: false},
		{key: "up", wantWave: WaveformManual, wantAuto: false},
		{key: "right", wantWave: WaveformManual, wantAuto: false},
		{key: "down", wantWave: WaveformManual, wantAuto: false},
		{key: "left", wantWave: WaveformManual, wantAuto: false},
		{key: " ", wantWave: WaveformManual, wantAuto: false, wantPaused: true},
		{key: "?", wantWave: WaveformManual, wantAuto: false, wantPaused: true, wantHelp: true},
		{key: "m", wantWave: WaveformManual, wantAuto: false, wantPaused: true, wantHelp: true, wantMaterial: true},
	}
	initialMaterial := m.material.Name
	for _, step := range sequence {
		m = applyTUIE2EKey(t, m, step.key)
		if m.waveform != step.wantWave || m.autoMode != step.wantAuto || m.paused != step.wantPaused || m.showHelp != step.wantHelp {
			t.Fatalf("after key %q state = waveform=%s auto=%v paused=%v help=%v", step.key, m.waveform, m.autoMode, m.paused, m.showHelp)
		}
		if step.wantMaterial && len(m.materials) > 1 && m.material.Name == initialMaterial {
			t.Fatalf("material key did not cycle material from %q", initialMaterial)
		}
		assertTUIE2EStateFinite(t, m)
	}

	historyBeforePauseTick := len(m.eHistory)
	m.lastTick = time.Now().Add(-50 * time.Millisecond)
	m = applyTUIE2EMsg(t, m, tickMsg(time.Now()))
	if len(m.eHistory) != historyBeforePauseTick {
		t.Fatalf("paused tick changed history length from %d to %d", historyBeforePauseTick, len(m.eHistory))
	}

	m = applyTUIE2EKey(t, m, "r")
	if m.electricField != 0 || m.polarization != 0 || m.normalizedP != 0 || m.discreteLevel != 15 || len(m.eHistory) != 0 || len(m.pHistory) != 0 || m.simTime != 0 {
		t.Fatalf("reset left dirty state: E=%g P=%g norm=%g level=%d hist=%d/%d time=%g", m.electricField, m.polarization, m.normalizedP, m.discreteLevel, len(m.eHistory), len(m.pHistory), m.simTime)
	}

	view := m.View()
	for _, marker := range []string{"FeCIM Hysteresis Visualizer", "Material:", "Discrete Level:", "30-Level Simulation Baseline", "Press [q] to quit"} {
		if !strings.Contains(view, marker) {
			t.Fatalf("View() missing marker %q\n%s", marker, view)
		}
	}
}

func TestModule1TUIE2EHistoryRenderAndMaterialMatrix(t *testing.T) {
	base := NewModel()
	if len(base.materials) == 0 {
		t.Fatal("no TUI materials available")
	}
	limit := len(base.materials)
	if limit > 4 {
		limit = 4
	}
	waveforms := []WaveformType{WaveformSine, WaveformTriangle, WaveformSquare, WaveformManual}

	for materialIdx := 0; materialIdx < limit; materialIdx++ {
		materialName := base.materials[materialIdx].Name
		for _, waveform := range waveforms {
			t.Run(materialName+"/"+waveform.String(), func(t *testing.T) {
				m := NewModelWithMaterial(materialName)
				m.waveform = waveform
				m.autoMode = waveform != WaveformManual
				m.plotWidth = 42
				m.plotHeight = 16
				if waveform == WaveformManual {
					m.electricField = -0.5 * m.material.Ec
				}

				for i := 0; i < m.maxHistory+35; i++ {
					m.lastTick = time.Now().Add(-50 * time.Millisecond)
					m.updateSimulation()
				}
				assertTUIE2EStateFinite(t, m)
				if len(m.eHistory) != m.maxHistory || len(m.pHistory) != m.maxHistory {
					t.Fatalf("history window = %d/%d, want %d", len(m.eHistory), len(m.pHistory), m.maxHistory)
				}

				plot := m.renderPEPlot()
				info := m.renderInfoPanel()
				bar := m.renderLevelBar()
				status := m.renderStatusBar()
				for name, rendered := range map[string]string{"plot": plot, "info": info, "bar": bar, "status": status} {
					if strings.TrimSpace(rendered) == "" {
						t.Fatalf("%s render is empty", name)
					}
				}
				for _, marker := range []string{materialName, waveform.String(), "Discrete Level:", "Auto Mode:"} {
					if !strings.Contains(info, marker) {
						t.Fatalf("info panel missing %q\n%s", marker, info)
					}
				}
				if !strings.Contains(plot, "P (µC/cm²)") || !strings.Contains(plot, "E (MV/cm)") {
					t.Fatalf("plot missing axes labels\n%s", plot)
				}
				if !strings.Contains(bar, "30-Level Simulation Baseline") || !strings.Contains(bar, "bits") {
					t.Fatalf("level bar missing baseline markers\n%s", bar)
				}
				if !strings.Contains(status, "Press [q] to quit") {
					t.Fatalf("status missing quit hint\n%s", status)
				}
			})
		}
	}
}

func TestModule1TUIE2EWindowAndUnknownMaterialBoundaries(t *testing.T) {
	m := NewModelWithMaterial("not a real material")
	if m.matIndex != 0 || m.material == nil {
		t.Fatalf("unknown material should fall back to first material, idx=%d material=%v", m.matIndex, m.material)
	}

	m = applyTUIE2EMsg(t, m, tea.WindowSizeMsg{Width: 55, Height: 28})
	if m.plotWidth != 25 || m.plotHeight != 13 {
		t.Fatalf("medium window sizing = %dx%d, want 25x13", m.plotWidth, m.plotHeight)
	}
	m = applyTUIE2EMsg(t, m, tea.WindowSizeMsg{Width: 200, Height: 100})
	if m.plotWidth != 80 || m.plotHeight != 25 {
		t.Fatalf("large window sizing = %dx%d, want capped 80x25", m.plotWidth, m.plotHeight)
	}

	for _, key := range []string{"enter", "x", "ctrl+z"} {
		before := tuiE2ESnapshot(m)
		m = applyTUIE2EKey(t, m, key)
		after := tuiE2ESnapshot(m)
		if before != after {
			t.Fatalf("unhandled key %q changed snapshot from %+v to %+v", key, before, after)
		}
	}
}

type tuiE2EComparableSnapshot struct {
	Material      string
	Waveform      WaveformType
	AutoMode      bool
	Paused        bool
	ShowHelp      bool
	ElectricField float64
	Level         int
	History       int
	PlotWidth     int
	PlotHeight    int
}

func tuiE2ESnapshot(m Model) tuiE2EComparableSnapshot {
	material := ""
	if m.material != nil {
		material = m.material.Name
	}
	return tuiE2EComparableSnapshot{
		Material:      material,
		Waveform:      m.waveform,
		AutoMode:      m.autoMode,
		Paused:        m.paused,
		ShowHelp:      m.showHelp,
		ElectricField: m.electricField,
		Level:         m.discreteLevel,
		History:       len(m.eHistory),
		PlotWidth:     m.plotWidth,
		PlotHeight:    m.plotHeight,
	}
}

func applyTUIE2EKey(t *testing.T, m Model, key string) Model {
	t.Helper()
	return applyTUIE2EMsg(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key), Alt: false})
}

func applyTUIE2EMsg(t *testing.T, m Model, msg tea.Msg) Model {
	t.Helper()
	updated, _ := m.Update(msg)
	next, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update(%T) returned %T, want tui.Model", msg, updated)
	}
	return next
}

func assertTUIE2EStateFinite(t *testing.T, m Model) {
	t.Helper()
	values := map[string]float64{
		"electricField": m.electricField,
		"polarization":  m.polarization,
		"normalizedP":   m.normalizedP,
		"simTime":       m.simTime,
	}
	for name, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			t.Fatalf("%s invalid: %g", name, value)
		}
	}
	if m.discreteLevel < 0 || m.discreteLevel > 29 {
		t.Fatalf("discreteLevel = %d, want [0,29]", m.discreteLevel)
	}
	if len(m.eHistory) != len(m.pHistory) || len(m.eHistory) > m.maxHistory {
		t.Fatalf("history lengths invalid: E=%d P=%d max=%d", len(m.eHistory), len(m.pHistory), m.maxHistory)
	}
}
