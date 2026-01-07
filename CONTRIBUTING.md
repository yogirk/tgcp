# Contributing to TGCP

First off, thanks for taking the time to contribute! 🎉

The following is a set of guidelines for contributing to TGCP. These are mostly guidelines, not rules. Use your best judgment, and feel free to propose changes to this document in a pull request.

## Getting Started

1.  **Fork the repository** on GitHub.
2.  **Clone your fork** locally:
    ```bash
    git clone https://github.com/yourusername/tgcp.git
    cd tgcp
    ```
3.  **Install dependencies**:
    ```bash
    go mod download
    ```
4.  **Create a branch** for your feature or fix:
    ```bash
    git checkout -b feature/amazing-feature
    ```

## Development Workflow

- **Run Locally**: `go run ./cmd/tgcp`
- **Debug Mode**: `go run ./cmd/tgcp --debug` (Output: `~/.tgcp/debug.log`)
- **Format Code**: `go fmt ./...`
- **Run Tests**: `go test ./...`

## Code Style

- We follow standard Go conventions.
- Use `gofmt` to format your code.
- Ensure your code passes `go vet`.
- Keep the UI responsive. Avoid blocking the main Bubbletea update loop. use `tea.Cmd` for long-running tasks.

## Project Structure

- `cmd/tgcp`: Main entry point.
- `internal/core`: Core logic (Auth, Caching, Event Bus).
- `internal/ui`: UI components and main Model.
- `internal/services`: Service-specific implementations (GCE, Cloud SQL, IAM).

## Pull Requests

1.  Push your branch to GitHub.
2.  Open a Pull Request against the `main` branch.
3.  Describe your changes clearly.
4.  Ensure all tests pass.

## License

By contributing, you agree that your contributions will be licensed under its MIT License.
