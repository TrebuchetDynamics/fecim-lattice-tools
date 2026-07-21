package fecim

import (
	"context"
	"fmt"
	"math"

	"fecim-lattice-tools/shared/peripherals"
	"fecim-lattice-tools/shared/system"
	"fecim-lattice-tools/workbench/experiment"
	"fecim-lattice-tools/workbench/project"
)

const EvaluatorVersion = "fecim-system-v1"

func Evaluate(ctx context.Context, design project.Design, seed int64) (experiment.Result, error) {
	if err := ctx.Err(); err != nil {
		return experiment.Result{}, err
	}
	if design.Device.Material != "HZO" {
		return failed("unsupported-material", fmt.Sprintf("material %q is not supported by %s", design.Device.Material, EvaluatorVersion)), nil
	}
	node, ok := technologyNode(design.Circuit.TechNode)
	if !ok {
		return failed("unsupported-tech-node", fmt.Sprintf("technology node %q is unsupported", design.Circuit.TechNode)), nil
	}
	if err := project.ValidateDesign(design); err != nil {
		return failed("model-domain", err.Error()), nil
	}
	_ = seed // The current evaluator is deterministic; seed remains part of run identity.

	dac := peripherals.DefaultDAC()
	dac.Bits = design.Circuit.DACBits
	adc := peripherals.DefaultADC()
	adc.Bits = design.Circuit.ADCBits
	tia := peripherals.DefaultTIA()
	tia.Gain = design.Circuit.TIAGainOhm

	latencyModel := system.NewLatencyModel(design.Array.Rows, design.Array.Cols, node)
	latencyNS := dac.SettleTime + latencyModel.CrossbarSettlingNS() + latencyModel.ADCLatencyNS(adc.Bits) + tia.SettlingTime()*1e9

	energy := system.EstimateMLPEnergyJ(system.MLPEnergyConfig{
		LayerSizes:     []int{design.Array.Rows, design.Array.Cols},
		LevelsPerLayer: []int{design.Device.ConductanceLevels},
		EnergyPerDACJ:  dac.EnergyPerConversion(),
		EnergyPerADCJ:  adc.EnergyPerConversion(),
	})
	tiaEnergyJ := float64(design.Array.Cols) * tia.PowerConsumption() * tia.SettlingTime()
	energyPJ := (energy.TotalJ + tiaEnergyJ) * 1e12

	area := system.NewCrossbarAreaModel(design.Array.Rows, design.Array.Cols, node, system.CellFeFET).TotalAreaUM2(adc.Bits)
	meanConductance := (design.Device.GMinS + design.Device.GMaxS) / 2
	columnCurrent := float64(design.Array.Rows) * design.Array.ReadVoltageV * meanConductance
	snrDB := tia.SNR(columnCurrent)
	if math.IsNaN(snrDB) || math.IsInf(snrDB, 0) {
		return failed("model-domain", "TIA SNR is non-finite"), nil
	}

	assumptions := []string{"literature-calibrated pre-silicon model", "not measured silicon"}
	return experiment.Result{
		Status: experiment.StatusSuccess,
		Metrics: []experiment.Metric{
			{Name: "area_um2", Value: area, Unit: "um2", Model: "shared/system.CrossbarAreaModel", Assumptions: assumptions, Evidence: experiment.EvidenceDefault},
			{Name: "energy_pj", Value: energyPJ, Unit: "pJ", Model: "shared/system.EstimateMLPEnergyJ+shared/peripherals.TIA", Assumptions: assumptions, Evidence: experiment.EvidenceDerived},
			{Name: "latency_ns", Value: latencyNS, Unit: "ns", Model: "shared/system.LatencyModel+shared/peripherals", Assumptions: assumptions, Evidence: experiment.EvidenceDerived},
			{Name: "tia_snr_db", Value: snrDB, Unit: "dB", Model: "shared/peripherals.TIA", Assumptions: assumptions, Evidence: experiment.EvidenceDefault},
		},
		Warnings: []string{"simulation estimates only; calibrate before quantitative device claims"},
	}, nil
}

func failed(kind, message string) experiment.Result {
	return experiment.Result{Status: experiment.StatusFailed, Failure: &experiment.Failure{Kind: kind, Message: message}}
}

func technologyNode(raw string) (system.TechnologyNode, bool) {
	switch raw {
	case "130nm":
		return system.Node130nm, true
	case "65nm":
		return system.Node65nm, true
	case "28nm":
		return system.Node28nm, true
	case "22nm":
		return system.Node22nm, true
	case "14nm":
		return system.Node14nm, true
	default:
		return "", false
	}
}
