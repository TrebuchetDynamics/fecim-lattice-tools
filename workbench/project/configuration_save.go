package project

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

var (
	ErrRevisionRequired     = errors.New("configuration revision is required")
	ErrInvalidRevision      = errors.New("configuration revision must be an exact strong cfg-v1 revision")
	configurationRevisionRE = regexp.MustCompile(`^"cfg-v1-[0-9a-f]{64}"$`)
)

// ConfigurationSaveResult describes the committed (or unchanged) bundle bytes.
type ConfigurationSaveResult struct {
	Revision string
	Changed  bool
}

// ConfigurationFieldConflict retains literal input and the current scalar text.
// OnDiskAvailable distinguishes a missing/ambiguous value from an empty string
// or zero. Reason explains unavailable values; these are not display/speech units.
type ConfigurationFieldConflict struct {
	Path, Yours, OnDisk string
	OnDiskAvailable     bool
	Reason              string
}

// ConfigurationConflict reports a stale revision, including for no-op saves.
type ConfigurationConflict struct {
	CurrentRevision string
	Fields          []ConfigurationFieldConflict
}

func (e *ConfigurationConflict) Error() string { return "configuration revision is stale" }

// ConfigurationEditor serializes saves participating in this editor only. Use
// one editor per project handle, and keep the caller-owned roots open throughout
// its lifetime. It must not be copied after use. External writers do not share
// its lock: the final revision recheck and rename are not an atomic CAS.
type ConfigurationEditor struct {
	mu        sync.Mutex
	root      *os.Root
	citations *os.Root
	fs        configurationSaveFS
}

// Private filesystem seams permit deterministic persistence faults/interleavings.
// Production instances always use real root operations and file methods.
type configurationSaveFS struct {
	open   func(string, int, os.FileMode) (*os.File, error)
	stat   func(*os.File) (os.FileInfo, error)
	write  func(*os.File, []byte) (int, error)
	sync   func(*os.File) error
	close  func(*os.File) error
	rename func(string, string) error
}

// NewConfigurationEditor borrows roots; it never opens roots by name or closes
// the caller's handles. Nil citations skips citation existence checks.
func NewConfigurationEditor(root, citations *os.Root) (*ConfigurationEditor, error) {
	if root == nil {
		return nil, errors.New("project root is required")
	}
	for _, handle := range []*os.Root{root, citations} {
		if handle != nil {
			if _, err := handle.Stat("."); err != nil {
				return nil, err
			}
		}
	}
	return &ConfigurationEditor{
		root: root, citations: citations,
		fs: configurationSaveFS{
			open: root.OpenFile, stat: (*os.File).Stat, write: (*os.File).Write, sync: (*os.File).Sync,
			close: (*os.File).Close, rename: root.Rename,
		},
	}, nil
}

// Save validates a single-document field patch and guards it with the exact
// bundle revision. A no-op never creates a temporary file or encodes YAML.
// For changed saves, successful root.Rename is the commit point: cancellation
// after it cannot report an uncommitted save. Atomic replacement is qualified
// for local Linux filesystems, not power-loss durability or cross-process CAS.
func (e *ConfigurationEditor) Save(ctx context.Context, expectedRevision, document string, changes []FieldChange) (ConfigurationSaveResult, error) {
	if expectedRevision == "" {
		return ConfigurationSaveResult{}, ErrRevisionRequired
	}
	if !configurationRevisionRE.MatchString(expectedRevision) {
		return ConfigurationSaveResult{}, ErrInvalidRevision
	}
	if err := ctx.Err(); err != nil {
		return ConfigurationSaveResult{}, err
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return ConfigurationSaveResult{}, err
	}
	base, err := e.checkRevision(expectedRevision, document, changes)
	if err != nil {
		return ConfigurationSaveResult{}, err
	}
	prepared, err := PrepareConfiguration(e.root, e.citations, base, document, changes)
	if err != nil {
		return ConfigurationSaveResult{}, err
	}
	if !prepared.Changed() {
		// Validation can read inputs/citations; observe intervening config edits
		// even on this path, without touching file bytes or timestamps.
		if _, err := e.checkRevision(expectedRevision, document, changes); err != nil {
			return ConfigurationSaveResult{}, err
		}
		if err := ctx.Err(); err != nil {
			return ConfigurationSaveResult{}, err
		}
		return ConfigurationSaveResult{Revision: base.Revision()}, nil
	}
	if err := ctx.Err(); err != nil {
		return ConfigurationSaveResult{}, err
	}
	info, err := e.root.Lstat(document)
	if err != nil {
		return ConfigurationSaveResult{}, err
	}
	if !info.Mode().IsRegular() {
		return ConfigurationSaveResult{}, errors.New("configuration must be a regular file")
	}
	// A cryptographic name, O_EXCL, and an initially private permission subset
	// avoid following links or exposing candidate bytes before replacement.
	name := fmt.Sprintf(".fecim-config-%x.tmp", randomConfigurationName())
	file, err := e.fs.open(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, info.Mode().Perm()&0o600)
	if err != nil {
		return ConfigurationSaveResult{}, err // Never remove a file we didn't create.
	}
	defer file.Close()
	var owned os.FileInfo
	committed := false
	defer func() {
		if !committed {
			// Register cleanup before the first metadata check. If it failed,
			// retry through the still-open descriptor to establish ownership.
			// If identity cannot be recovered, do not guess from the path alone.
			if owned == nil {
				owned, _ = file.Stat()
			}
			// Do not delete a nonparticipating writer's substituted temp path.
			if current, err := e.root.Lstat(name); err == nil && owned != nil && os.SameFile(owned, current) {
				e.root.Remove(name)
			}
		}
	}()
	owned, err = e.fs.stat(file)
	if err != nil {
		return ConfigurationSaveResult{}, err
	}
	if !owned.Mode().IsRegular() {
		return ConfigurationSaveResult{}, errors.New("configuration temporary must be a regular file")
	}
	data := []byte(prepared.snapshot.documents[configurationDocumentIndex(document)])
	if n, err := e.fs.write(file, data); err != nil {
		return ConfigurationSaveResult{}, err
	} else if n != len(data) {
		return ConfigurationSaveResult{}, io.ErrShortWrite
	}
	// Preserve permission bits, not setuid/setgid/sticky bits or ownership/ACLs.
	if err := file.Chmod(info.Mode().Perm()); err != nil {
		return ConfigurationSaveResult{}, err
	}
	if err := e.fs.sync(file); err != nil {
		return ConfigurationSaveResult{}, err
	}
	if err := e.fs.close(file); err != nil {
		return ConfigurationSaveResult{}, err
	}
	current, err := e.root.Lstat(name)
	if err != nil {
		return ConfigurationSaveResult{}, err
	}
	if !current.Mode().IsRegular() || !os.SameFile(owned, current) {
		return ConfigurationSaveResult{}, errors.New("configuration temporary changed before replacement")
	}
	// Do not broaden permissions if they were restricted during preparation.
	current, err = e.root.Lstat(document)
	if err != nil {
		return ConfigurationSaveResult{}, err
	}
	if !current.Mode().IsRegular() || current.Mode().Perm() != info.Mode().Perm() {
		return ConfigurationSaveResult{}, errors.New("configuration permissions or file type changed before replacement")
	}
	// Last bundle read before replacement; a nonparticipating writer can still
	// change a document after this check. This deliberately makes no CAS claim.
	if _, err := e.checkRevision(expectedRevision, document, changes); err != nil {
		return ConfigurationSaveResult{}, err
	}
	if err := ctx.Err(); err != nil {
		return ConfigurationSaveResult{}, err
	}
	if err := e.fs.rename(name, document); err != nil {
		return ConfigurationSaveResult{}, err
	}
	committed = true
	return ConfigurationSaveResult{Revision: prepared.snapshot.Revision(), Changed: true}, nil
}

func randomConfigurationName() []byte {
	// crypto/rand.Read fills the buffer or terminates on entropy failure.
	var name [16]byte
	rand.Read(name[:])
	return name[:]
}

func (e *ConfigurationEditor) checkRevision(expected, document string, changes []FieldChange) (ConfigurationSnapshot, error) {
	current, err := ReadConfiguration(e.root)
	if err != nil {
		return ConfigurationSnapshot{}, err
	}
	if current.Revision() == expected {
		return current, nil
	}
	fields := make([]ConfigurationFieldConflict, 0, len(changes))
	var tree yaml.Node
	index := configurationDocumentIndex(document)
	var treeErr error
	if index < 0 {
		treeErr = errors.New("unsupported configuration document")
	} else if treeErr = decodeStrict(strings.NewReader(current.documents[index]), &tree); treeErr == nil {
		treeErr = editableConfigurationTree(&tree)
	}
	for _, change := range changes {
		field := ConfigurationFieldConflict{Path: change.Path, Yours: change.Value}
		err := treeErr
		if err == nil {
			var parts []string
			parts, err = configurationPath(change.Path)
			if err == nil && configurationFieldKind(document, parts) == "" {
				err = errors.New("unsupported configuration field path")
			}
			if err == nil {
				var scalar *yaml.Node
				scalar, err = configurationScalar(&tree, parts)
				if err == nil {
					field.OnDisk, field.OnDiskAvailable = scalar.Value, true
				}
			}
		}
		if err != nil {
			field.Reason = err.Error()
		}
		fields = append(fields, field)
	}
	return ConfigurationSnapshot{}, &ConfigurationConflict{CurrentRevision: current.Revision(), Fields: fields}
}
