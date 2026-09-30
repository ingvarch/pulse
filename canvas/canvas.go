package canvas

import "strings"

// Canvas is a W x H rune grid. Origin (0,0) is top-left.
type Canvas struct {
	w, h  int
	cells [][]rune
}

// New creates a canvas filled with spaces.
func New(w, h int) *Canvas {
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	cells := make([][]rune, h)
	for y := range cells {
		cells[y] = make([]rune, w)
		for x := range cells[y] {
			cells[y][x] = ' '
		}
	}
	return &Canvas{w: w, h: h, cells: cells}
}

// Set sets a rune at (x, y). Out of bounds coordinates are no-ops.
func (c *Canvas) Set(x, y int, r rune) {
	if x < 0 || y < 0 || x >= c.w || y >= c.h {
		return
	}
	c.cells[y][x] = r
}

// View returns the rendered canvas lines.
func (c *Canvas) View() string {
	var sb strings.Builder
	for y := 0; y < c.h; y++ {
		for x := 0; x < c.w; x++ {
			sb.WriteRune(c.cells[y][x])
		}
		if y < c.h-1 {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}
