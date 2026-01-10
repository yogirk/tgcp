# TGCP Developer Guide

Welcome to the TGCP development guide. This document covers the architecture, setup, and steps to contribute new services.

## Architecture Overview

TGCP is built using [Bubble Tea](https://github.com/charmbracelet/bubbletea), a Go framework for terminal user interfaces (TUI), based on The Elm Architecture (TEA).

### Core Components

-   **Model**: The application state (`internal/ui/model.go`).
-   **Update**: The logic that handles messages and updates the model.
-   **View**: Renders the UI based on the current model.

### Directory Structure

```
tgcp/
├── cmd/tgcp/           # Main entry point
├── internal/
│   ├── core/           # Core logic (Auth, Caching, Navigation)
│   ├── services/       # Service implementations (GCE, GKE, etc.)
│   ├── ui/             # UI components and views
│   │   ├── components/ # Reusable UI widgets (Table, Sidebar, etc.)
│   │   └── styles/     # Lipgloss styles
│   └── utils/          # Helper functions
└── docs/               # Documentation
```

## Getting Started

1.  **Prerequisites**: Go 1.21+, `gcloud` SDK.
2.  **Run Locally**:
    ```bash
    go run ./cmd/tgcp --debug
    ```
    Debug logs will be written to `~/.tgcp/debug.log`.

## Adding a New Service

To add a new GCP service (e.g., `Cloud Spanner`):

1.  **Create Service Directory**:
    Create `internal/services/spanner/`.

2.  **Implement Service Interface**:
    Your service must implement the `Service` interface defined in `internal/services/interface.go`.
    
    ```go
    type Service interface {
        Name() string
        Init(projectID string) error
        Update(msg tea.Msg) (Service, tea.Cmd)
        View() string
        // ... see interface.go for full definition
    }
    ```

3.  **Register Service**:
    Add your new service to the initialization list in `internal/ui/model.go`.

## UI Component System

TGCP uses a set of standard components to ensure consistency. See `docs/ui_patterns.md` for detailed usage.

### Core Components
-   **StandardTable**: Use `components.NewStandardTable()` for resource lists with built-in focus/blur styling.
-   **DetailCard**: Use `components.DetailCard()` for detail views with auto-status detection.
-   **Breadcrumb**: Use `components.Breadcrumb()` for navigation paths.
-   **FilterModel**: Use `components.NewFilterWithPlaceholder()` with `FilterSession` for list filtering.

### Utility Functions
-   **RenderStatus()**: Renders status strings as colored badges (RUNNING=green, STOPPED=red, etc.)
-   **RenderFooterHint()**: Renders keyboard hints as `[key] Action` format.
-   **RenderSpinner()**: Loading indicator with message.
-   **RenderError()**: Standardized error display with suggestions.
-   **RenderConfirmation()**: Confirmation dialog for destructive actions.

### Styles
Always use styles from `internal/styles/styles.go` instead of defining custom Lipgloss styles.

## Coding Standards

-   **Error Handling**: Return errors explicitly. Use `ui.NewErrorMsg` to show user-facing errors.
-   **Concurrency**: Use `tea.Cmd` for all async operations (API calls). Never block the main UI thread.
-   **Styling**: Adhere to the `ui_patterns.md` guidelines.
