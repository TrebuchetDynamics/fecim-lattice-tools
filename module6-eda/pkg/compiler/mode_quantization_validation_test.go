package compiler

import "testing"

func TestNewArrayConfig_ModeSpecificDefaultsAndIsolation(t *testing.T) {
	storage := NewArrayConfig(ModeStorage, 8, 8)
	if storage.StorageConfig == nil {
		t.Fatal("storage mode must initialize StorageConfig")
	}
	if storage.MemoryConfig != nil || storage.ComputeConfig != nil {
		t.Fatal("storage mode must not initialize memory/compute configs")
	}

	memory := NewArrayConfig(ModeMemory, 8, 8)
	if memory.MemoryConfig == nil {
		t.Fatal("memory mode must initialize MemoryConfig")
	}
	if memory.StorageConfig != nil || memory.ComputeConfig != nil {
		t.Fatal("memory mode must not initialize storage/compute configs")
	}

	compute := NewArrayConfig(ModeCompute, 8, 8)
	if compute.ComputeConfig == nil {
		t.Fatal("compute mode must initialize ComputeConfig")
	}
	if compute.StorageConfig != nil || compute.MemoryConfig != nil {
		t.Fatal("compute mode must not initialize storage/memory configs")
	}
}

func TestGenerateDesign_BlankBehaviorAcrossModes(t *testing.T) {
	modes := []OperationMode{ModeStorage, ModeMemory, ModeCompute}
	for _, mode := range modes {
		t.Run(mode.String(), func(t *testing.T) {
			cfg := NewArrayConfig(mode, 4, 4)
			design, err := GenerateDesign(cfg)
			if err != nil {
				t.Fatalf("GenerateDesign failed: %v", err)
			}
			if design.Stats.ActiveCells != 0 {
				t.Fatalf("expected 0 active cells for blank %s mode, got %d", mode.String(), design.Stats.ActiveCells)
			}
			for i, c := range design.Cells {
				if c.Level != 0 {
					t.Fatalf("cell %d expected level=0, got %d", i, c.Level)
				}
				if c.Conductance != cfg.GMin {
					t.Fatalf("cell %d expected conductance=%v, got %v", i, cfg.GMin, c.Conductance)
				}
			}
		})
	}
}

func TestMapWeights_QuantizationSignAndMonotonicity(t *testing.T) {
	weights := [][]float64{{-1.0, -0.5, 0.0, 0.5, 1.0}}
	cfg := NewComputeConfig(2, 8)
	cfg.Levels = 30
	cfg.WithWeights(weights)

	design, err := GenerateDesign(cfg)
	if err != nil {
		t.Fatalf("GenerateDesign failed: %v", err)
	}

	if design.Stats.ActiveCells != 5 {
		t.Fatalf("expected 5 active cells, got %d", design.Stats.ActiveCells)
	}

	lNeg1 := design.Cells[0].Level
	lNegHalf := design.Cells[1].Level
	lZero := design.Cells[2].Level
	lPosHalf := design.Cells[3].Level
	lPos1 := design.Cells[4].Level

	if !(lNeg1 < lNegHalf && lNegHalf < lZero && lZero < lPosHalf && lPosHalf < lPos1) {
		t.Fatalf("levels not strictly monotonic: %d %d %d %d %d", lNeg1, lNegHalf, lZero, lPosHalf, lPos1)
	}

	if lNeg1+lPos1 != cfg.Levels-1 {
		t.Fatalf("expected symmetric quantization for +/-1.0, sum=%d want %d", lNeg1+lPos1, cfg.Levels-1)
	}
	if lNegHalf+lPosHalf != cfg.Levels-1 {
		t.Fatalf("expected symmetric quantization for +/-0.5, sum=%d want %d", lNegHalf+lPosHalf, cfg.Levels-1)
	}
	if lZero != 15 { // round((0.5)*(30-1)) => round(14.5)=15
		t.Fatalf("expected zero weight to map to middle level 15, got %d", lZero)
	}

	for i := 0; i < 5; i++ {
		c := design.Cells[i]
		if c.Conductance < cfg.GMin || c.Conductance > cfg.GMax {
			t.Fatalf("cell %d conductance out of bounds: %v not in [%v, %v]", i, c.Conductance, cfg.GMin, cfg.GMax)
		}
		if c.ProgramV < cfg.VProgMin || c.ProgramV > cfg.VProgMax {
			t.Fatalf("cell %d programV out of bounds: %v not in [%v, %v]", i, c.ProgramV, cfg.VProgMin, cfg.VProgMax)
		}
	}
}
