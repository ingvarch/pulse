package theme_test

import (
	"testing"

	"github.com/ingvarch/pulse/theme"
)

func TestValueColor_Thresholds(t *testing.T) {
	cases := []struct {
		value float64
		want  string
	}{
		{10, theme.ColorOK},
		{59.9, theme.ColorOK},
		{60, theme.ColorWarn},
		{70, theme.ColorWarn},
		{84.9, theme.ColorWarn},
		{85, theme.ColorCrit},
		{99, theme.ColorCrit},
	}
	for _, c := range cases {
		if got := theme.ValueColor(c.value); got != c.want {
			t.Errorf("ValueColor(%v) = %q, want %q", c.value, got, c.want)
		}
	}
}

func TestPreset_HasLineAndAxis(t *testing.T) {
	p := theme.TokyoNight()
	if p.LineFor(10).Render("x") == "" {
		t.Fatal("LineFor renders empty")
	}
	if p.Axis.Render("x") == "" {
		t.Fatal("Axis renders empty")
	}
}

func TestSeriesColor_Distinct(t *testing.T) {
	a, b, c := theme.SeriesColor(0), theme.SeriesColor(1), theme.SeriesColor(2)
	if a == "" || b == "" || c == "" {
		t.Fatalf("empty series colors: %q %q %q", a, b, c)
	}
	if a == b || b == c || a == c {
		t.Fatalf("series colors must differ: %q %q %q", a, b, c)
	}
}
