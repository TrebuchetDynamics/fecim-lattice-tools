package research

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"fecim-lattice-tools/workbench/moduleapi"
	"fecim-lattice-tools/workbench/project"
	"fecim-lattice-tools/workbench/study"
)

func TestProductionResearchBridge(t *testing.T) {
	b := New()
	defer b.Close()
	var seq, rev uint64
	call := func(action string, input moduleapi.ResearchInput) moduleapi.ResearchReply {
		t.Helper()
		seq++
		data, _ := json.Marshal(moduleapi.ResearchRequest{Identity: Identity(), Session: "test", Request: seq, Revision: rev, Operation: action, Input: input})
		var reply moduleapi.ResearchReply
		if err := json.Unmarshal(b.Dispatch(data), &reply); err != nil {
			t.Fatal(err)
		}
		if reply.Error != "" {
			t.Fatal(reply.Error)
		}
		rev = reply.Revision
		return reply
	}
	opened := call("example", moduleapi.ResearchInput{})
	if opened.Result.Configuration == nil || len(opened.Result.Configuration.Fields) == 0 {
		t.Fatal("no Go fields")
	}
	call("start", moduleapi.ResearchInput{RepositoryRevision: "parity"})
	for i := 0; i < 1000; i++ {
		reply := call("state", moduleapi.ResearchInput{})
		if reply.Result.Job.State != "running" && reply.Result.Job.State != "cancelling" {
			if reply.Result.Job.State != "finished" {
				t.Fatalf("%+v", reply)
			}
			docs := ExampleDocuments()
			bundle, err := project.LoadBytesLimited(docs, 256, 32)
			if err != nil {
				t.Fatal(err)
			}
			want, err := study.RunMemory(context.Background(), bundle, study.RunOptions{Workers: 1, RepositoryRevision: "parity"})
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal([]byte(reply.Result.ResultsJSON), want.ResultsJSON) {
				t.Fatal("original serializer mismatch")
			}
			if reply.Result.Job.InputRevision != opened.Result.Configuration.Revision || !reply.Result.Job.AnalysisAvailable || reply.Result.Job.ResultsRevision != reply.Result.Page.Revision || !reflect.DeepEqual(reply.Result.Job.Counts, reply.Result.Counts) {
				t.Fatalf("terminal accepted analysis not retained: %+v", reply.Result.Job)
			}
			if reply.Result.Plot == nil || reply.Result.Plot.Revision != reply.Result.Page.Revision || reply.Result.Plot.Plotted != 8 {
				t.Fatalf("supplementary plot missing or mismatched: %+v", reply.Result.Plot)
			}
			if reply.Result.Page.Total != 8 {
				t.Fatalf("page %+v", reply.Result.Page)
			}
			break
		}
		if i == 999 {
			t.Fatal("no terminal result")
		}
		time.Sleep(time.Millisecond)
	}
	call("start", moduleapi.ResearchInput{})
	requested := call("cancel", moduleapi.ResearchInput{})
	if requested.Result.Job.State != "cancelling" && requested.Result.Job.State != "cancelled" {
		t.Fatalf("%+v", requested)
	}
	for i := 0; i < 1000; i++ {
		reply := call("state", moduleapi.ResearchInput{})
		if reply.Result.Job.State == "cancelled" {
			if reply.Result.Job.Progress.Cancelled == 0 {
				t.Fatal("cancel did no work")
			}
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("cancellation never settled")
}

func TestResearchAdmissionAndRevision(t *testing.T) {
	b := New()
	defer b.Close()
	docs := ExampleDocuments()
	docs["sweep.yaml"] = bytes.ReplaceAll(docs["sweep.yaml"], []byte("[16, 32]"), []byte("[32, 64]"))
	data, _ := json.Marshal(moduleapi.ResearchRequest{Identity: Identity(), Session: "test", Request: 1, Operation: "import", Input: moduleapi.ResearchInput{Documents: docs}})
	var reply moduleapi.ResearchReply
	json.Unmarshal(b.Dispatch(data), &reply)
	if reply.Error == "" || reply.Committed {
		t.Fatal("oversized input admitted")
	}
	data, _ = json.Marshal(moduleapi.ResearchRequest{Identity: Identity(), Session: "test", Request: 2, Revision: 99, Operation: "example"})
	json.Unmarshal(b.Dispatch(data), &reply)
	if reply.Error == "" {
		t.Fatal("stale request accepted")
	}
}
