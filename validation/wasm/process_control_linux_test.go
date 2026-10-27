//go:build linux

package wasm_test

import (
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// The gate's supported environment is Linux. Playwright detaches Firefox into
// another process group, so own descendants as well as the Node process group.
func protectProcessTree(cmd *exec.Cmd) error {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.WaitDelay = 2 * time.Second
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}
		if err := cmd.Process.Signal(syscall.Signal(0)); err != nil {
			return err
		}
		pid := cmd.Process.Pid
		// Freeze owned groups/descendants before taking another snapshot, so a
		// detached browser cannot keep forking while its parent is killed.
		_ = syscall.Kill(-pid, syscall.SIGSTOP)
		defer syscall.Kill(-pid, syscall.SIGKILL)
		var owned []int
		for pass := 0; pass < 3; pass++ {
			var err error
			owned, err = descendantPIDs(pid)
			if err != nil {
				return err
			}
			for _, child := range owned {
				_ = syscall.Kill(child, syscall.SIGSTOP)
			}
		}
		for i := len(owned) - 1; i >= 0; i-- {
			_ = syscall.Kill(owned[i], syscall.SIGKILL)
		}
		if err := syscall.Kill(-pid, syscall.SIGKILL); errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		} else {
			return err
		}
	}
	return nil
}

func descendantPIDs(root int) ([]int, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}
	children := map[int][]int{}
	for _, entry := range entries {
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}
		data, err := os.ReadFile("/proc/" + entry.Name() + "/stat")
		if err != nil {
			continue
		} // Process exited between enumeration and read.
		end := strings.LastIndexByte(string(data), ')')
		if end < 0 {
			continue
		}
		fields := strings.Fields(string(data[end+1:]))
		if len(fields) < 2 {
			continue
		}
		parent, err := strconv.Atoi(fields[1])
		if err == nil {
			children[parent] = append(children[parent], pid)
		}
	}
	queue := []int{root}
	for i := 0; i < len(queue); i++ {
		queue = append(queue, children[queue[i]]...)
	}
	return queue[1:], nil
}
