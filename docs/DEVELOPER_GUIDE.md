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
3.  **Run against fixtures (no GCP auth)**:
    ```bash
    go run ./cmd/tgcp --demo
    ```
    The `--demo` flag synthesizes an auth state and makes every service's API client short-circuit to embedded JSON fixtures under `internal/demo/data/`. Use it for fast UI iteration, generating the `demo.gif`, or powering the website's landing-page hero. Keep this flag out of end-user documentation — it is intentionally not advertised.

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

3.  **Register Service** (5 locations required):

    **a. Service Registry** (`internal/ui/model.go` → `registerAllServices()`):
    ```go
    registry.Register("spanner", func(cache *core.Cache) services.Service {
        return spanner.NewService(cache)
    })
    ```

    **b. Landing Screen** (`internal/ui/components/home_menu.go` → `NewHomeMenu()`):
    Add to the appropriate category in the `Categories` slice:
    ```go
    {Name: "Spanner", ShortName: "spanner"},
    ```
    Categories: Compute, Storage, Databases, Data & Analytics, Security & Networking, Observability, DevOps

    **c. Sidebar** (`internal/ui/components/sidebar.go` → `Items` slice):
    Add in category order with a Unicode icon (see Icon Guidelines below):
    ```go
    {Name: "Spanner", ShortName: "spanner", Icon: "⬡"},
    ```

    **d. Group Breaks** (if needed): Update `groupBreaks` map in `sidebar.go` if adding to a new category position.

    **e. Category Mapping** (`internal/ui/components/category.go` → `serviceCategory`):
    Map the service short-name to its category so its icon inherits the right accent colour:
    ```go
    "spanner": catDatabase,
    ```

4.  **Wire demo mode** (recommended, even if you don't author fixtures yet):
    In the service's `api.go`, gate `NewClient` and every `List*`/`Get*` method with an early return when `demo.Enabled` is true, returning empty typed results. This keeps `./tgcp --demo` from hitting the real GCP API with the synthetic project ID. To add real fixture data, drop a JSON file at `internal/demo/data/<service>.json` and load it via `demo.MustLoad()` in a small loader (see `internal/services/gce/demo.go` for the pattern).

## UI Component System

TGCP uses a set of standard components to ensure consistency. See `docs/ui_patterns.md` for detailed usage.

### Core Components
-   **StandardTable**: Use `components.NewStandardTable()` for resource lists with built-in focus/blur styling.
-   **DetailCard**: Use `components.DetailCard()` for detail views with auto-status detection.
-   **Breadcrumb**: Use `components.Breadcrumb()` for navigation paths.
-   **FilterModel**: Use `components.NewFilterWithPlaceholder()` with `FilterSession` for list filtering.

### Utility Functions
-   **RenderStatus()**: Renders status strings as coloured badges (RUNNING=green, STOPPED=red, etc.)
-   **StatusSummary()**: Renders a count-pill summary above a list — `✓ 4 Running · ✗ 1 Stopped · 5 total`. Pass a slice of state strings; empty statuses are filtered out.
-   **EmptyState()**: Playful italic one-liner for zero-row views. Pass a resource-type key (`"instances"`, `"buckets"`, `"clusters"`, …); unknown keys fall back to the `"default"` pool.
-   **RenderFooterHint()**: Renders keyboard hints as `[key] Action` format.
-   **InlineLoader()**: Inline, single-line loading indicator for embedding in detail cards while a sub-field loads. Use `SpinnerModel` for full-page loading states.
-   **RenderError()**: Standardised error display with suggestions.
-   **RenderConfirmation()**: Confirmation dialog for destructive actions.
-   **CategoryColor() / ServiceAccent()**: Map a category name or a service short-name to its accent colour — used to tint service icons and category headers.

### Toast Notifications
Use `core.ToastMsg` to provide action feedback:
```go
return s, func() tea.Msg {
    return core.ToastMsg{Message: "Instance started", Type: core.ToastSuccess}
}
```

### Styles
Always use styles from `internal/styles/styles.go` instead of defining custom Lipgloss styles.

**Border Hierarchy (three tiers):**
-   `PrimaryBoxStyle`: Main content cards, active panels (rounded border, brand accent)
-   `SecondaryBoxStyle`: Supporting content, metadata sections (normal border, subtle grey)
-   `OverlayBoxStyle`: Modals, dropdowns, confirmation dialogs (rounded border, caller-supplied accent)

**Typography Hierarchy (three tiers):**
-   `HeaderStyle`: Page-level titles rendered as a solid bar (with background fill).
-   `SectionStyle`: Card titles and in-box section headings (bold, brand accent).
-   `GroupStyle`: Muted uppercase dividers for list groups and categories.

**Selection (two canonical states):**
-   `SelectedActive`: Row is selected AND its list has focus (bold, brand accent, left border bar).
-   `SelectedBlur`: Row is selected but focus is elsewhere (brand accent, no bar).

**Spacing Scale:**
Use `styles.SpaceXS / SpaceS / SpaceM / SpaceL` (0, 1, 2, 4) instead of raw padding literals.

## Icon Guidelines

### Sidebar Service Icons

Service icons in the sidebar (`internal/ui/components/sidebar.go`) **must use Unicode symbols, NOT emojis**.

**Allowed:** Unicode geometric shapes, arrows, and miscellaneous symbols:
```
◉ ⚙ ☸ ▷ ▤ ◔ ⛁ ⬡ ▦ ◇ ◲ ⊞ ⇢ ⎈ ⇌ ⚿ ✦ ⇄
```

**Not allowed:** Emojis (e.g., 🔐 🖥️ 💾)

**Why:** Unicode symbols render consistently across terminals and themes, while emojis may vary in appearance, width, and color rendering. Sidebar icons should be monochromatic and uniform.

**Finding icons:** Use Unicode blocks like:
- Geometric Shapes (U+25A0–U+25FF): `◉ ◇ ◈ ▤ ▦ ◲`
- Arrows (U+2190–U+21FF): `⇢ ⇌ ⇄`
- Miscellaneous Symbols (U+2600–U+26FF): `⚙ ⚿ ⛁`
- Miscellaneous Technical (U+2300–U+23FF): `⎈`

### Dashboard/Content Icons

Emojis are acceptable in dashboard content views (like `overview/views.go`) and the landing-page header — these surfaces are authored, not reflections of GCP data. Prefer consistency within each view.

### Category Tinting

Service icons in both the sidebar and the home menu are tinted by category (Compute=yellow, Storage/Databases=red, Data & Analytics=cyan, Security/Networking=yellow, Observability=cyan, DevOps=green). The mapping lives in `internal/ui/components/category.go`. When you add a new service, register it in `serviceCategory` there so its icon inherits the right colour.

## Coding Standards

-   **Error Handling**: Return errors explicitly. Use `ui.NewErrorMsg` to show user-facing errors.
-   **Concurrency**: Use `tea.Cmd` for all async operations (API calls). Never block the main UI thread.
-   **Styling**: Adhere to the `ui_patterns.md` guidelines.
