package pulse

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

// DefaultEventColor is the default foreground color for timeline event pins and guidelines (#7dcfff).
var DefaultEventColor color.Color = lipgloss.Color("#7dcfff")

// Event represents a discrete timeline annotation or event marker.
type Event struct {
	ID     string         // Optional identifier for programmatic lookup or removal
	Label  string         // Human-readable description (e.g. "Deploy v1.4.2")
	Glyph  string         // Marker glyph (e.g. "▼", "🚀", "▲", "◆", "!", "⚠️"). Defaults to "▼".
	Style  lipgloss.Style // Style for the marker glyph and vertical guideline
	NoLine bool           // If true, omits the vertical guideline across the chart
}

// VisibleEvent pairs an Event with its current visible column in the chart window.
type VisibleEvent struct {
	Col   int   // Column index in current plot area (0 to Width()-1)
	Age   int   // How many time steps ago this event occurred
	Event Event // The event metadata
}

type eventRecord struct {
	event Event
	step  int
}

// AddEvent records an event at the current newest point on the timeline.
// As new points are pushed, the event moves left across the window.
func (m *Model) AddEvent(ev Event) {
	m.AddEventAt(0, ev)
}

// AddEventAt records an event at a historical offset (0 = newest point, 10 = 10 points ago).
func (m *Model) AddEventAt(offset int, ev Event) {
	if offset < 0 {
		offset = 0
	}
	m.events = append(m.events, eventRecord{
		event: ev,
		step:  m.step - offset,
	})
}

// ClearEvents removes all timeline events from the chart.
func (m *Model) ClearEvents() {
	m.events = nil
}

// VisibleEvents returns all timeline events currently visible within the chart window.
func (m *Model) VisibleEvents() []VisibleEvent {
	if len(m.events) == 0 {
		return nil
	}
	w := m.w
	out := make([]VisibleEvent, 0, len(m.events))
	pruned := make([]eventRecord, 0, len(m.events))
	for _, rec := range m.events {
		age := m.step - rec.step
		col := (w - 1) - age
		if col >= 0 && col < w {
			out = append(out, VisibleEvent{
				Col:   col,
				Age:   age,
				Event: rec.event,
			})
		}
		if col >= 0 {
			pruned = append(pruned, rec)
		}
	}
	m.events = pruned
	return out
}

// EventsBox returns a boxed card listing all currently visible timeline events.
// Returns an empty string if there are no visible events.
func (m *Model) EventsBox() string {
	events := m.VisibleEvents()
	if len(events) == 0 {
		return ""
	}
	lines := make([]string, 0, len(events))
	for _, ve := range events {
		glyph := ve.Event.Glyph
		if glyph == "" {
			glyph = "▼"
		}
		st := ve.Event.Style
		if st.GetForeground() == nil || st.GetForeground() == (lipgloss.NoColor{}) {
			st = st.Foreground(DefaultEventColor).Bold(true)
		}
		badge := st.Render(glyph)
		label := ve.Event.Label
		if label == "" {
			label = "Event"
		}
		timeAgo := fmt.Sprintf("(-%d)", ve.Age)
		if ve.Age == 0 {
			timeAgo = "(now)"
		}
		line := badge + " " + m.axisStyle.Render(label) + " " + m.axisStyle.Faint(true).Render(timeAgo)
		lines = append(lines, line)
	}
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Render(strings.Join(lines, "\n"))
}
