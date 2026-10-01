package pulse_test

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/ingvarch/pulse"
)

func TestLineChart_TimelineEventsSlidingWindow(t *testing.T) {
	m := pulse.New(40, 8)
	for i := 0; i < 40; i++ {
		m.Push(50)
	}
	// Add event at newest point
	m.AddEvent(pulse.Event{Label: "Deploy", Glyph: "🚀"})
	events := m.VisibleEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 visible event, got %d", len(events))
	}
	if events[0].Col != 39 {
		t.Fatalf("expected event at col 39, got %d", events[0].Col)
	}
	if events[0].Age != 0 {
		t.Fatalf("expected age 0, got %d", events[0].Age)
	}

	// Push 10 points
	for i := 0; i < 10; i++ {
		m.Push(55)
	}
	events = m.VisibleEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 visible event, got %d", len(events))
	}
	if events[0].Col != 29 {
		t.Fatalf("expected event at col 29, got %d", events[0].Col)
	}
	if events[0].Age != 10 {
		t.Fatalf("expected age 10, got %d", events[0].Age)
	}

	// Push 30 more points -> event scrolls off screen
	for i := 0; i < 30; i++ {
		m.Push(60)
	}
	events = m.VisibleEvents()
	if len(events) != 0 {
		t.Fatalf("expected 0 visible events after scrolling off screen, got %d", len(events))
	}
}

func TestLineChart_EventsBoxFormatting(t *testing.T) {
	m := pulse.New(40, 8)
	if box := m.EventsBox(); box != "" {
		t.Fatalf("expected empty EventsBox() when no events exist, got %q", box)
	}
	for i := 0; i < 40; i++ {
		m.Push(50)
	}
	m.AddEvent(pulse.Event{Label: "Deploy v1.2", Glyph: "🚀"})
	m.AddEventAt(10, pulse.Event{Label: "Alert Triggered", Glyph: "⚠️"})

	box := m.EventsBox()
	if !strings.Contains(box, "Deploy v1.2") || !strings.Contains(box, "🚀") {
		t.Fatalf("EventsBox missing Deploy event: %s", box)
	}
	if !strings.Contains(box, "Alert Triggered") || !strings.Contains(box, "⚠️") {
		t.Fatalf("EventsBox missing Alert event: %s", box)
	}
	if !strings.Contains(box, "(now)") || !strings.Contains(box, "(-10)") {
		t.Fatalf("EventsBox missing age offsets: %s", box)
	}

	m.ClearEvents()
	if box := m.EventsBox(); box != "" {
		t.Fatalf("expected empty EventsBox() after ClearEvents(), got %q", box)
	}
}

func TestLineChart_EventsRenderInLinesAndBraille(t *testing.T) {
	cyan := lipgloss.NewStyle().Foreground(lipgloss.Color("#7dcfff"))
	m := pulse.New(40, 8,
		pulse.WithRange(0, 100),
		pulse.WithLineStyle(cyan),
		pulse.WithFill(true),
		pulse.WithTintedFill(true),
	)
	for i := 0; i < 40; i++ {
		m.Push(50)
	}
	m.AddEventAt(15, pulse.Event{Label: "Deploy", Glyph: "🚀"})

	// Lines renderer check
	viewLines := m.View()
	if !strings.Contains(viewLines, "🚀") {
		t.Fatalf("ModeLines View() must render event glyph 🚀")
	}
	if !strings.Contains(viewLines, "┆") {
		t.Fatalf("ModeLines View() must render event guideline ┆")
	}

	// Braille renderer check
	m.SetMode(pulse.ModeBraille)
	viewBraille := m.View()
	if !strings.Contains(viewBraille, "🚀") {
		t.Fatalf("ModeBraille View() must render event glyph 🚀")
	}
	if !strings.Contains(viewBraille, "┆") {
		t.Fatalf("ModeBraille View() must render event guideline ┆")
	}
}

func TestLineChart_WithEventOption(t *testing.T) {
	m := pulse.New(30, 8,
		pulse.WithEvent(5, pulse.Event{Label: "InitEvent", Glyph: "★"}),
	)
	for i := 0; i < 10; i++ {
		m.Push(42)
	}
	events := m.VisibleEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 event from WithEvent option, got %d", len(events))
	}
	if events[0].Event.Label != "InitEvent" || events[0].Event.Glyph != "★" {
		t.Fatalf("unexpected event metadata: %+v", events[0])
	}
	if events[0].Col != 14 {
		t.Fatalf("expected col 14 (29 - 15), got %d", events[0].Col)
	}
}
