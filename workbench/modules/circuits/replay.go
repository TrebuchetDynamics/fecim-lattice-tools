package circuits

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"fecim-lattice-tools/workbench/hysteresis"
	"fecim-lattice-tools/workbench/moduleapi"
)

// Scenario is a command fixture for the original production protocol. Each
// scientific operation uses the cancellable job seam in every runtime.
type Scenario struct {
	Identity hysteresis.Identity `json:"identity"`
	Session  string              `json:"session"`
	Commands []moduleapi.Command `json:"commands"`
}

func Replay(ctx context.Context, scenario Scenario) ([]string, error) {
	b := NewBridge()
	defer b.Close()
	req := Request{Identity: scenario.Identity, Session: scenario.Session, Request: 1}
	outputs := []string{}
	call := func(op string, c moduleapi.Command) (Reply, error) {
		if err := ctx.Err(); err != nil {
			return Reply{}, err
		}
		req.Operation = op
		req.Input.Command = c
		data, _ := json.Marshal(req)
		var reply Reply
		if err := json.Unmarshal(b.Dispatch(data), &reply); err != nil {
			return Reply{}, err
		}
		if reply.Error != "" {
			return Reply{}, errors.New(reply.Error)
		}
		req.Request++
		req.Revision = reply.Revision
		return reply, nil
	}
	for _, command := range scenario.Commands {
		op := "action"
		science := command.Action == "write" || command.Action == "read" || command.Action == "compute"
		if science {
			op = "start"
		}
		reply, err := call(op, command)
		if err != nil {
			return nil, err
		}
		if science {
			for reply.Result.Job.State == "running" || reply.Result.Job.State == "cancelling" {
				time.Sleep(time.Millisecond)
				reply, err = call("state", moduleapi.Command{})
				if err != nil {
					return nil, err
				}
			}
			if reply.Result.Job.State != "completed" {
				return nil, errors.New(reply.Result.Job.Message)
			}
			outputs = append(outputs, reply.Result.ResultsJSON)
		}
	}
	return outputs, nil
}
