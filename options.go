package pulse

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Option configures a Model.
type Option func(*Model)

// WithRange sets the fixed Y range. For example, WithRange(0, 100) for CPU/RAM.
func WithRange(min, max float64) Option {
	return func(m *Model) { m.SetRange(min, max) }
}

// WithLineStyle sets the default line style.
func WithLineStyle(s lipgloss.Style) Option {
	return func(m *Model) { m.SetLineStyle(s) }
}

// WithAxisStyle sets the axes and labels style.
func WithAxisStyle(s lipgloss.Style) Option {
	return func(m *Model) { m.SetAxisStyle(s) }
}

// WithSeriesStyle sets the style for a named series.
func WithSeriesStyle(name string, s lipgloss.Style) Option {
	return func(m *Model) { m.SetSeriesStyle(name, s) }
}

// WithGrid enables or disables gridlines at tick positions.
func WithGrid(on bool) Option {
	return func(m *Model) { m.SetGrid(on) }
}

// WithFill enables or disables area fill under the line.
func WithFill(on bool) Option {
	return func(m *Model) { m.SetFill(on) }
}

// WithTintedFill enables or disables tinted background fill under the line,
// seamlessly bridging the gap between box-drawing characters and area fill.
func WithTintedFill(on bool) Option {
	return func(m *Model) { m.SetTintedFill(on) }
}

// WithSolidFill enables or disables solid background fill (spaces with background color)
// instead of stippled glyphs (░).
func WithSolidFill(on bool) Option {
	return func(m *Model) { m.SetSolidFill(on) }
}

// WithTintColor sets a custom background tint color for tinted fill mode.
// If nil, DefaultTintColor (#1f2335) is used.
func WithTintColor(c color.Color) Option {
	return func(m *Model) { m.SetTintColor(c) }
}

// WithEvent registers an initial timeline event marker at a historical offset (0 = newest point).
func WithEvent(offset int, ev Event) Option {
	return func(m *Model) { m.AddEventAt(offset, ev) }
}

// WithSmooth enables or disables line smoothing (rounded corners or spline).
func WithSmooth(on bool) Option {
	return func(m *Model) { m.SetSmooth(on) }
}

// WithLineWidth sets line width: 1 thin, 2 bold.
func WithLineWidth(w int) Option {
	return func(m *Model) { m.SetLineWidth(w) }
}

// WithMode sets rendering mode (ModeLines or ModeBraille).
func WithMode(mode RenderMode) Option {
	return func(m *Model) { m.SetMode(mode) }
}

// WithTicks sets explicit Y-axis tick values (e.g. 0, 25, 50, 75, 100 or 0, 50, 100).
func WithTicks(ticks ...float64) Option {
	return func(m *Model) { m.SetTicks(ticks...) }
}

// WithLabelFormatter sets a custom formatter for Y-axis tick values.
func WithLabelFormatter(fn func(float64) string) Option {
	return func(m *Model) { m.SetLabelFormatter(fn) }
}

// WithLabelWidth sets an explicit width for Y-axis labels.
func WithLabelWidth(w int) Option {
	return func(m *Model) { m.SetLabelWidth(w) }
}

// WithRenderer sets a custom chart renderer.
func WithRenderer(r Renderer) Option {
	return func(m *Model) { m.SetRenderer(r) }
}

// WithZeroBaseline enables or disables explicit zero baseline rendering (├ on Y-axis and baseline ruling at Y=0).
func WithZeroBaseline(on bool) Option {
	return func(m *Model) { m.SetZeroBaseline(on) }
}

// WithSymmetric normalizes the Y range to be symmetric around zero ([-max, +max]),
// ensuring the zero baseline remains anchored directly in the center of the chart.
func WithSymmetric(on bool) Option {
	return func(m *Model) { m.SetSymmetric(on) }
}

// WithNegativeStyle sets the style for negative values (< 0) of the default series.
func WithNegativeStyle(s lipgloss.Style) Option {
	return func(m *Model) { m.SetNegativeStyle(s) }
}

// WithSeriesNegativeStyle sets the style for negative values (< 0) of a named series.
func WithSeriesNegativeStyle(name string, s lipgloss.Style) Option {
	return func(m *Model) { m.SetSeriesNegativeStyle(name, s) }
}

// WithSeriesInverted configures a named series to invert its values when plotted (v -> -v).
func WithSeriesInverted(name string, inverted bool) Option {
	return func(m *Model) { m.SetSeriesInverted(name, inverted) }
}

// WithInverted configures the default series to invert its values when plotted (v -> -v).
func WithInverted(inverted bool) Option {
	return func(m *Model) { m.SetInverted(inverted) }
}
