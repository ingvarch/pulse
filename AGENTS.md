# AGENTS.md

Pulse is a Go library that draws live line charts in the terminal for Lip Gloss and Bubble Tea apps. The root package `pulse` holds the chart model and its two renderers, `scale/` the tick and byte-rate formatters, `theme/` the Tokyo Night preset, `examples/` the Bubble Tea demos and `cmd/gen-shots/` the screenshot generator. A change that users can see updates the guide in `docs/` that describes it, and its screenshot when the picture changes.

- [docs/development.md](docs/development.md): read before running checks or demos, regenerating screenshots, tagging a release, or looking for the file that holds a feature. It has the CI checks, the demo keys, the screenshot command and its traps, the release steps and the file map.
- [CONTRIBUTING.md](CONTRIBUTING.md): read before starting a change, a commit or a pull request. It has the branch names, the doc comment rule for exported names, the commit style with the `pulse` scope and `Closes #N`.
- [docs/getting-started.md](docs/getting-started.md): read before changing `New`, `Push`, series, gaps, styles, legends or the runtime setters in `pulse.go`, `options.go` and `legend.go`.
- [docs/rendering-modes.md](docs/rendering-modes.md): read before changing `renderer*.go`: line glyphs, line width, smoothing, braille and the fill options.
- [docs/axes-and-grid.md](docs/axes-and-grid.md): read before changing Y ticks, the grid, the time axis, the symmetric range, the log scale, the right axis or label formatters in `view.go`, `xaxis.go`, `logscale.go`, `rightaxis.go` and `scale/`.
- [docs/zero-crossing-fill.md](docs/zero-crossing-fill.md): read before changing the zero baseline, inverted series, negative styles or the byte-rate formatters.
- [docs/timeline-events.md](docs/timeline-events.md): read before changing `event.go`: event placement, guidelines, pruning and `EventsBox`.
- [docs/thresholds.md](docs/thresholds.md): read before changing `threshold.go`: threshold lines, labels and colors, and how `view.go` and `renderer*.go` use them.
- [docs/cursor-and-scrolling.md](docs/cursor-and-scrolling.md): read before changing `cursor.go` or how the window follows the offset: `SetCursor`, `SetOffset`, `CursorValues`, `CursorBox`.
- [docs/sparklines.md](docs/sparklines.md): read before changing `sparkline.go`: `Sparkline`, `SparklineBraille`, `(*Model).Sparkline` and the range the method uses.
- [docs/bubbletea.md](docs/bubbletea.md): read before changing `examples/` or how the chart is embedded in a Bubble Tea model.

## Rules

- The library packages import Lip Gloss, never Bubble Tea or gopsutil: only `examples/` use them, and `examples/` is a separate module (`go -C examples build ./...`), so the root `go build ./...` does not build the demos. CI fails when the root module graph lists Bubble Tea or gopsutil.
- Code samples in the README and `docs/` are not compiled. When you change an API, grep the docs for it and update the samples. Charm packages use the v2 paths `charm.land/lipgloss/v2` and `charm.land/bubbletea/v2`; `github.com/charmbracelet/bubbletea` is the old v1 API.
- The plan lives in GitHub milestones and issues (`gh issue list --milestone v0.8.0`), not in the repository.
