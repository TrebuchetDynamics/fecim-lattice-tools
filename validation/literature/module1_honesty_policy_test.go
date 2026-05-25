package literature

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestModule1MaterialClaimWordingAvoidsUnverifiedDemonstratedLanguage(t *testing.T) {
	t.Parallel()

	repoRoot := filepath.Join("..", "..")
	checks := []struct {
		path      string
		forbidden []string
	}{
		{
			path: filepath.Join(repoRoot, "shared", "physics", "material.go"),
			forbidden: []string{
				"conference-baseline demonstrated values",
				"only demonstrated values",
				"Best-case from Cheema et al. 2020",
				"CAN theoretically achieve",
			},
		},
		{
			path: filepath.Join(repoRoot, "shared", "physics", "material_presets.go"),
			forbidden: []string{
				"DEMONSTRATED: 30 states",
				"NumLevels:           30, // Conference claim; pending peer review",
				"EnduranceCycles:     1e9, // DEMONSTRATED",
				"RetentionTime:       1e7, // DEMONSTRATED",
				"RECORD",
				"verified; 10^12",
				"verified",
				"demonstrated ~10ns",
				"cycles demonstrated",
				"VERIFIED (",
				"THEORETICAL BEST",
				"Best-in-class",
			},
		},
		{
			path: filepath.Join(repoRoot, "shared", "physics", "AGENTS.md"),
			forbidden: []string{
				"FeCIMMaterial (conservative)",
				"LiteratureSuperlattice (best-case)",
				"Best-case: HfO2/ZrO2",
				"\"World-class\" performance benchmarks",
				"Conservative: Only values supported",
				"Eight worldclass benchmarks validate against literature claims",
				"Use for claims validation.",
			},
		},
		{
			path: filepath.Join(repoRoot, "shared", "physics", "material_test.go"),
			forbidden: []string{
				"FeCIM demonstrated endurance should be 1e9",
			},
		},
		{
			path: filepath.Join(repoRoot, "module1-hysteresis", "pkg", "ferroelectric", "material.go"),
			forbidden: []string{
				"demonstrating 32 analog states",
			},
		},
		{
			path: filepath.Join(repoRoot, "module1-hysteresis", "pkg", "ferroelectric", "AGENTS.md"),
			forbidden: []string{
				"FeCIMMaterial() for conservative demonstrated values",
				"best-case academic exploration",
			},
		},
		{
			path: filepath.Join(repoRoot, "module1-hysteresis", "AGENTS.md"),
			forbidden: []string{
				"32 analog states (conference demonstration)",
			},
		},
		{
			path: filepath.Join(repoRoot, "docs", "2-learn", "module1-hysteresis", "features.md"),
			forbidden: []string{
				"World-Class Physics Features",
				"research-grade characterization",
				"world-class features",
				"These features are implemented in `shared/physics/worldclass_*.go` and provide research-grade characterization",
				"Wake-up physics model",
			},
		},
		{
			path: filepath.Join(repoRoot, "docs", "2-learn", "module1-hysteresis", "materials.md"),
			forbidden: []string{
				"Each ferroelectric cell can be programmed to one of 30 distinct polarization states",
				"FeCIM: 4.91 bits/cell",
				"FeCIM (demonstrated)",
				"FeCIM HZO (demonstrated values)",
				"| Advantage | **1×** | 50× worse | 1000× worse | 5× worse |",
				"FeCIM: 25.5 TOPS/W (5× better than TPU)",
			},
		},
		{
			path: filepath.Join(repoRoot, "docs", "2-learn", "module1-hysteresis", "eli5.md"),
			forbidden: []string{
				"More states per cell = store more information!",
				"30 Discrete Levels",
				"Demo baseline for discrete storage states",
			},
		},
		{
			path: filepath.Join(repoRoot, "module1-hysteresis", "pkg", "gui", "info.go"),
			forbidden: []string{
				"Endurance: %.0e cycles [demonstrated: 10⁹-10¹²]",
				"Same chip, 5× more storage!",
				"Note: Ranges from peer-reviewed literature.",
			},
		},
		{
			path: filepath.Join(repoRoot, "module1-hysteresis", "pkg", "ferroelectric", "render.go"),
			forbidden: []string{
				"Discrete Analog States (demo baseline; conference claim)",
				"Total states: %d  (%.1f bits/cell)",
			},
		},
		{
			path: filepath.Join(repoRoot, "module1-hysteresis", "pkg", "render", "render.go"),
			forbidden: []string{
				"30 discrete FeCIM levels",
				"30-level indicator (conference-claim baseline)",
			},
		},
		{
			path: filepath.Join(repoRoot, "module1-hysteresis", "cmd", "hysteresis", "main.go"),
			forbidden: []string{
				"30 Discrete States (conference-claim baseline)",
				"Peer-reviewed ferroelectric devices report roughly 32-140 states in related literature.",
			},
		},
		{
			path: filepath.Join(repoRoot, "module1-hysteresis", "cmd", "hysteresis-fyne", "main.go"),
			forbidden: []string{
				"30 Discrete States (conference-claim baseline)",
				"Peer-reviewed ferroelectric devices report roughly 32-140 states in related literature.",
			},
		},
		{
			path: filepath.Join(repoRoot, "docs", "2-learn", "module1-hysteresis", "physics.md"),
			forbidden: []string{
				"30 levels | Linear discretization of P | ✅ Simple & correct",
				"Minor loops | Implicit via hysteron states | ✅ Works correctly",
				"Physics must correctly predict intermediate states.",
			},
		},
		{
			path: filepath.Join(repoRoot, "module1-hysteresis", "pkg", "ferroelectric", "physics_validation_test.go"),
			forbidden: []string{
				"demonstrated in literature",
				"outside demonstrated range",
			},
		},
		{
			path: filepath.Join(repoRoot, "module1-hysteresis", "pkg", "gui", "sim_loop.go"),
			forbidden: []string{
				"peer-reviewed literature reports 25-100×",
				"~50× est.",
				"~1000× est.",
				"FeCIM: only 50 cells! (5× denser)",
			},
		},
		{
			path: filepath.Join(repoRoot, "shared", "physics", "quantization.go"),
			forbidden: []string{
				"discrete FeCIM levels",
				"standard function for FeCIM devices with 30 levels",
			},
		},
		{
			path: filepath.Join(repoRoot, "module1-hysteresis", "pkg", "tui", "tui.go"),
			forbidden: []string{
				"30 Levels (claim)",
			},
		},
	}

	for _, check := range checks {
		check := check
		t.Run(filepath.ToSlash(check.path), func(t *testing.T) {
			t.Parallel()
			contents, err := os.ReadFile(check.path)
			if err != nil {
				t.Fatalf("read claim source: %v", err)
			}
			text := string(contents)
			for _, forbidden := range check.forbidden {
				if strings.Contains(text, forbidden) {
					t.Errorf("%s contains unverified-as-demonstrated wording %q", check.path, forbidden)
				}
			}
		})
	}
}
