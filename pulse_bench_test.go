package pulse_test

import (
	"math"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/ingvarch/pulse"
)

func seedBenchmarkData(m *pulse.Model, count int) {
	for i := 0; i < count; i++ {
		v := 50.0 + 35.0*math.Sin(float64(i)*0.15)
		m.Push(v)
	}
}

func BenchmarkView_ModeLines(b *testing.B) {
	m := pulse.New(80, 20,
		pulse.WithRange(0, 100),
		pulse.WithMode(pulse.ModeLines),
		pulse.WithFill(false),
	)
	seedBenchmarkData(m, 80)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.View()
	}
}

func BenchmarkView_ModeLines_TintedFill(b *testing.B) {
	m := pulse.New(80, 20,
		pulse.WithRange(0, 100),
		pulse.WithMode(pulse.ModeLines),
		pulse.WithFill(true),
		pulse.WithTintedFill(true),
	)
	seedBenchmarkData(m, 80)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.View()
	}
}

func BenchmarkView_ModeBraille(b *testing.B) {
	m := pulse.New(80, 20,
		pulse.WithRange(0, 100),
		pulse.WithMode(pulse.ModeBraille),
	)
	seedBenchmarkData(m, 80)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.View()
	}
}

func BenchmarkModel_Push(b *testing.B) {
	m := pulse.New(80, 20, pulse.WithRange(0, 100))

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Push(float64(i % 100))
	}
}

func BenchmarkModel_LegendBox(b *testing.B) {
	m := pulse.New(80, 20, pulse.WithRange(0, 100))
	m.PushSeries("cpu", 75.5)
	m.PushSeries("mem", 42.1)
	m.PushSeries("net", 12.8)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.LegendBox()
	}
}

func BenchmarkModel_EventsBox(b *testing.B) {
	m := pulse.New(80, 20, pulse.WithRange(0, 100))
	seedBenchmarkData(m, 80)
	m.AddEventAt(10, pulse.Event{Label: "Deploy v1.4.2", Glyph: "🚀"})
	m.AddEventAt(30, pulse.Event{Label: "Traffic Spike", Glyph: "⚡", Style: lipgloss.NewStyle().Foreground(lipgloss.Color("#e0af68"))})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.EventsBox()
	}
}
