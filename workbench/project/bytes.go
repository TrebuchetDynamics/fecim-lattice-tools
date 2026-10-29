package project

import (
	"bytes"
	"fmt"
	"path"
	"strings"
)

// LoadBytes loads the three fixed YAML documents and their referenced inputs.
// Keys are canonical slash-separated relative names; unreferenced documents are
// rejected. The bundle has no filesystem authority (Root is empty), citations
// are syntax-checked, and input hashes are verified against actual supplied bytes.
// Each YAML document is limited to 1 MiB; all supplied bytes total at most 8 MiB.
func LoadBytes(documents map[string][]byte) (Bundle, error) {
	return loadBytes(documents, 0, 0)
}

// LoadBytesLimited enforces admission before the common loader first expands a
// sweep. Limits must be positive; callers still validate scientific outcomes.
func LoadBytesLimited(documents map[string][]byte, maxPoints, maxDimension int) (Bundle, error) {
	if maxPoints < 1 || maxPoints > 100000 || maxDimension < 1 {
		return Bundle{}, fmt.Errorf("invalid admission limits")
	}
	return loadBytes(documents, maxPoints, maxDimension)
}

func loadBytes(documents map[string][]byte, maxPoints, maxDimension int) (Bundle, error) {
	const maxBundleBytes = 8 << 20
	total := 0
	for name, data := range documents {
		if name == "." || path.IsAbs(name) || path.Clean(name) != name || strings.HasPrefix(name, "../") || strings.ContainsAny(name, "\\\x00:") {
			return Bundle{}, fmt.Errorf("invalid document name")
		}
		if len(data) > maxBundleBytes-total {
			return Bundle{}, fmt.Errorf("bundle exceeds %d bytes", maxBundleBytes)
		}
		total += len(data)
	}
	read := func(name string) ([]byte, error) {
		data, ok := documents[name]
		if !ok {
			return nil, fmt.Errorf("missing bundle document %q", name)
		}
		return data, nil
	}
	bundle, err := loadBundle("", func(name string, dst any) error {
		data, err := read(name)
		if err != nil {
			return err
		}
		if err := decodeStrict(bytes.NewReader(data), dst); err != nil {
			return err
		}
		if maxPoints > 0 {
			switch v := dst.(type) {
			case *Design:
				if v.Array.Rows > maxDimension || v.Array.Cols > maxDimension {
					return fmt.Errorf("array exceeds admitted dimension")
				}
			case *Sweep:
				limited := *v
				if limited.MaxPoints <= 0 || limited.MaxPoints > maxPoints {
					limited.MaxPoints = maxPoints
				}
				if err := preflightConfigurationSweep(limited); err != nil {
					return err
				}
				for _, p := range v.Parameters {
					if p.Path != "array.rows" && p.Path != "array.cols" {
						continue
					}
					for _, value := range p.Values {
						if value > float64(maxDimension) {
							return fmt.Errorf("sweep exceeds admitted dimension")
						}
					}
					if p.Range != nil && (p.Range.Start > float64(maxDimension) || p.Range.Stop > float64(maxDimension)) {
						return fmt.Errorf("sweep exceeds admitted dimension")
					}
				}
			}
		}
		return nil
	}, func(refs []InputRef) ([]ResolvedInput, error) { return resolveInputBytes(read, refs) }, nil, true)
	if err != nil {
		return Bundle{}, err
	}
	allowed := map[string]bool{"project.yaml": true, "design.yaml": true, "sweep.yaml": true}
	for _, input := range bundle.Inputs {
		allowed[input.Path] = true
	}
	for name := range documents {
		if !allowed[name] {
			return Bundle{}, fmt.Errorf("unreferenced bundle document %q", name)
		}
	}
	return bundle, nil
}
