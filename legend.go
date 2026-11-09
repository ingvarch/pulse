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
		parts = append(parts, swatch+" "+m.legendName(name))
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
		v, ok := m.Last(name)
		lines = append(lines, m.legendEntry(name, v, ok))
	}
	return boxed(lines)
}

// legendEntry renders a series line of a box: the swatch, the name and the value, left empty when ok is false.
func (m *Model) legendEntry(name string, v float64, ok bool) string {
	value := ""
	if ok {
		value = m.formatValue(v)
	}
	text := m.legendName(name) + " " + value
	if name == "" {
		text = value
	}
	return m.SeriesStyle(name).Render("■") + " " + m.axisStyle.Render(text)
}

// boxed returns lines inside a rounded border.
func boxed(lines []string) string {
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Render(strings.Join(lines, "\n"))
}
