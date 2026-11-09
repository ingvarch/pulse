package pulse_test

import (
	"math"
	"strings"
	"testing"

	"github.com/ingvarch/pulse"
	"github.com/ingvarch/pulse/scale"
)

func forModes(t *testing.T, f func(t *testing.T, mode pulse.RenderMode)) {
	t.Helper()
	for _, tc := range historyModes {
		t.Run(tc.name, func(t *testing.T) { f(t, tc.mode) })
	}
}

func firstRow(m *pulse.Model) string {
	return strings.SplitN(m.View(), "\n", 2)[0]
}

func TestAutoRange_RoundTopAboveMaxIsLabeled(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := pulse.New(20, 10, pulse.WithMode(mode), pulse.WithAutoRange(true))
		for _, v := range []float64{120, 300, 730, 410, 90} {
			m.Push(v)
		}
		m.View()
		if m.Min() != 0 {
			t.Fatalf("Min = %v, want 0", m.Min())
		}
		if m.Max() <= 730 || math.Mod(m.Max(), 50) != 0 {
			t.Fatalf("Max = %v, want a round top above 730", m.Max())
		}
		if got, want := strings.TrimSpace(strings.SplitN(firstRow(m), "│", 2)[0]), scale.FormatTick(m.Max()); got != want {
			t.Fatalf("top label = %q, want %q", got, want)
		}
	})
}

func TestAutoRange_ShrinksAfterSpikeLeavesWindow(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := pulse.New(10, 10, pulse.WithMode(mode), pulse.WithAutoRange(true), pulse.WithHistory(50))
		m.Push(5000)
		for range 5 {
			m.Push(10)
		}
		m.View()
		high := m.Max()
		for range 10 {
			m.Push(10)
		}
		m.View()
		if m.Max() >= high || m.Max() > 20 {
			t.Fatalf("Max = %v after the spike left the window, was %v", m.Max(), high)
		}
	})
}

func TestAutoRange_AllPositiveStartsAtZero(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := pulse.New(10, 10, pulse.WithMode(mode), pulse.WithAutoRange(true))
		for _, v := range []float64{500, 600, 700} {
			m.Push(v)
		}
		m.View()
		if m.Min() != 0 {
			t.Fatalf("Min = %v, want 0", m.Min())
		}
	})
}

func TestAutoRange_CoversBothSidesOfZero(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := pulse.New(10, 10, pulse.WithMode(mode), pulse.WithAutoRange(true))
		for _, v := range []float64{-30, 70, -10} {
			m.Push(v)
		}
		m.View()
		if m.Min() > -30 || m.Max() < 70 {
			t.Fatalf("range %v..%v does not cover -30..70", m.Min(), m.Max())
		}
	})
}

func TestAutoRange_SymmetricIsMirrored(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := pulse.New(10, 11, pulse.WithMode(mode), pulse.WithAutoRange(true), pulse.WithSymmetric(true))
		for _, v := range []float64{-30, 70, -10} {
			m.Push(v)
		}
		m.View()
		if m.Min() != -m.Max() || m.Max() < 70 {
			t.Fatalf("range %v..%v, want mirrored and covering 70", m.Min(), m.Max())
		}
		if !m.AutoRange() {
			t.Fatal("WithSymmetric turned auto range off")
		}
	})
}

func TestAutoRange_InvertedSeriesCountsWithFlippedSign(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := pulse.New(10, 10, pulse.WithMode(mode), pulse.WithAutoRange(true),
			pulse.WithSeriesInverted("tx", true))
		for _, v := range []float64{40, 80} {
			m.PushSeries("tx", v)
		}
		m.View()
		if m.Max() != 0 || m.Min() > -80 {
			t.Fatalf("range %v..%v, want to cover -80..0", m.Min(), m.Max())
		}
	})
}

func TestAutoRange_CountsAllSeriesAndTheDefault(t *testing.T) {
	m := pulse.New(10, 10, pulse.WithAutoRange(true))
	m.Push(900)
	m.PushSeries("a", 10)
	m.PushSeries("b", 300)
	m.View()
	if m.Max() < 900 {
		t.Fatalf("Max = %v, want at least 900", m.Max())
	}
	neg := pulse.New(10, 10, pulse.WithAutoRange(true))
	neg.Push(-900)
	neg.PushSeries("a", 10)
	neg.View()
	if neg.Min() > -900 {
		t.Fatalf("Min = %v, want at most -900", neg.Min())
	}
}

func TestAutoRange_NoFiniteValuesFallsBackToZeroOne(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := pulse.New(10, 10, pulse.WithMode(mode), pulse.WithAutoRange(true))
		m.Push(math.NaN())
		m.Push(math.Inf(1))
		m.Push(math.Inf(-1))
		m.View()
		if m.Min() != 0 || m.Max() != 1 {
			t.Fatalf("range %v..%v, want 0..1", m.Min(), m.Max())
		}
		sym := pulse.New(10, 10, pulse.WithMode(mode), pulse.WithAutoRange(true), pulse.WithSymmetric(true))
		sym.Push(math.NaN())
		sym.View()
		if sym.Min() != -1 || sym.Max() != 1 {
			t.Fatalf("symmetric range %v..%v, want -1..1", sym.Min(), sym.Max())
		}
	})
}

func TestAutoRange_EmptyChartFallsBackToZeroOne(t *testing.T) {
	m := pulse.New(10, 10, pulse.WithAutoRange(true))
	m.View()
	if m.Min() != 0 || m.Max() != 1 {
		t.Fatalf("range %v..%v, want 0..1", m.Min(), m.Max())
	}
}

func TestAutoRange_FlatZeroFallsBackToZeroOne(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := pulse.New(10, 10, pulse.WithMode(mode), pulse.WithAutoRange(true))
		for range 5 {
			m.Push(0)
		}
		m.View()
		if m.Min() != 0 || m.Max() != 1 {
			t.Fatalf("range %v..%v, want 0..1", m.Min(), m.Max())
		}
	})
}

func TestAutoRange_IgnoresGaps(t *testing.T) {
	m := pulse.New(10, 10, pulse.WithAutoRange(true))
	m.Push(40)
	m.Push(math.Inf(1))
	m.Push(math.NaN())
	m.View()
	if m.Max() < 40 || m.Max() > 50 || math.IsInf(m.Max(), 0) {
		t.Fatalf("Max = %v, want a round top near 40", m.Max())
	}
}

func TestAutoRange_WithRangeAfterTurnsItOff(t *testing.T) {
	m := pulse.New(10, 10, pulse.WithAutoRange(true), pulse.WithRange(0, 100))
	m.Push(730)
	m.View()
	if m.AutoRange() || m.Min() != 0 || m.Max() != 100 {
		t.Fatalf("AutoRange=%v range %v..%v, want off and 0..100", m.AutoRange(), m.Min(), m.Max())
	}
}

func TestAutoRange_SetRangeTurnsItOff(t *testing.T) {
	m := pulse.New(10, 10, pulse.WithAutoRange(true))
	m.SetRange(0, 50)
	if m.AutoRange() {
		t.Fatal("SetRange left auto range on")
	}
}

func TestAutoRange_IgnoredSetRangeKeepsItOn(t *testing.T) {
	m := pulse.New(10, 10, pulse.WithAutoRange(true))
	m.SetRange(math.NaN(), 5)
	if !m.AutoRange() {
		t.Fatal("an ignored SetRange turned auto range off")
	}
}

func TestAutoRange_WithAutoRangeAfterRangeTurnsItOn(t *testing.T) {
	m := pulse.New(10, 10, pulse.WithRange(0, 100), pulse.WithAutoRange(true))
	m.Push(730)
	m.View()
	if !m.AutoRange() || m.Max() < 730 {
		t.Fatalf("AutoRange=%v Max=%v, want on and at least 730", m.AutoRange(), m.Max())
	}
}

func TestAutoRange_SetSymmetricKeepsItOn(t *testing.T) {
	m := pulse.New(10, 10, pulse.WithAutoRange(true))
	m.SetSymmetric(true)
	if !m.AutoRange() {
		t.Fatal("SetSymmetric turned auto range off")
	}
}

func TestAutoRange_SetAutoRangeOffKeepsLastRange(t *testing.T) {
	m := pulse.New(10, 10, pulse.WithAutoRange(true))
	m.Push(730)
	m.View()
	max := m.Max()
	m.SetAutoRange(false)
	m.Push(5000)
	m.View()
	if m.Max() != max {
		t.Fatalf("Max = %v, want the last auto range %v", m.Max(), max)
	}
}

func TestAutoRange_SetSymmetricAfterOffMirrorsTheLastRange(t *testing.T) {
	m := pulse.New(10, 10, pulse.WithAutoRange(true))
	m.Push(120)
	m.Push(730)
	m.View()
	max := m.Max()
	m.SetAutoRange(false)
	m.SetSymmetric(true)
	if m.Min() != -max || m.Max() != max {
		t.Fatalf("range = %v..%v, want the last auto range mirrored: %v..%v", m.Min(), m.Max(), -max, max)
	}
}

func TestAutoRange_CustomTicksLabelOnlyInsideRange(t *testing.T) {
	m := pulse.New(10, 10, pulse.WithAutoRange(true), pulse.WithTicks(0, 500, 5000))
	m.Push(730)
	view := m.View()
	if strings.Contains(view, "5000") {
		t.Fatalf("a tick outside the range got a label:\n%s", view)
	}
	if !strings.Contains(view, "500") {
		t.Fatalf("a tick inside the range has no label:\n%s", view)
	}
}

func TestAutoRange_ZeroBaselineAndXAxis(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := pulse.New(20, 11, pulse.WithMode(mode), pulse.WithAutoRange(true), pulse.WithSymmetric(true),
			pulse.WithZeroBaseline(true), pulse.WithXAxis(func(age int) string { return "" }))
		for _, v := range []float64{-300, 730, 20} {
			m.Push(v)
		}
		view := m.View()
		if !strings.Contains(view, "├") {
			t.Fatalf("no zero baseline in:\n%s", view)
		}
		rows := strings.Split(view, "\n")
		if len(rows) != m.Height()+2 {
			t.Fatalf("%d rows, want %d with the X axis", len(rows), m.Height()+2)
		}
		if !strings.Contains(rows[(m.Height()-1)/2], "├") {
			t.Fatalf("zero baseline is not in the middle row:\n%s", view)
		}
		if m.Min() != -m.Max() || m.Max() < 730 {
			t.Fatalf("range %v..%v", m.Min(), m.Max())
		}
	})
}

func TestAutoRange_HugeValueDoesNotHang(t *testing.T) {
	m := pulse.New(10, 10, pulse.WithAutoRange(true))
	m.Push(1.7e308)
	m.View()
	if math.IsInf(m.Max(), 0) || m.Max() < 1.7e308 {
		t.Fatalf("Max = %v", m.Max())
	}
}

func TestAutoRange_RenderersDrawAgainstTheAutoRange(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		// Fill and grid off: only the line can mark the bottom row.
		m := pulse.New(2, 10, pulse.WithMode(mode), pulse.WithAutoRange(true),
			pulse.WithFill(false), pulse.WithGrid(false))
		m.Push(10)
		m.Push(730)
		view := m.View()
		rows := strings.Split(view, "\n")
		// On 0..750 the 10 point is in the bottom row; on a stale 0..100 it is one row higher.
		bottom := []rune(strings.SplitN(rows[len(rows)-1], "│", 2)[1])
		if bottom[0] == ' ' {
			t.Fatalf("the 10 point is missing from the bottom row:\n%s", view)
		}
	})
}

func TestAutoRange_IdleInvertedSeriesKeepsPositiveZero(t *testing.T) {
	forModes(t, func(t *testing.T, mode pulse.RenderMode) {
		m := pulse.New(10, 10, pulse.WithMode(mode), pulse.WithAutoRange(true),
			pulse.WithSeriesInverted("tx", true))
		for range 3 {
			m.PushSeries("rx", 300)
			m.PushSeries("tx", 0)
		}
		rows := strings.Split(m.View(), "\n")
		if math.Signbit(m.Min()) {
			t.Fatalf("Min = %v, want +0", m.Min())
		}
		if got := strings.TrimSpace(strings.SplitN(rows[len(rows)-1], "│", 2)[0]); got != "0" {
			t.Fatalf("bottom label = %q, want %q", got, "0")
		}
	})
}
