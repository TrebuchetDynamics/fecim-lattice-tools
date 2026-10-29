package study

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// CreateStarterProjectRoot exclusively creates one new child directory beneath
// the caller's root. Existing directories (including empty ones) and links are
// refused. Only root methods are used; the borrowed root is never closed.
func CreateStarterProjectRoot(root *os.Root, name string, options StarterSweepOptions) error {
	if root == nil {
		return errors.New("project root is required")
	}
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, "/\\\x00:") {
		return errors.New("project name must be a single directory name")
	}
	contents, err := starterDocuments(options)
	if err != nil {
		return err
	}
	if err := root.Mkdir(name, 0755); err != nil {
		return fmt.Errorf("create project: %w", err)
	}
	child, err := root.OpenRoot(name)
	if err != nil {
		return errors.Join(err, root.Remove(name))
	}
	created := make([]string, 0, len(contents))
	rollback := func(cause error) error {
		for _, path := range created {
			cause = errors.Join(cause, child.Remove(path))
		}
		cause = errors.Join(cause, child.Close())
		return errors.Join(cause, root.Remove(name))
	}
	for _, item := range contents {
		file, err := child.OpenFile(item.name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			return rollback(err)
		}
		created = append(created, item.name)
		_, writeErr := file.WriteString(item.body)
		closeErr := file.Close()
		if writeErr != nil || closeErr != nil {
			return rollback(errors.Join(writeErr, closeErr))
		}
	}
	return child.Close()
}
