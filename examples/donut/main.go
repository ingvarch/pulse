package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/ingvarch/pulse/donut"
	"github.com/ingvarch/pulse/theme"
)

type tickMsg time.Time

type phase struct {
	name  string
	pods  float64
	color string
}

type model struct {
	ring   *donut.Model
	frame  *donut.Model
	phases []phase
	rng    *rand.Rand
}

func tick() tea.Msg {
	time.Sleep(time.Second)
	return tickMsg{}
}

func (m *model) sync() {
	sections := make([]donut.Section, len(m.phases))
	for i, p := range m.phases {
		sections[i] = donut.Section{Name: p.name, Value: p.pods, Style: lipgloss.NewStyle().Foreground(lipgloss.Color(p.color))}
	}
	m.ring.SetSections(sections...)
	m.frame.SetSections(sections...)
}

func (m *model) cycleGlyphs() {
	next := (m.ring.Glyphs() + 1) % 3
	m.ring.SetGlyphs(next)
	m.frame.SetGlyphs(next)
}

func (m model) Init() tea.Cmd { return tick }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "g":
			m.cycleGlyphs()
		}
	case tickMsg:
		for i := range m.phases {
			m.phases[i].pods = max(0, m.phases[i].pods+float64(m.rng.Intn(7)-3))
		}
		m.phases[0].pods += math.Round((60 - m.phases[0].pods) / 10) // Running pods drift back to about 60
		m.sync()
		return m, tick
	}
	return m, nil
}

func (m model) View() tea.View {
	names := [...]string{"braille", "octants", "half blocks"}
	header := fmt.Sprintf("Pods by phase   [%s]   (g: glyphs, q: quit)\n\n", names[m.ring.Glyphs()])
	body := lipgloss.JoinHorizontal(lipgloss.Top, m.ring.View(), "    ", m.frame.View())
	return tea.NewView(header + body + "\n\n" + m.ring.Legend() + "\n")
}

func main() {
	m := model{
		ring:  donut.New(12, donut.WithCenterLabel("Pods")),
		frame: donut.New(12, donut.WithFrame(40, 12, 4, 8)),
		phases: []phase{
			{"Running", 52, theme.SeriesColor(0)},
			{"Pending", 8, theme.SeriesColor(4)},
			{"Failed", 4, theme.SeriesColor(5)},
		},
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	m.sync()
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
