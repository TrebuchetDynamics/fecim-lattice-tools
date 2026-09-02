package experiment

import (
	"bytes"
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
)

const maxRunArtifactBytes int64 = 16 << 20

var (
	runIDRE            = regexp.MustCompile(`^[0-9a-f]{64}$`)
	ErrUnverifiableRun = errors.New("experiment: unverifiable cached run")
)

type store struct {
	root   string
	rename func(string, string) error
}

func (s store) load(id string) (RunRecord, bool, error) {
	if !runIDRE.MatchString(id) {
		return RunRecord{}, false, fmt.Errorf("invalid run ID %q", id)
	}
	dir := filepath.Join(s.root, "runs", id)
	info, err := os.Stat(dir)
	if errors.Is(err, os.ErrNotExist) {
		return RunRecord{}, false, nil
	}
	if err != nil {
		return RunRecord{}, false, err
	}
	if !info.IsDir() {
		return RunRecord{}, false, fmt.Errorf("run path %s is not a directory", dir)
	}

	var record RunRecord
	manifestBytes, err := readRunArtifact(filepath.Join(dir, "manifest.json"))
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
	designBytes, err := readRunArtifact(filepath.Join(dir, "resolved-design.json"))
	if err != nil {
		return RunRecord{}, false, fmt.Errorf("%w: read resolved-design.json for run %s: %v", ErrUnverifiableRun, id, err)
	}
	if err := decodeRunArtifact(designBytes, &record.Design); err != nil {
		return RunRecord{}, false, fmt.Errorf("%w: decode resolved-design.json for run %s: %v", ErrUnverifiableRun, id, err)
	}
	resultBytes, err := readRunArtifact(filepath.Join(dir, "result.json"))
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
	expectedID, err := ID(DesignPoint{Design: record.Design, Seed: record.Manifest.Seed}, record.Manifest.EvaluatorVersion, record.Manifest.Inputs)
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

func (s store) commit(record RunRecord) (RunRecord, error) {
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

	runsDir := filepath.Join(s.root, "runs")
	if err := os.MkdirAll(runsDir, 0o755); err != nil {
		return RunRecord{}, err
	}
	if existing, ok, err := s.load(id); err != nil {
		return RunRecord{}, err
	} else if ok {
		existing.Reused = true
		return existing, nil
	}

	tempDir, err := os.MkdirTemp(runsDir, ".tmp-"+id+"-")
	if err != nil {
		return RunRecord{}, err
	}
	defer os.RemoveAll(tempDir)
	if err := writeExclusive(filepath.Join(tempDir, "manifest.json"), manifestBytes); err != nil {
		return RunRecord{}, fmt.Errorf("write manifest.json: %w", err)
	}
	if err := writeExclusive(filepath.Join(tempDir, "resolved-design.json"), designBytes); err != nil {
		return RunRecord{}, fmt.Errorf("write resolved-design.json: %w", err)
	}
	if err := writeExclusive(filepath.Join(tempDir, "result.json"), resultBytes); err != nil {
		return RunRecord{}, fmt.Errorf("write result.json: %w", err)
	}
	finalDir := filepath.Join(runsDir, id)
	rename := s.rename
	if rename == nil {
		rename = os.Rename
	}
	if err := rename(tempDir, finalDir); err != nil {
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
	return record, nil
}

func LoadRuns(root string) ([]RunRecord, error) {
	runsDir := filepath.Join(root, "runs")
	entries, err := os.ReadDir(runsDir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	storage := store{root: root}
	runs := make([]RunRecord, 0, len(entries))
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".tmp-") {
			continue
		}
		if !entry.IsDir() {
			return nil, fmt.Errorf("run path %s is not a directory", filepath.Join(runsDir, entry.Name()))
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

func readRunArtifact(path string) ([]byte, error) {
	file, err := os.Open(path)
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

func writeExclusive(path string, data []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}
