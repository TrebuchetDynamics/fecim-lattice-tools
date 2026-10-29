package server

import (
	"context"
	"encoding/json"
	"errors"
	"sync"

	"fecim-lattice-tools/shared/viewmodel"
	vmc "fecim-lattice-tools/shared/viewmodel/circuits"
	vmp "fecim-lattice-tools/shared/viewmodel/comparison"
	vmx "fecim-lattice-tools/shared/viewmodel/crossbar"
	vmd "fecim-lattice-tools/shared/viewmodel/docs"
	vme "fecim-lattice-tools/shared/viewmodel/eda"
	vmh "fecim-lattice-tools/shared/viewmodel/hysteresis"
	vmm "fecim-lattice-tools/shared/viewmodel/mnist"
	"fecim-lattice-tools/workbench/hysteresis"
	"fecim-lattice-tools/workbench/moduleapi"
)

type moduleFactory struct {
	id, title string
	construct func() any
}

// Sole native constructor registry. No Start, cwd/home reads, persistence or
// external-tool acquisition. Future ports self-supply via Port/Provider.
var nativeFactories = []moduleFactory{
	{"hysteresis", "Hysteresis", func() any { return vmh.New() }},
	{"crossbar", "Crossbar", func() any { return vmx.New(8, 8) }},
	{"mnist", "MNIST", func() any { return vmm.New() }},
	{"circuits", "Circuits", func() any { return vmc.New() }},
	{"comparison", "Comparison", func() any { return vmp.New() }},
	{"eda", "EDA", func() any { return vme.New() }},
	{"docs", "Documentation", func() any { return vmd.New() }},
}
var _ viewmodel.ModulePort = (*vmx.Module)(nil)

func nativePort(f moduleFactory) moduleapi.Port {
	if f.id == "hysteresis" {
		return &hysteresisPort{bridge: hysteresis.NewBridge(), id: randomSecret()}
	}
	candidate := f.construct()
	if port, ok := candidate.(moduleapi.Port); ok {
		return port
	}
	if provider, ok := candidate.(moduleapi.Provider); ok {
		return provider.WebPort()
	}
	return nil
}
func moduleDescriptor(f moduleFactory, port moduleapi.Port) moduleapi.ModuleDescriptor {
	if port != nil {
		return port.Descriptor()
	}
	reason := "This module has no implemented web port; legacy display metrics are not scientific transport."
	return moduleapi.ModuleDescriptor{ID: f.id, Title: f.title, Reason: reason, SchemaVersion: moduleapi.SchemaVersion, ModelVersion: "unavailable", Features: []string{}, Capabilities: []moduleapi.Capability{{Profile: "local", Reason: reason}, {Profile: "hosted", Reason: reason}}}
}

// Probe once per server, not on every list request. No legacy snapshot is read.
func (s *Server) initModules() {
	s.moduleDescriptors = []moduleapi.ModuleDescriptor{}
	for _, f := range nativeFactories {
		port := nativePort(f)
		s.moduleDescriptors = append(s.moduleDescriptors, moduleDescriptor(f, port))
		if port != nil {
			port.Close()
		}
	}
}

type moduleSession struct {
	mu         sync.Mutex
	id, module string
	port       moduleapi.Port
	ctx        context.Context
	cancel     context.CancelFunc
}

type hysteresisPort struct {
	bridge            *hysteresis.Bridge
	id                string
	request, revision uint64
	last              moduleapi.Snapshot
	initialized       bool
}

func (p *hysteresisPort) Descriptor() moduleapi.ModuleDescriptor {
	features := []string{}
	for _, op := range hysteresis.Operations() {
		features = append(features, op.Name)
	}
	return moduleapi.ModuleDescriptor{ID: "hysteresis", Title: "Hysteresis", Available: true, SchemaVersion: moduleapi.SchemaVersion, ModelVersion: hysteresis.BridgeIdentity().Model, Features: features, Capabilities: []moduleapi.Capability{{Profile: "local", Available: true}, {Profile: "hosted", Available: true}}}
}
func (p *hysteresisPort) Snapshot(ctx context.Context) (moduleapi.Snapshot, error) {
	if !p.initialized {
		payload, _ := json.Marshal(hysteresis.DefaultConfig())
		return p.Apply(ctx, moduleapi.Command{Action: "session.new", Payload: payload})
	}
	return p.last, nil
}
func (p *hysteresisPort) Apply(ctx context.Context, c moduleapi.Command) (moduleapi.Snapshot, error) {
	if err := ctx.Err(); err != nil {
		return moduleapi.Snapshot{}, err
	}
	if c.Revision != p.revision {
		return moduleapi.Snapshot{}, errors.New("stale module revision")
	}
	p.request++
	payload := c.Payload
	if len(payload) == 0 {
		payload = json.RawMessage("null")
	}
	data, err := json.Marshal(hysteresis.Envelope{Identity: hysteresis.BridgeIdentity(), Session: p.id, Request: p.request, Revision: p.revision, Operation: c.Action, Input: payload})
	if err != nil {
		return moduleapi.Snapshot{}, err
	}
	var reply hysteresis.Reply
	if err = json.Unmarshal(p.bridge.Dispatch(data), &reply); err != nil {
		return moduleapi.Snapshot{}, errors.New("module response unavailable")
	}
	if reply.Error != "" && !reply.Committed {
		return moduleapi.Snapshot{}, errors.New("module command rejected; check action, revision and typed input")
	}
	p.revision = reply.Revision
	p.initialized = true
	actions := []moduleapi.Action{}
	for _, op := range hysteresis.Operations() {
		actions = append(actions, moduleapi.Action{ID: op.Name, Label: op.Name, Available: true})
	}
	status := "ready"
	if reply.Error != "" {
		status = "failed"
	}
	p.last = moduleapi.Snapshot{Session: p.id, Revision: p.revision, Descriptor: p.Descriptor(), Status: status, Evidence: []string{"simulation-default"}, Fields: []moduleapi.Field{}, Actions: actions, Tables: []moduleapi.TablePage{}, Plots: []moduleapi.Plot{}, Bridge: &reply}
	return p.last, nil
}
func (p *hysteresisPort) Close() { p.bridge = nil }
