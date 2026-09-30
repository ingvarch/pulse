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
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

type tickMsg time.Time

type model struct {
	chart  *pulse.Model
	preset theme.Preset
	cpu    float64
	ram    float64
}

func tick() tea.Msg {
	time.Sleep(500 * time.Millisecond)
	return tickMsg{}
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
		if pct, err := cpu.Percent(0, false); err == nil && len(pct) > 0 {
			m.cpu = pct[0]
			m.chart.PushSeries("cpu", m.cpu)
			m.chart.SetSeriesStyle("cpu", m.preset.LineFor(m.cpu))
		}
		if vm, err := mem.VirtualMemory(); err == nil {
			m.ram = vm.UsedPercent
			m.chart.PushSeries("mem", m.ram)
		}
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

	header := fmt.Sprintf("CPU %5.1f%%   RAM %5.1f%%   [%s]   (t: rounded/bold, m: braille, q: quit)\n",
		m.cpu, m.ram, modeStr)

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
		pulse.WithSeriesStyle("cpu", preset.LineFor(0)),
		pulse.WithSeriesStyle("mem", lipgloss.NewStyle().Foreground(lipgloss.Color(theme.SeriesColor(1)))),
	)

	// Sample current telemetry to establish baseline
	curCPU := 15.0
	if pct, err := cpu.Percent(200*time.Millisecond, false); err == nil && len(pct) > 0 {
		curCPU = pct[0]
	}
	curRAM := 50.0
	if vm, err := mem.VirtualMemory(); err == nil {
		curRAM = vm.UsedPercent
	}

	// Pre-fill full window so chart renders completely on launch
	for i := 0; i < chartW; i++ {
		// subtle variation leading up to current metric
		drift := float64(i-chartW) * 0.1
		noise := (rand.Float64() - 0.5) * 3
		c := math.Max(0, math.Min(100, curCPU+drift+noise))
		r := math.Max(0, math.Min(100, curRAM+(rand.Float64()-0.5)*1.5))
		lc.PushSeries("cpu", c)
		lc.PushSeries("mem", r)
	}
	lc.SetSeriesStyle("cpu", preset.LineFor(curCPU))

	p := tea.NewProgram(model{
		chart:  lc,
		preset: preset,
		cpu:    curCPU,
		ram:    curRAM,
	})

	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
