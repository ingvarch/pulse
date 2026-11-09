package main

import (
	"flag"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/ingvarch/pulse"
	"github.com/ingvarch/pulse/scale"
	"github.com/ingvarch/pulse/theme"
)

func main() {
	mode := flag.String("mode", "hero", "hero | quickstart | styled | braille | rxtx | events | fill | zerocross | xaxis | xaxisat | autorange | thresholds | logscale | rightaxis | cursor | sparkline")
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
	case "xaxis":
		renderXAxis()
	case "xaxisat":
		renderXAxisAt()
	case "autorange":
		renderAutoRange()
	case "thresholds":
		renderThresholds()
	case "logscale":
		renderLogScale()
	case "rightaxis":
		renderRightAxis()
	case "cursor":
		renderCursor()
	case "sparkline":
		renderSparkline()
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

func renderXAxis() {
	newest := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	step := 7 * 24 * time.Hour / xAxisShotW
	renderXAxisShot("CPU, LAST 7 DAYS (GAP = NO DATA)", pulse.WithXAxis(func(age int) string {
		return newest.Add(-time.Duration(age) * step).Format("Jan 2")
	}))
}

func renderXAxisAt() {
	newest := time.Date(2026, 10, 15, 14, 47, 0, 0, time.UTC)
	step := 24 * time.Hour / xAxisShotW
	renderXAxisShot("CPU, LAST 24 HOURS (GAP = NO DATA)", pulse.WithXAxisAt(func(age int) string {
		if age == 0 {
			return "now"
		}
		t := newest.Add(-time.Duration(age) * step)
		mark := time.Date(t.Year(), t.Month(), t.Day(), t.Hour()/6*6, 0, 0, 0, t.Location())
		if t.Sub(mark) < step {
			return mark.Format("15:04")
		}
		return ""
	}))
}

func renderAutoRange() {
	preset := theme.TokyoNight()

	lc := pulse.New(80, 13,
		pulse.WithAutoRange(true),
		pulse.WithLineWidth(1),
		pulse.WithFill(true),
		pulse.WithTintedFill(true),
		pulse.WithGrid(true),
		pulse.WithAxisStyle(preset.Axis),
		pulse.WithLabelFormatter(func(v float64) string { return fmt.Sprintf("%.0f MB/s", v) }),
		pulse.WithLineStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#7aa2f7"))),
	)

	rng := rand.New(rand.NewSource(11))
	for i := 0; i < 80; i++ {
		v := 210 + 60*math.Sin(float64(i)*0.15) + (rng.Float64()-0.5)*40
		if i >= 48 && i < 56 {
			v += 420 * math.Sin(float64(i-48)/8*math.Pi)
		}
		lc.Push(math.Max(0, v))
	}

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#c0caf5")).
		Render("NETWORK THROUGHPUT (AUTO RANGE)")

	content := fmt.Sprintf("%s\n\n%s", header, lc.View())

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3b4261")).
		Padding(1, 2).
		Render(content)

	fmt.Println(box)
}

func renderThresholds() {
	preset := theme.TokyoNight()
	warn := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.ColorWarn))
	crit := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.ColorCrit))

	lc := pulse.New(80, 13,
		pulse.WithRange(0, 100),
		pulse.WithTicks(0, 25, 50, 75, 100),
		pulse.WithLineWidth(1),
		pulse.WithFill(true),
		pulse.WithGrid(true),
		pulse.WithAxisStyle(preset.Axis),
		pulse.WithLineStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#7aa2f7"))),
		pulse.WithThresholds(
			pulse.Threshold{Value: 80, Style: warn, Label: "warn 80", Line: true},
			pulse.Threshold{Value: 90, Style: crit, Label: "crit 90", Line: true},
		),
		pulse.WithThresholdColors(true),
	)

	rng := rand.New(rand.NewSource(5))
	for i := 0; i < 80; i++ {
		v := 52 + 30*math.Sin(float64(i)*0.13) + 12*math.Sin(float64(i)*0.41) + (rng.Float64()-0.5)*6
		lc.Push(math.Max(0, math.Min(100, v)))
	}

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#c0caf5")).
		Render("CPU (WARNING AT 80, CRITICAL AT 90)")

	content := fmt.Sprintf("%s\n\n%s", header, lc.View())

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3b4261")).
		Padding(1, 2).
		Render(content)

	fmt.Println(box)
}

const xAxisShotW = 80

func renderXAxisShot(title string, axis pulse.Option) {
	preset := theme.TokyoNight()

	lc := pulse.New(xAxisShotW, 13,
		pulse.WithRange(0, 100),
		pulse.WithTicks(0, 25, 50, 75, 100),
		pulse.WithLineWidth(1),
		pulse.WithFill(true),
		pulse.WithTintedFill(true),
		pulse.WithGrid(true),
		pulse.WithAxisStyle(preset.Axis),
		pulse.WithLineStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#7aa2f7"))),
		axis,
	)

	rng := rand.New(rand.NewSource(7))
	for i := 0; i < xAxisShotW; i++ {
		if i >= 52 && i < 58 {
			lc.Push(math.NaN())
			continue
		}
		v := 38 + 24*math.Sin(float64(i)*0.2) + 8*math.Cos(float64(i)*0.07) + (rng.Float64()-0.5)*6
		lc.Push(math.Max(0, math.Min(100, v)))
	}

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#c0caf5")).
		Render(title)

	content := fmt.Sprintf("%s\n\n%s", header, lc.View())

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3b4261")).
		Padding(1, 2).
		Render(content)

	fmt.Println(box)
}

func renderLogScale() {
	preset := theme.TokyoNight()

	lc := pulse.New(80, 13,
		pulse.WithLogScale(true),
		pulse.WithAutoRange(true),
		pulse.WithLineWidth(1),
		pulse.WithFill(false),
		pulse.WithGrid(true),
		pulse.WithAxisStyle(preset.Axis),
		pulse.WithLabelFormatter(func(v float64) string { return scale.FormatTick(v) + " ms" }),
		pulse.WithSeriesStyle("p50", lipgloss.NewStyle().Foreground(lipgloss.Color(theme.SeriesColor(0)))),
		pulse.WithSeriesStyle("p99", lipgloss.NewStyle().Foreground(lipgloss.Color(theme.SeriesColor(4)))),
	)

	rng := rand.New(rand.NewSource(23))
	for i := 0; i < 80; i++ {
		p50 := 12 + 6*math.Sin(float64(i)*0.2) + (rng.Float64()-0.5)*4
		p99 := p50 * (4 + 2*rng.Float64())
		if i >= 30 && i < 38 {
			p99 += 300 * math.Sin(float64(i-30)/8*math.Pi)
		}
		lc.PushSeries("p50", p50)
		lc.PushSeries("p99", p99)
	}

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#c0caf5")).
		Render("REQUEST LATENCY (LOG SCALE)")

	content := fmt.Sprintf("%s\n\n%s\n%s", header, lc.View(), lc.LegendBox())

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3b4261")).
		Padding(1, 2).
		Render(content)

	fmt.Println(box)
}

func renderRightAxis() {
	preset := theme.TokyoNight()

	lc := pulse.New(80, 13,
		pulse.WithRange(0, 100),
		pulse.WithTicks(0, 25, 50, 75, 100),
		pulse.WithRightAxis(pulse.Axis{
			Min: 0, Max: 5000,
			Ticks:     []float64{0, 1250, 2500, 3750, 5000},
			Formatter: func(v float64) string { return fmt.Sprintf("%.0f/s", v) },
		}),
		pulse.WithLineWidth(1),
		pulse.WithFill(false),
		pulse.WithAxisStyle(preset.Axis),
		pulse.WithSeriesStyle("cpu", lipgloss.NewStyle().Foreground(lipgloss.Color(theme.SeriesColor(1)))),
		pulse.WithSeriesStyle("requests", lipgloss.NewStyle().Foreground(lipgloss.Color(theme.SeriesColor(4)))),
	)

	rng := rand.New(rand.NewSource(31))
	for i := 0; i < 80; i++ {
		rps := 2400 + 1200*math.Sin(float64(i)*0.12) + (rng.Float64()-0.5)*300
		if i >= 50 && i < 62 {
			rps += 1100
		}
		cpu := 14 + rps/5000*60 + (rng.Float64()-0.5)*6
		lc.PushSeries("cpu", math.Max(0, math.Min(100, cpu)))
		lc.PushSeries("requests", math.Max(0, math.Min(5000, rps)))
	}
	lc.SetSeriesRightAxis("requests", true)

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#c0caf5")).
		Render("CPU (LEFT) AND REQUESTS PER SECOND (RIGHT)")

	content := fmt.Sprintf("%s\n\n%s\n%s", header, lc.View(), lc.Legend())

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3b4261")).
		Padding(1, 2).
		Render(content)

	fmt.Println(box)
}

func renderCursor() {
	preset := theme.TokyoNight()
	newest := time.Date(2026, 10, 15, 14, 47, 0, 0, time.UTC)

	lc := pulse.New(80, 13,
		pulse.WithRange(0, 100),
		pulse.WithTicks(0, 25, 50, 75, 100),
		pulse.WithHistory(200),
		pulse.WithLineWidth(1),
		pulse.WithAxisStyle(preset.Axis),
		pulse.WithXAxis(func(age int) string {
			return newest.Add(-time.Duration(age) * time.Minute).Format("15:04")
		}),
		pulse.WithSeriesStyle("cpu", lipgloss.NewStyle().Foreground(lipgloss.Color(theme.SeriesColor(1)))),
		pulse.WithSeriesStyle("mem", lipgloss.NewStyle().Foreground(lipgloss.Color(theme.SeriesColor(0)))),
	)

	rng := rand.New(rand.NewSource(17))
	for i := 0; i < 160; i++ {
		cpu := 52 + 30*math.Sin(float64(i)*0.13) + 10*math.Sin(float64(i)*0.41) + (rng.Float64()-0.5)*6
		mem := 40 + 12*math.Sin(float64(i)*0.05) + (rng.Float64()-0.5)*2
		lc.PushSeries("cpu", math.Max(0, math.Min(100, cpu)))
		lc.PushSeries("mem", math.Max(0, math.Min(100, mem)))
	}
	lc.SetOffset(30)
	lc.SetCursor(45)

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#c0caf5")).
		Render("CPU AND MEMORY, SCROLLED BACK 30 POINTS")

	content := fmt.Sprintf("%s\n\n%s\n%s", header, lc.View(), lc.CursorBox())

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3b4261")).
		Padding(1, 2).
		Render(content)

	fmt.Println(box)
}

func renderSparkline() {
	const points = 24
	nodes := []struct {
		name       string
		base, wave float64
		color      string
	}{
		{"node-1", 40, 28, theme.SeriesColor(0)},
		{"node-2", 62, 25, theme.SeriesColor(1)},
		{"node-3", 80, 14, theme.SeriesColor(2)},
	}
	rng := rand.New(rand.NewSource(7))

	head := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#c0caf5"))
	text := lipgloss.NewStyle().Foreground(lipgloss.Color("#a9b1d6"))
	rows := []string{head.Render(fmt.Sprintf("%-8s %5s  %-*s  %s", "NODE", "CPU", points, "BLOCKS", "BRAILLE"))}
	for _, n := range nodes {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(n.color))
		lc := pulse.New(points, 1, pulse.WithRange(0, 100), pulse.WithSeriesStyle("cpu", style))
		for i := 0; i < points; i++ {
			v := n.base + n.wave*math.Sin(float64(i)*0.45) + (rng.Float64()-0.5)*8
			lc.PushSeries("cpu", math.Max(0, math.Min(100, v)))
		}
		last, _ := lc.Last("cpu")
		rows = append(rows, fmt.Sprintf("%s %s  %s  %s",
			text.Render(fmt.Sprintf("%-8s", n.name)),
			style.Render(fmt.Sprintf("%4.0f%%", last)),
			lc.Sparkline("cpu"),
			style.Render(pulse.SparklineBraille(lc.SeriesData("cpu"), lc.Min(), lc.Max())),
		))
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3b4261")).
		Padding(1, 2).
		Render(strings.Join(rows, "\n"))

	fmt.Println(box)
}
