package main

import (
	"strings"
	"testing"
)

func runPkgSum(jsonl string) (pass, fail, skip, total int) {
	states := make(map[string]*pkgState)
	for _, line := range strings.Split(strings.TrimSpace(jsonl), "\n") {
		if line == "" {
			continue
		}
		var ev TestEvent
		if err := unmarshalEvent([]byte(line), &ev); err != nil {
			continue
		}
		if ev.Package == "" {
			continue
		}
		sw := ev.Action
		if sw != "pass" && sw != "fail" && sw != "skip" {
			continue
		}
		st := states[ev.Package]
		if st == nil {
			st = &pkgState{}
			states[ev.Package] = st
		}
		if sw == "fail" {
			st.seenFail = true
			continue
		}
		st.last = sw
	}
	for _, st := range states {
		total++
		if st.seenFail {
			fail++
		} else if st.last == "skip" {
			skip++
		} else {
			pass++
		}
	}
	return
}

func TestPkgSum_AllPass(t *testing.T) {
	jsonl := `{"Action":"pass","Package":"foo/bar"}` + "\n" +
		`{"Action":"pass","Package":"foo/baz"}`
	p, f, s, tot := runPkgSum(jsonl)
	if p != 2 || f != 0 || s != 0 || tot != 2 {
		t.Fatalf("want 2/0/0/2 got %d/%d/%d/%d", p, f, s, tot)
	}
}

func TestPkgSum_WithFail(t *testing.T) {
	jsonl := `{"Action":"pass","Package":"foo/bar"}` + "\n" +
		`{"Action":"fail","Package":"foo/baz"}` + "\n" +
		`{"Action":"pass","Package":"foo/baz"}`
	p, f, s, tot := runPkgSum(jsonl)
	if p != 1 || f != 1 || s != 0 || tot != 2 {
		t.Fatalf("want 1/1/0/2 got %d/%d/%d/%d", p, f, s, tot)
	}
}

func TestPkgSum_WithSkip(t *testing.T) {
	jsonl := `{"Action":"skip","Package":"foo/bar"}` + "\n" +
		`{"Action":"pass","Package":"foo/baz"}`
	p, f, s, tot := runPkgSum(jsonl)
	if p != 1 || f != 0 || s != 1 || tot != 2 {
		t.Fatalf("want 1/0/1/2 got %d/%d/%d/%d", p, f, s, tot)
	}
}
