package docs

import (
	"regexp"
	"sort"
	"strings"
)

// Canonical names from the native glossary; definitions come from the curated
// glossary document, not the legacy GUI's unsupported quantitative claims.
var glossaryNames = []string{"FeCIM", "Ec", "Pr", "HZO", "Hysteresis Loop", "Preisach Model", "Endurance", "1T1R", "MVM", "MAC", "BEOL", "Sneak Path", "IR Drop", "DAC", "ADC", "TIA", "Sense Amplifier", "TRL", "TOPS/W", "Bits per Cell", "MNIST Accuracy", "Retention Time", "Write Energy"}

func DetectGlossaryTerms(content string) []string {
	// Convert content to lowercase for case-insensitive matching
	lowerContent := strings.ToLower(content)

	// Find unique terms that appear as whole words
	foundTerms := make(map[string]bool)

	for _, term := range glossaryNames {
		termLower := strings.ToLower(term)
		// Create regex for whole-word match
		pattern := `\b` + regexp.QuoteMeta(termLower) + `\b`
		matched, err := regexp.MatchString(pattern, lowerContent)
		if err == nil && matched {
			// Store with original casing
			foundTerms[term] = true
		}
	}

	// Common single-word aliases that should map to canonical glossary entries.
	// This improves discoverability when docs mention generic forms
	// ("ferroelectric", "hysteresis", "polarization") instead of full term names.
	aliasToCanonical := map[string]string{
		"ferroelectric": "FeCIM",
		"hysteresis":    "Hysteresis Loop",
		"polarization":  "Remnant Polarization",
	}
	for alias, canonical := range aliasToCanonical {
		pattern := `\b` + regexp.QuoteMeta(alias) + `\b`
		matched, err := regexp.MatchString(pattern, lowerContent)
		if err == nil && matched {
			foundTerms[canonical] = true
		}
	}

	// Convert to sorted slice
	terms := make([]string, 0, len(foundTerms))
	for term := range foundTerms {
		terms = append(terms, term)
	}
	sort.Strings(terms)

	return terms
}
