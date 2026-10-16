package pulse_test

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/ingvarch/pulse"
)

const gapW, gapH = 40, 5

// plotRows returns the plot area of the chart as plain-text rows, without labels and axis.
func plotRows(c *pulse.Model) [][]rune {
	lines := strings.Split(ansi.Strip(c.View()), "\n")
	rows := make([][]rune, len(lines))
	for i, l := range lines {
		r := []rune(l)
		rows[i] = r[len(r)-c.Width():]
	}
	return rows
}

func colRunes(rows [][]rune, col int) string {
	var sb strings.Builder
	for _, r := range rows {
		sb.WriteRune(r[col])
	}
	return sb.String()
}

func isBlankCol(rows [][]rune, col int) bool {
	return strings.TrimSpace(colRunes(rows, col)) == ""
}

func assertBlankCols(t *testing.T, rows [][]rune, from, to int) {
	t.Helper()
	for col := from; col < to; col++ {
		if !isBlankCol(rows, col) {
			t.Fatalf("column %d must be blank, got %q", col, colRunes(rows, col))
		}
	}
}

func assertDrawnCols(t *testing.T, rows [][]rune, from, to int) {
	t.Helper()
	for col := from; col < to; col++ {
		if isBlankCol(rows, col) {
			t.Fatalf("column %d must have a line", col)
		}
	}
}

func gapValue(i, from, to int) float64 {
	if i >= from && i < to {
		return math.NaN()
	}
	return 40 + 30*math.Sin(float64(i)*0.3)
}

func newGapChart(mode pulse.RenderMode, fill bool, opts ...pulse.Option) *pulse.Model {
	base := []pulse.Option{
		pulse.WithRange(0, 100), pulse.WithTicks(0, 50, 100),
		pulse.WithGrid(false), pulse.WithFill(fill), pulse.WithLineWidth(1),
		pulse.WithMode(mode),
	}
	return pulse.New(gapW, gapH, append(base, opts...)...)
}

var gapModes = []struct {
	name string
	mode pulse.RenderMode
}{
	{"lines", pulse.ModeLines},
	{"braille", pulse.ModeBraille},
}

func TestGap_MiddleStartEnd(t *testing.T) {
	cases := []struct {
		name     string
		from, to int
	}{
		{"middle", 15, 22},
		{"start", 0, 8},
		{"end", 30, 40},
	}
	for _, m := range gapModes {
		for _, fill := range []bool{false, true} {
			for _, tc := range cases {
				name := m.name + "/" + tc.name
				if fill {
					name += "/fill"
				}
				t.Run(name, func(t *testing.T) {
					c := newGapChart(m.mode, fill)
					for i := 0; i < gapW; i++ {
						c.Push(gapValue(i, tc.from, tc.to))
					}
					rows := plotRows(c)
					assertBlankCols(t, rows, tc.from, tc.to)
					assertDrawnCols(t, rows, 0, tc.from)
					assertDrawnCols(t, rows, tc.to, gapW)
				})
			}
		}
	}
}

func TestGap_LonePoint(t *testing.T) {
	t.Run("lines", func(t *testing.T) {
		c := newGapChart(pulse.ModeLines, false)
		for i := 0; i < 10; i++ {
			switch i {
			case 5:
				c.Push(50)
			default:
				c.Push(math.NaN())
			}
		}
		rows := plotRows(c)
		// 10 points occupy the last 10 columns, the lone point is at column gapW-5.
		col := gapW - 10 + 5
		if got := strings.TrimSpace(colRunes(rows, col)); got != "─" {
			t.Fatalf("lone point must be a single ─, got %q", got)
		}
		for i := 0; i < gapW; i++ {
			if i != col {
				assertBlankCols(t, rows, i, i+1)
			}
		}
	})
	t.Run("braille", func(t *testing.T) {
		c := newGapChart(pulse.ModeBraille, false)
		for i := 0; i < 10; i++ {
			switch i {
			case 5:
				c.Push(50)
			default:
				c.Push(math.NaN())
			}
		}
		rows := plotRows(c)
		nonEmpty := 0
		for _, row := range rows {
			for _, ch := range row {
				if ch != ' ' {
					nonEmpty++
				}
			}
		}
		if nonEmpty != 1 {
			t.Fatalf("lone point must be one braille cell, got %d", nonEmpty)
		}
	})
}

func TestGap_AllGapsDrawNothing(t *testing.T) {
	for _, m := range gapModes {
		for _, fill := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/fill=%v", m.name, fill), func(t *testing.T) {
				c := newGapChart(m.mode, fill)
				for i := 0; i < gapW; i++ {
					c.Push(math.NaN())
				}
				assertBlankCols(t, plotRows(c), 0, gapW)
			})
		}
	}
}

func TestGap_InfIsGap(t *testing.T) {
	for _, m := range gapModes {
		for _, inf := range []float64{math.Inf(1), math.Inf(-1)} {
			t.Run(fmt.Sprintf("%s/%v", m.name, inf), func(t *testing.T) {
				c := newGapChart(m.mode, true)
				for i := 0; i < gapW; i++ {
					v := gapValue(i, 15, 22)
					if math.IsNaN(v) {
						v = inf
					}
					c.Push(v)
				}
				rows := plotRows(c)
				assertBlankCols(t, rows, 15, 22)
				assertDrawnCols(t, rows, 0, 15)
				assertDrawnCols(t, rows, 22, gapW)
			})
		}
	}
}

func TestGap_OtherSeriesStillDrawn(t *testing.T) {
	for _, m := range gapModes {
		t.Run(m.name, func(t *testing.T) {
			c := newGapChart(m.mode, false)
			for i := 0; i < gapW; i++ {
				c.PushSeries("a", gapValue(i, 15, 22))
				c.PushSeries("b", 80)
			}
			rows := plotRows(c)
			assertDrawnCols(t, rows, 15, 22)
			solo := newGapChart(m.mode, false)
			for i := 0; i < gapW; i++ {
				solo.PushSeries("b", 80)
			}
			want := plotRows(solo)
			for col := 15; col < 22; col++ {
				if colRunes(rows, col) != colRunes(want, col) {
					t.Fatalf("column %d: got %q, want only series b %q", col, colRunes(rows, col), colRunes(want, col))
				}
			}
		})
	}
}

func TestGap_GridAndEventStillShow(t *testing.T) {
	for _, m := range gapModes {
		t.Run(m.name, func(t *testing.T) {
			c := newGapChart(m.mode, false, pulse.WithGrid(true))
			for i := 0; i < gapW; i++ {
				c.Push(gapValue(i, 15, 22))
			}
			c.AddEventAt(gapW-1-18, pulse.Event{Glyph: "!"})
			empty := newGapChart(m.mode, false, pulse.WithGrid(true))
			empty.AddEventAt(gapW-1-18, pulse.Event{Glyph: "!"})
			rows, want := plotRows(c), plotRows(empty)
			for col := 15; col < 22; col++ {
				if colRunes(rows, col) != colRunes(want, col) {
					t.Fatalf("column %d: got %q, want the empty chart's %q", col, colRunes(rows, col), colRunes(want, col))
				}
			}
		})
	}
}

func TestGap_ZeroBaselineNegativeStyle(t *testing.T) {
	neg := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	for _, m := range gapModes {
		t.Run(m.name, func(t *testing.T) {
			c := pulse.New(gapW, 9, pulse.WithRange(-50, 50), pulse.WithZeroBaseline(true),
				pulse.WithNegativeStyle(neg), pulse.WithFill(true), pulse.WithGrid(false),
				pulse.WithLineWidth(1), pulse.WithMode(m.mode))
			for i := 0; i < gapW; i++ {
				v := 40 * math.Sin(float64(i)*0.3)
				if i >= 15 && i < 22 {
					v = math.NaN()
				}
				c.Push(v)
			}
			rows := plotRows(c)
			for col := 15; col < 22; col++ {
				for r, row := range rows {
					want := ' '
					if r == 4 {
						want = '─'
					}
					if row[col] != want {
						t.Fatalf("row %d col %d: got %q, want %q", r, col, row[col], want)
					}
				}
			}
			assertDrawnCols(t, rows, 0, 15)
			assertDrawnCols(t, rows, 22, gapW)
		})
	}
}

func TestGap_LastSkipsTrailingGaps(t *testing.T) {
	c := pulse.New(10, 5, pulse.WithRange(0, 100))
	c.PushSeries("cpu", 12.5)
	c.PushSeries("cpu", math.NaN())
	c.PushSeries("cpu", math.Inf(1))
	if v, ok := c.Last("cpu"); !ok || v != 12.5 {
		t.Fatalf("Last() = %v, %v; want 12.5, true", v, ok)
	}
	if box := c.LegendBox(); !strings.Contains(box, "12.5") {
		t.Fatalf("LegendBox() must show the last finite value, got:\n%s", box)
	}
}

func TestGap_LastOfAllGapsIsFalse(t *testing.T) {
	c := pulse.New(10, 5, pulse.WithRange(0, 100))
	c.PushSeries("cpu", math.NaN())
	c.PushSeries("cpu", math.Inf(-1))
	if v, ok := c.Last("cpu"); ok || v != 0 {
		t.Fatalf("Last() = %v, %v; want 0, false", v, ok)
	}
	box := ansi.Strip(c.LegendBox())
	for _, bad := range []string{"NaN", "Inf"} {
		if strings.Contains(box, bad) {
			t.Fatalf("LegendBox() must show no value, got:\n%s", box)
		}
	}
}

func TestGap_SeriesDataKeepsRawValues(t *testing.T) {
	c := pulse.New(10, 5)
	c.Push(1)
	c.Push(math.NaN())
	c.Push(math.Inf(1))
	got := c.SeriesData("")
	if len(got) != 3 || !math.IsNaN(got[1]) || !math.IsInf(got[2], 1) {
		t.Fatalf("SeriesData() = %v, want raw values with NaN and +Inf", got)
	}
}
