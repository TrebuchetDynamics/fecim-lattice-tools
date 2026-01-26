// Package gui provides info panel creation and management for the hysteresis demo.
package gui

import (
	"fmt"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"multilayer-ferroelectric-cim-visualizer/module1-hysteresis/pkg/gui/widgets"
)

// createInfoPanel creates the state and material information panel
func (a *App) createInfoPanel() fyne.CanvasObject {
	a.pLabel = widget.NewLabel("P: 0.00 µC/cm²")
	a.levelLabel = widget.NewLabel("Level: 15/30")
	a.modeIndicator = widgets.NewModeIndicator()
	a.modeIndicator.SetMinSize(fyne.NewSize(180, 60))

	// State display - compact
	stateGrid := container.NewGridWithColumns(2,
		widget.NewLabel("P:"), a.pLabel,
		widget.NewLabel("Level:"), a.levelLabel,
	)

	// Material params - single line
	matParams := widget.NewLabel(fmt.Sprintf(
		"Pr=%.0f Ps=%.0f Ec=%.2f  End: %.0e",
		a.material.Pr*100, a.material.Ps*100,
		a.material.Ec/1e8,
		a.material.EnduranceCycles,
	))
	matParams.Wrapping = fyne.TextWrapOff

	// Wake-up/Fatigue display
	a.cyclesLabel = widget.NewLabel("0")
	a.wakeupLabel = widget.NewLabel("80%")
	a.fatigueLabel = widget.NewLabel("0.0%")

	fatigueRow := container.NewHBox(
		widget.NewLabel("Cyc:"), a.cyclesLabel,
		widget.NewLabel("Wake:"), a.wakeupLabel,
		widget.NewLabel("Fat:"), a.fatigueLabel,
	)

	return container.NewVBox(
		stateGrid,
		a.modeIndicator,
		matParams,
		fatigueRow,
	)
}

// createSlidePanel creates the explanation panel
func (a *App) createSlidePanel() fyne.CanvasObject {
	a.slideTitle = widget.NewLabel("") // Keep for compatibility
	a.slideText = widget.NewLabel(a.getSlideText())
	a.slideText.Wrapping = fyne.TextWrapWord
	return a.slideText
}

// createLogPanel creates the memory operations log panel
func (a *App) createLogPanel() fyne.CanvasObject {
	a.logText = widget.NewLabel("Waiting...")
	a.logText.Wrapping = fyne.TextWrapWord
	return a.logText
}

// getSlideText returns the contextual explanation text based on current waveform mode
func (a *App) getSlideText() string {
	a.mu.RLock()
	level := a.discreteLevel + 1 // 1-indexed for display
	wrdPhase := a.wrdPhase
	wrdTarget := a.wrdTargetLevel
	isWrite := math.Abs(a.electricField) > a.material.Ec
	waveform := a.waveform
	wrdTotalWrites := a.wrdTotalWrites
	wrdSuccessWrites := a.wrdSuccessWrites
	wrdTotalEnergyfJ := a.wrdTotalEnergyfJ
	a.mu.RUnlock()

	switch waveform {
	case WaveformManual:
		a.mu.RLock()
		animating := a.manualAnimating
		a.mu.RUnlock()

		if animating {
			return fmt.Sprintf("🎯 WRITING LEVEL %d\n\n"+
				"Click-to-level animation:\n"+
				"SATURATE → SETTLE → HOLD\n\n"+
				"Watch the P-E plot trace\n"+
				"the physics path!\n\n"+
				"Click any level on the bar\n"+
				"to program a new value.", level)
		}
		if isWrite {
			return fmt.Sprintf("██ WRITING LEVEL %d ██\n\n"+
				"Electric field E > Ec.\n"+
				"Domains are switching.\n"+
				"Polarization is changing.\n\n"+
				"Use slider OR click\n"+
				"level bar to program!", level)
		}
		return fmt.Sprintf("░░ HOLDING LEVEL %d ░░\n\n"+
			"E-field is low or zero.\n"+
			"Polarization PERSISTS.\n"+
			"No power needed.\n\n"+
			"MANUAL MODE:\n"+
			"• Drag slider to apply E-field\n"+
			"• Click level bar to auto-program", level)

	case WaveformSine, WaveformTriangle:
		phaseText := "░░ READING ░░"
		if isWrite {
			phaseText = "██ WRITING ██"
		}
		return fmt.Sprintf("%s\n\n"+
			"Level: %d/30\n\n"+
			"The P-E loop shows hysteresis:\n"+
			"• Upper branch: E increasing\n"+
			"• Lower branch: E decreasing\n"+
			"• Area inside = energy loss\n\n"+
			"The SQUARE shape means:\n"+
			"sharp switching at ±Ec.", phaseText, level)

	case WaveformWriteReadDemo:
		// Calculate stats (using local copies from RLock above)
		successRate := 0.0
		if wrdTotalWrites > 0 {
			successRate = float64(wrdSuccessWrites) / float64(wrdTotalWrites) * 100
		}
		energyPerOp := 10.0 // ~10 fJ per operation (FeFET switching energy)

		var phaseExplanation string
		switch wrdPhase {
		case 0: // SATURATE
			phaseExplanation = fmt.Sprintf("▓▓ SATURATE → %d ▓▓\n\n"+
				"|E| = 2×Ec (maximum field)\n"+
				"ALL ferroelectric domains\n"+
				"switching simultaneously.\n\n"+
				"Energy: ~%.0f fJ\n"+
				"(10M× less than NAND!)\n\n"+
				"\"The same device does memory\n"+
				"AND computation.\" - Dr. Tour", wrdTarget, energyPerOp)
		case 1: // SETTLE
			phaseExplanation = fmt.Sprintf("██ SETTLE → %d ██\n\n"+
				"MINOR LOOP formation:\n"+
				"Partial domain reversal\n"+
				"sets the analog level.\n\n"+
				"This is how we store\n"+
				"4.91 BITS in ONE cell!\n"+
				"(Binary = only 1 bit)\n\n"+
				"30 levels = 5× density", wrdTarget)
		case 2: // HOLD
			phaseExplanation = fmt.Sprintf("░░ HOLD LEVEL %d ░░\n\n"+
				"E = 0, P persists!\n\n"+
				"ZERO POWER NEEDED.\n"+
				"Data retention: 10+ years\n"+
				"(demonstrated: 10⁷ sec)\n\n"+
				"This is TRUE non-volatile:\n"+
				"No refresh like DRAM.\n"+
				"No charge leakage.\n\n"+
				"Just ferroelectric\n"+
				"polarization.", level)
		case 3: // READ
			phaseExplanation = fmt.Sprintf("▒▒ READING LEVEL %d ▒▒\n\n"+
				"Sense pulse: |E| < Ec\n"+
				"State UNCHANGED!\n\n"+
				"Non-destructive read:\n"+
				"Unlike NAND, data stays.\n"+
				"No rewrite needed.\n\n"+
				"Read energy: ~%.0f fJ\n"+
				"(1000× less than DRAM)", level, energyPerOp*0.1)
		case 4: // DISPLAY
			status := "✓ SUCCESS"
			accuracy := ""
			if level != wrdTarget {
				status = fmt.Sprintf("△ Level %d (target %d)", level, wrdTarget)
				accuracy = "\n(Within ±1 is normal)"
			} else {
				accuracy = "\nPerfect analog storage!"
			}
			phaseExplanation = fmt.Sprintf("%s%s\n\n"+
				"─── SESSION STATS ───\n"+
				"Writes: %d\n"+
				"Success: %.0f%%\n"+
				"Energy: %.1f pJ total\n\n"+
				"Each write: ~10 fJ\n"+
				"Each read: ~1 fJ\n"+
				"─────────────────\n"+
				"Next level coming...", status, accuracy,
				wrdTotalWrites, successRate, wrdTotalEnergyfJ/1000)
		}
		// Add Dr. Tour footer
		return phaseExplanation + "\n\n═══════════════════\n" +
			"FeCIM ADVANTAGE:\n" +
			"• 30 levels = 4.9 bits/cell\n" +
			"• 10M× better than NAND\n" +
			"• 1000× better than DRAM\n" +
			"• CMOS compatible"

	default:
		return "Select a waveform mode\nto see explanation."
	}
}

// addLogEntry adds a timestamped entry to the memory log
func (a *App) addLogEntry(entry string) {
	// Add timestamp prefix
	timestamp := fmt.Sprintf("t=%.1fs", a.simTime)
	fullEntry := fmt.Sprintf("%s %s", timestamp, entry)
	a.logEntries = append(a.logEntries, fullEntry)
	if len(a.logEntries) > a.maxLogLines {
		a.logEntries = a.logEntries[1:]
	}
}

// getLogText returns the formatted log text
func (a *App) getLogText() string {
	if len(a.logEntries) == 0 {
		return "Waiting for operations..."
	}
	result := ""
	for _, e := range a.logEntries {
		result += e + "\n"
	}
	return result
}

// showELI5Dialog displays the "Explain Like I'm 5" hysteresis guide
func (a *App) showELI5Dialog() {
	// Create content with key concepts from the ELI5 guide
	content := widget.NewLabel(
		"HYSTERESIS EXPLAINED LIKE YOU'RE 5\n\n" +
			"🔁 What is Hysteresis?\n" +
			"Like a rubber band that \"remembers\" being stretched.\n" +
			"The path going UP is different from the path coming DOWN.\n\n" +
			"💾 Why It Matters for Memory?\n" +
			"• Regular memory (DRAM): Like a whiteboard - erase & gone\n" +
			"• FeCIM memory: Like carving in clay - stays after power off!\n\n" +
			"📊 The P-E Loop:\n" +
			"• E = Electric Field (the \"push\" you apply)\n" +
			"• P = Polarization (material's response)\n" +
			"• When E = 0, P stays at ±Pr → MEMORY!\n\n" +
			"🎚️ Why 30 Levels?\n" +
			"• Binary: Like a light switch (ON/OFF) = 1 bit\n" +
			"• FeCIM: Like a dimmer with 30 positions = ~5 bits\n" +
			"• Same chip, 5× more storage!\n\n" +
			"📝 Write vs Read:\n" +
			"• WRITE: |E| > Ec → Data changes\n" +
			"• READ: |E| < Ec → Data unchanged, just sense\n" +
			"• HOLD: E = 0 → Data persists (no power!)\n\n" +
			"🎯 The Key Insight:\n" +
			"Hysteresis isn't a bug - it's the FEATURE that\n" +
			"enables memory! The loop REMEMBERS which\n" +
			"way you pushed it.\n\n" +
			"📚 Full Documentation:\n" +
			"See docs/hysteresis/hysteresis.ELI5.md for\n" +
			"detailed explanations with diagrams.")
	content.Wrapping = fyne.TextWrapWord

	// Create scrollable container
	scroll := container.NewScroll(content)
	scroll.SetMinSize(fyne.NewSize(600, 500))

	// Create dialog (declare as var first so button callback can reference it)
	var dialog *widget.PopUp
	closeBtn := widget.NewButton("Got it!", func() {
		if dialog != nil {
			dialog.Hide()
		}
	})

	dialog = widget.NewModalPopUp(
		container.NewVBox(
			container.NewPadded(scroll),
			closeBtn,
		),
		a.mainWindow.Canvas(),
	)

	dialog.Show()
}
