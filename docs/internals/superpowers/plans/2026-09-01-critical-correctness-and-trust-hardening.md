# Critical Correctness and Trust Hardening Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove the five P0 defects that currently allow incorrect peripheral noise, unverifiable cached research runs, corrupt or oversized model binaries, shell interpretation of EDA arguments, and false-positive external Verilog validation.

**Architecture:** Keep the existing Go modules and interfaces. Deepen each existing trust seam with the smallest enforceable invariant: sampled/clipped noise at the peripheral, schema-v2 integrity metadata at the run store, bounded rectangular model serialization, fixed shell programs with dynamic positional arguments, and one required CI lane that executes a real Verilog tool. No broad UI or package restructure belongs in this plan.

**Tech Stack:** Go 1.25, Go standard library, Fyne-independent package tests, GitHub Actions, and Icarus Verilog (`iverilog`) installed by the CI runner with its exact resolved version recorded. Immutable CI runner/tool pinning remains a TR-CI-01 follow-up.

## Global Constraints

- Follow test-driven development: record RED before production changes, then GREEN, then refactor.
- Do not regenerate physics golden files; none of these tasks intentionally changes ferroelectric model calibration.
- Add no Go dependency.
- Preserve existing public behavior except where it is unsafe or falsely documented.
- Reject unverifiable legacy cached runs; do not silently reuse them and do not delete them automatically.
- Detect accidental/local-file corruption. Do not claim protection against an attacker who can rewrite every unsigned local artifact and manifest.
- Treat model binaries and EDA paths/arguments as untrusted input.
- External validation counts only when the named executable actually ran and exited successfully.
- Keep unrelated dirty owner files out of every commit.

---

## Scope Decomposition

This plan is the first independently shippable slice of the broader [Technical Recovery Roadmap](../../audits/2026-09-01-technical-recovery-roadmap.md). Follow-up plans cover:

1. executable project-bundle and committed-result invariants;
2. claim registry, pinned OpenLane image, CI vulnerability/format gates;
3. launcher consolidation, `ModulePort`/Fyne convergence, and electrical-solver ownership;
4. log retention, documentation drift, and research-PDF licensing.

Do not fold those follow-ups into this plan.

## File Structure

Modify:

- `shared/peripherals/charge_amplifier.go` — sample zero-mean Gaussian noise and clip after noise.
- `shared/peripherals/charge_amplifier_test.go` — deterministic injected-noise and rail regression tests.
- `workbench/experiment/types.go` — schema-v2 run artifact digest metadata.
- `workbench/experiment/store.go` — canonical artifact bytes, digest checks, identity recomputation, legacy-cache rejection.
- `workbench/experiment/runner.go` — commit schema-v2 manifests.
- `workbench/experiment/runner_test.go` — design/result corruption and legacy-cache tests.
- `module2-crossbar/pkg/weights/serialization.go` — rectangularity checks and bounded binary decoding.
- `module2-crossbar/pkg/weights/weights_test.go` — ragged-save and oversized-length tests.
- `module6-eda/pkg/openlane/runner.go` — fixed Docker shell programs and pure argument builders.
- `module6-eda/pkg/openlane/openlane_test.go` — adversarial argument-construction tests.
- `validation/external/eda/verilog_sanity_test.go` — real iverilog/verilator invocation.
- `.github/workflows/ci.yml` — required installed external Verilog validator.
- `scripts/ci_contract_test.go` — new workflow contract proving the required validator is installed and gating.
- `TODO.md` and `docs/internals/audits/2026-09-01-technical-recovery-roadmap.md` — completion receipts after all gates pass.

No new production package is required.

---

### Task 1: Correct charge-amplifier noise semantics

**Files:**
- Modify: `shared/peripherals/charge_amplifier.go:19,96-111`
- Modify: `shared/peripherals/charge_amplifier_test.go:23-49`

**Interfaces:**
- Consumes: `ChargeAmplifier.Sense(float64) float64`, existing noise parameters and output rails.
- Produces: unchanged public `SenseWithNoise(float64) float64`; private `senseWithGaussian(float64, func() float64) float64` test seam.

**Trust wording after completion:** The behavioral model samples zero-mean Gaussian noise and applies output rails after signal-plus-noise. It is still a simulation default, not measured hardware calibration.  
**Rollback:** Revert the task commit; no persisted format or user data changes.

- [ ] **Step 1: Write deterministic failing noise tests**

Append to `shared/peripherals/charge_amplifier_test.go`:

```go
func TestChargeAmplifier_SenseWithNoiseUsesSignedGaussianSample(t *testing.T) {
    ca := DefaultChargeAmplifier()
    sigmaV := ca.InputChargeNoiseRMS * math.Sqrt(ca.Bandwidth) / ca.Cfb

    positive := ca.senseWithGaussian(0, func() float64 { return 1 })
    negative := ca.senseWithGaussian(0, func() float64 { return -1 })

    if math.Abs(positive-sigmaV) > 1e-15 {
        t.Fatalf("positive sample=%e want %e", positive, sigmaV)
    }
    if math.Abs(negative+sigmaV) > 1e-15 {
        t.Fatalf("negative sample=%e want -%e", negative, sigmaV)
    }
    if math.Abs(positive+negative) > 1e-15 {
        t.Fatalf("injected symmetric samples are biased: +%e -%e", positive, negative)
    }
}

func TestChargeAmplifier_SenseWithNoiseSamplesPublicSource(t *testing.T) {
    ca := DefaultChargeAmplifier()
    ca.Bandwidth = 1
    first := ca.SenseWithNoise(0)
    second := ca.SenseWithNoise(0)
    if first == second {
        t.Fatalf("two sampled outputs are identical: %e", first)
    }
}

func TestChargeAmplifier_SenseWithNoiseClipsOnceAfterNoise(t *testing.T) {
    ca := DefaultChargeAmplifier()
    sigmaV := ca.InputChargeNoiseRMS * math.Sqrt(ca.Bandwidth) / ca.Cfb

    gotHigh := ca.senseWithGaussian(0.99*ca.MaxInputCharge, func() float64 { return 1000 })
    gotLow := ca.senseWithGaussian(-0.99*ca.MaxInputCharge, func() float64 { return -1000 })
    if gotHigh != ca.MaxOutputVoltage {
        t.Fatalf("high noisy output=%e want rail %e", gotHigh, ca.MaxOutputVoltage)
    }
    if gotLow != -ca.MaxOutputVoltage {
        t.Fatalf("low noisy output=%e want rail -%e", gotLow, ca.MaxOutputVoltage)
    }

    inwardSample := -0.02
    gotInward := ca.senseWithGaussian(1.01*ca.MaxInputCharge, func() float64 { return inwardSample })
    wantInward := 1.01*ca.MaxInputCharge/ca.Cfb + inwardSample*sigmaV
    if math.Abs(gotInward-wantInward) > 1e-15 {
        t.Fatalf("inward noisy output=%e want unclipped post-noise value %e", gotInward, wantInward)
    }
}
```

- [ ] **Step 2: Run RED**

Run:

```bash
go test -count=1 ./shared/peripherals -run 'TestChargeAmplifier_SenseWithNoise'
```

Expected: FAIL to compile because `senseWithGaussian` does not exist.

- [ ] **Step 3: Implement sampled, clipped noise**

Replace the single import with:

```go
import (
    "math"
    "math/rand"
)
```

Replace `SenseWithNoise` with:

```go
func (ca *ChargeAmplifier) SenseWithNoise(totalCharge float64) float64 {
    return ca.senseWithGaussian(totalCharge, rand.NormFloat64)
}

func (ca *ChargeAmplifier) senseWithGaussian(totalCharge float64, gaussian func() float64) float64 {
    if ca.Cfb <= 0 {
        return 0
    }
    vOut := totalCharge / ca.Cfb
    if gaussian != nil && ca.InputChargeNoiseRMS > 0 && ca.Bandwidth > 0 {
        sigmaQ := ca.InputChargeNoiseRMS * math.Sqrt(ca.Bandwidth)
        vOut += (sigmaQ / ca.Cfb) * gaussian()
    }
    if vOut > ca.MaxOutputVoltage {
        return ca.MaxOutputVoltage
    }
    if vOut < -ca.MaxOutputVoltage {
        return -ca.MaxOutputVoltage
    }
    return vOut
}
```

Update the method comment to state that the package-level source is used and that deterministic package tests inject the Gaussian sample through the private helper. Do not promise that global `rand.Seed` is the preferred reproducibility interface.

- [ ] **Step 4: Run GREEN and package regression**

```bash
go test -count=1 ./shared/peripherals -run 'TestChargeAmplifier'
go test -race -count=1 ./shared/peripherals
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add shared/peripherals/charge_amplifier.go shared/peripherals/charge_amplifier_test.go
git commit -m "fix(peripherals): sample and clip charge amplifier noise"
```

Record RED, GREEN, and race commands in the commit/PR evidence.

---

### Task 2: Bind reusable run caches to schema-v2 integrity metadata

**Files:**
- Modify: `workbench/experiment/types.go:54-65`
- Modify: `workbench/experiment/store.go:15-90,129-158`
- Modify: `workbench/experiment/runner.go:154-169`
- Modify: `workbench/experiment/runner_test.go:45-83,121-264`

**Interfaces:**
- Consumes: `ID(DesignPoint, string, []project.ResolvedInput)`, `RunRecord`, atomic directory commit.
- Produces: manifest schema 2 with artifact and record digests; recognizable `ErrUnverifiableRun`; load-time input-identity and artifact-integrity checks.

**Trust wording after completion:** Schema-2 caches detect accidental corruption and edits not accompanied by a complete digest rewrite. They are not signed and are not tamper-proof against an actor controlling all local files.  
**Rollback:** Archive schema-2 run directories before reverting. Never downgrade or rewrite them as schema 1; rerun evaluations under the reverted version if rollback is required.

**Compatibility decision:** Existing schema-1 run directories remain untouched but are not reused. The command identifies the offending run ID and tells the user to archive/remove only that run’s `runs/RUN_ID/` directory before rerunning. Automatic migration would assert integrity that the old format cannot prove.

- [ ] **Step 1: Add failing cache-corruption tests**

Add the test import `encoding/json` if not already present. Append to `workbench/experiment/runner_test.go`:

```go
func rewriteRunJSON(t *testing.T, path string, value any) {
    t.Helper()
    data, err := json.MarshalIndent(value, "", "  ")
    if err != nil {
        t.Fatal(err)
    }
    data = append(data, '\n')
    if err := os.WriteFile(path, data, 0o644); err != nil {
        t.Fatal(err)
    }
}

func TestRunRejectsTamperedCachedResult(t *testing.T) {
    bundle := runnerBundle(t, 1)
    opts := deterministicRunOptions(func(_ context.Context, _ project.Design, seed int64) (Result, error) {
        return successResult(float64(seed)), nil
    })
    first, err := Run(context.Background(), bundle, opts)
    if err != nil {
        t.Fatal(err)
    }

    record := first.Runs[0]
    record.Result.Metrics[0].Value = 999
    resultPath := filepath.Join(bundle.Root, "runs", record.Manifest.RunID, "result.json")
    rewriteRunJSON(t, resultPath, record.Result)

    if _, err := Run(context.Background(), bundle, opts); !errors.Is(err, ErrUnverifiableRun) {
        t.Fatalf("Run error=%v want ErrUnverifiableRun", err)
    }
}

func TestRunRejectsMalformedCachedArtifacts(t *testing.T) {
    for _, artifact := range []string{"manifest.json", "resolved-design.json", "result.json"} {
        t.Run(artifact, func(t *testing.T) {
            bundle := runnerBundle(t, 1)
            opts := deterministicRunOptions(func(_ context.Context, _ project.Design, seed int64) (Result, error) {
                return successResult(float64(seed)), nil
            })
            first, err := Run(context.Background(), bundle, opts)
            if err != nil {
                t.Fatal(err)
            }

            path := filepath.Join(bundle.Root, "runs", first.Runs[0].Manifest.RunID, artifact)
            if err := os.Truncate(path, 1); err != nil {
                t.Fatal(err)
            }
            if _, err := Run(context.Background(), bundle, opts); !errors.Is(err, ErrUnverifiableRun) {
                t.Fatalf("Run error=%v want ErrUnverifiableRun", err)
            }
        })
    }
}

func TestRunRejectsOversizedCachedArtifacts(t *testing.T) {
    for _, artifact := range []string{"manifest.json", "resolved-design.json", "result.json"} {
        t.Run(artifact, func(t *testing.T) {
            bundle := runnerBundle(t, 1)
            opts := deterministicRunOptions(func(_ context.Context, _ project.Design, seed int64) (Result, error) {
                return successResult(float64(seed)), nil
            })
            first, err := Run(context.Background(), bundle, opts)
            if err != nil {
                t.Fatal(err)
            }

            path := filepath.Join(bundle.Root, "runs", first.Runs[0].Manifest.RunID, artifact)
            if err := os.Truncate(path, maxRunArtifactBytes+1); err != nil {
                t.Fatal(err)
            }
            if _, err := Run(context.Background(), bundle, opts); !errors.Is(err, ErrUnverifiableRun) {
                t.Fatalf("Run error=%v want ErrUnverifiableRun", err)
            }
        })
    }
}

func TestRunRejectsCachedDesignThatDoesNotMatchRunID(t *testing.T) {
    bundle := runnerBundle(t, 1)
    opts := deterministicRunOptions(func(_ context.Context, _ project.Design, seed int64) (Result, error) {
        return successResult(float64(seed)), nil
    })
    first, err := Run(context.Background(), bundle, opts)
    if err != nil {
        t.Fatal(err)
    }

    record := first.Runs[0]
    record.Design.Array.Rows++
    designPath := filepath.Join(bundle.Root, "runs", record.Manifest.RunID, "resolved-design.json")
    rewriteRunJSON(t, designPath, record.Design)

    if _, err := Run(context.Background(), bundle, opts); !errors.Is(err, ErrUnverifiableRun) {
        t.Fatalf("Run error=%v want ErrUnverifiableRun", err)
    }
}

func TestRunRejectsTamperedCachedManifest(t *testing.T) {
    bundle := runnerBundle(t, 1)
    opts := deterministicRunOptions(func(_ context.Context, _ project.Design, seed int64) (Result, error) {
        return successResult(float64(seed)), nil
    })
    first, err := Run(context.Background(), bundle, opts)
    if err != nil {
        t.Fatal(err)
    }

    record := first.Runs[0]
    record.Manifest.RepositoryRevision = "tampered-revision"
    manifestPath := filepath.Join(bundle.Root, "runs", record.Manifest.RunID, "manifest.json")
    rewriteRunJSON(t, manifestPath, record.Manifest)
    if _, err := Run(context.Background(), bundle, opts); !errors.Is(err, ErrUnverifiableRun) {
        t.Fatalf("Run error=%v want ErrUnverifiableRun", err)
    }
}

func TestRunRejectsLegacySchemaOneCache(t *testing.T) {
    bundle := runnerBundle(t, 1)
    opts := deterministicRunOptions(func(_ context.Context, _ project.Design, seed int64) (Result, error) {
        return successResult(float64(seed)), nil
    })
    first, err := Run(context.Background(), bundle, opts)
    if err != nil {
        t.Fatal(err)
    }

    record := first.Runs[0]
    record.Manifest.SchemaVersion = 1
    record.Manifest.ArtifactSHA256 = nil
    record.Manifest.RecordSHA256 = ""
    manifestPath := filepath.Join(bundle.Root, "runs", record.Manifest.RunID, "manifest.json")
    rewriteRunJSON(t, manifestPath, record.Manifest)

    if _, err := Run(context.Background(), bundle, opts); !errors.Is(err, ErrUnverifiableRun) {
        t.Fatalf("Run error=%v want ErrUnverifiableRun", err)
    }
}
```

- [ ] **Step 2: Run RED**

```bash
go test -count=1 ./workbench/experiment -run 'TestRunRejects(Tampered|Malformed|Oversized|CachedDesign|Legacy)'
```

Expected: FAIL because `ErrUnverifiableRun` and `ArtifactSHA256` do not exist.

- [ ] **Step 3: Add schema-v2 manifest metadata**

In `workbench/experiment/types.go`, add to `RunManifest`:

```go
ArtifactSHA256 map[string]string `json:"artifact_sha256"`
RecordSHA256   string            `json:"record_sha256"`
```

In `workbench/experiment/store.go`, add imports `crypto/sha256` and `encoding/hex`, then define:

```go
var ErrUnverifiableRun = errors.New("experiment: unverifiable cached run")

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
```

Refactor `commit` so it canonicalizes `record.Design` and `record.Result`, sets:

```go
if record.Manifest.SchemaVersion != 2 {
    return RunRecord{}, fmt.Errorf("commit requires manifest schema 2, got %d", record.Manifest.SchemaVersion)
}
record.Manifest.ArtifactSHA256 = map[string]string{
    "resolved-design.json": sha256Hex(designBytes),
    "result.json":          sha256Hex(resultBytes),
}
recordDigest, err := recordSHA256(record.Manifest, designBytes, resultBytes)
if err != nil {
    return fmt.Errorf("digest run record: %w", err)
}
record.Manifest.RecordSHA256 = recordDigest
```

Then canonicalize the updated manifest and write the three prepared byte slices with a new exclusive helper:

```go
func writeExclusive(path string, data []byte) error {
    file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
    if err != nil {
        return err
    }
    if _, err := file.Write(data); err != nil {
        file.Close()
        return err
    }
    return file.Close()
}
```

Do not use a map iteration for the three writes; write `manifest.json`, `resolved-design.json`, and `result.json` explicitly so error messages and behavior are stable.

- [ ] **Step 4: Verify identity and digests during load**

In `store.load`, read each file as bytes with a bounded helper before decoding:

```go
const maxRunArtifactBytes int64 = 16 << 20

func readRunArtifact(path string) ([]byte, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()
    data, err := io.ReadAll(io.LimitReader(file, maxRunArtifactBytes+1))
    if err != nil {
        return nil, err
    }
    if int64(len(data)) > maxRunArtifactBytes {
        return nil, fmt.Errorf("%w: %s exceeds 16 MiB", ErrUnverifiableRun, filepath.Base(path))
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
```

Add `bytes` and `io` imports. Load `manifest.json`, `resolved-design.json`, and `result.json` through these helpers. Wrap every read/decode failure after the run directory is found with `fmt.Errorf("%w: read/decode manifest.json: %v", ErrUnverifiableRun, err)`, `fmt.Errorf("%w: read/decode resolved-design.json: %v", ErrUnverifiableRun, err)`, or `fmt.Errorf("%w: read/decode result.json: %v", ErrUnverifiableRun, err)` as applicable. This makes missing, truncated, malformed, and oversized cache artifacts recognizable without hiding the underlying message. After decoding the manifest:

```go
if record.Manifest.SchemaVersion != 2 {
    return RunRecord{}, false, fmt.Errorf("%w: run %s uses manifest schema %d; archive or remove runs/%s/ and rerun", ErrUnverifiableRun, id, record.Manifest.SchemaVersion, id)
}
```

After decoding design and result, verify both raw byte digests, then recompute the input identity:

```go
expectedID, err := ID(DesignPoint{
    Design: record.Design,
    Seed:   record.Manifest.Seed,
}, record.Manifest.EvaluatorVersion, record.Manifest.Inputs)
if err != nil {
    return RunRecord{}, false, fmt.Errorf("%w: recompute run %s identity: %v", ErrUnverifiableRun, id, err)
}
if expectedID != id || record.Manifest.RunID != id {
    return RunRecord{}, false, fmt.Errorf("%w: run %s identity mismatch", ErrUnverifiableRun, id)
}
if err := verifyArtifact("resolved-design.json", designBytes, record.Manifest); err != nil {
    return RunRecord{}, false, err
}
if err := verifyArtifact("result.json", resultBytes, record.Manifest); err != nil {
    return RunRecord{}, false, err
}
recordDigest, err := recordSHA256(record.Manifest, designBytes, resultBytes)
if err != nil || record.Manifest.RecordSHA256 == "" || recordDigest != record.Manifest.RecordSHA256 {
    return RunRecord{}, false, fmt.Errorf("%w: run %s record digest mismatch", ErrUnverifiableRun, id)
}
```

Retain the existing manifest/result status consistency check. This digest binds manifest provenance fields as well as the two artifact digests. It still does not provide signed protection against a same-user attacker who rewrites all three files.

- [ ] **Step 5: Make new runs schema 2 and run GREEN**

Change `SchemaVersion: 1` to `SchemaVersion: 2` in `evaluateJob`.

Run:

```bash
go test -count=1 ./workbench/experiment
go test -race -count=1 ./workbench/experiment
go test -count=1 ./workbench/...
```

Expected: PASS. Existing persistence tests must assert schema 2 and both digest keys.

- [ ] **Step 6: Commit**

```bash
git add workbench/experiment/types.go workbench/experiment/store.go \
  workbench/experiment/runner.go workbench/experiment/runner_test.go
git commit -m "fix(workbench): verify cached run artifacts before reuse"
```

Document the schema-1 cache invalidation in the commit body.

---

### Task 3: Bound and validate model binary serialization

**Files:**
- Modify: `module2-crossbar/pkg/weights/serialization.go:14-40,113-220,223-363`
- Modify: `module2-crossbar/pkg/weights/weights_test.go`

**Interfaces:**
- Consumes: existing `Model.SaveBinary(string) error` and `LoadModelBinary(string) (*Model, error)`.
- Produces: unchanged public signatures; rectangularity validation before writing; bounded decoding before every allocation.

**Trust wording after completion:** Binary model parsing is bounded against malformed/untrusted lengths and rejects non-rectangular weight matrices. It does not authenticate who created a valid file.  
**Rollback:** Revert the task commit. Format version remains 1, so valid files continue to round-trip; do not restore acceptance of a file that exceeded the documented limits without a separate reviewed limit change.

- [ ] **Step 1: Add failing ragged and oversized tests**

Append to `weights_test.go`:

```go
func TestSaveBinaryRejectsRaggedWeightsWithoutTouchingDestination(t *testing.T) {
    model := NewModel("ragged", "linear")
    model.AddLayer("fc1", "linear", [][]float64{{1, 2}, {3}}, nil)
    path := filepath.Join(t.TempDir(), "ragged.bin")
    if err := os.WriteFile(path, []byte("preserve-me"), 0o644); err != nil {
        t.Fatal(err)
    }

    err := model.SaveBinary(path)
    if err == nil || !strings.Contains(err.Error(), "rectangular") {
        t.Fatalf("SaveBinary error=%v want rectangularity error", err)
    }
    got, err := os.ReadFile(path)
    if err != nil {
        t.Fatal(err)
    }
    if string(got) != "preserve-me" {
        t.Fatalf("destination changed on validation failure: %q", got)
    }
}

func TestLoadModelBinaryRejectsOversizedMetadataLength(t *testing.T) {
    path := filepath.Join(t.TempDir(), "oversized.bin")
    file, err := os.Create(path)
    if err != nil {
        t.Fatal(err)
    }
    gz := gzip.NewWriter(file)
    _ = binary.Write(gz, binary.LittleEndian, uint32(0x4D4F444C))
    _ = binary.Write(gz, binary.LittleEndian, uint32(1))
    _ = binary.Write(gz, binary.LittleEndian, uint32((1<<20)+1))
    if err := gz.Close(); err != nil {
        t.Fatal(err)
    }
    if err := file.Close(); err != nil {
        t.Fatal(err)
    }

    if _, err := LoadModelBinary(path); err == nil || !strings.Contains(err.Error(), "metadata length") {
        t.Fatalf("LoadModelBinary error=%v want metadata length error", err)
    }
}

func TestLoadModelBinaryRejectsOversizedMatrixDimensions(t *testing.T) {
    var payload bytes.Buffer
    gz := gzip.NewWriter(&payload)
    _ = binary.Write(gz, binary.LittleEndian, uint32(0x4D4F444C))
    _ = binary.Write(gz, binary.LittleEndian, uint32(1))
    meta, _ := json.Marshal(ModelMetadata{Name: "oversized"})
    _ = binary.Write(gz, binary.LittleEndian, uint32(len(meta)))
    _, _ = gz.Write(meta)
    _ = binary.Write(gz, binary.LittleEndian, uint32(1))
    _ = binary.Write(gz, binary.LittleEndian, uint32(1))
    _, _ = gz.Write([]byte("x"))
    _ = binary.Write(gz, binary.LittleEndian, uint32(6))
    _, _ = gz.Write([]byte("linear"))
    _ = binary.Write(gz, binary.LittleEndian, uint32(0))
    _ = binary.Write(gz, binary.LittleEndian, uint32(16<<20))
    _ = binary.Write(gz, binary.LittleEndian, uint32(2))
    _ = gz.Close()

    path := filepath.Join(t.TempDir(), "oversized-matrix.bin")
    if err := os.WriteFile(path, payload.Bytes(), 0o644); err != nil {
        t.Fatal(err)
    }
    if _, err := LoadModelBinary(path); err == nil || !strings.Contains(err.Error(), "matrix dimensions") {
        t.Fatalf("LoadModelBinary error=%v want matrix dimensions error", err)
    }
}
```

Add required test imports: `bytes`, `compress/gzip`, `encoding/binary`, `encoding/json`, and `strings` if absent.

- [ ] **Step 2: Run RED**

```bash
go test -count=1 ./module2-crossbar/pkg/weights -run 'Test(SaveBinaryRejectsRagged|LoadModelBinaryRejectsOversized)'
```

Expected: FAIL for behavior: ragged rows are accepted/truncate the destination, and oversized lengths are not rejected with the required bounded-input errors.

- [ ] **Step 3: Add explicit format limits and layer validation**

Add after the imports in `serialization.go`:

```go
const (
    maxDecodedModelBytes   int64  = 512 << 20
    maxModelMetadataBytes  uint32 = 1 << 20
    maxModelLayers         uint32 = 4096
    maxModelStringBytes    uint32 = 64 << 10
    maxModelShapeDims      uint32 = 8
    maxModelMatrixElements uint32 = 16 << 20
    maxModelBiases         uint32 = 1 << 20
)

func validateSerializedLayer(layer *SerializedLayer) error {
    if len(layer.Name) > int(maxModelStringBytes) {
        return fmt.Errorf("layer name length %d exceeds limit", len(layer.Name))
    }
    if len(layer.Type) > int(maxModelStringBytes) {
        return fmt.Errorf("layer type length %d exceeds limit", len(layer.Type))
    }
    if len(layer.Shape) > int(maxModelShapeDims) {
        return fmt.Errorf("shape dimension count %d exceeds limit", len(layer.Shape))
    }
    for index, dim := range layer.Shape {
        if dim < 0 || uint64(dim) > uint64(maxModelMatrixElements) {
            return fmt.Errorf("shape dimension %d value %d is invalid", index, dim)
        }
    }
    rows := len(layer.Weights)
    cols := 0
    if rows > 0 {
        cols = len(layer.Weights[0])
    }
    if uint64(rows)*uint64(cols) > uint64(maxModelMatrixElements) {
        return fmt.Errorf("matrix dimensions %dx%d exceed limit", rows, cols)
    }
    for row := range layer.Weights {
        if len(layer.Weights[row]) != cols {
            return fmt.Errorf("weights must be rectangular: row %d has %d columns, want %d", row, len(layer.Weights[row]), cols)
        }
    }
    if len(layer.Biases) > int(maxModelBiases) {
        return fmt.Errorf("bias length %d exceeds limit", len(layer.Biases))
    }
    return nil
}

func checkedLength(label string, value, limit uint32) error {
    if value > limit {
        return fmt.Errorf("%s %d exceeds limit %d", label, value, limit)
    }
    return nil
}

func validateModelForBinary(model *Model) ([]byte, error) {
    if len(model.Layers) > int(maxModelLayers) {
        return nil, fmt.Errorf("layer count %d exceeds limit", len(model.Layers))
    }
    metadata, err := json.Marshal(model.Metadata)
    if err != nil {
        return nil, fmt.Errorf("marshal metadata: %w", err)
    }
    if len(metadata) > int(maxModelMetadataBytes) {
        return nil, fmt.Errorf("metadata length %d exceeds limit", len(metadata))
    }
    for index := range model.Layers {
        if err := validateSerializedLayer(&model.Layers[index]); err != nil {
            return nil, fmt.Errorf("layer %d: %w", index, err)
        }
    }
    return metadata, nil
}
```

At the first line of `SaveBinary`, call `validateModelForBinary` and return its error before `os.Create`. Reuse the returned metadata bytes instead of marshaling after opening the destination. Keep `writeSerializedLayerBinary` defensive by also calling `validateSerializedLayer` before it emits any layer bytes.

Remove the deferred `gzWriter.Close` and `file.Close`. After the last successful layer write, finalize explicitly so compression and filesystem errors reach the caller:

```go
if err := gzWriter.Close(); err != nil {
    _ = file.Close()
    return fmt.Errorf("finalize gzip stream: %w", err)
}
if err := file.Close(); err != nil {
    return fmt.Errorf("close model file: %w", err)
}
return nil
```

On every earlier write error, close both writers before returning the original error. The preflight guarantee applies to validation failures; later device/filesystem failures may leave an incomplete destination and must be reported.

- [ ] **Step 4: Bound every binary allocation**

In `LoadModelBinary`, wrap the gzip reader:

```go
limited := &io.LimitedReader{R: gzReader, N: maxDecodedModelBytes + 1}
```

Use `limited` for all subsequent reads. Before allocating metadata and layers:

```go
if err := checkedLength("metadata length", metaLen, maxModelMetadataBytes); err != nil {
    return nil, err
}
if err := checkedLength("layer count", numLayers, maxModelLayers); err != nil {
    return nil, err
}
```

In `readSerializedLayerBinary`, validate before each allocation:

```go
if err := checkedLength("layer name length", nameLen, maxModelStringBytes); err != nil { return nil, err }
if err := checkedLength("layer type length", typeLen, maxModelStringBytes); err != nil { return nil, err }
if err := checkedLength("shape dimension count", shapeDims, maxModelShapeDims); err != nil { return nil, err }
// Inside the shape loop, before converting uint32 to int:
if dim > maxModelMatrixElements { return nil, fmt.Errorf("shape dimension %d exceeds limit", dim) }
if rows > 0 {
    if cols == 0 || uint64(rows)*uint64(cols) > uint64(maxModelMatrixElements) {
        return nil, fmt.Errorf("matrix dimensions %dx%d exceed limit", rows, cols)
    }
}
if err := checkedLength("bias length", biasLen, maxModelBiases); err != nil { return nil, err }
```

After all layers are read, attempt one extra byte. Accept only `io.EOF`; reject trailing decoded data. Also reject `limited.N <= 0` as `decoded model exceeds 512 MiB limit`.

- [ ] **Step 5: Run GREEN, round-trip, and fuzz-adjacent malformed tests**

```bash
go test -count=1 ./module2-crossbar/pkg/weights -run 'Test(SaveBinaryRejectsRagged|LoadModelBinaryRejectsOversized)'
go test -count=1 ./module2-crossbar/pkg/weights
go test -race -count=1 ./module2-crossbar/pkg/weights
```

Expected: PASS. Existing JSON and binary round trips must remain green.

- [ ] **Step 6: Commit**

```bash
git add module2-crossbar/pkg/weights/serialization.go module2-crossbar/pkg/weights/weights_test.go
git commit -m "fix(weights): bound and validate model binaries"
```

---

### Task 4: Make Docker EDA argument construction non-interpreting

**Files:**
- Modify: `module6-eda/pkg/openlane/runner.go:51-81,162-192`
- Modify: `module6-eda/pkg/openlane/openlane_test.go:461-609`

**Interfaces:**
- Consumes: existing `Runner.RunOpenROAD` and `Runner.RunKLayout` public methods.
- Produces: pure `dockerOpenROADArgs` and `dockerKLayoutArgs` helpers; fixed shell programs whose dynamic data is supplied through positional arguments.

**Trust wording after completion:** Public script names and KLayout variables are not interpreted as shell program text in these two Docker paths. This does not make the container or mounted project a general security sandbox.  
**Rollback:** Revert the task commit only if the Docker integrations regress; immediately restore the roadmap trust blocker because the former interpolation behavior is unsafe.

- [ ] **Step 1: Add failing adversarial argument tests**

Append to `openlane_test.go`:

```go
func valueAfterArg(t *testing.T, args []string, token string) string {
    t.Helper()
    for i := 0; i+1 < len(args); i++ {
        if args[i] == token {
            return args[i+1]
        }
    }
    t.Fatalf("argument %q not found in %q", token, args)
    return ""
}

func TestDockerOpenROADArgsDoNotInterpolateScriptName(t *testing.T) {
    script := "/tmp/layout;touch PWNED.tcl"
    args := dockerOpenROADArgs("image@sha256:test", "/tmp/work", script, nil)
    shellProgram := valueAfterArg(t, args, "-c")
    if strings.Contains(shellProgram, filepath.Base(script)) || strings.Contains(shellProgram, "touch") {
        t.Fatalf("script name interpolated into shell program: %q", shellProgram)
    }
    if args[len(args)-1] != "/design/layout;touch PWNED.tcl" {
        t.Fatalf("script positional argument=%q", args[len(args)-1])
    }

    traversal := dockerOpenROADArgs("image@sha256:test", "/tmp/work", "../../escape.tcl", nil)
    if traversal[len(traversal)-1] != "/design/escape.tcl" {
        t.Fatalf("traversal was not confined to mounted design basename: %q", traversal[len(traversal)-1])
    }
}

func TestDockerKLayoutArgsKeepVariablesOutOfShellProgram(t *testing.T) {
    vars := map[string]string{"output_png": "/design/out;touch /design/PWNED.png"}
    args := dockerKLayoutArgs("image@sha256:test", "/tmp/work", "layout.rb", vars)
    shellProgram := valueAfterArg(t, args, "-c")
    if strings.Contains(shellProgram, "touch") || strings.Contains(shellProgram, vars["output_png"]) {
        t.Fatalf("dynamic value interpolated into shell program: %q", shellProgram)
    }
    joined := strings.Join(args, "\x00")
    if !strings.Contains(joined, "output_png="+vars["output_png"]) {
        t.Fatalf("dynamic value missing from argv: %q", args)
    }
}
```

Replace the existing `TestXvfbCommandConstruction` assertion that requires direct script interpolation. Its new assertion must obtain the `-c` value with `valueAfterArg`, require the fixed `Xvfb`/`exec openroad` text, and reject the literal script filename in that shell program. This prevents an obsolete test from preserving the vulnerability.

- [ ] **Step 2: Run RED**

```bash
go test -count=1 ./module6-eda/pkg/openlane -run 'Test(Docker(OpenROAD|KLayout)Args|XvfbCommandConstruction)'
```

Expected: FAIL because the pure helper functions do not exist.

- [ ] **Step 3: Implement fixed shell programs with positional arguments**

Add deterministic key ordering via `sort.Strings`. Implement:

```go
const xvfbPrefix = `Xvfb :99 -screen 0 1024x768x24 -nolisten tcp >/dev/null 2>&1 & sleep 1; export DISPLAY=:99; `

func dockerOpenROADArgs(image, absWorkDir, scriptName string, envVars map[string]string) []string {
    args := []string{"run", "--rm", "--entrypoint", "sh", "-v", absWorkDir + ":/design", "-w", "/design"}
    keys := make([]string, 0, len(envVars))
    for key := range envVars { keys = append(keys, key) }
    sort.Strings(keys)
    for _, key := range keys { args = append(args, "-e", key+"="+envVars[key]) }
    return append(args, image, "-c", xvfbPrefix+`exec openroad -no_splash -exit "$1"`, "fecim-openroad", "/design/"+filepath.Base(scriptName))
}

func dockerKLayoutArgs(image, absWorkDir, scriptPath string, envVars map[string]string) []string {
    args := []string{"run", "--rm", "--entrypoint", "sh", "-v", absWorkDir + ":/design", "-w", "/design", image, "-c", xvfbPrefix+`exec klayout "$@"`, "fecim-klayout", "-z"}
    keys := make([]string, 0, len(envVars))
    for key := range envVars { keys = append(keys, key) }
    sort.Strings(keys)
    for _, key := range keys { args = append(args, "-rd", key+"="+envVars[key]) }
    return append(args, "-r", "/design/"+filepath.Base(scriptPath))
}
```

Refactor `runDockerOpenROAD` and `runDockerKLayout` to compute `absWorkDir`, call the helper, and pass its result to `runWithTimeout`. No dynamic value may be inserted into `xvfbPrefix` or the fixed command suffix.

- [ ] **Step 4: Run GREEN and OpenLane regression**

```bash
go test -count=1 ./module6-eda/pkg/openlane -run 'Test(Docker(OpenROAD|KLayout)Args|XvfbCommandConstruction)'
go test -count=1 ./module6-eda/pkg/openlane
go vet ./module6-eda/pkg/openlane
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add module6-eda/pkg/openlane/runner.go module6-eda/pkg/openlane/openlane_test.go
git commit -m "fix(eda): pass docker tool inputs as arguments"
```

---

### Task 5: Replace no-op Verilog validation with a required external gate

**Files:**
- Modify: `validation/external/eda/verilog_sanity_test.go`
- Modify: `.github/workflows/ci.yml:65-117`
- Create: `scripts/ci_contract_test.go`

**Interfaces:**
- Consumes: `makeTestDesign`, `export.GenerateVerilogWithDefaults`, installed `iverilog` or `verilator`.
- Produces: a test that executes a validator; a required CI job that installs and runs iverilog; a repository test that prevents the lane from becoming optional/no-op again.

**Trust wording after completion:** The generated passive-array Verilog and explicit cell stub parse/elaborate under the recorded validator version. This is syntax/elaboration evidence, not synthesis, timing, physical-design, or hardware validation.  
**Rollback:** If CI availability forces a temporary revert, mark TR-VAL-01 open in the same change and remove any release claim that external Verilog validation is gating.

- [ ] **Step 1: Add a failing CI workflow contract**

Create `scripts/ci_contract_test.go` with complete content:

```go
package scripts_test

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
)

func readCIWorkflow(t *testing.T) string {
    t.Helper()
    data, err := os.ReadFile(filepath.Join(repoRoot(t), ".github", "workflows", "ci.yml"))
    if err != nil {
        t.Fatal(err)
    }
    return string(data)
}

func TestCIRequiresRealExternalVerilogValidation(t *testing.T) {
    workflow := readCIWorkflow(t)
    required := []string{
        "apt-get install -y libgl1-mesa-dev xorg-dev iverilog",
        "go test -v -count=1 ./validation/external/eda -run TestVerilogSanityCheck",
    }
    for _, token := range required {
        if !strings.Contains(workflow, token) {
            t.Errorf("ci.yml missing required external-validation token %q", token)
        }
    }
    externalStart := strings.Index(workflow, "external-validation:")
    if externalStart < 0 {
        t.Fatal("ci.yml missing external-validation job")
    }
    external := workflow[externalStart:]
    if strings.Contains(external, "continue-on-error: true") {
        t.Error("external-validation job must be gating")
    }

    sanityPath := filepath.Join(repoRoot(t), "validation", "external", "eda", "verilog_sanity_test.go")
    sanity := readFile(t, sanityPath)
    if strings.Contains(sanity, "would invoke") {
        t.Error("Verilog sanity test still contains no-op validation wording")
    }
    if !strings.Contains(sanity, `exec.Command("iverilog"`) {
        t.Error("Verilog sanity test must execute iverilog")
    }
}
```

The test reuses `repoRoot` from `scripts/check_architecture_test.go`; both files use package `scripts_test`.

- [ ] **Step 2: Run RED**

```bash
go test -count=1 ./scripts -run TestCIRequiresRealExternalVerilogValidation
```

Expected: FAIL because CI neither installs iverilog nor runs the exact required test as a gating step.

- [ ] **Step 3: Replace the no-op external test**

Replace `TestVerilogSanityCheck` with:

```go
func TestVerilogSanityCheck(t *testing.T) {
    design := makeTestDesign(4, 4, compiler.ArchPassive)
    dir := t.TempDir()
    verilogPath := filepath.Join(dir, "fecim_crossbar.v")
    cellPath := filepath.Join(dir, "fecim_cells.v")
    if err := os.WriteFile(verilogPath, []byte(export.GenerateVerilogWithDefaults(design)), 0o644); err != nil {
        t.Fatal(err)
    }
    const passiveCellStub = `module fecim_bit #(
    parameter LEVEL = 0
) (
    input wire WL,
    inout wire BL,
    inout wire VPWR,
    inout wire VGND
);
endmodule
`
    if err := os.WriteFile(cellPath, []byte(passiveCellStub), 0o644); err != nil {
        t.Fatal(err)
    }

    switch {
    case testsupport.HasCommand("iverilog"):
        cmd := exec.Command("iverilog", "-g2012", "-t", "null", "-o", filepath.Join(dir, "lint.out"), cellPath, verilogPath)
        output, err := cmd.CombinedOutput()
        if err != nil {
            t.Fatalf("iverilog validation failed: %v\n%s", err, output)
        }
        version, _ := exec.Command("iverilog", "-V").CombinedOutput()
        t.Logf("iverilog executed successfully: %s", strings.TrimSpace(string(version)))
    case testsupport.HasCommand("verilator"):
        cmd := exec.Command("verilator", "--lint-only", cellPath, verilogPath)
        output, err := cmd.CombinedOutput()
        if err != nil {
            t.Fatalf("verilator validation failed: %v\n%s", err, output)
        }
        version, _ := exec.Command("verilator", "--version").CombinedOutput()
        t.Logf("verilator executed successfully: %s", strings.TrimSpace(string(version)))
    default:
        t.Skip("iverilog or verilator is required for external Verilog validation")
    }
}
```

Use imports:

```go
import (
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "testing"

    "fecim-lattice-tools/module6-eda/pkg/compiler"
    "fecim-lattice-tools/module6-eda/pkg/export"
    "fecim-lattice-tools/validation/external/internal/testsupport"
)
```

This test must never log “would run.” It either executes a tool successfully, fails, or explicitly skips because the local tool is absent.

- [ ] **Step 4: Make external validation required in CI**

In `.github/workflows/ci.yml`:

- remove `continue-on-error: true` from `external-validation`;
- install `iverilog` alongside GUI dependencies:

```yaml
      - name: Install GUI and external validation dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y libgl1-mesa-dev xorg-dev iverilog
```

- replace conditional detection-only Verilog steps with:

```yaml
      - name: Record Icarus Verilog version
        run: iverilog -V

      - name: Run required external Verilog validation
        run: go test -v -count=1 ./validation/external/eda -run TestVerilogSanityCheck
```

Keep ngspice checks optional until a separate roadmap item installs and gates them. Do not describe the whole external job as optional after this change.

- [ ] **Step 5: Run GREEN and local validation**

```bash
go test -count=1 ./scripts -run TestCIRequiresRealExternalVerilogValidation
go test -v -count=1 ./validation/external/eda -run TestVerilogSanityCheck
go test -v -count=1 ./validation/external/eda
```

Expected locally: the first test passes; the Verilog test passes when iverilog/verilator is installed or explicitly skips otherwise. CI acceptance requires PASS with iverilog installed; a local skip is not release evidence.

- [ ] **Step 6: Commit**

```bash
git add .github/workflows/ci.yml scripts/ci_contract_test.go \
  validation/external/eda/verilog_sanity_test.go
git commit -m "test(eda): require real external Verilog validation"
```

---

### Task 6: Close Phase 1 with aggregate verification and roadmap receipts

**Files:**
- Modify: `TODO.md`
- Modify: `docs/internals/audits/2026-09-01-technical-recovery-roadmap.md`

**Interfaces:**
- Consumes: completion evidence from Tasks 1–5.
- Produces: honest status showing exactly which gates passed, skipped, or remain open.

- [ ] **Step 1: Run focused aggregate gates**

```bash
go test -v -count=1 ./shared/peripherals ./module2-crossbar/pkg/weights \
  ./workbench/experiment ./module6-eda/pkg/openlane ./validation/external/eda
go test -race -count=1 ./shared/peripherals ./workbench/experiment
go vet ./...
```

Expected: PASS. If the local Verilog validator is absent, record the skip and do not close TR-VAL-01 locally.

- [ ] **Step 2: Run repository regression**

```bash
go test ./...
```

Expected: PASS.

- [ ] **Step 3: Confirm required CI evidence**

After the branch CI run, attach the complete `Record Icarus Verilog version` step output and the successful `Run required external Verilog validation` step URL to the recovery item. Record the literal status lines `TestVerilogSanityCheck: PASS` and `external-validation job: PASS (gating)` in the roadmap receipt. Do not close Phase 1 before this evidence exists.

- [ ] **Step 4: Update roadmap status**

In `TODO.md`, mark only completed IDs done and link their commit/PR evidence. In the roadmap, change Phase 1 status from Active to Complete only when all five IDs and the aggregate gate pass. For each ID, record the exact RED and GREEN commands already specified in its task, the failing/passing test names, and the implementing commit hash. Under final verification, record the focused aggregate, race, vet, and full-suite commands from this task with their exit codes. For TR-VAL-01, also link the required CI job and copy its reported iverilog version. Do not summarize a skipped tool check as passing evidence.

- [ ] **Step 5: Commit documentation receipts**

```bash
git add TODO.md docs/internals/audits/2026-09-01-technical-recovery-roadmap.md
git commit -m "docs: record phase one recovery evidence"
```

---

## Final self-review checklist

Before implementation begins:

- [ ] Every task has one focused RED signal.
- [ ] Public signatures remain stable except the manifest schema, whose compatibility policy is explicit.
- [ ] Dynamic EDA values never enter shell program text.
- [ ] Binary limits are checked before allocation.
- [ ] Cache integrity wording does not claim signed tamper resistance.
- [ ] A local skipped external test cannot close the CI/release gate.
- [ ] No unrelated dirty files appear in any task commit.

## Completion definition

This plan is complete only when:

1. Tasks 1–5 have recorded RED/GREEN evidence;
2. focused, race, vet, and full-suite gates pass;
3. required CI executes iverilog and passes;
4. schema-1 cache behavior is documented and tested;
5. `TODO.md` and the recovery roadmap contain actual receipts rather than test-count claims.
