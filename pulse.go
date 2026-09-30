package pulse

import (
	"fmt"
	"math"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/ingvarch/pulse/scale"
)

// RenderMode defines how line curves are rendered on the chart.
type RenderMode int

const (
	// ModeLines renders smooth continuous lines (box-drawing ╭, ╮, ╯, ╰, ─, │) with Grafana-style area fill ░.
	ModeLines RenderMode = iota
	// ModeBraille renders using a 2x4 dot braille matrix (Unicode Braille Patterns).
	ModeBraille
)

// Model is a streaming terminal line chart.
type Model struct {
	w, h     int
	min, max float64

	series map[string]*series
	order  []string

	grid      bool
	fill      bool
	smooth    bool
	lineWidth int
	mode      RenderMode

	lineStyle lipgloss.Style
	axisStyle lipgloss.Style

	customTicks []float64

	labelWidth     int
	labelFormatter func(float64) string
}

type series struct {
	data     []float64
	style    lipgloss.Style
	styled   bool
	last     float64
	hasValue bool
}

// Option configures a Model.
type Option func(*Model)

// WithRange sets the fixed Y range. For example, WithRange(0, 100) for CPU/RAM.
func WithRange(min, max float64) Option {
	return func(m *Model) {
		m.min, m.max = min, max
	}
}

// WithLineStyle sets the default line style.
func WithLineStyle(s lipgloss.Style) Option {
	return func(m *Model) { m.lineStyle = s }
}

// WithAxisStyle sets the axes and labels style.
func WithAxisStyle(s lipgloss.Style) Option {
	return func(m *Model) { m.axisStyle = s }
}

// WithSeriesStyle sets the style for a named series.
func WithSeriesStyle(name string, s lipgloss.Style) Option {
	return func(m *Model) { m.SetSeriesStyle(name, s) }
}

// WithGrid enables or disables gridlines at tick positions.
func WithGrid(on bool) Option {
	return func(m *Model) { m.grid = on }
}

// WithFill enables or disables area fill under the line.
func WithFill(on bool) Option {
	return func(m *Model) { m.fill = on }
}

// WithSmooth enables or disables line smoothing (rounded corners or spline).
func WithSmooth(on bool) Option {
	return func(m *Model) { m.smooth = on }
}

// WithLineWidth sets line width: 1 thin, 2 bold.
func WithLineWidth(w int) Option {
	return func(m *Model) { m.SetLineWidth(w) }
}

// WithMode sets rendering mode (ModeLines or ModeBraille).
func WithMode(mode RenderMode) Option {
	return func(m *Model) { m.mode = mode }
}

// WithLines enables continuous line rendering mode.
func WithLines() Option {
	return func(m *Model) { m.mode = ModeLines }
}

// WithTicks sets explicit Y-axis tick values (e.g. 0, 25, 50, 75, 100 or 0, 50, 100).
func WithTicks(ticks ...float64) Option {
	return func(m *Model) { m.customTicks = ticks }
}

// WithLabelFormatter sets a custom formatter for Y-axis tick values.
func WithLabelFormatter(fn func(float64) string) Option {
	return func(m *Model) { m.labelFormatter = fn }
}

// WithLabelWidth sets an explicit width for Y-axis labels.
func WithLabelWidth(w int) Option {
	return func(m *Model) { m.labelWidth = w }
}

// WithBraille enables braille dot matrix mode (2x4 dots).
func WithBraille() Option {
	return func(m *Model) { m.mode = ModeBraille }
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
		w:         w,
		h:         h,
		min:       0,
		max:       100,
		series:    map[string]*series{},
		grid:      true,
		fill:      true,
		smooth:    true,
		lineWidth: 2,
		mode:      ModeLines,
	}
	for _, o := range opts {
		o(m)
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
	s.last, s.hasValue = v, true
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
func (m *Model) SetTicks(ticks ...float64) {
	m.customTicks = ticks
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
func (m *Model) SetMode(mode RenderMode) {
	m.mode = mode
}

// ToggleRenderMode switches between ModeLines and ModeBraille.
func (m *Model) ToggleRenderMode() {
	if m.mode == ModeLines {
		m.mode = ModeBraille
	} else {
		m.mode = ModeLines
	}
}

// Mode returns the current rendering mode.
func (m *Model) Mode() RenderMode {
	return m.mode
}

func (m *Model) seriesStyle(name string) lipgloss.Style {
	if s, ok := m.series[name]; ok && s.styled {
		return s.style
	}
	return m.lineStyle
}

// Legend returns a formatted legend for named series, or empty string if none exist.
func (m *Model) Legend() string {
	parts := make([]string, 0, len(m.order))
	for _, name := range m.order {
		swatch := m.seriesStyle(name).Render("■")
		parts = append(parts, swatch+" "+name)
	}
	return strings.Join(parts, "  ")
}

// Last returns the most recent value of a named series.
func (m *Model) Last(name string) (float64, bool) {
	s, ok := m.series[name]
	if !ok || !s.hasValue {
		return 0, false
	}
	return s.last, true
}

// LegendBox returns a boxed legend with borders, swatches, series names, and recent values.
func (m *Model) LegendBox() string {
	if len(m.order) == 0 {
		return ""
	}
	lines := make([]string, 0, len(m.order))
	for _, name := range m.order {
		swatch := m.seriesStyle(name).Render("■")
		value := ""
		if v, ok := m.Last(name); ok {
			if m.labelFormatter != nil {
				value = m.labelFormatter(v)
			} else {
				value = scale.FormatTick(v)
			}
		}
		lines = append(lines, swatch+" "+m.axisStyle.Render(name+" "+value))
	}
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Render(strings.Join(lines, "\n"))
}

var brailleBits = [4][2]uint8{
	{0x01, 0x08},
	{0x02, 0x10},
	{0x04, 0x20},
	{0x40, 0x80},
}

// dotCell maps a 2w x 4h coordinate into a Braille cell and its internal bit.
func (m *Model) dotCell(xd, yd int) (row, col int, bit uint8, ok bool) {
	dw, dh := m.w*2, m.h*4
	if xd < 0 || yd < 0 || xd >= dw || yd >= dh {
		return 0, 0, 0, false
	}
	rowFromBottom := yd / 4
	dy := yd % 4
	cx, dx := xd/2, xd%2
	return m.h - 1 - rowFromBottom, cx, brailleBits[dy][dx], true
}

func (m *Model) setDot(grid [][]uint8, owner [][]int, idx, xd, yd int) {
	pts := [][2]int{{xd, yd}}
	if m.lineWidth > 1 {
		pts = append(pts, [2]int{xd + 1, yd}, [2]int{xd, yd + 1}, [2]int{xd + 1, yd + 1})
	}
	for _, p := range pts {
		row, col, bit, ok := m.dotCell(p[0], p[1])
		if !ok {
			continue
		}
		grid[row][col] |= bit
		owner[row][col] = idx
	}
}

func (m *Model) span() float64 {
	if m.max == m.min {
		return 1
	}
	return m.max - m.min
}

func (m *Model) dotH() int {
	if dh := m.h*4 - 1; dh > 0 {
		return dh
	}
	return 0
}

func (m *Model) norm(v float64) float64 {
	norm := (v - m.min) / m.span()
	if norm < 0 {
		return 0
	}
	if norm > 1 {
		return 1
	}
	return norm
}

// dotY maps a value to dot row offset from the bottom (for Braille).
func (m *Model) dotY(v float64) int {
	return int(math.Round(m.norm(v) * float64(m.dotH())))
}

// tickRow returns the terminal row index (0 at top to h-1 at bottom) for value v.
func (m *Model) tickRow(v float64) int {
	return int(math.Round(float64(m.h-1) * (1.0 - m.norm(v))))
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

type dot struct{ x, y int }

// catmullRom computes a Catmull-Rom spline value for segment p1-p2 with neighbors p0, p3.
func catmullRom(p0, p1, p2, p3, t float64) float64 {
	t2 := t * t
	t3 := t2 * t
	return 0.5 * ((2 * p1) + (-p0+p2)*t + (2*p0-5*p1+4*p2-p3)*t2 + (-p0+3*p1-3*p2+p3)*t3)
}

func (m *Model) clampDot(xd, yd int) (int, int) {
	maxX := m.w*2 - 1
	if maxX < 0 {
		maxX = 0
	}
	if xd < 0 {
		xd = 0
	}
	if xd > maxX {
		xd = maxX
	}
	if yd < 0 {
		yd = 0
	}
	if yd > m.dotH() {
		yd = m.dotH()
	}
	return xd, yd
}

// sampleLine interpolates the polyline into dots: Catmull-Rom spline if smooth, otherwise line segments.
func (m *Model) sampleLine(xs, ys []int) []dot {
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
		if m.smooth {
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
			if m.smooth {
				fx = catmullRom(px0, x0, x1, px3, t)
				fy = catmullRom(py0, y0, y1, py3, t)
			} else {
				fx = x0 + dx*t
				fy = y0 + dy*t
			}
			xd, yd := m.clampDot(int(math.Round(fx)), int(math.Round(fy)))
			out = append(out, dot{xd, yd})
		}
	}
	return out
}

func (m *Model) plotSeriesBraille(grid [][]uint8, owner [][]int, idx int, data []float64) []int {
	colTop := make([]int, m.w*2)
	for i := range colTop {
		colTop[i] = -1
	}
	n := len(data)
	if n == 0 {
		return colTop
	}
	window := m.w
	if window < 1 {
		window = 1
	}
	dw := m.w*2 - 1
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
		xs[j], ys[j] = xd, m.dotY(v)
	}
	top := func(xd, yd int) {
		if xd >= 0 && xd < len(colTop) && yd > colTop[xd] {
			colTop[xd] = yd
		}
	}
	for _, d := range m.sampleLine(xs, ys) {
		m.setDot(grid, owner, idx, d.x, d.y)
		top(d.x, d.y)
	}
	return colTop
}

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

func (m *Model) glyphs() lineGlyphs {
	if m.lineWidth > 1 {
		if !m.smooth {
			return glyphsBoldSharp
		}
		return glyphsBoldRounded
	}
	if m.smooth {
		return glyphsThinRounded
	}
	return glyphsThinSharp
}

func (m *Model) runeForMask(mask uint8) rune {
	g := m.glyphs()
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
func (m *Model) plotSeriesLines(lineMask [][]uint8, lineOwner [][]int, fillMask [][]bool, fillOwner [][]int, idx int, data []float64) {
	n := len(data)
	if n == 0 {
		return
	}
	window := m.w
	startCol := window - n
	if startCol < 0 {
		startCol = 0
		data = data[n-window:]
		n = window
	}

	ys := make([]int, window)
	for j, v := range data {
		c := startCol + j
		ys[c] = int(math.Round(float64(m.h-1) * (1.0 - m.norm(v))))
	}

	for c := startCol + 1; c < window; c++ {
		y0 := ys[c-1]
		y1 := ys[c]

		// Horizontal segment at row y0 between col c-1 and col c
		lineMask[y0][c-1] |= armRight
		lineOwner[y0][c-1] = idx

		lineMask[y0][c] |= armLeft
		lineOwner[y0][c] = idx

		// Vertical transition in col c between y0 and y1
		if y0 > y1 { // line goes upwards (row index decreases)
			for y := y1; y < y0; y++ {
				lineMask[y][c] |= armDown
				lineOwner[y][c] = idx

				lineMask[y+1][c] |= armUp
				lineOwner[y+1][c] = idx
			}
		} else if y0 < y1 { // line goes downwards (row index increases)
			for y := y0; y < y1; y++ {
				lineMask[y][c] |= armDown
				lineOwner[y][c] = idx

				lineMask[y+1][c] |= armUp
				lineOwner[y+1][c] = idx
			}
		}
	}

	// Terminate endpoints
	if n == 1 {
		lineMask[ys[startCol]][startCol] |= (armLeft | armRight)
		lineOwner[ys[startCol]][startCol] = idx
	} else if ys[startCol] >= 0 && ys[startCol] < m.h {
		lineMask[ys[startCol]][startCol] |= armLeft
		lineOwner[ys[startCol]][startCol] = idx
	}

	// Baseline for area fill (0 by default)
	baseRow := int(math.Round(float64(m.h-1) * (1.0 - m.norm(0))))

	// Shaded area fill ░ below the line (strictly beneath the line's lower boundary in each column)
	if m.fill {
		for c := startCol; c < window; c++ {
			bottomY := -1
			for r := m.h - 1; r >= 0; r-- {
				if lineOwner[r][c] == idx {
					bottomY = r
					break
				}
			}
			if bottomY < 0 {
				bottomY = ys[c]
			}
			lo, hi := bottomY+1, baseRow
			if lo > hi {
				lo, hi = baseRow, bottomY-1
			}
			for y := lo; y <= hi && y < m.h; y++ {
				if y >= 0 {
					fillMask[y][c] = true
					fillOwner[y][c] = idx
				}
			}
		}
	}
}

func (m *Model) styleFor(owner int, styles []lipgloss.Style) lipgloss.Style {
	if owner >= 0 && owner < len(styles) {
		return styles[owner]
	}
	return m.lineStyle
}

func (m *Model) faintStyleFor(owner int, styles []lipgloss.Style) lipgloss.Style {
	return m.styleFor(owner, styles).Faint(true)
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
	cells := make([][]T, h)
	for i := range cells {
		cells[i] = make([]T, w)
		for j := range cells[i] {
			cells[i][j] = fill
		}
	}
	return cells
}

// View renders the chart: Y-axis labels + axis border + plot area.
func (m *Model) View() string {
	ticks := m.gridTicks()
	rowLabel := map[int]string{}
	for i := len(ticks) - 1; i >= 0; i-- {
		t := ticks[i]
		if t < m.min || t > m.max {
			continue
		}
		row := m.tickRow(t)
		if _, taken := rowLabel[row]; !taken {
			if m.labelFormatter != nil {
				rowLabel[row] = m.labelFormatter(t)
			} else {
				rowLabel[row] = scale.FormatTick(t)
			}
		}
	}

	labelWidth := m.labelWidth
	if labelWidth <= 0 {
		labelWidth = 4
		for _, lbl := range rowLabel {
			if len(lbl) > labelWidth {
				labelWidth = len(lbl)
			}
		}
	}
	fmtStr := fmt.Sprintf("%%%ds", labelWidth)

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
	for idx, name := range names {
		styles[idx] = m.seriesStyle(name)
	}

	gridStyle := m.axisStyle.Faint(true)

	if m.mode == ModeLines {
		return m.renderLines(fmtStr, rowLabel, hGrid, vGrid, names, styles, gridStyle)
	}
	return m.renderBraille(fmtStr, rowLabel, hGrid, vGrid, names, styles, gridStyle)
}

func (m *Model) renderLines(fmtStr string, rowLabel map[int]string, hGrid, vGrid []bool, names []string, styles []lipgloss.Style, gridStyle lipgloss.Style) string {
	lineMask := makeCells(m.h, m.w, uint8(0))
	lineOwner := makeCells(m.h, m.w, -1)
	fillMask := makeCells(m.h, m.w, false)
	fillOwner := makeCells(m.h, m.w, -1)

	for idx, name := range names {
		s, ok := m.series[name]
		if !ok {
			continue
		}
		m.plotSeriesLines(lineMask, lineOwner, fillMask, fillOwner, idx, s.data)
	}

	var sb strings.Builder
	for r := 0; r < m.h; r++ {
		sb.WriteString(m.axisStyle.Render(fmt.Sprintf(fmtStr, rowLabel[r])))
		sb.WriteString(m.axisStyle.Render("│"))
		for c := 0; c < m.w; c++ {
			if lineMask[r][c] != 0 {
				st := m.styleFor(lineOwner[r][c], styles)
				ru := m.runeForMask(lineMask[r][c])
				sb.WriteString(st.Render(string(ru)))
				continue
			}
			if fillMask[r][c] {
				st := m.faintStyleFor(fillOwner[r][c], styles)
				sb.WriteString(st.Render("░"))
				continue
			}
			if hGrid[r] || vGrid[c] {
				sb.WriteString(gridStyle.Render(gridCell(hGrid[r], vGrid[c])))
				continue
			}
			sb.WriteByte(' ')
		}
		if r < m.h-1 {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

func (m *Model) renderBraille(fmtStr string, rowLabel map[int]string, hGrid, vGrid []bool, names []string, styles []lipgloss.Style, gridStyle lipgloss.Style) string {
	grid := makeCells(m.h, m.w, uint8(0))
	owner := makeCells(m.h, m.w, -1)
	fillMask := makeCells(m.h, m.w, uint8(0))
	fillOwner := makeCells(m.h, m.w, -1)

	base := m.dotY(0)
	for idx, name := range names {
		s, ok := m.series[name]
		if !ok {
			continue
		}
		colTop := m.plotSeriesBraille(grid, owner, idx, s.data)
		if !m.fill {
			continue
		}
		for xd, top := range colTop {
			for yd := base; yd < top; yd++ {
				if xd%2 != 0 || yd%2 != 0 {
					continue
				}
				row, col, bit, ok := m.dotCell(xd, yd)
				if !ok {
					continue
				}
				fillMask[row][col] |= bit
				fillOwner[row][col] = idx
			}
		}
	}

	var sb strings.Builder
	for r := 0; r < m.h; r++ {
		sb.WriteString(m.axisStyle.Render(fmt.Sprintf(fmtStr, rowLabel[r])))
		sb.WriteString(m.axisStyle.Render("│"))
		for c := 0; c < m.w; c++ {
			if grid[r][c] != 0 {
				st := m.styleFor(owner[r][c], styles)
				sb.WriteString(st.Render(string(rune(0x2800 + int(grid[r][c])))))
				continue
			}
			if fillMask[r][c] != 0 {
				st := m.faintStyleFor(fillOwner[r][c], styles)
				sb.WriteString(st.Render(string(rune(0x2800 + int(fillMask[r][c])))))
				continue
			}
			if hGrid[r] || vGrid[c] {
				sb.WriteString(gridStyle.Render(gridCell(hGrid[r], vGrid[c])))
				continue
			}
			sb.WriteByte(' ')
		}
		if r < m.h-1 {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}
