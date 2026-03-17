# Contributing to llmtop

Thank you for your interest in contributing! llmtop is built by and for the LLM inference community.

## Getting Started

### Prerequisites

- Go 1.22+
- A terminal that supports 256 colors (most modern terminals do)
- Optional: a running vLLM, SGLang, or LMCache instance for real testing

### Development Setup

```bash
# Clone the repo
git clone https://github.com/InfraWhisperer/llmtop.git
cd llmtop

# Download dependencies
go mod download

# Build
make build

# Run
./llmtop --endpoint http://localhost:8000

# Run tests
make test
```

### Project Structure

```
cmd/llmtop/     — CLI entry point (cobra)
internal/
  collector/    — Concurrent metrics polling
  discovery/    — Auto-discovery of local workers
  metrics/      — Prometheus parser + data models
  ui/           — Bubbletea TUI
pkg/config/     — Config file parsing
```

## Adding a New Backend

1. Add constants to `internal/metrics/models.go`
2. Add a parser function `internal/collector/<backend>.go`
3. Wire it up in `internal/collector/collector.go` `parseWorkerMetrics()`
4. Update the detection logic in `detectBackendAndModel()`
5. Add tests in `*_test.go` files
6. Update README.md backend support table

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Run `golangci-lint run` before submitting
- All exported types and functions must have doc comments
- Prefer explicit error handling over panics

## Pull Request Process

1. Fork the repo and create a feature branch
2. Write tests for new functionality
3. Run `make test` and `make lint`
4. Open a PR with a clear description of the change
5. Link any related issues

## Reporting Issues

Please include:
- Your OS and terminal emulator
- Go version (`go version`)
- llmtop version (`llmtop version`)
- The command you ran
- Expected vs. actual behavior
- Any error messages

## Feature Requests

Open an issue with the `enhancement` label. Describe:
- The use case / problem you're solving
- What backend(s) it applies to
- Any mockups or examples

## License

By contributing, you agree that your contributions will be licensed under the Apache 2.0 License.
