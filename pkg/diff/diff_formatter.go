package diff

import (
	"fmt"
	"strings"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

// FormatUnifiedDiff formats the drift as a unified diff (Git-style)
func (f *DiffFormatter) FormatUnifiedDiff(alert *types.DriftAlert) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("--- terraform/%s.%s\t(Terraform State)\n",
		alert.ResourceType, alert.ResourceName))
	b.WriteString(fmt.Sprintf("+++ runtime/%s.%s\t(Actual Configuration)\n",
		alert.ResourceType, alert.ResourceName))
	b.WriteString("@@ -1,1 +1,1 @@\n")

	oldStr := f.formatValue(alert.OldValue)
	newStr := f.formatValue(alert.NewValue)

	// Format as unified diff
	for _, line := range strings.Split(oldStr, "\n") {
		b.WriteString(f.color(ColorRed, fmt.Sprintf("-%s\n", line)))
	}
	for _, line := range strings.Split(newStr, "\n") {
		b.WriteString(f.color(ColorGreen, fmt.Sprintf("+%s\n", line)))
	}

	return b.String()
}

// FormatSideBySide formats the drift as side-by-side comparison
func (f *DiffFormatter) FormatSideBySide(alert *types.DriftAlert) string {
	var b strings.Builder

	oldStr := f.formatValue(alert.OldValue)
	newStr := f.formatValue(alert.NewValue)

	oldLines := strings.Split(oldStr, "\n")
	newLines := strings.Split(newStr, "\n")

	maxLines := len(oldLines)
	if len(newLines) > maxLines {
		maxLines = len(newLines)
	}

	// Header
	b.WriteString(fmt.Sprintf("%-40s | %s\n", "Terraform State", "Actual Configuration"))
	b.WriteString(strings.Repeat("─", 40) + "─┼─" + strings.Repeat("─", 40) + "\n")

	// Side-by-side comparison
	for i := 0; i < maxLines; i++ {
		oldLine := ""
		if i < len(oldLines) {
			oldLine = oldLines[i]
		}

		newLine := ""
		if i < len(newLines) {
			newLine = newLines[i]
		}

		if oldLine == newLine {
			b.WriteString(fmt.Sprintf("  %-38s │   %-38s\n", oldLine, newLine))
		} else {
			b.WriteString(fmt.Sprintf("%s │ %s\n",
				f.color(ColorRed, fmt.Sprintf("- %-38s", oldLine)),
				f.color(ColorGreen, fmt.Sprintf("+ %-38s", newLine)),
			))
		}
	}

	return b.String()
}
