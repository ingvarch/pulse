package pulse

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// Legend returns a formatted legend for named series, or empty string if none exist.
func (m *Model) Legend() string {
	parts := make([]string, 0, len(m.order))
	for _, name := range m.order {
		swatch := m.SeriesStyle(name).Render("■")
		parts = append(parts, swatch+" "+name)
	}
	return strings.Join(parts, "  ")
}

// LegendBox returns a boxed legend with borders, swatches, series names, and recent values.
func (m *Model) LegendBox() string {
	if len(m.order) == 0 {
		return ""
	}
	lines := make([]string, 0, len(m.order))
	for _, name := range m.order {
		swatch := m.SeriesStyle(name).Render("■")
		value := ""
		if v, ok := m.Last(name); ok {
			value = m.formatValue(v)
		}
		lines = append(lines, swatch+" "+m.axisStyle.Render(name+" "+value))
	}
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Render(strings.Join(lines, "\n"))
}
