package pulse_test

import (
	"math"
	"regexp"
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/ingvarch/pulse"
)

const thW, thH = 20, 11

var (
	thBase = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00"))
	thWarn = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffff00"))
	thCrit = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	thNeg  = lipgloss.NewStyle().Foreground(lipgloss.Color("#0000ff"))
)

func thOpts(mode pulse.RenderMode, opts ...pulse.Option) []pulse.Option {
	return append([]pulse.Option{
		pulse.WithMode(mode),
		pulse.WithRange(0, 100),
		pulse.WithGrid(false),
		pulse.WithFill(false),
		pulse.WithLineWidth(1),
		pulse.WithLineStyle(thBase),
	}, opts...)
}

var cellRe = regexp.MustCompile(`(?:\x1b\[[0-9;]*m)?[^\x1b](?:\x1b\[m)?`)

// cellAt returns the styled plot cell at (row, col) of a rendered chart.
func cellAt(c *pulse.Model, row, col int) string {
	line := strings.Split(c.View(), "\n")[row]
	col += lipgloss.Width(line) - c.Width()
	at := 0
	for _, cell := range cellRe.FindAllString(line, -1) {
		if at == col {
			return cell
		}
		at += lipgloss.Width(ansi.Strip(cell))
	}
	return ""
}

// styledAs reports whether cell is a non-blank glyph drawn in style st.
func styledAs(cell string, st lipgloss.Style) bool {
	g := ansi.Strip(cell)
	return strings.TrimSpace(g) != "" && cell == st.Render(g)
}

// rowOfGlyph returns the only row where column col is not blank.
func rowOfGlyph(t *testing.T, c *pulse.Model, col int) int {
	t.Helper()
	found := -1
	for r, line := range strings.Split(ansi.Strip(c.View()), "\n") {
		if strings.TrimSpace(string([]rune(line)[len([]rune(line))-c.Width()+col])) != "" {
			if found >= 0 {
				t.Fatalf("column %d has glyphs in rows %d and %d", col, found, r)
			}
			found = r
		}
	}
	if found < 0 {
		t.Fatalf("column %d is blank", col)
	}
	return found
}

func TestThresholds_LineOnRowOfValue(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := pulse.New(thW, thH, thOpts(mode,
			pulse.WithGrid(true),
			pulse.WithThresholds(pulse.Threshold{Value: 50, Style: thWarn, Line: true}),
		)...)
		rows := plotRows(m)
		for r, row := range rows {
			if r == 5 {
				if got := string(row); got != strings.Repeat("┄", thW) {
					t.Fatalf("row 5 = %q, want a dashed line", got)
				}
				continue
			}
			if strings.ContainsRune(string(row), '┄') {
				t.Fatalf("row %d has a threshold glyph: %q", r, string(row))
			}
		}
		for col := range thW {
			if c := cellAt(m, 5, col); c != thWarn.Render("┄") {
				t.Fatalf("cell %d = %q, want the threshold style", col, c)
			}
		}
	})
}

func TestThresholds_NoLineWithoutLineFlag(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := pulse.New(thW, thH, thOpts(mode,
			pulse.WithThresholds(pulse.Threshold{Value: 50, Style: thWarn, Label: "WARN"}),
		)...)
		if strings.Contains(ansi.Strip(m.View()), "┄") || strings.Contains(ansi.Strip(m.View()), "WARN") {
			t.Fatalf("a threshold without Line must draw nothing:\n%s", m.View())
		}
	})
}

func TestThresholds_SeriesCoversLine(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := pulse.New(thW, thH, thOpts(mode,
			pulse.WithThresholds(pulse.Threshold{Value: 50, Style: thWarn, Line: true}),
		)...)
		for range 10 {
			m.Push(50)
		}
		row := string(plotRows(m)[5])
		if got := strings.Count(row, "┄"); got != 10 {
			t.Fatalf("row 5 = %q, want the line on the 10 empty columns only", row)
		}
		for col := 10; col < thW; col++ {
			if g := []rune(row)[col]; g == '┄' || g == ' ' {
				t.Fatalf("column %d = %q, want the series", col, string(g))
			}
		}
	})
}

func TestThresholds_LineShowsInGaps(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := pulse.New(thW, thH, thOpts(mode,
			pulse.WithThresholds(pulse.Threshold{Value: 50, Style: thWarn, Line: true}),
		)...)
		for i := range thW {
			if i == 8 {
				m.Push(math.NaN())
				continue
			}
			m.Push(50)
		}
		if g := plotRows(m)[5][8]; g != '┄' {
			t.Fatalf("gap cell = %q, want the threshold line", string(g))
		}
	})
}

func TestThresholds_HighestValueWinsOnSameRow(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		for _, order := range [][2]pulse.Threshold{
			{{Value: 50, Style: thWarn, Line: true}, {Value: 51, Style: thCrit, Line: true}},
			{{Value: 51, Style: thCrit, Line: true}, {Value: 50, Style: thWarn, Line: true}},
		} {
			m := pulse.New(thW, thH, thOpts(mode, pulse.WithThresholds(order[0], order[1]))...)
			if c := cellAt(m, 5, 0); c != thCrit.Render("┄") {
				t.Fatalf("cell = %q, want the style of the higher value", c)
			}
		}
	})
}

func TestThresholds_LineReplacesZeroBaselineNotAxis(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := pulse.New(thW, thH, pulse.WithMode(mode), pulse.WithRange(-50, 50), pulse.WithGrid(false), pulse.WithFill(false), pulse.WithZeroBaseline(true),
			pulse.WithThresholds(pulse.Threshold{Value: 0, Style: thWarn, Line: true}))
		line := strings.Split(ansi.Strip(m.View()), "\n")[5]
		if !strings.Contains(line, "├") || !strings.HasSuffix(line, strings.Repeat("┄", thW)) {
			t.Fatalf("row = %q, want the axis tick and the threshold line", line)
		}
	})
}

func TestThresholds_LabelFitsAtRightEnd(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := pulse.New(thW, thH, thOpts(mode,
			pulse.WithThresholds(pulse.Threshold{Value: 50, Style: thWarn, Line: true, Label: "WARN"}),
		)...)
		if got, want := string(plotRows(m)[5]), strings.Repeat("┄", thW-4)+"WARN"; got != want {
			t.Fatalf("row = %q, want %q", got, want)
		}
		for col := thW - 4; col < thW; col++ {
			if !styledAs(cellAt(m, 5, col), thWarn) {
				t.Fatalf("label cell %d is not in the threshold style: %q", col, cellAt(m, 5, col))
			}
		}
	})
}

func TestThresholds_LabelCoveredBySeries(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := pulse.New(thW, thH, thOpts(mode,
			pulse.WithThresholds(pulse.Threshold{Value: 50, Style: thWarn, Line: true, Label: "WARN"}),
		)...)
		for range thW {
			m.Push(50)
		}
		if row := string(plotRows(m)[5]); strings.ContainsAny(row, "┄WARN") {
			t.Fatalf("row = %q, want the series over line and label", row)
		}
	})
}

func TestThresholds_LabelWidthRules(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		const w = 10
		draw := func(label string) string {
			m := pulse.New(w, thH, thOpts(mode,
				pulse.WithThresholds(pulse.Threshold{Value: 50, Style: thWarn, Line: true, Label: label}),
			)...)
			return plotPart(m, 5)
		}
		if got, want := draw("ABCDEFGH"), "┄┄ABCDEFGH"; got != want {
			t.Fatalf("label of w-2 cells: row = %q, want %q", got, want)
		}
		if got, want := draw("ABCDEFGHI"), strings.Repeat("┄", w); got != want {
			t.Fatalf("label wider than w-2: row = %q, want the line only", got)
		}
		if got, want := draw("日本語本"), "┄┄日本語本"; got != want {
			t.Fatalf("wide runes: row = %q, want %q", got, want)
		}
		if got, want := draw("日本語本人"), strings.Repeat("┄", w); got != want {
			t.Fatalf("wide label of 10 cells: row = %q, want the line only", got)
		}
	})
}

func TestThresholds_OutsideRangeDrawsNothing(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		plain := pulse.New(thW, thH, thOpts(mode)...)
		m := pulse.New(thW, thH, thOpts(mode,
			pulse.WithThresholds(
				pulse.Threshold{Value: 150, Style: thCrit, Line: true, Label: "HIGH"},
				pulse.Threshold{Value: -1, Style: thWarn, Line: true, Label: "LOW"},
			),
			pulse.WithThresholdColors(true),
		)...)
		for _, c := range []*pulse.Model{plain, m} {
			for i := range thW {
				c.Push(float64(i) * 5)
			}
		}
		if got, want := ansi.Strip(m.View()), ansi.Strip(plain.View()); got != want {
			t.Fatalf("view changed:\n%s\nwant\n%s", got, want)
		}
	})
}

func TestThresholdColors_AboveMaxColorsNothing(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := pulse.New(thW, thH, thOpts(mode,
			pulse.WithThresholds(pulse.Threshold{Value: 150, Style: thCrit}),
			pulse.WithThresholdColors(true),
		)...)
		for i := range thW {
			m.Push(98 + float64(i%3))
		}
		drawn := 0
		for col := range thW {
			c := cellAt(m, 0, col)
			if strings.TrimSpace(ansi.Strip(c)) == "" {
				continue
			}
			drawn++
			if !styledAs(c, thBase) {
				t.Fatalf("row 0, column %d = %q, want the series style", col, c)
			}
		}
		if drawn == 0 {
			t.Fatal("row 0 is blank, want the series")
		}
	})
}

func TestThresholds_DoNotChangeAutoRange(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := pulse.New(thW, thH, pulse.WithMode(mode), pulse.WithAutoRange(true),
			pulse.WithThresholds(pulse.Threshold{Value: 5000, Style: thCrit, Line: true}))
		m.Push(10)
		m.View()
		if m.Max() > 20 {
			t.Fatalf("Max = %v, want a range that fits the data only", m.Max())
		}
	})
}

// crossingChart plots 50, 85 and 96 in three plateaus, with warning at 80 and critical at 90.
func crossingChart(mode pulse.RenderMode, opts ...pulse.Option) *pulse.Model {
	m := pulse.New(15, 21, thOpts(mode, append([]pulse.Option{
		pulse.WithThresholds(
			pulse.Threshold{Value: 80, Style: thWarn},
			pulse.Threshold{Value: 90, Style: thCrit},
		),
		pulse.WithThresholdColors(true),
	}, opts...)...)...)
	for _, v := range []float64{50, 85, 96} {
		for range 5 {
			m.Push(v)
		}
	}
	return m
}

func TestThresholdColors_SwitchAtCrossing(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := crossingChart(mode)
		for _, tc := range []struct {
			col  int
			want lipgloss.Style
		}{{2, thBase}, {7, thWarn}, {12, thCrit}} {
			row := rowOfGlyph(t, m, tc.col)
			if c := cellAt(m, row, tc.col); !styledAs(c, tc.want) {
				t.Fatalf("column %d, row %d = %q, want the style %q", tc.col, row, c, tc.want.Render(ansi.Strip(c)))
			}
		}
	})
}

func TestThresholdColors_CellOnThresholdRowTakesItsStyle(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := pulse.New(15, 21, thOpts(mode,
			pulse.WithThresholds(
				pulse.Threshold{Value: 80, Style: thWarn},
				pulse.Threshold{Value: 90, Style: thCrit},
			),
			pulse.WithThresholdColors(true),
		)...)
		for _, v := range []float64{50, 80, 90} {
			for range 5 {
				m.Push(v)
			}
		}
		for _, tc := range []struct {
			col  int
			want lipgloss.Style
		}{{7, thWarn}, {12, thCrit}} {
			row := rowOfGlyph(t, m, tc.col)
			if c := cellAt(m, row, tc.col); !styledAs(c, tc.want) {
				t.Fatalf("column %d, row %d = %q, want the style %q", tc.col, row, c, tc.want.Render(ansi.Strip(c)))
			}
		}
	})
}

func TestThresholdColors_OffKeepsSeriesStyle(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		on := crossingChart(mode)
		off := crossingChart(mode, pulse.WithThresholdColors(false))
		plain := pulse.New(15, 21, thOpts(mode)...)
		for _, v := range []float64{50, 85, 96} {
			for range 5 {
				plain.Push(v)
			}
		}
		if off.View() != plain.View() {
			t.Fatalf("colors off must leave the series styles alone:\n%s", off.View())
		}
		if on.View() == plain.View() {
			t.Fatal("colors on must change the view")
		}
	})
}

func TestThresholdColors_FillUsesFaintThresholdStyle(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := crossingChart(mode, pulse.WithFill(true))
		var crit, warn, base bool
		for r := range 21 {
			c := cellAt(m, r, 12)
			g := ansi.Strip(c)
			if strings.TrimSpace(g) == "" {
				continue
			}
			switch c {
			case thCrit.Faint(true).Render(g):
				crit = true
			case thWarn.Faint(true).Render(g):
				warn = true
			case thBase.Faint(true).Render(g):
				base = true
			}
		}
		if !crit || !warn || !base {
			t.Fatalf("column 12 fill: critical=%v warning=%v series=%v, want all", crit, warn, base)
		}
	})
}

func TestThresholdColors_TintedFillTakesTintFromThresholdStyle(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		bg := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).Background(lipgloss.Color("#300000"))
		m := pulse.New(15, 21, thOpts(mode,
			pulse.WithFill(true), pulse.WithTintedFill(true),
			pulse.WithThresholds(pulse.Threshold{Value: 90, Style: bg}),
			pulse.WithThresholdColors(true),
		)...)
		for range 15 {
			m.Push(96)
		}
		const tint = "48;2;48;0;0"
		if c := cellAt(m, 2, 5); !strings.Contains(c, tint) {
			t.Fatalf("fill cell in a threshold row must take the threshold tint: %q", c)
		}
		if c := cellAt(m, 8, 5); strings.Contains(c, tint) || !strings.Contains(c, "48;2;") {
			t.Fatalf("fill cell below the thresholds must keep the series tint: %q", c)
		}
	})
}

func TestThresholdColors_ZeroBaselineJunctions(t *testing.T) {
	m := pulse.New(thW, thH, pulse.WithMode(pulse.ModeLines), pulse.WithRange(-50, 50), pulse.WithGrid(false), pulse.WithLineWidth(1),
		pulse.WithFill(true), pulse.WithZeroBaseline(true), pulse.WithLineStyle(thBase),
		pulse.WithThresholds(pulse.Threshold{Value: 0, Style: thWarn}), pulse.WithThresholdColors(true))
	for i := range thW {
		if i < thW/2 {
			m.Push(30)
		} else {
			m.Push(-30)
		}
	}
	for _, tc := range []struct {
		col   int
		glyph string
	}{{4, "┴"}, {15, "┬"}} {
		if c := cellAt(m, 5, tc.col); c != thWarn.Render(tc.glyph) {
			t.Fatalf("zero row, column %d = %q, want %q in the threshold style", tc.col, c, tc.glyph)
		}
	}
}

func TestThresholdColors_EventGuidelineTakesThresholdTint(t *testing.T) {
	bg := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).Background(lipgloss.Color("#300000"))
	m := pulse.New(15, 21, thOpts(pulse.ModeLines,
		pulse.WithFill(true), pulse.WithTintedFill(true),
		pulse.WithThresholds(pulse.Threshold{Value: 90, Style: bg}),
		pulse.WithThresholdColors(true),
	)...)
	for range 15 {
		m.Push(96)
	}
	m.AddEventAt(9, pulse.Event{Label: "x", Glyph: "!"})
	const tint = "48;2;48;0;0"
	if c := cellAt(m, 2, 5); ansi.Strip(c) != "┆" || !strings.Contains(c, tint) {
		t.Fatalf("guideline in a threshold row must take the threshold tint: %q", c)
	}
	if c := cellAt(m, 8, 5); ansi.Strip(c) != "┆" || strings.Contains(c, tint) {
		t.Fatalf("guideline below the thresholds must keep the series tint: %q", c)
	}
}

func negChart(mode pulse.RenderMode, opts ...pulse.Option) *pulse.Model {
	m := pulse.New(10, 21, append([]pulse.Option{
		pulse.WithMode(mode), pulse.WithRange(-50, 50), pulse.WithGrid(false), pulse.WithFill(false), pulse.WithLineWidth(1),
		pulse.WithLineStyle(thBase),
		pulse.WithThresholds(pulse.Threshold{Value: -20, Style: thWarn}),
		pulse.WithThresholdColors(true),
	}, opts...)...)
	for range 5 {
		m.Push(25)
	}
	for range 5 {
		m.Push(-10)
	}
	return m
}

func TestThresholdColors_NegativeStyleWins(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		withNeg := negChart(mode, pulse.WithNegativeStyle(thNeg))
		without := negChart(mode)
		for _, tc := range []struct {
			name string
			m    *pulse.Model
			col  int
			want lipgloss.Style
		}{
			{"positive cell above the threshold", withNeg, 2, thWarn},
			{"negative cell with negative style", withNeg, 7, thNeg},
			{"negative cell without negative style", without, 7, thWarn},
		} {
			row := rowOfGlyph(t, tc.m, tc.col)
			if c := cellAt(tc.m, row, tc.col); !styledAs(c, tc.want) {
				t.Fatalf("%s: %q, want style %q", tc.name, c, tc.want.Render(ansi.Strip(c)))
			}
		}
	})
}

func TestThresholdColors_SeriesWithNegativeStyle(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := pulse.New(10, 21, pulse.WithMode(mode), pulse.WithRange(-50, 50), pulse.WithGrid(false), pulse.WithFill(false), pulse.WithLineWidth(1),
			pulse.WithSeriesStyle("a", thBase), pulse.WithSeriesNegativeStyle("a", thNeg),
			pulse.WithThresholds(pulse.Threshold{Value: -20, Style: thWarn}), pulse.WithThresholdColors(true))
		for range 10 {
			m.PushSeries("a", -10)
		}
		row := rowOfGlyph(t, m, 5)
		if c := cellAt(m, row, 5); !styledAs(c, thNeg) {
			t.Fatalf("cell = %q, want the series negative style", c)
		}
	})
}

func TestSetThresholds_ReplacesAndCopies(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := pulse.New(thW, thH, thOpts(mode, pulse.WithThresholds(pulse.Threshold{Value: 20, Style: thWarn, Line: true}))...)
		ts := []pulse.Threshold{{Value: 50, Style: thCrit, Line: true}}
		m.SetThresholds(ts...)
		ts[0].Value = 80
		rows := plotRows(m)
		for r, row := range rows {
			if has := strings.ContainsRune(string(row), '┄'); has != (r == 5) {
				t.Fatalf("row %d = %q: only row 5 may have a line", r, string(row))
			}
		}
		m.SetThresholds()
		if strings.Contains(ansi.Strip(m.View()), "┄") {
			t.Fatal("SetThresholds() must clear the list")
		}
	})
}

func TestSetThresholds_SkipsNonFinite(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		finite := pulse.Threshold{Value: 50, Style: thWarn, Line: true}
		want := pulse.New(thW, thH, thOpts(mode, pulse.WithThresholdColors(true), pulse.WithThresholds(finite))...)
		m := pulse.New(thW, thH, thOpts(mode, pulse.WithThresholdColors(true))...)
		m.SetThresholds(
			pulse.Threshold{Value: math.NaN(), Style: thCrit, Line: true},
			pulse.Threshold{Value: math.Inf(1), Style: thCrit, Line: true},
			pulse.Threshold{Value: math.Inf(-1), Style: thCrit, Line: true},
			finite,
		)
		for _, c := range []*pulse.Model{want, m} {
			for i := range thW {
				c.Push(float64(i) * 5)
			}
		}
		if got := m.View(); got != want.View() {
			t.Fatalf("a non-finite threshold changed the view:\n%s\nwant\n%s", got, want.View())
		}
	})
}

func TestThresholdColors_Accessors(t *testing.T) {
	m := pulse.New(thW, thH)
	if m.ThresholdColors() {
		t.Fatal("threshold colors must be off by default")
	}
	m.SetThresholdColors(true)
	if !m.ThresholdColors() {
		t.Fatal("SetThresholdColors(true) did not stick")
	}
	if !pulse.New(thW, thH, pulse.WithThresholdColors(true)).ThresholdColors() {
		t.Fatal("WithThresholdColors(true) did not stick")
	}
}

// plotPart returns the plot area of one plain-text row of the view.
func plotPart(m *pulse.Model, row int) string {
	line := strings.Split(ansi.Strip(m.View()), "\n")[row]
	return strings.SplitN(line, "│", 2)[1]
}

// viewWithin renders m and fails the test when View does not return in time.
func viewWithin(t *testing.T, m *pulse.Model) string {
	t.Helper()
	done := make(chan string, 1)
	go func() { done <- m.View() }()
	select {
	case v := <-done:
		return v
	case <-time.After(2 * time.Second):
		t.Fatal("View did not return")
		return ""
	}
}

func TestThresholds_LabelGraphemes(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		for _, tc := range []struct {
			name, label string
			cells       int // 0: label is dropped
		}{
			{"ansi", "\x1b[1mBOLD\x1b[m", 4},
			{"nfd accent", "cafe\u0301", 4},
			{"vs16 emoji", "x⚠\ufe0f", 3},
			{"zwj emoji", "\U0001F468\u200d\U0001F469\u200d\U0001F467", 2},
			{"flag", "\U0001F1E9\U0001F1EA", 2},
			{"newline", "x\ny", 0},
			{"tab", "x\ty", 0},
			{"lone zero-width", "\ufe0fx", 0},
		} {
			m := pulse.New(thW, thH, thOpts(mode,
				pulse.WithThresholds(pulse.Threshold{Value: 50, Style: thWarn, Line: true, Label: tc.label}),
			)...)
			viewWithin(t, m)
			got := plotPart(m, 5)
			want := strings.Repeat("┄", thW)
			if tc.cells > 0 {
				want = strings.Repeat("┄", thW-tc.cells) + ansi.Strip(tc.label)
			}
			if got != want {
				t.Fatalf("%s: row = %q, want %q", tc.name, got, want)
			}
			if w := lipgloss.Width(got); w != thW {
				t.Fatalf("%s: plot row is %d cells wide, want %d", tc.name, w, thW)
			}
		}
	})
}

func TestThresholds_SeriesCoversSecondColumnOfWideLabel(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		const w = 10
		m := pulse.New(w, thH, thOpts(mode,
			pulse.WithThresholds(pulse.Threshold{Value: 50, Style: thWarn, Line: true, Label: "日本"}),
		)...)
		for i := range w {
			if i == 7 {
				m.Push(50)
			} else {
				m.Push(math.NaN())
			}
		}
		row := []rune(plotPart(m, 5))
		if len(row) != 9 || string(row[:7]) != "┄┄┄┄┄┄┄" || row[7] == '┄' || row[7] == ' ' || row[8] != '本' {
			t.Fatalf("row = %q, want a dash at column 6, the series at column 7 and the second label rune", string(row))
		}
	})
}

func TestThresholds_SeriesCoversFirstColumnOfWideLabel(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		const w = 10
		m := pulse.New(w, thH, thOpts(mode,
			pulse.WithThresholds(pulse.Threshold{Value: 50, Style: thWarn, Line: true, Label: "日本"}),
		)...)
		for i := range w {
			if i == 6 {
				m.Push(50)
			} else {
				m.Push(math.NaN())
			}
		}
		got := plotPart(m, 5)
		row := []rune(got)
		if len(row) != 9 || string(row[:6]) != "┄┄┄┄┄┄" || row[6] == '┄' || row[6] == ' ' || row[7] != '┄' || row[8] != '本' {
			t.Fatalf("row = %q, want the series at column 6, a dash at column 7 and the second label rune", got)
		}
		if lw := lipgloss.Width(got); lw != w {
			t.Fatalf("plot row is %d cells wide, want %d", lw, w)
		}
	})
}

func TestThresholds_EventGuidelineOverLine(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := pulse.New(thW, thH, thOpts(mode,
			pulse.WithThresholds(pulse.Threshold{Value: 50, Style: thWarn, Line: true}),
		)...)
		for range thW {
			m.Push(10)
		}
		m.AddEventAt(5, pulse.Event{Label: "x", Glyph: "!"})
		row := plotPart(m, 5)
		if strings.Count(row, "┆") != 1 || strings.Count(row, "┄") != thW-1 {
			t.Fatalf("row = %q, want one guideline over the line", row)
		}
	})
}

func TestThresholdColors_SeveralAndInvertedSeries(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := pulse.New(10, 21, pulse.WithMode(mode), pulse.WithRange(-100, 100), pulse.WithGrid(false), pulse.WithFill(false), pulse.WithLineWidth(1),
			pulse.WithSeriesStyle("inv", thBase), pulse.WithSeriesInverted("inv", true),
			pulse.WithSeriesStyle("low", thNeg),
			pulse.WithThresholds(pulse.Threshold{Value: 50, Style: thWarn}), pulse.WithThresholdColors(true))
		for range 10 {
			m.PushSeries("inv", -60)
			m.PushSeries("low", 10)
		}
		var warn, own int
		for r := range 21 {
			c := cellAt(m, r, 5)
			switch {
			case strings.TrimSpace(ansi.Strip(c)) == "":
			case styledAs(c, thWarn):
				warn++
			case styledAs(c, thNeg):
				own++
			default:
				t.Fatalf("row %d: unexpected style %q", r, c)
			}
		}
		if warn != 1 || own != 1 {
			t.Fatalf("warning cells = %d, own-style cells = %d, want 1 and 1", warn, own)
		}
	})
}
