package tabs

import (
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/module6-eda/pkg/config"
	"fecim-lattice-tools/module6-eda/pkg/export"
)

var exportFormats = []string{"LEF", "Liberty", "Verilog", "DEF", "Config (JSON)", "SDC", "Design Summary", "SPICE"}

// MakeExportViewerTab creates a read-only export preview tab for LEF/Liberty/Verilog/DEF/SPICE.
func MakeExportViewerTab(cfg *config.ArrayConfig, window fyne.Window) fyne.CanvasObject {
	if cfg == nil {
		cfg = &config.ArrayConfig{Rows: 4, Cols: 4, Mode: "storage", Architecture: "passive", CellWidth: 0.46, CellHeight: 2.72}
	}

	formatSelect := widget.NewSelect(exportFormats, nil)
	formatSelect.SetSelected("LEF")

	status := widget.NewLabel("Ready")
	preview := widget.NewMultiLineEntry()
	preview.Wrapping = fyne.TextWrapOff
	preview.TextStyle.Monospace = true
	preview.Disable()

	refresh := func() {
		content, source := loadExportPreviewContent(formatSelect.Selected, cfg)
		preview.SetText(content)
		status.SetText("Source: " + source)
	}

	formatSelect.OnChanged = func(string) { refresh() }

	refreshBtn := widget.NewButton("Refresh", refresh)

	saveBtn := widget.NewButton("Save to File…", func() {
		ext := formatExtension(formatSelect.Selected)
		design := fmt.Sprintf("fecim_crossbar_%dx%d", cfg.Rows, cfg.Cols)
		defaultName := design + ext

		dlg := dialog.NewFileSave(func(w fyne.URIWriteCloser, err error) {
			if err != nil || w == nil {
				return
			}
			defer w.Close()
			if _, werr := w.Write([]byte(preview.Text)); werr != nil {
				dialog.ShowError(werr, window)
			}
		}, window)
		dlg.SetFileName(defaultName)
		dlg.Show()
	})

	header := container.NewHBox(
		widget.NewLabel("Format:"),
		formatSelect,
		refreshBtn,
		saveBtn,
		widget.NewSeparator(),
		status,
	)

	refresh()

	return container.NewBorder(header, nil, nil, nil, container.NewScroll(preview))
}

// formatExtension returns the canonical file extension for the given format name.
func formatExtension(format string) string {
	switch format {
	case "LEF":
		return ".lef"
	case "Liberty":
		return ".lib"
	case "Verilog":
		return ".v"
	case "DEF":
		return ".def"
	case "Config (JSON)":
		return ".json"
	case "SDC":
		return ".sdc"
	case "Design Summary":
		return ".txt"
	case "SPICE":
		return ".sp"
	default:
		return ".txt"
	}
}

func loadExportPreviewContent(format string, cfg *config.ArrayConfig) (content string, source string) {
	tech := cfg.Technology
	if tech == "" {
		tech = "sky130"
	}

	// Derive cell name and directory from architecture so LEF/Liberty reflect the correct cell type.
	cellName := "fecim_bitcell"
	cellDir := "fecim_bitcell"
	switch cfg.Architecture {
	case "1t1r":
		cellName = "fecim_1t1r_bitcell"
		cellDir = "fecim_1t1r_bitcell"
	case "2t1r":
		cellName = "fecim_2t1r_bitcell"
		cellDir = "fecim_2t1r_bitcell"
	}

	cellCfg := config.CellConfig{
		Name:         cellName,
		Width:        cfg.CellWidth,
		Height:       cfg.CellHeight,
		CellType:     cfg.Architecture,
		Technology:   tech,
		RiseTime:     10.0,
		FallTime:     10.0,
		InputCap:     0.015,
		LeakagePower: 0.0003,
	}

	design := fmt.Sprintf("fecim_crossbar_%dx%d", cfg.Rows, cfg.Cols)
	dataDir := "data"

	tryRead := func(path string) (string, bool) {
		b, err := os.ReadFile(path)
		if err != nil {
			return "", false
		}
		return string(b), true
	}

	switch format {
	case "LEF":
		// Try the architecture-specific cell file first.
		archLEF := filepath.Join("cells", cellDir, cellName+".lef")
		if s, ok := tryRead(archLEF); ok {
			return s, archLEF
		}
		// Fall back to the appropriate in-memory generator for the architecture.
		switch cfg.Architecture {
		case "1t1r":
			return export.Generate1T1RLEF(cellCfg), "generated (in-memory)"
		case "2t1r":
			return export.Generate2T1RLEF(cellCfg), "generated (in-memory)"
		default:
			return export.GenerateLEF(cellCfg), "generated (in-memory)"
		}

	case "Liberty":
		archLib := filepath.Join("cells", cellDir, cellName+".lib")
		if s, ok := tryRead(archLib); ok {
			return s, archLib
		}
		return export.GenerateLiberty(cellCfg), "generated (in-memory)"

	case "Verilog":
		p := filepath.Join(dataDir, design+".v")
		if s, ok := tryRead(p); ok {
			return s, p
		}
		return export.GenerateArrayVerilog(*cfg), "generated (in-memory)"

	case "DEF":
		paths := []string{
			filepath.Join(dataDir, design+".def"),
			filepath.Join("output", design+".def"),
		}
		for _, p := range paths {
			if s, ok := tryRead(p); ok {
				return s, p
			}
		}
		// Generate structural DEF in-memory (no compiled design required).
		return export.GenerateLatticeDEF(cfg.Rows, cfg.Cols), "generated (in-memory)"

	case "Config (JSON)":
		if s, ok := tryRead(filepath.Join(dataDir, "config.json")); ok {
			return s, filepath.Join(dataDir, "config.json")
		}
		return export.GenerateLibreLaneConfig(*cfg), "generated (in-memory)"

	case "SDC":
		if s, ok := tryRead(filepath.Join(dataDir, "constraints.sdc")); ok {
			return s, filepath.Join(dataDir, "constraints.sdc")
		}
		return export.GenerateSDC(*cfg), "generated (in-memory)"

	case "Design Summary":
		if s, ok := tryRead(filepath.Join(dataDir, "design_summary.txt")); ok {
			return s, filepath.Join(dataDir, "design_summary.txt")
		}
		return export.GenerateDesignSummary(*cfg), "generated (in-memory)"

	case "SPICE":
		paths := []string{
			filepath.Join(dataDir, design+".sp"),
			filepath.Join(dataDir, "fecim_array.sp"),
		}
		for _, p := range paths {
			if s, ok := tryRead(p); ok {
				return s, p
			}
		}
		// Generate a subcircuit preview: FeFET definition + architecture-specific bitcell.
		mat := export.DefaultHzoFeFETMaterial()
		preview := fmt.Sprintf(
			"* FeCIM Array SPICE Subcircuit Preview\n"+
				"* Array: %dx%d  Architecture: %s  Technology: %s\n"+
				"* NOTE: This shows cell subcircuit definitions only.\n"+
				"*       Full array netlist: use CLI with --spice flag.\n\n",
			cfg.Rows, cfg.Cols, cfg.Architecture, tech)
		preview += export.GenerateFeFETSubcircuit(mat)
		switch cfg.Architecture {
		case "1t1r":
			preview += export.Generate1T1RSubcircuit()
		case "2t1r":
			preview += export.Generate2T1RSubcircuit()
		}
		return preview, "generated (subcircuit preview)"

	default:
		return "", "unknown format"
	}
}
