# TGCP Agent Guide

This document is intended to help AI agents understand the TGCP codebase, its context, and common tasks.

## Project Context

**TGCP** is a Terminal User Interface (TUI) for Google Cloud Platform. It is written in Go using the Bubble Tea framework. It aims to provide a fast, keyboard-driven alternative to the GCP Console for observability and basic management tasks.

## Key Files & Directories

-   **`cmd/tgcp/main.go`**: Entry point. Handles flag parsing and starts the Bubble Tea program.
-   **`internal/ui/model.go`**: The main application state (`MainModel`). This is the "brain" of the TUI.
-   **`internal/services/`**: Contains subdirectories for each GCP service (e.g., `gce`, `gke`).
    -   Each service MUST implement the `Service` interface (`internal/services/interface.go`).
-   **`internal/styles/styles.go`**: The central source of truth for all UI styling.
-   **`docs/spec.md`**: The original Product Requirements Document (PRD).

## Architectural Guidelines

1.  **Immutability**: The Bubble Tea update loop is functional. State changes return a *new* model.
2.  **Async Operations**: All API calls (fetching instances, starting VMs) MUST be done via formatting a `tea.Cmd`. **NEVER** block the `Update` function.
3.  **Service Isolation**: Service implementations should not know about each other. They interact through the main model or shared utility messages.
4.  **Error Handling**: Use `ui.NewErrorMsg` to send error messages to the UI to be displayed as toast notifications.

## Common Tasks for Agents

### Adding a New Service
1.  Read `internal/services/interface.go` to understand the contract.
2.  Create a new directory `internal/services/<service_name>`.
3.  Implement `Init`, `Update`, `View`, `Name`, etc.
4.  Add the service to the `initialServices` map in `internal/ui/model.go`.

### Fixing a Bug in a View
1.  Identify the component in `internal/ui/components/` or the specific service view in `internal/services/<service_type>/view.go`.
2.  Check for logic errors in the `Update` function or rendering issues in `View`.

### Polishing UI
1.  Consult `docs/ui_patterns.md`.
2.  Use `internal/styles/styles.go` definitions.
3.  Ensure layout responsiveness (use `tea.WindowSizeMsg`).

## Important Constraints
-   **No "Coming Soon"**: If a feature is claimed, it must be implemented (even if basic).
-   **Zero Config**: The app should work out-of-the-box with ADC (`gcloud auth application-default login`).
-   **Performance**: Large lists should be virtualized or paginated (standard table component handles this).
