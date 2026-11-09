package pulse_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/ingvarch/pulse"
)

func logOpts(mode pulse.RenderMode, opts ...pulse.Option) []pulse.Option {
	return append([]pulse.Option{
		pulse.WithMode(mode),
		pulse.WithLogScale(true),
		pulse.WithRange(1, 1000),
		pulse.WithGrid(false),
		pulse.WithFill(false),
		pulse.WithLineWidth(1),
	}, opts...)
}

// yLabels returns the Y label of each row that has one.
func yLabels(m *pulse.Model) map[int]string {
	out := map[int]string{}
	for r, line := range strings.Split(ansi.Strip(m.View()), "\n") {
		if label, _, ok := strings.Cut(line, "│"); ok && strings.TrimSpace(label) != "" {
			out[r] = strings.TrimSpace(label)
		}
	}
	return out
}

func assertLabels(t *testing.T, m *pulse.Model, want map[int]string) {
	t.Helper()
	got := yLabels(m)
	if len(got) != len(want) {
		t.Fatalf("labels = %v, want %v", got, want)
	}
	for r, w := range want {
		if got[r] != w {
			t.Fatalf("labels = %v, want %v", got, want)
		}
	}
}

func TestLogScale_GetterAndSetter(t *testing.T) {
	m := pulse.New(10, 5)
	if m.LogScale() {
		t.Fatal("log scale must be off by default")
	}
	m.SetLogScale(true)
	if !m.LogScale() {
		t.Fatal("SetLogScale(true) did not turn it on")
	}
	if !pulse.New(10, 5, pulse.WithLogScale(true)).LogScale() {
		t.Fatal("WithLogScale(true) did not turn it on")
	}
}

func TestLogScale_PowersOfTenOnEvenRows(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		for v, want := range map[float64]int{1: 9, 10: 6, 100: 3, 1000: 0} {
			m := pulse.New(1, 10, logOpts(mode)...)
			m.Push(v)
			if got := rowOfGlyph(t, m, 0); got != want {
				t.Errorf("value %v on row %d, want %d", v, got, want)
			}
		}
	})
}

func TestLogScale_TickRowsAndLabelsForSeveralDecades(t *testing.T) {
	m := pulse.New(10, 13, pulse.WithLogScale(true), pulse.WithRange(1, 1000))
	assertLabels(t, m, map[int]string{12: "1", 8: "10", 4: "100", 0: "1000"})

	m = pulse.New(10, 13, pulse.WithLogScale(true), pulse.WithRange(1, 1e6))
	assertLabels(t, m, map[int]string{12: "1", 8: "100", 4: "1e+04", 0: "1e+06"})
}

func TestLogScale_TickRowsAndLabelsForRangeNarrowerThanADecade(t *testing.T) {
	m := pulse.New(10, 10, pulse.WithLogScale(true), pulse.WithRange(2, 8))
	assertLabels(t, m, map[int]string{9: "2", 3: "5"})
}

func TestLogScale_GridFollowsTickRows(t *testing.T) {
	m := pulse.New(5, 13, pulse.WithLogScale(true), pulse.WithRange(1, 1000), pulse.WithGrid(true))
	rows := plotRows(m)
	for r, row := range rows {
		hasGrid := strings.ContainsRune(string(row), '─') || strings.ContainsRune(string(row), '┼')
		if want := r == 0 || r == 4 || r == 8 || r == 12; hasGrid != want {
			t.Fatalf("row %d grid = %v, want %v", r, hasGrid, want)
		}
	}
}

func TestLogScale_CustomTicksMapThroughLog(t *testing.T) {
	m := pulse.New(10, 10,
		pulse.WithLogScale(true), pulse.WithRange(1, 1000),
		pulse.WithTicks(-5, 0, 10, 5000),
		pulse.WithLabelFormatter(func(v float64) string { return "v" + strconv.FormatFloat(v, 'f', -1, 64) }),
	)
	assertLabels(t, m, map[int]string{6: "v10"})
}

func TestLogScale_ValueAtOrBelowZeroIsAGap(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := pulse.New(5, 10, logOpts(mode)...)
		for _, v := range []float64{10, 0, 10, -5, 10} {
			m.Push(v)
		}
		rows := plotRows(m)
		for _, col := range []int{1, 3} {
			if got := strings.TrimSpace(colRunes(rows, col)); got != "" {
				t.Errorf("column %d = %q, want a gap", col, got)
			}
		}
		if got := strings.TrimSpace(colRunes(rows, 2)); got == "" {
			t.Error("column 2 is blank, want the value 10")
		}
		if v, ok := m.Last(""); !ok || v != 10 {
			t.Errorf("Last = %v, %v, want 10", v, ok)
		}
	})
}

func TestLogScale_NonPositiveMinimumBecomesMaxOverTen(t *testing.T) {
	for _, min := range []float64{0, -50} {
		m := pulse.New(10, 10, pulse.WithLogScale(true), pulse.WithRange(min, 1000))
		assertLabels(t, m, map[int]string{9: "100", 0: "1000"})
		if m.Min() != min || m.Max() != 1000 {
			t.Errorf("range = %v..%v, want the stored %v..1000", m.Min(), m.Max(), min)
		}
	}
}

func TestLogScale_NonPositiveBoundsBecomeOneToTen(t *testing.T) {
	m := pulse.New(10, 10, pulse.WithLogScale(true), pulse.WithRange(-10, -1))
	assertLabels(t, m, map[int]string{9: "1", 0: "10"})
	if m.Min() != -10 || m.Max() != -1 {
		t.Errorf("range = %v..%v, want the stored -10..-1", m.Min(), m.Max())
	}
}

func TestLogScale_FillReachesTheBottomRow(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		for _, fill := range []bool{true, false} {
			m := pulse.New(4, 10, logOpts(mode, pulse.WithFill(fill))...)
			for _, v := range []float64{10, 100, 1000, 100} {
				m.Push(v)
			}
			bottom := string(plotRows(m)[9])
			if got := strings.TrimSpace(bottom) != ""; got != fill {
				t.Errorf("fill=%v: bottom row %q", fill, bottom)
			}
			if fill && strings.Contains(bottom, " ") {
				t.Errorf("fill=true: bottom row %q has a blank cell", bottom)
			}
		}
	})
}

func TestLogScale_FillIgnoresTheLinearZeroRow(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := pulse.New(4, 10, logOpts(mode, pulse.WithRange(-1000, 1000), pulse.WithFill(true))...)
		for _, v := range []float64{200, 500, 1000, 300} {
			m.Push(v)
		}
		if bottom := string(plotRows(m)[9]); strings.Contains(bottom, " ") {
			t.Errorf("bottom row %q has a blank cell, want the fill to reach it", bottom)
		}
	})
}

func TestLogScale_SeveralSeries(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := pulse.New(1, 10, logOpts(mode)...)
		m.PushSeries("a", 10)
		m.PushSeries("b", 1000)
		rows := plotRows(m)
		for _, r := range []int{0, 6} {
			if strings.TrimSpace(string(rows[r])) == "" {
				t.Errorf("row %d is blank, want a series point", r)
			}
		}
	})
}

func TestLogScale_AutoRangeWidensToDecades(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := pulse.New(10, 10, pulse.WithMode(mode), pulse.WithLogScale(true), pulse.WithAutoRange(true))
		for _, v := range []float64{5, 300, 0, -2} {
			m.Push(v)
		}
		m.View()
		if m.Min() != 1 || m.Max() != 1000 {
			t.Fatalf("range = %v..%v, want 1..1000", m.Min(), m.Max())
		}
	})
}

func TestLogScale_AfterLinearAutoRangeDrawsTheLastRange(t *testing.T) {
	m := pulse.New(10, 10, pulse.WithAutoRange(true))
	m.Push(120)
	m.Push(730)
	m.View()
	m.SetAutoRange(false)
	m.SetLogScale(true)
	if m.Min() != 0 || m.Max() != 750 {
		t.Fatalf("range = %v..%v, want the stored 0..750", m.Min(), m.Max())
	}
	assertLabels(t, m, map[int]string{8: "100", 5: "200", 2: "500"})
}

func TestLogScale_AutoRangeEdgeCases(t *testing.T) {
	m := pulse.New(10, 10, pulse.WithLogScale(true), pulse.WithAutoRange(true))
	m.View()
	if m.Min() != 1 || m.Max() != 10 {
		t.Errorf("no data: range = %v..%v, want 1..10", m.Min(), m.Max())
	}
	m.Push(0)
	m.Push(-3)
	m.View()
	if m.Min() != 1 || m.Max() != 10 {
		t.Errorf("no positive values: range = %v..%v, want 1..10", m.Min(), m.Max())
	}
	m.Push(100)
	m.View()
	if m.Min() != 100 || m.Max() != 1000 {
		t.Errorf("one power: range = %v..%v, want 100..1000", m.Min(), m.Max())
	}
}

func TestLogScale_ThresholdFollowsTheLogMapping(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := pulse.New(10, 10, logOpts(mode,
			pulse.WithThresholds(pulse.Threshold{Value: 100, Style: thWarn, Line: true}),
		)...)
		rows := plotRows(m)
		for r, row := range rows {
			if has := strings.ContainsRune(string(row), '┄'); has != (r == 3) {
				t.Fatalf("row %d threshold glyph = %v, want %v", r, has, r == 3)
			}
		}
	})
}

func TestLogScale_ThresholdColorsFollowTheLogMapping(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		opts := logOpts(mode,
			pulse.WithLineStyle(thBase),
			pulse.WithThresholds(pulse.Threshold{Value: 100, Style: thWarn}),
			pulse.WithThresholdColors(true),
		)
		m := pulse.New(1, 10, opts...)
		m.Push(1000)
		if c := cellAt(m, 0, 0); !styledAs(c, thWarn) {
			t.Fatalf("cell at 1000 = %q, want the threshold style", c)
		}
		m = pulse.New(1, 10, opts...)
		m.Push(10)
		if c := cellAt(m, 6, 0); !styledAs(c, thBase) {
			t.Fatalf("cell at 10 = %q, want the series style", c)
		}
	})
}

func TestLogScale_NonPositiveThresholdDrawsNothing(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := pulse.New(1, 10, logOpts(mode,
			pulse.WithLineStyle(thBase),
			pulse.WithThresholds(
				pulse.Threshold{Value: 0, Style: thWarn, Line: true},
				pulse.Threshold{Value: -5, Style: thCrit, Line: true},
			),
			pulse.WithThresholdColors(true),
		)...)
		m.Push(1000)
		if strings.Contains(ansi.Strip(m.View()), "┄") {
			t.Fatalf("a non-positive threshold drew a line:\n%s", m.View())
		}
		if c := cellAt(m, 0, 0); !styledAs(c, thBase) {
			t.Fatalf("cell = %q, want the series style", c)
		}
	})
}

func TestLogScale_OffRestoresTheLinearChart(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		linear := pulse.New(10, 10, pulse.WithMode(mode), pulse.WithRange(0, 1000))
		m := pulse.New(10, 10, pulse.WithMode(mode), pulse.WithRange(0, 1000), pulse.WithLogScale(true))
		for _, v := range []float64{5, 40, 300, 900, 20} {
			linear.Push(v)
			m.Push(v)
		}
		if m.View() == linear.View() {
			t.Fatal("the log chart looks like the linear one")
		}
		m.SetLogScale(false)
		if m.View() != linear.View() {
			t.Fatalf("log scale off:\n%s\nwant:\n%s", m.View(), linear.View())
		}
	})
}

func TestLogScale_IgnoresLinearOnlyOptions(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		base := []pulse.Option{pulse.WithMode(mode), pulse.WithLogScale(true), pulse.WithRange(0, 1000), pulse.WithFill(true)}
		plain := pulse.New(6, 10, base...)
		m := pulse.New(6, 10, append(base,
			pulse.WithZeroBaseline(true),
			pulse.WithSymmetric(true),
			pulse.WithNegativeStyle(thNeg),
		)...)
		for _, v := range []float64{3, 30, 300, 30, 3, 30} {
			plain.Push(v)
			m.Push(v)
		}
		if m.View() != plain.View() {
			t.Fatalf("linear-only options changed the log chart:\n%s\nwant:\n%s", m.View(), plain.View())
		}
	})
}

func TestLogScale_SymmetricKeepsTheFixedRange(t *testing.T) {
	want := map[int]string{12: "1", 8: "10", 4: "100", 0: "1000"}
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := pulse.New(10, 13, pulse.WithMode(mode), pulse.WithSymmetric(true), pulse.WithRange(1, 1000), pulse.WithLogScale(true))
		assertLabels(t, m, want)
		m = pulse.New(10, 13, pulse.WithMode(mode), pulse.WithLogScale(true), pulse.WithRange(1, 1000))
		m.SetSymmetric(true)
		assertLabels(t, m, want)
		m.SetRange(1, 1000)
		assertLabels(t, m, want)
	})
}

func TestLogScale_ValuesLandOnRowsWhenTheMinimumIsNotPositive(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		for _, tc := range []struct {
			name string
			opts []pulse.Option
			rows map[float64]int
		}{
			{"zero minimum", []pulse.Option{pulse.WithRange(0, 1000)}, map[float64]int{100: 9, 1000: 0}},
			{"negative bounds", []pulse.Option{pulse.WithRange(-10, -1)}, map[float64]int{1: 9, 10: 0}},
			{"default range", nil, map[float64]int{10: 9, 100: 0}},
		} {
			for v, want := range tc.rows {
				m := pulse.New(1, 10, append([]pulse.Option{pulse.WithMode(mode), pulse.WithLogScale(true), pulse.WithFill(false), pulse.WithGrid(false), pulse.WithLineWidth(1)}, tc.opts...)...)
				m.Push(v)
				if got := rowOfGlyph(t, m, 0); got != want {
					t.Errorf("%s: value %v on row %d, want %d", tc.name, v, got, want)
				}
			}
		}
	})
}

func TestLogScale_ManyDecadesDoNotOverflow(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := pulse.New(1, 11, logOpts(mode, pulse.WithRange(1e-200, 1e200))...)
		m.Push(1)
		if got := rowOfGlyph(t, m, 0); got != 5 {
			t.Fatalf("value 1 on row %d, want the middle row 5", got)
		}
	})
}

func TestLogScale_InvertedSeriesIsDrawnAtItsRealValue(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := pulse.New(1, 10, logOpts(mode, pulse.WithInverted(true))...)
		m.Push(100)
		if got := rowOfGlyph(t, m, 0); got != 3 {
			t.Fatalf("inverted 100 on row %d, want 3", got)
		}
		if v, _ := m.Last(""); v != 100 {
			t.Fatalf("Last = %v, want 100", v)
		}
	})
}
