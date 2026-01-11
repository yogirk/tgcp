# Contributing to TGCP

First off, thanks for taking the time to contribute! 🎉

The following is a set of guidelines for contributing to TGCP. These are mostly guidelines, not rules. Use your best judgment, and feel free to propose changes to this document in a pull request.

## Quick Links

- **[Developer Guide](docs/DEVELOPER_GUIDE.md)** - Comprehensive guide for adding services, patterns, and debugging
- **[Features Guide](FEATURES.md)** - Overview of all features
- **[Codebase Review](CODEBASE_REVIEW.md)** - Architecture and design decisions
- **[UI Patterns](docs/ui_patterns.md)** - Standard UI layout rules for lists and detail views

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
- Keep the UI responsive. Avoid blocking the main Bubbletea update loop. Use `tea.Cmd` for long-running tasks.

## Project Structure

- `cmd/tgcp`: Main entry point.
- `internal/core`: Core logic (Auth, Caching, Event Bus, Service Registry).
- `internal/ui`: UI components and main Model.
- `internal/services`: Service-specific implementations (GCE, Cloud SQL, IAM, etc.).

## Adding a New Service

**See the [Developer Guide](docs/DEVELOPER_GUIDE.md) for detailed instructions.**

Quick steps:
1. Copy `internal/services/service_template.go.txt` as a starting point
2. Implement the `services.Service` interface
3. Register in `internal/ui/model.go`
4. Add navigation command and sidebar item

**Estimated time:** 2-4 hours for a basic service.

## Pull Requests

1.  Push your branch to GitHub.
2.  Open a Pull Request against the `main` branch.
3.  Describe your changes clearly.
4.  Ensure all tests pass.
5.  Follow the code style guidelines.

## Common Patterns

- Use shared components: `components.RenderError()`, `components.RenderSpinner()`, `components.RenderConfirmation()`
- Always use `StandardTable` for tables
- Follow `docs/ui_patterns.md` for breadcrumbs, filters, and detail layout decisions
- Implement proper caching with TTL
- Handle errors gracefully
- Keep UI responsive (use `tea.Cmd` for async operations)

For more details, see the [Developer Guide](docs/DEVELOPER_GUIDE.md).

## License

By contributing, you agree that your contributions will be licensed under its MIT License.
