package main

import (
	"flag"
	"fmt"
	"math"
	"math/rand"

	"charm.land/lipgloss/v2"
	"github.com/ingvarch/pulse"
	"github.com/ingvarch/pulse/scale"
	"github.com/ingvarch/pulse/theme"
)

func main() {
	mode := flag.String("mode", "hero", "hero | quickstart | styled | braille | rxtx | events | fill | zerocross")
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
	case "rxtx":
		renderRXTX()
	case "events":
		renderEvents()
	case "fill":
		renderFill()
	case "zerocross":
		renderZeroCross()
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

func renderRXTX() {
	preset := theme.TokyoNight()
	chartW, chartH := 80, 15
	const MB = 1024 * 1024

	lc := pulse.New(chartW, chartH,
		pulse.WithLineWidth(1),
		pulse.WithRange(0, 100*MB),
		pulse.WithSymmetric(true),
		pulse.WithZeroBaseline(true),
		pulse.WithTicks(-100*MB, -50*MB, 0, 50*MB, 100*MB),
		pulse.WithLabelFormatter(scale.BytesRateFormatter(true)),
		pulse.WithSeriesStyle("rx", lipgloss.NewStyle().Foreground(lipgloss.Color("#9ece6a"))),
		pulse.WithSeriesStyle("tx", lipgloss.NewStyle().Foreground(lipgloss.Color("#bb9af7"))),
		pulse.WithSeriesInverted("tx", true),
		pulse.WithAxisStyle(preset.Axis),
		pulse.WithFill(true),
		pulse.WithTintedFill(true),
		pulse.WithGrid(true),
	)

	rng := rand.New(rand.NewSource(123))
	for i := 1; i <= chartW; i++ {
		rx := (48 + 26*math.Sin(float64(i)*0.13) + 8*math.Cos(float64(i)*0.07) + (rng.Float64()-0.5)*4) * MB
		tx := (32 + 20*math.Cos(float64(i)*0.11) + 6*math.Sin(float64(i)*0.09) + (rng.Float64()-0.5)*3) * MB
		lc.PushSeries("rx", rx)
		lc.PushSeries("tx", tx)
	}

	rxVal, _ := lc.Last("rx")
	txVal, _ := lc.Last("tx")

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#c0caf5")).
		Render(fmt.Sprintf("NETWORK THROUGHPUT  ▲ RX %s  |  ▼ TX %s", scale.FormatBytesRate(rxVal), scale.FormatBytesRate(txVal)))

	content := fmt.Sprintf("%s\n\n%s\n%s", header, lc.View(), lc.LegendBox())

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3b4261")).
		Padding(1, 2).
		Render(content)

	fmt.Println(box)
}

func renderEvents() {
	preset := theme.TokyoNight()
	chartW, chartH := 80, 13

	lc := pulse.New(chartW, chartH,
		pulse.WithRange(0, 100),
		pulse.WithTicks(0, 25, 50, 75, 100),
		pulse.WithFill(true),
		pulse.WithTintedFill(true),
		pulse.WithGrid(true),
		pulse.WithAxisStyle(preset.Axis),
		pulse.WithLineStyle(preset.LineFor(60)),
	)

	for i := 0; i < chartW; i++ {
		val := 42.0 + 32.0*math.Sin(float64(i)*0.14) + 6.0*math.Cos(float64(i)*0.06)
		lc.Push(val)
	}

	lc.AddEventAt(52, pulse.Event{
		Label: "Deploy v1.4.2",
		Glyph: "🚀",
		Style: lipgloss.NewStyle().Foreground(lipgloss.Color("#7dcfff")).Bold(true),
	})

	lc.AddEventAt(26, pulse.Event{
		Label: "Traffic Surge",
		Glyph: "⚡",
		Style: lipgloss.NewStyle().Foreground(lipgloss.Color("#e0af68")).Bold(true),
	})

	lc.AddEventAt(10, pulse.Event{
		Label: "Cache Eviction",
		Glyph: "⚠️",
		Style: lipgloss.NewStyle().Foreground(lipgloss.Color("#f7768e")).Bold(true),
	})

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#c0caf5")).
		Render("PRODUCTION SERVICE LATENCY & TIMELINE EVENTS")

	content := fmt.Sprintf("%s\n\n%s\n%s", header, lc.View(), lc.EventsBox())

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3b4261")).
		Padding(1, 2).
		Render(content)

	fmt.Println(box)
}

func renderFill() {
	preset := theme.TokyoNight()
	chartW, chartH := 80, 13

	lc := pulse.New(chartW, chartH,
		pulse.WithRange(0, 100),
		pulse.WithTicks(0, 25, 50, 75, 100),
		pulse.WithLineWidth(1),
		pulse.WithFill(true),
		pulse.WithTintedFill(true),
		pulse.WithGrid(true),
		pulse.WithAxisStyle(preset.Axis),
		pulse.WithLineStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#7aa2f7"))),
	)

	for i := 0; i < chartW; i++ {
		val := 45 + 35*math.Sin(float64(i)*0.12) + 10*math.Cos(float64(i)*0.05)
		lc.Push(val)
	}

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#c0caf5")).
		Render("SHADED & TINTED AREA FILL (GRID ABSORPTION)")

	content := fmt.Sprintf("%s\n\n%s", header, lc.View())

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3b4261")).
		Padding(1, 2).
		Render(content)

	fmt.Println(box)
}

func renderZeroCross() {
	preset := theme.TokyoNight()
	chartW, chartH := 80, 13

	green := lipgloss.NewStyle().Foreground(lipgloss.Color("#9ece6a"))
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#f7768e"))

	lc := pulse.New(chartW, chartH,
		pulse.WithLineWidth(1),
		pulse.WithRange(-40, 40),
		pulse.WithSymmetric(true),
		pulse.WithZeroBaseline(true),
		pulse.WithTicks(-40, -20, 0, 20, 40),
		pulse.WithLineStyle(green),
		pulse.WithNegativeStyle(red),
		pulse.WithFill(true),
		pulse.WithTintedFill(true),
		pulse.WithGrid(true),
		pulse.WithAxisStyle(preset.Axis),
		pulse.WithLabelFormatter(func(v float64) string {
			if v > 0 {
				return fmt.Sprintf("+%.0f ms", v)
			}
			return fmt.Sprintf("%.0f ms", v)
		}),
	)

	for i := 0; i < chartW; i++ {
		delta := 28.0*math.Sin(float64(i)*0.16) + 6.0*math.Cos(float64(i)*0.08)
		lc.Push(delta)
	}

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#c0caf5")).
		Render("BIDIRECTIONAL ZERO-CROSSING (LATENCY DRIFT)")

	content := fmt.Sprintf("%s\n\n%s", header, lc.View())

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3b4261")).
		Padding(1, 2).
		Render(content)

	fmt.Println(box)
}
