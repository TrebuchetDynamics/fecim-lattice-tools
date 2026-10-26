//go:build cgo

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/workbench/experiment"
	"fecim-lattice-tools/workbench/study"
)

func parseSweepValues(raw string) ([]int, error) {
	parts := strings.Split(raw, ",")
	if len(parts) == 0 || len(parts) > 16 {
		return nil, errors.New("enter 1 to 16 comma-separated integers")
	}
	values := make([]int, len(parts))
	for i, part := range parts {
		value, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return nil, fmt.Errorf("invalid integer %q", strings.TrimSpace(part))
		}
		values[i] = value
	}
	return values, nil
}

func joinInts(values []int) string {
	parts := make([]string, len(values))
	for i, value := range values {
		parts[i] = strconv.Itoa(value)
	}
	return strings.Join(parts, ",")
}

func CreateResearchWorkbenchContent(window fyne.Window) fyne.CanvasObject {
	return createResearchWorkbenchContent(window, study.Run)
}

func createResearchWorkbenchContent(window fyne.Window, runStudy func(context.Context, string, study.RunOptions) (study.RunOutcome, error)) fyne.CanvasObject {
	projectPath := widget.NewEntry()
	projectPath.SetPlaceHolder("Select an existing project folder or an empty folder to initialize")
	status := widget.NewLabel("Local projects only · simulation estimates, not measured hardware")
	output := widget.NewMultiLineEntry()
	output.SetMinRowsVisible(12)

	var busy bool
	var cancelTask context.CancelFunc
	var buttons []*widget.Button
	var exportButton *widget.Button
	runSelect := widget.NewSelect(nil, func(selected string) {
		if selected != "" && !busy {
			exportButton.Enable()
		} else {
			exportButton.Disable()
		}
	})
	runSelect.PlaceHolder = "Run ID (successful runs only)"
	refreshRuns := func(root string) {
		entries, err := os.ReadDir(filepath.Join(root, "runs"))
		if err != nil {
			runSelect.SetOptions(nil)
			runSelect.SetSelected("")
			return
		}
		ids := make([]string, 0, len(entries))
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			if _, err := experiment.LoadRun(root, entry.Name()); err == nil {
				ids = append(ids, entry.Name())
			}
		}
		previous := runSelect.Selected
		runSelect.SetOptions(ids)
		if previous != "" {
			runSelect.SetSelected(previous)
		} else if len(ids) > 0 {
			runSelect.SetSelected(ids[len(ids)-1])
		}
		if runSelect.Selected == "" {
			exportButton.Disable()
		} else {
			exportButton.Enable()
		}
	}
	runTaskWithOutcome := func(label string, cancellable bool, operation func(context.Context, string) (status, details string)) {
		root := strings.TrimSpace(projectPath.Text)
		if root == "" {
			status.SetText("Choose a project folder first")
			return
		}
		if busy {
			return
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancelTask = cancel
		busy = true
		for _, button := range buttons {
			button.Disable()
		}
		runSelect.Disable()
		if cancellable {
			// Cancel remains available while long-running simulations execute.
			buttons[len(buttons)-1].Enable()
		}
		status.SetText(label + "…")
		go func() {
			terminalStatus, result := operation(ctx, root)
			cancel()
			fyne.Do(func() {
				busy = false
				cancelTask = nil
				for _, button := range buttons {
					button.Enable()
				}
				runSelect.Enable()
				refreshRuns(root)
				buttons[len(buttons)-1].Disable()
				status.SetText(terminalStatus)
				output.SetText(result)
			})
		}()
	}

	// Non-simulation actions retain their existing text/error presentation.
	runTask := func(label string, cancellable bool, operation func(context.Context, string) (string, error)) {
		runTaskWithOutcome(label, cancellable, func(ctx context.Context, root string) (string, string) {
			result, err := operation(ctx, root)
			if err != nil {
				if workbenchCancellationOnly(err) {
					return label + " cancelled · committed runs are preserved", err.Error()
				}
				return label + " failed", err.Error()
			}
			return label + " complete", result
		})
	}

	choose := widget.NewButton("Choose folder…", func() {
		dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				status.SetText("Folder selection failed: " + err.Error())
				return
			}
			if uri != nil {
				projectPath.SetText(uri.Path())
				refreshRuns(uri.Path())
			}
		}, window).Show()
	})
	initialize := widget.NewButton("Create starter project", func() {
		runTask("Initialize", false, func(_ context.Context, root string) (string, error) {
			var stdout, stderr strings.Builder
			err := runProjectSubcommand([]string{"init", root}, &stdout, &stderr)
			return stdout.String() + stderr.String(), err
		})
	})
	configureStarter := widget.NewButton("Configure starter sweep…", func() {
		defaults := study.DefaultStarterSweepOptions()
		levels := widget.NewEntry()
		levels.SetText(joinInts(defaults.ConductanceLevels))
		rows := widget.NewEntry()
		rows.SetText(joinInts(defaults.Rows))
		adcBits := widget.NewEntry()
		adcBits.SetText(joinInts(defaults.ADCBits))
		form := dialog.NewForm("Configure HZO starter sweep", "Create project", "Cancel", []*widget.FormItem{
			widget.NewFormItem("Conductance levels (2–64)", levels),
			widget.NewFormItem("Array rows (1–4096)", rows),
			widget.NewFormItem("ADC bits (1–16)", adcBits),
		}, func(confirm bool) {
			if !confirm {
				return
			}
			conductance, err := parseSweepValues(levels.Text)
			if err != nil {
				status.SetText("Conductance levels: " + err.Error())
				return
			}
			rowValues, err := parseSweepValues(rows.Text)
			if err != nil {
				status.SetText("Array rows: " + err.Error())
				return
			}
			adcValues, err := parseSweepValues(adcBits.Text)
			if err != nil {
				status.SetText("ADC bits: " + err.Error())
				return
			}
			options := study.StarterSweepOptions{ConductanceLevels: conductance, Rows: rowValues, ADCBits: adcValues}
			runTask("Initialize", false, func(_ context.Context, root string) (string, error) {
				if err := study.CreateStarterProject(root, options); err != nil {
					return "", err
				}
				return "initialized configured HZO sweep: " + filepath.Clean(root) + "\n", nil
			})
		}, window)
		form.Show()
	})
	validate := widget.NewButton("Validate project", func() {
		runTask("Validate", false, func(_ context.Context, root string) (string, error) {
			bundle, pointCount, err := study.ValidateProject(root, "")
			if err != nil {
				return "", err
			}
			var details strings.Builder
			fmt.Fprintf(&details, "Project: %s\nHypothesis: %s\nModel version: %s\nDesign points: %d\nObjectives: %d\n", bundle.Project.ID, bundle.Project.Hypothesis, bundle.Project.ModelVersion, pointCount, len(bundle.Project.Objectives))
			for _, input := range bundle.Inputs {
				fmt.Fprintf(&details, "Input: %s  sha256:%s  %s\n", input.Path, input.SHA256, input.Evidence)
			}
			for _, citation := range bundle.Project.Citations {
				fmt.Fprintf(&details, "Citation: %s\n", citation)
			}
			return details.String(), nil
		})
	})
	simulate := widget.NewButton("Simulate sweep", func() {
		runTaskWithOutcome("Simulation", true, func(ctx context.Context, root string) (string, string) {
			outcome, err := runStudy(ctx, root, study.RunOptions{Workers: 1, RepositoryRevision: buildRevision()})
			return formatWorkbenchOutcome(outcome, err)
		})
	})
	analyze := widget.NewButton("Analyze · generate reports", func() {
		runTask("Analysis", false, func(_ context.Context, root string) (string, error) {
			analysis, err := study.GenerateReport(root, "")
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("generated: %s/reports · %d runs\n", root, len(analysis.Runs)), nil
		})
	})
	verifyRuns := widget.NewButton("Validate · run integrity/provenance", func() {
		runTask("Run validation", false, func(_ context.Context, root string) (string, error) {
			if _, _, err := study.ValidateProject(root, ""); err != nil {
				return "", err
			}
			runs, err := experiment.LoadRuns(root)
			if err != nil {
				return "", err
			}
			if len(runs) == 0 {
				return "", errors.New("project has no committed runs to validate")
			}
			var details strings.Builder
			fmt.Fprintf(&details, "Verified %d persisted runs against stored identities and artifact hashes.\n", len(runs))
			for _, run := range runs {
				fmt.Fprintf(&details, "%s  status=%s  record_sha256=%s\n", run.Manifest.RunID, run.Result.Status, run.Manifest.RecordSHA256)
			}
			return details.String(), nil
		})
	})
	cancel := widget.NewButton("Cancel simulation", func() {
		if cancelTask != nil {
			cancelTask()
			status.SetText("Cancel requested · completed runs are preserved and can be resumed")
			buttons[len(buttons)-1].Disable()
		}
	})
	exportButton = widget.NewButton("Export selected run · SPICE/Verilog/DEF/LEF/Liberty", func() {
		runID := runSelect.Selected
		if runID == "" {
			return
		}
		runTask("EDA export", false, func(_ context.Context, root string) (string, error) {
			outputDir, err := study.ExportRun(root, runID)
			if err != nil {
				return "", err
			}
			return "exported: " + outputDir + "\n", nil
		})
	})
	exportButton.Disable()
	cancel.Disable()
	buttons = []*widget.Button{initialize, configureStarter, validate, simulate, analyze, verifyRuns, exportButton, cancel}

	controls := container.NewVBox(choose, projectPath, container.NewGridWithColumns(2, initialize, configureStarter), validate, container.NewGridWithColumns(2, simulate, analyze), container.NewGridWithColumns(2, verifyRuns, exportButton), runSelect, cancel, status)
	return container.NewBorder(widget.NewLabel("Research and design workbench"), nil, nil, nil, container.NewVBox(controls, output))
}
