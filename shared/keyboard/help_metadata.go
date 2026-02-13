package keyboard

import "strings"

// ShortcutMetadata represents one user-visible keyboard binding.
type ShortcutMetadata struct {
	Key         string
	Description string
}

// ShortcutSection groups shortcuts by functional area.
type ShortcutSection struct {
	Title     string
	Shortcuts []ShortcutMetadata
}

// HelpMetadata is a structured model for module keyboard help.
type HelpMetadata struct {
	Sections []ShortcutSection
	Tips     []string
}

// FormatHelpMetadata converts structured help metadata into display text.
func FormatHelpMetadata(meta HelpMetadata) string {
	var b strings.Builder
	b.WriteString("Keyboard Shortcuts:\n\n")

	for _, section := range meta.Sections {
		if section.Title == "" || len(section.Shortcuts) == 0 {
			continue
		}
		b.WriteString(section.Title)
		b.WriteString(":\n")
		for _, sc := range section.Shortcuts {
			if sc.Key == "" || sc.Description == "" {
				continue
			}
			b.WriteString("  ")
			b.WriteString(sc.Key)
			b.WriteString("    ")
			b.WriteString(sc.Description)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	if len(meta.Tips) > 0 {
		b.WriteString("Tips:\n")
		for _, tip := range meta.Tips {
			if tip == "" {
				continue
			}
			b.WriteString("• ")
			b.WriteString(tip)
			b.WriteString("\n")
		}
	}

	return strings.TrimSpace(b.String())
}
