# TGCP UI Patterns & Style Guide

This document outlines the standard UI patterns, colors, and components used in TGCP to maintain a consistent and professional aesthetic.

## Color System

TGCP uses a semantic color system defined in `internal/styles/styles.go`. Avoid using raw hex codes in views; always use the exported Lipgloss colors.

| Semantic Name | Description | Use Case |
| :--- | :--- | :--- |
| `ColorBrandPrimary` | GCP Blue | Main titles, primary focus rings, key branding. |
| `ColorBrandAccent` | Light Blue | Selected items, active tabs, highlights. |
| `ColorTextPrimary` | Near White | Standard body text. |
| `ColorTextMuted` | Muted Grey | Secondary text, labels, inactive states. |
| `ColorSuccess` | Green | "RUNNING", "HEALTHY", success messages. |
| `ColorWarning` | Orange | "STOPPED", "PENDING", warning alerts. |
| `ColorError` | Red | "ERROR", "FAILED", critical alerts. |
| `ColorBorderSubtle` | Dark Grey | Panels borders, dividers. |

## Typography

-   **Headings**: Bold, `ColorBrandPrimary`. Use `TitleStyle`.
-   **Labels**: Bold, `ColorTextMuted`. Use `LabelStyle`.
-   **Values**: Regular, `ColorTextPrimary`. Use `ValueStyle`.

## Border Hierarchy

TGCP uses a two-tier border system to create visual hierarchy:

| Style | Use Case | Visual |
| :--- | :--- | :--- |
| `PrimaryBoxStyle` | Main content (detail cards, modals) | Rounded border, accent color (#75), more padding |
| `SecondaryBoxStyle` | Supporting content (hints, sections) | Normal border, subtle grey (#240), less padding |

```go
// Main content - prominent
styles.PrimaryBoxStyle.Render(mainContent)

// Supporting content - subtle
styles.SecondaryBoxStyle.Render(hints)
```

This ensures users can quickly identify the primary focus area vs. supporting information.

## Standard Components

### 1. Main Layout
The application follows a standard layout:
-   **Sidebar**: Left panel (25 chars width) listing services with semantic icons.
-   **Content Area**: Right panel taking up remaining space.
-   **Status Bar**: Bottom bar showing project, region, and service context.

### 2. Lists (Tables)
Resource lists should use `StandardTable` (`internal/ui/components/table.go`).
-   **Headers**: Bold, primary text, subtle background (#237).
-   **Selection (Focused)**: Dark grey background (#236), blue accent text (#39), bold.
-   **Selection (Blurred)**: Lighter grey background (#240), muted text (#245).
-   **Status Column**: Use `components.RenderStatus()` for badge-style indicators.

### 3. Detail Views
Use `DetailCard` (`internal/ui/components/detail.go`) for consistent styling.

```go
components.DetailCard(components.DetailCardOpts{
    Title: "Instance Details",
    Rows: []components.KeyValue{
        {Key: "Name", Value: instance.Name},
        {Key: "Status", Value: instance.Status}, // Auto-styled as badge
        {Key: "Zone", Value: instance.Zone},
    },
})
```

-   **Auto-Status Detection**: Fields named "Status" or "State" are automatically rendered as badges.
-   **Breadcrumbs**: Use `components.Breadcrumb()` - renders with `›` separator, muted path, bold current location.

### 4. Status Indicators
Use `components.RenderStatus()` for consistent status badges across all services.

```go
components.RenderStatus("RUNNING")  // ✓ RUNNING (green badge)
components.RenderStatus("STOPPED")  // ✗ STOPPED (red badge)
components.RenderStatus("PENDING")  // ◐ PENDING (yellow badge)
```

Recognized states:
-   **Running** (green): RUNNING, ACTIVE, HEALTHY, READY, RUNNABLE
-   **Stopped** (red): STOPPED, TERMINATED, FAILED, ERROR, DELETED
-   **Pending** (yellow): PENDING, STARTING, STOPPING, PROVISIONING, UPDATING
-   **Unknown** (grey): Any other state

### 5. Footer Hints (Keyboard Shortcuts)
Use `components.RenderFooterHint()` for styled keyboard hints.

```go
components.RenderFooterHint("s Start | x Stop | q Back")
// Renders as: [s] Start  [x] Stop  [q] Back
```

### 6. Filter Bar
The filter component (`components.FilterModel`) has three visual states:
-   **Inactive**: Dimmed placeholder with `/` hint
-   **Active**: Full input field with cursor
-   **Applied**: Badge showing filter count (e.g., `Filter: "prod" (3 of 10)`)

### 7. Overlays
-   **Command Palette**: Modal overlay centered on screen.
-   **Dialogs**: Use `components.RenderConfirmation()` for destructive actions.

### 8. Toast Notifications
Use `core.ToastMsg` to show temporary notifications for action feedback.

```go
// From a service action result:
return s, func() tea.Msg {
    return core.ToastMsg{
        Message: "Starting instance prod-web-1...",
        Type:    core.ToastSuccess,  // ToastSuccess, ToastError, ToastInfo
    }
}
```

Toast types:
-   **ToastSuccess** (green): Action completed successfully
-   **ToastError** (red): Action failed
-   **ToastInfo** (blue): Informational message

Toasts auto-dismiss after 3 seconds (default) or custom duration.

## Interaction Patterns

-   **Navigation**: `j/k` (Vim style) and `Arrow Keys` must both filter/navigate.
-   **Tabs**: `[` and `]` for switching internal tabs (e.g., Cloud Run Services <-> Jobs).
-   **Back**: `Esc` or `q` should always return to the previous context.
-   **Filters**: `/` should focus the filter input in list views.
