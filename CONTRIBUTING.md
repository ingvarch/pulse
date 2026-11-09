# Contributing to Pulse

Bug reports, feature ideas and pull requests are welcome.

## Reporting Bugs

Open a [bug report](https://github.com/ingvarch/pulse/issues/new?template=bug_report.yml) with:
- A clear description of the issue.
- Your terminal emulator, its font and your operating system.
- The Pulse and Go versions.
- A minimal code snippet that reproduces it.
- Expected and actual output: a screenshot or the text of the chart.

Report security issues privately, as [SECURITY.md](SECURITY.md) describes.

## Suggesting Enhancements

Open a [feature request](https://github.com/ingvarch/pulse/issues/new?template=feature_request.yml) that describes:
- The problem you are trying to solve.
- How you see the API.
- Any alternatives you considered.

## Development Setup

Setup, checks, demos and screenshots are in [docs/development.md](docs/development.md).

## Making a Change

1. Create a branch named `<type>/<short-name>`, for example `feat/xaxis-at` or `fix/gap-fill`.
2. Write a failing test for the new behavior or the bug first, then the code that makes it pass.
3. Document every exported function, method and type in English, following the [Go doc conventions](https://go.dev/doc/comment).
4. Update the guide in `docs/` that describes the behavior, and its screenshot when the picture changes.
5. Run the checks from [docs/development.md](docs/development.md#checks): `gofmt`, `go vet`, `go test -race` and `go build`, in the root and in the `examples` module.
6. Write commits as [Conventional Commits](https://www.conventionalcommits.org/) with the `pulse` scope, for example `feat(pulse): add WithXAxisAt` or `docs(pulse): fix the Bubble Tea example`.
7. Open a pull request that fills in the template and closes its issue with `Closes #<number>`. Keep it to one fix or one feature.

## Code of Conduct

This project follows the [Code of Conduct](CODE_OF_CONDUCT.md). By taking part you agree to it.

## License

By contributing, you agree that your contributions are licensed under the [MIT License](LICENSE).
