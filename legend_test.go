package pulse_test

import (
	"strings"
	"testing"

	"github.com/ingvarch/pulse"
)

func TestLineChart_LegendEmptyWithoutNamedSeries(t *testing.T) {
	lc := pulse.New(10, 5, pulse.WithRange(0, 100))
	lc.Push(50)
	if got := lc.Legend(); strings.TrimSpace(got) != "" {
		t.Fatalf("Legend() must be empty without named series, got %q", got)
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
