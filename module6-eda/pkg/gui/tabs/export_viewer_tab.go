package tabs

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/module6-eda/pkg/config"
	"fecim-lattice-tools/module6-eda/pkg/export"
)

var exportFormats = []string{"LEF", "Liberty", "Verilog", "SPICE"}

// MakeExportViewerTab creates a read-only export preview tab for LEF/Liberty/Verilog/SPICE.
func MakeExportViewerTab(cfg *config.ArrayConfig, _ fyne.Window) fyne.CanvasObject {
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

	selectedArtifact := ""
	refresh := func() {
		content, source := loadExportPreviewContent(formatSelect.Selected, cfg)
		preview.SetText(content)
		if selectedArtifact != "" {
			status.SetText("Artifact: " + selectedArtifact + " | Source: " + source)
			return
		}
		status.SetText("Source: " + source)
	}

	artifactModel := NewArtifactTreeModel(exportPreviewArtifactPaths(cfg))
	formatChangeFromTree := false
	formatSelect.OnChanged = func(string) {
		if !formatChangeFromTree {
			selectedArtifact = ""
		}
		refresh()
	}

	artifactTree := newExportArtifactTree(artifactModel, func(uid string) {
		if artifactModel.IsBranch(uid) {
			return
		}
		format := exportFormatForArtifactID(uid)
		if format == "" {
			return
		}
		selectedArtifact = uid
		if formatSelect.Selected != format {
			formatChangeFromTree = true
			formatSelect.SetSelected(format)
			formatChangeFromTree = false
			return
		}
		refresh()
	})
	artifactTree.OpenBranch("cells")
	artifactTree.OpenBranch("data")

	refreshBtn := widget.NewButton("Refresh", refresh)

	header := container.NewHBox(
		widget.NewLabel("Format:"),
		formatSelect,
		refreshBtn,
		widget.NewSeparator(),
		status,
	)

	artifactBrowser := container.NewBorder(
		widget.NewLabelWithStyle("Artifacts", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		nil,
		nil,
		nil,
		container.NewScroll(artifactTree),
	)
	content := container.NewHSplit(artifactBrowser, container.NewScroll(preview))
	content.Offset = 0.3

	refresh()

	return container.NewBorder(header, nil, nil, nil, content)
}

func exportPreviewArtifactPaths(cfg *config.ArrayConfig) []string {
	design := fmt.Sprintf("fecim_crossbar_%dx%d", cfg.Rows, cfg.Cols)
	return []string{
		path.Join("cells", "fecim_bitcell", "fecim_bitcell.lef"),
		path.Join("cells", "fecim_bitcell", "fecim_bitcell.lib"),
		path.Join("cells", "fecim_1t1r_bitcell", "fecim_1t1r_bitcell.lef"),
		path.Join("cells", "fecim_1t1r_bitcell", "fecim_1t1r_bitcell.lib"),
		path.Join("cells", "fecim_2t1r_bitcell", "fecim_2t1r_bitcell.lef"),
		path.Join("cells", "fecim_2t1r_bitcell", "fecim_2t1r_bitcell.lib"),
		path.Join("data", design+".v"),
		path.Join("data", design+".sp"),
		path.Join("data", "fecim_array.sp"),
	}
}

func newExportArtifactTree(model *ArtifactTreeModel, onSelected func(string)) *widget.Tree {
	tree := widget.NewTree(
		model.ChildUIDs,
		model.IsBranch,
		func(bool) fyne.CanvasObject {
			label := widget.NewLabel("artifact")
			label.Truncation = fyne.TextTruncateEllipsis
			return label
		},
		func(uid widget.TreeNodeID, branch bool, node fyne.CanvasObject) {
			label := node.(*widget.Label)
			text := model.Label(uid)
			if branch {
				text = "▸ " + text
			}
			label.SetText(text)
		},
	)
	tree.HideSeparators = true
	tree.OnSelected = onSelected
	return tree
}

func exportFormatForArtifactID(uid string) string {
	switch path.Ext(uid) {
	case ".lef":
		return "LEF"
	case ".lib":
		return "Liberty"
	case ".v":
		return "Verilog"
	case ".sp", ".spice":
		return "SPICE"
	default:
		return ""
	}
}

func loadExportPreviewContent(format string, cfg *config.ArrayConfig) (content string, source string) {
	cellCfg := config.CellConfig{
		Name:         "fecim_bitcell",
		Width:        cfg.CellWidth,
		Height:       cfg.CellHeight,
		CellType:     cfg.Architecture,
		Technology:   "sky130",
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
		paths := []string{
			filepath.Join("cells", "fecim_bitcell", "fecim_bitcell.lef"),
			filepath.Join("cells", "fecim_1t1r_bitcell", "fecim_1t1r_bitcell.lef"),
			filepath.Join("cells", "fecim_2t1r_bitcell", "fecim_2t1r_bitcell.lef"),
		}
		for _, p := range paths {
			if s, ok := tryRead(p); ok {
				return s, p
			}
		}
		return export.GenerateLEF(cellCfg), "generated (in-memory)"
	case "Liberty":
		paths := []string{
			filepath.Join("cells", "fecim_bitcell", "fecim_bitcell.lib"),
			filepath.Join("cells", "fecim_1t1r_bitcell", "fecim_1t1r_bitcell.lib"),
			filepath.Join("cells", "fecim_2t1r_bitcell", "fecim_2t1r_bitcell.lib"),
		}
		for _, p := range paths {
			if s, ok := tryRead(p); ok {
				return s, p
			}
		}
		return export.GenerateLiberty(cellCfg), "generated (in-memory)"
	case "Verilog":
		p := filepath.Join(dataDir, design+".v")
		if s, ok := tryRead(p); ok {
			return s, p
		}
		return export.GenerateArrayVerilog(*cfg), "generated (in-memory)"
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
		return "* SPICE export not found on disk\n* Expected: data/" + design + ".sp\n* CLI export supports SPICE output.", "not found"
	default:
		return "", "unknown format"
	}
}
