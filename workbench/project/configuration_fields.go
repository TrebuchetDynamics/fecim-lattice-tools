package project

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// ConfigurationField exposes one existing allowlisted scalar, not YAML source
// or filesystem authority. Document is a fixed YAML name and Path is its schema
// field path. Kind is string/int/int64/float; Value is literal editor input.
type ConfigurationField struct {
	Document string `json:"document"`
	Path     string `json:"path"`
	Kind     string `json:"kind"`
	Value    string `json:"value"`
}

// ReadConfigurationBytes captures the same immutable three-document snapshot as
// ReadConfiguration, without disk. Other map entries (e.g. input bytes) are not
// part of the configuration revision. Like ReadConfiguration, this is not proof
// of scientific validity; use LoadBytes for whole-bundle validation.
func ReadConfigurationBytes(documents map[string][]byte) (ConfigurationSnapshot, error) {
	var original [3]string
	for i, name := range configurationDocuments {
		data, ok := documents[name]
		if !ok {
			return ConfigurationSnapshot{}, fmt.Errorf("missing configuration document %s", name)
		}
		if int64(len(data)) > maxProjectFileBytes {
			return ConfigurationSnapshot{}, fmt.Errorf("configuration document exceeds %d bytes", maxProjectFileBytes)
		}
		original[i] = string(data)
	}
	return configurationSnapshot(original), nil
}

// Fields derives fresh scalar values in document order and source field order.
// Aliases/merges/ambiguous keys are refused, just as for configuration edits.
// Numeric input is canonical Go text (hexadecimal YAML integers become decimal)
// so clients never need to parse YAML or infer schema types.
func (s ConfigurationSnapshot) Fields() ([]ConfigurationField, error) {
	if s.revision == "" {
		return nil, fmt.Errorf("configuration snapshot is required")
	}
	fields := make([]ConfigurationField, 0)
	for i, document := range configurationDocuments {
		var tree yaml.Node
		if err := decodeStrict(strings.NewReader(s.documents[i]), &tree); err != nil {
			return nil, fmt.Errorf("invalid configuration document %s", document)
		}
		if err := editableConfigurationTree(&tree); err != nil {
			return nil, fmt.Errorf("ambiguous configuration document %s", document)
		}
		if tree.Kind != yaml.DocumentNode || len(tree.Content) != 1 {
			return nil, fmt.Errorf("invalid configuration document %s", document)
		}
		var walk func(*yaml.Node, []string) error
		walk = func(node *yaml.Node, path []string) error {
			if node.Kind == yaml.ScalarNode {
				kind := configurationFieldKind(document, path)
				if kind == "" {
					return nil
				}
				var value string
				switch kind {
				case "string":
					if err := node.Decode(&value); err != nil {
						return fmt.Errorf("invalid configuration scalar")
					}
				case "int", "int64":
					var number int64
					if err := node.Decode(&number); err != nil {
						return fmt.Errorf("invalid configuration integer")
					}
					if kind == "int" && int64(int(number)) != number {
						return fmt.Errorf("configuration integer overflows")
					}
					value = strconv.FormatInt(number, 10)
				case "float":
					var number float64
					if err := node.Decode(&number); err != nil || !finite(number) {
						return fmt.Errorf("invalid configuration number")
					}
					value = strconv.FormatFloat(number, 'g', -1, 64)
				}
				escaped := make([]string, len(path))
				for i, part := range path {
					escaped[i] = strings.ReplaceAll(strings.ReplaceAll(part, "~", "~0"), "/", "~1")
				}
				fields = append(fields, ConfigurationField{Document: document, Path: "/" + strings.Join(escaped, "/"), Kind: kind, Value: value})
				return nil
			}
			switch node.Kind {
			case yaml.MappingNode:
				for i := 0; i < len(node.Content); i += 2 {
					if err := walk(node.Content[i+1], append(path, node.Content[i].Value)); err != nil {
						return err
					}
				}
			case yaml.SequenceNode:
				for i, child := range node.Content {
					if err := walk(child, append(path, strconv.Itoa(i))); err != nil {
						return err
					}
				}
			}
			return nil
		}
		if err := walk(tree.Content[0], nil); err != nil {
			return nil, err
		}
	}
	return fields, nil
}
