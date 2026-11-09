package pulse_test

import (
	"math"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/ingvarch/pulse"
)

func TestSparkline_Levels(t *testing.T) {
	cases := []struct {
		name     string
		values   []float64
		min, max float64
		want     string
	}{
		{"min is the lowest block", []float64{0}, 0, 100, "▁"},
		{"max is the full block", []float64{100}, 0, 100, "█"},
		{"middle is the fifth block", []float64{50}, 0, 100, "▅"},
		{"all eight blocks", []float64{0, 1, 2, 3, 4, 5, 6, 7}, 0, 7, "▁▂▃▄▅▆▇█"},
		{"below min clamps", []float64{-50}, 0, 100, "▁"},
		{"above max clamps", []float64{500}, 0, 100, "█"},
		{"NaN is a space", []float64{0, math.NaN(), 100}, 0, 100, "▁ █"},
		{"+Inf is a space", []float64{math.Inf(1)}, 0, 100, " "},
		{"-Inf is a space", []float64{math.Inf(-1)}, 0, 100, " "},
		{"empty", []float64{}, 0, 100, ""},
		{"nil", nil, 0, 100, ""},
		{"equal bounds give the lowest block", []float64{5, 5}, 5, 5, "▁▁"},
		{"min above max is swapped", []float64{0, 100}, 100, 0, "▁█"},
		{"equal bounds work as min..min+1", []float64{5, 5.5, 6}, 5, 5, "▁▅█"},
		{"NaN min makes every cell a gap", []float64{1, 2}, math.NaN(), 10, "  "},
		{"infinite bounds make every cell a gap", []float64{1, 2}, math.Inf(-1), math.Inf(1), "  "},
		{"a value whose distance from min overflows is a gap", []float64{math.MaxFloat64}, -math.MaxFloat64, math.MaxFloat64, " "},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := pulse.Sparkline(tc.values, tc.min, tc.max); got != tc.want {
				t.Fatalf("Sparkline = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestSparklineBraille(t *testing.T) {
	nan := math.NaN()
	cases := []struct {
		name     string
		values   []float64
		min, max float64
		want     string
	}{
		{"min in the left column, max in the right", []float64{0, 100}, 0, 100, "⣸"},
		{"max in the left column, min in the right", []float64{100, 0}, 0, 100, "\u28c7"},
		{"odd count leaves the right column empty", []float64{0, 100, 100}, 0, 100, "⣸⡇"},
		{"gap in the left column", []float64{nan, 100}, 0, 100, "⢸"},
		{"gap in the right column", []float64{100, nan}, 0, 100, "⡇"},
		{"two gaps are a space", []float64{0, 100, nan, math.Inf(1)}, 0, 100, "⣸ "},
		{"lone trailing gap is a space", []float64{100, nan, nan}, 0, 100, "⡇ "},
		{"middle value is three dots high", []float64{50, 50}, 0, 100, "\u28f6"},
		{"clamps", []float64{-5, 500}, 0, 100, "⣸"},
		{"empty", nil, 0, 100, ""},
		{"min above max is swapped", []float64{0, 100}, 100, 0, "⣸"},
		{"equal bounds give the bottom dots", []float64{3, 3}, 3, 3, "⣀"},
		{"equal bounds work as min..min+1", []float64{5, 6}, 5, 5, "⣸"},
		{"NaN max makes every cell a gap", []float64{1, 2, 3}, 0, math.NaN(), "  "},
		{"infinite bounds make every cell a gap", []float64{1, 2}, math.Inf(-1), math.Inf(1), " "},
		{"a value whose distance from min overflows is a gap", []float64{math.MaxFloat64}, -math.MaxFloat64, math.MaxFloat64, " "},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := pulse.SparklineBraille(tc.values, tc.min, tc.max); got != tc.want {
				t.Fatalf("SparklineBraille = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestSparklineBraille_Width(t *testing.T) {
	for n, want := range map[int]int{1: 1, 2: 1, 3: 2, 8: 4} {
		if got := lipgloss.Width(pulse.SparklineBraille(make([]float64, n), 0, 1)); got != want {
			t.Errorf("%d values: width %d, want %d", n, got, want)
		}
	}
}

func newSparkChart(opts ...pulse.Option) *pulse.Model {
	return pulse.New(8, 5, append([]pulse.Option{pulse.WithHistory(100)}, opts...)...)
}

func TestModelSparkline_FixedRange(t *testing.T) {
	m := newSparkChart(pulse.WithRange(0, 100))
	for _, v := range []float64{0, 50, 100, 200} {
		m.Push(v)
	}
	if got, want := m.Sparkline(""), m.SeriesStyle("").Render("▁▅██"); got != want {
		t.Fatalf("Sparkline = %q, want %q", got, want)
	}
}

func TestModelSparkline_UsesSeriesStyle(t *testing.T) {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).Bold(true)
	m := newSparkChart(pulse.WithRange(0, 100), pulse.WithSeriesStyle("cpu", style))
	m.PushSeries("cpu", 0)
	m.PushSeries("cpu", 100)
	if got, want := m.Sparkline("cpu"), style.Render("▁█"); got != want {
		t.Fatalf("Sparkline = %q, want %q", got, want)
	}
}

func TestModelSparkline_AutoRangeWithoutView(t *testing.T) {
	m := newSparkChart(pulse.WithAutoRange(true))
	for _, v := range []float64{2, 4, 6, 8, 10} {
		m.Push(v)
	}
	got := m.Sparkline("")
	lo, hi := m.Min(), m.Max()
	if lo == 0 && hi == 100 {
		t.Fatal("Sparkline did not store the auto range in Min and Max")
	}
	m.View()
	if m.Min() != lo || m.Max() != hi {
		t.Fatalf("View range %v..%v, Sparkline stored %v..%v", m.Min(), m.Max(), lo, hi)
	}
	want := m.SeriesStyle("").Render(pulse.Sparkline([]float64{2, 4, 6, 8, 10}, lo, hi))
	if got != want {
		t.Fatalf("Sparkline = %q, want %q (range %v..%v)", got, want, lo, hi)
	}
	if got == m.SeriesStyle("").Render(pulse.Sparkline([]float64{2, 4, 6, 8, 10}, 0, 100)) {
		t.Fatal("the auto range was not applied")
	}
}

func TestModelSparkline_FollowsOffset(t *testing.T) {
	m := newSparkChart(pulse.WithRange(0, 7))
	for i := range 20 {
		m.Push(float64(i % 8))
	}
	m.SetOffset(4)
	want := m.SeriesStyle("").Render("▁▂▃▄▅▆▇█")
	if got := m.Sparkline(""); got != want {
		t.Fatalf("Sparkline = %q, want %q", got, want)
	}
}

func TestModelSparkline_Inverted(t *testing.T) {
	m := newSparkChart(pulse.WithRange(-100, 0))
	m.SetInverted(true)
	m.Push(0)
	m.Push(100)
	if got, want := m.Sparkline(""), m.SeriesStyle("").Render("█▁"); got != want {
		t.Fatalf("Sparkline = %q, want %q", got, want)
	}

	m = newSparkChart(pulse.WithRange(-100, 0), pulse.WithSeriesInverted("rx", true))
	m.PushSeries("rx", 0)
	m.PushSeries("rx", 100)
	if got, want := m.Sparkline("rx"), m.SeriesStyle("rx").Render("█▁"); got != want {
		t.Fatalf("named: Sparkline = %q, want %q", got, want)
	}
}

func TestModelSparkline_RightAxis(t *testing.T) {
	m := newSparkChart(
		pulse.WithRange(0, 100),
		pulse.WithRightAxis(pulse.Axis{Min: 0, Max: 1000}),
		pulse.WithSeriesRightAxis("net"),
	)
	m.PushSeries("net", 0)
	m.PushSeries("net", 500)
	m.PushSeries("net", 1000)
	if got, want := m.Sparkline("net"), m.SeriesStyle("net").Render("▁▅█"); got != want {
		t.Fatalf("Sparkline = %q, want %q", got, want)
	}
}

func TestModelSparkline_RightAxisStaysLinearOnLogScale(t *testing.T) {
	m := newSparkChart(
		pulse.WithLogScale(true),
		pulse.WithRange(1, 1000),
		pulse.WithRightAxis(pulse.Axis{Min: 0, Max: 1000}),
		pulse.WithSeriesRightAxis("net"),
	)
	m.PushSeries("net", 0)
	m.PushSeries("net", 500)
	m.PushSeries("net", 1000)
	if got, want := m.Sparkline("net"), m.SeriesStyle("net").Render("▁▅█"); got != want {
		t.Fatalf("Sparkline = %q, want %q", got, want)
	}
}

func TestModelSparkline_InvertedRightAxis(t *testing.T) {
	m := newSparkChart(
		pulse.WithRightAxis(pulse.Axis{Min: -1000, Max: 0}),
		pulse.WithSeriesRightAxis("tx"),
		pulse.WithSeriesInverted("tx", true),
	)
	m.PushSeries("tx", 0)
	m.PushSeries("tx", 1000)
	if got, want := m.Sparkline("tx"), m.SeriesStyle("tx").Render("█▁"); got != want {
		t.Fatalf("Sparkline = %q, want %q", got, want)
	}
}

func TestModelSparkline_LogScale(t *testing.T) {
	m := newSparkChart(pulse.WithRange(1, 1000), pulse.WithLogScale(true))
	for _, v := range []float64{1, 10, 100, 1000, 0, -5} {
		m.Push(v)
	}
	if got, want := m.Sparkline(""), m.SeriesStyle("").Render("▁▃▆█  "); got != want {
		t.Fatalf("Sparkline = %q, want %q", got, want)
	}
}

func TestModelSparkline_Empty(t *testing.T) {
	m := newSparkChart(pulse.WithRange(0, 100))
	if got := m.Sparkline(""); got != "" {
		t.Fatalf("no points: Sparkline = %q, want empty", got)
	}
	m.PushSeries("cpu", 10)
	if got := m.Sparkline("nope"); got != "" {
		t.Fatalf("unknown name: Sparkline = %q, want empty", got)
	}
	if m.Sparkline("cpu") == "" {
		t.Fatal("known series returned empty")
	}
}

func TestModelSparkline_DefaultSeriesByEmptyName(t *testing.T) {
	m := newSparkChart(pulse.WithRange(0, 100))
	m.Push(100)
	if got, want := m.Sparkline(""), m.SeriesStyle("").Render("█"); got != want {
		t.Fatalf("Sparkline = %q, want %q", got, want)
	}
}
