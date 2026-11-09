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
	case tea.WindowSizeMsg:
		// Leave room for the label column, the axis and the legend
		m.chart.SetSize(max(msg.Width-6, 1), max(msg.Height-6, 1))
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

## Cursor and Scrolling

Bind keys to `SetCursor` and `SetOffset`, and show `CursorBox` under the chart. Create the chart with `pulse.WithHistory(n)` so there is history to scroll through. See [Cursor and Scrolling](cursor-and-scrolling.md).

```go
case "left": // older; a cursor outside the view starts at its right edge
	if c, o := m.chart.Cursor(), m.chart.Offset(); c < o || c >= o+m.chart.Width() {
		m.chart.SetCursor(o)
	} else {
		m.chart.SetCursor(c + 1)
	}
case "right": // newer; at age 0 the cursor hides
	m.chart.SetCursor(m.chart.Cursor() - 1)
case "[":
	m.chart.SetOffset(m.chart.Offset() + max(m.chart.Width()/4, 1))
case "]":
	m.chart.SetOffset(m.chart.Offset() - max(m.chart.Width()/4, 1))
case "esc":
	m.chart.SetCursor(-1)
	m.chart.SetOffset(0)
```

## Resizing

Terminal size changes arrive as `tea.WindowSizeMsg`. Call `SetSize` with the plot size: the window width minus the label column and the axis (and the right axis line and its label column, with `WithRightAxis`), and the height minus the lines you draw around the chart. The data stays in the chart, so you do not create a new one.

Create the chart with `pulse.WithHistory(n)` to keep points that a narrow window hides. When the terminal gets wider, the chart shows them again. Without it, the new columns on the left stay empty until new points arrive. See [History and Resizing](getting-started.md#history-and-resizing).

See [examples/livedemo](../examples/livedemo/main.go) for a complete interactive implementation.
