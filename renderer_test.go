package pulse_test

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/ingvarch/pulse"
)

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

func TestLineChart_TintedFillToggleChangesRender(t *testing.T) {
	cyan := lipgloss.NewStyle().Foreground(lipgloss.Color("#7dcfff"))
	a := pulse.New(20, 8, pulse.WithRange(0, 100), pulse.WithLineStyle(cyan), pulse.WithFill(true))
	seedCPU(a)
	b := pulse.New(20, 8, pulse.WithRange(0, 100), pulse.WithLineStyle(cyan), pulse.WithFill(true), pulse.WithTintedFill(true))
	seedCPU(b)

	if !b.TintedFill() {
		t.Fatal("expected TintedFill() to be true")
	}
	if a.View() == b.View() {
		t.Fatal("View() with tinted fill on and off must differ")
	}

	b.SetTintedFill(false)
	if b.TintedFill() {
		t.Fatal("expected TintedFill() to be false after SetTintedFill(false)")
	}
	if a.View() != b.View() {
		t.Fatal("expected View() to match after disabling tinted fill")
	}
}

func TestLineChart_SolidFillToggleChangesRender(t *testing.T) {
	cyan := lipgloss.NewStyle().Foreground(lipgloss.Color("#7dcfff"))
	a := pulse.New(20, 8, pulse.WithRange(0, 100), pulse.WithLineStyle(cyan), pulse.WithFill(true), pulse.WithTintedFill(true))
	seedCPU(a)
	b := pulse.New(20, 8, pulse.WithRange(0, 100), pulse.WithLineStyle(cyan), pulse.WithFill(true), pulse.WithTintedFill(true), pulse.WithSolidFill(true))
	seedCPU(b)

	if !b.SolidFill() {
		t.Fatal("expected SolidFill() to be true")
	}
	if a.View() == b.View() {
		t.Fatal("View() with solid fill and textured fill must differ")
	}

	b.SetSolidFill(false)
	if b.SolidFill() {
		t.Fatal("expected SolidFill() to be false after SetSolidFill(false)")
	}
	if a.View() != b.View() {
		t.Fatal("expected View() to match after disabling solid fill")
	}
}

func TestLineChart_CustomTintColorChangesRender(t *testing.T) {
	cyan := lipgloss.NewStyle().Foreground(lipgloss.Color("#7dcfff"))
	a := pulse.New(20, 8, pulse.WithRange(0, 100), pulse.WithLineStyle(cyan), pulse.WithFill(true), pulse.WithTintedFill(true))
	seedCPU(a)

	customTint := lipgloss.Color("#251c3d")
	b := pulse.New(20, 8, pulse.WithRange(0, 100), pulse.WithLineStyle(cyan), pulse.WithFill(true), pulse.WithTintedFill(true), pulse.WithTintColor(customTint))
	seedCPU(b)

	if b.TintColor() != customTint {
		t.Fatalf("expected TintColor() to be %v, got %v", customTint, b.TintColor())
	}
	if a.View() == b.View() {
		t.Fatal("View() with custom tint color and default tint color must differ")
	}

	b.SetTintColor(nil)
	if b.TintColor() != nil {
		t.Fatal("expected TintColor() to be nil after SetTintColor(nil)")
	}
	if a.View() != b.View() {
		t.Fatal("expected View() to match after resetting tint color to nil (default)")
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
		return string([]byte{byte('0' + int(v)/1000)}) + " GHz"
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
		sep := strings.Index(ln, "│")
		if sep < 0 {
			continue
		}
		lbl := ln[:sep]
		if len([]rune(lbl)) != 8 {
			t.Fatalf("line %d label width = %d, want 8: %q", i, len([]rune(lbl)), lbl)
		}
	}
	if !strings.Contains(view, "4 GHz") || !strings.Contains(view, "2 GHz") || !strings.Contains(view, "0 GHz") {
		t.Fatalf("view must contain formatted labels: 4 GHz, 2 GHz, 0 GHz:\n%s", view)
	}
}
