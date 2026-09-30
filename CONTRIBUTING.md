# Contributing to Pulse

Thank you for your interest in contributing to **Pulse**! We welcome contributions, bug reports, feature proposals, and pull requests.

## Development Setup

Pulse requires Go 1.22 or newer.

1. Clone the repository:
   ```bash
   git clone https://github.com/ingvarch/pulse.git
   cd pulse
   ```

2. Run tests to ensure everything is working:
   ```bash
   go test -v ./...
   ```

3. Try out the live interactive demo:
   ```bash
   go run ./examples/livedemo
   ```

## Pull Request Guidelines

1. **Keep it focused**: One bugfix or feature per PR.
2. **Add tests**: Include unit tests for any new behavior or bugfix.
3. **Follow Go style**:
   - Format code using standard formatting: `gofmt -s -w .`
   - Run vet: `go vet ./...`
   - Run test suite with race detector: `go test -race ./...`
4. **Preserve documentation**: Document all exported functions, methods, and types with clean docstrings in English according to [Go Doc Conventions](https://go.dev/blog/godoc).
5. **Commit style**: We follow [Conventional Commits](https://www.conventionalcommits.org/) (e.g. `feat: ...`, `fix: ...`, `docs: ...`, `refactor: ...`).

## Reporting Bugs

Please open an issue on GitHub with:
- A clear description of the issue.
- Your terminal emulator and operating system.
- A minimal reproducible code snippet.
- Expected vs actual output (terminal screenshots or ASCII text are appreciated).

## Suggesting Enhancements

Have an idea for a new feature (e.g., custom rendering styles, new chart components, scale improvements)? Open an issue describing:
- What problem you are trying to solve.
- How you envision the API looking.
- Any alternative solutions considered.
