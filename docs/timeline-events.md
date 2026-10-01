# Timeline Annotations & Event Markers

**Pulse** is the first Go terminal charting library to feature **first-class Timeline Annotations & Event Markers**, bringing a core capability from cloud observability platforms like **Grafana** and **Datadog** straight into your terminal user interface.

---

## 🎯 Why Timeline Annotations?

When monitoring metrics in real time (e.g. CPU spikes, memory leaks, error rate surges), raw numbers only tell you *that* something happened—not *why* it happened.

In modern operations and incident response, correlating metrics with discrete engineering events is essential:
- **Deployments**: *"Did latency increase right after deploying `v2.4.1`?"*
- **Autoscaling & Restarts**: *"Was this CPU drop caused by a pod OOM-kill or HPA scale-up?"*
- **Configuration Changes**: *"Did DB connection pool saturation start when feature flag `X` was enabled?"*
- **Chaos / Load Tests**: *"How did response times behave during the simulated network partition?"*

Pulse enables you to pin events directly onto the time-series grid with vertical guidelines (`┆`), custom badges or emojis (`▼`, `🚀`, `⚡`, `⚠️`), and dynamic aging cards.

```text
 100│─────────────▼─────── ╭──────────╮  [Deploy v1.4.2 (-24)]
    │─────────────┆─────╭──╯░░░░░░░░░░╰──
  50│─────────────┆───╭─╯░░░░░░░░░░░░░░░░
   0│─────────────┆───╯░░░░░░░░░░░░░░░░░░
```

---

## 📦 Core Types & API

### `Event`

```go
type Event struct {
    ID     string         // Optional identifier for programmatic lookup or removal
    Label  string         // Human-readable description (e.g., "Deploy v1.4.2")
    Glyph  string         // Marker glyph (e.g. "▼", "🚀", "⚡", "▲", "◆"). Defaults to "▼"
    Style  lipgloss.Style // Lip Gloss style for the glyph and guideline
    NoLine bool           // If true, omits the vertical guideline across the chart
}
```

### `VisibleEvent`

```go
type VisibleEvent struct {
    Col   int   // Column index within the visible chart window (0 to Width()-1)
    Age   int   // How many time steps ago this event occurred
    Event Event // The original event metadata
}
```

### Chart Methods

| Method | Description |
| ------ | ----------- |
| `chart.AddEvent(ev Event)` | Pins a new event at the current newest point (`now`, right edge). |
| `chart.AddEventAt(offset int, ev Event)` | Pins an event at a historical offset (`0` = newest point, `15` = 15 points ago). |
| `chart.VisibleEvents() []VisibleEvent` | Returns all events currently visible within the sliding window. Automatically prunes off-screen events. |
| `chart.EventsBox() string` | Returns a formatted, bordered Lip Gloss card displaying all active visible events with age tags (`(now)`, `(-14)`). |
| `chart.ClearEvents()` | Clears all registered events from the chart. |

### Functional Option

```go
pulse.New(width, height,
    // Pin an initial event 25 points in the past
    pulse.WithEvent(25, pulse.Event{
        Label: "Traffic Spike",
        Glyph: "⚡",
        Style: lipgloss.NewStyle().Foreground(lipgloss.Color("#e0af68")).Bold(true),
    }),
)
```

---

## 🛠️ How It Works: The Sliding Window Lifecycle

Pulse charts maintain an internal time step (`m.step`). Every time you push a data point (`chart.Push(v)` or `chart.PushSeries(name, v)`), the timeline advances by one tick:

1. **Recording**: When you call `chart.AddEvent(ev)`, Pulse stamps the event with the current `step`.
2. **Aging**: At any moment, the event's age is calculated as:
   $$\text{age} = \text{current\_step} - \text{event\_step}$$
3. **Column Mapping**: In a chart of width $W$, the screen column is:
   $$\text{col} = (W - 1) - \text{age}$$
   - When $\text{age} = 0$, the marker appears at the extreme right edge: `col = W - 1`.
   - As new points are pushed, the marker smoothly travels leftward across the grid.
4. **Rendering**:
   - The pin glyph (`▼`, `🚀`, `⚡`) is anchored at row 0 (the top margin).
   - If `NoLine` is `false`, a subtle vertical guideline (`┆`) cascades down through empty space and shaded/tinted area fill, preserving line curves (`╭`, `╯`, `─`) without disruption.
5. **Garbage Collection**: Once an event scrolls past the left edge ($\text{col} < 0$), Pulse automatically purges it from memory.

---

## 💡 Practical Examples

### 1. Static Report / CLI Incident Summary

Pin past events on a static time-series chart before rendering:

```go
package main

import (
	"fmt"
	"math"

	"charm.land/lipgloss/v2"
	"github.com/ingvarch/pulse"
	"github.com/ingvarch/pulse/theme"
)

func main() {
	preset := theme.TokyoNight()
	chart := pulse.New(60, 12,
		pulse.WithRange(0, 100),
		pulse.WithTicks(0, 25, 50, 75, 100),
		pulse.WithFill(true),
		pulse.WithTintedFill(true),
		pulse.WithAxisStyle(preset.Axis),
		pulse.WithLineStyle(preset.LineFor(60)),
	)

	// Stream synthetic metric values
	for i := 0; i < 60; i++ {
		val := 40.0 + 35.0*math.Sin(float64(i)*0.15)
		chart.Push(val)
	}

	// Annotate historical events
	chart.AddEventAt(35, pulse.Event{
		Label: "Deploy v1.4.2",
		Glyph: "🚀",
		Style: lipgloss.NewStyle().Foreground(lipgloss.Color("#7dcfff")).Bold(true),
	})

	chart.AddEventAt(15, pulse.Event{
		Label: "Spike Detected",
		Glyph: "⚡",
		Style: lipgloss.NewStyle().Foreground(lipgloss.Color("#e0af68")).Bold(true),
	})

	// Render chart and events summary card
	fmt.Println(chart.View())
	fmt.Println(chart.EventsBox())
}
```

---

### 2. Live Bubble Tea Dashboard with Dynamic Event Injection

Trigger events dynamically via user hotkeys or incoming background messages:

```go
package main

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"charm.land/lipgloss/v2"
	"github.com/ingvarch/pulse"
)

type tickMsg time.Time

type model struct {
	chart *pulse.Model
	step  int
}

func (m model) Init() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "e": // Drop a deploy event at (now)
			m.chart.AddEvent(pulse.Event{
				Label: fmt.Sprintf("Deploy #%d", m.step),
				Glyph: "🚀",
				Style: lipgloss.NewStyle().Foreground(lipgloss.Color("#7dcfff")).Bold(true),
			})
		case "c": // Clear all events
			m.chart.ClearEvents()
		}

	case tickMsg:
		m.step++
		// Push latest metric
		m.chart.Push(computeCurrentMetric(m.step))
		return m, tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	}

	return m, nil
}

func (m model) View() string {
	chartView := m.chart.View()
	eventsCard := m.chart.EventsBox()

	if eventsCard != "" {
		return lipgloss.JoinVertical(lipgloss.Left, chartView, eventsCard)
	}
	return chartView
}
```

---

## 🎨 Best Practices & UI Tips

### 1. Choosing Marker Glyphs

Pulse supports single-character Unicode symbols, ASCII, and multi-width emojis:

| Glyph | Recommended Purpose | Sample Style Color |
| :---: | ------------------- | ------------------ |
| `▼` | General event marker / pin | `#7dcfff` (Cyan) |
| `🚀` | Application deployment or rollout | `#bb9af7` (Purple) |
| `⚡` | Traffic spike or load increase | `#e0af68` (Amber) |
| `⚠️` | Warning or threshold breach | `#f7768e` (Red) |
| `🏷️` | Release or milestone tag | `#73daca` (Teal) |
| `🔄` | Pod restart or container recycle | `#2ac3de` (Sky Blue) |
| `◆` | Discrete log event or audit action | `#9ece6a` (Green) |

> [!TIP]
> Pulse automatically measures terminal display width via `lipgloss.Width(glyph)`. If you use double-width emojis (such as `🚀` or `⚡`), Pulse offsets surrounding grid cells to ensure the vertical guideline `┆` stays strictly aligned with the marker.

---

### 2. Guideline Visibility (`NoLine`)

By default, every event drops a subtle dotted vertical line (`┆`) through the chart:
- **`NoLine: false` (Default)**: Best for single or spaced-out events where you need to trace the exact point in time down to the baseline or across filled areas.
- **`NoLine: true`**: Best when you have high-frequency events or dense timelines, avoiding vertical stripes across the plot area.

```go
chart.AddEvent(pulse.Event{
    Label:  "Minor Warning",
    Glyph:  "⚠️",
    NoLine: true, // Marker only on row 0; no vertical line
})
```

---

### 3. Pairing with `LegendBox` and `EventsBox`

For a production-grade TUI layout, place the `LegendBox()` and `EventsBox()` side-by-side beneath the chart using `lipgloss.JoinHorizontal`:

```go
footer := lipgloss.JoinHorizontal(
    lipgloss.Top,
    chart.LegendBox(),
    "  ",
    chart.EventsBox(),
)

fullDashboard := lipgloss.JoinVertical(
    lipgloss.Left,
    chart.View(),
    footer,
)
```

---

### 4. Compatibility with Rendering Engines

Timeline annotations work seamlessly across all Pulse rendering modes:
- **`pulse.ModeLines`**: Guidelines run through box-drawing turns without breaking corner connections. In tinted fill mode (`WithTintedFill(true)`), guidelines automatically inherit the underlying tint background.
- **`pulse.ModeBraille`**: Markers and guidelines cleanly overlay the 2×4 sub-pixel Braille matrix.
