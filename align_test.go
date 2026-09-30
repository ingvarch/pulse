package pulse

import (
	"strings"
	"testing"

	"github.com/ingvarch/pulse/scale"
)

// TestLineChart_LabelsSitOnGridlines verifies each label sits exactly on its gridline row.
func TestLineChart_LabelsSitOnGridlines(t *testing.T) {
	m := New(20, 10, WithRange(0, 100), WithGrid(true), WithFill(false))
	rows := strings.Split(m.View(), "\n")
	labels := make(map[int]string)
	for r, ln := range rows {
		runes := []rune(ln)
		if len(runes) < 5 {
			t.Fatalf("line %d too short: %q", r, ln)
		}
		labels[r] = strings.TrimSpace(string(runes[:4]))
	}
	for _, tick := range scale.NiceTicks(0, 100, 4) {
		if tick < 0 || tick > 100 {
			continue
		}
		row := m.tickRow(tick)
		if got := labels[row]; got != scale.FormatTick(tick) && !collides(t, m, tick, labels) {
			t.Errorf("tick %v must label its gridline row %d, row labels %q", tick, row, labels[row])
		}
	}
}

// collides verifies that a row is occupied by the larger tick when two ticks collide.
func collides(t *testing.T, m *Model, tick float64, labels map[int]string) bool {
	t.Helper()
	row := m.tickRow(tick)
	for _, other := range scale.NiceTicks(0, 100, 4) {
		if other > tick && other <= 100 && m.tickRow(other) == row {
			return labels[row] == scale.FormatTick(other)
		}
	}
	return false
}

func TestLineChart_SmoothToggleChangesRender(t *testing.T) {
	data := []float64{10, 30, 50, 70, 90, 70, 50, 30}
	a := New(20, 8, WithRange(0, 100))
	for _, v := range data {
		a.Push(v)
	}
	b := New(20, 8, WithRange(0, 100), WithSmooth(false))
	for _, v := range data {
		b.Push(v)
	}
	if a.View() == b.View() {
		t.Fatal("View() with smooth on and off must differ")
	}
}

// TestLineChart_FillHidesGrid verifies that filled area ░ does not show grid lines ┼ or │.
func TestLineChart_FillHidesGrid(t *testing.T) {
	m := New(40, 10, WithRange(0, 100), WithGrid(true), WithFill(true))
	for i := 0; i < 40; i++ {
		m.Push(50) // bottom half of chart is filled
	}
	view := m.View()
	lines := strings.Split(view, "\n")
	// On lower rows (below line 50), there should be fill ░ and no ┼
	for r := 6; r < len(lines); r++ {
		ln := lines[r]
		if i := strings.Index(ln, "│"); i >= 0 {
			plotArea := ln[i+len("│"):]
			if strings.Contains(plotArea, "┼") {
				t.Fatalf("row %d in filled area contains grid cross ┼: %q", r, plotArea)
			}
			if !strings.Contains(plotArea, "░") {
				t.Fatalf("row %d in filled area should contain fill ░: %q", r, plotArea)
			}
		}
	}
}
