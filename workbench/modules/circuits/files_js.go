//go:build js && wasm

package circuits

import (
	"context"
	"errors"
	"fecim-lattice-tools/workbench/moduleapi"
)

func (p *Port) save(context.Context, moduleapi.Command) error {
	return errors.New("file exports require authenticated fecim-lattice-tools serve and an opened rooted project")
}
