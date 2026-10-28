package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestServeCommand(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want string
		ok   bool
	}{
		{[]string{"serve", "--help"}, "--root", true},
		{[]string{"serve"}, "--root", false},
		{[]string{"serve", "--root", t.TempDir()}, "browser-assets-unavailable", false},
		{[]string{"serve", "--root", t.TempDir(), "--port", "-1"}, "port", false},
	} {
		var out, errs bytes.Buffer
		err := dispatchSubcommandWithWriters(tc.args, &out, &errs)
		if (err == nil) != tc.ok {
			t.Fatalf("%v: %v", tc.args, err)
		}
		text := out.String() + errs.String()
		if err != nil {
			text += err.Error()
		}
		if !strings.Contains(text, tc.want) {
			t.Fatalf("%v: %q lacks %q", tc.args, text, tc.want)
		}
	}
	var out bytes.Buffer
	printRootUsage(&out)
	if !strings.Contains(out.String(), "serve") {
		t.Fatal("serve absent from help")
	}
}
