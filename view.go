package pulse

import (
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

// View renders the chart: Y-axis labels + axis border + plot area.
func (m *Model) View() string {
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
		if t < m.min || t > m.max {
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

	hGrid := make([]bool, m.h)
	vGrid := make([]bool, m.w)
	if m.grid {
		for r := range hGrid {
			hGrid[r] = true
		}
		for k := 1; k < 4; k++ {
			if col := k * m.w / 4; col < m.w {
				vGrid[col] = true
			}
		}
	}

	names := make([]string, 0, len(m.order)+1)
	names = append(names, "")
	names = append(names, m.order...)
	styles := make([]lipgloss.Style, len(names))
	negativeStyles := make([]lipgloss.Style, len(names))
	hasNegStyles := make([]bool, len(names))
	seriesData := make([][]float64, len(names))
	for idx, name := range names {
		styles[idx] = m.SeriesStyle(name)
		negativeStyles[idx] = m.SeriesNegativeStyle(name)
		hasNegStyles[idx] = m.hasSeriesNegativeStyle(name)
		if s, ok := m.series[name]; ok {
			isInverted := s.inverted || (name == "" && m.inverted)
			if isInverted {
				inv := make([]float64, len(s.data))
				for i, val := range s.data {
					inv[i] = -val
				}
				seriesData[idx] = inv
			} else {
				seriesData[idx] = s.data
			}
		}
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
	return r.Render(m, ctx)
}

// String implements fmt.Stringer, returning the rendered chart View().
func (m *Model) String() string {
	return m.View()
}
