package pulse

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// LinesRenderer renders smooth box-drawing lines with optional area fill.
type LinesRenderer struct{}

var _ Renderer = LinesRenderer{}

const (
	armUp    = 1
	armRight = 2
	armDown  = 4
	armLeft  = 8
)

type lineGlyphs struct {
	vert, horiz    rune
	dr, dl, ur, ul rune
	tr, tl, td, tu rune
	cross          rune
}

var (
	glyphsThinRounded = lineGlyphs{
		vert: '│', horiz: '─',
		dr: '╭', dl: '╮', ur: '╰', ul: '╯',
		tr: '├', tl: '┤', td: '┬', tu: '┴',
		cross: '┼',
	}
	glyphsThinSharp = lineGlyphs{
		vert: '│', horiz: '─',
		dr: '┌', dl: '┐', ur: '└', ul: '┘',
		tr: '├', tl: '┤', td: '┬', tu: '┴',
		cross: '┼',
	}
	glyphsBoldRounded = lineGlyphs{
		vert: '┃', horiz: '━',
		dr: '┏', dl: '┓', ur: '┗', ul: '┛',
		tr: '┣', tl: '┫', td: '┳', tu: '┻',
		cross: '╋',
	}
	glyphsBoldSharp = lineGlyphs{
		vert: '┃', horiz: '━',
		dr: '┌', dl: '┐', ur: '└', ul: '┘',
		tr: '├', tl: '┤', td: '┬', tu: '┴',
		cross: '┼',
	}
)

func glyphsFor(lineWidth int, smooth bool) lineGlyphs {
	if lineWidth > 1 {
		if !smooth {
			return glyphsBoldSharp
		}
		return glyphsBoldRounded
	}
	if smooth {
		return glyphsThinRounded
	}
	return glyphsThinSharp
}

func runeForMask(lineWidth int, smooth bool, mask uint8) rune {
	g := glyphsFor(lineWidth, smooth)
	switch mask {
	case armUp, armDown, armUp | armDown:
		return g.vert
	case armLeft, armRight, armLeft | armRight:
		return g.horiz
	case armDown | armRight:
		return g.dr
	case armDown | armLeft:
		return g.dl
	case armUp | armRight:
		return g.ur
	case armUp | armLeft:
		return g.ul
	case armUp | armDown | armRight:
		return g.tr
	case armUp | armDown | armLeft:
		return g.tl
	case armDown | armLeft | armRight:
		return g.td
	case armUp | armLeft | armRight:
		return g.tu
	case armUp | armDown | armLeft | armRight:
		return g.cross
	default:
		return g.horiz
	}
}

// plotSeriesLines renders a smooth continuous box-drawing line and shaded area fill ░.
func plotSeriesLines(c Chart, lineMask [][]uint8, lineOwner [][]int, fillMask [][]bool, fillOwner [][]int, idx int, data []float64) {
	w, h := c.Width(), c.Height()
	min, max := c.Min(), c.Max()
	n := len(data)
	if n == 0 {
		return
	}
	window := w
	startCol := window - n
	if startCol < 0 {
		startCol = 0
		data = data[n-window:]
		n = window
	}

	ys := make([]int, window)
	for j, v := range data {
		col := startCol + j
		ys[col] = tickRowOf(h, min, max, v)
	}

	for col := startCol + 1; col < window; col++ {
		y0 := ys[col-1]
		y1 := ys[col]

		// Horizontal segment at row y0 between col-1 and col
		lineMask[y0][col-1] |= armRight
		lineOwner[y0][col-1] = idx

		lineMask[y0][col] |= armLeft
		lineOwner[y0][col] = idx

		// Vertical transition in col between y0 and y1
		if y0 > y1 { // line goes upwards (row index decreases)
			for y := y1; y < y0; y++ {
				lineMask[y][col] |= armDown
				lineOwner[y][col] = idx

				lineMask[y+1][col] |= armUp
				lineOwner[y+1][col] = idx
			}
		} else if y0 < y1 { // line goes downwards (row index increases)
			for y := y0; y < y1; y++ {
				lineMask[y][col] |= armDown
				lineOwner[y][col] = idx

				lineMask[y+1][col] |= armUp
				lineOwner[y+1][col] = idx
			}
		}
	}

	// Terminate endpoints
	if n == 1 {
		lineMask[ys[startCol]][startCol] |= (armLeft | armRight)
		lineOwner[ys[startCol]][startCol] = idx
	} else if ys[startCol] >= 0 && ys[startCol] < h {
		lineMask[ys[startCol]][startCol] |= armLeft
		lineOwner[ys[startCol]][startCol] = idx
	}

	// Baseline for area fill (0 by default)
	baseRow := tickRowOf(h, min, max, 0)

	// Shaded area fill ░ below the line (strictly beneath the line's lower boundary in each column)
	if c.Fill() {
		for col := startCol; col < window; col++ {
			bottomY := -1
			for r := h - 1; r >= 0; r-- {
				if lineOwner[r][col] == idx {
					bottomY = r
					break
				}
			}
			if bottomY < 0 {
				bottomY = ys[col]
			}
			lo, hi := bottomY+1, baseRow
			if lo > hi {
				lo, hi = baseRow, bottomY-1
			}
			for y := lo; y <= hi && y < h; y++ {
				if y >= 0 {
					fillMask[y][col] = true
					fillOwner[y][col] = idx
				}
			}
		}
	}
}

// Render implements Renderer for LinesRenderer.
func (LinesRenderer) Render(c Chart, ctx RenderContext) string {
	w, h := c.Width(), c.Height()
	lineMask := makeCells(h, w, uint8(0))
	lineOwner := makeCells(h, w, -1)
	fillMask := makeCells(h, w, false)
	fillOwner := makeCells(h, w, -1)

	for idx, name := range ctx.Names {
		var data []float64
		if idx < len(ctx.SeriesData) && ctx.SeriesData[idx] != nil {
			data = ctx.SeriesData[idx]
		} else {
			data = c.SeriesData(name)
		}
		plotSeriesLines(c, lineMask, lineOwner, fillMask, fillOwner, idx, data)
	}

	eventMap := eventsByColumn(c.VisibleEvents())

	fallback, glyphWidth, glyphSmooth := c.LineStyle(), c.LineWidth(), c.Smooth()
	tinted := c.TintedFill()
	solid := c.SolidFill()
	return renderGrid(c, ctx, func(r, col int) (string, bool) {
		ve, hasEvent := eventMap[col]
		if hasEvent && r == 0 {
			return renderEventPin(ve), true
		}
		if lineMask[r][col] != 0 {
			st := styleFor(lineOwner[r][col], ctx.Styles, fallback)
			if tinted {
				st = st.Background(tintColorFor(c, st))
			}
			ru := runeForMask(glyphWidth, glyphSmooth, lineMask[r][col])
			return st.Render(string(ru)), true
		}
		if hasEvent {
			var bg color.Color
			if fillMask[r][col] && tinted {
				baseStyle := styleFor(fillOwner[r][col], ctx.Styles, fallback)
				bg = tintColorFor(c, baseStyle)
			}
			if s, ok := renderEventGuideline(ve, bg); ok {
				return s, true
			}
		}
		if fillMask[r][col] {
			owner := fillOwner[r][col]
			st := faintStyleFor(owner, ctx.Styles, fallback)
			if tinted {
				baseStyle := styleFor(owner, ctx.Styles, fallback)
				bg := tintColorFor(c, baseStyle)
				if solid {
					return lipgloss.NewStyle().Background(bg).Render(" "), true
				}
				return st.Background(bg).Render("░"), true
			}
			if solid {
				return " ", true
			}
			return st.Render("░"), true
		}
		return "", false
	})
}
