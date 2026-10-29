package scripts_test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

const (
	installCommand    = "sudo apt-get install -y libgl1-mesa-dev xorg-dev iverilog"
	versionCommand    = "iverilog -V"
	verilogCommand    = "go test -v -count=1 ./validation/external/eda -run TestVerilogSanityCheck"
	ngspiceCommand    = "go test -v ./validation/external/circuit -run '^TestNgspiceCrossbar_'"
	oldNgspiceCommand = "go test -v ./validation/... -run 'NGSpice|Spice|External'"
)

type ciWorkflow struct {
	Jobs map[string]ciJob `yaml:"jobs"`
}

type ciJob struct {
	If              yaml.Node `yaml:"if"`
	ContinueOnError yaml.Node `yaml:"continue-on-error"`
	Steps           []ciStep  `yaml:"steps"`
}

type ciStep struct {
	Name             string    `yaml:"name"`
	Run              string    `yaml:"run"`
	WorkingDirectory string    `yaml:"working-directory"`
	If               yaml.Node `yaml:"if"`
	ContinueOnError  yaml.Node `yaml:"continue-on-error"`
}

func TestCIChecksFormattingWithoutMutatingFiles(t *testing.T) {
	makefile := readFile(t, filepath.Join(repoRoot(t), "Makefile"))
	ciLine := "ci: fmt-check vet govulncheck test-short arch-check test-research research-audit"
	if !strings.Contains(makefile, ciLine) {
		t.Fatalf("ci target must use %q", ciLine)
	}
	if strings.Contains(makefile, "ci: fmt vet") {
		t.Fatal("ci target must not run the mutating fmt target")
	}
	if !strings.Contains(makefile, "fmt-check:\n\t@test -z \"$$($(GOFMT) -l .)\"") {
		t.Fatal("fmt-check must fail on gofmt output without modifying files")
	}
}

func TestCIRequiresVulnerabilityScanAndScheduledDependencyUpdates(t *testing.T) {
	root := repoRoot(t)
	makefile := readFile(t, filepath.Join(root, "Makefile"))
	if !strings.Contains(makefile, "ci: fmt-check vet govulncheck test-short") || !strings.Contains(makefile, "govulncheck:\n\t$(GO) run golang.org/x/vuln/cmd/govulncheck@v1.1.4 ./...") {
		t.Fatal("CI must run the pinned govulncheck tool")
	}
	dependabot := readFile(t, filepath.Join(root, ".github", "dependabot.yml"))
	for _, ecosystem := range []string{"gomod", "github-actions"} {
		if !strings.Contains(dependabot, "package-ecosystem: "+ecosystem) {
			t.Errorf("scheduled dependency updates missing ecosystem %q", ecosystem)
		}
	}
	if !strings.Contains(dependabot, "interval: weekly") {
		t.Fatal("dependency updates must be scheduled weekly")
	}
}

func TestFyneLaunchersUseOneCanonicalImplementation(t *testing.T) {
	root := repoRoot(t)
	shared := filepath.Join(root, "internal", "launcher", "launcher.go")
	body := readFile(t, shared)
	for _, symbol := range []string{"type DemoInfo struct", "func CreateLauncherContent", "func DrawPreviewThumbnail"} {
		if !strings.Contains(body, symbol) {
			t.Errorf("shared launcher missing %q", symbol)
		}
	}
	for _, path := range []string{
		filepath.Join(root, "cmd", "fecim-lattice-tools-fyne", "launcher.go"),
		filepath.Join(root, "cmd", "fecim-lattice-tools", "fyne_launcher.go"),
	} {
		adapter := readFile(t, path)
		if !strings.Contains(adapter, `"fecim-lattice-tools/internal/launcher"`) || strings.Contains(adapter, "type demoCardRenderer struct") {
			t.Errorf("%s must delegate to the shared launcher rather than retain its implementation", path)
		}
	}
}

func TestCIRequiresRealExternalVerilogValidation(t *testing.T) {
	workflow := readFile(t, filepath.Join(repoRoot(t), ".github", "workflows", "ci.yml"))
	if err := validateExternalValidationWorkflow(workflow); err != nil {
		t.Fatal(err)
	}

	sanityPath := filepath.Join(repoRoot(t), "validation", "external", "eda", "verilog_sanity_test.go")
	sanity := readFile(t, sanityPath)
	if strings.Contains(sanity, "would run") || strings.Contains(sanity, "would invoke") {
		t.Error("Verilog sanity test still contains no-op validation wording")
	}
	for _, token := range []string{
		`exec.Command("iverilog"`,
		"GenerateVerilogWithDefaults",
		"GenerateArrayVerilog",
		"GenerateCellVerilog",
	} {
		if !strings.Contains(sanity, token) {
			t.Errorf("Verilog sanity test missing required source use %q", token)
		}
	}
}

func TestExternalValidationWorkflowRejectsNonGatingMutations(t *testing.T) {
	workflow := readFile(t, filepath.Join(repoRoot(t), ".github", "workflows", "ci.yml"))
	if err := validateExternalValidationWorkflow(workflow); err != nil {
		t.Fatalf("baseline workflow invalid: %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*testing.T, string) string
	}{
		{
			name: "job if false",
			mutate: func(t *testing.T, workflow string) string {
				return replaceOnce(t, workflow, "  external-validation:\n    runs-on:", "  external-validation:\n    if: false\n    runs-on:")
			},
		},
		{
			name: "job continue-on-error",
			mutate: func(t *testing.T, workflow string) string {
				return replaceOnce(t, workflow, "  external-validation:\n    runs-on:", "  external-validation:\n    continue-on-error: true\n    runs-on:")
			},
		},
		{
			name: "version step if false",
			mutate: func(t *testing.T, workflow string) string {
				return replaceOnce(t, workflow, "      - name: Record Icarus Verilog version\n        run: iverilog -V", "      - name: Record Icarus Verilog version\n        if: false\n        run: iverilog -V")
			},
		},
		{
			name: "Verilog step continue-on-error",
			mutate: func(t *testing.T, workflow string) string {
				return replaceOnce(t, workflow, "      - name: Run required external Verilog validation\n        run: "+verilogCommand, "      - name: Run required external Verilog validation\n        continue-on-error: true\n        run: "+verilogCommand)
			},
		},
		{
			name: "moved version step",
			mutate: func(t *testing.T, workflow string) string {
				workflow = replaceOnce(t, workflow, "      - name: Record Icarus Verilog version\n        run: iverilog -V\n\n", "")
				ciStart := "  ci:\n    runs-on: ubuntu-latest\n    timeout-minutes: 20\n\n    steps:\n      - name: Checkout"
				ciWithVersion := "  ci:\n    runs-on: ubuntu-latest\n    timeout-minutes: 20\n\n    steps:\n      - name: Record Icarus Verilog version\n        run: iverilog -V\n\n      - name: Checkout"
				return replaceOnce(t, workflow, ciStart, ciWithVersion)
			},
		},
		{
			name: "missing Verilog step",
			mutate: func(t *testing.T, workflow string) string {
				return replaceOnce(t, workflow, "      - name: Run required external Verilog validation\n        run: "+verilogCommand+"\n\n", "")
			},
		},
		{
			name: "Verilog test before install",
			mutate: func(t *testing.T, workflow string) string {
				verilogStep := "      - name: Run required external Verilog validation\n        run: " + verilogCommand
				workflow = replaceOnce(t, workflow, verilogStep+"\n\n", "")
				installStep := "      - name: Install GUI and external validation dependencies"
				return replaceOnce(t, workflow, installStep, verilogStep+"\n\n"+installStep)
			},
		},
		{
			name: "masked install command",
			mutate: func(t *testing.T, workflow string) string {
				return replaceOnce(t, workflow, installCommand, installCommand+" || true")
			},
		},
		{
			name: "masked version command",
			mutate: func(t *testing.T, workflow string) string {
				return replaceOnce(t, workflow, "run: "+versionCommand, "run: "+versionCommand+" || true")
			},
		},
		{
			name: "masked Verilog command",
			mutate: func(t *testing.T, workflow string) string {
				return replaceOnce(t, workflow, "run: "+verilogCommand, "run: "+verilogCommand+" || true")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mutated := test.mutate(t, workflow)
			if err := validateExternalValidationWorkflow(mutated); err == nil {
				t.Fatal("mutated workflow unexpectedly satisfied external-validation contract")
			}
		})
	}
}

func validateExternalValidationWorkflow(data string) error {
	var workflow ciWorkflow
	if err := yaml.Unmarshal([]byte(data), &workflow); err != nil {
		return fmt.Errorf("parse ci.yml: %w", err)
	}
	external, ok := workflow.Jobs["external-validation"]
	if !ok {
		return fmt.Errorf("ci.yml missing external-validation job")
	}
	if nodePresent(external.If) {
		return fmt.Errorf("external-validation job must not have if")
	}
	if nodePresent(external.ContinueOnError) {
		return fmt.Errorf("external-validation job must not have continue-on-error")
	}

	required := []struct {
		name  string
		match func(string) bool
	}{
		{name: "iverilog install", match: func(run string) bool { return hasCommandLine(run, installCommand) }},
		{name: "iverilog version", match: func(run string) bool { return strings.TrimSpace(run) == versionCommand }},
		{name: "Verilog test", match: func(run string) bool { return strings.TrimSpace(run) == verilogCommand }},
	}
	requiredIndices := make([]int, 0, len(required))
	for _, requirement := range required {
		step, index, err := findRunStep(external.Steps, requirement.name, requirement.match)
		if err != nil {
			return err
		}
		requiredIndices = append(requiredIndices, index)
		if nodePresent(step.If) {
			return fmt.Errorf("required %s step must not have if", requirement.name)
		}
		if nodePresent(step.ContinueOnError) {
			return fmt.Errorf("required %s step must not have continue-on-error", requirement.name)
		}
		if masksFailure(step.Run) {
			return fmt.Errorf("required %s step masks command failure", requirement.name)
		}
	}
	if requiredIndices[0] >= requiredIndices[1] || requiredIndices[1] >= requiredIndices[2] {
		return fmt.Errorf("external-validation requires install, version, then Verilog test step order")
	}

	ngspice, _, err := findRunStep(external.Steps, "ngspice test", func(run string) bool {
		return strings.TrimSpace(run) == ngspiceCommand
	})
	if err != nil {
		return err
	}
	if !nodePresent(ngspice.If) {
		return fmt.Errorf("ngspice test must remain optional with an if condition")
	}
	if nodePresent(ngspice.ContinueOnError) {
		return fmt.Errorf("ngspice test must not use continue-on-error")
	}

	foundOptionalNgspiceSummary := false
	for _, step := range external.Steps {
		if strings.TrimSpace(step.Run) == oldNgspiceCommand {
			return fmt.Errorf("external-validation contains obsolete zero-match ngspice filter")
		}
		text := step.Name + "\n" + step.Run + "\n" + step.If.Value
		foundOptionalNgspiceSummary = foundOptionalNgspiceSummary || strings.Contains(text, "Optional ngspice validation checks were skipped")
		for _, forbidden := range []string{
			"has_iverilog",
			"iverilog missing",
			"Optional external validation checks",
			"external validation tests will be skipped",
			"verilog simulation checks will be skipped",
		} {
			if strings.Contains(text, forbidden) {
				return fmt.Errorf("external-validation contains optional Verilog wording %q", forbidden)
			}
		}
	}
	if !foundOptionalNgspiceSummary {
		return fmt.Errorf("external-validation missing optional-ngspice-only summary")
	}
	return nil
}

func findRunStep(steps []ciStep, name string, match func(string) bool) (ciStep, int, error) {
	var found ciStep
	foundIndex := -1
	count := 0
	for index, step := range steps {
		if match(step.Run) {
			found = step
			foundIndex = index
			count++
		}
	}
	if count != 1 {
		return ciStep{}, -1, fmt.Errorf("external-validation requires exactly one %s run step, got %d", name, count)
	}
	return found, foundIndex, nil
}

func nodePresent(node yaml.Node) bool {
	return node.Kind != 0
}

func masksFailure(script string) bool {
	return strings.Contains(script, "|| true") || strings.Contains(script, "|| :")
}

func hasCommandLine(script, command string) bool {
	for _, line := range strings.Split(script, "\n") {
		if strings.TrimSpace(line) == command {
			return true
		}
	}
	return false
}

func replaceOnce(t *testing.T, input, old, replacement string) string {
	t.Helper()
	if strings.Count(input, old) != 1 {
		t.Fatalf("mutation target count for %q = %d, want 1", old, strings.Count(input, old))
	}
	return strings.Replace(input, old, replacement, 1)
}

func TestCIRunsWebTestsAndProductionBuild(t *testing.T) {
	workflow := readFile(t, filepath.Join(repoRoot(t), ".github", "workflows", "ci.yml"))
	var parsed ciWorkflow
	if err := yaml.Unmarshal([]byte(workflow), &parsed); err != nil {
		t.Fatal(err)
	}
	required := map[string]struct{ command, directory string }{
		"Test and build web app": {
			"npm ci && npm run generate:api && npm test && npm run build", "web",
		},
		"Reject generated API drift": {
			"git diff --exit-code -- workbench/moduleapi/openapi.json web/src/api/generated.ts", "",
		},
		"Typecheck and build both shared React profiles": {
			"npm run typecheck && npm run build:hosted && npm run build:local", "web",
		},
	}
	for _, step := range parsed.Jobs["ci"].Steps {
		want, ok := required[step.Name]
		if !ok {
			continue
		}
		if step.Run != want.command || step.WorkingDirectory != want.directory {
			t.Fatalf("%s: command=%q working-directory=%q", step.Name, step.Run, step.WorkingDirectory)
		}
		if nodePresent(step.ContinueOnError) {
			t.Fatalf("%s must be required", step.Name)
		}
		delete(required, step.Name)
	}
	for name := range required {
		t.Errorf("CI is missing required web step %q", name)
	}
}
