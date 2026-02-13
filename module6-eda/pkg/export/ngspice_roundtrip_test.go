package export

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestNgspiceRoundTrip(t *testing.T) {
	mapping := getTestMapping()
	netlist := GenerateSPICE(mapping, 1.8)

	if err := validateBasicSpiceStructure(netlist); err != nil {
		t.Fatalf("generated netlist structure invalid: %v", err)
	}

	tmpDir := t.TempDir()
	netlistPath := filepath.Join(tmpDir, "fecim_roundtrip.sp")
	if err := os.WriteFile(netlistPath, []byte(netlist), 0o644); err != nil {
		t.Fatalf("write netlist: %v", err)
	}

	if _, err := exec.LookPath("ngspice"); err != nil {
		t.Log("ngspice not installed; performed structural syntax validation only")
		return
	}

	cmd := exec.Command("ngspice", "-b", netlistPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ngspice batch run failed: %v\n%s", err, string(out))
	}
	output := string(out)
	if strings.Contains(strings.ToLower(output), "error") {
		t.Fatalf("ngspice reported error output:\n%s", output)
	}
	if !strings.Contains(output, "No. of Data Rows") && !strings.Contains(strings.ToLower(output), "ngspice") {
		t.Fatalf("unexpected ngspice output sanity check failed:\n%s", output)
	}
	t.Logf("ngspice round-trip sanity output:\n%s", output)
}

func validateBasicSpiceStructure(netlist string) error {
	trimmed := strings.TrimSpace(netlist)
	if !strings.HasPrefix(trimmed, "*") {
		return errString("missing header comment")
	}
	if !strings.Contains(trimmed, ".param VDD") {
		return errString("missing VDD parameter")
	}
	if !strings.Contains(trimmed, "FeFET Cells") {
		return errString("missing FeFET section")
	}
	if !strings.HasSuffix(trimmed, ".end") {
		return errString("missing .end terminator")
	}
	return nil
}

type errString string

func (e errString) Error() string { return string(e) }
