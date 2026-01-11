# TGCP - Terminal UI for Google Cloud Platform
## Product Requirements Document (PRD)

**Version:** 0.1
**Last Updated:** 2026-01-06
**Status:** Initial Specification

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Project Vision & Goals](#project-vision--goals)
3. [Target Audience](#target-audience)
4. [Scope & Boundaries](#scope--boundaries)
5. [Technical Architecture](#technical-architecture)
6. [User Experience & Interface Design](#user-experience--interface-design)
7. [Feature Specifications](#feature-specifications)
8. [Service Modules](#service-modules)
9. [Authentication & Authorization](#authentication--authorization)
10. [Error Handling & Resilience](#error-handling--resilience)
11. [Performance & Scalability](#performance--scalability)
12. [Configuration & Customization](#configuration--customization)
13. [Development Roadmap](#development-roadmap)
14. [Technical Stack](#technical-stack)
15. [Success Metrics](#success-metrics)
16. [Future Enhancements](#future-enhancements)

---

## Executive Summary

**TGCP** (Terminal UI for Google Cloud Platform) is a terminal-based user interface for observing and managing GCP resources, inspired by K9s (Kubernetes TUI) and TAWS (Terminal AWS). The tool prioritizes read-heavy observability workflows for DevOps engineers working with small-to-medium GCP environments.

**Core Value Proposition:**
- Fast, keyboard-driven navigation of GCP resources
- Eliminate context-switching to web console for routine observability tasks
- Beautiful, minimalist terminal interface with K9s-inspired aesthetics
- Modular architecture supporting incremental service additions

**Key Differentiators:**
- Terminal-native experience (no browser required)
- Focused on observability over infrastructure provisioning
- Single binary distribution with zero configuration required
- Project-scoped view (not cross-project aggregation)

---

## Project Vision & Goals

### Vision Statement
Empower DevOps engineers to observe and interact with GCP resources entirely from the terminal, reducing cognitive overhead and improving operational efficiency.

### Primary Goals
1. **Observability First:** Provide read-heavy views of GCP resources with instant access to metadata, status, and logs
2. **Developer Delight:** Create a beautiful, intuitive TUI that feels native to terminal power users
3. **Extensibility:** Design a modular architecture where adding new GCP services doesn't break existing functionality
4. **Safety:** Prevent destructive operations while allowing safe administrative actions (start/stop, SSH, log viewing)

### Non-Goals
1. **Not an IaC replacement:** No resource creation or deletion (use Terraform/Pulumi for that)
2. **Not for massive multi-project environments:** Optimized for single-project views, not fleet-wide orchestration
3. **Not a billing management tool:** (though billing summary is a future enhancement)
4. **Not a complete GCP Console replacement:** Focused on common DevOps workflows, not every GCP feature

---

## Target Audience

### Primary Persona: DevOps Engineer
- **Role:** SRE, Platform Engineer, Backend Developer with ops responsibilities
- **Environment:** Small-to-medium GCP projects (not enterprise-scale multi-org setups)
- **Workflows:**
  - Check VM status before deployments
  - Tail logs from Cloud Run services
  - Verify IAM permissions for service accounts
  - Monitor Cloud SQL instance health
  - SSH into instances for debugging
- **Pain Points:**
  - GCP Console is slow and requires browser context-switching
  - `gcloud` CLI commands are verbose and lack interactive exploration
  - Existing tools (K9s, TAWS) don't support GCP

### Secondary Persona: Backend Developer
- **Use Case:** Quick resource lookups without leaving terminal workflow
- **Needs:** Fast answers to "Is my Cloud Run service up?" or "What's the instance IP?"

---

## Scope & Boundaries

### In Scope (MVP - Phase 1)

#### Services (Initial 3 for MVP validation)
1. **Google Compute Engine (GCE)**
   - List instances with [Name | Zone | Status | Internal IP | External IP | Tags]
   - View instance metadata
   - Start/Stop instances
   - SSH into instances (external session, not embedded)

2. **Cloud SQL**
   - List instances with [Name | Region | Database Version | Status | Connection Name]
   - View instance configuration
   - Start/Stop instances (for development environments)

3. **Identity & Access Management (IAM)**
   - List service accounts
   - List IAM users
   - View role bindings (read-only)

#### Core Features
- Application Default Credentials (ADC) authentication
- Project selection and switching
- Command palette (`:` trigger) with fuzzy search
- Service sidebar navigation
- Table views with inline filtering (`/` trigger)
- Pagination (50 items per page)
- Basic error handling with toast notifications
- Version update checking on launch
- Debug logging (`--debug` flag)

### In Scope (Phase 2 - Post-MVP)

#### Additional Services
4. **Cloud Run**
   - List services with [Name | Region | URL | Revision]
   - View service metadata
   - Restart services

5. **Cloud Storage (GCS)**
   - List buckets
   - Browse objects (read-only)

6. **BigQuery**
   - List datasets
   - List tables
   - View table schemas (read-only)

7. **Networking**
   - List VPCs
   - List firewall rules (read-only)

8. **Cloud Functions**
   - List functions
   - View function metadata

#### Advanced Features
- Log tailing integration (Cloud Logging API streaming)
- IAP tunnel support for SSH to private instances
- Cross-project switching (within same organization)
- Custom keybinding configuration (`~/.tgcprc`)
- API polling interval customization

### Out of Scope

#### Operations NOT Supported
- ❌ Resource deletion (destructive operations)
- ❌ Resource creation (use IaC tools)
- ❌ Billing account administration
- ❌ Organization-level IAM management
- ❌ Cross-project resource aggregation
- ❌ Real-time alerting/notifications
- ❌ Write operations on BigQuery (no query execution)

#### Future Considerations (Not Committed)
- AI-powered insights ("Instance high CPU for 7 days")
- Multi-pane views (split screens)
- Embedded SSH terminal (vs. external session)
- Billing cost summaries
- Resource dependency graphs

---

## Technical Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        TGCP (Main)                           │
│  ┌────────────────────────────────────────────────────────┐ │
│  │         Bubbletea Application (TUI Framework)          │ │
│  └────────────────────────────────────────────────────────┘ │
│                              │                               │
│         ┌────────────────────┼────────────────────┐         │
│         ▼                    ▼                    ▼         │
│  ┌─────────────┐      ┌─────────────┐     ┌─────────────┐  │
│  │ Navigation  │      │  UI Layer   │     │   Cache     │  │
│  │  - Palette  │      │ - Tables    │     │  - Local    │  │
│  │  - Sidebar  │      │ - Forms     │     │  - Refresh  │  │
│  │  - Router   │      │ - Toasts    │     │  - TTL      │  │
│  └─────────────┘      └─────────────┘     └─────────────┘  │
│                              │                               │
│                    ┌─────────┴─────────┐                    │
│                    ▼                   ▼                    │
│            ┌───────────────┐   ┌──────────────┐            │
│            │ Service Layer │   │  API Client  │            │
│            │  (Interface)  │   │  - Auth      │            │
│            └───────────────┘   │  - Rate Limit│            │
│                    │            └──────────────┘            │
│       ┌────────────┼────────────┐                           │
│       ▼            ▼            ▼                           │
│  ┌────────┐  ┌─────────┐  ┌─────────┐                      │
│  │  GCE   │  │CloudSQL │  │   IAM   │  ... (more services) │
│  │ Module │  │ Module  │  │ Module  │                      │
│  └────────┘  └─────────┘  └─────────┘                      │
│       │            │            │                           │
│       └────────────┼────────────┘                           │
│                    ▼                                        │
│         ┌─────────────────────┐                             │
│         │   GCP APIs          │                             │
│         │ - compute.v1        │                             │
│         │ - sqladmin.v1       │                             │
│         │ - iam.v1            │                             │
│         └─────────────────────┘                             │
└─────────────────────────────────────────────────────────────┘
```

### Directory Structure

```
tgcp/
├── main.go                    # Entry point, Bubbletea app initialization
├── go.mod
├── go.sum
├── README.md
├── LICENSE
│
├── cmd/
│   └── tgcp/
│       └── main.go            # CLI flag parsing (--debug, --version)
│
├── internal/
│   ├── core/
│   │   ├── auth.go            # ADC authentication, token management
│   │   ├── cache.go           # Local caching with TTL
│   │   ├── client.go          # GCP API client wrapper (rate limiting)
│   │   ├── navigation.go      # Command palette, routing logic
│   │   └── config.go          # Future: ~/.tgcprc parsing
│   │
│   ├── ui/
│   │   ├── components/
│   │   │   ├── table.go       # Reusable table component
│   │   │   ├── toast.go       # Toast notification component
│   │   │   ├── statusbar.go   # Top metadata bar
│   │   │   ├── sidebar.go     # Left service sidebar
│   │   │   └── filter.go      # Inline filter input
│   │   ├── home.go            # Home screen (ASCII banner + metadata)
│   │   ├── help.go            # Help view
│   │   └── styles.go          # Color schemes, styles
│   │
│   ├── services/
│   │   ├── interface.go       # Service interface definition
│   │   │
│   │   ├── gce/
│   │   │   ├── gce.go         # GCE service implementation
│   │   │   ├── views.go       # Instance list, detail views
│   │   │   ├── actions.go     # Start, stop, SSH logic
│   │   │   └── api.go         # GCE API calls
│   │   │
│   │   ├── cloudsql/
│   │   │   ├── cloudsql.go
│   │   │   ├── views.go
│   │   │   ├── actions.go
│   │   │   └── api.go
│   │   │
│   │   ├── iam/
│   │   │   ├── iam.go
│   │   │   ├── views.go
│   │   │   └── api.go
│   │   │
│   │   └── ... (future services: cloudrun, gcs, bigquery, etc.)
│   │
│   └── utils/
│       ├── logger.go          # Debug logging to ~/.tgcp/debug.log
│       └── version.go         # Version checking logic
│
└── assets/
    └── banner.txt             # ASCII art for home screen
```

### Service Interface Design

All GCP service modules implement this common interface:

```go
package services

import "github.com/charmbracelet/bubbletea"

type Service interface {
    // Name returns the service display name (e.g., "Google Compute Engine")
    Name() string

    // ShortName returns the CLI shorthand (e.g., "gce")
    ShortName() string

    // Init initializes the service (setup API clients)
    Init(projectID string) error

    // List returns the main resource listing view (Bubbletea Model)
    List() tea.Model

    // Detail returns a detail view for a specific resource
    Detail(resourceID string) tea.Model

    // Actions returns available actions for this service
    // (e.g., ["start", "stop", "ssh"] for GCE)
    Actions() []Action

    // Refresh fetches latest data from GCP API
    Refresh() error

    // SupportsFilter indicates if service supports inline filtering
    SupportsFilter() bool
}

type Action struct {
    Name        string
    Description string
    Keybinding  string
    Handler     func(resourceID string) error
}
```

**Why this works:**
- Each service is fully encapsulated (API logic, views, actions)
- Adding a new service = create new directory + implement interface
- Core navigation/UI layers interact only via interface (no tight coupling)
- Testing is isolated per service module

---

## User Experience & Interface Design

### Home Screen

```
┌─────────────────────────────────────────────────────────────────────┐
│  ████████╗ ██████╗  ██████╗██████╗                                  │
│  ╚══██╔══╝██╔════╝ ██╔════╝██╔══██╗                                 │
│     ██║   ██║  ███╗██║     ██████╔╝    Terminal GCP Explorer        │
│     ██║   ██║   ██║██║     ██╔═══╝                                  │
│     ██║   ╚██████╔╝╚██████╗██║          v1.0.0                      │
│     ╚═╝    ╚═════╝  ╚═════╝╚═╝                                      │
│                                                                      │
│  👤 User: devops@example.com                                         │
│  📁 Project: my-production-project (project-id-12345)                │
│  🌍 Region: us-central1 (default from ADC)                           │
│                                                                      │
├──────────────────────────────────────────────────────────────────────┤
│  📦 Available Services:                                              │
│                                                                      │
│    ▸ Google Compute Engine (GCE)                                    │
│    ▸ Cloud SQL                                                      │
│    ▸ Identity & Access Management (IAM)                             │
│    ▸ Cloud Run                                    [Coming Soon]     │
│    ▸ Cloud Storage (GCS)                          [Coming Soon]     │
│    ▸ BigQuery                                     [Coming Soon]     │
│    ▸ Networking                                   [Coming Soon]     │
│                                                                      │
├──────────────────────────────────────────────────────────────────────┤
│  ⌨️  Navigation:                                                      │
│    ↑↓ / j/k     Navigate services                                   │
│    Enter        Select service                                      │
│    :            Open command palette                                │
│    ?            Show help                                           │
│    q            Quit                                                │
│                                                                      │
│  💡 Tip: Type ':gce list' to jump directly to GCE instances          │
└──────────────────────────────────────────────────────────────────────┘
```

**Design Notes:**
- ASCII banner rendered with color gradient (using Lipgloss)
- Top metadata bar persists across all views (collapsed to single line after home)
- "Coming Soon" tags for Phase 2 services
- Keybinding hints always visible at bottom

### Service List View (Example: GCE Instances)

```
┌─────────────────────────────────────────────────────────────────────┐
│  TGCP | devops@example.com | my-production-project                  │
├──────┬──────────────────────────────────────────────────────────────┤
│ GCE  │  Instances (42 total, showing 1-42)                          │
│      │                                                              │
│ [▸]  │  Filter: /________________________________________            │
│      │                                                              │
│ SQL  │  NAME              ZONE          STATUS   INT_IP      EXT_IP │
│ IAM  │  ─────────────────────────────────────────────────────────── │
│ Run  │  prod-web-1        us-central1-a RUNNING  10.0.1.5   34.1... │
│ GCS  │  prod-web-2        us-central1-a RUNNING  10.0.1.6   34.2... │
│ BQ   │  prod-worker-1     us-central1-b RUNNING  10.0.2.10  <none> │
│ Net  │  staging-api       us-east1-c    STOPPED  10.1.1.20  35.1... │
│      │  dev-test-vm       us-west1-a    RUNNING  10.2.1.5   34.5... │
│      │  ...                                                         │
│      │                                                              │
│      │  [1/1 pages]                                                 │
│      │                                                              │
├──────┴──────────────────────────────────────────────────────────────┤
│  ↑↓/j/k Navigate | Enter Details | s Start | x Stop | h SSH        │
│  / Filter | : Command Palette | q Back | ? Help                    │
└──────────────────────────────────────────────────────────────────────┘
```

**Design Notes:**
- Left sidebar collapsible (press `Tab` to hide/show)
- Current service highlighted in sidebar
- Inline filter (`/` key) appears at top, filters table in real-time
- Table columns auto-sized based on terminal width
- Status color-coded (RUNNING=green, STOPPED=yellow, ERROR=red)
- Bottom bar shows contextual keybindings for current view

### Resource Detail View (Example: GCE Instance)

```
┌─────────────────────────────────────────────────────────────────────┐
│  TGCP | devops@example.com | my-production-project                  │
├─────────────────────────────────────────────────────────────────────┤
│  Google Compute Engine > Instances > prod-web-1                     │
│                                                                      │
│  ╭─ Instance Details ─────────────────────────────────────────────╮ │
│  │                                                                 │ │
│  │  Name:           prod-web-1                                    │ │
│  │  Status:         🟢 RUNNING (uptime: 23d 14h)                   │ │
│  │  Zone:           us-central1-a                                 │ │
│  │  Machine Type:   n1-standard-2 (2 vCPU, 7.5 GB RAM)            │ │
│  │  Internal IP:    10.0.1.5                                      │ │
│  │  External IP:    34.123.45.67                                  │ │
│  │  Disk:           pd-standard, 50 GB                            │ │
│  │  Tags:           http-server, production, web-tier             │ │
│  │  Service Acct:   prod-web-sa@project.iam.gserviceaccount.com   │ │
│  │                                                                 │ │
│  ╰─────────────────────────────────────────────────────────────────╯ │
│                                                                      │
│  ╭─ Metadata ──────────────────────────────────────────────────────╮ │
│  │  startup-script:  #!/bin/bash...                               │ │
│  │  env:             production                                   │ │
│  │  app-version:     v2.3.1                                       │ │
│  ╰─────────────────────────────────────────────────────────────────╯ │
│                                                                      │
├─────────────────────────────────────────────────────────────────────┤
│  s Start | x Stop | h SSH to instance | l View logs                │
│  q Back to list | : Command Palette | ? Help                       │
└─────────────────────────────────────────────────────────────────────┘
```

**Design Notes:**
- Breadcrumb navigation shows path (GCE > Instances > prod-web-1)
- Grouped sections with borders (using Lipgloss box styling)
- Status indicators with emoji/color
- Actions contextual to resource state (can't start if already running)

### Command Palette

```
┌─────────────────────────────────────────────────────────────────────┐
│  TGCP | devops@example.com | my-production-project                  │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ╭─ Command Palette ────────────────────────────────────────────╮   │
│  │                                                              │   │
│  │  > gce list_____________________________________             │   │
│  │                                                              │   │
│  │  Suggestions:                                                │   │
│  │    📦 gce list          - List GCE instances                 │   │
│  │    🔍 gce search        - Search GCE instances               │   │
│  │    📦 cloudrun list     - List Cloud Run services            │   │
│  │    📦 sql list          - List Cloud SQL instances           │   │
│  │    📦 iam list          - List IAM service accounts          │   │
│  │                                                              │   │
│  ╰──────────────────────────────────────────────────────────────╯   │
│                                                                      │
│  💡 Fuzzy search enabled | Esc to cancel                             │
└─────────────────────────────────────────────────────────────────────┘
```

**Design Notes:**
- Overlay appears on top of current view
- Fuzzy matching (typing "crls" matches "cloudrun list")
- Real-time filtering as user types
- Arrow keys to select, Enter to execute

### Help Screen

```
┌─────────────────────────────────────────────────────────────────────┐
│  TGCP - Help & Keybindings                                           │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ╭─ Global Navigation ─────────────────────────────────────────╮    │
│  │  :            Open command palette                          │    │
│  │  ?            Toggle help screen                            │    │
│  │  q            Back / Quit (context-dependent)               │    │
│  │  Ctrl+C       Force quit application                        │    │
│  │  Tab          Toggle sidebar visibility                     │    │
│  ╰─────────────────────────────────────────────────────────────╯    │
│                                                                      │
│  ╭─ List View ──────────────────────────────────────────────────╮   │
│  │  ↑↓ or j/k    Navigate items                                │   │
│  │  Enter        View item details                             │   │
│  │  /            Filter current list                           │   │
│  │  n / p        Next/Previous page                            │   │
│  ╰─────────────────────────────────────────────────────────────╯    │
│                                                                      │
│  ╭─ Service-Specific Actions ──────────────────────────────────╮    │
│  │  GCE Instances:                                             │    │
│  │    s          Start instance                                │    │
│  │    x          Stop instance                                 │    │
│  │    h          SSH to instance (external session)            │    │
│  │                                                             │    │
│  │  Cloud SQL:                                                 │    │
│  │    s          Start instance                                │    │
│  │    x          Stop instance                                 │    │
│  ╰─────────────────────────────────────────────────────────────╯    │
│                                                                      │
│  ╭─ Command Palette Commands ──────────────────────────────────╮    │
│  │  gce list             List GCE instances                    │    │
│  │  sql list             List Cloud SQL instances              │    │
│  │  iam list             List IAM service accounts             │    │
│  │  cloudrun list        List Cloud Run services               │    │
│  ╰─────────────────────────────────────────────────────────────╯    │
│                                                                      │
├─────────────────────────────────────────────────────────────────────┤
│  q Back | Esc Close                                                 │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Feature Specifications

### 1. Authentication & Project Selection

**Authentication Method:**
- Use Google Cloud Application Default Credentials (ADC)
- Load credentials from standard locations:
  1. `GOOGLE_APPLICATION_CREDENTIALS` environment variable
  2. `gcloud auth application-default login` credentials
  3. GCE/Cloud Run metadata server (if running on GCP)

**Project Detection:**
- Default project from ADC configuration
- Allow override via `--project` flag: `tgcp --project=my-project-id`
- Future: Project switcher in-app (command palette: `project switch`)

**Error Handling:**
- If no ADC found: Display error with instructions to run `gcloud auth application-default login`
- If insufficient permissions: Show specific missing IAM roles

### 2. Service Sidebar Navigation

**Behavior:**
- Left sidebar shows all available services (both implemented and "Coming Soon")
- Arrow keys (↑↓) or vim keys (j/k) to navigate
- Enter to select service → loads resource list in main pane
- Tab key to hide/show sidebar (gives more screen space for tables)

**Visual States:**
- Current service: Highlighted with background color
- Available services: Normal text
- Coming Soon services: Grayed out with `[Coming Soon]` suffix
- Services with errors: Red indicator (e.g., "⚠ GCE [API Error]")

### 3. Command Palette

**Trigger:** `:` key (like Vim command mode)

**Functionality:**
- Fuzzy search across:
  - Service names ("gce", "cloudrun", "sql")
  - Commands ("list", "search")
  - Combined ("gce list", "sql instances")
- Real-time suggestions as user types
- Arrow keys to select suggestion, Enter to execute

**Supported Commands (MVP):**
```
:gce list          → Navigate to GCE instances list
:sql list          → Navigate to Cloud SQL instances list
:iam list          → Navigate to IAM service accounts list
:help              → Open help screen
:quit              → Quit application
```

**Future Commands:**
```
:gce ssh <name>    → SSH directly to instance by name
:project switch    → Switch to different project
:refresh           → Force refresh current view
```

### 4. Resource List Views

**Table Features:**
- Auto-sized columns based on terminal width
- Sortable by column (future enhancement)
- Color-coded status fields:
  - Green: RUNNING, ACTIVE, HEALTHY
  - Yellow: STOPPED, PAUSED, PENDING
  - Red: FAILED, ERROR, TERMINATED

**Inline Filtering:**
- `/` key activates filter input at top of table
- Filters applied in real-time as user types
- Filter syntax: simple substring match (case-insensitive)
- Future: Support regex or field-specific filters (e.g., `status:running`)

**Pagination:**
- 50 items per page (configurable in future via config file)
- `n` / `p` keys for next/previous page
- Page indicator at bottom: `[Page 2/5]`
- If < 50 items, no pagination controls shown

### 5. Resource Detail Views

**Layout:**
- Single-pane drill-down (replaces list view)
- Breadcrumb navigation at top: `Service > Resource Type > Resource Name`
- Grouped sections with borders (e.g., "Instance Details", "Metadata", "Labels")

**Actions:**
- Context-sensitive keybindings shown at bottom
- Example for GCE instance:
  - `s` Start (only if stopped)
  - `x` Stop (only if running)
  - `h` SSH (only if running and has external IP)
  - `l` View logs (future)

### 6. Safe Administrative Actions

**Allowed Operations:**

| Service | Action | Implementation |
|---------|--------|----------------|
| GCE | Start instance | `compute.instances.start()` API call |
| GCE | Stop instance | `compute.instances.stop()` API call (graceful shutdown) |
| GCE | SSH | Spawn external terminal: `gcloud compute ssh <instance>` |
| Cloud SQL | Start instance | `sqladmin.instances.patch()` with activation policy |
| Cloud SQL | Stop instance | `sqladmin.instances.patch()` with deactivation |
| Cloud Run | Restart service | Deploy new revision (future) |

**Confirmation Prompts:**
- Any state-changing action shows confirmation dialog:
  ```
  ╭─ Confirm Action ─────────────────╮
  │ Stop instance "prod-web-1"?      │
  │                                  │
  │ This will gracefully shutdown    │
  │ the VM. Continue?                │
  │                                  │
  │   [Yes (y)]    [No (n)]          │
  ╰──────────────────────────────────╯
  ```

**Forbidden Operations:**
- ❌ Delete (instances, buckets, datasets)
- ❌ Create (no resource provisioning)
- ❌ Modify firewall rules, VPCs (networking changes)
- ❌ IAM role binding changes

### 7. Caching & Refresh Strategy

**Local Cache:**
- In-memory cache with TTL per service type:
  - GCE instances: 30 seconds
  - Cloud SQL: 60 seconds (slower state changes)
  - IAM: 5 minutes (rarely changes)
- Cache stored in Go map with `sync.RWMutex` for thread safety

**Refresh Triggers:**
- Automatic: Background goroutine refreshes on TTL expiration
- Manual: `r` key forces immediate refresh of current view
- On navigation: Switching services triggers refresh if cache expired

**API Rate Limiting:**
- Implement token bucket algorithm:
  - Max 10 requests per second per service
  - Burst allowance: 20 requests
- If rate limit hit, show warning: "⚠ API rate limit reached, using cached data"

**Offline Mode:**
- If API unreachable, show last cached data with warning banner:
  ```
  ⚠ OFFLINE MODE - Showing cached data from 2 minutes ago
  ```

---

## Service Modules

### Service 1: Google Compute Engine (GCE)

**Resources Supported:**
- VM Instances (only; not instance templates, instance groups)

**List View Columns:**
- Name
- Zone
- Status (RUNNING, STOPPED, TERMINATED)
- Internal IP
- External IP (or `<none>`)
- Tags (comma-separated)

**Detail View:**
- Machine type (e.g., `n1-standard-2`)
- CPU/Memory specs
- Boot disk type and size
- Network tags
- Service account
- Metadata key-value pairs
- Uptime (if running)

**Actions:**
- Start instance (if stopped)
- Stop instance (if running)
- SSH (spawns `gcloud compute ssh <instance> --zone=<zone>`)

**API Calls:**
- List: `compute.instances.aggregatedList(project)` (all zones)
- Get: `compute.instances.get(project, zone, instance)`
- Start: `compute.instances.start(project, zone, instance)`
- Stop: `compute.instances.stop(project, zone, instance)`

### Service 2: Cloud SQL

**Resources Supported:**
- SQL instances (MySQL, PostgreSQL, SQL Server)

**List View Columns:**
- Name
- Region
- Database Version (e.g., `POSTGRES_14`)
- Status (RUNNABLE, STOPPED, SUSPENDED)
- Connection Name

**Detail View:**
- Tier (machine type, e.g., `db-n1-standard-1`)
- Storage size and type
- Backup configuration
- Authorized networks
- Private IP / Public IP
- Maintenance window

**Actions:**
- Start instance (activate)
- Stop instance (deactivate)

**API Calls:**
- List: `sqladmin.instances.list(project)`
- Get: `sqladmin.instances.get(project, instance)`
- Patch: `sqladmin.instances.patch(project, instance, body)` for start/stop

### Service 3: Identity & Access Management (IAM)

**Resources Supported:**
- Service Accounts
- IAM Users (from project IAM policy)

**List View (Service Accounts):**
- Email
- Display Name
- Status (Enabled/Disabled)
- Creation Time

**List View (IAM Users):**
- Email
- Roles (comma-separated)

**Detail View:**
- Service Account Keys (list only, no display of private keys)
- IAM role bindings for the service account
- Resource hierarchy (project-level only)

**Actions:**
- None (read-only for MVP)

**API Calls:**
- List SAs: `iam.projects.serviceAccounts.list(project)`
- Get SA: `iam.projects.serviceAccounts.get(name)`
- Get IAM Policy: `cloudresourcemanager.projects.getIamPolicy(project)`

---

## Authentication & Authorization

### Authentication Flow

```
1. TGCP starts
2. Check for ADC credentials:
   - ENV: GOOGLE_APPLICATION_CREDENTIALS
   - File: ~/.config/gcloud/application_default_credentials.json
   - Metadata: GCE/Cloud Run instance metadata server
3. If found:
   - Load credentials
   - Detect default project from ADC config
   - Display home screen with user email + project
4. If not found:
   - Display error screen:
     "No credentials found. Run: gcloud auth application-default login"
   - Exit with code 1
```

### Authorization (IAM Permissions)

**Required Roles (Minimum):**
- `roles/viewer` (project-level) for basic read access
- Service-specific read permissions:
  - `compute.instances.list`
  - `compute.instances.get`
  - `sqladmin.instances.list`
  - `sqladmin.instances.get`
  - `iam.serviceAccounts.list`

**For Administrative Actions:**
- `compute.instances.start`
- `compute.instances.stop`
- `sqladmin.instances.update`

**Permission Error Handling:**
- When listing resources: If 403 Forbidden, show in sidebar as "⚠ GCE [Insufficient Permissions]"
- When performing action: Show error toast with specific missing permission

---

## Error Handling & Resilience

### Error Categories & Responses

| Error Type | User Experience | Technical Handling |
|------------|----------------|-------------------|
| **No ADC credentials** | Error screen with setup instructions | Exit application with code 1 |
| **Invalid project ID** | Toast: "Project not found or inaccessible" | Prompt to change project |
| **API HTTP 403** | Service grayed out: "Insufficient permissions" | Cache last known state, disable actions |
| **API HTTP 429** | Toast: "Rate limit reached, using cached data" | Exponential backoff, use cache |
| **API HTTP 5xx** | Toast: "GCP API error, retrying..." | Retry with backoff (max 3 attempts) |
| **Network timeout** | Banner: "Offline mode - showing cached data" | Use cache, retry in background |
| **Invalid API response** | Toast: "Failed to parse response" | Log error to debug log, skip resource |

### Graceful Degradation

**Partial Service Availability:**
- If GCE API is down but Cloud SQL works:
  - Show "⚠ GCE [API Unavailable]" in sidebar
  - Other services remain functional
  - User can still navigate, just sees stale GCE data

**Cache Staleness Indicators:**
- Show data age in status bar: `Last updated: 2m ago`
- If cache > 5 minutes old, show warning icon

**User-Triggered Recovery:**
- `r` key forces refresh (bypasses cache)
- If refresh fails, show error but retain old cache

---

## Performance & Scalability

### Target Performance Metrics

- **Startup time:** < 2 seconds (from launch to home screen)
- **Service navigation:** < 500ms (switching between services)
- **Resource list loading:** < 3 seconds for 100 resources
- **Filter responsiveness:** < 100ms (real-time typing)
- **Memory footprint:** < 100 MB RAM for typical usage

### Optimization Strategies

**1. Lazy Loading:**
- Don't load all services on startup
- Only fetch resources when user navigates to that service

**2. Pagination:**
- Fetch first 50 items immediately
- Load next page on-demand (when user presses `n`)

**3. API Request Batching:**
- For GCE, use `aggregatedList` instead of per-zone calls
- Batch multiple resource details into single request (where API supports)

**4. Background Refresh:**
- Run refresh goroutines only for currently visible service
- Pause background refresh when TUI is in background (future)

**5. Caching:**
- In-memory cache reduces redundant API calls
- Cache invalidation based on TTL + user actions

### Scalability Limits

**Expected Environment:**
- Projects with 100-1000 resources per service (not 10,000+)
- Single-project view (not multi-project aggregation)

**If User Exceeds Limits:**
- For 500+ instances: Pagination keeps UI responsive
- For 5000+ instances: Recommend using `gcloud` CLI or filtering in Console
- Future: Add server-side filtering (API query parameters)

---

## Configuration & Customization

### Phase 1 (MVP): Zero Configuration
- No config file required
- All settings are defaults:
  - Authentication: ADC
  - Project: From ADC default
  - Refresh intervals: Hardcoded per service
  - Keybindings: Fixed (no customization)

### Phase 2 (Future): `~/.tgcprc` Config File

**Format:** YAML

**Example:**
```yaml
# TGCP Configuration File

# Default project (overrides ADC default)
default_project: "my-prod-project"

# Default region (for region-specific services)
default_region: "us-central1"

# API refresh intervals (seconds)
refresh_intervals:
  gce: 30
  cloudsql: 60
  iam: 300
  cloudrun: 20

# Services to show in sidebar (hide unused services)
enabled_services:
  - gce
  - cloudsql
  - cloudrun
  # iam hidden

# Keybindings customization
keybindings:
  quit: "q"
  help: "?"
  command_palette: ":"
  filter: "/"
  refresh: "r"
  # Service-specific
  gce_ssh: "h"
  gce_start: "s"
  gce_stop: "x"

# UI preferences
ui:
  color_scheme: "auto"  # auto, dark, light
  sidebar_visible: true  # Show sidebar on startup
  page_size: 50          # Items per page
```

**Config Loading Priority:**
1. CLI flags (e.g., `--project=my-project`)
2. `~/.tgcprc` file
3. Defaults

---

## Development Roadmap

### Phase 1: MVP (Weeks 1-4)

**Week 1: Foundation**
- [ ] Project structure setup
- [ ] Bubbletea basic app skeleton
- [ ] Home screen with ASCII banner
- [ ] ADC authentication integration
- [ ] Project detection and display

**Week 2: Core UI Components**
- [ ] Service sidebar navigation
- [ ] Command palette implementation
- [ ] Table component (reusable)
- [ ] Inline filter component
- [ ] Help screen

**Week 3: First Service (GCE)**
- [ ] GCE service module structure
- [ ] List instances view
- [ ] Instance detail view
- [ ] Start/Stop actions with confirmation
- [ ] SSH integration (external spawn)

**Week 4: Services 2 & 3 + Polish**
- [ ] Cloud SQL module (list, detail, start/stop)
- [ ] IAM module (list service accounts, read-only)
- [ ] Error handling (toasts, status bar)
- [ ] Caching layer with TTL
- [ ] Debug logging (`--debug` flag)

**MVP Release Criteria:**
- ✅ All 3 services functional
- ✅ Authentication works via ADC
- ✅ Start/Stop actions confirmed working
- ✅ No crashes on common error scenarios
- ✅ Help documentation complete

### Phase 2: Expansion (Weeks 5-8)

**Services:**
- [ ] Cloud Run (list, restart)
- [ ] Cloud Storage (list buckets, browse objects)
- [ ] BigQuery (list datasets, tables, schemas)
- [ ] Networking (VPCs, firewall rules - read-only)

**Features:**
- [ ] Log tailing (Cloud Logging API integration)
- [ ] Project switcher (command palette)
- [ ] Config file support (`~/.tgcprc`)
- [ ] Custom keybindings
- [ ] Column sorting in tables

### Phase 3: Advanced Features (Weeks 9-12)

- [ ] IAP tunnel support for SSH
- [ ] Cross-project switching (within org)
- [ ] Multi-pane views (experimental)
- [ ] Billing summary dashboard
- [ ] Resource search across all services
- [ ] Export to JSON/CSV

### Future (Post v1.0)

- [ ] AI-powered insights (Gemini integration)
- [ ] Anomaly detection (high CPU, long-running instances)
- [ ] Cloud Functions (list, view logs)
- [ ] GKE clusters (basic info, not full K9s replacement)
- [ ] Pub/Sub topics and subscriptions

---

## Technical Stack

### Core Technologies

| Component | Technology | Justification |
|-----------|-----------|---------------|
| **Language** | Go 1.21+ | - Excellent GCP library support<br>- Single binary distribution<br>- Strong concurrency for API calls<br>- K9s is Go-based (familiar patterns) |
| **TUI Framework** | [Bubbletea](https://github.com/charmbracelet/bubbletea) | - Elm-inspired architecture (predictable state)<br>- Rich ecosystem (Lipgloss for styling, Bubbles for components)<br>- Active community, well-documented<br>- AI-friendly codebase |
| **Styling** | [Lipgloss](https://github.com/charmbracelet/lipgloss) | - Declarative styling (like CSS for terminal)<br>- Color, borders, layout primitives<br>- Integrates with Bubbletea |
| **GCP APIs** | `google.golang.org/api`<br>`cloud.google.com/go` | - Official Google libraries<br>- Auto-generated from API specs<br>- Auth handled via ADC |
| **Authentication** | `golang.org/x/oauth2/google` | - ADC support out-of-the-box<br>- Token refresh handling |

### Key Dependencies

```go
// go.mod (simplified)
module github.com/yourusername/tgcp

go 1.21

require (
    github.com/charmbracelet/bubbletea v0.25.0
    github.com/charmbracelet/lipgloss v0.9.1
    github.com/charmbracelet/bubbles v0.18.0

    google.golang.org/api v0.150.0
    cloud.google.com/go/compute v1.23.0
    cloud.google.com/go/iam v1.1.0

    golang.org/x/oauth2 v0.15.0

    // Utilities
    github.com/sahilm/fuzzy v0.1.0  // Fuzzy search for command palette
    gopkg.in/yaml.v3 v3.0.1         // Config file parsing (future)
)
```

### Development Tools

- **Build:** `go build` (no additional tooling required)
- **Testing:** `go test` (unit tests per service module)
- **Linting:** `golangci-lint` (enforce code quality)
- **Debugging:** `--debug` flag writes to `~/.tgcp/debug.log`

### Distribution

**Method 1: Go Install**
```bash
go install github.com/yourusername/tgcp@latest
```

**Method 2: Homebrew (macOS/Linux)**
```bash
brew tap yourusername/tgcp
brew install tgcp
```

**Method 3: Direct Binary Download**
- GitHub Releases with binaries for:
  - `tgcp-darwin-amd64` (macOS Intel)
  - `tgcp-darwin-arm64` (macOS Apple Silicon)
  - `tgcp-linux-amd64`
  - `tgcp-linux-arm64`
  - `tgcp-windows-amd64.exe`

### Version Management

- **Semantic Versioning:** `v1.0.0`, `v1.1.0`, etc.
- **Update Check:** On launch, TGCP checks GitHub API for latest release
  ```
  💡 New version available: v1.2.0 (current: v1.1.0)
     Run: brew upgrade tgcp
  ```

---

## Success Metrics

### MVP Success Criteria

**Functional Requirements:**
- [ ] User can authenticate via ADC without manual token entry
- [ ] All 3 MVP services (GCE, Cloud SQL, IAM) display correctly
- [ ] Start/Stop actions execute without errors
- [ ] SSH spawns successfully for GCE instances
- [ ] No crashes on common workflows (list → detail → action → back)

**User Experience:**
- [ ] Home screen loads in < 2 seconds
- [ ] Service switching feels instant (< 500ms)
- [ ] Inline filtering responds in real-time (< 100ms)
- [ ] Error messages are actionable (tell user what to do)

**Code Quality:**
- [ ] Service modules are independent (adding GCS doesn't require changing GCE code)
- [ ] Test coverage > 60% for core logic (API calls, caching)
- [ ] No hardcoded project IDs or credentials in source

### Post-Launch Metrics (for yourself/team)

**Adoption:**
- Number of times TGCP is launched per week (track via opt-in telemetry or manual logs)
- Percentage of GCP Console visits replaced by TGCP

**Efficiency:**
- Time saved: "How long did it take to check VM status?" (Console: 30s, TGCP: 5s)
- Workflow completion: "Can I complete task X entirely in terminal?"

**Community (if open-sourced):**
- GitHub stars, forks, issues
- Contributor PRs for new services
- Feature requests

---

## Future Enhancements

### Short-Term (Post-MVP, within 6 months)

1. **Log Tailing**
   - Integrate Cloud Logging API
   - Stream logs in real-time (like `tail -f`)
   - Filter by severity, timestamp

2. **Resource Filtering**
   - Server-side filtering via API query params (reduce data transfer)
   - Regex support in inline filter

3. **Custom Dashboards**
   - User-defined views (e.g., "My Production VMs" showing specific labels)

4. **Batch Actions**
   - Multi-select resources (checkboxes)
   - Start/Stop multiple instances at once

### Medium-Term (6-12 months)

5. **AI-Powered Insights**
   - Gemini integration for anomaly detection
   - Suggestions: "Instance prod-web-1 has been idle for 5 days, consider stopping"

6. **Cloud Functions**
   - List functions
   - View invocation logs
   - Trigger test invocations

7. **GKE Integration**
   - List clusters
   - Basic node/pod info (not full K9s replacement)

8. **Billing Dashboard**
   - Monthly cost summary per service
   - Anomaly alerts ("BigQuery costs up 300% this month")

### Long-Term (12+ months)

9. **Multi-Project Aggregation**
   - View resources across multiple projects
   - Org-level IAM management

10. **Plugin System**
    - Third-party service modules
    - User-contributed plugins for niche services

11. **Remote Mode**
    - TGCP server/client architecture
    - Share terminal session with team (like `screen`)

12. **Mobile Companion**
    - iOS/Android app for on-call emergencies
    - "Approve production VM restart" push notification

---

## Appendix

### A. Glossary

- **ADC:** Application Default Credentials (Google Cloud authentication method)
- **TUI:** Terminal User Interface
- **IAM:** Identity and Access Management
- **GCE:** Google Compute Engine
- **GCS:** Google Cloud Storage
- **GKE:** Google Kubernetes Engine
- **IAP:** Identity-Aware Proxy
- **TTL:** Time To Live (cache expiration)

### B. References

- [K9s GitHub](https://github.com/derailed/k9s) - Kubernetes TUI inspiration
- [TAWS GitHub](https://github.com/huseyinbabal/taws) - AWS TUI reference
- [Bubbletea Documentation](https://github.com/charmbracelet/bubbletea)
- [GCP Go Client Libraries](https://cloud.google.com/go/docs/reference)
- [Google Cloud APIs](https://cloud.google.com/apis)

### C. Open Questions (To Resolve During Development)

1. **SSH via IAP:** Should we support `gcloud compute start-iap-tunnel` for private instances, or only external IP SSH in MVP?
   - **Decision pending:** Start with external IP only, add IAP in Phase 2

2. **Config File Format:** YAML vs TOML vs JSON?
   - **Decision:** YAML (most readable, widely supported)

3. **Telemetry:** Opt-in usage analytics for improving tool?
   - **Decision pending:** Discuss with team, prioritize privacy

4. **Windows Support:** Terminal rendering differences, worth the effort?
   - **Decision:** Support in binaries, but don't test extensively until user demand

---

## Document Metadata

**Author:** Technical Product Manager (AI-assisted)
**Created:** 2026-01-06
**Version:** 1.0 (Initial PRD)
**Next Review:** After MVP completion (Week 4)
**Status:** Approved for Development

---

## Sign-off

This PRD serves as the foundational blueprint for TGCP. Any major scope changes (e.g., adding services beyond planned roadmap, changing core architecture) require PRD update and stakeholder review.

**Approved by:**
- [ ] Product Owner (you)
- [ ] Lead Developer (you)
- [ ] Team Review (if applicable)

**Development may commence:** ✅ YES

---

*End of Document*
