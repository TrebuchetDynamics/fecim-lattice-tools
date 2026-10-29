package project

import (
	"errors"
	"fmt"
	"io"
	"os"
)

// Native project inputs have an 8 MiB per-file and aggregate byte budget.
// Larger model/dataset reads belong to opaque native resource capabilities,
// not project configuration/hash loading. YAML remains limited to 1 MiB.
const maxInputFileBytes int64 = 8 << 20
const maxInputTotalBytes int64 = 8 << 20

// Preserve rooted traversal/errno identity without returning supplied paths.
func safeReadError(err error) error {
	var p *os.PathError
	if errors.As(err, &p) {
		return &os.PathError{Op: "read", Path: "granted-file", Err: p.Err}
	}
	return errors.New("granted file unavailable")
}
func readRegularRoot(root *os.Root, name string, limit int64, followLinks bool) ([]byte, error) {
	if root == nil {
		return nil, errors.New("project root is required")
	}
	before, err := root.Lstat(name)
	if err != nil {
		return nil, safeReadError(err)
	}
	if followLinks && before.Mode()&os.ModeSymlink != 0 {
		before, err = root.Stat(name)
		if err != nil {
			return nil, safeReadError(err)
		}
	}
	if !before.Mode().IsRegular() {
		return nil, errors.New("file must be regular")
	}
	if before.Size() > limit {
		return nil, fmt.Errorf("file exceeds %d byte read budget", limit)
	}
	file, err := openRegularRead(root, name)
	if err != nil {
		return nil, safeReadError(err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, safeReadError(err)
	}
	if !info.Mode().IsRegular() || !os.SameFile(before, info) {
		return nil, errors.New("regular file changed during admission")
	}
	if info.Size() > limit {
		return nil, fmt.Errorf("file exceeds %d byte read budget", limit)
	}
	data, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		return nil, safeReadError(err)
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("file exceeds %d byte read budget", limit)
	}
	after, err := root.Lstat(name)
	if err != nil {
		return nil, safeReadError(err)
	}
	if followLinks && after.Mode()&os.ModeSymlink != 0 {
		after, err = root.Stat(name)
		if err != nil {
			return nil, safeReadError(err)
		}
	}
	if !after.Mode().IsRegular() || !os.SameFile(info, after) {
		return nil, errors.New("regular file changed during read")
	}
	return data, nil
}
