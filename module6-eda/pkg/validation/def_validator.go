// pkg/validation/def_validator.go
package validation

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// ValidateDEF performs basic DEF syntax validation
// Checks for required keywords, component count, and basic structure
func ValidateDEF(defPath string) error {
	// Check if file exists
	file, err := os.Open(defPath)
	if err != nil {
		return fmt.Errorf("cannot open DEF file: %v", err)
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading DEF file: %v", err)
	}
	
	content := strings.Join(lines, "\n")
	
	// Check for required keywords
	requiredKeywords := []string{"VERSION", "DESIGN", "UNITS", "DIEAREA", "COMPONENTS", "END COMPONENTS"}
	for _, keyword := range requiredKeywords {
		if !strings.Contains(content, keyword) {
			return fmt.Errorf("missing required keyword: %s", keyword)
		}
	}
	
	// Validate COMPONENTS count
	componentsLine := ""
	endComponentsFound := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "COMPONENTS") {
			componentsLine = trimmed
		}
		if strings.HasPrefix(trimmed, "END COMPONENTS") {
			endComponentsFound = true
		}
	}
	
	if componentsLine == "" {
		return fmt.Errorf("COMPONENTS declaration not found")
	}
	if !endComponentsFound {
		return fmt.Errorf("END COMPONENTS not found")
	}
	
	// Extract declared component count
	re := regexp.MustCompile(`COMPONENTS\s+(\d+)`)
	matches := re.FindStringSubmatch(componentsLine)
	if len(matches) < 2 {
		return fmt.Errorf("invalid COMPONENTS declaration format")
	}
	
	declaredCount, _ := strconv.Atoi(matches[1])
	
	// Count actual component instances (lines starting with "  -")
	actualCount := 0
	inComponents := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "COMPONENTS") {
			inComponents = true
			continue
		}
		if strings.HasPrefix(trimmed, "END COMPONENTS") {
			inComponents = false
			break
		}
		if inComponents && strings.HasPrefix(trimmed, "-") {
			actualCount++
		}
	}
	
	if declaredCount != actualCount {
		return fmt.Errorf("component count mismatch: declared %d, found %d instances", declaredCount, actualCount)
	}
	
	return nil
}

// GetDEFStats returns basic statistics from a DEF file
func GetDEFStats(defPath string) (map[string]interface{}, error) {
	file, err := os.Open(defPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	stats := make(map[string]interface{})
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Extract design name
		if strings.HasPrefix(line, "DESIGN") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				stats["design_name"] = parts[1]
			}
		}
		
		// Extract component count
		if strings.HasPrefix(line, "COMPONENTS") {
			re := regexp.MustCompile(`COMPONENTS\s+(\d+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) >= 2 {
				count, _ := strconv.Atoi(matches[1])
				stats["component_count"] = count
			}
		}
		
		// Extract die area
		if strings.HasPrefix(line, "DIEAREA") {
			stats["die_area"] = line
		}
	}
	
	return stats, nil
}
