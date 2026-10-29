package docs

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

// The input cap must also imply linear work for many short code lines, not
// repeated whole-prefix copies. This is an allocation guard, not a latency SLA.
func TestDocsWebParserWork(t *testing.T) {
	for _, size := range []int{32 << 10, MaxDocumentBytes - 64} {
		for _, fenced := range []bool{true, false} {
			line := "x\n"
			if !fenced {
				line = "    x\n"
			}
			body := strings.Repeat(line, size/len(line))
			if fenced {
				body = "```text\n" + body + "```\n"
			}
			source := Source{ID: "code", Path: "docs/guides/code.md", Text: body, SHA256: Hash([]byte(body)), Rights: "MIT", Attribution: "fixture"}
			runtime.GC()
			var before, after runtime.MemStats
			runtime.ReadMemStats(&before)
			corpus, err := Build([]Source{source})
			runtime.ReadMemStats(&after)
			if err != nil {
				t.Fatal(err)
			}
			if len(corpus.Documents[0].Nodes) != 1 || corpus.Documents[0].Nodes[0].Kind != "pre" {
				t.Fatal("code block lost")
			}
			if used := after.TotalAlloc - before.TotalAlloc; used > uint64(len(body))*192+8<<20 {
				t.Fatal(fmt.Sprintf("nonlinear parser work: %d bytes allocated for %d input bytes (fenced=%v)", used, len(body), fenced))
			}
		}
	}
}
