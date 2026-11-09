package pulse

import (
	"cmp"
	"slices"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// Threshold marks a limit on the Y axis, such as a warning at 80 or a critical level at 90.
type Threshold struct {
	// Value is the Y value of the limit.
	Value float64
	// Style colors the line and label, and the series at or above Value when threshold colors are on.
	Style lipgloss.Style
	// Label is drawn at the right end of the line. It needs Line and is dropped when it does not fit.
	Label string
	// Line draws a dashed horizontal line at Value.
	Line bool
}

// SetThresholds replaces the thresholds with a copy of ts. A threshold with a NaN or infinite Value is ignored.
func (m *Model) SetThresholds(ts ...Threshold) {
	m.thresholds = nil
	for _, t := range ts {
		if !isGap(t.Value) {
			m.thresholds = append(m.thresholds, t)
		}
	}
}

// SetThresholdColors turns on or off coloring of series by the thresholds.
func (m *Model) SetThresholdColors(on bool) { m.thresholdColors = on }

// ThresholdColors returns whether series are colored by the thresholds.
func (m *Model) ThresholdColors() bool { return m.thresholdColors }

const thresholdGlyph = "┄"

// thresholdCells prepares the threshold lines and the row styles for the renderer.
func (m *Model) thresholdCells() (lines map[int][]string, styles map[int]lipgloss.Style) {
	lines, styles = map[int][]string{}, map[int]lipgloss.Style{}
	ts := slices.Clone(m.thresholds)
	slices.SortStableFunc(ts, func(a, b Threshold) int { return cmp.Compare(b.Value, a.Value) })
	for _, t := range ts {
		if t.Line && t.Value >= m.min && t.Value <= m.max {
			row := m.tickRow(t.Value)
			if _, taken := lines[row]; !taken {
				lines[row] = m.thresholdLine(t)
			}
		}
		if m.thresholdColors && t.Value <= m.max {
			for r := m.tickRow(t.Value); r >= 0; r-- {
				if _, taken := styles[r]; !taken {
					styles[r] = t.Style
				}
			}
		}
	}
	return lines, styles
}

// thresholdLine returns one rendered cell per column: dashes, with the label at the right end if it fits.
// A wide label cell is followed by an empty cell, and the first cell is always a dash.
func (m *Model) thresholdLine(t Threshold) []string {
	cells := make([]string, m.w)
	dash := t.Style.Render(thresholdGlyph)
	for i := range cells {
		cells[i] = dash
	}
	label, width := labelClusters(t.Label)
	if width == 0 || width > m.w-2 {
		return cells
	}
	col := m.w - width
	for _, g := range label {
		cells[col] = t.Style.Render(g.text)
		for k := 1; k < g.width; k++ {
			cells[col+k] = ""
		}
		col += g.width
	}
	return cells
}

type cluster struct {
	text  string
	width int
}

// labelClusters splits the label into grapheme clusters, each with its cell width, and returns their total width.
// ANSI sequences are stripped. A label with a control character or a zero-width cluster gets width 0.
func labelClusters(label string) ([]cluster, int) {
	s := ansi.Strip(label)
	var out []cluster
	total := 0
	for s != "" {
		text, w := ansi.FirstGraphemeCluster(s, ansi.GraphemeWidth)
		if w < 1 {
			return nil, 0
		}
		out = append(out, cluster{text, w})
		total += w
		s = s[len(text):]
	}
	return out, total
}
