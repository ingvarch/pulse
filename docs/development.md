# Development

How to build, check and release Pulse, run the demos and regenerate the screenshots. To use the library in your own code, start with [Getting Started](getting-started.md).

## Prerequisites

- Go 1.26 or newer (`go.mod` says `go 1.26.0`).
- [freeze](https://github.com/charmbracelet/freeze), only to regenerate the screenshots in `assets/`.

```bash
git clone https://github.com/ingvarch/pulse.git
cd pulse
```

## Checks

CI runs these on Ubuntu and macOS for every push and pull request to `main`. Run them before you open a pull request:

```bash
gofmt -l .                     # must print nothing; gofmt -s -w . fixes it
go vet ./...
go -C examples vet ./...
go test -race ./...
go build ./...                 # the library and cmd/gen-shots
go -C examples build ./...     # the demos
```

The demos in `examples/` are a separate module, `github.com/ingvarch/pulse/examples`, so Bubble Tea and gopsutil stay out of the module graph of every program that imports pulse. Its `go.mod` replaces `github.com/ingvarch/pulse` with `../`, so the demos build against the working tree. CI fails when the root `go list -m all` lists Bubble Tea or gopsutil.

## Run the Demos

Each demo is a Bubble Tea program in the `examples/` module. Run them from the repository root with `go -C examples run`. `q` or `Ctrl+C` quits, and so does `Esc`, except in `livedemo`.

| Command | What it shows | Keys |
| ------- | ------------- | ---- |
| `go -C examples run ./livedemo` | One generated signal. Resizes with the terminal and keeps 1000 points | `t` line width, `m` lines or braille, `f` tinted fill, `s` solid fill, `e` add an event, `Left` and `Right` move the cursor, `[` and `]` scroll back and forward, `Esc` hides the cursor and returns to live |
| `go -C examples run ./multidemo` | CPU, MEM and NET series | `t` line width, `m` lines or braille |
| `go -C examples run ./rxtxdemo` | Network RX and TX around a zero baseline | `w` line width, `s` smoothing, `m` lines or braille, `f` fill, `t` tinted fill, `z` zero baseline, `y` symmetric range |
| `go -C examples run ./cpulive` | CPU and RAM of this machine, read with gopsutil | `t` line width, `m` lines or braille |
| `go -C examples run ./donut` | Pods by phase as a ring and a frame, changing every second | `g` cycles braille, octants and half blocks |
| `go -C examples run ./cpudemo` | Prints one chart and exits | none |

## Screenshots

The PNG files in `assets/` come from `cmd/gen-shots`, which prints one chart per `-mode`, piped into freeze with its default config:

```bash
go run ./cmd/gen-shots -mode xaxis | env PATH=/usr/bin:/bin "$(command -v freeze)" -o assets/xaxis.png
```

- The modes are `hero`, `quickstart`, `styled`, `braille`, `rxtx`, `events`, `fill`, `zerocross`, `xaxis`, `xaxisat`, `autorange`, `thresholds`, `logscale`, `rightaxis`, `cursor`, `sparkline` and `donut`. Each one writes `assets/<mode>.png`, except `xaxisat`, which writes `assets/xaxis-at.png`.
- The data is generated with fixed random seeds, so a mode draws the same chart on every run.
- `env PATH=/usr/bin:/bin` hides `rsvg-convert` from freeze. When librsvg is installed, freeze converts the image with `rsvg-convert`, which ignores freeze's embedded monospace font and draws the chart in a proportional font.
- The embedded font of freeze has no braille glyphs, so the `braille`, `sparkline` and `donut` modes draw empty boxes with the command above. For these modes install a monospace font with the Braille block (for example JetBrainsMono Nerd Font Mono) and run freeze without the `PATH` override: `go run ./cmd/gen-shots -mode sparkline | freeze -o assets/sparkline.png --font.family "JetBrainsMono Nerd Font Mono"`. Then open the PNG and check the dots.
- `freeze --execute` hangs in a non-interactive shell; pipe the output in instead.
- For a new image, add a mode to `cmd/gen-shots/main.go`.

## Release

Pushing a `v*` tag runs `.github/workflows/release.yml`. GoReleaser builds `pulse-livedemo` and `pulse-multidemo` from the `examples` module for Linux, macOS and Windows (amd64 and arm64) and publishes them with a changelog on the GitHub release. The changelog leaves out `docs:` and `test:` commits.

Tag the merged commit on `main`: pull first and check `git log -1` before you tag. The Go module proxy keeps a published version, so a wrong tag cannot be moved; retract it in `go.mod` and tag a new version. That happened to v0.6.0, which was tagged on the v0.5.0 commit.

```bash
git switch main && git pull
git log -1 --oneline
git tag v0.8.0
git push origin v0.8.0
```

## Project Structure

| Path | Contents |
| ---- | -------- |
| `pulse.go` | `Model`, `New`, `Push`, `PushSeries`, the setters and getters |
| `options.go` | The `With...` options for `New` |
| `view.go` | `View`, Y ticks and labels |
| `renderer.go` | The `Renderer` interface, grid and fill shared by both modes |
| `renderer_lines.go` | `ModeLines`: box-drawing curves |
| `renderer_braille.go` | `ModeBraille`: braille dots with Catmull-Rom splines |
| `logscale.go` | `SetLogScale`, the log range fix-ups and the plot-space `Chart` for renderers |
| `rightaxis.go` | `Axis`, `WithRightAxis`, `SetSeriesRightAxis`, and the mapping of right-axis series into the plot |
| `threshold.go` | `Threshold`, `SetThresholds`, threshold lines and row styles |
| `cursor.go` | `SetCursor`, `SetOffset`, `CursorValues`, `CursorBox` |
| `sparkline.go` | `Sparkline`, `SparklineBraille` and `(*Model).Sparkline` |
| `event.go` | Timeline events, `EventsBox` |
| `legend.go` | `Legend`, `LegendBox` |
| `xaxis.go` | The time axis for `WithXAxis` and `WithXAxisAt` |
| `scale/` | `NiceTicks`, `NiceRange`, `LogTicks`, `LogRange`, tick and byte-rate formatters |
| `donut/` | The donut chart: `Model`, `Section`, glyph sets, the ring and frame geometry |
| `theme/` | The Tokyo Night preset and value colors |
| `examples/` | The demos, a separate module with its own `go.mod` |
| `cmd/gen-shots/` | The screenshot generator |
| `assets/` | Images used by the README and the guides |
| `docs/` | The guides |
