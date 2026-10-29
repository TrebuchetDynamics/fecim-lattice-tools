package comparison

import (
	"encoding/json"
	"fecim-lattice-tools/workbench/hysteresis"
	"strings"
	"testing"
)

func TestComparisonWebAdmission(t *testing.T) {
	old := hysteresis.BuildID
	defer func() { hysteresis.BuildID = old }()
	hysteresis.BuildID = "comparison-admission"
	q := Request{Identity: Identity(), Session: "test", Request: 1, Operation: "snapshot"}
	valid, _ := json.Marshal(q)
	for name, data := range map[string][]byte{
		"duplicate":        []byte(strings.Replace(string(valid), `"revision":0`, `"revision":0,"revision":0`, 1)),
		"missing revision": []byte(strings.Replace(string(valid), `"revision":0,`, "", 1)),
		"oversized":        []byte(strings.Repeat(" ", (16<<10)+1)),
	} {
		t.Run(name, func(t *testing.T) {
			b := NewBridge()
			defer b.Close()
			var r Reply
			json.Unmarshal(b.Dispatch(data), &r)
			if r.Error == "" || r.Committed {
				t.Fatalf("admitted %s", name)
			}
		})
	}
	hysteresis.BuildID = "unqualified"
	q.Identity = Identity()
	unqualified, _ := json.Marshal(q)
	b := NewBridge()
	defer b.Close()
	var r Reply
	json.Unmarshal(b.Dispatch(unqualified), &r)
	if r.Error == "" {
		t.Fatal("unqualified Worker admitted")
	}
}
