// Package edaexport owns local-only educational EDA publication. It never installs tools.
package edaexport

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
	config "fecim-lattice-tools/module6-eda/pkg/config"
	export "fecim-lattice-tools/module6-eda/pkg/export"
	"fecim-lattice-tools/module6-eda/pkg/validation"
)

const (
	MaxConfigBytes         = 8 << 20
	MaxWeightsBytes        = 32 << 20
	MaxArtifactBytes int64 = 32 << 20
	MaxTotalBytes    int64 = 128 << 20
	MaxArtifacts           = 64
	MaxStagingBytes  int64 = 512 << 20
	CommandTimeout         = 10 * time.Minute
	JobTimeout             = 30 * time.Minute
	Warning                = "EDUCATIONAL: Placeholder cell Verilog; ESTIMATED_SCALAR characterization. Structural checks only, not extracted LVS, signoff or foundry-mapped silicon. No PDK mapping."
)

var identifier = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]{0,63}$`)
var exportID = regexp.MustCompile(`^[a-f0-9]{64}$`)
var artifactNames = map[string]bool{"design.spice": true, "design.v": true, "design.def": true, "cell.lef": true, "cell.lib": true, "cell.v": true, "config.json": true, "package.zip": true, "manifest.json": true}

type Builder struct {
	Config *compiler.ArrayConfig `json:"config"`
	Cell   config.CellConfig     `json:"cell"`
}
type Artifact struct {
	Name   string `json:"name"`
	Bytes  int64  `json:"bytes"`
	SHA256 string `json:"sha256"`
}
type Manifest struct {
	SchemaVersion int               `json:"schema_version"`
	Origin        string            `json:"origin"`
	RunID         string            `json:"run_id,omitempty"`
	Builder       *Builder          `json:"builder,omitempty"`
	Warning       string            `json:"warning"`
	FilesSHA256   map[string]string `json:"files_sha256"`
}
type Result struct {
	ID        string     `json:"id"`
	Origin    string     `json:"origin"`
	Warning   string     `json:"warning"`
	Artifacts []Artifact `json:"artifacts"`
}

func Defaults(mode compiler.OperationMode, rows, cols int, arch string) Builder {
	c := compiler.NewArrayConfig(mode, rows, cols)
	if arch == compiler.Arch1T1R {
		c.With1T1R()
	} else if arch == compiler.Arch2T1R {
		c.With2T1R()
	} else {
		c.Architecture = arch
	}
	cell := config.DefaultCellConfig()
	cell.CellType = arch
	cell.Width = c.CellPitch
	cell.Height = c.RowHeight
	return Builder{c, cell}
}
func DecodeBuilder(data []byte) (Builder, error) {
	var b Builder
	if len(data) > MaxConfigBytes {
		return b, errors.New("EDA config exceeds 8 MiB")
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(&b); err != nil {
		return b, err
	}
	if d.Decode(new(any)) != io.EOF {
		return b, errors.New("one JSON value required")
	}
	return b, b.Validate()
}
func (b Builder) Validate() error {
	c := b.Config
	if c == nil || c.ArrayRows < 1 || c.ArrayRows > 256 || c.ArrayCols < 1 || c.ArrayCols > 256 || c.ArrayRows > 65536/c.ArrayCols {
		return errors.New("freeform dimensions must be 1..256, at most 65536 cells")
	}
	if c.Mode < compiler.ModeStorage || c.Mode > compiler.ModeCompute {
		return errors.New("unsupported mode")
	}
	if c.Architecture != compiler.ArchPassive && c.Architecture != compiler.Arch1T1R && c.Architecture != compiler.Arch2T1R {
		return errors.New("unsupported architecture")
	}
	if c.Technology != compiler.TechSKY130 || b.Cell.Technology != "sky130" {
		return errors.New("initial web profile supports SKY130 educational defaults only; no PDK mapping")
	}
	if !identifier.MatchString(c.Name) || !identifier.MatchString(b.Cell.Name) || b.Cell.CellType != c.Architecture {
		return errors.New("invalid identifiers or mismatched cell architecture")
	}
	// Recursively bound every scalar before allocation/formatting. JSON already excludes NaN/Inf;
	// direct native callers must meet the same rule. Strings are enumerated below.
	if err := boundedScalars(reflect.ValueOf(b)); err != nil {
		return err
	}
	if c.Levels < 2 || c.Levels > 256 || c.GMin <= 0 || c.GMax <= c.GMin || c.CellPitch <= 0 || c.RowHeight <= 0 || c.TPulse <= 0 || c.Peripherals.VDD <= 0 || c.Peripherals.ADCBits < 1 || c.Peripherals.ADCBits > 32 || c.Peripherals.DACBits < 1 || c.Peripherals.DACBits > 32 {
		return errors.New("invalid electrical/geometry/quantization bounds")
	}
	if b.Cell.Width <= 0 || b.Cell.Height <= 0 || b.Cell.Voltage <= 0 || b.Cell.Process <= 0 || b.Cell.MetalPitch <= 0 || b.Cell.MetalWidth <= 0 {
		return errors.New("invalid cell geometry/conditions")
	}
	if c.Mode == compiler.ModeStorage {
		if c.StorageConfig == nil || c.MemoryConfig != nil || c.ComputeConfig != nil {
			return errors.New("storage config required exclusively")
		}
		e := c.StorageConfig.ErrorCorrection
		if e != "none" && e != "SECDED" && e != "BCH" {
			return errors.New("unsupported ECC")
		}
	}
	if c.Mode == compiler.ModeMemory && (c.MemoryConfig == nil || c.ComputeConfig != nil || c.StorageConfig != nil) {
		return errors.New("memory config required exclusively")
	}
	if c.Mode == compiler.ModeCompute {
		if c.ComputeConfig == nil || c.MemoryConfig != nil || c.StorageConfig != nil {
			return errors.New("compute config required exclusively")
		}
		a := c.ComputeConfig.ActivationFunc
		if a != "none" && a != "relu" && a != "sigmoid" {
			return errors.New("unsupported activation")
		}
		if c.ComputeConfig.QuantLevels < 2 || c.ComputeConfig.QuantLevels > 256 || c.ComputeConfig.AccumulatorBits < 1 || c.ComputeConfig.AccumulatorBits > 64 {
			return errors.New("invalid compute bit/level bounds")
		}
		if w := c.ComputeConfig.InitialWeights; w != nil {
			if len(w) != c.ArrayRows {
				return errors.New("weights must exactly match rows")
			}
			for _, row := range w {
				if len(row) != c.ArrayCols {
					return errors.New("weights must exactly match columns")
				}
			}
		}
	}
	return nil
}
func boundedScalars(v reflect.Value) error {
	switch v.Kind() {
	case reflect.Pointer:
		if !v.IsNil() {
			return boundedScalars(v.Elem())
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if e := boundedScalars(v.Field(i)); e != nil {
				return e
			}
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			if e := boundedScalars(v.Index(i)); e != nil {
				return e
			}
		}
	case reflect.Float64:
		x := v.Float()
		if x != x || x > 1e12 || x < -1e12 {
			return errors.New("nonfinite or excessive scalar")
		}
	}
	return nil
}

// CheckArtifacts admits only enumerated flat server-owned names, counts and byte budgets.
func CheckArtifacts(sizes map[string]int64) error {
	if len(sizes) > MaxArtifacts {
		return errors.New("more than 64 artifacts")
	}
	var total int64
	for name, n := range sizes {
		if !artifactNames[name] || n < 0 || n > MaxArtifactBytes {
			return errors.New("artifact not enumerated or exceeds 32 MiB")
		}
		total += n
		if total > MaxTotalBytes {
			return errors.New("artifacts exceed 128 MiB")
		}
	}
	return nil
}
func AllowedArtifact(name string) bool { return artifactNames[name] }

func GenerateBuilderRoot(ctx context.Context, root *os.Root, id string, b Builder) (Result, error) {
	if !exportID.MatchString(id) {
		return Result{}, errors.New("invalid export ID")
	}
	if err := b.Validate(); err != nil {
		return Result{}, err
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	design, err := compiler.GenerateDesign(b.Config)
	if err != nil {
		return Result{}, err
	}
	cellName := b.Cell.Name
	if cellName == "fecim_bitcell" {
		if b.Config.Architecture == compiler.Arch1T1R {
			cellName = "fecim_1t1r_bitcell"
		}
		if b.Config.Architecture == compiler.Arch2T1R {
			cellName = "fecim_2t1r_bitcell"
		}
	}
	vc := export.VerilogConfig{ModuleName: b.Config.Name, CellName: cellName}
	dc := export.DEFConfigFrom(design)
	dc.CellName = cellName
	files := map[string][]byte{}
	sizes := map[string]int64{}
	add := func(name, s string) error {
		files[name] = []byte(s)
		sizes[name] = int64(len(s))
		return CheckArtifacts(sizes)
	}
	generators := []struct {
		name string
		f    func() string
	}{
		{"design.spice", func() string { return export.GenerateSPICE(design, b.Config.Peripherals.VDD) }},
		{"design.v", func() string { return export.GenerateVerilog(design, vc) }},
		{"design.def", func() string { return export.GenerateDEF(design, dc) }},
		{"cell.v", func() string { return export.GenerateCellVerilogWithLevel(b.Cell) }},
		{"cell.lef", func() string { return export.GenerateLEF(b.Cell) }},
		{"cell.lib", func() string { return export.GenerateLiberty(b.Cell) }},
	}
	for _, g := range generators {
		if err = ctx.Err(); err != nil {
			return Result{}, err
		}
		if err = add(g.name, g.f()); err != nil {
			return Result{}, err
		}
	}
	if err = validateContents(files); err != nil {
		return Result{}, err
	}
	resolved, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return Result{}, err
	}
	if err = add("config.json", string(resolved)); err != nil {
		return Result{}, err
	}
	// Package only enumerated generated files; never scan arbitrary project entries.
	var pack limitedBuffer
	pack.limit = MaxArtifactBytes
	zw := zip.NewWriter(&pack)
	names := sortedNames(files)
	for _, name := range names {
		w, e := zw.Create(name)
		if e != nil {
			return Result{}, e
		}
		if _, e = w.Write(files[name]); e != nil {
			return Result{}, e
		}
	}
	if err = zw.Close(); err != nil {
		return Result{}, err
	}
	if err = add("package.zip", pack.String()); err != nil {
		return Result{}, err
	}
	hashes := map[string]string{}
	for n, data := range files {
		h := sha256.Sum256(data)
		hashes[n] = hex.EncodeToString(h[:])
	}
	m := Manifest{SchemaVersion: 1, Origin: "freeform", Builder: &b, Warning: Warning, FilesSHA256: hashes}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return Result{}, err
	}
	if err = add("manifest.json", string(data)+"\n"); err != nil {
		return Result{}, err
	}
	if err = root.MkdirAll("exports", 0755); err != nil {
		return Result{}, err
	}
	parent, err := root.OpenRoot("exports")
	if err != nil {
		return Result{}, err
	}
	defer parent.Close()
	if err = parent.Mkdir(id, 0755); err != nil {
		return Result{}, err
	}
	dir, err := parent.OpenRoot(id)
	if err != nil {
		return Result{}, err
	}
	defer dir.Close()
	created := map[string]os.FileInfo{}
	complete := false
	defer func() {
		if complete {
			return
		}
		for n, owned := range created {
			entry, e := dir.Lstat(n)
			if e == nil && os.SameFile(owned, entry) {
				_ = dir.Remove(n)
			}
		}
		owned, e := dir.Stat(".")
		entry, e2 := parent.Lstat(id)
		if e == nil && e2 == nil && os.SameFile(owned, entry) {
			_ = parent.Remove(id)
		}
	}()
	names = sortedNames(files)
	for _, name := range names {
		if name == "manifest.json" {
			continue
		}
		if err = ctx.Err(); err != nil {
			return Result{}, err
		}
		info, e := writeExportFile(dir, name, files[name])
		if info != nil {
			created[name] = info
		}
		if e != nil {
			return Result{}, e
		}
	}
	if err = ctx.Err(); err != nil {
		return Result{}, err
	}
	info, err := writeExportFile(dir, "manifest.json", files["manifest.json"])
	if info != nil {
		created["manifest.json"] = info
	}
	if err != nil {
		return Result{}, err
	}
	complete = true
	return InspectRoot(root, id)
}
func sortedNames[T any](m map[string]T) []string {
	names := make([]string, 0, len(m))
	for n := range m {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

type limitedBuffer struct {
	bytes.Buffer
	limit int64
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if int64(b.Len())+int64(len(p)) > b.limit {
		return 0, errors.New("serialization exceeds artifact budget")
	}
	return b.Buffer.Write(p)
}

func readRegular(root *os.Root, name string, limit int64) ([]byte, error) {
	info, err := root.Lstat(name)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > limit {
		return nil, errors.New("artifact is not a bounded regular file")
	}
	f, err := root.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	after, err := f.Stat()
	if err != nil || !os.SameFile(info, after) {
		return nil, errors.New("artifact changed")
	}
	data, err := io.ReadAll(io.LimitReader(f, limit+1))
	if int64(len(data)) > limit {
		return nil, errors.New("artifact exceeds bound")
	}
	return data, err
}
func InspectRoot(root *os.Root, id string) (Result, error) {
	if !exportID.MatchString(id) {
		return Result{}, errors.New("invalid export ID")
	}
	dir, err := root.OpenRoot(filepath.Join("exports", id))
	if err != nil {
		return Result{}, err
	}
	defer dir.Close()
	raw, err := readRegular(dir, "manifest.json", MaxConfigBytes)
	if err != nil {
		return Result{}, err
	}
	var m Manifest
	if err = json.Unmarshal(raw, &m); err != nil {
		return Result{}, err
	}
	if m.SchemaVersion != 1 || len(m.FilesSHA256) == 0 {
		return Result{}, errors.New("invalid manifest")
	}
	origin := m.Origin
	if origin == "" && m.RunID == id {
		origin = "selected-run"
	}
	if origin != "freeform" && origin != "selected-run" {
		return Result{}, errors.New("invalid origin")
	}
	required := []string{"design.spice", "design.v", "design.def", "cell.lef", "cell.lib"}
	if origin == "freeform" {
		required = append(required, "cell.v", "config.json", "package.zip")
	}
	for _, name := range required {
		if m.FilesSHA256[name] == "" {
			return Result{}, errors.New("completion manifest omits required artifact")
		}
	}
	out := Result{ID: id, Origin: origin, Warning: Warning + " " + m.Warning, Artifacts: []Artifact{}}
	sizes := map[string]int64{"manifest.json": int64(len(raw))}
	for _, name := range sortedNames(m.FilesSHA256) {
		if name == "manifest.json" || !artifactNames[name] {
			return Result{}, errors.New("invalid manifest artifact")
		}
		data, e := readRegular(dir, name, MaxArtifactBytes)
		if e != nil {
			return Result{}, e
		}
		h := sha256.Sum256(data)
		hash := hex.EncodeToString(h[:])
		if hash != m.FilesSHA256[name] {
			return Result{}, fmt.Errorf("artifact integrity failed: %s", name)
		}
		sizes[name] = int64(len(data))
		if e = CheckArtifacts(sizes); e != nil {
			return Result{}, e
		}
		out.Artifacts = append(out.Artifacts, Artifact{name, int64(len(data)), hash})
	}
	h := sha256.Sum256(raw)
	out.Artifacts = append(out.Artifacts, Artifact{"manifest.json", int64(len(raw)), hex.EncodeToString(h[:])})
	return out, nil
}
func ReadArtifactRoot(root *os.Root, id, name string) ([]byte, error) {
	if !artifactNames[name] {
		return nil, errors.New("artifact not enumerated")
	}
	r, err := InspectRoot(root, id)
	if err != nil {
		return nil, err
	}
	found := false
	for _, a := range r.Artifacts {
		found = found || a.Name == name
	}
	if !found {
		return nil, errors.New("artifact absent from manifest")
	}
	dir, err := root.OpenRoot(filepath.Join("exports", id))
	if err != nil {
		return nil, err
	}
	defer dir.Close()
	data, err := readRegular(dir, name, MaxArtifactBytes)
	if err != nil {
		return nil, err
	}
	h := sha256.Sum256(data)
	for _, a := range r.Artifacts {
		if a.Name == name && a.SHA256 != hex.EncodeToString(h[:]) {
			return nil, errors.New("artifact changed during read")
		}
	}
	return data, nil
}
func ValidateRoot(root *os.Root, id string) error {
	r, err := InspectRoot(root, id)
	if err != nil {
		return err
	}
	files := map[string][]byte{}
	for _, a := range r.Artifacts {
		if a.Name == "design.def" || a.Name == "cell.lef" || a.Name == "cell.lib" || a.Name == "cell.v" {
			files[a.Name], err = ReadArtifactRoot(root, id, a.Name)
			if err != nil {
				return err
			}
		}
	}
	return validateContents(files)
}
func validateContents(files map[string][]byte) error {
	if err := validation.ValidateDEFReader(bytes.NewReader(files["design.def"])); err != nil {
		return err
	}
	// Selected-run's historical bundle has no cell.v; do not invent a cross-file pass.
	if files["cell.v"] == nil {
		return nil
	}
	// The legacy parser includes comment text. Strip comments just as its native
	// bundle oracle does; otherwise a commented example module masks the real cell.
	lines := strings.Split(string(files["cell.v"]), "\n")
	for i, l := range lines {
		lines[i] = strings.SplitN(l, "//", 2)[0]
	}
	return validation.CrossCheckReaders(bytes.NewReader(files["cell.lef"]), bytes.NewReader(files["cell.lib"]), strings.NewReader(strings.Join(lines, "\n")))
}
