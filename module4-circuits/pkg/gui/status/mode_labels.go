//go:build legacy_fyne

package status

// Operation mode values mirror gui.OpMode without importing the stateful gui package.
const (
	OperationRead = iota
	OperationWrite
	OperationCompute
)

// DAC mode values mirror gui.DACMode without importing the stateful gui package.
const (
	DACManualMode = iota
	DACReadPresetMode
	DACWritePresetMode
	DACInputVectorMode
	DACRandomMode
)

// DAC range values mirror gui.DACRangeMode without importing the stateful gui package.
const (
	DACRangeReadMode = iota
	DACRangeWriteMode
)

// OpModeLabel returns the canonical short label for an operation mode.
func OpModeLabel(mode int) string {
	switch mode {
	case OperationRead:
		return "READ"
	case OperationWrite:
		return "WRITE"
	case OperationCompute:
		return "COMPUTE"
	default:
		return "IDLE"
	}
}

// DACModeLabel returns the canonical short label for a DAC mode.
func DACModeLabel(mode int) string {
	switch mode {
	case DACManualMode:
		return "MANUAL"
	case DACReadPresetMode:
		return "READ_PRESET"
	case DACWritePresetMode:
		return "WRITE_PRESET"
	case DACInputVectorMode:
		return "INPUT_VECTOR"
	case DACRandomMode:
		return "RANDOM"
	default:
		return "UNKNOWN"
	}
}

// DACRangeLabel returns the canonical short label for a DAC range mode.
func DACRangeLabel(mode int) string {
	switch mode {
	case DACRangeReadMode:
		return "READ"
	case DACRangeWriteMode:
		return "WRITE"
	default:
		return "UNKNOWN"
	}
}
