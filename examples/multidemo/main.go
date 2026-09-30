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
	"github.com/ingvarch/pulse/theme"
)

type tickMsg time.Time

type model struct {
	chart  *pulse.Model
	preset theme.Preset
	step   int
}

func tick() tea.Msg {
	time.Sleep(200 * time.Millisecond)
	return tickMsg{}
}

func cpuSignal(step int) float64 {
	v := 55 + 25*math.Sin(float64(step)*0.14) + 12*math.Cos(float64(step)*0.06)
	v += (rand.Float64() - 0.5) * 4
	return math.Max(0, math.Min(100, v))
}

func memSignal(step int) float64 {
	v := 45 + 15*math.Sin(float64(step)*0.04) + 5*math.Sin(float64(step)*0.1)
	return math.Max(0, math.Min(100, v))
}

func netSignal(step int) float64 {
	// Base low activity with periodic bursts
	base := 18 + 8*math.Sin(float64(step)*0.08)
	if step%17 < 4 {
		base += 35 + 20*rand.Float64()
	}
	return math.Max(0, math.Min(100, base))
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
			return m, nil
		}
		if s == "t" {
			if m.chart.LineWidth() == 1 {
				m.chart.SetLineWidth(2)
			} else {
				m.chart.SetLineWidth(1)
			}
			return m, nil
		}
	case tickMsg:
		m.step++
		c := cpuSignal(m.step)
		mem := memSignal(m.step)
		net := netSignal(m.step)

		m.chart.PushSeries("cpu", c)
		m.chart.PushSeries("mem", mem)
		m.chart.PushSeries("net", net)

		return m, tick
	}
	return m, nil
}

func (m model) View() tea.View {
	modeStr := "rounded ╭╯"
	if m.chart.Mode() == pulse.ModeBraille {
		modeStr = "braille ⠒"
	} else if m.chart.LineWidth() > 1 {
		modeStr = "bold ┏┛"
	}

	cpuVal, _ := m.chart.Last("cpu")
	memVal, _ := m.chart.Last("mem")
	netVal, _ := m.chart.Last("net")

	header := fmt.Sprintf("CPU %4.1f%%   MEM %4.1f%%   NET %4.1f%%   [%s]   (t: rounded/bold, m: braille, q: quit)\n",
		cpuVal, memVal, netVal, modeStr)

	return tea.NewView(header + m.chart.View() + "\n" + m.chart.LegendBox() + "\n")
}

func main() {
	preset := theme.TokyoNight()
	chartW, chartH := 80, 13

	lc := pulse.New(chartW, chartH,
		pulse.WithRange(0, 100),
		pulse.WithTicks(0, 25, 50, 75, 100),
		pulse.WithLineWidth(1),
		pulse.WithAxisStyle(preset.Axis),
		pulse.WithSeriesStyle("cpu", lipgloss.NewStyle().Foreground(lipgloss.Color(theme.SeriesColor(1)))), // blue
		pulse.WithSeriesStyle("mem", lipgloss.NewStyle().Foreground(lipgloss.Color(theme.SeriesColor(0)))), // green
		pulse.WithSeriesStyle("net", lipgloss.NewStyle().Foreground(lipgloss.Color(theme.SeriesColor(4)))), // yellow
	)

	// Pre-fill full window for immediate complete rendering
	initStep := chartW
	for i := 1; i <= initStep; i++ {
		lc.PushSeries("cpu", cpuSignal(i))
		lc.PushSeries("mem", memSignal(i))
		lc.PushSeries("net", netSignal(i))
	}

	p := tea.NewProgram(model{
		chart:  lc,
		preset: preset,
		step:   initStep,
	})

	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
