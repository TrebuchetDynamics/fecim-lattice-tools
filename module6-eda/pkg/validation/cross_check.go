// pkg/validation/cross_check.go
package validation

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"fecim-lattice-tools/shared/logging"
)

var logCrossCheck = logging.NewLogger("eda-validation-crosscheck")

// CrossCheckFiles performs cross-file consistency checks between LEF, LIB, and Verilog
// Verifies that pin names and cell names match across all three files
func CrossCheckFiles(lefPath, libPath, verilogPath string) error {
	logCrossCheck.Input("CrossCheckFiles", map[string]interface{}{
		"lefPath": lefPath, "libPath": libPath, "verilogPath": verilogPath,
	})

	// Extract data from each file
	lefPins, lefCellName, err := extractLEFData(lefPath)
	if err != nil {
		return fmt.Errorf("LEF parsing error: %v", err)
	}
	
	libPins, libCellName, err := extractLibData(libPath)
	if err != nil {
		return fmt.Errorf("Liberty parsing error: %v", err)
	}
	
	verilogPins, verilogCellName, err := extractVerilogData(verilogPath)
	if err != nil {
		return fmt.Errorf("Verilog parsing error: %v", err)
	}
	
	// Check cell name consistency
	if lefCellName != libCellName || lefCellName != verilogCellName {
		return fmt.Errorf("cell name mismatch: LEF=%s, LIB=%s, Verilog=%s", 
			lefCellName, libCellName, verilogCellName)
	}
	
	// Check pin consistency
	if !slicesEqual(lefPins, libPins) {
		return fmt.Errorf("pin mismatch between LEF and LIB:\nLEF: %v\nLIB: %v", lefPins, libPins)
	}
	
	if !slicesEqual(lefPins, verilogPins) {
		return fmt.Errorf("pin mismatch between LEF and Verilog:\nLEF: %v\nVerilog: %v", lefPins, verilogPins)
	}
	
	return nil
}

// extractLEFData extracts cell name and pin names from LEF file
func extractLEFData(path string) ([]string, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()
	
	var pins []string
	var cellName string
	scanner := bufio.NewScanner(file)
	
	macroRe := regexp.MustCompile(`MACRO\s+(\w+)`)
	pinRe := regexp.MustCompile(`PIN\s+(\w+)`)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		if matches := macroRe.FindStringSubmatch(line); matches != nil {
			cellName = matches[1]
		}
		
		if matches := pinRe.FindStringSubmatch(line); matches != nil {
			pins = append(pins, matches[1])
		}
	}
	
	return pins, cellName, nil
}

// extractLibData extracts cell name and pin names from Liberty file
func extractLibData(path string) ([]string, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()
	
	var pins []string
	var cellName string
	scanner := bufio.NewScanner(file)
	
	cellRe := regexp.MustCompile(`cell\((\w+)\)`)
	pinRe := regexp.MustCompile(`pin\((\w+)\)`)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		if matches := cellRe.FindStringSubmatch(line); matches != nil {
			cellName = matches[1]
		}
		
		if matches := pinRe.FindStringSubmatch(line); matches != nil {
			pins = append(pins, matches[1])
		}
	}
	
	return pins, cellName, nil
}

// extractVerilogData extracts module name and port names from Verilog file
func extractVerilogData(path string) ([]string, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()
	
	var pins []string
	var moduleName string
	scanner := bufio.NewScanner(file)
	
	moduleRe := regexp.MustCompile(`module\s+(\w+)`)
	portRe := regexp.MustCompile(`(input|output|inout)\s+\w+\s+(\w+)`)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		if matches := moduleRe.FindStringSubmatch(line); matches != nil {
			moduleName = matches[1]
		}
		
		if matches := portRe.FindStringSubmatch(line); matches != nil {
			pins = append(pins, matches[2])
		}
	}
	
	return pins, moduleName, nil
}

// slicesEqual checks if two string slices contain the same elements (order-independent)
func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	
	aMap := make(map[string]bool)
	for _, s := range a {
		aMap[s] = true
	}
	
	for _, s := range b {
		if !aMap[s] {
			return false
		}
	}
	
	return true
}
