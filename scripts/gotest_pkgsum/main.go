package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// go test -json emits one JSON object per line (TestEvent).
// We aggregate package-level final status deterministically.
//
// Counting rule (per spec):
// - Every observed non-empty Package must have a pass/fail/skip with empty Test.
// - Final status per package is the last package-level pass/skip,
//   BUT any fail (including a test/subtest failure) dominates.

type TestEvent struct {
	Action  string `json:"Action"`
	Package string `json:"Package"`
	Test    string `json:"Test"`
}

type pkgState struct {
	completed bool
	seenFail  bool
	last      string // last package-level pass/skip
}

func main() {
	os.Exit(runPkgSum(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func runPkgSum(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	var r io.Reader = stdin
	if len(args) >= 1 && args[0] != "-" {
		f, err := os.Open(args[0])
		if err != nil {
			fmt.Fprintf(stderr, "open %s: %v\n", args[0], err)
			return 2
		}
		defer f.Close()
		r = f
	}

	states := make(map[string]*pkgState)

	s := bufio.NewScanner(r)
	// Some packages can emit very long Output lines; increase scanner buffer.
	s.Buffer(make([]byte, 0, 256*1024), 64*1024*1024)

	invalidLines := 0
	for s.Scan() {
		line := s.Bytes()
		var ev TestEvent
		if err := json.Unmarshal(line, &ev); err != nil {
			// Be tolerant of occasional non-JSON noise from test subprocess output.
			invalidLines++
			continue
		}
		if ev.Package == "" {
			continue
		}
		st := states[ev.Package]
		if st == nil {
			st = &pkgState{}
			states[ev.Package] = st
		}
		if ev.Action == "fail" {
			st.seenFail = true
		}
		if ev.Test != "" || (ev.Action != "pass" && ev.Action != "fail" && ev.Action != "skip") {
			continue
		}
		st.completed = true
		if ev.Action != "fail" {
			st.last = ev.Action
		}
	}
	if err := s.Err(); err != nil {
		fmt.Fprintf(stderr, "scan: %v\n", err)
		return 2
	}

	pass := 0
	fail := 0
	skip := 0
	total := 0

	var incomplete []string
	for pkg, st := range states {
		if !st.completed {
			incomplete = append(incomplete, pkg)
			continue
		}
		total++
		if st.seenFail {
			fail++
			continue
		}
		if st.last == "skip" {
			skip++
			continue
		}
		pass++
	}

	if invalidLines > 0 {
		fmt.Fprintf(stderr, "WARN: skipped %d non-JSON line(s) while aggregating go test -json output\n", invalidLines)
	}
	if len(incomplete) > 0 {
		sort.Strings(incomplete)
		fmt.Fprintf(stderr, "ERROR: incomplete go test -json stream: packages without terminal summaries: %s\n", strings.Join(incomplete, ", "))
		return 2
	}
	if total == 0 {
		fmt.Fprintf(stderr, "ERROR: no package summaries parsed from go test -json stream\n")
		return 2
	}

	fmt.Fprintf(stdout, "PKG_SUM pass=%d fail=%d skip=%d total=%d\n", pass, fail, skip, total)
	if fail > 0 {
		return 1
	}
	return 0
}
