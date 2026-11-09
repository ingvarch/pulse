package pulse_test

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/ingvarch/pulse"
)

const cursorW = 20

// newCursorChart returns a chart with n points 0..n-1 in the default series; the line sits at the bottom row.
func newCursorChart(n int, opts ...pulse.Option) *pulse.Model {
	m := pulse.New(cursorW, 5, append([]pulse.Option{
		pulse.WithRange(0, 1000), pulse.WithGrid(false), pulse.WithFill(false), pulse.WithHistory(100),
	}, opts...)...)
	for i := range n {
		m.Push(float64(i))
	}
	return m
}

// cursorCol returns the plot column of the cursor glyph in the top row, or -1.
func cursorCol(m *pulse.Model) int { return cursorColAt(m, 0) }

func cursorColAt(m *pulse.Model, line int) int {
	row := []rune(ansi.Strip(strings.Split(m.View(), "\n")[line]))
	axis := slices.Index(row, '│')
	glyph := slices.Index(row, '┆')
	if glyph < 0 {
		return -1
	}
	return glyph - axis - 1
}

func TestCursor_ColumnForAges(t *testing.T) {
	for _, mode := range []pulse.RenderMode{pulse.ModeLines, pulse.ModeBraille} {
		m := newCursorChart(50)
		m.SetMode(mode)
		for age, want := range map[int]int{0: 19, 3: 16, 19: 0, 20: -1, 70: -1} {
			m.SetCursor(age)
			if got := cursorCol(m); got != want {
				t.Errorf("mode %v age %d: column %d, want %d", mode, age, got, want)
			}
		}
	}
}

func TestCursor_ColumnMatchesEvents(t *testing.T) {
	for _, mode := range []pulse.RenderMode{pulse.ModeLines, pulse.ModeBraille} {
		m := newCursorChart(50)
		m.SetMode(mode)
		m.AddEventAt(6, pulse.Event{NoLine: true})
		m.SetCursor(6)
		ev := m.VisibleEvents()
		if len(ev) != 1 || cursorColAt(m, 1) != ev[0].Col {
			t.Fatalf("mode %v: events %v, cursor column %d", mode, ev, cursorColAt(m, 1))
		}
	}
}

func TestCursor_HiddenByDefaultAndBelowMinusOne(t *testing.T) {
	m := newCursorChart(50)
	if m.Cursor() != -1 || m.Offset() != 0 {
		t.Fatalf("Cursor %d Offset %d, want -1 and 0", m.Cursor(), m.Offset())
	}
	if strings.Contains(m.View(), "┆") {
		t.Fatal("hidden cursor drew a glyph")
	}
	m.SetCursor(3)
	m.SetCursor(-5)
	if m.Cursor() != -1 || strings.Contains(m.View(), "┆") {
		t.Fatalf("Cursor %d after SetCursor(-5)", m.Cursor())
	}
}

func TestCursor_DrawnOnEveryPlotRowOverThresholdAndGrid(t *testing.T) {
	m := newCursorChart(50, pulse.WithGrid(true))
	m.SetThresholds(pulse.Threshold{Value: 500, Line: true})
	m.SetCursor(4)
	rows := strings.Split(ansi.Strip(m.View()), "\n")
	count := 0
	for _, r := range rows[:4] {
		if strings.ContainsRune(r, '┆') {
			count++
		}
	}
	if count != 4 {
		t.Fatalf("cursor on %d of the 4 rows above the line:\n%s", count, strings.Join(rows, "\n"))
	}
}

func TestCursorValues(t *testing.T) {
	m := pulse.New(cursorW, 5, pulse.WithHistory(100))
	if m.CursorValues() != nil {
		t.Fatal("CursorValues must be nil while hidden")
	}
	for i := range 10 {
		m.PushSeries("a", float64(i))
		m.PushSeries("b", float64(i)*10)
	}
	m.PushSeries("a", math.NaN())
	m.PushSeries("b", 5)
	m.PushSeries("short", 1)
	m.SetSeriesInverted("b", true)

	m.SetCursor(1)
	got := m.CursorValues()
	if len(got) != 3 || got["a"] != 9 || got["b"] != 90 || !math.IsNaN(got["short"]) {
		t.Fatalf("age 1: %v, want a 9, b 90, short NaN", got)
	}
	m.SetCursor(0)
	got = m.CursorValues()
	if !math.IsNaN(got["a"]) || got["b"] != 5 || got["short"] != 1 {
		t.Fatalf("age 0: %v, want a NaN, b 5, short 1", got)
	}
	if _, ok := got[""]; ok {
		t.Fatal("default series without data must not be listed")
	}
	m.SetCursor(500)
	for name, v := range m.CursorValues() {
		if !math.IsNaN(v) {
			t.Fatalf("%s beyond the history = %v, want NaN", name, v)
		}
	}
}

func TestCursorValues_OffScreenAndDefaultSeries(t *testing.T) {
	m := newCursorChart(50)
	m.SetCursor(30)
	if cursorCol(m) != -1 {
		t.Fatal("cursor outside the window drew a glyph")
	}
	if got := m.CursorValues()[""]; got != 19 {
		t.Fatalf("value at age 30 = %v, want 19", got)
	}
}

func TestCursorValues_LogScaleReturnsPushedValue(t *testing.T) {
	m := newCursorChart(10, pulse.WithLogScale(true))
	m.SetCursor(2)
	if got := m.CursorValues()[""]; got != 7 {
		t.Fatalf("value = %v, want 7", got)
	}
}

func TestCursorBox(t *testing.T) {
	m := pulse.New(cursorW, 5, pulse.WithHistory(100))
	if m.CursorBox() != "" {
		t.Fatal("CursorBox must be empty while hidden")
	}
	for i := range 30 {
		m.PushSeries("cpu", float64(i))
		m.PushSeries("mem", float64(i)*2)
	}
	m.SetCursor(4)
	box := ansi.Strip(m.CursorBox())
	for _, want := range []string{"cpu 25", "mem 50", "╭", "■"} {
		if !strings.Contains(box, want) {
			t.Fatalf("box lacks %q:\n%s", want, box)
		}
	}
	if strings.Contains(box, "age") || len(strings.Split(box, "\n")) != 4 {
		t.Fatalf("box without an X axis must hold only the series lines:\n%s", box)
	}

	m.SetXAxis(func(age int) string { return fmt.Sprintf("age %d", age) })
	lines := strings.Split(ansi.Strip(m.CursorBox()), "\n")
	if len(lines) != 5 || !strings.Contains(lines[1], "age 4") {
		t.Fatalf("box with an X axis:\n%s", strings.Join(lines, "\n"))
	}
}

func TestCursorBox_DefaultSeriesHasNoNameGap(t *testing.T) {
	m := newCursorChart(30)
	m.SetCursor(2)
	box := ansi.Strip(m.CursorBox())
	if !strings.Contains(box, "■ 27") {
		t.Fatalf("an unnamed series must show the swatch and the value with one space:\n%s", box)
	}
	m.SetLabelFormatter(func(v float64) string { return fmt.Sprintf("%4.0f", v) })
	if box := ansi.Strip(m.CursorBox()); !strings.Contains(box, "■   27") {
		t.Fatalf("the formatter's padding must stay:\n%s", box)
	}
}

func TestCursorBox_NoLabelLineWhenTheLabelIsEmpty(t *testing.T) {
	m := newCursorChart(30)
	m.SetXAxisAt(func(age int) string {
		if age%10 == 0 {
			return "x"
		}
		return ""
	})
	m.SetCursor(3)
	if box := ansi.Strip(m.CursorBox()); len(strings.Split(box, "\n")) != 3 {
		t.Fatalf("an empty label must not add a line:\n%s", box)
	}
}

func TestCursorValues_DefaultSeriesWithoutData(t *testing.T) {
	m := pulse.New(cursorW, 5)
	m.SetSeriesStyle("", lipgloss.NewStyle())
	m.PushSeries("a", 1)
	m.SetCursor(0)
	if got := m.CursorValues(); len(got) != 1 || got["a"] != 1 {
		t.Fatalf("CursorValues = %v, want only a 1", got)
	}
	if box := ansi.Strip(m.CursorBox()); strings.Count(box, "■") != 1 {
		t.Fatalf("the default series without data must not be listed:\n%s", box)
	}
}

func TestCursorBox_RightAxisMarkAndLegendFormat(t *testing.T) {
	m := pulse.New(cursorW, 5,
		pulse.WithRightAxis(pulse.Axis{Min: 0, Max: 10}),
		pulse.WithSeriesRightAxis("temp"),
		pulse.WithLabelFormatter(func(v float64) string { return fmt.Sprintf("%.1f%%", v) }),
	)
	m.PushSeries("temp", 3)
	m.SetCursor(0)
	if box := ansi.Strip(m.CursorBox()); !strings.Contains(box, "temp → 3.0%") {
		t.Fatalf("box:\n%s", box)
	}
}

func TestOffset_ShowsOlderPointsAndClamps(t *testing.T) {
	m := newCursorChart(50)
	m.SetOffset(5)
	data := m.SeriesData("")
	if m.Offset() != 5 || len(data) != cursorW || data[cursorW-1] != 44 || data[0] != 25 {
		t.Fatalf("offset %d data %v", m.Offset(), data)
	}
	if v, _ := m.Last(""); v != 49 {
		t.Fatalf("Last = %v, want the newest value 49", v)
	}
	m.SetOffset(1000)
	if m.Offset() != 30 {
		t.Fatalf("Offset = %d, want 30 (50 points - 20 columns)", m.Offset())
	}
	m.SetOffset(-3)
	if m.Offset() != 0 {
		t.Fatalf("Offset = %d, want 0", m.Offset())
	}
	short := newCursorChart(10)
	short.SetOffset(4)
	if short.Offset() != 0 {
		t.Fatalf("a history shorter than the width cannot scroll, Offset = %d", short.Offset())
	}
}

func TestOffset_ShorterSeriesShowsNothingPastItsHistory(t *testing.T) {
	m := newCursorChart(50)
	for i := range 3 {
		m.PushSeries("new", float64(i))
	}
	m.SetOffset(10)
	if m.LenSeries("new") != 0 {
		t.Fatalf("LenSeries = %d, want 0", m.LenSeries("new"))
	}
}

func TestOffset_PushesKeepTheSamePointsInView(t *testing.T) {
	m := newCursorChart(50)
	m.SetOffset(5)
	before := m.SeriesData("")
	for i := range 4 {
		m.Push(float64(100 + i))
	}
	if m.Offset() != 9 || !slices.Equal(m.SeriesData(""), before) {
		t.Fatalf("offset %d data %v, want 9 and %v", m.Offset(), m.SeriesData(""), before)
	}
}

func TestOffset_StopsAtEndOfHistory(t *testing.T) {
	m := newCursorChart(25, pulse.WithHistory(25))
	m.SetOffset(5)
	m.SetCursor(8)
	m.Push(25)
	if m.Offset() != 5 || m.Cursor() != 8 {
		t.Fatalf("Offset %d Cursor %d, want 5 and 8 at the history limit", m.Offset(), m.Cursor())
	}
	m.SetOffset(2)
	m.SetHistory(0)
	if m.Offset() != 0 {
		t.Fatalf("Offset = %d after the history shrank to the width", m.Offset())
	}
}

func TestOffset_PushDoesNotScrollWhenLive(t *testing.T) {
	m := newCursorChart(50)
	m.Push(1)
	if m.Offset() != 0 {
		t.Fatalf("Offset = %d, want 0", m.Offset())
	}
}

func TestOffset_SetSizeReclamps(t *testing.T) {
	m := newCursorChart(50)
	m.SetOffset(30)
	m.SetSize(40, 5)
	if m.Offset() != 10 {
		t.Fatalf("Offset = %d, want 10", m.Offset())
	}
}

func TestCursor_AgeFollowsPushesWhileScrolled(t *testing.T) {
	m := newCursorChart(50)
	m.SetCursor(3)
	m.Push(1)
	if m.Cursor() != 3 {
		t.Fatalf("live cursor age = %d, want 3", m.Cursor())
	}
	m.SetOffset(5)
	m.SetCursor(7)
	want := m.CursorValues()[""]
	m.Push(1)
	m.Push(1)
	if m.Cursor() != 9 || m.CursorValues()[""] != want {
		t.Fatalf("scrolled cursor age %d value %v, want 9 and %v", m.Cursor(), m.CursorValues()[""], want)
	}
	hidden := newCursorChart(50)
	hidden.SetOffset(5)
	hidden.Push(1)
	if hidden.Cursor() != -1 {
		t.Fatalf("hidden cursor moved to %d", hidden.Cursor())
	}
}

func TestCursor_ColumnAtOffset(t *testing.T) {
	m := newCursorChart(50)
	m.SetOffset(5)
	for age, want := range map[int]int{0: -1, 4: -1, 5: 19, 8: 16, 24: 0, 25: -1} {
		m.SetCursor(age)
		if got := cursorCol(m); got != want {
			t.Errorf("age %d: column %d, want %d", age, got, want)
		}
	}
}

func TestOffset_XAxisLabelsGetRealAges(t *testing.T) {
	m := newCursorChart(50, pulse.WithXAxis(func(age int) string { return fmt.Sprint(age) }))
	m.SetOffset(5)
	lines := strings.Split(ansi.Strip(m.View()), "\n")
	labels := lines[len(lines)-1]
	if !strings.HasSuffix(strings.TrimRight(labels, " "), "5") || !strings.Contains(labels, "24") {
		t.Fatalf("labels %q, want the right edge 5 and the left edge 24", labels)
	}
}

func TestOffset_EventsFollowTheOffset(t *testing.T) {
	m := newCursorChart(50)
	m.AddEventAt(7, pulse.Event{ID: "e"})
	m.SetOffset(5)
	ev := m.VisibleEvents()
	if len(ev) != 1 || ev[0].Col != 17 || ev[0].Age != 7 {
		t.Fatalf("events %v, want col 17 age 7", ev)
	}
	m.SetOffset(8)
	if len(m.VisibleEvents()) != 0 {
		t.Fatalf("an event newer than the offset must not draw: %v", m.VisibleEvents())
	}
	m.SetOffset(30)
	if len(m.VisibleEvents()) != 0 {
		t.Fatalf("an event newer than the offset must not draw: %v", m.VisibleEvents())
	}
}

func TestOffset_EventOlderThanTheWindow(t *testing.T) {
	m := newCursorChart(50)
	m.AddEventAt(30, pulse.Event{ID: "old"})
	m.SetOffset(5)
	if len(m.VisibleEvents()) != 0 {
		t.Fatalf("an event older than the window must not draw: %v", m.VisibleEvents())
	}
	m.SetOffset(15)
	ev := m.VisibleEvents()
	if len(ev) != 1 || ev[0].Col != 4 || ev[0].Age != 30 {
		t.Fatalf("events %v, want col 4 age 30", ev)
	}
}

func TestCursor_BrailleColumnMatchesThePoint(t *testing.T) {
	m := pulse.New(cursorW, 5, pulse.WithMode(pulse.ModeBraille), pulse.WithRange(0, 100),
		pulse.WithGrid(false), pulse.WithFill(false), pulse.WithHistory(100))
	for i := range 40 {
		v := 0.0
		if i == 40-1-11 {
			v = 100
		}
		m.Push(v)
	}
	m.SetOffset(5)
	m.SetCursor(11)
	want := cursorColAt(m, 4)
	top := []rune(ansi.Strip(strings.Split(m.View(), "\n")[0]))
	axis := slices.Index(top, '│')
	first, last := -1, -1
	for i, r := range top[axis+1:] {
		if r > 0x2800 && r <= 0x28FF {
			if first < 0 {
				first = i
			}
			last = i
		}
	}
	if want < 0 || first < 0 || (first+last)/2 != want {
		t.Fatalf("spike in columns %d-%d, cursor in column %d\n%s", first, last, want, ansi.Strip(m.View()))
	}
}

func TestOffset_AutoRangeFollows(t *testing.T) {
	m := pulse.New(cursorW, 5, pulse.WithAutoRange(true), pulse.WithHistory(100))
	for i := range 40 {
		v := 1.0
		if i < 10 {
			v = 900
		}
		m.Push(v)
	}
	m.View()
	live := m.Max()
	m.SetOffset(20)
	m.View()
	if m.Max() <= live {
		t.Fatalf("Max at offset %d = %v, live %v: auto range must see the older, larger points", m.Offset(), m.Max(), live)
	}
}

func TestCursor_DrawnOverTheFill(t *testing.T) {
	for _, tinted := range []bool{false, true} {
		m := newCursorChart(0, pulse.WithFill(true), pulse.WithTintedFill(tinted))
		for range 30 {
			m.Push(1000)
		}
		m.SetCursor(4)
		want := "┆"
		if tinted {
			want = lipgloss.NewStyle().Background(pulse.DefaultTintColor).Render("┆")
		}
		for r := 1; r < 5; r++ {
			if c := cellAt(m, r, cursorW-1-4); c != want {
				t.Errorf("tinted %v row %d: %q, want %q", tinted, r, c, want)
			}
		}
	}
}

func TestCursor_DrawOrder(t *testing.T) {
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	const col = cursorW - 1 - 4
	for _, mode := range []pulse.RenderMode{pulse.ModeLines, pulse.ModeBraille} {
		m := newCursorChart(50, pulse.WithMode(mode))
		m.SetCursor(4)
		if c := cellAt(m, 4, col); ansi.Strip(c) == "┆" {
			t.Errorf("mode %v: the line must cover the cursor, got %q", mode, c)
		}
		m.AddEventAt(4, pulse.Event{Style: red})
		if c := cellAt(m, 2, col); !styledAs(c, red) {
			t.Errorf("mode %v: an event guideline must cover the cursor, got %q", mode, c)
		}
	}
	for _, tc := range []struct {
		mode pulse.RenderMode
		fill bool
	}{{pulse.ModeLines, false}, {pulse.ModeLines, true}, {pulse.ModeBraille, false}} {
		m := pulse.New(cursorW, 5, pulse.WithMode(tc.mode), pulse.WithRange(-1000, 1000), pulse.WithGrid(false),
			pulse.WithZeroBaseline(true), pulse.WithFill(tc.fill))
		for range 30 {
			m.Push(1000)
		}
		m.SetCursor(4)
		if c := cellAt(m, 2, col); c != "┆" {
			t.Errorf("mode %v fill %v: the cursor must cover the zero baseline, got %q", tc.mode, tc.fill, c)
		}
	}
}

func TestOffset_OneStepPerTickWithDefaultAndNamedSeries(t *testing.T) {
	m := pulse.New(cursorW, 5, pulse.WithHistory(100))
	for i := range 50 {
		m.Push(float64(i))
		m.PushSeries("a", float64(i))
	}
	m.SetOffset(5)
	before := m.SeriesData("a")
	m.Push(1)
	m.PushSeries("a", 1)
	if m.Offset() != 6 || !slices.Equal(m.SeriesData("a"), before) {
		t.Fatalf("offset %d data %v, want 6 and %v", m.Offset(), m.SeriesData("a"), before)
	}
}
