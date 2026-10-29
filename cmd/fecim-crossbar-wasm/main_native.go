//go:build !js

package main

import (
	"encoding/json"
	"fecim-lattice-tools/workbench/modules/crossbar"
	"fmt"
	"os"
	"time"
)

// Native oracle for the identical production protocol used by the Worker.
func main() {
	if len(os.Args) != 2 {
		panic("provide one JSON request fixture")
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var requests []crossbar.Request
	if err = json.Unmarshal(data, &requests); err != nil {
		panic(err)
	}
	results := []string{}
	for _, request := range requests {
		b := crossbar.NewBridge()
		call := func(op string) crossbar.Reply {
			request.Operation = op
			data, _ := json.Marshal(request)
			var reply crossbar.Reply
			if err := json.Unmarshal(b.Dispatch(data), &reply); err != nil {
				panic(err)
			}
			if reply.Error != "" {
				panic(reply.Error)
			}
			request.Request++
			request.Revision = reply.Revision
			return reply
		}
		call("action")
		call("start")
		deadline := time.Now().Add(30 * time.Second)
		for {
			reply := call("state")
			if reply.Result.Job.State == "succeeded" {
				results = append(results, reply.Result.ResultsJSON)
				break
			}
			if reply.Result.Job.State == "failed" || time.Now().After(deadline) {
				panic(fmt.Sprintf("native science failed: %+v", reply.Result.Job))
			}
			time.Sleep(time.Millisecond)
		}
		b.Close()
	}
	data, err = json.Marshal(results)
	if err != nil {
		panic(err)
	}
	fmt.Print(string(data))
}
