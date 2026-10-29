//go:build !js

package comparison

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	model "fecim-lattice-tools/module5-comparison/pkg/comparison"
	"fecim-lattice-tools/workbench/moduleapi"
	nativeapi "fecim-lattice-tools/workbench/moduleapi/native"
	"fmt"
	"io"
	"os"
)

func localExportAvailable(ctx context.Context) bool {
	_, ok := nativeapi.Projects(ctx)
	return ok
}

// Export authority is leased only from an existing authenticated project. The
// browser cannot choose paths. Immutable content names refuse clobbers/symlinks.
func exportRooted(ctx context.Context, ref moduleapi.ResourceReference, result model.ScenarioEstimate, format string) (string, error) {
	resolver, ok := nativeapi.Projects(ctx)
	if !ok {
		return "", errors.New("local export requires a granted project resolver")
	}
	data, err := model.MarshalScenario(result, format)
	if err != nil {
		return "", err
	}
	name := fmt.Sprintf("exports/comparison/comparison-v2-%x.%s", sha256.Sum256(data), format)
	err = resolver.WithProject(ctx, ref, true, func(root *os.Root) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := root.MkdirAll("exports/comparison", 0700); err != nil {
			return err
		}
		f, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if errors.Is(err, os.ErrExist) {
			info, e := root.Lstat(name)
			if e != nil {
				return e
			}
			if !info.Mode().IsRegular() {
				return errors.New("immutable comparison artifact must be a regular file, not a symlink")
			}
			old, e := root.Open(name)
			if e != nil {
				return e
			}
			defer old.Close()
			prior, e := io.ReadAll(io.LimitReader(old, int64(len(data)+1)))
			if e != nil {
				return e
			}
			if !bytes.Equal(data, prior) {
				return errors.New("immutable comparison artifact already exists with different bytes")
			}
			return nil
		}
		if err != nil {
			return err
		}
		_, err = f.Write(data)
		closeErr := f.Close()
		if err != nil {
			root.Remove(name)
			return err
		}
		if closeErr != nil {
			root.Remove(name)
			return closeErr
		}
		return nil
	})
	return name, err
}
