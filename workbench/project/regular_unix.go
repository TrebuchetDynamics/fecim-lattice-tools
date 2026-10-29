//go:build unix

package project

import (
	"os"
	"syscall"
)

// Nonblocking open closes the regular-check/open FIFO substitution window.
// O_NONBLOCK does not alter reads of regular files; type is checked again.
func openRegularRead(root *os.Root, name string) (*os.File, error) {
	return root.OpenFile(name, os.O_RDONLY|syscall.O_NONBLOCK, 0)
}
