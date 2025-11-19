package diff

// DiffFormatter formats drift differences in various output formats
type DiffFormatter struct {
	colorEnabled bool
}

// NewFormatter creates a new diff formatter
func NewFormatter(colorEnabled bool) *DiffFormatter {
	return &DiffFormatter{
		colorEnabled: colorEnabled,
	}
}

// ANSI color codes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorGray   = "\033[37m"
	ColorBold   = "\033[1m"
)
