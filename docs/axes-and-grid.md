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
- **Vertical lines**: Drawn at quarter intervals of chart width (`w / 4`, `2w / 4`, `3w / 4`).
- **Intersections**: Rendered with `┼`.
- **Faint styling**: Grid lines automatically inherit `axisStyle.Faint(true)` so they remain visually subtle beneath data curves.

```go
chart := pulse.New(60, 12,
    pulse.WithGrid(true),
    pulse.WithAxisStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#444b6a"))),
)
```
