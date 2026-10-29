package circuits

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"fecim-lattice-tools/workbench/moduleapi"
)

func TestCircuitsBridgeTransaction(t *testing.T) {
	b := NewBridge()
	defer b.Close()
	req := Request{Identity: Identity(), Session: "test", Operation: "action", Request: 1, Input: BridgeInput{Command: moduleapi.Command{Action: "configure", Input: map[string]string{"rows": "1", "cols": "1", "architecture": "1T1R", "active_rows": "[true]", "input": "[0]"}}}}
	call := func() {
		t.Helper()
		data, _ := json.Marshal(req)
		var reply Reply
		if err := json.Unmarshal(b.Dispatch(data), &reply); err != nil {
			t.Fatalf("bridge did not serialize valid zero-signal science: %v", err)
		}
		if !reply.Committed || reply.Error != "" {
			t.Fatalf("bridge rejected action: %+v", reply)
		}
		req.Request++
		req.Revision = reply.Revision
	}
	call()
	req.Input.Command = moduleapi.Command{Action: "compute"}
	req.Operation = "start"
	call()
	req.Operation = "state"
	for i := 0; i < 100; i++ {
		time.Sleep(time.Millisecond)
		call()
	}
	if b.port.Canonical() == "" {
		t.Fatal("zero-signal job produced no scientific result")
	}
	s, _ := b.port.Snapshot(context.Background())
	for _, table := range s.Tables {
		if table.ID == "sense" && table.Rows[0][5].Value != nil {
			t.Fatal("zero input current must not report a positive SNR from TIA output offset")
		}
	}
}
