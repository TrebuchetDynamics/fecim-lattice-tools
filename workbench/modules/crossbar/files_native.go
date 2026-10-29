//go:build !js

package crossbar

import (
	"context"
	"fecim-lattice-tools/workbench/moduleapi"
	"fecim-lattice-tools/workbench/moduleapi/native"
	"fecim-lattice-tools/workbench/project"
	"fmt"
	"io"
	"os"
	"strings"
)

func (p *Port) readInput(ctx context.Context, c moduleapi.Command) ([]byte, error) {
	if c.Action == "import" {
		if document, ok := c.Input["document"]; ok {
			return []byte(document), nil
		}
		if len(c.Payload) > 0 {
			return c.Payload, nil
		}
		resolver, ok := native.Resources(ctx)
		if !ok || len(c.Resources) != 1 {
			return nil, fmt.Errorf("bind one crossbar input resource")
		}
		reader, _, err := resolver.Open(ctx, c.Resources[0])
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		data, err := io.ReadAll(io.LimitReader(reader, maxDocumentBytes+1))
		return data, err
	}
	resolver, ok := native.Projects(ctx)
	if !ok || len(c.Resources) != 1 {
		return nil, fmt.Errorf("open a rooted local project first")
	}
	var data []byte
	err := resolver.WithProject(ctx, c.Resources[0], false, func(root *os.Root) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		var err error
		data, err = project.ReadRegularInput(ctx, root, c.Input["path"], maxDocumentBytes)
		return err
	})
	return data, err
}
func (p *Port) save(ctx context.Context, c moduleapi.Command, data []byte) error {
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
		temporary := path + ".crossbar-tmp"
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
