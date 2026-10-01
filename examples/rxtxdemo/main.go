package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/ingvarch/pulse"
	"github.com/ingvarch/pulse/scale"
	"github.com/ingvarch/pulse/theme"
)

const MB = 1024.0 * 1024.0

type tickMsg time.Time

type model struct {
	chart *pulse.Model
	step  int
}

func tick() tea.Msg {
	time.Sleep(150 * time.Millisecond)
	return tickMsg{}
}

func rxSignal(step int) float64 {
	base := 45 + 30*math.Sin(float64(step)*0.12) + 15*math.Cos(float64(step)*0.05)
	if step%23 < 5 {
		base += 25 + 15*rand.Float64()
	}
	return math.Max(5, math.Min(95, base)) * MB
}

func txSignal(step int) float64 {
	base := 35 + 25*math.Cos(float64(step)*0.09) + 10*math.Sin(float64(step)*0.04)
	if (step+10)%27 < 4 {
		base += 30 + 10*rand.Float64()
	}
	return math.Max(5, math.Min(95, base)) * MB
}

func (m model) Init() tea.Cmd { return tick }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		s := msg.String()
		if s == "q" || s == "ctrl+c" || s == "esc" {
			return m, tea.Quit
		}
		if s == "m" {
			m.chart.ToggleRenderMode()
		}
		if s == "w" {
			if m.chart.LineWidth() == 1 {
				m.chart.SetLineWidth(2)
			} else {
				m.chart.SetLineWidth(1)
			}
		}
		if s == "s" {
			m.chart.SetSmooth(!m.chart.Smooth())
		}
		if s == "y" {
			m.chart.SetSymmetric(!m.chart.Symmetric())
		}
		if s == "f" {
			m.chart.SetFill(!m.chart.Fill())
		}
		if s == "t" {
			m.chart.SetTintedFill(!m.chart.TintedFill())
		}
		if s == "z" {
			m.chart.SetZeroBaseline(!m.chart.ZeroBaseline())
		}
	case tickMsg:
		m.step++
		rx := rxSignal(m.step)
		tx := txSignal(m.step)
		m.chart.PushSeries("rx", rx)
		m.chart.PushSeries("tx", tx)
		return m, tick
	}
	return m, nil
}

func (m model) View() tea.View {
	rx, _ := m.chart.Last("rx")
	tx, _ := m.chart.Last("tx")

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7dcfff"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#565f89"))
	rxStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#9ece6a")).Bold(true)
	txStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#bb9af7")).Bold(true)

	modeStr := "Lines"
	if m.chart.Mode() == pulse.ModeBraille {
		modeStr = "Braille"
	}
	cornerStr := "rounded ╭╯"
	if !m.chart.Smooth() {
		cornerStr = "sharp ┌┘"
	}
	if m.chart.LineWidth() > 1 {
		cornerStr = "bold ┏┛"
	}

	fillStr := "off"
	if m.chart.Fill() {
		if m.chart.TintedFill() {
			fillStr = "tinted"
		} else {
			fillStr = "on"
		}
	}

	symStr := "asym"
	if m.chart.Symmetric() {
		symStr = "sym"
	}

	header := lipgloss.JoinHorizontal(
		lipgloss.Center,
		titleStyle.Render("⚡ Network Interface (eth0) Traffic"),
		dimStyle.Render(fmt.Sprintf("  [%s (w/s) | %s (m) | %s (y) | fill: %s (f/t) | zero: %v (z) | q: quit]",
			cornerStr, modeStr, symStr, fillStr, m.chart.ZeroBaseline())),
	)

	stats := lipgloss.JoinHorizontal(
		lipgloss.Center,
		rxStyle.Render(fmt.Sprintf(" ▲ RX: %s", scale.FormatBytesRate(rx))),
		"    ",
		txStyle.Render(fmt.Sprintf(" ▼ TX: %s", scale.FormatBytesRate(tx))),
		"    ",
		m.chart.LegendBox(),
	)

	out := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		m.chart.View(),
		"",
		stats,
	) + "\n"

	return tea.NewView(out)
}

func main() {
	preset := theme.TokyoNight()

	chart := pulse.New(65, 15,
		pulse.WithLineWidth(1),
		pulse.WithRange(-100*MB, 100*MB),
		pulse.WithSymmetric(true),
		pulse.WithZeroBaseline(true),
		pulse.WithTicks(-100*MB, -50*MB, 0, 50*MB, 100*MB),
		pulse.WithLabelFormatter(scale.FormatBytesRateAbs),
		pulse.WithSeriesStyle("rx", lipgloss.NewStyle().Foreground(lipgloss.Color("#9ece6a"))),
		pulse.WithSeriesStyle("tx", lipgloss.NewStyle().Foreground(lipgloss.Color("#bb9af7"))),
		pulse.WithSeriesInverted("tx", true),
		pulse.WithAxisStyle(preset.Axis),
		pulse.WithFill(true),
		pulse.WithTintedFill(true),
	)

	// Pre-seed historical window
	for i := 0; i < 65; i++ {
		chart.PushSeries("rx", rxSignal(i))
		chart.PushSeries("tx", txSignal(i))
	}

	p := tea.NewProgram(model{chart: chart, step: 65})
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
