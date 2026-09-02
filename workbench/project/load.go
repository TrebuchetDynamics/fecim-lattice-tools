package project

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

const maxProjectFileBytes int64 = 1 << 20

var idRE = regexp.MustCompile(`^[a-z][a-z0-9-]{1,62}$`)
var citationRE = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]*$`)

func Load(root string, opts LoadOptions) (Bundle, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return Bundle{}, err
	}
	bundle := Bundle{Root: abs}
	if err := decodeStrict(filepath.Join(abs, "project.yaml"), &bundle.Project); err != nil {
		return Bundle{}, fmt.Errorf("project.yaml: %w", err)
	}
	if err := decodeStrict(filepath.Join(abs, "design.yaml"), &bundle.Design); err != nil {
		return Bundle{}, fmt.Errorf("design.yaml: %w", err)
	}
	if err := decodeStrict(filepath.Join(abs, "sweep.yaml"), &bundle.Sweep); err != nil {
		return Bundle{}, fmt.Errorf("sweep.yaml: %w", err)
	}
	if bundle.Project.SchemaVersion != 1 || bundle.Design.SchemaVersion != 1 || bundle.Sweep.SchemaVersion != 1 {
		return Bundle{}, errors.New("all schema_version values must be 1")
	}
	if !idRE.MatchString(bundle.Project.ID) || strings.TrimSpace(bundle.Project.Hypothesis) == "" || len(bundle.Project.Objectives) == 0 {
		return Bundle{}, errors.New("project requires a valid id, hypothesis, and objective")
	}
	if bundle.Project.ModelVersion == "" {
		return Bundle{}, errors.New("project model_version is required")
	}
	for _, objective := range bundle.Project.Objectives {
		if objective.Metric == "" || (objective.Direction != Minimize && objective.Direction != Maximize) {
			return Bundle{}, fmt.Errorf("invalid objective %+v", objective)
		}
	}
	for _, constraint := range bundle.Project.Constraints {
		if constraint.Metric == "" || constraint.Unit == "" || !finite(constraint.Value) {
			return Bundle{}, fmt.Errorf("invalid constraint %+v", constraint)
		}
		switch constraint.Operator {
		case "<", "<=", ">", ">=", "==":
		default:
			return Bundle{}, fmt.Errorf("unsupported constraint operator %q", constraint.Operator)
		}
	}
	if err := ValidateDesign(bundle.Design); err != nil {
		return Bundle{}, err
	}
	if bundle.Sweep.MaxPoints <= 0 {
		bundle.Sweep.MaxPoints = 10_000
	}
	if bundle.Sweep.MaxPoints > 100_000 {
		return Bundle{}, errors.New("max_points must not exceed 100000")
	}
	if len(bundle.Sweep.Parameters) == 0 {
		return Bundle{}, errors.New("sweep requires at least one parameter")
	}

	citations := append([]string(nil), bundle.Project.Citations...)
	citations = append(citations, inputCitations(bundle.Project.Inputs)...)
	for _, key := range citations {
		if !citationRE.MatchString(key) {
			return Bundle{}, fmt.Errorf("invalid citation key %q", key)
		}
		if opts.CitationDir != "" {
			if _, err := os.Stat(filepath.Join(opts.CitationDir, key+".md")); err != nil {
				return Bundle{}, fmt.Errorf("citation %s: %w", key, err)
			}
		}
	}
	inputs, err := resolveInputs(abs, bundle.Project.Inputs)
	if err != nil {
		return Bundle{}, err
	}
	bundle.Inputs = inputs
	return bundle, nil
}

func decodeStrict(path string, dst any) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.Size() > maxProjectFileBytes {
		return fmt.Errorf("file exceeds %d bytes", maxProjectFileBytes)
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(io.LimitReader(file, maxProjectFileBytes+1))
	decoder.KnownFields(true)
	if err := decoder.Decode(dst); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("multiple YAML documents are not allowed")
	}
	return nil
}

func ValidateDesign(design Design) error {
	if design.Device.Material == "" || design.Device.ConductanceLevels < 2 || !finite(design.Device.GMinS) || !finite(design.Device.GMaxS) || design.Device.GMinS <= 0 || design.Device.GMaxS <= design.Device.GMinS {
		return errors.New("invalid device configuration")
	}
	if design.Array.Rows <= 0 || design.Array.Cols <= 0 || !finite(design.Array.ReadVoltageV) || design.Array.ReadVoltageV <= 0 {
		return errors.New("invalid array configuration")
	}
	if design.Circuit.ADCBits < 1 || design.Circuit.ADCBits > 16 || design.Circuit.DACBits < 1 || design.Circuit.DACBits > 16 || !finite(design.Circuit.TIAGainOhm) || design.Circuit.TIAGainOhm <= 0 {
		return errors.New("invalid circuit configuration")
	}
	switch design.Circuit.TechNode {
	case "130nm", "65nm", "28nm", "22nm", "14nm":
	default:
		return fmt.Errorf("unsupported tech_node %q", design.Circuit.TechNode)
	}
	return nil
}

func resolveInputs(root string, refs []InputRef) ([]ResolvedInput, error) {
	out := make([]ResolvedInput, 0, len(refs))
	for _, ref := range refs {
		clean := filepath.Clean(ref.Path)
		if filepath.IsAbs(clean) || clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
			return nil, fmt.Errorf("unsafe input path %q", ref.Path)
		}
		resolved, err := filepath.EvalSymlinks(filepath.Join(root, clean))
		if err != nil {
			return nil, err
		}
		rel, err := filepath.Rel(root, resolved)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return nil, fmt.Errorf("input escapes project root: %q", ref.Path)
		}
		data, err := os.ReadFile(resolved)
		if err != nil {
			return nil, err
		}
		sum := sha256.Sum256(data)
		actual := hex.EncodeToString(sum[:])
		if actual != strings.ToLower(ref.SHA256) {
			return nil, fmt.Errorf("input %s sha256 mismatch: got %s", ref.Path, actual)
		}
		switch ref.Evidence {
		case "literature-backed", "experiment-calibrated", "simulation-default":
		default:
			return nil, fmt.Errorf("input %s has invalid evidence %q", ref.Path, ref.Evidence)
		}
		out = append(out, ResolvedInput{
			Path:     filepath.ToSlash(clean),
			SHA256:   actual,
			Citation: ref.Citation,
			Evidence: ref.Evidence,
		})
	}
	return out, nil
}

func inputCitations(inputs []InputRef) []string {
	out := make([]string, 0, len(inputs))
	for _, input := range inputs {
		if input.Citation != "" {
			out = append(out, input.Citation)
		}
	}
	return out
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
