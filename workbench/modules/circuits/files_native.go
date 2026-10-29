//go:build !js

package circuits

import (
	"context"
	"encoding/json"
	"fecim-lattice-tools/workbench/moduleapi"
	"fecim-lattice-tools/workbench/moduleapi/native"
	"fmt"
	"os"
	"strings"
)

func (p *Port) save(ctx context.Context, c moduleapi.Command) error {
	if p.limit != 128 {
		return fmt.Errorf("exports require local profile")
	}
	data, err := json.Marshal(struct {
		Schema string
		Config Config
		State  State
		Result *Result
		Logs   []string
	}{"fecim.circuits.export.v1", p.cfg, p.state, p.result, p.logs})
	if err != nil {
		return err
	}
	return p.saveBytes(ctx, c, data)
}
func (p *Port) saveBytes(ctx context.Context, c moduleapi.Command, data []byte) error {
	resolver, ok := native.Projects(ctx)
	if !ok || len(c.Resources) != 1 {
		return fmt.Errorf("open a rooted local project first")
	}
	path := c.Input["path"]
	if path == "" || !strings.HasSuffix(path, ".json") {
		return fmt.Errorf("choose an explicit JSON path inside the project")
	}
	return resolver.WithProject(ctx, c.Resources[0], true, func(root *os.Root) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		temporary := path + ".circuits-tmp"
		f, err := root.OpenFile(temporary, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return err
		}
		defer root.Remove(temporary)
		_, err = f.Write(data)
		closeErr := f.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
		if err = ctx.Err(); err != nil {
			return err
		}
		return root.Rename(temporary, path)
	})
}
