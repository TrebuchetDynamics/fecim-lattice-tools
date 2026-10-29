//go:build !linux

package eda

import (
	"context"
	"errors"
	"os"
	"time"
)

func runProcess(context.Context, *os.Root, string, []string, *boundedLog, time.Duration, int64) error {
	return errors.New("bounded installed-tool runner unavailable on this platform; Linux pinned-process profile required")
}
