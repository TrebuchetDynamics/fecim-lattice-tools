package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestDispatchResearchSubcommandUsesResearchRunner(t *testing.T) {
	var got []string
	previous := researchRunner
	researchRunner = func(args []string) error {
		got = append([]string(nil), args...)
		return nil
	}
	defer func() { researchRunner = previous }()

	if err := dispatchSubcommand([]string{"research", "search", "HZO coercive field"}); err != nil {
		t.Fatalf("dispatch research: %v", err)
	}
	want := []string{"search", "HZO coercive field"}
	if len(got) != len(want) {
		t.Fatalf("research runner args len=%d want=%d args=%v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("arg %d=%q want %q", i, got[i], want[i])
		}
	}
}

func TestDispatchResearchSubcommandPropagatesRunnerError(t *testing.T) {
	previous := researchRunner
	researchRunner = func(args []string) error {
		return errors.New("research tool failed")
	}
	defer func() { researchRunner = previous }()

	err := dispatchSubcommand([]string{"research", "ingest"})
	if err == nil || !strings.Contains(err.Error(), "research tool failed") {
		t.Fatalf("expected runner error, got %v", err)
	}
}

func TestRootUsageListsResearchSubcommand(t *testing.T) {
	var buf bytes.Buffer
	printRootUsage(&buf)
	text := buf.String()
	if !strings.Contains(text, "research") {
		t.Fatalf("root usage must mention research subcommand:\n%s", text)
	}
	if !strings.Contains(text, "research ingest") {
		t.Fatalf("root usage must include research example:\n%s", text)
	}
}
