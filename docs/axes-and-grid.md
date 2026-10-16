# Axes, Scales & Grid Alignment

## Y-Axis Auto-Scaling (`NiceTicks`)

By default, Pulse calculates visually balanced tick intervals using the **Nice Numbers** algorithm (`scale.NiceTicks`).

```go
chart := pulse.New(60, 10, pulse.WithRange(0, 100))
```

The algorithm adapts to chart height, picking clean, human-friendly steps such as multiples of 1, 2, 5, 10, 20, 25, or 50.

## Custom Ticks

For fixed ranges (e.g. CPU or RAM percent `0%..100%`), you can specify explicit quarter-step ticks:

```go
chart := pulse.New(60, 13,
    pulse.WithRange(0, 100),
    pulse.WithTicks(0, 25, 50, 75, 100),
)
```

> [!TIP]
> For perfectly even row spacing across 5 ticks (`0, 25, 50, 75, 100`), use a height of `13` terminal rows. Each quarter-step will be spaced exactly 3 terminal rows apart (`12 / 4 = 3`).

## Grafana-Style Coordinate Grid

When `pulse.WithGrid(true)` is enabled:
- **Horizontal lines**: Drawn as subtle `─` aligned precisely with each Y-axis label.
- **Vertical lines**: Drawn at quarter intervals of chart width (`w / 4`, `2w / 4`, `3w / 4`). With `WithXAxisAt`, they are drawn at the columns of the placed middle time labels instead.
- **Intersections**: Rendered with `┼`.
- **Faint styling**: Grid lines automatically inherit `axisStyle.Faint(true)` so they remain visually subtle beneath data curves.

```go
chart := pulse.New(60, 12,
    pulse.WithGrid(true),
    pulse.WithAxisStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#444b6a"))),
)
```

<p align="center">
  <img src="../assets/fill.png" alt="Grafana-Style Coordinate Grid" width="100%" />
</p>

## Time Axis (`WithXAxis`)

`pulse.WithXAxis(label)` draws an axis row and a label row under the plot. `label(age)` returns the text for the point `age` steps before the newest one, so age 0 is the newest point. The function runs on every `View()` and gets only the age. To keep the labels with the data, compute them from the time of the newest point and update that time on each push.

```go
var newest time.Time            // time of the newest point
step := 7 * 24 * time.Hour / 60 // 60 points cover 7 days

chart := pulse.New(60, 10,
    pulse.WithRange(0, 100),
    pulse.WithXAxis(func(age int) string {
        return newest.Add(-time.Duration(age) * step).Format("Jan 2")
    }),
)

// For each sample:
newest = sample.Time
chart.Push(sample.Value)
```

Labels sit at the left edge, the right edge, and at the quarter columns (`w / 4`, `2w / 4`, `3w / 4`), the same columns as the vertical grid. The left label starts at column 0, the right label ends at the last column, and the middle labels are centered on their column with a `┴` mark on the axis row.

- An empty string means no label and no mark at that column.
- A label that does not fit inside the plot width is dropped.
- Labels keep at least one space between them. The right edge is placed first, then the left edge, then the quarters. A label that would touch one already placed is dropped, so narrow charts show only the edge labels.

`Height()` stays the plot height. With the axis on, `View()` returns `Height()+2` lines: the plot rows, the axis row, and the label row. Leave room for the two extra lines when you lay the chart out. The axis rows do not depend on the grid, the render mode, or events. `chart.SetXAxis(nil)` turns the axis off.

<p align="center">
  <img src="../assets/xaxis.png" alt="Time axis under a 7-day CPU chart with a gap" width="100%" />
</p>

## Labels at Chosen Columns (`WithXAxisAt`)

`WithXAxis` puts labels at the edges and the quarter columns, so they show whatever time falls there, for example `12:45  18:33  00:37  06:41  now`. `pulse.WithXAxisAt(label)` draws the same two rows, but asks `label` for every column, ages 0 to `w-1`. A non-empty text proposes a label at that column. To get round times, return text only for the column that holds one:

```go
var newest time.Time        // time of the newest point
step := 24 * time.Hour / 80 // 80 points cover 24 hours

chart := pulse.New(80, 13,
    pulse.WithRange(0, 100),
    pulse.WithXAxisAt(func(age int) string {
        if age == 0 {
            return "now"
        }
        t := newest.Add(-time.Duration(age) * step)
        // The last 6-hour mark on the clock of t's time zone; the column holds it if it falls in (t-step, t].
        mark := time.Date(t.Year(), t.Month(), t.Day(), t.Hour()/6*6, 0, 0, 0, t.Location())
        if t.Sub(mark) < step {
            return mark.Format("15:04")
        }
        return ""
    }),
)
```

- Labels are aligned as with `WithXAxis`: the left edge label starts at column 0, the right edge label ends at the last column, and the others are centered on their column with a `┴` mark.
- The right edge is placed first, then the other columns from right to left. A label that would run past an edge or touch one already placed is dropped, so when the candidates are too close, the newer ones stay. Pick a step that leaves room: a 5-character label needs 6 columns.
- The vertical grid lines follow the placed middle labels instead of the quarters. The edge labels get no grid line, so with no middle labels there are no vertical grid lines.
- `t.Truncate(6 * time.Hour)` rounds in UTC, not in the time zone of `t`. Build the mark from the wall clock, as above, so the labels read `18:00` and not `17:30` in a zone with a half-hour offset.
- `label` runs `w` times on every `View()`, so keep it cheap.
- `chart.SetXAxisAt(nil)` turns the axis off. `SetXAxis` and `SetXAxisAt` replace each other.

<p align="center">
  <img src="../assets/xaxis-at.png" alt="Time axis with labels every 6 hours under a 24-hour CPU chart" width="100%" />
</p>

## Symmetric Y-Axis (`WithSymmetric`)

When plotting bidirectional metrics (such as network RX/TX, disk read/write, or profit/loss), you often need zero to stay directly in the vertical center of the chart with equal positive and negative scaling.

`pulse.WithSymmetric(true)` normalizes bounds to $[-\text{limit}, +\text{limit}]$ where $\text{limit} = \max(|min|, |max|)$:

```go
chart := pulse.New(60, 13,
    pulse.WithRange(0, 100),   // Automatically normalized to [-100, +100]
    pulse.WithSymmetric(true),
    pulse.WithZeroBaseline(true),
)
```

<p align="center">
  <img src="../assets/rxtx.png" alt="Symmetric Y-Axis Centered Baseline" width="100%" />
</p>

You can also dynamically toggle symmetry at runtime via `chart.SetSymmetric(bool)` or inspect it with `chart.Symmetric() bool`.

## Label Formatters & Network Rates (`scale`)

Custom tick labels can be formatted using `pulse.WithLabelFormatter(...)`. The `scale` package provides built-in formatters for human-readable bytes and throughput rates:

```go
import "github.com/ingvarch/pulse/scale"

// 1. Throughput rates with sign (e.g. "+50 MB/s", "-25 MB/s")
pulse.WithLabelFormatter(scale.FormatBytesRate)

// 2. Mirrored absolute throughput rates without minus sign (e.g. "50 MB/s" for both RX and TX)
pulse.WithLabelFormatter(scale.FormatBytesRateAbs)

// 3. Or use the configurable helper:
pulse.WithLabelFormatter(scale.BytesRateFormatter(true)) // true for absolute, false for signed
```

For static data size metrics (RAM, disk space), `scale.FormatBytes` formats values as `"512 B"`, `"1.5 KB"`, `"10 MB"`, `"2 GB"`, etc.

