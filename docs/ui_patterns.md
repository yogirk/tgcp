# TGCP UI Patterns & Style Guide

This document outlines the standard UI patterns, colors, and components used in TGCP to maintain a consistent and professional aesthetic.

## Color System

TGCP uses a semantic color system defined in `internal/styles/styles.go`. Avoid using raw hex codes in views; always use the exported Lipgloss colors.

### Brand & Semantic

| Semantic Name | Description | Use Case |
| :--- | :--- | :--- |
| `ColorBrandPrimary` | GCP Blue | Main titles, primary focus rings, key branding. |
| `ColorBrandAccent` | Light Blue | Selected items, active tabs, highlights. |
| `ColorTextPrimary` | Near White | Standard body text. |
| `ColorTextMuted` | Muted Grey | Secondary text, labels, inactive states. |
| `ColorSuccess` | Green | "RUNNING", "HEALTHY", success messages. |
| `ColorWarning` | Orange | Genuine warnings (reserved — do not use for mode badges). |
| `ColorError` | Red | "ERROR", "FAILED", critical alerts. |
| `ColorInfo` | Cyan | Informational mode badges (FILTER), info toasts. |
| `ColorBorderSubtle` | Dark Grey | Panel borders, dividers. |

### Accent Palette (lifted from the banner)

| Semantic Name | Description | Use Case |
| :--- | :--- | :--- |
| `ColorAccentRed` | Google Red | Databases / Storage category tint. |
| `ColorAccentYellow` | Google Yellow | Compute / Security category tint. |
| `ColorAccentGreen` | Google Green | DevOps category tint. |

Categories map to accents in `internal/ui/components/category.go`. Use `CategoryColor(name)` or `ServiceAccent(shortName)` to look up a tint — do not hardcode.

### Surfaces

| Semantic Name | Description | Use Case |
| :--- | :--- | :--- |
| `ColorSurfaceBar` | #235 | Status bar background. |
| `ColorSurfaceHeader` | #237 | Header-block background (page titles). |
| `ColorTextOnBar` | #246 | Status bar foreground (≥4.5:1 over `ColorSurfaceBar`). |

## Typography

TGCP uses a three-tier header scale — one rule per tier:

| Style | Tier | Use Case |
| :--- | :--- | :--- |
| `HeaderStyle` | Page | Solid bar with background fill. Page-level titles, detail-card header bar. |
| `SectionStyle` | Section | Bold, brand accent. Card titles, in-box section headings. `TitleStyle` is an alias. |
| `GroupStyle` | Group | Muted uppercase dividers for list groups and categories. |

Body text uses `LabelStyle` (muted, bold) for keys and `ValueStyle` (primary text) for values.

## Border Hierarchy

TGCP uses a three-tier border system:

| Style | Use Case | Visual |
| :--- | :--- | :--- |
| `PrimaryBoxStyle` | Main content cards, active panels | Rounded border, brand accent (#75) |
| `SecondaryBoxStyle` | Supporting content (metadata sections) | Normal border, subtle grey (#240) |
| `OverlayBoxStyle` | Modals, dialogs, dropdowns | Rounded border, caller-supplied accent colour |

```go
// Main content — prominent
styles.PrimaryBoxStyle.Render(mainContent)

// Supporting content — subtle
styles.SecondaryBoxStyle.Render(metadata)

// Overlay — caller recolors the border per action
styles.OverlayBoxStyle.Copy().BorderForeground(styles.ColorError).Render(errorDialog)
```

## Spacing Scale

Reference `styles.SpaceXS` / `SpaceS` / `SpaceM` / `SpaceL` (0, 1, 2, 4) instead of raw padding literals. Keeps rhythm consistent across screens.

## Selection States

Two canonical states for any list row (sidebar, home menu, command palette):

| Style | When |
| :--- | :--- |
| `SelectedActive` | Row is selected AND its list has focus. Bold brand accent, left border bar. |
| `SelectedBlur` | Row is selected but focus is elsewhere. Brand accent foreground, no bar. |

Do not hand-roll selection styling per component — consistency comes from these two tokens.

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
-   **Summary Pills**: Render `components.StatusSummary(states)` above the table to give readers an at-a-glance breakdown before they scan rows. Empty categories are omitted; total is always shown.
-   **Empty Row Sets**: When the list is empty, render `components.EmptyState("<resource-type>")` instead of leaving the area blank.

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
    FooterHint: "s Start | x Stop | q Back",
})
```

**Visual structure:**
```
 Instance Details                    ← Header bar (background + bold)
╭──────────────────────────────────╮
│ Name:     my-instance            │ ← Primary box with key-value rows
│ Status:   ✓ RUNNING              │
╰──────────────────────────────────╯
[s] Start  [x] Stop  [q] Back       ← Footer hints (styled keys)
```

-   **Header Bar**: Title rendered with background (matches table headers).
-   **Auto-Status Detection**: Fields named "Status" or "State" are automatically rendered as badges.
-   **Footer Hints**: Use `FooterHint` option for keyboard shortcuts.
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

### 9. Status Summary Pills
Sits above a list view to give readers an at-a-glance breakdown of status distribution.

```go
states := make([]string, 0, len(instances))
for _, i := range instances {
    states = append(states, string(i.State))
}
summary := components.StatusSummary(states)
// Renders: ✓ 4 Running  ·  ◐ 1 Pending  ·  ✗ 1 Stopped  ·  6 total
```

Pills reuse the colour system from `RenderStatus()` so per-row badges and the summary stay in visual lockstep.

### 10. Empty States
Render a subtle italic one-liner instead of a blank box when a list has zero rows.

```go
if len(instances) == 0 {
    return components.EmptyState("instances")
}
```

Message pools live in `internal/ui/components/empty.go`, keyed by resource type (`"instances"`, `"buckets"`, `"clusters"`, `"databases"`, `"logs"`, `"recommendations"`, `"budgets"`, ...). Unknown keys fall back to the `"default"` pool. Messages rotate deterministically by hour-of-day so the copy varies day-to-day without changing within a session.

## Landing Page (Home Menu)

The landing page displays a playful ASCII banner, user/project context, and a fuzzy-filterable service menu.

### Current Layout: Fuzzy-Filterable Flat List

```
          ████ TGCP ████                   ← ASCII banner (brand moment)
   👤 rk@...   ·   📁 cloudside-academy    ← Muted inline identity line

┌─ Services ─────────────────────────────┐
│ / Filter...                            │  ← Fuzzy filter bar
│                                        │
│   Overview (Command Center)            │  ← Top item (selected row = accent bar)
│                                        │
│ COMPUTE                                │  ← Category header (category colour)
│   ⚙  Compute Engine (GCE)              │  ← Service icons tinted by category
│   ☸  Kubernetes Engine (GKE)           │
│   ▷  Cloud Run                         │
│ STORAGE                                │
│   ▤  Cloud Storage (GCS)               │
│   ◔  Disks                             │
│ DATABASES                              │
│   ⛁  Cloud SQL                         │
│   ...                                  │
│ DEVOPS                                 │
│   ◈  Cloud Build                       │
│   ▣  Artifact Registry                 │
└────────────────────────────────────────┘
  ▼ 8 more                                ← Scroll indicator

↑/↓ navigate   / filter   Enter select   q clear filter   ? help
```

**Visual notes:**
- The identity line is rendered inline (no box) so the banner and the menu breathe.
- Category headers use `GroupStyle` tinted with `CategoryColor(name)`.
- Service icons are tinted with `ServiceAccent(shortName)` — the label text stays neutral for readability.
- The selected row uses `SelectedActive` (left accent bar); no `▸` prefix.

**Structure:**
- **Filter Bar**: Fuzzy search input at top, activated with `/`
- **Top Item**: Overview sits above categories as the primary dashboard entry point
- **Category Headers**: Visual grouping labels (Compute, Storage, Databases, Data & Analytics, Security & Networking, Observability, DevOps) — not selectable
- **Services**: Flat list under category headers, all visible (scrollable)
- **Scroll Indicators**: Show when more items exist above/below viewport

**Categories (8):**
Compute, Storage, Databases, Data & Analytics, Security & Networking, Observability, DevOps

**Navigation:**
- `↑/↓` or `j/k`: Move through service items
- `/`: Activate fuzzy filter — type to narrow the list
- `Enter`: Navigate to selected service
- `q` or `Esc`: Clear active filter

## Interaction Patterns

-   **Navigation**: `j/k` (Vim style) and `Arrow Keys` must both filter/navigate.
-   **Tabs**: `[` and `]` for switching internal tabs (e.g., Cloud Run Services <-> Jobs).
-   **Back**: `Esc` or `q` should always return to the previous context.
-   **Filters**: `/` should focus the filter input in list views.
