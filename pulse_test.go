package pulse_test

import (
	"fmt"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/ingvarch/pulse"
)

func seedCPU(lc *pulse.Model) {
	for _, v := range []float64{10, 30, 50, 70, 90, 70, 50, 30} {
		lc.Push(v)
	}
}

func TestLineChart_RendersCPUValues(t *testing.T) {
	lc := pulse.New(20, 8, pulse.WithRange(0, 100))
	seedCPU(lc)
	view := lc.View()
	if strings.TrimSpace(view) == "" {
		t.Fatal("View() returned empty string")
	}
	lines := strings.Split(strings.TrimRight(view, "\n"), "\n")
	if len(lines) != 8 {
		t.Fatalf("expected 8 lines, got %d", len(lines))
	}
	// Check Y-axis border exists on every line
	for i, ln := range lines {
		if !strings.Contains(ln, "│") {
			t.Fatalf("line %d missing Y-axis border: %q", i, ln)
		}
	}
}

func TestLineChart_EmptyRendersAxes(t *testing.T) {
	lc := pulse.New(20, 8, pulse.WithRange(0, 100))
	view := lc.View()
	lines := strings.Split(strings.TrimRight(view, "\n"), "\n")
	if len(lines) != 8 {
		t.Fatalf("expected 8 lines, got %d", len(lines))
	}
}

func TestLineChart_NoDuplicateAdjacentLabels(t *testing.T) {
	lc := pulse.New(20, 6, pulse.WithRange(0, 100))
	lines := strings.Split(strings.TrimRight(lc.View(), "\n"), "\n")
	prev := ""
	for i, ln := range lines {
		sep := strings.Index(ln, "│")
		if sep < 0 {
			continue
		}
		lbl := strings.TrimSpace(ln[:sep])
		if lbl == "" {
			prev = ""
			continue
		}
		if lbl == prev {
			t.Fatalf("duplicate adjacent label %q at line %d", lbl, i)
		}
		prev = lbl
	}
}

func TestLineChart_SlidingWindow(t *testing.T) {
	lc := pulse.New(5, 5, pulse.WithRange(0, 100))
	for i := 0; i < 10; i++ {
		lc.Push(float64(i * 10))
	}
	if lc.Len() != 5 {
		t.Fatalf("expected sliding window len 5, got %d", lc.Len())
	}
}

func TestLineChart_SetLineStyle(t *testing.T) {
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	lc := pulse.New(10, 5, pulse.WithRange(0, 100), pulse.WithLineStyle(red))
	lc.Push(50)
	if lc.LineStyle().GetForeground() != red.GetForeground() {
		t.Fatal("LineStyle() did not retain style set via option")
	}
}

func TestLineChart_MultiSeries(t *testing.T) {
	lc := pulse.New(20, 8, pulse.WithRange(0, 100))
	lc.PushSeries("cpu", 50)
	lc.PushSeries("mem", 75)
	if lc.LenSeries("cpu") != 1 || lc.LenSeries("mem") != 1 {
		t.Fatal("named series not tracked correctly")
	}
	names := lc.SeriesNames()
	if len(names) != 2 || names[0] != "cpu" || names[1] != "mem" {
		t.Fatalf("SeriesNames() = %v, want [cpu mem]", names)
	}
	legend := lc.Legend()
	if !strings.Contains(legend, "cpu") || !strings.Contains(legend, "mem") {
		t.Fatalf("Legend() must list cpu and mem, got %q", legend)
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

func TestLineChart_Stringer(t *testing.T) {
	lc := pulse.New(20, 8, pulse.WithRange(0, 100))
	seedCPU(lc)
	str := fmt.Sprint(lc)
	if str != lc.View() {
		t.Fatalf("fmt.Sprint(lc) must match lc.View()")
	}
}

type staticRenderer struct{}

func (staticRenderer) Render(c pulse.Chart, ctx pulse.RenderContext) string {
	return "STATIC_CHART"
}

func TestLineChart_CustomRenderer(t *testing.T) {
	lc := pulse.New(20, 8, pulse.WithRenderer(staticRenderer{}))
	if got := lc.View(); got != "STATIC_CHART" {
		t.Fatalf("expected custom renderer output, got %q", got)
	}
}

func TestLineChart_UnicodeLabelAlignment(t *testing.T) {
	tests := []struct {
		name       string
		formatter  func(float64) string
		labelWidth int
		maxVal     float64
	}{
		{
			name:       "Celsius",
			formatter:  func(v float64) string { return fmt.Sprintf("%.0f°C", v) },
			labelWidth: 5,
			maxVal:     100,
		},
		{
			name:       "CJK",
			formatter:  func(v float64) string { return fmt.Sprintf("%.0f度", v) },
			labelWidth: 5,
			maxVal:     100,
		},
		{
			name:       "Emoji",
			formatter:  func(v float64) string { return fmt.Sprintf("%.0f🌡️", v) },
			labelWidth: 6,
			maxVal:     100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lc := pulse.New(20, 8,
				pulse.WithRange(0, tt.maxVal),
				pulse.WithLabelFormatter(tt.formatter),
				pulse.WithLabelWidth(tt.labelWidth),
			)
			view := lc.View()
			for _, line := range strings.Split(view, "\n") {
				if len(line) == 0 {
					continue
				}
				sepIdx := strings.Index(line, "│")
				if sepIdx < 0 {
					t.Fatalf("missing axis border in line: %q", line)
				}
				prefix := line[:sepIdx]
				w := lipgloss.Width(prefix)
				if w != tt.labelWidth {
					t.Errorf("visual width mismatch for prefix %q: got %d, want %d", prefix, w, tt.labelWidth)
				}
			}
		})
	}
}

func TestLineChart_CustomTicks_SortAndDeduplicate(t *testing.T) {
	lc := pulse.New(20, 8,
		pulse.WithRange(0, 100),
		pulse.WithTicks(100, 50, 0, 50, 25, 0, 75, 100),
	)
	view := lc.View()
	if strings.TrimSpace(view) == "" {
		t.Fatal("View() returned empty string with custom ticks")
	}
	lines := strings.Split(strings.TrimRight(view, "\n"), "\n")
	seen := make(map[string]int)
	for _, ln := range lines {
		if sep := strings.Index(ln, "│"); sep >= 0 {
			lbl := strings.TrimSpace(ln[:sep])
			if lbl != "" {
				seen[lbl]++
				if seen[lbl] > 1 {
					t.Errorf("tick label %q appeared %d times", lbl, seen[lbl])
				}
			}
		}
	}
}

func TestRenderer_CustomOverridesViewAndMode(t *testing.T) {
	m := pulse.New(20, 8, pulse.WithRenderer(staticRenderer{}))
	if m.Mode() != pulse.ModeCustom {
		t.Fatalf("expected ModeCustom for staticRenderer, got %v", m.Mode())
	}
	if got := m.View(); got != "STATIC_CHART" {
		t.Fatalf("expected STATIC_CHART, got %q", got)
	}
}

func TestLineChart_InvertedRangeNormalization(t *testing.T) {
	m := pulse.New(20, 8, pulse.WithRange(100, 0))
	if m.Min() != 0 || m.Max() != 100 {
		t.Fatalf("expected WithRange(100, 0) to normalize to (0, 100), got (%v, %v)", m.Min(), m.Max())
	}
	m.SetRange(80, 20)
	if m.Min() != 20 || m.Max() != 80 {
		t.Fatalf("expected SetRange(80, 20) to normalize to (20, 80), got (%v, %v)", m.Min(), m.Max())
	}
}
