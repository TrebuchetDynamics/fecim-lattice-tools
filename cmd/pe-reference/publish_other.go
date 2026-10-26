//go:build !linux

package main

import "fmt"

// Fail closed where an atomic no-replace directory rename is not implemented.
// In particular, os.Rename can overwrite a caller-owned empty directory on Unix.
func publishComparison(staging, output string) error {
	return fmt.Errorf("atomic no-replace comparison publication requires Linux renameat2; output was not published")
}
