package pulse_test

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"

	"github.com/ingvarch/pulse"
)

// CPU/RAM case: 0-100% values, width W window, rendered as terminal lines.
func TestLineChart_RendersCPUValues(t *testing.T) {
	lc := pulse.New(20, 8, pulse.WithRange(0, 100))
	for _, v := range []float64{10, 30, 50, 70, 90, 70, 50, 30} {
		lc.Push(v)
	}
	view := lc.View()
	if strings.TrimSpace(view) == "" {
		t.Fatal("View() is empty")
	}
	// At least one non-space plot rune (braille, lines, or blocks).
	hasMark := false
	for _, r := range view {
		if r != ' ' && r != '\n' && r != '\r' && r != '│' && r != '─' && r != '┼' && r != '0' && r != '1' && r != '2' && r != '3' && r != '4' && r != '5' && r != '6' && r != '7' && r != '8' && r != '9' {
			hasMark = true
			break
		}
	}
	if !hasMark {
		t.Fatalf("View() has no chart marks:\n%s", view)
	}
}

func TestLineChart_EmptyRendersAxes(t *testing.T) {
	lc := pulse.New(10, 5, pulse.WithRange(0, 100))
	view := lc.View()
	lines := strings.Split(strings.TrimRight(view, "\n"), "\n")
	if len(lines) != 5 {
		t.Fatalf("expected 5 lines, got %d:\n%s", len(lines), view)
	}
}

func TestLineChart_NoDuplicateAdjacentLabels(t *testing.T) {
	lc := pulse.New(20, 10, pulse.WithRange(0, 100))
	lc.Push(50)
	view := lc.View()
	lines := strings.Split(strings.TrimRight(view, "\n"), "\n")
	prev := ""
	for i, ln := range lines {
		runes := []rune(ln)
		if len(runes) < 4 {
			t.Fatalf("line %d too short: %q", i, ln)
		}
		label := strings.TrimSpace(string(runes[:4]))
		if label != "" && label == prev {
			t.Fatalf("duplicate adjacent label %q on lines %d and %d:\n%s", label, i-1, i, view)
		}
		if label != "" {
			prev = label
		}
	}
}

func TestLineChart_SlidingWindow(t *testing.T) {
	lc := pulse.New(4, 4, pulse.WithRange(0, 100))
	for i := 0; i < 10; i++ {
		lc.Push(float64(i * 10))
	}
	if got := lc.Len(); got != 4 {
		t.Fatalf("Len() = %d, want 4 (window = width)", got)
	}
}

func TestLineChart_SetLineStyle(t *testing.T) {
	lc := pulse.New(10, 5, pulse.WithRange(0, 100))
	lc.Push(50)
	lc.SetLineStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")))
	if view := lc.View(); strings.TrimSpace(view) == "" {
		t.Fatal("View() is empty after SetLineStyle")
	}
}

func TestLineChart_MultiSeries(t *testing.T) {
	lc := pulse.New(20, 8, pulse.WithRange(0, 100))
	for i := 0; i < 10; i++ {
		lc.PushSeries("cpu", float64(i*10))
		lc.PushSeries("mem", float64(100-i*5))
	}
	if got := lc.LenSeries("cpu"); got != 10 {
		t.Fatalf("LenSeries(cpu) = %d, want 10", got)
	}
	if got := lc.LenSeries("mem"); got != 10 {
		t.Fatalf("LenSeries(mem) = %d, want 10", got)
	}
	if got := lc.Len(); got != 0 {
		t.Fatalf("Len() = %d, want 0 (nothing pushed to default series)", got)
	}
	view := lc.View()
	if strings.TrimSpace(view) == "" {
		t.Fatal("View() is empty")
	}
	legend := lc.Legend()
	if !strings.Contains(legend, "cpu") || !strings.Contains(legend, "mem") {
		t.Fatalf("Legend() must list cpu and mem, got %q", legend)
	}
}

func TestLineChart_LegendEmptyWithoutNamedSeries(t *testing.T) {
	lc := pulse.New(10, 5, pulse.WithRange(0, 100))
	lc.Push(50)
	if got := lc.Legend(); strings.TrimSpace(got) != "" {
		t.Fatalf("Legend() must be empty without named series, got %q", got)
	}
}

func TestLineChart_Last(t *testing.T) {
	lc := pulse.New(10, 5, pulse.WithRange(0, 100))
	if _, ok := lc.Last("cpu"); ok {
		t.Fatal("Last(cpu) must report false before any push")
	}
	lc.PushSeries("cpu", 42.5)
	lc.PushSeries("cpu", 43.5)
	v, ok := lc.Last("cpu")
	if !ok {
		t.Fatal("Last(cpu) must report true after push")
	}
	if v != 43.5 {
		t.Fatalf("Last(cpu) = %v, want 43.5", v)
	}
}

func TestLineChart_LegendBox(t *testing.T) {
	lc := pulse.New(20, 8, pulse.WithRange(0, 100))
	lc.PushSeries("cpu", 12.5)
	lc.PushSeries("mem", 45)
	box := lc.LegendBox()
	if strings.TrimSpace(box) == "" {
		t.Fatal("LegendBox() is empty")
	}
	for _, want := range []string{"cpu", "12.5", "mem", "45", "╭", "╰"} {
		if !strings.Contains(box, want) {
			t.Fatalf("LegendBox() must contain %q, got:\n%s", want, box)
		}
	}
}

func TestLineChart_LegendBoxEmptyWithoutNamedSeries(t *testing.T) {
	lc := pulse.New(10, 5, pulse.WithRange(0, 100))
	lc.Push(50)
	if got := lc.LegendBox(); strings.TrimSpace(got) != "" {
		t.Fatalf("LegendBox() must be empty without named series, got %q", got)
	}
}

func seedCPU(lc *pulse.Model) {
	for _, v := range []float64{10, 30, 50, 70, 90, 70, 50, 30} {
		lc.Push(v)
	}
}

func TestLineChart_GridToggleChangesRender(t *testing.T) {
	a := pulse.New(20, 8, pulse.WithRange(0, 100))
	seedCPU(a)
	b := pulse.New(20, 8, pulse.WithRange(0, 100), pulse.WithGrid(false))
	seedCPU(b)
	if a.View() == b.View() {
		t.Fatal("View() with grid on and off must differ")
	}
}

func TestLineChart_FillToggleChangesRender(t *testing.T) {
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	a := pulse.New(20, 8, pulse.WithRange(0, 100), pulse.WithLineStyle(red))
	seedCPU(a)
	b := pulse.New(20, 8, pulse.WithRange(0, 100), pulse.WithLineStyle(red), pulse.WithFill(false))
	seedCPU(b)
	if a.View() == b.View() {
		t.Fatal("View() with fill on and off must differ")
	}
}

func TestLineChart_PlainRenderKeepsDimensions(t *testing.T) {
	lc := pulse.New(20, 8, pulse.WithRange(0, 100), pulse.WithGrid(false), pulse.WithFill(false))
	seedCPU(lc)
	lines := strings.Split(strings.TrimRight(lc.View(), "\n"), "\n")
	if len(lines) != 8 {
		t.Fatalf("expected 8 lines, got %d", len(lines))
	}
}

func TestLineChart_GridLinesExactAndEven(t *testing.T) {
	lc := pulse.New(20, 8, pulse.WithRange(0, 100))
	rows := strings.Split(lc.View(), "\n")
	plot := func(r int) string {
		runes := []rune(rows[r])
		i := -1
		for j, ru := range runes {
			if ru == '│' {
				i = j
				break
			}
		}
		if i < 0 {
			t.Fatalf("row %d has no axis", r)
		}
		return string(runes[i+1:])
	}
	// Solid ruling: every row in the plot has no spaces.
	for r := range rows {
		for _, ru := range plot(r) {
			if ru == ' ' {
				t.Fatalf("grid row %d must be solid, got %q", r, plot(r))
			}
		}
	}
	// Solid grid made of box characters.
	for _, want := range []string{"─", "│", "┼"} {
		if !strings.Contains(lc.View(), want) {
			t.Fatalf("View() with grid must contain %q", want)
		}
	}
	plain := pulse.New(20, 8, pulse.WithRange(0, 100), pulse.WithGrid(false), pulse.WithFill(false))
	seedCPU(plain)
	for _, ln := range strings.Split(plain.View(), "\n") {
		if i := strings.Index(ln, "│"); i >= 0 {
			ln = ln[i+len("│"):]
		}
		for _, bad := range []string{"─", "│", "┼"} {
			if strings.Contains(ln, bad) {
				t.Fatalf("plot without grid must not contain %q", bad)
			}
		}
	}
}

func TestLineChart_DenseGridIsEven(t *testing.T) {
	lc := pulse.New(60, 12, pulse.WithRange(0, 100))
	rows := strings.Split(lc.View(), "\n")
	if len(rows) != 12 {
		t.Fatalf("expected 12 rows, got %d", len(rows))
	}
	labelRows := []int{}
	for r, ln := range rows {
		runes := []rune(ln)
		if strings.TrimSpace(string(runes[:4])) != "" {
			labelRows = append(labelRows, r)
		}
	}
	// Gaps between adjacent labels differ by at most 1 row.
	for i := 1; i < len(labelRows); i++ {
		gap := labelRows[i] - labelRows[i-1]
		if gap < 2 || gap > 3 {
			t.Fatalf("label rows %v have uneven gap %d", labelRows, gap)
		}
	}
}

func TestLineChart_LineWidthToggleChangesRender(t *testing.T) {
	a := pulse.New(20, 8, pulse.WithRange(0, 100))
	seedCPU(a)
	b := pulse.New(20, 8, pulse.WithRange(0, 100), pulse.WithLineWidth(1))
	seedCPU(b)
	if a.View() == b.View() {
		t.Fatal("View() with default and thin line must differ")
	}
}
