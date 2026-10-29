package docs

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"

	"fecim-lattice-tools/workbench/moduleapi"
	nativeapi "fecim-lattice-tools/workbench/moduleapi/native"
	core "fecim-lattice-tools/workbench/modules/docs"
)

type nativePort struct {
	port            *core.Port
	err             error
	revision        uint64
	historyRoot     *moduleapi.ResourceReference
	historyRevision string
}

func newNativePort() moduleapi.Port                          { p, err := core.NewPort(); return &nativePort{port: p, err: err} }
func (p *nativePort) Descriptor() moduleapi.ModuleDescriptor { return core.Descriptor() }
func (p *nativePort) Snapshot(ctx context.Context) (moduleapi.Snapshot, error) {
	if p.err != nil {
		return moduleapi.Snapshot{}, p.err
	}
	s, err := p.port.Snapshot(ctx)
	s.Revision = p.revision
	return s, err
}
func (p *nativePort) Apply(ctx context.Context, c moduleapi.Command) (moduleapi.Snapshot, error) {
	if p.err != nil {
		return moduleapi.Snapshot{}, p.err
	}
	if c.Revision != p.revision {
		return moduleapi.Snapshot{}, errors.New("stale docs revision")
	}
	if err := ctx.Err(); err != nil {
		return moduleapi.Snapshot{}, err
	}
	var err error
	switch c.Action {
	case "import_granted":
		if len(c.Payload) > 0 || len(c.Input) != 1 || c.Input["id"] == "" {
			return moduleapi.Snapshot{}, errors.New("choose a document ID, not an arbitrary payload")
		}
	case "history_optin":
		if len(c.Payload) > 0 || len(c.Input) != 1 || (c.Input["enabled"] != "true" && c.Input["enabled"] != "false") || c.Input["enabled"] == "false" && len(c.Resources) > 0 {
			return moduleapi.Snapshot{}, errors.New("explicit true/false history intent required")
		}
	case "history_load", "history_save":
		if len(c.Payload) > 0 || len(c.Input) > 0 || len(c.Resources) > 0 {
			return moduleapi.Snapshot{}, errors.New("history uses only the opted-in grant")
		}
	}
	switch c.Action {
	case "import_granted":
		if len(c.Resources) != 1 || c.Resources[0].Kind != "reference" {
			return moduleapi.Snapshot{}, errors.New("choose one granted Markdown reference")
		}
		resolver, ok := nativeapi.Resources(ctx)
		if !ok {
			return moduleapi.Snapshot{}, errors.New("local document grants unavailable")
		}
		reader, info, e := resolver.Open(ctx, c.Resources[0])
		if e != nil {
			return moduleapi.Snapshot{}, e
		}
		defer reader.Close()
		if info.Size > core.MaxDocumentBytes {
			return moduleapi.Snapshot{}, errors.New("document exceeds 1 MiB")
		}
		data, e := io.ReadAll(io.LimitReader(reader, core.MaxDocumentBytes+1))
		if e != nil {
			return moduleapi.Snapshot{}, e
		}
		if len(data) > core.MaxDocumentBytes {
			return moduleapi.Snapshot{}, errors.New("document exceeds 1 MiB")
		}
		err = p.port.AddGranted(c.Input["id"], data)
	case "history_optin":
		if c.Input["enabled"] != "true" {
			p.historyRoot = nil
			p.historyRevision = ""
			break
		}
		if len(c.Resources) != 1 || c.Resources[0].Kind != "project" {
			return moduleapi.Snapshot{}, errors.New("choose an already opened granted project")
		}
		resolver, ok := nativeapi.Projects(ctx)
		if !ok {
			return moduleapi.Snapshot{}, errors.New("local project grants unavailable")
		}
		// Validate the grant without creating directories or reading/writing history.
		err = resolver.WithProject(ctx, c.Resources[0], false, func(root *os.Root) error { return nil })
		if err == nil {
			ref := c.Resources[0]
			p.historyRoot = &ref
			p.historyRevision = ""
		}
	case "history_load", "history_save":
		if p.historyRoot == nil {
			return moduleapi.Snapshot{}, errors.New("history requires explicit opt-in and a granted project")
		}
		resolver, ok := nativeapi.Projects(ctx)
		if !ok {
			return moduleapi.Snapshot{}, errors.New("local project grants unavailable")
		}
		err = resolver.WithProject(ctx, *p.historyRoot, c.Action == "history_save", func(root *os.Root) error {
			if c.Action == "history_load" {
				data, e := readHistory(root)
				if e != nil {
					return e
				}
				var h core.History
				decoder := json.NewDecoder(bytes.NewReader(data))
				decoder.DisallowUnknownFields()
				if e = decoder.Decode(&h); e != nil {
					return e
				}
				if decoder.Decode(new(any)) != io.EOF {
					return errors.New("trailing history data")
				}
				if e = p.port.RestoreHistory(h); e != nil {
					return e
				}
				p.historyRevision = core.Hash(data)
				return nil
			}
			// Project serialization plus the last loaded/saved byte hash protects
			// against stale writers. An existing file must be loaded first.
			current, e := readHistory(root)
			actual := ""
			if e == nil {
				actual = core.Hash(current)
			} else if !errors.Is(e, os.ErrNotExist) {
				return e
			}
			if actual != p.historyRevision {
				return errors.New("history changed or already exists; load and review it before saving")
			}
			out, e := p.port.Output()
			if e != nil {
				return e
			}
			data, e := json.Marshal(out.History)
			if e != nil {
				return e
			}
			var random [12]byte
			if _, e = rand.Read(random[:]); e != nil {
				return e
			}
			name := "docs-history-" + hex.EncodeToString(random[:]) + ".tmp"
			f, e := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
			if e != nil {
				return e
			}
			defer root.Remove(name)
			_, e = f.Write(data)
			if e != nil {
				f.Close()
				return e
			}
			if e = f.Sync(); e != nil {
				f.Close()
				return e
			}
			if e = f.Close(); e != nil {
				return e
			}
			if e = root.Rename(name, "docs-history.json"); e != nil {
				return e
			}
			p.historyRevision = core.Hash(data)
			return nil
		})
	default:
		s, e := p.port.Snapshot(ctx)
		if e != nil {
			return s, e
		}
		c.Revision = s.Revision
		_, err = p.port.Apply(ctx, c)
	}
	if err != nil {
		return moduleapi.Snapshot{}, err
	}
	p.revision++
	return p.Snapshot(ctx)
}
func readHistory(root *os.Root) ([]byte, error) {
	info, err := root.Lstat("docs-history.json")
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > 16384 {
		return nil, errors.New("history must be a bounded regular file")
	}
	f, err := root.Open("docs-history.json")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	after, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !after.Mode().IsRegular() || !os.SameFile(info, after) {
		return nil, errors.New("history file changed during read")
	}
	data, err := io.ReadAll(io.LimitReader(f, 16385))
	if err != nil {
		return nil, err
	}
	if len(data) > 16384 {
		return nil, errors.New("history byte limit")
	}
	return data, nil
}
func (p *nativePort) Close() {
	p.err = errors.New("docs session closed")
	p.historyRoot = nil
	if p.port != nil {
		p.port.Close()
	}
}
