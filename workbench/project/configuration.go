package project

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

var configurationDocuments = [...]string{"project.yaml", "design.yaml", "sweep.yaml"}

// ConfigurationSnapshot owns the original bytes of the three configuration
// documents. It is a revision snapshot, not proof of semantic validity.
// Strings keep copies of snapshots from sharing caller-mutable bytes or nodes.
type ConfigurationSnapshot struct {
	documents [3]string
	revision  string
}

// Revision returns the quoted, strong cfg-v1 ETag over the original bytes.
func (s ConfigurationSnapshot) Revision() string { return s.revision }

func configurationSnapshot(documents [3]string) ConfigurationSnapshot {
	hash := sha256.New()
	hash.Write([]byte("fecim-config-v1\x00"))
	for i, name := range configurationDocuments {
		binary.Write(hash, binary.BigEndian, uint64(len(name)))
		io.WriteString(hash, name)
		binary.Write(hash, binary.BigEndian, uint64(len(documents[i])))
		io.WriteString(hash, documents[i])
	}
	return ConfigurationSnapshot{documents: documents, revision: fmt.Sprintf(`"cfg-v1-%x"`, hash.Sum(nil))}
}

// ReadConfiguration captures bounded regular configuration files through the
// caller's root. Unlike LoadRoot's legacy reads, configuration links are refused.
// The caller retains ownership of root.
func ReadConfiguration(root *os.Root) (ConfigurationSnapshot, error) {
	if root == nil {
		return ConfigurationSnapshot{}, fmt.Errorf("project root is required")
	}
	var documents [3]string
	for i, name := range configurationDocuments {
		data, err := readConfigurationDocument(root, name)
		if err != nil {
			return ConfigurationSnapshot{}, fmt.Errorf("%s: %w", name, err)
		}
		documents[i] = string(data)
	}
	return configurationSnapshot(documents), nil
}

func readConfigurationDocument(root *os.Root, name string) ([]byte, error) {
	return readRegularRoot(root, name, maxProjectFileBytes, false)
}

// FieldChange replaces an existing allowlisted scalar with typed literal input,
// never with a YAML fragment. Path uses slash-separated schema keys and indices.
type FieldChange struct {
	Path  string
	Value string
}

// ConfigurationFieldIssue identifies a scalar input/allowlist error without
// exposing YAML source, input literals, filesystem paths or parser internals.
type ConfigurationFieldIssue struct{ Document, Path, Reason string }

func (e *ConfigurationFieldIssue) Error() string { return e.Reason }

// PreparedConfiguration privately retains candidate bytes and a validated bundle
// for persistence inside this package. Preparation itself performs no writes.
type PreparedConfiguration struct {
	snapshot ConfigurationSnapshot
	bundle   Bundle
	document string
	changed  bool
}

func (p PreparedConfiguration) Changed() bool { return p.changed }

// PrepareConfiguration edits only the named document and validates the entire
// candidate using the caller's root-bound input and citation authority. A no-op
// retains every original byte without invoking a YAML encoder.
func PrepareConfiguration(root, citations *os.Root, base ConfigurationSnapshot, document string, changes []FieldChange) (PreparedConfiguration, error) {
	index := configurationDocumentIndex(document)
	if index < 0 {
		return PreparedConfiguration{}, fmt.Errorf("unsupported configuration document")
	}
	if base.revision == "" {
		return PreparedConfiguration{}, fmt.Errorf("configuration snapshot is required")
	}
	if len(changes) > 64 {
		return PreparedConfiguration{}, fmt.Errorf("patch exceeds 64 changes")
	}
	candidate := base
	changed := false
	if len(changes) != 0 {
		// Inspect all three documents: even a semantic no-op edit must not guess
		// at ambiguous keys, aliases or merges in the candidate bundle.
		var nodes [3]yaml.Node
		for i, data := range base.documents {
			if err := decodeStrict(strings.NewReader(data), &nodes[i]); err != nil {
				return PreparedConfiguration{}, err
			}
			if err := editableConfigurationTree(&nodes[i]); err != nil {
				return PreparedConfiguration{}, err
			}
		}
		seen := make(map[string]bool, len(changes))
		for _, change := range changes {
			if len(change.Value) > 4096 || !utf8.ValidString(change.Value) {
				return PreparedConfiguration{}, fmt.Errorf("scalar input must be valid UTF-8 and at most 4096 bytes")
			}
			parts, err := configurationPath(change.Path)
			if err != nil {
				return PreparedConfiguration{}, &ConfigurationFieldIssue{Document: document, Path: change.Path, Reason: "Use an existing editable field path"}
			}
			kind := configurationFieldKind(document, parts)
			if kind == "" {
				return PreparedConfiguration{}, &ConfigurationFieldIssue{Document: document, Path: change.Path, Reason: "This field is not editable"}
			}
			key := strings.Join(parts, "/")
			if seen[key] {
				return PreparedConfiguration{}, fmt.Errorf("duplicate field target")
			}
			seen[key] = true
			node, err := configurationScalar(&nodes[index], parts)
			if err != nil {
				return PreparedConfiguration{}, &ConfigurationFieldIssue{Document: document, Path: change.Path, Reason: "Choose an existing scalar field"}
			}
			different, err := replaceConfigurationScalar(node, kind, change.Value)
			if err != nil {
				reason := "Enter a valid literal text value"
				if kind == "float" {
					reason = "Enter a finite number in the displayed base unit"
				}
				if kind == "int" || kind == "int64" {
					reason = "Enter a base-10 integer within the supported range"
				}
				return PreparedConfiguration{}, &ConfigurationFieldIssue{Document: document, Path: change.Path, Reason: reason}
			}
			changed = changed || different
		}
		if changed {
			var output bytes.Buffer
			encoder := yaml.NewEncoder(&output)
			encoder.SetIndent(2)
			if err := encoder.Encode(&nodes[index]); err != nil {
				return PreparedConfiguration{}, err
			}
			if err := encoder.Close(); err != nil {
				return PreparedConfiguration{}, err
			}
			if int64(output.Len()) > maxProjectFileBytes {
				return PreparedConfiguration{}, fmt.Errorf("file exceeds %d bytes", maxProjectFileBytes)
			}
			documents := base.documents
			documents[index] = output.String()
			candidate = configurationSnapshot(documents)
		}
	}
	bundle, err := loadConfigurationRoot(root, citations, &candidate)
	if err != nil {
		return PreparedConfiguration{}, err
	}
	return PreparedConfiguration{snapshot: candidate, bundle: bundle, document: document, changed: changed}, nil
}

func configurationDocumentIndex(name string) int {
	for i, document := range configurationDocuments {
		if name == document {
			return i
		}
	}
	return -1
}

func editableConfigurationTree(node *yaml.Node) error {
	if node.Kind == yaml.AliasNode || node.Tag == "!!merge" {
		return fmt.Errorf("alias and merge edits are not supported")
	}
	if node.Kind == yaml.MappingNode {
		seen := make(map[string]bool, len(node.Content)/2)
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i]
			if key.Kind != yaml.ScalarNode || key.Tag != "!!str" || seen[key.Value] {
				return fmt.Errorf("ambiguous configuration mapping key")
			}
			seen[key.Value] = true
		}
	}
	for _, child := range node.Content {
		if err := editableConfigurationTree(child); err != nil {
			return err
		}
	}
	return nil
}

func configurationPath(path string) ([]string, error) {
	if !strings.HasPrefix(path, "/") {
		return nil, fmt.Errorf("field path must start with a slash")
	}
	parts := strings.Split(path[1:], "/")
	for i, part := range parts {
		for j := 0; j < len(part); j++ {
			if part[j] == '~' {
				j++
				if j == len(part) || (part[j] != '0' && part[j] != '1') {
					return nil, fmt.Errorf("invalid field path escape")
				}
			}
		}
		parts[i] = strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")
	}
	return parts, nil
}

func configurationIndex(part string) (int, bool) {
	index, err := strconv.Atoi(part)
	return index, err == nil && index >= 0 && strconv.Itoa(index) == part
}

// Kinds are schema types, not the input's inferred YAML tag.
func configurationFieldKind(document string, p []string) string {
	path := strings.Join(p, "/")
	switch document {
	case "project.yaml":
		if len(p) == 1 && (path == "name" || path == "hypothesis") {
			return "string"
		}
		if len(p) == 3 && p[0] == "constraints" {
			if _, ok := configurationIndex(p[1]); ok {
				switch p[2] {
				case "value":
					return "float"
				case "operator":
					return "string"
				}
			}
		}
	case "design.yaml":
		if len(p) != 2 {
			return ""
		}
		switch path {
		case "device/material", "circuit/tech_node":
			return "string"
		case "device/conductance_levels", "array/rows", "array/cols", "circuit/adc_bits", "circuit/dac_bits":
			return "int"
		case "device/g_min_s", "device/g_max_s", "array/read_voltage_v", "circuit/tia_gain_ohm":
			return "float"
		}
	case "sweep.yaml":
		if len(p) == 1 {
			switch path {
			case "seed":
				return "int64"
			case "max_points":
				return "int"
			}
		}
		if len(p) == 4 && p[0] == "parameters" {
			if _, ok := configurationIndex(p[1]); ok {
				if p[2] == "values" {
					if _, ok := configurationIndex(p[3]); ok {
						return "float"
					}
				}
				if p[2] == "range" {
					switch p[3] {
					case "start", "stop":
						return "float"
					case "count":
						return "int"
					}
				}
			}
		}
	}
	return ""
}

func configurationScalar(node *yaml.Node, parts []string) (*yaml.Node, error) {
	if node.Kind != yaml.DocumentNode || len(node.Content) != 1 {
		return nil, fmt.Errorf("configuration must contain one document")
	}
	node = node.Content[0]
	for _, part := range parts {
		var next *yaml.Node
		switch node.Kind {
		case yaml.MappingNode:
			for i := 0; i < len(node.Content); i += 2 {
				if node.Content[i].Value == part {
					next = node.Content[i+1]
					break
				}
			}
		case yaml.SequenceNode:
			if index, ok := configurationIndex(part); ok && index < len(node.Content) {
				next = node.Content[index]
			}
		}
		if next == nil {
			return nil, fmt.Errorf("field path does not identify an existing node")
		}
		node = next
	}
	if node.Kind != yaml.ScalarNode {
		return nil, fmt.Errorf("field path must identify an existing scalar")
	}
	return node, nil
}

func replaceConfigurationScalar(node *yaml.Node, kind, input string) (bool, error) {
	var value any
	var old any
	switch kind {
	case "string":
		var existing string
		if err := node.Decode(&existing); err != nil {
			return false, err
		}
		old, value = existing, input
	case "int", "int64":
		bits := 64
		if kind == "int" {
			bits = strconv.IntSize
		}
		number, err := strconv.ParseInt(input, 10, bits)
		if err != nil {
			return false, fmt.Errorf("field requires a base-10 integer within range")
		}
		var existing int64
		if err := node.Decode(&existing); err != nil {
			return false, err
		}
		old, value = existing, number
	case "float":
		number, err := strconv.ParseFloat(input, 64)
		if err != nil || !finite(number) {
			return false, fmt.Errorf("field requires a finite float64")
		}
		var existing float64
		if err := node.Decode(&existing); err != nil {
			return false, err
		}
		old, value = existing, number
	}
	if old == value {
		return false, nil
	}
	var replacement yaml.Node
	if err := replacement.Encode(value); err != nil {
		return false, err
	}
	// Keep comments, ordering, anchors and presentation style on the owned node.
	// Only the scalar's schema tag and value change.
	node.Tag, node.Value = replacement.Tag, replacement.Value
	return true, nil
}

// Preflight only the editor path before Expand can allocate range/value copies
// or convert integer coordinates. Scientific validation and default resolution
// remain in the common loader.
func preflightConfigurationSweep(sweep Sweep) error {
	// The upper bound is exclusive: float64(MaxInt) can round up to 2^63.
	integerLimit := math.Ldexp(1, strconv.IntSize-1)
	inIntegerRange := func(value float64) bool {
		return value >= -integerLimit && value < integerLimit
	}
	points := 1
	for _, parameter := range sweep.Parameters {
		if configurationFieldKind("design.yaml", strings.Split(parameter.Path, ".")) == "int" {
			for _, value := range parameter.Values {
				if !inIntegerRange(value) {
					return fmt.Errorf("integer sweep coordinate overflows")
				}
			}
			if r := parameter.Range; r != nil && (!inIntegerRange(r.Start) || !inIntegerRange(r.Stop)) {
				return fmt.Errorf("integer sweep coordinate overflows")
			}
		}
		count := len(parameter.Values)
		if r := parameter.Range; r != nil {
			if count != 0 || r.Count < 1 || !finite(r.Start) || !finite(r.Stop) {
				return fmt.Errorf("invalid sweep range")
			}
			count = r.Count
			if count > 1 && !finite(r.Stop-r.Start) {
				return fmt.Errorf("sweep range arithmetic overflows")
			}
		}
		if count < 1 || count > sweep.MaxPoints || points > sweep.MaxPoints/count {
			return fmt.Errorf("sweep exceeds max_points %d", sweep.MaxPoints)
		}
		points *= count
	}
	return nil
}
