package pulse

import (
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
)

type placedLabel struct {
	col, start, width int
	text              string
	middle            bool
}

func (l placedLabel) end() int { return l.start + l.width }

// xAxisCols returns the candidate label columns in placement order:
// every column from right to left with WithXAxisAt, else the right edge, the left edge, then the quarters.
func (m *Model) xAxisCols() []int {
	if m.xAxisAt {
		cols := make([]int, m.w)
		for i := range cols {
			cols[i] = m.w - 1 - i
		}
		return cols
	}
	cols := []int{m.w - 1, 0}
	for _, col := range quarterCols(m.w) {
		if col > 0 && col < m.w-1 {
			cols = append(cols, col)
		}
	}
	return cols
}

// xAxisLabels places the labels for the candidate columns in order.
// A label that does not fit in the plot or touches a placed one is dropped.
func (m *Model) xAxisLabels() []placedLabel {
	var placed []placedLabel
	for _, col := range m.xAxisCols() {
		text := m.xAxisLabel(m.w - 1 - col)
		width := lipgloss.Width(text)
		if width == 0 {
			continue
		}
		l := placedLabel{col: col, width: width, text: text}
		switch col {
		case m.w - 1:
			l.start = m.w - width
		case 0:
			l.start = 0
		default:
			l.start = col - (width-1)/2
			l.middle = true
		}
		if l.start < 0 || l.end() > m.w {
			continue
		}
		if !slices.ContainsFunc(placed, func(p placedLabel) bool {
			return l.start <= p.end() && p.start <= l.end()
		}) {
			placed = append(placed, l)
		}
	}
	slices.SortFunc(placed, func(a, b placedLabel) int { return a.start - b.start })
	return placed
}

// vGridCols returns the vertical grid columns: the ticks of the placed labels with WithXAxisAt, else the quarters.
func (m *Model) vGridCols(placed []placedLabel) []int {
	if !m.xAxisAt {
		q := quarterCols(m.w)
		return q[:]
	}
	var cols []int
	for _, l := range placed {
		if l.middle {
			cols = append(cols, l.col)
		}
	}
	return cols
}

// xAxisRows renders the axis row and the label row under the plot.
func (m *Model) xAxisRows(labelWidth int, placed []placedLabel) (axis, labels string) {
	ticks := make([]bool, m.w)
	for _, l := range placed {
		ticks[l.col] = l.middle
	}

	var ax strings.Builder
	ax.WriteString("└")
	for _, t := range ticks {
		if t {
			ax.WriteString("┴")
		} else {
			ax.WriteString("─")
		}
	}

	var lb strings.Builder
	lb.WriteString(strings.Repeat(" ", labelWidth+1))
	pos := 0
	for _, l := range placed {
		lb.WriteString(strings.Repeat(" ", l.start-pos))
		lb.WriteString(m.axisStyle.Render(l.text))
		pos = l.end()
	}
	lb.WriteString(strings.Repeat(" ", m.w-pos))

	return strings.Repeat(" ", labelWidth) + m.axisStyle.Render(ax.String()), lb.String()
}
