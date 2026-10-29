//go:build linux

package eda

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

// A monitored staging ceiling, NOT an OS sandbox or hard adversarial disk quota.
// The directory descriptor remains pinned through exec; root.Name is never reopened.
func runProcess(ctx context.Context, stage *os.Root, exe string, args []string, log *boundedLog, limit time.Duration, stagingLimit int64) error {
	ctx, cancel := context.WithTimeout(ctx, limit)
	defer cancel()
	dir, err := stage.Open(".")
	if err != nil {
		return err
	}
	defer dir.Close()
	if err = stageSize(stage, stagingLimit); err != nil {
		return err
	}
	cmd := exec.Command(exe, args...)
	// Go's child chdir happens before ExtraFiles remapping. Use the inherited
	// original descriptor for chdir, while fd 3 pins it across the final exec.
	cmd.Dir = "/proc/self/fd/" + strconv.FormatUint(uint64(dir.Fd()), 10)
	cmd.ExtraFiles = []*os.File{dir}
	cmd.Env = []string{"PATH=/usr/bin:/bin", "HOME=.", "TMPDIR=.", "LC_ALL=C"}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdout = log
	cmd.Stderr = log
	cmd.WaitDelay = time.Second
	if err = cmd.Start(); err != nil {
		return err
	}
	kill := func() { _ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	defer kill()
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	tick := time.NewTicker(20 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			kill()
			<-done
			return ctx.Err()
		case err := <-done:
			kill()
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if e := stageSize(stage, stagingLimit); e != nil {
				return e
			}
			return err
		case <-tick.C:
			if e := stageSize(stage, stagingLimit); e != nil {
				kill()
				<-done
				return e
			}
		}
	}
}
func stageSize(root *os.Root, limit int64) error {
	var size int64
	var count int
	return fs.WalkDir(root.FS(), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		count++
		if count > 4096 {
			return errors.New("staging entry limit exceeded")
		}
		if d.Type()&os.ModeSymlink != 0 {
			return errors.New("tool staging symlink rejected")
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return errors.New("nonregular tool staging entry")
		}
		size += info.Size()
		if size > limit {
			return errors.New("staging exceeded monitored ceiling; aborted and cleaned")
		}
		return nil
	})
}
