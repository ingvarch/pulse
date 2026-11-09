package pulse

import (
	"slices"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/ingvarch/pulse/scale"
)

// Axis is a Y axis with its own range and labels, such as the right axis of WithRightAxis.
type Axis struct {
	// Min and Max are the range of the axis.
	Min, Max float64
	// Ticks are the labeled values. When empty, scale.NiceTicks picks them.
	Ticks []float64
	// Formatter formats a tick label. When nil, scale.FormatTick is used.
	Formatter func(float64) string
}

// WithRightAxis adds a Y axis on the right with its own range and labels, as SetRightAxis does.
func WithRightAxis(a Axis) Option {
	return func(m *Model) { m.SetRightAxis(a) }
}

// WithSeriesRightAxis plots the named series against the right axis.
func WithSeriesRightAxis(name string) Option {
	return func(m *Model) { m.SetSeriesRightAxis(name, true) }
}

// SetRightAxis adds or replaces the right Y axis.
// A call with a NaN or infinite bound is ignored. Min above Max is swapped, and equal bounds widen to Min..Min+1.
// Ticks are copied, sorted and deduplicated.
// Series on the right axis are drawn against its range; the horizontal grid follows the left ticks.
func (m *Model) SetRightAxis(a Axis) {
	if isGap(a.Min) || isGap(a.Max) {
		return
	}
	if a.Min > a.Max {
		a.Min, a.Max = a.Max, a.Min
	}
	if a.Min == a.Max {
		a.Max = a.Min + 1
	}
	a.Ticks = sortedTicks(a.Ticks)
	m.rightAxis, m.hasRightAxis = a, true
}

// RightAxis returns the right axis and whether one is set.
func (m *Model) RightAxis() (Axis, bool) {
	a := m.rightAxis
	a.Ticks = slices.Clone(a.Ticks)
	return a, m.hasRightAxis
}

// SetSeriesRightAxis plots the named series against the right axis, or back against the left one.
// It has no effect until a right axis is set.
func (m *Model) SetSeriesRightAxis(name string, on bool) {
	m.getOrCreate(name).rightAxis = on
}

// SeriesRightAxis returns whether the named series is set to the right axis.
func (m *Model) SeriesRightAxis(name string) bool {
	s, ok := m.series[name]
	return ok && s.rightAxis
}

// onRight reports whether the named series is drawn against the right axis.
func (m *Model) onRight(name string) bool {
	return m.hasRightAxis && m.SeriesRightAxis(name)
}

// legendName returns the series name for a legend, marked when the series is on the right axis.
func (m *Model) legendName(name string) string {
	if m.onRight(name) {
		return name + " →"
	}
	return name
}

// rightRowLabels returns the right axis labels by row, and the width of the widest one.
func (m *Model) rightRowLabels() (map[int]string, int) {
	a := m.rightAxis
	ticks := a.Ticks
	if len(ticks) == 0 {
		ticks = scale.NiceTicks(a.Min, a.Max, m.maxGridTicks())
	}
	format := a.Formatter
	if format == nil {
		format = scale.FormatTick
	}
	labels := rowLabels(ticks, a.Min, a.Max, func(v float64) int { return tickRowOf(m.h, a.Min, a.Max, v) }, format)
	width := 0
	for _, lbl := range labels {
		width = max(width, lipgloss.Width(lbl))
	}
	return labels, width
}

// appendRightAxis adds the right axis line and the label column to every plot line of out.
func (m *Model) appendRightAxis(out string, labels map[int]string, width, zeroRow int) string {
	style := m.axisStyle.Width(width).Align(lipgloss.Left)
	lines := strings.Split(out, "\n")
	for r := range lines {
		axis := "│"
		if m.zeroBaseline && r == zeroRow {
			axis = "┤"
		}
		lines[r] += m.axisStyle.Render(axis) + style.Render(labels[r])
	}
	return strings.Join(lines, "\n")
}

// withoutRight returns the series data with the right-axis series left out as nil.
func withoutRight(seriesData [][]float64, onRight []bool) [][]float64 {
	if !slices.Contains(onRight, true) {
		return seriesData
	}
	out := slices.Clone(seriesData)
	for i, right := range onRight {
		if right {
			out[i] = nil
		}
	}
	return out
}

// rightChart is the Chart a renderer sees when series are on the right axis: their values are already in the left drawing space.
type rightChart struct {
	Chart
	data map[string][]float64
}

// SeriesData returns the mapped values of a right-axis series.
func (c rightChart) SeriesData(name string) []float64 {
	if d, ok := c.data[name]; ok {
		return d
	}
	return c.Chart.SeriesData(name)
}

// SeriesInverted is false for a right-axis series: its values are negated before the mapping.
func (c rightChart) SeriesInverted(name string) bool {
	if _, ok := c.data[name]; ok {
		return false
	}
	return c.Chart.SeriesInverted(name)
}

// toRightPlot maps the right-axis series into the range chart sees and stores them in plot. Gaps stay gaps.
// It returns the Chart the renderers get.
func (m *Model) toRightPlot(names []string, onRight []bool, seriesData, plot [][]float64, chart Chart) Chart {
	lo, hi := chart.Min(), chart.Max()
	a := m.rightAxis
	byName := map[string][]float64{}
	for i, right := range onRight {
		if !right || seriesData[i] == nil {
			continue
		}
		d := make([]float64, len(seriesData[i]))
		for j, v := range seriesData[i] {
			d[j] = v
			if !isGap(v) {
				d[j] = lo + (v-a.Min)/(a.Max-a.Min)*(hi-lo)
			}
		}
		plot[i], byName[names[i]] = d, d
	}
	if len(byName) == 0 {
		return chart
	}
	return rightChart{Chart: chart, data: byName}
}
