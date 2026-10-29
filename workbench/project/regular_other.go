//go:build !unix

package project

import "os"

func openRegularRead(root *os.Root, name string) (*os.File, error) { return root.Open(name) }
