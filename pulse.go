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
	min, max float64

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

	step   int
	events []eventRecord

	zeroBaseline  bool
	symmetric     bool
	negativeStyle lipgloss.Style
	hasNegStyle   bool
	inverted      bool
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
		m.SetRange(m.min, m.max)
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

// Push adds a data point to the default series. Maintains a sliding window of width w.
func (m *Model) Push(v float64) {
	m.PushSeries("", v)
}

// PushSeries adds a data point to a named series.
func (m *Model) PushSeries(name string, v float64) {
	s := m.getOrCreate(name)
	s.data = append(s.data, v)
	if len(s.data) > m.w {
		s.data = s.data[len(s.data)-m.w:]
	}
	if len(m.order) == 0 || name == m.order[0] || name == "" {
		m.step++
	}
}

// Len returns the number of points in the default series within the window.
func (m *Model) Len() int { return m.LenSeries("") }

// LenSeries returns the number of points in the named series within the window.
func (m *Model) LenSeries(name string) int {
	s, ok := m.series[name]
	if !ok {
		return 0
	}
	return len(s.data)
}

// SetRange dynamically sets the Y-axis range.
func (m *Model) SetRange(min, max float64) {
	if min > max {
		min, max = max, min
	}
	if min == max {
		max = min + 1
	}
	if m.symmetric {
		limit := math.Max(math.Abs(min), math.Abs(max))
		if limit == 0 {
			limit = 1
		}
		min, max = -limit, limit
	}
	m.min, m.max = min, max
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

// Width returns the chart plot width in terminal cells.
func (m *Model) Width() int { return m.w }

// Height returns the chart plot height in terminal cells.
func (m *Model) Height() int { return m.h }

// Min returns the chart minimum Y value.
func (m *Model) Min() float64 { return m.min }

// Max returns the chart maximum Y value.
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

// SeriesData returns a copy of data points for a series.
func (m *Model) SeriesData(name string) []float64 {
	s, ok := m.series[name]
	if !ok {
		return nil
	}
	out := make([]float64, len(s.data))
	copy(out, s.data)
	return out
}

// SeriesStyle returns the style for a named series, or default line style.
func (m *Model) SeriesStyle(name string) lipgloss.Style {
	if s, ok := m.series[name]; ok && s.styled {
		return s.style
	}
	return m.lineStyle
}

// Last returns the most recent value of a named series.
func (m *Model) Last(name string) (float64, bool) {
	s, ok := m.series[name]
	if !ok || len(s.data) == 0 {
		return 0, false
	}
	return s.data[len(s.data)-1], true
}

// SetZeroBaseline enables or disables explicit zero baseline rendering.
func (m *Model) SetZeroBaseline(on bool) { m.zeroBaseline = on }

// ZeroBaseline returns whether explicit zero baseline rendering is enabled.
func (m *Model) ZeroBaseline() bool { return m.zeroBaseline }

// SetSymmetric enables or disables symmetric Y range normalization around zero ([-max, +max]).
func (m *Model) SetSymmetric(on bool) {
	m.symmetric = on
	if on {
		m.SetRange(m.min, m.max)
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
