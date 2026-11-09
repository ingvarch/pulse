package pulse

import (
	"image/color"
	"math"
	"strings"

	"charm.land/lipgloss/v2"
)

// RenderMode defines how line curves are rendered on the chart.
type RenderMode int

const (
	// ModeLines renders smooth continuous lines (box-drawing ╭, ╮, ╯, ╰, ─, │) with Grafana-style area fill ░.
	ModeLines RenderMode = iota
	// ModeBraille renders using a 2x4 dot braille matrix (Unicode Braille Patterns).
	ModeBraille
	// ModeCustom represents a custom user-supplied Renderer.
	ModeCustom
)

// DefaultTintColor is the default background tint used when TintedFill is active
// and no custom tint color or style background is specified. Defaults to Tokyo Night slate (#1f2335).
var DefaultTintColor color.Color = lipgloss.Color("#1f2335")

// Chart provides the read-only contract for inspecting chart dimensions, data, and styles.
type Chart interface {
	Width() int
	Height() int
	Min() float64
	Max() float64
	Fill() bool
	TintedFill() bool
	SolidFill() bool
	TintColor() color.Color
	Grid() bool
	Smooth() bool
	LineWidth() int
	AxisStyle() lipgloss.Style
	LineStyle() lipgloss.Style
	SeriesStyle(name string) lipgloss.Style
	SeriesNames() []string
	// SeriesData returns raw values, including NaN and Inf for missing points.
	SeriesData(name string) []float64
	VisibleEvents() []VisibleEvent
	ZeroBaseline() bool
	Symmetric() bool
	NegativeStyle() lipgloss.Style
	SeriesNegativeStyle(name string) lipgloss.Style
	SeriesInverted(name string) bool
}

// RenderContext contains prepared axis, tick, label, and style metadata passed to a Renderer.
type RenderContext struct {
	LabelWidth     int
	LabelStyle     lipgloss.Style
	RowLabel       map[int]string
	HGrid          []bool
	VGrid          []bool
	Names          []string
	Styles         []lipgloss.Style
	NegativeStyles []lipgloss.Style
	HasNegStyles   []bool
	SeriesData     [][]float64
	GridStyle      lipgloss.Style
	ZeroRow        int
	// ThresholdStyles maps a row to the threshold style of series cells in it. It is empty when threshold colors are off.
	ThresholdStyles map[int]lipgloss.Style
	// ThresholdLines maps a row to its threshold line: one rendered cell per plot column.
	// A wide cell is followed by empty cells, and the first cell is always a dash.
	ThresholdLines map[int][]string
}

// isNegativeRow reports whether terminal row r represents negative values.
func (ctx RenderContext) isNegativeRow(r int, maxVal float64) bool {
	return (ctx.ZeroRow >= 0 && r > ctx.ZeroRow) || (ctx.ZeroRow < 0 && maxVal < 0)
}

// Renderer defines how chart series data and axes are rendered into a string.
// Custom renderers only need the Chart contract: build one with New and Push
// in tests, no mocks required.
type Renderer interface {
	Render(c Chart, ctx RenderContext) string
}

func spanOf(min, max float64) float64 {
	if max == min {
		return 1
	}
	return max - min
}

// isGap reports whether v is a missing point (NaN or ±Inf) that is drawn as a gap.
func isGap(v float64) bool {
	return math.IsNaN(v) || math.IsInf(v, 0)
}

func normOf(min, max, v float64) float64 {
	norm := (v - min) / spanOf(min, max)
	if norm < 0 {
		return 0
	}
	if norm > 1 {
		return 1
	}
	return norm
}

// tickRowOf returns the terminal row index (0 at top to h-1 at bottom) for value v.
func tickRowOf(h int, min, max, v float64) int {
	return int(math.Round(float64(h-1) * (1.0 - normOf(min, max, v))))
}

// callbackCovers reports whether the callback draws a cell in columns col+1 to col+width-1.
func callbackCovers(cellAt func(r, c int) (string, bool), r, col, width int) bool {
	for k := 1; k < width; k++ {
		if _, ok := cellAt(r, col+k); ok {
			return true
		}
	}
	return false
}

func renderGrid(c Chart, ctx RenderContext, cellAt func(r, c int) (string, bool)) string {
	w, h := c.Width(), c.Height()
	axisNormal := c.AxisStyle().Render("│")
	axisZero := c.AxisStyle().Render("├")
	isZeroRow := c.ZeroBaseline() && ctx.ZeroRow >= 0
	var sb strings.Builder
	for r := 0; r < h; r++ {
		sb.WriteString(ctx.LabelStyle.Render(ctx.RowLabel[r]))
		if isZeroRow && r == ctx.ZeroRow {
			sb.WriteString(axisZero)
		} else {
			sb.WriteString(axisNormal)
		}
		for col := 0; col < w; col++ {
			if s, ok := cellAt(r, col); ok {
				sb.WriteString(s)
				if sw := lipgloss.Width(s); sw > 1 {
					col += sw - 1
				}
				continue
			}
			if cells := ctx.ThresholdLines[r]; col < len(cells) {
				cell := cells[col]
				sw := lipgloss.Width(cell)
				if cell == "" || (sw > 1 && callbackCovers(cellAt, r, col, sw)) {
					cell, sw = cells[0], 1
				}
				sb.WriteString(cell)
				if sw > 1 {
					col += sw - 1
				}
				continue
			}
			if isZeroRow && r == ctx.ZeroRow {
				if ctx.VGrid[col] {
					sb.WriteString(c.AxisStyle().Render("┼"))
				} else {
					sb.WriteString(c.AxisStyle().Render("─"))
				}
				continue
			}
			if ctx.HGrid[r] || ctx.VGrid[col] {
				sb.WriteString(ctx.GridStyle.Render(gridCell(ctx.HGrid[r], ctx.VGrid[col])))
				continue
			}
			sb.WriteByte(' ')
		}
		if r < h-1 {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

// styleFor picks the style of a series cell in row r: the negative style of its series,
// else the threshold style of the row, else the series style.
func styleFor(owner, r int, isNeg bool, ctx RenderContext, fallback lipgloss.Style) lipgloss.Style {
	known := owner >= 0 && owner < len(ctx.Styles)
	if known && isNeg && owner < len(ctx.HasNegStyles) && ctx.HasNegStyles[owner] {
		return ctx.NegativeStyles[owner]
	}
	if st, ok := ctx.ThresholdStyles[r]; ok {
		return st
	}
	if known {
		return ctx.Styles[owner]
	}
	return fallback
}

func faintStyleFor(owner, r int, isNeg bool, ctx RenderContext, fallback lipgloss.Style) lipgloss.Style {
	return styleFor(owner, r, isNeg, ctx, fallback).Faint(true)
}

func gridCell(h, v bool) string {
	switch {
	case h && v:
		return "┼"
	case h:
		return "─"
	default:
		return "│"
	}
}

func makeCells[T any](h, w int, fill T) [][]T {
	buf := make([]T, h*w)
	for i := range buf {
		buf[i] = fill
	}
	cells := make([][]T, h)
	for i := range cells {
		cells[i] = buf[i*w : (i+1)*w]
	}
	return cells
}

func tintColorFor(c Chart, st lipgloss.Style) color.Color {
	if tc := c.TintColor(); tc != nil && tc != (lipgloss.NoColor{}) {
		return tc
	}
	if bg := st.GetBackground(); bg != (lipgloss.NoColor{}) {
		return bg
	}
	return DefaultTintColor
}
