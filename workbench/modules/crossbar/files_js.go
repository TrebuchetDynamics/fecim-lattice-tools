//go:build js && wasm

package crossbar

import (
	"context"
	"fecim-lattice-tools/workbench/moduleapi"
	"fmt"
)

func (p *Port) readInput(_ context.Context, c moduleapi.Command) ([]byte, error) {
	if c.Action != "import" || len(c.Resources) != 0 {
		return nil, fmt.Errorf("rooted files require fecim-lattice-tools serve")
	}
	if document, ok := c.Input["document"]; ok {
		return []byte(document), nil
	}
	return c.Payload, nil
}
func (p *Port) save(context.Context, moduleapi.Command, []byte) error {
	return fmt.Errorf("file writes require fecim-lattice-tools serve")
}
