// pkg/gui/tabs/validation_tab.go
package tabs

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/config"
	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/validation"
)

// MakeValidationTab creates Tab 5: Validation
// Provides Yosys, DEF, and cross-file validation
func MakeValidationTab(cfg *config.ArrayConfig) fyne.CanvasObject {
	// Log output
	logOutput := widget.NewMultiLineEntry()
	logOutput.Wrapping = fyne.TextWrapWord

	// Result labels
	yosysResult := widget.NewLabel("⏸️ Not run")
	defResult := widget.NewLabel("⏸️ Not run")
	crossResult := widget.NewLabel("⏸️ Not run")
	overallResult := widget.NewLabel("")

	addLog := func(msg string) {
		logOutput.SetText(logOutput.Text + msg + "\n")
	}

	// Validation 1: Yosys
	runYosysBtn := widget.NewButton("Run Yosys Validation", func() {
		logOutput.SetText("")
		addLog("=== Yosys Verilog Validation ===")
		
		arrayPath := fmt.Sprintf("output/fecim_crossbar_%dx%d.v", cfg.Rows, cfg.Cols)
		cellPath := "cells/fecim_bitcell/fecim_bitcell.v"
		
		addLog(fmt.Sprintf("Validating: %s", arrayPath))
		addLog(fmt.Sprintf("Cell library: %s", cellPath))
		
		err := validation.ValidateVerilogWithCell(arrayPath, cellPath)
		if err != nil {
			yosysResult.SetText("❌ Failed")
			addLog(fmt.Sprintf("ERROR: %v", err))
		} else {
			yosysResult.SetText("✅ Passed")
			addLog("SUCCESS: Verilog syntax is valid")
		}
	})

	// Validation 2: DEF
	runDEFBtn := widget.NewButton("Run DEF Validation", func() {
		logOutput.SetText("")
		addLog("=== DEF Syntax Validation ===")
		
		defPath := fmt.Sprintf("output/fecim_crossbar_%dx%d.def", cfg.Rows, cfg.Cols)
		addLog(fmt.Sprintf("Validating: %s", defPath))
		
		err := validation.ValidateDEF(defPath)
		if err != nil {
			defResult.SetText("❌ Failed")
			addLog(fmt.Sprintf("ERROR: %v", err))
		} else {
			defResult.SetText("✅ Passed")
			stats, _ := validation.GetDEFStats(defPath)
			addLog(fmt.Sprintf("SUCCESS: DEF is valid"))
			addLog(fmt.Sprintf("Design: %v", stats["design_name"]))
			addLog(fmt.Sprintf("Components: %v", stats["component_count"]))
		}
	})

	// Validation 3: Cross-check
	runCrossBtn := widget.NewButton("Run Cross-Check", func() {
		logOutput.SetText("")
		addLog("=== LEF/LIB/V Cross-Check ===")
		
		lefPath := "cells/fecim_bitcell/fecim_bitcell.lef"
		libPath := "cells/fecim_bitcell/fecim_bitcell.lib"
		vPath := "cells/fecim_bitcell/fecim_bitcell.v"
		
		addLog(fmt.Sprintf("Checking: %s", lefPath))
		addLog(fmt.Sprintf("          %s", libPath))
		addLog(fmt.Sprintf("          %s", vPath))
		
		err := validation.CrossCheckFiles(lefPath, libPath, vPath)
		if err != nil {
			crossResult.SetText("❌ Failed")
			addLog(fmt.Sprintf("ERROR: %v", err))
		} else {
			crossResult.SetText("✅ Passed")
			addLog("SUCCESS: Pin names and cell names match across all files")
		}
	})

	// Run all button
	runAllBtn := widget.NewButton("▶ Run All Validations", func() {
		logOutput.SetText("")
		addLog("=== Running All Validations ===\n")
		
		// Run Yosys
		addLog("1. Yosys Verilog Validation...")
		arrayPath := fmt.Sprintf("output/fecim_crossbar_%dx%d.v", cfg.Rows, cfg.Cols)
		cellPath := "cells/fecim_bitcell/fecim_bitcell.v"
		err1 := validation.ValidateVerilogWithCell(arrayPath, cellPath)
		if err1 != nil {
			yosysResult.SetText("❌ Failed")
			addLog(fmt.Sprintf("   ERROR: %v\n", err1))
		} else {
			yosysResult.SetText("✅ Passed")
			addLog("   PASSED\n")
		}
		
		// Run DEF
		addLog("2. DEF Syntax Validation...")
		defPath := fmt.Sprintf("output/fecim_crossbar_%dx%d.def", cfg.Rows, cfg.Cols)
		err2 := validation.ValidateDEF(defPath)
		if err2 != nil {
			defResult.SetText("❌ Failed")
			addLog(fmt.Sprintf("   ERROR: %v\n", err2))
		} else {
			defResult.SetText("✅ Passed")
			addLog("   PASSED\n")
		}
		
		// Run cross-check
		addLog("3. LEF/LIB/V Cross-Check...")
		err3 := validation.CrossCheckFiles(
			"cells/fecim_bitcell/fecim_bitcell.lef",
			"cells/fecim_bitcell/fecim_bitcell.lib",
			"cells/fecim_bitcell/fecim_bitcell.v",
		)
		if err3 != nil {
			crossResult.SetText("❌ Failed")
			addLog(fmt.Sprintf("   ERROR: %v\n", err3))
		} else {
			crossResult.SetText("✅ Passed")
			addLog("   PASSED\n")
		}
		
		// Overall result
		if err1 == nil && err2 == nil && err3 == nil {
			overallResult.SetText("✅ ALL VALIDATIONS PASSED")
			addLog("=== ✅ ALL VALIDATIONS PASSED ===")
		} else {
			overallResult.SetText("❌ SOME VALIDATIONS FAILED")
			addLog("=== ❌ SOME VALIDATIONS FAILED ===")
		}
	})

	// Validation list
	validationList := container.NewVBox(
		widget.NewLabel("Validation Checks"),
		widget.NewSeparator(),
		container.NewHBox(widget.NewLabel("1. Yosys Verilog:"), yosysResult, runYosysBtn),
		container.NewHBox(widget.NewLabel("2. DEF Syntax:"), defResult, runDEFBtn),
		container.NewHBox(widget.NewLabel("3. LEF/LIB/V Cross-Check:"), crossResult, runCrossBtn),
		widget.NewSeparator(),
		overallResult,
		runAllBtn,
	)

	return container.NewBorder(
		validationList,
		nil, nil, nil,
		container.NewScroll(logOutput),
	)
}
