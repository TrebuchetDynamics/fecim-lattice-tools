package experiment

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
)

const maxRunArtifactBytes int64 = 16 << 20

var (
	runIDRE            = regexp.MustCompile(`^[0-9a-f]{64}$`)
	ErrUnverifiableRun = errors.New("experiment: unverifiable cached run")
)

// store owns only child handles. Once opened, the runs namespace is never
// resolved again, including during publication failure recovery.
type store struct {
	root     *os.Root
	runsPath string
	mu       sync.Mutex
	runs     *os.Root
	rename   func(string, string) error
}

func openStore(root *os.Root) (*store, error) {
	if root == nil {
		return nil, errors.New("experiment root is required")
	}
	owned, err := root.OpenRoot(".")
	if err != nil {
		return nil, err
	}
	return pinStore(owned, "runs")
}

// Legacy Run may target a workspace that does not yet exist. Pin its nearest
// existing ancestor now, then create only the missing relative path at commit.
// No later I/O reopens the original string path.
func openStorePath(path string) (*store, error) {
	ancestor, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	runsPath := "runs"
	for {
		root, err := os.OpenRoot(ancestor)
		if err == nil {
			return pinStore(root, runsPath)
		}
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		parent := filepath.Dir(ancestor)
		if parent == ancestor {
			return nil, err
		}
		runsPath = filepath.Join(filepath.Base(ancestor), runsPath)
		ancestor = parent
	}
}

// pinStore takes ownership of the already-open root, including on failure.
func pinStore(owned *os.Root, runsPath string) (*store, error) {
	s := &store{root: owned, runsPath: runsPath}
	if _, err := s.runsRoot(false); err != nil {
		s.close()
		return nil, err
	}
	return s, nil
}

// runsRoot serializes lazy creation across workers. Invalid/oversized evaluator
// results must not create storage, but an existing namespace is pinned up front.
func (s *store) runsRoot(create bool) (*os.Root, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.runs != nil {
		return s.runs, nil
	}
	if create {
		if err := s.root.MkdirAll(s.runsPath, 0o755); err != nil {
			return nil, err
		}
	}
	runs, err := s.root.OpenRoot(s.runsPath)
	if !create && errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	s.runs = runs
	return runs, nil
}

// close is called only after all workers have settled.
func (s *store) close() {
	if s.runs != nil {
		_ = s.runs.Close()
	}
	_ = s.root.Close()
}

func (s *store) load(id string) (RunRecord, bool, error) {
	if !runIDRE.MatchString(id) {
		return RunRecord{}, false, fmt.Errorf("invalid run ID %q", id)
	}
	runs, err := s.runsRoot(false)
	if err != nil || runs == nil {
		return RunRecord{}, false, err
	}
	dir, err := runs.OpenRoot(id)
	if errors.Is(err, os.ErrNotExist) {
		return RunRecord{}, false, nil
	}
	if err != nil {
		return RunRecord{}, false, err
	}
	defer dir.Close()

	var record RunRecord
	manifestBytes, err := readRunArtifact(dir, "manifest.json")
	if err != nil {
		return RunRecord{}, false, fmt.Errorf("%w: read manifest.json for run %s: %v", ErrUnverifiableRun, id, err)
	}
	if err := decodeRunArtifact(manifestBytes, &record.Manifest); err != nil {
		return RunRecord{}, false, fmt.Errorf("%w: decode manifest.json for run %s: %v", ErrUnverifiableRun, id, err)
	}
	if record.Manifest.SchemaVersion != 2 {
		return RunRecord{}, false, fmt.Errorf("%w: run %s uses manifest schema %d; archive or remove runs/%s/ and rerun", ErrUnverifiableRun, id, record.Manifest.SchemaVersion, id)
	}
	canonicalManifest, err := canonicalJSON(record.Manifest)
	if err != nil {
		return RunRecord{}, false, fmt.Errorf("%w: canonicalize manifest.json for run %s: %v", ErrUnverifiableRun, id, err)
	}
	if !bytes.Equal(manifestBytes, canonicalManifest) {
		return RunRecord{}, false, fmt.Errorf("%w: run %s manifest.json is not canonical", ErrUnverifiableRun, id)
	}
	designBytes, err := readRunArtifact(dir, "resolved-design.json")
	if err != nil {
		return RunRecord{}, false, fmt.Errorf("%w: read resolved-design.json for run %s: %v", ErrUnverifiableRun, id, err)
	}
	if err := decodeRunArtifact(designBytes, &record.Design); err != nil {
		return RunRecord{}, false, fmt.Errorf("%w: decode resolved-design.json for run %s: %v", ErrUnverifiableRun, id, err)
	}
	resultBytes, err := readRunArtifact(dir, "result.json")
	if err != nil {
		return RunRecord{}, false, fmt.Errorf("%w: read result.json for run %s: %v", ErrUnverifiableRun, id, err)
	}
	if err := decodeRunArtifact(resultBytes, &record.Result); err != nil {
		return RunRecord{}, false, fmt.Errorf("%w: decode result.json for run %s: %v", ErrUnverifiableRun, id, err)
	}

	if err := verifyArtifact("resolved-design.json", designBytes, record.Manifest); err != nil {
		return RunRecord{}, false, err
	}
	if err := verifyArtifact("result.json", resultBytes, record.Manifest); err != nil {
		return RunRecord{}, false, err
	}
	expectedID, err := ID(DesignPoint{Design: record.Design, ModelVersion: record.Manifest.ModelVersion, Seed: record.Manifest.Seed}, record.Manifest.EvaluatorVersion, record.Manifest.Inputs)
	if err != nil {
		return RunRecord{}, false, fmt.Errorf("%w: recompute run %s identity: %v", ErrUnverifiableRun, id, err)
	}
	if expectedID != id || record.Manifest.RunID != id {
		return RunRecord{}, false, fmt.Errorf("%w: run %s identity mismatch", ErrUnverifiableRun, id)
	}
	if record.Manifest.Status != record.Result.Status {
		return RunRecord{}, false, fmt.Errorf("%w: run %s has inconsistent status", ErrUnverifiableRun, id)
	}
	recordDigest, err := recordSHA256(record.Manifest, designBytes, resultBytes)
	if err != nil || record.Manifest.RecordSHA256 == "" || recordDigest != record.Manifest.RecordSHA256 {
		return RunRecord{}, false, fmt.Errorf("%w: run %s record digest mismatch", ErrUnverifiableRun, id)
	}
	return record, true, nil
}

func (s *store) commit(record RunRecord) (RunRecord, error) {
	id := record.Manifest.RunID
	if !runIDRE.MatchString(id) {
		return RunRecord{}, fmt.Errorf("invalid run ID %q", id)
	}
	if record.Manifest.SchemaVersion != 2 {
		return RunRecord{}, fmt.Errorf("commit requires manifest schema 2, got %d", record.Manifest.SchemaVersion)
	}
	designBytes, err := canonicalJSON(record.Design)
	if err != nil {
		return RunRecord{}, fmt.Errorf("encode resolved design: %w", err)
	}
	if err := checkRunArtifactSize("resolved-design.json", designBytes); err != nil {
		return RunRecord{}, err
	}
	resultBytes, err := canonicalJSON(record.Result)
	if err != nil {
		return RunRecord{}, fmt.Errorf("encode result: %w", err)
	}
	if err := checkRunArtifactSize("result.json", resultBytes); err != nil {
		return RunRecord{}, err
	}
	record.Manifest.ArtifactSHA256 = map[string]string{
		"resolved-design.json": sha256Hex(designBytes),
		"result.json":          sha256Hex(resultBytes),
	}
	recordDigest, err := recordSHA256(record.Manifest, designBytes, resultBytes)
	if err != nil {
		return RunRecord{}, fmt.Errorf("digest run record: %w", err)
	}
	record.Manifest.RecordSHA256 = recordDigest
	manifestBytes, err := canonicalJSON(record.Manifest)
	if err != nil {
		return RunRecord{}, fmt.Errorf("encode manifest: %w", err)
	}
	if err := checkRunArtifactSize("manifest.json", manifestBytes); err != nil {
		return RunRecord{}, err
	}

	runs, err := s.runsRoot(true)
	if err != nil {
		return RunRecord{}, err
	}
	if existing, ok, err := s.load(id); err != nil {
		return RunRecord{}, err
	} else if ok {
		existing.Reused = true
		return existing, nil
	}

	// Mkdir is exclusive and the unpredictable name is relative to the pinned
	// namespace. A collision never grants ownership of an existing entry.
	var tempDir string
	for {
		tempDir = ".tmp-" + id + "-" + rand.Text()
		err = runs.Mkdir(tempDir, 0o700)
		if !errors.Is(err, os.ErrExist) {
			break
		}
	}
	if err != nil {
		return RunRecord{}, err
	}
	scratch, err := runs.OpenRoot(tempDir)
	if err != nil {
		_ = runs.Remove(tempDir) // Never recursively clean an unpinned entry.
		return RunRecord{}, err
	}
	published := false
	defer func() {
		if !published {
			// Clean through our handle, not a possibly retargeted scratch name.
			for _, name := range []string{"manifest.json", "resolved-design.json", "result.json"} {
				_ = scratch.Remove(name)
			}
			owned, ownErr := scratch.Stat(".")
			entry, entryErr := runs.Lstat(tempDir)
			if ownErr == nil && entryErr == nil && os.SameFile(owned, entry) {
				_ = runs.Remove(tempDir)
			}
		}
		_ = scratch.Close()
	}()
	if err := writeExclusive(scratch, "manifest.json", manifestBytes); err != nil {
		return RunRecord{}, fmt.Errorf("write manifest.json: %w", err)
	}
	if err := writeExclusive(scratch, "resolved-design.json", designBytes); err != nil {
		return RunRecord{}, fmt.Errorf("write resolved-design.json: %w", err)
	}
	if err := writeExclusive(scratch, "result.json", resultBytes); err != nil {
		return RunRecord{}, fmt.Errorf("write result.json: %w", err)
	}
	rename := s.rename
	if rename == nil {
		rename = runs.Rename
	}
	if err := rename(tempDir, id); err != nil {
		existing, ok, loadErr := s.load(id)
		if loadErr != nil {
			return RunRecord{}, loadErr
		}
		if ok {
			existing.Reused = true
			return existing, nil
		}
		return RunRecord{}, err
	}
	published = true
	return record, nil
}

func LoadRuns(path string) ([]RunRecord, error) {
	root, err := os.OpenRoot(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer root.Close()
	return LoadRunsRoot(root)
}

// LoadRunsRoot verifies and lists committed runs through pinned child handles.
// The caller retains root and must keep it open until this operation returns.
func LoadRunsRoot(root *os.Root) ([]RunRecord, error) {
	storage, err := openStore(root)
	if err != nil {
		return nil, err
	}
	defer storage.close()
	runsRoot, err := storage.runsRoot(false)
	if err != nil || runsRoot == nil {
		return nil, err
	}
	dir, err := runsRoot.Open(".")
	if err != nil {
		return nil, err
	}
	entries, readErr := dir.ReadDir(-1)
	closeErr := dir.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	runs := make([]RunRecord, 0, len(entries))
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".tmp-") {
			continue
		}
		if !entry.IsDir() {
			return nil, fmt.Errorf("run path %s is not a directory", filepath.Join("runs", entry.Name()))
		}
		record, ok, err := storage.load(entry.Name())
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("run %s disappeared while loading", entry.Name())
		}
		runs = append(runs, record)
	}
	sort.Slice(runs, func(i, j int) bool {
		if runs[i].Manifest.PointIndex == runs[j].Manifest.PointIndex {
			return runs[i].Manifest.RunID < runs[j].Manifest.RunID
		}
		return runs[i].Manifest.PointIndex < runs[j].Manifest.PointIndex
	})
	return runs, nil
}

func canonicalJSON(value any) ([]byte, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func verifyArtifact(name string, data []byte, manifest RunManifest) error {
	want := manifest.ArtifactSHA256[name]
	if want == "" || sha256Hex(data) != want {
		return fmt.Errorf("%w: %s digest mismatch", ErrUnverifiableRun, name)
	}
	return nil
}

func recordSHA256(manifest RunManifest, designBytes, resultBytes []byte) (string, error) {
	manifest.RecordSHA256 = ""
	payload := struct {
		Manifest RunManifest     `json:"manifest"`
		Design   json.RawMessage `json:"resolved_design"`
		Result   json.RawMessage `json:"result"`
	}{
		Manifest: manifest,
		Design:   json.RawMessage(designBytes),
		Result:   json.RawMessage(resultBytes),
	}
	data, err := canonicalJSON(payload)
	if err != nil {
		return "", err
	}
	return sha256Hex(data), nil
}

func checkRunArtifactSize(name string, data []byte) error {
	if int64(len(data)) > maxRunArtifactBytes {
		return fmt.Errorf("%s exceeds 16 MiB", name)
	}
	return nil
}

func readRunArtifact(root *os.Root, path string) ([]byte, error) {
	file, err := root.Open(path)
	if err != nil {
		return nil, err
	}
	data, readErr := io.ReadAll(io.LimitReader(file, maxRunArtifactBytes+1))
	closeErr := file.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	if err := checkRunArtifactSize(filepath.Base(path), data); err != nil {
		return nil, err
	}
	return data, nil
}

func decodeRunArtifact(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("multiple JSON values")
		}
		return fmt.Errorf("trailing JSON data: %w", err)
	}
	return nil
}

func writeExclusive(root *os.Root, path string, data []byte) error {
	file, err := root.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}
