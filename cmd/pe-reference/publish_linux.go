package main

import "golang.org/x/sys/unix"

// RENAME_NOREPLACE also protects empty directories and dangling symlinks created
// after validation. Never fall back to a check-then-Rename on unsupported filesystems.
func publishComparison(staging, output string) error {
	return unix.Renameat2(unix.AT_FDCWD, staging, unix.AT_FDCWD, output, unix.RENAME_NOREPLACE)
}
