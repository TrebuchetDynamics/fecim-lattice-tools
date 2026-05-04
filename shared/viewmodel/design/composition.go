package design

import (
	"fmt"

	"fecim-lattice-tools/shared/viewmodel"
)

// Composition holds references to all active module ports and provides
// cross-module design aggregation.
type Composition struct {
	Hysteresis viewmodel.ModulePort
	Crossbar   viewmodel.ModulePort
	Circuits   viewmodel.ModulePort
	EDA        viewmodel.ModulePort
}

// DesignSnapshot aggregates state across the design pipeline:
// Material → Array → Circuits → Export.
type DesignSnapshot struct {
	Material      string
	ArrayRows     int
	ArrayCols     int
	ADCResolution int
	DACResolution int
	ProcessNode   string
	DesignName    string
	Summary       string
}

// Snapshot computes a unified design state from all modules.
func (c *Composition) Snapshot() DesignSnapshot {
	ds := DesignSnapshot{}

	if c.Hysteresis != nil {
		for _, m := range c.Hysteresis.Snapshot().Metrics {
			if m.ID == "material" {
				ds.Material = m.Value
			}
		}
	}
	if c.Crossbar != nil {
		for _, m := range c.Crossbar.Snapshot().Metrics {
			switch m.ID {
			case "rows":
				fmt.Sscanf(m.Value, "%d", &ds.ArrayRows)
			case "cols":
				fmt.Sscanf(m.Value, "%d", &ds.ArrayCols)
			}
		}
	}
	if c.Circuits != nil {
		for _, m := range c.Circuits.Snapshot().Metrics {
			switch m.ID {
			case "adc":
				fmt.Sscanf(m.Value, "%d-bit", &ds.ADCResolution)
			case "dac":
				fmt.Sscanf(m.Value, "%d-bit", &ds.DACResolution)
			}
		}
	}
	if c.EDA != nil {
		for _, m := range c.EDA.Snapshot().Metrics {
			switch m.ID {
			case "process":
				ds.ProcessNode = m.Value
			case "design":
				ds.DesignName = m.Value
			}
		}
	}

	ds.Summary = fmt.Sprintf("Design: %s | %s × %d×%d (%d-bit ADC/%d-bit DAC) @ %s",
		ds.DesignName, ds.Material, ds.ArrayRows, ds.ArrayCols,
		ds.ADCResolution, ds.DACResolution, ds.ProcessNode)
	return ds
}

// ExportDesign triggers export across the EDA module.
func (c *Composition) ExportDesign() error {
	if c.EDA == nil {
		return fmt.Errorf("design: EDA module not connected")
	}
	return c.EDA.ApplyAction(viewmodel.Action{
		ID:   "generate_all",
		Kind: viewmodel.ActionCommand,
	})
}
