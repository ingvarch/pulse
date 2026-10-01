package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/ingvarch/pulse"
	"github.com/ingvarch/pulse/theme"
)

type tickMsg time.Time

type model struct {
	chart  *pulse.Model
	preset theme.Preset
	step   int
	last   float64
}

func tick() tea.Msg {
	time.Sleep(200 * time.Millisecond)
	return tickMsg{}
}

// signal generates smooth rolling hills with light noise in range 0-100.
func signal(step int) float64 {
	v := 50 + 28*math.Sin(float64(step)*0.12) + 8*math.Sin(float64(step)*0.05+1)
	v += (rand.Float64() - 0.5) * 2
	if v < 0 {
		v = 0
	}
	if v > 100 {
		v = 100
	}
	return v
}

func (m model) Init() tea.Cmd { return tick }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		s := msg.String()
		if s == "q" || s == "ctrl+c" || s == "esc" {
			return m, tea.Quit
		}
		if s == "f" {
			m.chart.SetTintedFill(!m.chart.TintedFill())
			return m, nil
		}
		if s == "s" {
			m.chart.SetSolidFill(!m.chart.SolidFill())
			return m, nil
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
		m.last = signal(m.step)
		m.chart.PushSeries("signal", m.last)
		m.chart.SetSeriesStyle("signal", m.preset.LineFor(m.last))
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
	tintStr := "tinted: on"
	if !m.chart.TintedFill() {
		tintStr = "tinted: off"
	}
	fillStr := "textured ░"
	if m.chart.SolidFill() {
		fillStr = "solid fill"
	}
	header := fmt.Sprintf("signal %5.1f  [%s, %s, %s]  (f: tint, s: solid/textured, t: width, m: braille, q: quit)\n",
		m.last, modeStr, tintStr, fillStr)
	return tea.NewView(header + m.chart.View() + "\n" + m.chart.LegendBox() + "\n")
}

func main() {
	preset := theme.TokyoNight()
	chartW, chartH := 80, 13

	lc := pulse.New(chartW, chartH,
		pulse.WithRange(0, 100),
		pulse.WithTicks(0, 25, 50, 75, 100),
		pulse.WithLineWidth(2),
		pulse.WithTintedFill(true),
		pulse.WithAxisStyle(preset.Axis),
		pulse.WithSeriesStyle("signal", preset.LineFor(50)),
	)

	// Pre-fill initial data for immediate graph rendering
	initStep := chartW
	for i := 1; i <= initStep; i++ {
		v := signal(i)
		lc.PushSeries("signal", v)
	}
	lastVal := signal(initStep)
	lc.SetSeriesStyle("signal", preset.LineFor(lastVal))

	p := tea.NewProgram(model{
		chart:  lc,
		preset: preset,
		step:   initStep,
		last:   lastVal,
	})
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
