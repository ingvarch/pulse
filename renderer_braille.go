package pulse

import (
	"math"

	"charm.land/lipgloss/v2"
)

// BrailleRenderer renders 2x4 dot Braille patterns.
type BrailleRenderer struct{}

var _ Renderer = BrailleRenderer{}

var brailleBits = [4][2]uint8{
	{0x01, 0x08},
	{0x02, 0x10},
	{0x04, 0x20},
	{0x40, 0x80},
}

// dotCellOf maps a 2w x 4h coordinate into a Braille cell and its internal bit.
func dotCellOf(w, h, xd, yd int) (row, col int, bit uint8, ok bool) {
	dw, dh := w*2, h*4
	if xd < 0 || yd < 0 || xd >= dw || yd >= dh {
		return 0, 0, 0, false
	}
	rowFromBottom := yd / 4
	dy := yd % 4
	cx, dx := xd/2, xd%2
	return h - 1 - rowFromBottom, cx, brailleBits[dy][dx], true
}

func plotDots(grid [][]uint8, owner [][]int, idx, lineWidth, w, h, xd, yd int) {
	pts := [][2]int{{xd, yd}}
	if lineWidth > 1 {
		pts = append(pts, [2]int{xd + 1, yd}, [2]int{xd, yd + 1}, [2]int{xd + 1, yd + 1})
	}
	for _, p := range pts {
		row, col, bit, ok := dotCellOf(w, h, p[0], p[1])
		if !ok {
			continue
		}
		grid[row][col] |= bit
		owner[row][col] = idx
	}
}

func dotHOf(h int) int {
	if dh := h*4 - 1; dh > 0 {
		return dh
	}
	return 0
}

// dotYOf maps a value to dot row offset from the bottom (for Braille).
func dotYOf(h int, min, max, v float64) int {
	return int(math.Round(normOf(min, max, v) * float64(dotHOf(h))))
}

type dot struct{ x, y int }

// catmullRom computes a Catmull-Rom spline value for segment p1-p2 with neighbors p0, p3.
func catmullRom(p0, p1, p2, p3, t float64) float64 {
	t2 := t * t
	t3 := t2 * t
	return 0.5 * ((2 * p1) + (-p0+p2)*t + (2*p0-5*p1+4*p2-p3)*t2 + (-p0+3*p1-3*p2+p3)*t3)
}

func clampDotOf(w, h, xd, yd int) (int, int) {
	maxX := w*2 - 1
	if maxX < 0 {
		maxX = 0
	}
	if xd < 0 {
		xd = 0
	}
	if xd > maxX {
		xd = maxX
	}
	dh := dotHOf(h)
	if yd < 0 {
		yd = 0
	}
	if yd > dh {
		yd = dh
	}
	return xd, yd
}

// sampleLineOf interpolates the polyline into dots: Catmull-Rom spline if smooth, otherwise line segments.
func sampleLineOf(w, h int, smooth bool, xs, ys []int) []dot {
	out := []dot{{xs[0], ys[0]}}
	n := len(xs)
	for i := 0; i < n-1; i++ {
		x0, y0, x1, y1 := float64(xs[i]), float64(ys[i]), float64(xs[i+1]), float64(ys[i+1])
		var px0, py0, px3, py3 float64
		if i == 0 {
			px0, py0 = x0, y0
		} else {
			px0, py0 = float64(xs[i-1]), float64(ys[i-1])
		}
		if i+2 >= n {
			px3, py3 = x1, y1
		} else {
			px3, py3 = float64(xs[i+2]), float64(ys[i+2])
		}
		dx, dy := x1-x0, y1-y0
		steps := 1
		if smooth {
			steps = int((math.Abs(dx) + math.Abs(dy)) * 2)
			if steps < 8 {
				steps = 8
			}
			if steps > 128 {
				steps = 128
			}
		} else {
			steps = int(math.Abs(dx))
			if a := int(math.Abs(dy)); a > steps {
				steps = a
			}
			if steps < 1 {
				steps = 1
			}
		}
		for s := 1; s <= steps; s++ {
			t := float64(s) / float64(steps)
			var fx, fy float64
			if smooth {
				fx = catmullRom(px0, x0, x1, px3, t)
				fy = catmullRom(py0, y0, y1, py3, t)
			} else {
				fx = x0 + dx*t
				fy = y0 + dy*t
			}
			xd, yd := clampDotOf(w, h, int(math.Round(fx)), int(math.Round(fy)))
			out = append(out, dot{xd, yd})
		}
	}
	return out
}

func plotSeriesBraille(c Chart, grid [][]uint8, owner [][]int, idx int, data []float64) []int {
	w, h := c.Width(), c.Height()
	min, max := c.Min(), c.Max()
	colTop := make([]int, w*2)
	for i := range colTop {
		colTop[i] = -1
	}
	n := len(data)
	if n == 0 {
		return colTop
	}
	window := w
	if window < 1 {
		window = 1
	}
	dw := w*2 - 1
	if dw < 0 {
		dw = 0
	}
	den := window - 1
	if den < 1 {
		den = 1
	}
	xs := make([]int, n)
	ys := make([]int, n)
	for j, v := range data {
		xd := (window - n + j) * dw / den
		xs[j], ys[j] = xd, dotYOf(h, min, max, v)
	}
	top := func(xd, yd int) {
		if xd >= 0 && xd < len(colTop) && yd > colTop[xd] {
			colTop[xd] = yd
		}
	}
	for _, d := range sampleLineOf(w, h, c.Smooth(), xs, ys) {
		plotDots(grid, owner, idx, c.LineWidth(), w, h, d.x, d.y)
		top(d.x, d.y)
	}
	return colTop
}

// Render implements Renderer for BrailleRenderer.
func (BrailleRenderer) Render(c Chart, ctx RenderContext) string {
	w, h := c.Width(), c.Height()
	grid := makeCells(h, w, uint8(0))
	owner := makeCells(h, w, -1)
	fillMask := makeCells(h, w, uint8(0))
	fillOwner := makeCells(h, w, -1)

	base := dotYOf(h, c.Min(), c.Max(), 0)
	for idx, name := range ctx.Names {
		colTop := plotSeriesBraille(c, grid, owner, idx, c.SeriesData(name))
		if !c.Fill() {
			continue
		}
		for xd, top := range colTop {
			for yd := base; yd < top; yd++ {
				if xd%2 != 0 || yd%2 != 0 {
					continue
				}
				row, col, bit, ok := dotCellOf(w, h, xd, yd)
				if !ok {
					continue
				}
				fillMask[row][col] |= bit
				fillOwner[row][col] = idx
			}
		}
	}

	visibleEvents := c.VisibleEvents()
	eventMap := make(map[int]VisibleEvent, len(visibleEvents))
	for _, ve := range visibleEvents {
		eventMap[ve.Col] = ve
	}

	fallback := c.LineStyle()
	tinted := c.TintedFill()
	return renderGrid(c, ctx, func(r, col int) (string, bool) {
		ve, hasEvent := eventMap[col]
		if hasEvent && r == 0 {
			glyph := ve.Event.Glyph
			if glyph == "" {
				glyph = "▼"
			}
			st := ve.Event.Style
			if st.GetForeground() == nil || st.GetForeground() == (lipgloss.NoColor{}) {
				st = st.Foreground(lipgloss.Color("#7dcfff")).Bold(true)
			}
			return st.Render(glyph), true
		}
		lineDots := grid[r][col]
		fillDots := fillMask[r][col]
		dots := lineDots | fillDots
		if dots != 0 {
			if lineDots != 0 {
				st := styleFor(owner[r][col], ctx.Styles, fallback)
				if tinted {
					st = st.Background(tintColorFor(c, st))
				}
				return st.Render(string(rune(0x2800 + int(dots)))), true
			}
			ownerIdx := fillOwner[r][col]
			st := faintStyleFor(ownerIdx, ctx.Styles, fallback)
			if tinted {
				baseStyle := styleFor(ownerIdx, ctx.Styles, fallback)
				st = st.Background(tintColorFor(c, baseStyle))
			}
			return st.Render(string(rune(0x2800 + int(dots)))), true
		}
		if hasEvent && !ve.Event.NoLine {
			guideStyle := ve.Event.Style
			if guideStyle.GetForeground() == nil || guideStyle.GetForeground() == (lipgloss.NoColor{}) {
				guideStyle = guideStyle.Foreground(lipgloss.Color("#7dcfff")).Faint(true)
			}
			return guideStyle.Render("┆"), true
		}
		return "", false
	})
}
