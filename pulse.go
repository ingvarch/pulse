package pulse

import (
	"image/color"
	"math"
	"slices"

	"charm.land/lipgloss/v2"
)

var _ Chart = (*Model)(nil)

// Model is a streaming terminal line chart.
type Model struct {
	w, h     int
	history  int
	min, max float64
	// rangeMin and rangeMax are the range before the symmetric rewrite, as set or as auto range computed it.
	rangeMin, rangeMax float64

	series map[string]*series
	order  []string

	grid       bool
	fill       bool
	tintedFill bool
	solidFill  bool
	tintColor  color.Color
	smooth     bool
	lineWidth  int
	renderer   Renderer

	lineStyle lipgloss.Style
	axisStyle lipgloss.Style

	customTicks []float64

	labelWidth     int
	labelFormatter func(float64) string
	xAxisLabel     func(age int) string
	xAxisAt        bool

	step   int
	events []eventRecord

	autoRange     bool
	logScale      bool
	zeroBaseline  bool
	symmetric     bool
	negativeStyle lipgloss.Style
	hasNegStyle   bool
	inverted      bool

	thresholds      []Threshold
	thresholdColors bool
}

type series struct {
	data          []float64
	style         lipgloss.Style
	styled        bool
	negativeStyle lipgloss.Style
	hasNegStyle   bool
	inverted      bool
}

// New creates a chart for a w x h cells plot area.
func New(w, h int, opts ...Option) *Model {
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	m := &Model{
		w:          w,
		h:          h,
		min:        0,
		max:        100,
		rangeMax:   100,
		series:     map[string]*series{},
		grid:       true,
		fill:       true,
		tintedFill: false,
		solidFill:  false,
		tintColor:  nil,
		smooth:     true,
		lineWidth:  2,
		renderer:   LinesRenderer{},
	}
	for _, o := range opts {
		o(m)
	}
	if m.symmetric {
		m.setRange(m.rangeMin, m.rangeMax)
	}
	if m.min > m.max {
		m.min, m.max = m.max, m.min
	}
	if m.max == m.min {
		m.max = m.min + 1
	}
	return m
}

func (m *Model) getOrCreate(name string) *series {
	s, ok := m.series[name]
	if !ok {
		s = &series{}
		if name == "" && m.hasNegStyle {
			s.negativeStyle = m.negativeStyle
			s.hasNegStyle = true
		}
		m.series[name] = s
		if name != "" {
			m.order = append(m.order, name)
		}
	}
	return s
}

// Push adds a data point to the default series. The chart shows the last w points
// and keeps up to max(history, w) of them.
// NaN, +Inf and -Inf are kept as given and drawn as gaps.
func (m *Model) Push(v float64) {
	m.PushSeries("", v)
}

// PushSeries adds a data point to a named series.
// NaN, +Inf and -Inf are kept as given and drawn as gaps.
func (m *Model) PushSeries(name string, v float64) {
	s := m.getOrCreate(name)
	s.data = append(s.data, v)
	s.data = lastN(s.data, m.capacity())
	if len(m.order) == 0 || name == m.order[0] || name == "" {
		m.step++
		m.pruneEvents()
	}
}

// SetSize changes the plot size and clamps w and h to at least 1, as New does.
// Points stay in the history: a wider chart shows older points again.
// A smaller capacity drops the oldest points.
func (m *Model) SetSize(w, h int) {
	m.w, m.h = max(w, 1), max(h, 1)
	m.trim()
}

// SetHistory keeps up to n points per series, and never fewer than the plot width.
// A negative n counts as 0.
func (m *Model) SetHistory(n int) {
	m.history = max(n, 0)
	m.trim()
}

// capacity returns how many points each series keeps.
func (m *Model) capacity() int { return max(m.history, m.w) }

// trim drops the points and events that no longer fit in the capacity.
func (m *Model) trim() {
	m.pruneEvents()
	for _, s := range m.series {
		s.data = lastN(s.data, m.capacity())
	}
}

// lastN returns the last n elements of data, or all of it when it is shorter.
func lastN(data []float64, n int) []float64 {
	if len(data) > n {
		return data[len(data)-n:]
	}
	return data
}

// window returns the points of s inside the visible window.
func (m *Model) window(s *series) []float64 { return lastN(s.data, m.w) }

// Len returns the number of points in the default series within the window.
func (m *Model) Len() int { return m.LenSeries("") }

// LenSeries returns the number of points in the named series within the window.
func (m *Model) LenSeries(name string) int {
	s, ok := m.series[name]
	if !ok {
		return 0
	}
	return len(m.window(s))
}

// SetRange sets a fixed Y-axis range and turns auto range off.
// A call with a NaN or infinite bound is ignored and keeps the current range.
func (m *Model) SetRange(min, max float64) {
	if m.setRange(min, max) {
		m.autoRange = false
	}
}

// SetAutoRange turns the automatic Y range on or off. See WithAutoRange.
// Turning it off keeps the range of the last render.
func (m *Model) SetAutoRange(on bool) { m.autoRange = on }

// AutoRange returns whether the automatic Y range is on.
func (m *Model) AutoRange() bool { return m.autoRange }

// setRange stores the range without touching auto range; it reports whether the range was applied.
func (m *Model) setRange(min, max float64) bool {
	if isGap(min) || isGap(max) {
		return false
	}
	if min > max {
		min, max = max, min
	}
	if min == max {
		max = min + 1
	}
	m.rangeMin, m.rangeMax = min, max
	if m.symmetric {
		limit := math.Max(math.Abs(min), math.Abs(max))
		if limit == 0 {
			limit = 1
		}
		min, max = -limit, limit
	}
	m.min, m.max = min, max
	return true
}

// SetLineStyle dynamically updates the line style.
func (m *Model) SetLineStyle(s lipgloss.Style) { m.lineStyle = s }

// SetAxisStyle dynamically updates the axis style.
func (m *Model) SetAxisStyle(s lipgloss.Style) { m.axisStyle = s }

// SetSeriesStyle dynamically sets the style for a named series.
func (m *Model) SetSeriesStyle(name string, s lipgloss.Style) {
	ms := m.getOrCreate(name)
	ms.style, ms.styled = s, true
}

// SetGrid enables or disables the grid dynamically.
func (m *Model) SetGrid(on bool) { m.grid = on }

// SetFill enables or disables area fill dynamically.
func (m *Model) SetFill(on bool) { m.fill = on }

// SetTintedFill enables or disables tinted background fill dynamically.
func (m *Model) SetTintedFill(on bool) { m.tintedFill = on }

// SetSolidFill enables or disables solid background fill dynamically.
func (m *Model) SetSolidFill(on bool) { m.solidFill = on }

// SetTintColor dynamically updates the chart background tint color.
func (m *Model) SetTintColor(c color.Color) { m.tintColor = c }

// SetSmooth enables or disables smoothing dynamically.
func (m *Model) SetSmooth(on bool) { m.smooth = on }

// SetLineWidth sets the line width: 1 for thin, 2 for bold.
func (m *Model) SetLineWidth(w int) {
	if w < 1 {
		w = 1
	}
	if w > 2 {
		w = 2
	}
	m.lineWidth = w
}

// LineWidth returns the current line width.
func (m *Model) LineWidth() int {
	return m.lineWidth
}

// SetTicks dynamically sets explicit tick values for the Y axis.
// Ticks are copied, sorted ascending, and deduplicated.
func (m *Model) SetTicks(ticks ...float64) {
	if len(ticks) == 0 {
		m.customTicks = nil
		return
	}
	cp := make([]float64, len(ticks))
	copy(cp, ticks)
	slices.Sort(cp)
	m.customTicks = slices.Compact(cp)
}

// SetLabelFormatter dynamically updates the Y-axis label formatter.
func (m *Model) SetLabelFormatter(fn func(float64) string) {
	m.labelFormatter = fn
}

// SetLabelWidth dynamically updates the Y-axis label column width.
func (m *Model) SetLabelWidth(w int) {
	m.labelWidth = w
}

// SetMode sets the rendering mode (ModeLines or ModeBraille).
// ModeCustom cannot be set: install the renderer with SetRenderer instead.
func (m *Model) SetMode(mode RenderMode) {
	switch mode {
	case ModeBraille:
		m.renderer = BrailleRenderer{}
	default:
		m.renderer = LinesRenderer{}
	}
}

// SetRenderer dynamically updates the chart renderer and synchronizes Mode().
// A nil renderer falls back to LinesRenderer.
func (m *Model) SetRenderer(r Renderer) {
	if r == nil {
		r = LinesRenderer{}
	}
	m.renderer = r
}

// Mode returns the current rendering mode derived from the active renderer.
func (m *Model) Mode() RenderMode {
	switch m.renderer.(type) {
	case BrailleRenderer, *BrailleRenderer:
		return ModeBraille
	case LinesRenderer, *LinesRenderer:
		return ModeLines
	default:
		return ModeCustom
	}
}

// ToggleRenderMode switches between ModeLines and ModeBraille.
func (m *Model) ToggleRenderMode() {
	if m.Mode() == ModeLines {
		m.SetMode(ModeBraille)
	} else {
		m.SetMode(ModeLines)
	}
}

// Renderer returns the current chart renderer.
func (m *Model) Renderer() Renderer {
	return m.renderer
}

// SetXAxis sets the X axis label function; nil turns the axis off.
// See WithXAxis.
func (m *Model) SetXAxis(label func(age int) string) { m.xAxisLabel, m.xAxisAt = label, false }

// SetXAxisAt sets the X axis label function asked for every column; nil turns the axis off.
// See WithXAxisAt.
func (m *Model) SetXAxisAt(label func(age int) string) { m.xAxisLabel, m.xAxisAt = label, label != nil }

// Width returns the chart plot width in terminal cells.
func (m *Model) Width() int { return m.w }

// Height returns the chart plot height in terminal cells.
// With an X axis, View returns Height()+2 lines.
func (m *Model) Height() int { return m.h }

// Min returns the chart minimum Y value. With auto range it is the minimum of the last render.
// On a log scale the drawn range can differ; see SetLogScale.
func (m *Model) Min() float64 { return m.min }

// Max returns the chart maximum Y value. With auto range it is the maximum of the last render.
// On a log scale the drawn range can differ; see SetLogScale.
func (m *Model) Max() float64 { return m.max }

// Fill returns whether area fill is enabled.
func (m *Model) Fill() bool { return m.fill }

// TintedFill returns whether tinted area fill is enabled.
func (m *Model) TintedFill() bool { return m.tintedFill }

// SolidFill returns whether solid background fill is enabled.
func (m *Model) SolidFill() bool { return m.solidFill }

// TintColor returns the custom background tint color, or nil if using default.
func (m *Model) TintColor() color.Color { return m.tintColor }

// Grid returns whether grid lines are enabled.
func (m *Model) Grid() bool { return m.grid }

// Smooth returns whether line smoothing is enabled.
func (m *Model) Smooth() bool { return m.smooth }

// AxisStyle returns the axis lipgloss style.
func (m *Model) AxisStyle() lipgloss.Style { return m.axisStyle }

// LineStyle returns the default line lipgloss style.
func (m *Model) LineStyle() lipgloss.Style { return m.lineStyle }

// SeriesNames returns registered series names in order.
func (m *Model) SeriesNames() []string {
	out := make([]string, len(m.order))
	copy(out, m.order)
	return out
}

// SeriesData returns a copy of the points of a series in the visible window (the last Width() points).
// Missing points are returned raw (NaN, +Inf, -Inf), so custom renderers see them.
func (m *Model) SeriesData(name string) []float64 {
	s, ok := m.series[name]
	if !ok {
		return nil
	}
	return slices.Clone(m.window(s))
}

// SeriesStyle returns the style for a named series, or default line style.
func (m *Model) SeriesStyle(name string) lipgloss.Style {
	if s, ok := m.series[name]; ok && s.styled {
		return s.style
	}
	return m.lineStyle
}

// Last returns the most recent finite value of a named series in the visible window.
// Trailing gaps (NaN, +Inf, -Inf) are skipped.
func (m *Model) Last(name string) (float64, bool) {
	s, ok := m.series[name]
	if !ok {
		return 0, false
	}
	data := m.window(s)
	for i := len(data) - 1; i >= 0; i-- {
		if !isGap(data[i]) {
			return data[i], true
		}
	}
	return 0, false
}

// SetZeroBaseline enables or disables explicit zero baseline rendering.
func (m *Model) SetZeroBaseline(on bool) { m.zeroBaseline = on }

// ZeroBaseline returns whether explicit zero baseline rendering is enabled.
func (m *Model) ZeroBaseline() bool { return m.zeroBaseline }

// SetSymmetric enables or disables symmetric Y range normalization around zero ([-max, +max]).
func (m *Model) SetSymmetric(on bool) {
	m.symmetric = on
	if on {
		m.setRange(m.rangeMin, m.rangeMax)
	}
}

// Symmetric returns whether symmetric Y range normalization is enabled.
func (m *Model) Symmetric() bool { return m.symmetric }

// SetNegativeStyle dynamically sets the negative line style for the default series.
func (m *Model) SetNegativeStyle(s lipgloss.Style) {
	m.negativeStyle = s
	m.hasNegStyle = true
	if ms, ok := m.series[""]; ok {
		ms.negativeStyle = s
		ms.hasNegStyle = true
	}
}

// NegativeStyle returns the negative line style for the default series, or default line style.
func (m *Model) NegativeStyle() lipgloss.Style {
	if s, ok := m.series[""]; ok && s.hasNegStyle {
		return s.negativeStyle
	}
	if m.hasNegStyle {
		return m.negativeStyle
	}
	return m.lineStyle
}

// SetSeriesNegativeStyle dynamically sets the negative style for a named series.
func (m *Model) SetSeriesNegativeStyle(name string, s lipgloss.Style) {
	ms := m.getOrCreate(name)
	ms.negativeStyle = s
	ms.hasNegStyle = true
	if name == "" {
		m.negativeStyle = s
		m.hasNegStyle = true
	}
}

// SeriesNegativeStyle returns the negative style for a named series.
func (m *Model) SeriesNegativeStyle(name string) lipgloss.Style {
	if name == "" {
		return m.NegativeStyle()
	}
	if s, ok := m.series[name]; ok && s.hasNegStyle {
		return s.negativeStyle
	}
	return m.SeriesStyle(name)
}

func (m *Model) hasSeriesNegativeStyle(name string) bool {
	if name == "" {
		if s, ok := m.series[""]; ok && s.hasNegStyle {
			return true
		}
		return m.hasNegStyle
	}
	if s, ok := m.series[name]; ok {
		return s.hasNegStyle
	}
	return false
}

// SetSeriesInverted enables or disables value inversion for a named series.
func (m *Model) SetSeriesInverted(name string, inverted bool) {
	ms := m.getOrCreate(name)
	ms.inverted = inverted
}

// SeriesInverted returns whether value inversion is enabled for a named series.
func (m *Model) SeriesInverted(name string) bool {
	if s, ok := m.series[name]; ok {
		return s.inverted
	}
	if name == "" {
		return m.inverted
	}
	return false
}

// SetInverted enables or disables value inversion for the default series.
func (m *Model) SetInverted(inverted bool) {
	m.inverted = inverted
}

// Inverted returns whether value inversion is enabled for the default series.
func (m *Model) Inverted() bool {
	return m.inverted
}
