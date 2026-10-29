// Command module_runner exercises the same production memory bridge natively.
package main

import (
	"encoding/json"
	"fecim-lattice-tools/workbench/moduleapi"
	"fecim-lattice-tools/workbench/moduleapi/research"
	"fmt"
	"os"
	"time"
)

func main() {
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var request moduleapi.ResearchRequest
	if err = json.Unmarshal(data, &request); err != nil {
		panic(err)
	}
	bridge := research.New()
	defer bridge.Close()
	call := func(operation string) moduleapi.ResearchResult {
		request.Operation = operation
		data, _ := json.Marshal(request)
		var reply moduleapi.ResearchReply
		if err := json.Unmarshal(bridge.Dispatch(data), &reply); err != nil {
			panic(err)
		}
		if reply.Error != "" {
			panic(reply.Error)
		}
		request.Request++
		request.Revision = reply.Revision
		return reply.Result
	}
	call("example")
	call("start")
	for i := 0; i < 5000; i++ {
		result := call("state")
		if result.Job.State == "finished" {
			fmt.Print(result.ResultsJSON)
			return
		}
		if result.Job.State != "running" {
			panic(result.Job.Message)
		}
		time.Sleep(time.Millisecond)
	}
	panic("module runner timeout")
}
