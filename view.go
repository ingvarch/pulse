package pulse

import (
	"math"
	"slices"

	"charm.land/lipgloss/v2"

	"github.com/ingvarch/pulse/scale"
)

// tickRow returns the terminal row index (0 at top to h-1 at bottom) for value v.
func (m *Model) tickRow(v float64) int {
	return tickRowOf(m.h, m.min, m.max, v)
}

// maxGridTicks calculates an appropriate grid tick count based on chart height.
func (m *Model) maxGridTicks() int {
	ticks := (m.h - 1) / 3
	if ticks < 2 {
		return 2
	}
	if ticks > 5 {
		return 5
	}
	return ticks
}

// gridTicks returns adaptive grid ticks or user-specified custom ticks.
func (m *Model) gridTicks() []float64 {
	if len(m.customTicks) > 0 {
		return m.customTicks
	}
	return scale.NiceTicks(m.min, m.max, m.maxGridTicks())
}

func (m *Model) formatValue(v float64) string {
	if m.labelFormatter != nil {
		return m.labelFormatter(v)
	}
	return scale.FormatTick(v)
}

// quarterCols returns the columns at 1/4, 2/4 and 3/4 of width w.
func quarterCols(w int) [3]int {
	return [3]int{w / 4, 2 * w / 4, 3 * w / 4}
}

// View renders the chart: Y-axis labels + axis border + plot area,
// then the X axis and label rows when WithXAxis is set.
func (m *Model) View() string {
	names := make([]string, 0, len(m.order)+1)
	names = append(names, "")
	names = append(names, m.order...)
	seriesData := make([][]float64, len(names))
	for idx, name := range names {
		if s, ok := m.series[name]; ok {
			data := m.window(s)
			if s.inverted || (name == "" && m.inverted) {
				data = negated(data)
			}
			seriesData[idx] = data
		}
	}
	if m.autoRange {
		m.min, m.max = m.fitRange(seriesData)
	}

	ticks := m.gridTicks()
	if m.zeroBaseline && m.min <= 0 && m.max >= 0 {
		hasZero := false
		for _, t := range ticks {
			if t == 0 {
				hasZero = true
				break
			}
		}
		if !hasZero {
			ticks = append(ticks, 0)
			slices.Sort(ticks)
		}
	}

	rowLabel := map[int]string{}
	for i := len(ticks) - 1; i >= 0; i-- {
		t := ticks[i]
		if isGap(t) || t < m.min || t > m.max {
			continue
		}
		row := m.tickRow(t)
		if _, taken := rowLabel[row]; !taken {
			rowLabel[row] = m.formatValue(t)
		}
	}

	labelWidth := m.labelWidth
	if labelWidth <= 0 {
		labelWidth = 4
		for _, lbl := range rowLabel {
			if w := lipgloss.Width(lbl); w > labelWidth {
				labelWidth = w
			}
		}
	}
	labelStyle := m.axisStyle.Width(labelWidth).Align(lipgloss.Right)

	zeroRow := -1
	if m.min <= 0 && m.max >= 0 {
		zeroRow = m.tickRow(0)
	}

	var placed []placedLabel
	if m.xAxisLabel != nil {
		placed = m.xAxisLabels()
	}

	hGrid := make([]bool, m.h)
	vGrid := make([]bool, m.w)
	if m.grid {
		for r := range rowLabel {
			hGrid[r] = true
		}
		for _, col := range m.vGridCols(placed) {
			vGrid[col] = true
		}
	}

	styles := make([]lipgloss.Style, len(names))
	negativeStyles := make([]lipgloss.Style, len(names))
	hasNegStyles := make([]bool, len(names))
	for idx, name := range names {
		styles[idx] = m.SeriesStyle(name)
		negativeStyles[idx] = m.SeriesNegativeStyle(name)
		hasNegStyles[idx] = m.hasSeriesNegativeStyle(name)
	}

	gridStyle := m.axisStyle.Faint(true)
	ctx := RenderContext{
		LabelWidth:     labelWidth,
		LabelStyle:     labelStyle,
		RowLabel:       rowLabel,
		HGrid:          hGrid,
		VGrid:          vGrid,
		Names:          names,
		Styles:         styles,
		NegativeStyles: negativeStyles,
		HasNegStyles:   hasNegStyles,
		SeriesData:     seriesData,
		GridStyle:      gridStyle,
		ZeroRow:        zeroRow,
	}

	r := m.renderer
	if r == nil {
		r = LinesRenderer{}
	}
	out := r.Render(m, ctx)
	if m.xAxisLabel == nil {
		return out
	}
	axis, labels := m.xAxisRows(labelWidth, placed)
	return out + "\n" + axis + "\n" + labels
}

// String implements fmt.Stringer, returning the rendered chart View().
func (m *Model) String() string {
	return m.View()
}

// negated returns a copy of data with every sign flipped.
func negated(data []float64) []float64 {
	out := make([]float64, len(data))
	for i, v := range data {
		out[i] = -v
	}
	return out
}

// fitRange returns the auto range for the plotted values: 0 stays inside,
// symmetric mirrors it, and the bounds are rounded out to grid ticks.
func (m *Model) fitRange(seriesData [][]float64) (lo, hi float64) {
	for _, data := range seriesData {
		for _, v := range data {
			if isGap(v) {
				continue
			}
			// Comparisons, not math.Min: a -0 must not replace the starting +0.
			if v < lo {
				lo = v
			}
			if v > hi {
				hi = v
			}
		}
	}
	if m.symmetric {
		hi = math.Max(-lo, hi)
		lo = -hi
	}
	if hi == lo {
		hi = 1
		if m.symmetric {
			lo = -1
		}
	}
	return scale.NiceRange(lo, hi, m.maxGridTicks())
}
