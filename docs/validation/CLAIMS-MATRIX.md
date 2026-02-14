# Scientific Claims Matrix

This matrix maps each reported scientific claim to executable verification artifacts.

| Claim | Test/Script | Evidence Path | Tolerance/Criterion |
|-------|-------------|---------------|---------------------|
| Pr=19.17 µC/cm² (HZO, Materlik 2015) | TestPhysicsRegression_Preisach | validation/testdata/ | ±5% |
| Ec=1.16 MV/cm | TestPhysicsRegression_LK | validation/testdata/ | ±5% |
| MNIST 80% accuracy | TestFullStackMNIST | validation/ | ≥80% |
| Energy 44.94 fJ/cell | TestTransientPulse | shared/physics/ | 10-100 fJ range |
| **(M4)** Power conservation < 1% | `scripts/module4_automation.sh --fast` (TestThermodynamics*) | module4-circuits/pkg/arraysim/ | < 1% |
| **(M4)** KCL residual < 1e-12 | `scripts/module4_automation.sh --fast` (TestKirchhoff*) | module4-circuits/pkg/arraysim/ | < 1e-12 |
| **(M4)** MVM accuracy < 5% (Tier-A) | `scripts/module4_automation.sh --full` (TestComputeMVM*) | module4-circuits/pkg/gui/ | < 5% |
| **(M4)** BER < 5% | `scripts/module4_automation.sh --full` (TestComputeBER*) | module4-circuits/pkg/gui/ | < 5% |
| **(M4)** Read margin > 3σ | `scripts/module4_automation.sh --full` (TestReadMarginBER*) | module4-circuits/pkg/gui/ | > 3σ |
| **(M4)** DAC INL < 1 LSB | `scripts/module4_automation.sh --full` (TestPeripheralsINLDNL*) | shared/peripherals/ | < 1 LSB |
| **(M4)** Retention ΔG/G < 1% | `scripts/module4_automation.sh --full` (TestRetention*) | module4-circuits/pkg/gui/ | < 1% |
| **(M4)** Write disturb bounded | `scripts/module4_automation.sh --full` (TestWriteDisturb*) | module4-circuits/pkg/gui/ | bounded |
| Preisach discontinuities physical | TestHeadlessISPPContinuityValidation | cmd/fecim-lattice-tools/ | 0 spurious |
| ISPP converges all targets | TestISPPConverges_Preisach | module1-hysteresis/pkg/controller/ | all levels hit |
| Array ISPP with disturb tracking | TestArrayISPP | shared/physics/ | MaxDisturb < 0.3 |

## Known Limits

- VK-1
- VK-2
- VK-3
- L07
- L08
- L09
- L10

