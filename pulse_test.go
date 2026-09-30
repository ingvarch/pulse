package pulse_test

import (
	"fmt"
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

func TestLineChart_NamedSeriesLeftEndpointOwner(t *testing.T) {
	blue := lipgloss.NewStyle().Foreground(lipgloss.Color("#0000ff"))
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))

	// Case 1: Multi-point line. Start point at startCol must be red, not blue.
	lc := pulse.New(20, 8,
		pulse.WithLineWidth(1),
		pulse.WithRange(0, 100),
		pulse.WithLineStyle(blue),
		pulse.WithSeriesStyle("alpha", red),
		pulse.WithGrid(false),
		pulse.WithFill(false),
	)
	for _, v := range []float64{50, 60, 70} {
		lc.PushSeries("alpha", v)
	}
	view := lc.View()

	// Default line style (blue) must NEVER appear anywhere in a chart with only named series "alpha"
	if strings.Contains(view, blue.Render("─")) || strings.Contains(view, blue.Render("╭")) || strings.Contains(view, blue.Render("╰")) {
		t.Fatalf("default blue style leaked into named series alpha:\n%s", view)
	}

	lines := strings.Split(strings.TrimRight(view, "\n"), "\n")
	// For 50% on height 8: row index = round(7 * 0.5) = 4 (counting from top).
	// With 3 points in window of 20: startCol = 20 - 3 = 17.
	sepIdx := strings.Index(lines[4], "│")
	if sepIdx < 0 {
		t.Fatalf("missing axis border in line 4: %q", lines[4])
	}
	plotArea := lines[4][sepIdx+len("│"):]
	if len(plotArea) < 17 {
		t.Fatalf("plotArea too short: %q", plotArea)
	}
	// Columns 0..16 must be blank spaces
	if plotArea[:17] != strings.Repeat(" ", 17) {
		t.Fatalf("expected 17 leading spaces before startCol, got %q", plotArea[:17])
	}
	// Column 17 (the starting cell) must strictly begin with red.Render("─")
	if !strings.HasPrefix(plotArea[17:], red.Render("─")) {
		t.Fatalf("starting cell at col 17 must be red.Render(\"─\"), got:\n%q", plotArea[17:])
	}

	// Case 2: Single-point line (n == 1). Must also be red, not blue.
	single := pulse.New(10, 5,
		pulse.WithLineWidth(1),
		pulse.WithRange(0, 100),
		pulse.WithLineStyle(blue),
		pulse.WithSeriesStyle("alpha", red),
		pulse.WithGrid(false),
		pulse.WithFill(false),
	)
	single.PushSeries("alpha", 50)
	singleView := single.View()
	if strings.Contains(singleView, blue.Render("─")) {
		t.Fatalf("single point left endpoint must not use default blue style")
	}
	if !strings.Contains(singleView, red.Render("─")) {
		t.Fatalf("single point left endpoint must use series red style")
	}
}

func TestLineChart_LabelFormatterAndWidth(t *testing.T) {
	fmtFn := func(v float64) string {
		return strings.Repeat(" ", 0) + string([]byte{byte('0' + int(v)/1000)}) + " GHz"
	}
	lc := pulse.New(15, 5,
		pulse.WithRange(0, 4000),
		pulse.WithTicks(0, 2000, 4000),
		pulse.WithLabelFormatter(fmtFn),
		pulse.WithLabelWidth(8),
	)
	view := lc.View()
	lines := strings.Split(strings.TrimRight(view, "\n"), "\n")
	for i, ln := range lines {
		runes := []rune(ln)
		if len(runes) < 9 {
			t.Fatalf("line %d too short: %q", i, ln)
		}
		if runes[8] != '│' {
			t.Fatalf("expected '│' at index 8 for line %d, got %q (line: %q)", i, string(runes[8]), ln)
		}
	}
	if !strings.Contains(view, "4 GHz") || !strings.Contains(view, "0 GHz") {
		t.Fatalf("expected custom formatted labels in View(), got:\n%s", view)
	}
}

func TestLineChart_Stringer(t *testing.T) {
	lc := pulse.New(20, 8, pulse.WithRange(0, 100))
	lc.Push(50)
	var stringer fmt.Stringer = lc
	if stringer.String() != lc.View() {
		t.Fatalf("expected String() to match View()")
	}
}

type testCustomRenderer struct{}

func (testCustomRenderer) Render(m pulse.Chart, ctx pulse.RenderContext) string {
	return fmt.Sprintf("custom[%dx%d]:%d", m.Width(), m.Height(), len(ctx.Names))
}

func TestLineChart_CustomRenderer(t *testing.T) {
	lc := pulse.New(30, 10, pulse.WithRenderer(testCustomRenderer{}))
	lc.PushSeries("s1", 42)
	got := lc.View()
	want := "custom[30x10]:2" // "" default series + "s1"
	if got != want {
		t.Fatalf("expected custom renderer output %q, got %q", want, got)
	}
}

func TestLineChart_UnicodeLabelAlignment(t *testing.T) {
	tests := []struct {
		name      string
		formatter func(float64) string
	}{
		{
			name: "Celsius",
			formatter: func(v float64) string {
				return fmt.Sprintf("%d°C", int(v))
			},
		},
		{
			name: "CJK",
			formatter: func(v float64) string {
				return fmt.Sprintf("%d度", int(v))
			},
		},
		{
			name: "Emoji",
			formatter: func(v float64) string {
				return fmt.Sprintf("%d🔥", int(v))
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lc := pulse.New(20, 6,
				pulse.WithRange(0, 100),
				pulse.WithTicks(0, 50, 100),
				pulse.WithLabelFormatter(tc.formatter),
			)
			view := lc.View()
			lines := strings.Split(strings.TrimRight(view, "\n"), "\n")
			var axisCol int
			for r, ln := range lines {
				idx := strings.Index(ln, "│")
				if idx < 0 {
					t.Fatalf("line %d missing axis '│': %q", r, ln)
				}
				// Visual width of the label column before '│' must be constant across all lines
				visualW := lipgloss.Width(ln[:idx])
				if r == 0 {
					axisCol = visualW
				} else if visualW != axisCol {
					t.Fatalf("line %d axis visual column mismatch: got %d, want %d (line: %q)", r, visualW, axisCol, ln)
				}
			}
		})
	}
}

func TestLineChart_CustomTicks_SortAndDeduplicate(t *testing.T) {
	// User provides unsorted ticks with duplicates: 100, 0, 50, 50
	lc := pulse.New(20, 6,
		pulse.WithRange(0, 100),
		pulse.WithTicks(100, 0, 50, 50),
	)
	view := lc.View()
	lines := strings.Split(strings.TrimRight(view, "\n"), "\n")
	// Must have labels at top (100) and bottom (0)
	if !strings.Contains(lines[0], "100") {
		t.Fatalf("expected top line to contain 100, got: %q", lines[0])
	}
	if !strings.Contains(lines[len(lines)-1], "0") {
		t.Fatalf("expected bottom line to contain 0, got: %q", lines[len(lines)-1])
	}
}

type stubRenderer struct{}

func (stubRenderer) Render(_ pulse.Chart, _ pulse.RenderContext) string {
	return "CUSTOM"
}

func TestRenderer_CustomOverridesViewAndMode(t *testing.T) {
	lc := pulse.New(20, 8, pulse.WithRange(0, 100))
	lc.Push(50)
	lc.SetRenderer(stubRenderer{})
	if got := lc.Mode(); got != pulse.ModeCustom {
		t.Fatalf("Mode() = %v, want ModeCustom", got)
	}
	if got := lc.View(); got != "CUSTOM" {
		t.Fatalf("View() = %q, want custom renderer output", got)
	}
}
