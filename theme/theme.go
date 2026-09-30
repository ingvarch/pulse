package theme

import "charm.land/lipgloss/v2"

// Percentage value threshold colors.
const (
	ColorOK   = "#9ece6a"
	ColorWarn = "#e0af68"
	ColorCrit = "#f7768e"
	ColorAxis = "#565f89"
)

// ValueColor returns a color based on thresholds: <60 OK, <85 warn, otherwise crit.
func ValueColor(v float64) string {
	switch {
	case v < 60:
		return ColorOK
	case v < 85:
		return ColorWarn
	default:
		return ColorCrit
	}
}

// SeriesPalette defines series colors in Tokyo Night palette order.
var SeriesPalette = []string{
	"#9ece6a",
	"#7aa2f7",
	"#bb9af7",
	"#7dcfff",
	"#e0af68",
	"#f7768e",
}

// SeriesColor returns the color for series i, cycling through the palette.
func SeriesColor(i int) string {
	if len(SeriesPalette) == 0 {
		return ""
	}
	if i < 0 {
		i = 0
	}
	return SeriesPalette[i%len(SeriesPalette)]
}

// Preset holds chart styling options.
type Preset struct {
	Axis lipgloss.Style
}

// LineFor returns line style for the current value based on thresholds.
func (p Preset) LineFor(v float64) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(ValueColor(v)))
}

// TokyoNight returns a preset in Tokyo Night palette.
func TokyoNight() Preset {
	return Preset{
		Axis: lipgloss.NewStyle().Foreground(lipgloss.Color(ColorAxis)),
	}
}
