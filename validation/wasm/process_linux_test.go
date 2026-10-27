//go:build linux

package wasm_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Playwright starts detached browser process groups. Killing only Node or its
// process group is insufficient; this regression deliberately detaches a child.
func TestParityCommandKillsDetachedDescendants(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestParityProcessHelper$")
	cmd.Env = append(os.Environ(), "FECIM_PARITY_PROCESS_HELPER=1")
	if err := protectProcessTree(cmd); err != nil {
		t.Fatal(err)
	}
	output, err := cmd.CombinedOutput()
	if err == nil || ctx.Err() == nil {
		t.Fatalf("expected command deadline, got %v", err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		t.Fatalf("missing detached child pid: %s", output)
	}
	t.Cleanup(func() { _ = syscall.Kill(pid, syscall.SIGKILL) })
	deadline := time.Now().Add(time.Second)
	for {
		data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
		if os.IsNotExist(err) {
			return
		}
		fields := strings.Fields(string(data[strings.LastIndexByte(string(data), ')')+1:]))
		if len(fields) > 0 && (fields[0] == "Z" || fields[0] == "X") {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("detached descendant %d survived command timeout", pid)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestParityProcessHelper(t *testing.T) {
	if os.Getenv("FECIM_PARITY_PROCESS_HELPER") != "1" {
		return
	}
	child := exec.Command("sleep", "60")
	child.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}
	fmt.Println(child.Process.Pid)
	time.Sleep(time.Minute)
	os.Exit(0)
}
