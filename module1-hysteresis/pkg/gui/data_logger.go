package gui

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"fecim-lattice-tools/shared/logging"
)

const (
	hysteresisDataLogBuffer = 8192
	mvPerCm                 = 1e8
	ucPerCm2                = 1e2
	// Downsample CSV logging by simulation time to avoid huge files and UI stalls.
	// Default: 250ms => ~4 samples/sec at most.
	// Override with FECIM_HYSTERESIS_LOG_INTERVAL_MS (float, milliseconds).
	hysteresisDataLogMinSimInterval = 2.5e-1

	// During ISPP write operations, use a much finer recording interval so that the
	// CSV log captures the actual E-field ramp and polarization trajectory instead of
	// showing apparent "teleportation" jumps between coarsely-sampled records.
	hysteresisDataLogISPPInterval = 1e-2 // 10ms => ~100 samples/sec during ISPP
)

type HysteresisDataLogger struct {
	path   string
	file   *os.File
	rows   chan HysteresisSnapshot
	wg     sync.WaitGroup
	closed uint32
	step   uint64

	minSimInterval  float64
	lastSimTimeBits uint64

	dropped     uint64
	lastDropLog int64
}

type HysteresisSnapshot struct {
	Step          uint64
	Timestamp     string
	SimTime       float64
	Dt            float64
	Waveform      string
	AutoMode      bool
	Material      string
	TemperatureK  float64
	EcMVcm        float64
	PsUcCm2       float64
	PrUcCm2       float64
	NumLevels     int
	LevelIndex    int
	Level         int
	StateBand     string
	EField        float64
	EFieldMVcm    float64
	Polarization  float64
	PolarizationU float64
	NormalizedP   float64

	WrdPhase       int
	WrdPhaseName   string
	WrdPhaseTimer  float64
	WrdTargetLevel int
	WrdReadLevel   int
	WrdRetryCount  int
	WrdCycleEnergy float64
	WrdTotalWrites int
	WrdSuccess     int
	WrdWriteE      float64
	WrdPrepE       float64
	WrdSettleE     float64
	WrdStartLevel  int

	ControllerState          string
	ControllerPhaseTimer     float64
	ControllerTargetLevel    int
	ControllerCurrentField   float64
	ControllerCurrentFieldMV float64
	ControllerPulseCount     int
	ControllerTotalPulses    int
	ControllerRetryCount     int
	ControllerOvershootCount int
	ControllerOvershootTotal int
	ControllerLastVerify     int
	ControllerLastError      int
	ControllerVMin           float64
	ControllerVMax           float64
	ControllerVMinEc         float64
	ControllerVMaxEc         float64
	ControllerInitialLevel   int
	ControllerFromSaturation bool
	ControllerResetDirection int
}

func NewHysteresisDataLogger(materialName string) (*HysteresisDataLogger, error) {
	if materialName == "" {
		materialName = "unknown"
	}
	logsDir := logging.LogsDir()
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, fmt.Errorf("create data log dir: %w", err)
	}

	// Include microseconds to avoid filename collisions when multiple headless
	// runs start within the same second (common in fast test suites).
	timestamp := time.Now().Format("2006-01-02_15-04-05.000000")
	safeMaterial := sanitizeMaterialName(materialName)
	filename := fmt.Sprintf("hysteresis-%s-%s.csv", safeMaterial, timestamp)
	path := filepath.Join(logsDir, filename)

	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("create data log file: %w", err)
	}

	minInterval := hysteresisDataLogMinSimInterval
	if v := strings.TrimSpace(os.Getenv("FECIM_HYSTERESIS_LOG_INTERVAL_MS")); v != "" {
		if ms, err := strconv.ParseFloat(v, 64); err == nil {
			if ms <= 0 {
				minInterval = 0
			} else {
				minInterval = ms / 1000.0
			}
		}
	}

	logger := &HysteresisDataLogger{
		path: path,
		file: file,
		rows: make(chan HysteresisSnapshot, hysteresisDataLogBuffer),
		// Throttle CSV logging to keep runtime smooth and files manageable.
		minSimInterval:  minInterval,
		lastSimTimeBits: math.Float64bits(-1),
	}
	logger.wg.Add(1)
	go logger.run()
	return logger, nil
}

func (l *HysteresisDataLogger) Path() string {
	if l == nil {
		return ""
	}
	return l.path
}

func (l *HysteresisDataLogger) shouldRecord(simTime float64) bool {
	return l.shouldRecordAt(simTime, l.minSimInterval)
}

// shouldRecordAt is like shouldRecord but accepts a custom minimum interval.
// Use a shorter interval for high-resolution phases (e.g., ISPP writes).
func (l *HysteresisDataLogger) shouldRecordAt(simTime float64, minInterval float64) bool {
	if l == nil || atomic.LoadUint32(&l.closed) == 1 {
		return false
	}
	if minInterval <= 0 {
		return true
	}
	lastBits := atomic.LoadUint64(&l.lastSimTimeBits)
	last := math.Float64frombits(lastBits)
	if last >= 0 && simTime >= last && (simTime-last) < minInterval {
		return false
	}
	atomic.StoreUint64(&l.lastSimTimeBits, math.Float64bits(simTime))
	return true
}

func (l *HysteresisDataLogger) Record(snapshot HysteresisSnapshot) {
	if l == nil || atomic.LoadUint32(&l.closed) == 1 {
		return
	}
	snapshot.Step = atomic.AddUint64(&l.step, 1)

	defer func() {
		_ = recover()
	}()
	select {
	case l.rows <- snapshot:
		return
	default:
		atomic.AddUint64(&l.dropped, 1)
		if log == nil {
			return
		}
		now := time.Now().UnixNano()
		last := atomic.LoadInt64(&l.lastDropLog)
		if last != 0 && now-last < int64(time.Second) {
			return
		}
		if atomic.CompareAndSwapInt64(&l.lastDropLog, last, now) {
			dropped := atomic.LoadUint64(&l.dropped)
			log.Printf("Hysteresis data log backlog: dropped %d samples", dropped)
		}
	}
}

func (l *HysteresisDataLogger) Close() error {
	if l == nil {
		return nil
	}
	if !atomic.CompareAndSwapUint32(&l.closed, 0, 1) {
		return nil
	}
	close(l.rows)
	l.wg.Wait()
	if err := l.file.Close(); err != nil {
		return fmt.Errorf("close data log file: %w", err)
	}
	return nil
}

func (l *HysteresisDataLogger) run() {
	defer l.wg.Done()

	writer := csv.NewWriter(l.file)
	if err := writer.Write(hysteresisDataHeader()); err != nil {
		if log != nil {
			log.Printf("Hysteresis data log header error: %v", err)
		}
		writer.Flush()
		return
	}
	writer.Flush()

	rowCount := 0
	for snapshot := range l.rows {
		if err := writer.Write(snapshot.toCSVRow()); err != nil {
			if log != nil {
				log.Printf("Hysteresis data log write error: %v", err)
			}
			continue
		}
		rowCount++
		if rowCount%256 == 0 {
			writer.Flush()
			if err := writer.Error(); err != nil && log != nil {
				log.Printf("Hysteresis data log flush error: %v", err)
			}
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil && log != nil {
		log.Printf("Hysteresis data log final flush error: %v", err)
	}
}

func hysteresisDataHeader() []string {
	return []string{
		"step",
		"timestamp",
		"sim_time_s",
		"dt_s",
		"waveform",
		"auto_mode",
		"material",
		"temperature_k",
		"ec_mv_cm",
		"ps_uc_cm2",
		"pr_uc_cm2",
		"num_levels",
		"level_index",
		"level",
		"state_band",
		"e_field_v_m",
		"e_field_mv_cm",
		"polarization_c_m2",
		"polarization_uc_cm2",
		"normalized_p",
		"wrd_phase",
		"wrd_phase_name",
		"wrd_phase_timer_s",
		"wrd_target_level",
		"wrd_read_level",
		"wrd_retry_count",
		"wrd_cycle_energy_fj",
		"wrd_total_writes",
		"wrd_success_writes",
		"wrd_write_e_v_m",
		"wrd_prep_e_v_m",
		"wrd_settle_e_v_m",
		"wrd_start_level",
		"controller_state",
		"controller_phase_timer_s",
		"controller_target_level",
		"controller_current_field_v_m",
		"controller_current_field_mv_cm",
		"controller_pulse_count",
		"controller_total_pulses",
		"controller_retry_count",
		"controller_overshoot_count",
		"controller_overshoot_total",
		"controller_last_verify_level",
		"controller_last_error",
		"controller_vmin_v_m",
		"controller_vmax_v_m",
		"controller_vmin_ec",
		"controller_vmax_ec",
		"controller_initial_level",
		"controller_from_saturation",
		"controller_reset_direction",
	}
}

func (s HysteresisSnapshot) toCSVRow() []string {
	return []string{
		strconv.FormatUint(s.Step, 10),
		s.Timestamp,
		formatFloat(s.SimTime),
		formatFloat(s.Dt),
		s.Waveform,
		strconv.FormatBool(s.AutoMode),
		s.Material,
		formatFloat(s.TemperatureK),
		formatFloat(s.EcMVcm),
		formatFloat(s.PsUcCm2),
		formatFloat(s.PrUcCm2),
		strconv.Itoa(s.NumLevels),
		strconv.Itoa(s.LevelIndex),
		strconv.Itoa(s.Level),
		s.StateBand,
		formatFloat(s.EField),
		formatFloat(s.EFieldMVcm),
		formatFloat(s.Polarization),
		formatFloat(s.PolarizationU),
		formatFloat(s.NormalizedP),
		strconv.Itoa(s.WrdPhase),
		s.WrdPhaseName,
		formatFloat(s.WrdPhaseTimer),
		strconv.Itoa(s.WrdTargetLevel),
		strconv.Itoa(s.WrdReadLevel),
		strconv.Itoa(s.WrdRetryCount),
		formatFloat(s.WrdCycleEnergy),
		strconv.Itoa(s.WrdTotalWrites),
		strconv.Itoa(s.WrdSuccess),
		formatFloat(s.WrdWriteE),
		formatFloat(s.WrdPrepE),
		formatFloat(s.WrdSettleE),
		strconv.Itoa(s.WrdStartLevel),
		s.ControllerState,
		formatFloat(s.ControllerPhaseTimer),
		strconv.Itoa(s.ControllerTargetLevel),
		formatFloat(s.ControllerCurrentField),
		formatFloat(s.ControllerCurrentFieldMV),
		strconv.Itoa(s.ControllerPulseCount),
		strconv.Itoa(s.ControllerTotalPulses),
		strconv.Itoa(s.ControllerRetryCount),
		strconv.Itoa(s.ControllerOvershootCount),
		strconv.Itoa(s.ControllerOvershootTotal),
		strconv.Itoa(s.ControllerLastVerify),
		strconv.Itoa(s.ControllerLastError),
		formatFloat(s.ControllerVMin),
		formatFloat(s.ControllerVMax),
		formatFloat(s.ControllerVMinEc),
		formatFloat(s.ControllerVMaxEc),
		strconv.Itoa(s.ControllerInitialLevel),
		strconv.FormatBool(s.ControllerFromSaturation),
		strconv.Itoa(s.ControllerResetDirection),
	}
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'g', -1, 64)
}

func sanitizeMaterialName(name string) string {
	if name == "" {
		return "unknown"
	}
	safe := strings.ToLower(name)
	safe = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == '-' || r == '_':
			return r
		case r == ' ' || r == '.' || r == '/':
			return '-'
		default:
			return -1
		}
	}, safe)
	if safe == "" {
		return "unknown"
	}
	return safe
}

func wrdPhaseName(phase int) string {
	switch phase {
	case 0:
		return "PREP"
	case 1:
		return "SETTLE"
	case 2:
		return "PROG_VERIFY"
	case 3:
		return "HOLD"
	case 4:
		return "READBACK"
	case 5:
		return "RESULT"
	case 6:
		return "RETRY"
	default:
		return "UNKNOWN"
	}
}

func stateBand(levelIndex int, numLevels int) string {
	if numLevels <= 0 {
		return "UNKNOWN"
	}
	lowThird := numLevels / 3
	highThird := numLevels * 2 / 3
	if levelIndex < lowThird {
		return "NEGATIVE"
	}
	if levelIndex >= highThird {
		return "POSITIVE"
	}
	return "INTERMEDIATE"
}

func (a *App) startDataLogger() {
	if a == nil || a.dataLogger != nil || !logging.IsFileLoggingEnabled() {
		return
	}
	materialName := "unknown"
	if a.material != nil {
		materialName = a.material.Name
	}
	logger, err := NewHysteresisDataLogger(materialName)
	if err != nil {
		if log != nil {
			log.Printf("Hysteresis data logger init failed: %v", err)
		}
		return
	}
	a.dataLogger = logger
	if log != nil {
		log.Info("Hysteresis data logging enabled: %s", logger.Path())
	}
}

func (a *App) stopDataLogger() {
	if a == nil || a.dataLogger == nil {
		return
	}
	if err := a.dataLogger.Close(); err != nil && log != nil {
		log.Printf("Hysteresis data logger close failed: %v", err)
	}
	a.dataLogger = nil
}

func (a *App) recordDataSnapshot(dt float64) {
	if a == nil || a.dataLogger == nil {
		return
	}
	mat := a.material
	if mat == nil {
		return
	}

	// During ISPP writes, use a finer recording interval so that the CSV log
	// captures the E-field ramp and P trajectory (prevents apparent teleportation).
	// Also force-record on controller state transitions (APPLY→WAIT→VERIFY etc.)
	// to capture exact transition points regardless of the throttle interval.
	if a.waveform == WaveformWriteReadDemo {
		forceRecord := false
		if a.writeController != nil && a.writeController.State != a.wrdLastLogState {
			forceRecord = true
			a.wrdLastLogState = a.writeController.State
		}
		if !forceRecord && !a.dataLogger.shouldRecordAt(a.simTime, hysteresisDataLogISPPInterval) {
			return
		}
	} else if !a.dataLogger.shouldRecord(a.simTime) {
		return
	}

	temp := a.currentTemperature()

	snapshot := HysteresisSnapshot{
		Timestamp:      time.Now().UTC().Format(time.RFC3339Nano),
		SimTime:        a.simTime,
		Dt:             dt,
		Waveform:       a.waveform.String(),
		AutoMode:       a.autoMode,
		Material:       mat.Name,
		TemperatureK:   temp,
		EcMVcm:         mat.Ec / mvPerCm,
		PsUcCm2:        mat.Ps * ucPerCm2,
		PrUcCm2:        mat.Pr * ucPerCm2,
		NumLevels:      a.numLevels,
		LevelIndex:     a.discreteLevel,
		Level:          a.discreteLevel + 1,
		StateBand:      stateBand(a.discreteLevel, a.numLevels),
		EField:         a.electricField,
		EFieldMVcm:     a.electricField / mvPerCm,
		Polarization:   a.polarization,
		PolarizationU:  a.polarization * ucPerCm2,
		NormalizedP:    a.normalizedP,
		WrdPhase:       a.wrdPhase,
		WrdPhaseName:   wrdPhaseName(a.wrdPhase),
		WrdPhaseTimer:  a.wrdPhaseTimer,
		WrdTargetLevel: a.wrdTargetLevel,
		WrdReadLevel:   a.wrdReadLevel,
		WrdRetryCount:  a.wrdRetryCount,
		WrdCycleEnergy: a.wrdCycleEnergy,
		WrdTotalWrites: a.wrdTotalWrites,
		WrdSuccess:     a.wrdSuccessWrites,
		WrdWriteE:      a.wrdWriteE,
		WrdPrepE:       a.wrdPrepE,
		WrdSettleE:     a.wrdSettleE,
		WrdStartLevel:  a.wrdStartLevel,
	}

	if a.writeController != nil {
		ctrl := a.writeController
		snapshot.ControllerState = ctrl.State.String()
		snapshot.ControllerPhaseTimer = ctrl.PhaseTimer
		snapshot.ControllerTargetLevel = ctrl.TargetLevel
		snapshot.ControllerCurrentField = ctrl.CurrentField
		snapshot.ControllerCurrentFieldMV = ctrl.CurrentField / mvPerCm
		snapshot.ControllerPulseCount = ctrl.PulseCount
		snapshot.ControllerTotalPulses = ctrl.TotalPulses
		snapshot.ControllerRetryCount = ctrl.RetryCount
		snapshot.ControllerOvershootCount = ctrl.OvershootCount
		snapshot.ControllerOvershootTotal = ctrl.OvershootTotal
		snapshot.ControllerLastVerify = ctrl.LastVerifyLevel
		snapshot.ControllerLastError = ctrl.LastError
		snapshot.ControllerVMin = ctrl.VMin
		snapshot.ControllerVMax = ctrl.VMax
		if ctrl.EcField != 0 {
			snapshot.ControllerVMinEc = ctrl.VMin / ctrl.EcField
			snapshot.ControllerVMaxEc = ctrl.VMax / ctrl.EcField
		}
		snapshot.ControllerInitialLevel = ctrl.InitialLevel
		snapshot.ControllerFromSaturation = ctrl.FromSaturation
		snapshot.ControllerResetDirection = ctrl.ResetDirection()
	}

	a.dataLogger.Record(snapshot)
}

// WriteReadDebugLog stores debug data for write/read operations
type WriteReadDebugLog struct {
	Timestamp string           `json:"timestamp"`
	Material  string           `json:"material"`
	Ec        float64          `json:"ec_v_per_m"`
	EcMVcm    float64          `json:"ec_mv_per_cm"`
	Ps        float64          `json:"ps_c_per_m2"`
	Cycles    []WriteReadCycle `json:"cycles"`
}

// WriteReadCycle stores one complete write/read cycle
type WriteReadCycle struct {
	CycleNum    int              `json:"cycle_num"`
	TargetLevel int              `json:"target_level"`
	StartLevel  int              `json:"start_level"`
	ReadLevel   int              `json:"read_level"`
	Success     bool             `json:"success"`
	Phases      []WriteReadPhase `json:"phases"`
}

// WriteReadPhase stores data for one phase of the write/read cycle
type WriteReadPhase struct {
	Phase       string           `json:"phase"`
	Duration    float64          `json:"duration_s"`
	EFieldStart float64          `json:"e_field_start_mv_cm"`
	EFieldEnd   float64          `json:"e_field_end_mv_cm"`
	EFieldPeak  float64          `json:"e_field_peak_mv_cm"`
	PStart      float64          `json:"p_start_uc_cm2"`
	PEnd        float64          `json:"p_end_uc_cm2"`
	LevelStart  int              `json:"level_start"`
	LevelEnd    int              `json:"level_end"`
	Samples     []PhaseDataPoint `json:"samples,omitempty"`
}

// PhaseDataPoint stores a single data point during a phase
type PhaseDataPoint struct {
	Time  float64 `json:"t"`
	E     float64 `json:"e"`
	P     float64 `json:"p"`
	Level int     `json:"level"`
}

// saveDebugLog saves the debug log to a JSON file
// THREAD-SAFE: No UI updates - only file I/O operations. Safe to call from goroutines without fyne.Do().
func (a *App) saveDebugLog() {
	if a.wrdDebugLog == nil || len(a.wrdDebugLog.Cycles) == 0 {
		return
	}

	// Create logs directory if it doesn't exist
	logsDir := "logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		log.Printf("Error creating logs dir: %v", err)
		return
	}

	// Generate filename with timestamp
	filename := filepath.Join(logsDir, fmt.Sprintf("hysteresis-%s.json",
		time.Now().Format("2006-01-02T15-04-05")))

	// Marshal to JSON
	data, err := json.MarshalIndent(a.wrdDebugLog, "", "  ")
	if err != nil {
		log.Printf("Error marshaling debug log: %v", err)
		return
	}

	// Write to file
	if err := os.WriteFile(filename, data, 0644); err != nil {
		log.Printf("Error writing debug log: %v", err)
		return
	}

	log.Info("Debug log saved: %s", filename)
}

// initDebugLog initializes the debug log
func (a *App) initDebugLog() {
	// NOTE: Caller must hold a.mu.Lock() before calling this function
	// Defensive: ensure material exists
	materialName := "Unknown"
	materialEc := 0.0
	materialPs := 0.0
	if a.material != nil {
		materialName = a.material.Name
		materialEc = a.material.Ec
		materialPs = a.material.Ps
	}

	a.wrdDebugLog = &WriteReadDebugLog{
		Timestamp: time.Now().Format(time.RFC3339),
		Material:  materialName,
		Ec:        materialEc,
		EcMVcm:    materialEc / 1e8,
		Ps:        materialPs,
		Cycles:    make([]WriteReadCycle, 0),
	}
	// Clear previous log entries and add impressive startup banner
	// Defensive: initialize logEntries if nil
	if a.logEntries == nil {
		a.logEntries = make([]string, 0, 12)
	} else {
		a.logEntries = a.logEntries[:0]
	}
	a.logEntries = append(a.logEntries, "══════════════════════")
	a.logEntries = append(a.logEntries, "  FeCIM WRITE/READ    ")
	a.logEntries = append(a.logEntries, "══════════════════════")
	a.logEntries = append(a.logEntries, fmt.Sprintf("Material: %s", materialName))
	a.logEntries = append(a.logEntries, fmt.Sprintf("Ec: %.1f MV/cm", materialEc/1e8))
	a.logEntries = append(a.logEntries, fmt.Sprintf("%d LEVELS = %.2f bits/cell", a.numLevels, math.Log2(float64(a.numLevels))))
	a.logEntries = append(a.logEntries, "──────────────────────")
}
