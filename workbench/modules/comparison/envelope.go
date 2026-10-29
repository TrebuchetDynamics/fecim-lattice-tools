package comparison

import (
	"encoding/json"
	"errors"
	"strings"
)

// Bound ambiguous JSON before typed decoding; no arrays or dynamic input maps
// may allocate beyond the small comparison command budget.
func uniqueEnvelope(d *json.Decoder, depth int) error {
	if depth > 16 {
		return errors.New("comparison JSON depth exceeds budget")
	}
	token, err := d.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	if delim != '{' && delim != '[' {
		return errors.New("unexpected delimiter")
	}
	keys := map[string]bool{}
	count := 0
	for d.More() {
		count++
		if count > 32 {
			return errors.New("comparison JSON collection exceeds budget")
		}
		if delim == '{' {
			key, err := d.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok {
				return errors.New("invalid JSON key")
			}
			name = strings.ToLower(name)
			if keys[name] {
				return errors.New("duplicate JSON field")
			}
			keys[name] = true
		}
		if err := uniqueEnvelope(d, depth+1); err != nil {
			return err
		}
	}
	_, err = d.Token()
	return err
}
