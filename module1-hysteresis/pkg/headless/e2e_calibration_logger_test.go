package headless

import (
	"encoding/csv"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestModule1HeadlessE2ECalibrationBundleMatrix(t *testing.T) {
	workspace := t.TempDir()
	t.Chdir(workspace)

	materials := ListMaterials()
	if len(materials) == 0 {
		t.Fatal("expected at least one material for calibration E2E")
	}
	material := materials[0]

	runs := []struct {
		name        string
		levels      int
		temperature float64
		force       bool
	}{
		{name: "initial-low-temperature", levels: 6, temperature: 250, force: true},
		{name: "overwrite-body-temperature", levels: 8, temperature: 310, force: true},
	}
	for _, run := range runs {
		t.Run(run.name, func(t *testing.T) {
			err := RunCLICalibration(CLICalibrationOptions{
				MaterialName: material,
				NumLevels:    run.levels,
				Temperature:  run.temperature,
				Force:        run.force,
				Verify:       false,
			})
			if err != nil {
				t.Fatalf("RunCLICalibration() error = %v", err)
			}

			data := readCalibrationE2EFile(t, material)
			if data.Version != calibrationVersion || data.MaterialName != material || data.NumLevels != run.levels {
				t.Fatalf("calibration metadata = version %d material %q levels %d, want %d/%q/%d", data.Version, data.MaterialName, data.NumLevels, calibrationVersion, material, run.levels)
			}
			tempKey := int(math.Round(run.temperature))
			cal := data.Calibrations[tempKey]
			if cal == nil {
				t.Fatalf("missing temperature calibration %dK in %#v", tempKey, data.Calibrations)
			}
			assertCalibrationArrayE2E(t, "up", cal.CalibrationUp, run.levels)
			assertCalibrationArrayE2E(t, "down", cal.CalibrationDown, run.levels)
			assertCalibrationArrayE2E(t, "up_low", cal.CalibUpLow, run.levels)
			assertCalibrationArrayE2E(t, "up_high", cal.CalibUpHigh, run.levels)
			assertCalibrationArrayE2E(t, "down_low", cal.CalibDownLow, run.levels)
			assertCalibrationArrayE2E(t, "down_high", cal.CalibDownHigh, run.levels)
			if len(cal.LastErrorUp) != run.levels || len(cal.LastErrorDown) != run.levels || len(cal.RelaxCompUp) != run.levels || len(cal.RelaxCompDown) != run.levels {
				t.Fatalf("state array lengths mismatch for levels=%d", run.levels)
			}
			assertStrictlyIncreasingFieldsE2E(t, "calibration_up", cal.CalibrationUp[1:])
			assertStrictlyIncreasingFieldsE2E(t, "calibration_down", cal.CalibrationDown[:len(cal.CalibrationDown)-1])
		})
	}

	before, err := os.ReadFile(calibrationFileForMaterial(material))
	if err != nil {
		t.Fatalf("read calibration before skip: %v", err)
	}
	if err := RunCLICalibration(CLICalibrationOptions{MaterialName: material, NumLevels: 12, Temperature: 390, Force: false}); err != nil {
		t.Fatalf("RunCLICalibration(skip existing) error = %v", err)
	}
	after, err := os.ReadFile(calibrationFileForMaterial(material))
	if err != nil {
		t.Fatalf("read calibration after skip: %v", err)
	}
	if string(before) != string(after) {
		t.Fatalf("Force=false changed existing calibration file")
	}

	missingErr := RunCLICalibration(CLICalibrationOptions{MaterialName: "definitely-not-a-module1-material", NumLevels: 4, Force: true})
	if missingErr == nil || !strings.Contains(missingErr.Error(), "material not found") {
		t.Fatalf("missing material error = %v, want material not found", missingErr)
	}
	matches, err := filepath.Glob(filepath.Join(calibrationDir, "definitely*.json"))
	if err != nil {
		t.Fatalf("glob missing material files: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("missing material produced calibration artifacts: %v", matches)
	}
}

func TestModule1HeadlessE2EDataLoggerWideCSVWorkflow(t *testing.T) {
	logsDir := t.TempDir()
	t.Setenv("FECIM_LOGS_DIR", logsDir)
	t.Setenv("FECIM_HYSTERESIS_LOG_INTERVAL_MS", "1000")

	logger, err := NewHysteresisDataLogger("HZO / logger.test")
	if err != nil {
		t.Fatalf("NewHysteresisDataLogger() error = %v", err)
	}
	if !strings.HasPrefix(logger.Path(), logsDir) || !strings.Contains(filepath.Base(logger.Path()), "hysteresis-hzo---logger-test-") {
		t.Fatalf("logger path %q not under sanitized logs dir %q", logger.Path(), logsDir)
	}
	if !logger.shouldRecord(0) || logger.shouldRecord(0.25) || !logger.shouldRecord(1.25) {
		t.Fatalf("logger simulation-time downsampling contract changed")
	}
	if !logger.shouldRecordAt(1.26, hysteresisDataLogISPPInterval) {
		t.Fatalf("ISPP interval should allow high-resolution critical sample")
	}

	snapshots := []HysteresisSnapshot{
		{Timestamp: time.RFC3339, SimTime: 0.00, Dt: 0.01, Waveform: "triangle", AutoMode: true, Material: "HZO", TemperatureK: 300, EcMVcm: 1.2, PsUcCm2: 30, PrUcCm2: 20, NumLevels: 30, LevelIndex: 0, Level: 0, StateBand: stateBand(0, 30), EField: -1.2e8, EFieldMVcm: -1.2, Polarization: -0.2, PolarizationU: -20, NormalizedP: -1},
		{Timestamp: time.RFC3339, SimTime: 0.25, Dt: 0.01, Waveform: "sine", AutoMode: true, Material: "HZO", TemperatureK: 300, EcMVcm: 1.2, PsUcCm2: 30, PrUcCm2: 20, NumLevels: 30, LevelIndex: 14, Level: 14, StateBand: stateBand(14, 30), EField: 0, EFieldMVcm: 0, Polarization: 0, PolarizationU: 0, NormalizedP: 0},
		{Timestamp: time.RFC3339, SimTime: 0.26, Dt: 0.01, Waveform: "ISPP", Material: "HZO", NumLevels: 30, LevelIndex: 15, Level: 15, StateBand: stateBand(15, 30), WrdPhase: 2, WrdPhaseName: wrdPhaseName(2), WrdTargetLevel: 15, WrdReadLevel: 14, WrdRetryCount: 1, WrdCycleEnergy: 3.5, WrdTotalWrites: 2, WrdSuccess: 1, WrdWriteE: 1.1e8, WrdStartLevel: 4, ControllerState: "VERIFY", ControllerTargetLevel: 15, ControllerCurrentField: 1.1e8, ControllerCurrentFieldMV: 1.1, ControllerPulseCount: 3, ControllerTotalPulses: 4, ControllerRetryCount: 1, ControllerOvershootCount: 2, ControllerOvershootTotal: 3, ControllerLastVerify: 14, ControllerLastError: -1, ControllerVMin: 0.3e8, ControllerVMax: 2e8, ControllerVMinEc: 0.3, ControllerVMaxEc: 2, ControllerInitialLevel: 4, ControllerFromSaturation: true, ControllerResetDirection: -1},
		{Timestamp: time.RFC3339, SimTime: 1.50, Dt: 0.01, Waveform: "square", Material: "HZO", NumLevels: 30, LevelIndex: 29, Level: 29, StateBand: stateBand(29, 30), EField: 1.2e8, EFieldMVcm: 1.2, Polarization: 0.2, PolarizationU: 20, NormalizedP: 1},
	}
	for _, snapshot := range snapshots {
		logger.Record(snapshot)
	}
	if err := logger.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	logger.Record(HysteresisSnapshot{Waveform: "after-close"})
	if err := logger.Close(); err != nil {
		t.Fatalf("second Close() error = %v", err)
	}

	rows := readLoggerCSVRowsE2E(t, logger.Path())
	if len(rows) != len(snapshots)+1 {
		t.Fatalf("CSV row count = %d, want header + %d", len(rows), len(snapshots))
	}
	header := rows[0]
	for _, column := range []string{"step", "waveform", "state_band", "wrd_target_level", "controller_state", "controller_reset_direction"} {
		if csvColumnE2E(header, column) < 0 {
			t.Fatalf("CSV header missing %q: %v", column, header)
		}
	}
	stepCol := csvColumnE2E(header, "step")
	waveCol := csvColumnE2E(header, "waveform")
	bandCol := csvColumnE2E(header, "state_band")
	targetCol := csvColumnE2E(header, "wrd_target_level")
	controllerCol := csvColumnE2E(header, "controller_state")
	resetDirCol := csvColumnE2E(header, "controller_reset_direction")
	wantWaves := []string{"triangle", "sine", "ISPP", "square"}
	for i, wantWave := range wantWaves {
		row := rows[i+1]
		if row[stepCol] != string(rune('1'+i)) || row[waveCol] != wantWave || row[bandCol] == "" {
			t.Fatalf("row %d key fields = step %q wave %q band %q", i+1, row[stepCol], row[waveCol], row[bandCol])
		}
	}
	ispp := rows[3]
	if ispp[targetCol] != "15" || ispp[controllerCol] != "VERIFY" || ispp[resetDirCol] != "-1" {
		t.Fatalf("ISPP/controller row fields = target %q state %q reset %q", ispp[targetCol], ispp[controllerCol], ispp[resetDirCol])
	}
}

func readCalibrationE2EFile(t *testing.T, material string) CalibrationData {
	t.Helper()
	f, err := os.Open(calibrationFileForMaterial(material))
	if err != nil {
		t.Fatalf("open calibration file: %v", err)
	}
	defer f.Close()
	var data CalibrationData
	if err := json.NewDecoder(f).Decode(&data); err != nil {
		t.Fatalf("decode calibration file: %v", err)
	}
	return data
}

func assertCalibrationArrayE2E(t *testing.T, name string, values []float64, wantLen int) {
	t.Helper()
	if len(values) != wantLen {
		t.Fatalf("%s length = %d, want %d", name, len(values), wantLen)
	}
	for i, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			t.Fatalf("%s[%d] is invalid: %g", name, i, value)
		}
	}
}

func assertStrictlyIncreasingFieldsE2E(t *testing.T, name string, values []float64) {
	t.Helper()
	for i := 1; i < len(values); i++ {
		if !(values[i] > values[i-1]) {
			t.Fatalf("%s is not strictly increasing at %d: %.6e <= %.6e", name, i, values[i], values[i-1])
		}
	}
}

func readLoggerCSVRowsE2E(t *testing.T, path string) [][]string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open CSV log: %v", err)
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatalf("read CSV log: %v", err)
	}
	return rows
}

func csvColumnE2E(header []string, name string) int {
	for i, column := range header {
		if column == name {
			return i
		}
	}
	return -1
}
