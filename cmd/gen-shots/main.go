package main

import (
	"flag"
	"fmt"
	"math"
	"math/rand"

	"charm.land/lipgloss/v2"
	"github.com/ingvarch/pulse"
	"github.com/ingvarch/pulse/theme"
)

func main() {
	mode := flag.String("mode", "hero", "hero | quickstart | styled | braille")
	flag.Parse()

	switch *mode {
	case "hero":
		renderHero()
	case "quickstart":
		renderQuickstart()
	case "styled":
		renderStyled()
	case "braille":
		renderBraille()
	}
}

func renderHero() {
	preset := theme.TokyoNight()
	chartW, chartH := 80, 13

	lc := pulse.New(chartW, chartH,
		pulse.WithRange(0, 100),
		pulse.WithTicks(0, 25, 50, 75, 100),
		pulse.WithLineWidth(1),
		pulse.WithAxisStyle(preset.Axis),
		pulse.WithSeriesStyle("cpu", lipgloss.NewStyle().Foreground(lipgloss.Color(theme.SeriesColor(1)))), // blue #7aa2f7
		pulse.WithSeriesStyle("mem", lipgloss.NewStyle().Foreground(lipgloss.Color(theme.SeriesColor(0)))), // green #9ece6a
		pulse.WithSeriesStyle("net", lipgloss.NewStyle().Foreground(lipgloss.Color(theme.SeriesColor(4)))), // yellow #e0af68
	)

	// Deterministic seed for reproducible screenshots
	rng := rand.New(rand.NewSource(42))

	for i := 1; i <= chartW; i++ {
		// CPU: rolling hills
		cpu := 55 + 24*math.Sin(float64(i)*0.14) + 10*math.Cos(float64(i)*0.07) + (rng.Float64()-0.5)*3
		// MEM: steady baseline
		mem := 44 + 14*math.Sin(float64(i)*0.05) + 4*math.Cos(float64(i)*0.1)
		// NET: base with burst
		net := 18 + 7*math.Sin(float64(i)*0.09)
		if i > 52 && i < 68 {
			net += 45 + rng.Float64()*15
		}

		lc.PushSeries("cpu", math.Max(0, math.Min(100, cpu)))
		lc.PushSeries("mem", math.Max(0, math.Min(100, mem)))
		lc.PushSeries("net", math.Max(0, math.Min(100, net)))
	}

	fmt.Println(lc.View())
	fmt.Println(lc.LegendBox())
}

func renderQuickstart() {
	preset := theme.TokyoNight()
	chartW, chartH := 80, 13

	lc := pulse.New(chartW, chartH,
		pulse.WithRange(0, 100),
		pulse.WithTicks(0, 25, 50, 75, 100),
		pulse.WithLineWidth(1),
		pulse.WithAxisStyle(preset.Axis),
		pulse.WithLineStyle(preset.LineFor(60)),
	)

	for i := 0; i < chartW; i++ {
		v := 50 + 38*math.Sin(float64(i)*0.16)
		lc.Push(v)
	}

	fmt.Println(lc.View())
}

func renderStyled() {
	preset := theme.TokyoNight()
	chartW, chartH := 80, 13

	lc := pulse.New(chartW, chartH,
		pulse.WithRange(0, 100),
		pulse.WithTicks(0, 25, 50, 75, 100),
		pulse.WithLineWidth(2), // Bold curves
		pulse.WithTintedFill(true),
		pulse.WithAxisStyle(preset.Axis),
		pulse.WithSeriesStyle("cluster-a", preset.LineFor(88.0)), // Red threshold warning
		pulse.WithSeriesStyle("cluster-b", lipgloss.NewStyle().Foreground(lipgloss.Color("#bb9af7"))), // Purple
	)

	rng := rand.New(rand.NewSource(99))
	for i := 1; i <= chartW; i++ {
		ca := 72 + 18*math.Sin(float64(i)*0.12) + (rng.Float64()-0.5)*4
		cb := 35 + 20*math.Cos(float64(i)*0.15) + (rng.Float64()-0.5)*3
		lc.PushSeries("cluster-a", math.Max(0, math.Min(100, ca)))
		lc.PushSeries("cluster-b", math.Max(0, math.Min(100, cb)))
	}

	// Container box with Lip Gloss border
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#c0caf5")).
		Render("  INFRASTRUCTURE TELEMETRY (LIVE)")

	chartStr := lc.View()
	legendStr := lc.LegendBox()

	content := fmt.Sprintf("%s\n\n%s\n%s", header, chartStr, legendStr)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3b4261")).
		Padding(1, 2).
		Render(content)

	fmt.Println(box)
}

func renderBraille() {
	preset := theme.TokyoNight()
	chartW, chartH := 80, 13

	lc := pulse.New(chartW, chartH,
		pulse.WithRange(0, 100),
		pulse.WithTicks(0, 25, 50, 75, 100),
		pulse.WithMode(pulse.ModeBraille),
		pulse.WithAxisStyle(preset.Axis),
		pulse.WithSeriesStyle("frequency", lipgloss.NewStyle().Foreground(lipgloss.Color("#7dcfff"))),
	)

	for i := 0; i < chartW; i++ {
		v := 50 + 35*math.Sin(float64(i)*0.18) + 10*math.Cos(float64(i)*0.35)
		lc.PushSeries("frequency", math.Max(0, math.Min(100, v)))
	}

	fmt.Println(lc.View())
	fmt.Println(lc.LegendBox())
}
