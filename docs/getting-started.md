# Getting Started with Pulse

**Pulse** is a terminal line charting component built for [Lip Gloss](https://github.com/charmbracelet/lipgloss) and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Installation

```bash
go get github.com/ingvarch/pulse
```

## Basic Usage

Create a new chart with `pulse.New(width, height, options...)`, push numeric values, and call `.View()` to produce the formatted string.

```go
package main

import (
	"fmt"
	"math"

	"github.com/ingvarch/pulse"
)

func main() {
	// Create a chart of 40 columns by 10 rows
	chart := pulse.New(40, 10,
		pulse.WithRange(0, 100),
		pulse.WithTicks(0, 25, 50, 75, 100),
		pulse.WithFill(true),
	)

	// Feed a sine wave
	for i := 0; i < 40; i++ {
		val := 50 + 40*math.Sin(float64(i)*0.4)
		chart.Push(val)
	}

	// Render to stdout
	fmt.Println(chart.View())
}
```

<p align="center">
  <img src="../assets/quickstart.png" alt="Pulse Basic Usage Output" width="100%" />
</p>

## Sliding Window

`pulse.Model` maintains an internal sliding buffer equal to the chart's width (`w`). When you call `Push(val)` or `PushSeries(name, val)` beyond the width, older points automatically roll off, creating a smooth real-time stream.

To query buffer lengths:
```go
totalPoints := chart.Len()             // Default series
cpuPoints   := chart.LenSeries("cpu")  // Named series
```

## Multi-Series Plotting

You can plot multiple concurrent time-series on the same coordinate space:

```go
chart.PushSeries("cpu", cpuUsage)
chart.PushSeries("mem", memUsage)
chart.PushSeries("net", netRps)
```

Each series can have its own Lip Gloss style:
```go
import "charm.land/lipgloss/v2"

chart.SetSeriesStyle("cpu", lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5555")))
chart.SetSeriesStyle("mem", lipgloss.NewStyle().Foreground(lipgloss.Color("#50fa7b")))
```

## Legends and Value Inspection

Pulse provides built-in legend helpers:

```go
// Inline legend: ■ cpu  ■ mem
fmt.Println(chart.Legend())

// Boxed legend with live values and borders:
// ╭─────────────╮
// │ ■ cpu 88.67 │
// │ ■ mem 41.00 │
// ╰─────────────╯
fmt.Println(chart.LegendBox())

// Inspect the last pushed value:
if val, ok := chart.Last("cpu"); ok {
    fmt.Printf("Current CPU: %.2f%%\n", val)
}
```

## Dynamic Settings

All styling and rendering options can be modified dynamically at runtime:
```go
chart.SetGrid(true)            // toggle grid
chart.SetFill(false)           // toggle shaded area fill
chart.SetLineWidth(2)          // switch between thin (1) and bold (2)
chart.SetMode(pulse.ModeBraille)// switch between ModeLines and ModeBraille
chart.ToggleRenderMode()       // flip between lines and braille
```
