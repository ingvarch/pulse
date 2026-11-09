package donut

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// minArc is the least arc length, in units, of a drawn section.
const minArc = 2

// grid returns the section of every sub-pixel, row by row, or -1 for the
// sub-pixels outside the shape. With nothing to draw the shape is section 0.
func (m *Model) grid() [][]int {
	shp := m.shape()
	cw, ch := shp.cells()
	sx, sy := m.glyphs.grid()

	values := make([]float64, len(m.sections))
	for i, s := range m.sections {
		values[i] = s.Value
	}
	share := shares(values, minArc/shp.midLength())
	bounds := make([]float64, len(share)) // cumulative end of each section
	last, acc := -1, 0.0
	for i, s := range share {
		acc += s
		bounds[i] = acc
		if s > 0 {
			last = i
		}
	}

	sectionAt := func(fraction float64) int {
		if last < 0 {
			return 0
		}
		for i, b := range bounds {
			if share[i] > 0 && fraction < b {
				return i
			}
		}
		return last
	}

	g := make([][]int, ch*sy)
	for j := range g {
		g[j] = make([]int, cw*sx)
		y := 2 * (float64(j) + 0.5) / float64(sy)
		for i := range g[j] {
			x := (float64(i) + 0.5) / float64(sx)
			g[j][i] = -1
			if shp.contains(x, y) {
				g[j][i] = sectionAt(shp.fraction(x, y))
			}
		}
	}
	return g
}

// View returns the donut as h lines of equal width.
func (m *Model) View() string {
	g := m.grid()
	cw, ch := m.shape().cells()
	sx, sy := m.glyphs.grid()
	idle := m.total() == 0

	cells := make([][]string, ch)
	sub := make([]int, sx*sy)
	for r := range cells {
		cells[r] = make([]string, cw)
		for c := range cells[r] {
			for k := range sub {
				sub[k] = g[r*sy+k/sx][c*sx+k%sx]
			}
			cells[r][c] = m.renderCell(sub, idle)
		}
	}

	covered := func(r, c0, c1 int) bool { // any shape sub-pixel under the cells?
		for c := c0; c <= c1; c++ {
			for k := 0; k < sy; k++ {
				for i := 0; i < sx; i++ {
					if g[r*sy+k][c*sx+i] >= 0 {
						return true
					}
				}
			}
		}
		return false
	}
	m.placeCenter(cells, covered)

	lines := make([]string, ch)
	for r := range lines {
		lines[r] = strings.Join(cells[r], "")
	}
	return strings.Join(lines, "\n")
}

func (m *Model) renderCell(sub []int, idle bool) string {
	p := pickCell(sub, m.glyphs != Braille)
	if p.fg < 0 {
		return " "
	}
	r, p := glyphFor(m.glyphs, p)
	style := lipgloss.NewStyle()
	if idle {
		style = style.Faint(true)
	} else {
		style = style.Foreground(m.sections[p.fg].Style.GetForeground())
		if p.bg >= 0 {
			style = style.Background(m.sections[p.bg].Style.GetForeground())
		}
	}
	return style.Render(string(r))
}

// placeCenter writes the total, and the label under it, in the middle of
// the cells. A line that would cover the shape is left out.
func (m *Model) placeCenter(cells [][]string, covered func(r, c0, c1 int) bool) {
	total := lipgloss.NewStyle().Bold(true).Render(formatValue(m.total()))
	layouts := [][]string{{total}}
	// Rendering expands tabs, so the width is measured on the rendered label.
	if label := lipgloss.NewStyle().Render(m.label); m.label != "" && lipgloss.Height(label) == 1 {
		layouts = [][]string{{total, label}, {total}}
	}
	h, w := len(cells), len(cells[0])
	place := func(lines []string) bool {
		row := (h - len(lines)) / 2
		for i, txt := range lines {
			tw := lipgloss.Width(txt)
			c := (w - tw) / 2
			if row < 0 || row+i >= h || tw > w || covered(row+i, c, c+tw-1) {
				return false
			}
		}
		for i, txt := range lines {
			tw := lipgloss.Width(txt)
			c := (w - tw) / 2
			cells[row+i][c] = txt
			for k := 1; k < tw; k++ {
				cells[row+i][c+k] = ""
			}
		}
		return true
	}
	for _, lines := range layouts {
		if place(lines) {
			return
		}
	}
}
