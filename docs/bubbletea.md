# Integrating with Bubble Tea

Pulse is designed to embed seamlessly inside [Charm Bubble Tea](https://github.com/charmbracelet/bubbletea) models.

## Architecture Pattern

Store a `*pulse.Model` inside your Bubble Tea state. In `Update`, push new values upon receiving tick messages. In `View`, embed `m.chart.View()`.

```go
package main

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/ingvarch/pulse"
	"github.com/ingvarch/pulse/theme"
)

type tickMsg time.Time

type model struct {
	chart  *pulse.Model
	preset theme.Preset
	cpu    float64
}

func tick() tea.Msg {
	time.Sleep(200 * time.Millisecond)
	return tickMsg{}
}

func (m model) Init() tea.Cmd { return tick }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "m":
			// Toggle between Braille and smooth box-drawing lines
			m.chart.ToggleRenderMode()
		case "t":
			// Toggle line thickness
			if m.chart.LineWidth() == 1 {
				m.chart.SetLineWidth(2)
			} else {
				m.chart.SetLineWidth(1)
			}
		}
	case tickMsg:
		// Push latest metric
		m.chart.PushSeries("cpu", m.cpu)
		m.chart.SetSeriesStyle("cpu", m.preset.LineFor(m.cpu))
		return m, tick
	}
	return m, nil
}

func (m model) View() tea.View {
	return tea.NewView(fmt.Sprintf("%s\n%s\n", m.chart.View(), m.chart.LegendBox()))
}
```

See [examples/livedemo](file:///examples/livedemo/main.go) for a complete interactive implementation.
