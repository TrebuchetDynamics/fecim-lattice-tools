package project

import (
	"context"
	"errors"
	"os"
)

// ReadRegularInput reuses native project admission for a bounded module input.
// The caller must hold its existing rooted capability lease. No links, special
// files, descriptor reopening or arbitrary root acquisition are allowed.
func ReadRegularInput(ctx context.Context, root *os.Root, name string, limit int64) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if limit < 1 || limit > maxInputFileBytes {
		return nil, errors.New("input byte budget is outside native admission")
	}
	data, err := readRegularRoot(root, name, limit, false)
	if err != nil {
		return nil, err
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	return data, nil
}
