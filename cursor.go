package pulse

import (
	"image/color"
	"math"

	"charm.land/lipgloss/v2"
)

// SetCursor shows a vertical cursor at the point age steps before the newest one.
// A negative age hides it. An age beyond the visible window draws nothing, but CursorValues still works.
func (m *Model) SetCursor(age int) { m.cursor = max(age, -1) }

// Cursor returns the age of the cursor, or -1 when it is hidden.
func (m *Model) Cursor() int { return m.cursor }

// SetOffset scrolls the view n points back into the history; 0 follows the newest point.
// The offset stops at the oldest point. While it is above zero, each step (a push to the default series,
// or to the first named series while the default has none) moves it by one, and a shown cursor with it,
// so the view stays on the same points.
func (m *Model) SetOffset(n int) { m.offset = min(max(n, 0), m.maxOffset()) }

// Offset returns how many points the view is scrolled back.
func (m *Model) Offset() int { return m.offset }

// maxOffset returns the largest offset: the oldest point of the longest series at the left edge.
func (m *Model) maxOffset() int {
	longest := 0
	for _, s := range m.series {
		longest = max(longest, len(s.data))
	}
	return max(longest-m.w, 0)
}

// followPush keeps the view and the cursor on the same points after a push while the view is scrolled.
func (m *Model) followPush() {
	if m.offset == 0 {
		return
	}
	moved := min(m.offset+1, m.maxOffset()) - m.offset
	m.offset += moved
	if m.cursor >= 0 {
		m.cursor += moved
	}
}

// colOfAge returns the plot column of the point age steps before the newest one. It can lie outside the plot.
func (m *Model) colOfAge(age int) int { return m.w - 1 - (age - m.offset) }

// ageOfCol returns the age of the point in plot column col; it is the inverse of colOfAge.
func (m *Model) ageOfCol(col int) int { return m.offset + m.w - 1 - col }

// cursorCol returns the plot column of the cursor, or -1 when it is hidden or outside the plot.
func (m *Model) cursorCol() int {
	if col := m.colOfAge(m.cursor); m.cursor >= 0 && col >= 0 && col < m.w {
		return col
	}
	return -1
}

// renderCursor renders a cursor cell in the axis style, on the background bg when it is set.
func renderCursor(c Chart, bg color.Color) string {
	return onBackground(c.AxisStyle(), bg).Render(guideGlyph)
}

// CursorValues returns each series' pushed value at the cursor, or nil when the cursor is hidden.
// A gap or a missing point is NaN. The default series is listed once it has data.
func (m *Model) CursorValues() map[string]float64 {
	if m.cursor < 0 {
		return nil
	}
	out := map[string]float64{}
	for _, name := range m.cursorNames() {
		out[name] = m.valueAt(m.series[name], m.cursor)
	}
	return out
}

// cursorNames returns the series shown at the cursor: the default series when it has data, then the named ones.
func (m *Model) cursorNames() []string {
	var names []string
	if s, ok := m.series[""]; ok && len(s.data) > 0 {
		names = append(names, "")
	}
	return append(names, m.order...)
}

// valueAt returns the value age steps before the newest point of s; a gap or a missing point is NaN.
func (m *Model) valueAt(s *series, age int) float64 {
	if i := len(s.data) - 1 - age; i >= 0 && !isGap(s.data[i]) {
		return s.data[i]
	}
	return math.NaN()
}

// CursorBox returns a boxed card for the cursor: the X axis label at the cursor, when an X axis is set
// and the label is not empty, then each series with its swatch and value. It returns an empty string while the cursor is hidden.
func (m *Model) CursorBox() string {
	if m.cursor < 0 {
		return ""
	}
	var lines []string
	if m.xAxisLabel != nil {
		if label := m.xAxisLabel(m.cursor); lipgloss.Width(label) > 0 {
			lines = append(lines, m.axisStyle.Render(label))
		}
	}
	for _, name := range m.cursorNames() {
		v := m.valueAt(m.series[name], m.cursor)
		lines = append(lines, m.legendEntry(name, v, !math.IsNaN(v)))
	}
	if len(lines) == 0 {
		return ""
	}
	return boxed(lines)
}
