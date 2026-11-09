package pulse_test

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/ingvarch/pulse"
	"github.com/ingvarch/pulse/scale"
)

const raW, raH = 20, 11

func raOpts(mode pulse.RenderMode, opts ...pulse.Option) []pulse.Option {
	return append([]pulse.Option{
		pulse.WithMode(mode),
		pulse.WithRange(0, 100),
		pulse.WithGrid(false),
		pulse.WithFill(false),
		pulse.WithLineWidth(1),
	}, opts...)
}

func raUnits(v float64) string { return fmt.Sprintf("%.0fu", v) }

// raPlotRows returns the plot cells of every View line, with the left label column and the right axis cut off.
func raPlotRows(m *pulse.Model, labelWidth int) []string {
	var rows []string
	for _, line := range strings.Split(ansi.Strip(m.View()), "\n") {
		rows = append(rows, string([]rune(line)[labelWidth+1:labelWidth+1+m.Width()]))
	}
	return rows[:m.Height()]
}

func pushConst(m *pulse.Model, name string, v float64) {
	for range raW {
		m.PushSeries(name, v)
	}
}

func TestRightAxis_SeriesUsesRightRange(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		right := pulse.New(raW, raH, raOpts(mode,
			pulse.WithRightAxis(pulse.Axis{Min: 0, Max: 250}),
			pulse.WithSeriesRightAxis("s"))...)
		onLeft := pulse.New(raW, raH, raOpts(mode)...)
		same := pulse.New(raW, raH, raOpts(mode)...)
		pushConst(right, "s", 50)
		pushConst(onLeft, "s", 50)
		pushConst(same, "s", 20)

		if got, want := raPlotRows(right, 4), raPlotRows(same, 4); strings.Join(got, "\n") != strings.Join(want, "\n") {
			t.Fatalf("right series 50 on 0..250 must draw like 20 on 0..100:\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
		}
		if got, left := raPlotRows(right, 4), raPlotRows(onLeft, 4); strings.Join(got, "\n") == strings.Join(left, "\n") {
			t.Fatal("the same value must land on different rows on the left and right axis")
		}
	})
}

func TestRightAxis_GapsStayGaps(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := pulse.New(raW, raH, raOpts(mode,
			pulse.WithRightAxis(pulse.Axis{Min: 0, Max: 250}),
			pulse.WithSeriesRightAxis("s"))...)
		ref := pulse.New(raW, raH, raOpts(mode)...)
		for i := range raW {
			switch i % 4 {
			case 1:
				m.PushSeries("s", math.NaN())
				ref.PushSeries("s", math.NaN())
			case 2:
				m.PushSeries("s", math.Inf(1))
				ref.PushSeries("s", math.Inf(1))
			case 3:
				m.PushSeries("s", math.Inf(-1))
				ref.PushSeries("s", math.Inf(-1))
			default:
				m.PushSeries("s", 125)
				ref.PushSeries("s", 50)
			}
		}
		got, want := strings.Join(raPlotRows(m, 4), "\n"), strings.Join(raPlotRows(ref, 4), "\n")
		if strings.TrimSpace(want) == "" || got != want {
			t.Fatalf("gaps must stay gaps and points must be drawn:\n%s\nwant\n%s", got, want)
		}
	})
}

func TestRightAxis_LinesHaveEqualWidth(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		for _, xaxis := range []string{"", "at", "every"} {
			opts := raOpts(mode, pulse.WithGrid(true))
			switch xaxis {
			case "every":
				opts = append(opts, pulse.WithXAxis(func(age int) string { return fmt.Sprint(age) }))
			case "at":
				opts = append(opts, pulse.WithXAxisAt(func(age int) string {
					if age%5 == 0 {
						return fmt.Sprint(age)
					}
					return ""
				}))
			}
			plain := pulse.New(raW, raH, opts...)
			m := pulse.New(raW, raH, append(opts, pulse.WithRightAxis(pulse.Axis{
				Min: 0, Max: 5000, Ticks: []float64{0, 2500, 5000}, Formatter: raUnits,
			}))...)
			for _, c := range []*pulse.Model{plain, m} {
				pushConst(c, "s", 50)
			}
			base := lipgloss.Width(strings.SplitN(plain.View(), "\n", 2)[0])
			lines := strings.Split(m.View(), "\n")
			if want := len(strings.Split(plain.View(), "\n")); len(lines) != want {
				t.Fatalf("xaxis=%q: %d lines, want %d", xaxis, len(lines), want)
			}
			for i, line := range lines {
				if got, want := lipgloss.Width(line), base+1+len("5000u"); got != want {
					t.Fatalf("xaxis=%q: line %d width = %d, want %d: %q", xaxis, i, got, want, ansi.Strip(line))
				}
			}
		}
	})
}

func TestRightAxis_LabelsOnTickRowsGridOnLeftTicks(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := pulse.New(raW, raH, raOpts(mode,
			pulse.WithGrid(true),
			pulse.WithTicks(0, 25, 75, 100),
			pulse.WithRightAxis(pulse.Axis{Min: 0, Max: 5000, Ticks: []float64{0, 2500, 5000, 9000}, Formatter: raUnits}))...)
		lines := strings.Split(ansi.Strip(m.View()), "\n")
		wantLabel := map[int]string{0: "5000u", 5: "2500u", 10: "0u"}
		for r, line := range lines {
			right := strings.TrimSpace(string([]rune(line)[4+1+raW+1:]))
			if right != wantLabel[r] {
				t.Fatalf("row %d: right label = %q, want %q", r, right, wantLabel[r])
			}
		}
		if strings.Contains(lines[5], "─") {
			t.Fatalf("row 5 is a right tick only and must have no grid line: %q", lines[5])
		}
		for _, r := range []int{0, 3, 8, 10} {
			if !strings.Contains(lines[r], "─") {
				t.Fatalf("row %d is a left tick and must have a grid line: %q", r, lines[r])
			}
		}
	})
}

func TestRightAxis_DefaultTicksAndFormatter(t *testing.T) {
	m := pulse.New(raW, raH, pulse.WithRange(0, 100), pulse.WithRightAxis(pulse.Axis{Min: 0, Max: 6000}))
	lines := strings.Split(ansi.Strip(m.View()), "\n")
	if got, want := strings.TrimSpace(string([]rune(lines[0])[4+1+raW+1:])), scale.FormatTick(6000); got != want {
		t.Fatalf("top right label = %q, want %q", got, want)
	}
}

func TestRightAxis_AxisLineMirrorsZeroBaseline(t *testing.T) {
	m := pulse.New(raW, raH, pulse.WithRange(-100, 100), pulse.WithZeroBaseline(true),
		pulse.WithRightAxis(pulse.Axis{Min: 0, Max: 10}))
	for r, line := range strings.Split(ansi.Strip(m.View()), "\n") {
		want := '│'
		if r == 5 {
			want = '┤'
		}
		if got := []rune(line)[4+1+raW]; got != want {
			t.Fatalf("row %d: right axis glyph = %q, want %q", r, got, want)
		}
	}
}

func TestRightAxis_XAxisRowClosesWithCorner(t *testing.T) {
	m := pulse.New(raW, raH, pulse.WithXAxis(func(age int) string { return fmt.Sprint(age) }),
		pulse.WithRightAxis(pulse.Axis{Min: 0, Max: 5000, Ticks: []float64{5000}, Formatter: raUnits}))
	lines := strings.Split(ansi.Strip(m.View()), "\n")
	axis := []rune(lines[raH])
	if got := axis[4+1+raW]; got != '┘' {
		t.Fatalf("X axis row corner = %q, want ┘: %q", got, lines[raH])
	}
	if rest := strings.TrimSpace(string(axis[4+1+raW+1:])); rest != "" {
		t.Fatalf("X axis row must be blank after the corner, got %q", rest)
	}
}

func TestRightAxis_HigherTickWinsOnSharedRow(t *testing.T) {
	m := pulse.New(raW, raH, pulse.WithRange(0, 100),
		pulse.WithRightAxis(pulse.Axis{Min: 0, Max: 5000, Ticks: []float64{100, 5000, 0}, Formatter: raUnits}))
	lines := strings.Split(ansi.Strip(m.View()), "\n")
	if got := strings.TrimSpace(string([]rune(lines[raH-1])[4+1+raW+1:])); got != "100u" {
		t.Fatalf("bottom right label = %q, want the higher tick 100u", got)
	}
}

func TestRightAxis_LegendMarksRightSeries(t *testing.T) {
	m := pulse.New(raW, raH, pulse.WithRightAxis(pulse.Axis{Min: 0, Max: 5000}))
	m.PushSeries("cpu", 10)
	m.PushSeries("rps", 20)
	m.SetSeriesRightAxis("rps", true)
	if got := ansi.Strip(m.Legend()); got != "■ cpu  ■ rps →" {
		t.Fatalf("Legend() = %q", got)
	}
	box := ansi.Strip(m.LegendBox())
	if !strings.Contains(box, "rps → ") || strings.Contains(box, "cpu →") {
		t.Fatalf("LegendBox() must mark rps only:\n%s", box)
	}
}

func TestRightAxis_AutoRangeIgnoresRightSeries(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := pulse.New(raW, raH, pulse.WithMode(mode), pulse.WithAutoRange(true),
			pulse.WithRightAxis(pulse.Axis{Min: 0, Max: 5000}),
			pulse.WithSeriesRightAxis("rps"))
		m.PushSeries("cpu", 10)
		m.PushSeries("rps", 4000)
		m.View()
		if m.Max() > 20 {
			t.Fatalf("Max = %v, want a range that fits the left series only", m.Max())
		}
	})
}

func TestRightAxis_ThresholdColorsSkipRightSeries(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		build := func(right bool) *pulse.Model {
			opts := raOpts(mode,
				pulse.WithSeriesStyle("s", thBase),
				pulse.WithThresholds(pulse.Threshold{Value: 10, Style: thCrit}),
				pulse.WithThresholdColors(true),
				pulse.WithRightAxis(pulse.Axis{Min: 0, Max: 100}))
			if right {
				opts = append(opts, pulse.WithSeriesRightAxis("s"))
			}
			m := pulse.New(raW, raH, opts...)
			pushConst(m, "s", 50)
			return m
		}
		row := rowOfGlyph(t, build(true), 10)
		if c := cellAt(build(false), row, 10); !styledAs(c, thCrit) {
			t.Fatalf("control: left series cell = %q, want the threshold style", c)
		}
		if c := cellAt(build(true), row, 10); !styledAs(c, thBase) {
			t.Fatalf("right series cell = %q, want the series style", c)
		}
	})
}

func TestRightAxis_InvertedSeriesIsNegatedBeforeMapping(t *testing.T) {
	m := pulse.New(raW, raH, pulse.WithRange(-100, 100), pulse.WithGrid(false), pulse.WithFill(false),
		pulse.WithRightAxis(pulse.Axis{Min: -250, Max: 250}),
		pulse.WithSeriesRightAxis("s"), pulse.WithSeriesInverted("s", true))
	ref := pulse.New(raW, raH, pulse.WithRange(-100, 100), pulse.WithGrid(false), pulse.WithFill(false))
	pushConst(m, "s", 125)
	pushConst(ref, "s", -50)
	if got, want := raPlotRows(m, 4), raPlotRows(ref, 4); strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("inverted right series 125 must draw like -50:\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestRightAxis_LogScaleKeepsRightAxisLinear(t *testing.T) {
	m := pulse.New(raW, raH, pulse.WithRange(1, 100), pulse.WithLogScale(true), pulse.WithGrid(false), pulse.WithFill(false),
		pulse.WithRightAxis(pulse.Axis{Min: 0, Max: 250}), pulse.WithSeriesRightAxis("s"))
	ref := pulse.New(raW, raH, pulse.WithRange(1, 100), pulse.WithLogScale(true), pulse.WithGrid(false), pulse.WithFill(false))
	pushConst(m, "s", 125)
	pushConst(ref, "s", 10) // the middle of a 1..100 log range
	if got, want := raPlotRows(m, 4), raPlotRows(ref, 4); strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("right series at the middle of 0..250 must sit at the middle row:\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestSetRightAxis_Validation(t *testing.T) {
	m := pulse.New(raW, raH)
	m.SetRightAxis(pulse.Axis{Min: math.NaN(), Max: 10})
	if _, ok := m.RightAxis(); ok {
		t.Fatal("a NaN bound must not create the axis")
	}
	m.SetRightAxis(pulse.Axis{Min: 0, Max: math.Inf(1)})
	if _, ok := m.RightAxis(); ok {
		t.Fatal("an infinite bound must not create the axis")
	}
	m.SetRightAxis(pulse.Axis{Min: 20, Max: 10})
	if a, ok := m.RightAxis(); !ok || a.Min != 10 || a.Max != 20 {
		t.Fatalf("RightAxis() = %+v, %v, want swapped 10..20", a, ok)
	}
	m.SetRightAxis(pulse.Axis{Min: math.NaN(), Max: 99})
	if a, _ := m.RightAxis(); a.Min != 10 || a.Max != 20 {
		t.Fatalf("a NaN bound must keep the range, got %+v", a)
	}
	m.SetRightAxis(pulse.Axis{Min: 5, Max: 5})
	if a, _ := m.RightAxis(); a.Min != 5 || a.Max != 6 {
		t.Fatalf("equal bounds must widen to 5..6, got %+v", a)
	}
}

func TestSetSeriesRightAxis_Toggle(t *testing.T) {
	m := pulse.New(raW, raH)
	if m.SeriesRightAxis("s") {
		t.Fatal("a new series is on the left axis")
	}
	m.SetSeriesRightAxis("s", true)
	if !m.SeriesRightAxis("s") {
		t.Fatal("SeriesRightAxis must report true")
	}
	m.SetSeriesRightAxis("s", false)
	if m.SeriesRightAxis("s") {
		t.Fatal("SeriesRightAxis must report false")
	}
}

func TestRightAxis_SeriesWithoutAxisStaysOnLeft(t *testing.T) {
	m := pulse.New(raW, raH, raOpts(pulse.ModeLines, pulse.WithSeriesRightAxis("s"))...)
	ref := pulse.New(raW, raH, raOpts(pulse.ModeLines)...)
	pushConst(m, "s", 50)
	pushConst(ref, "s", 50)
	if m.View() != ref.View() {
		t.Fatal("without a right axis the flag must change nothing")
	}
	if strings.Contains(m.Legend(), "→") {
		t.Fatal("without a right axis the legend must not mark the series")
	}
}

func TestRightAxis_NegativeStyleSkipsRightSeries(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		build := func(right bool, v float64) *pulse.Model {
			opts := []pulse.Option{
				pulse.WithMode(mode), pulse.WithRange(-100, 100), pulse.WithZeroBaseline(true),
				pulse.WithGrid(false), pulse.WithFill(false), pulse.WithLineWidth(1),
				pulse.WithSeriesStyle("s", thBase), pulse.WithSeriesNegativeStyle("s", thNeg),
				pulse.WithRightAxis(pulse.Axis{Min: 0, Max: 100}),
			}
			if right {
				opts = append(opts, pulse.WithSeriesRightAxis("s"))
			}
			m := pulse.New(raW, raH, opts...)
			pushConst(m, "s", v)
			return m
		}
		left := build(false, -80)
		if c := cellAt(left, 9, 10); !styledAs(c, thNeg) {
			t.Fatalf("control: left cell = %q, want the negative style", c)
		}
		right := build(true, 10) // maps to -80 on the left range
		if c := cellAt(right, 9, 10); !styledAs(c, thBase) {
			t.Fatalf("right cell = %q, want the series style", c)
		}
	})
}

type raRecorder struct {
	data     []float64
	inverted bool
	right    []bool
	min, max float64
}

func (r *raRecorder) Render(c pulse.Chart, ctx pulse.RenderContext) string {
	r.data, r.inverted, r.right = c.SeriesData("r"), c.SeriesInverted("r"), ctx.RightAxis
	r.min, r.max = c.Min(), c.Max()
	return ""
}

func TestRightAxis_CustomRendererGetsMappedValues(t *testing.T) {
	rec := &raRecorder{}
	m := pulse.New(4, raH, pulse.WithRenderer(rec), pulse.WithRange(0, 100),
		pulse.WithRightAxis(pulse.Axis{Min: 0, Max: 250}),
		pulse.WithSeriesRightAxis("r"), pulse.WithSeriesInverted("r", true))
	m.PushSeries("r", 125)
	m.PushSeries("r", math.NaN())
	_ = m.View()
	if len(rec.data) != 2 || rec.data[0] != -50 {
		t.Fatalf("Chart.SeriesData = %v, want the negated 125 mapped to -50", rec.data)
	}
	if !math.IsNaN(rec.data[1]) {
		t.Fatalf("a gap must stay a gap, got %v", rec.data[1])
	}
	if rec.inverted {
		t.Fatal("SeriesInverted must be false for a right-axis series")
	}
	if len(rec.right) != 2 || rec.right[0] || !rec.right[1] {
		t.Fatalf("ctx.RightAxis = %v, want [false true]", rec.right)
	}
}

func TestRightAxis_InvertedSeriesOnLogScale(t *testing.T) {
	m := pulse.New(raW, raH, pulse.WithRange(1, 100), pulse.WithLogScale(true), pulse.WithGrid(false), pulse.WithFill(false),
		pulse.WithRightAxis(pulse.Axis{Min: -250, Max: 250}),
		pulse.WithSeriesRightAxis("s"), pulse.WithSeriesInverted("s", true))
	ref := pulse.New(raW, raH, pulse.WithRange(1, 100), pulse.WithLogScale(true), pulse.WithGrid(false), pulse.WithFill(false),
		pulse.WithRightAxis(pulse.Axis{Min: -250, Max: 250}), pulse.WithSeriesRightAxis("s"))
	pushConst(m, "s", -125)
	pushConst(ref, "s", 125)
	if got, want := raPlotRows(m, 4), raPlotRows(ref, 4); strings.Join(got, "\n") != strings.Join(want, "\n") || strings.TrimSpace(strings.Join(want, "")) == "" {
		t.Fatalf("inverted right series -125 must draw like 125:\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}
