# Codebase Structure

**Analysis Date:** 2026-02-09

## Directory Layout

```
tgcp/
├── cmd/                           # Application entry points
│   └── tgcp/
│       └── main.go               # Main binary entry point
├── internal/                       # All application code
│   ├── config/                    # Configuration management
│   │   └── config.go             # Config loading from ~/.tgcprc
│   ├── core/                      # Core infrastructure
│   │   ├── auth.go               # Google Cloud ADC authentication
│   │   ├── cache.go              # TTL-based in-memory cache
│   │   ├── client.go             # GCP API client management
│   │   ├── events.go             # Event/message type definitions
│   │   ├── navigation.go         # Navigation model & command palette
│   │   ├── projects.go           # Project management/switching
│   │   ├── registry.go           # Service registry & lazy initialization
│   │   └── version.go            # Version information
│   ├── services/                  # GCP service implementations
│   │   ├── interface.go          # Service interface contract
│   │   ├── bigquery/             # BigQuery service (datasets, tables)
│   │   ├── bigtable/             # Bigtable service
│   │   ├── cloudrun/             # Cloud Run service
│   │   ├── cloudsql/             # Cloud SQL service
│   │   ├── dataflow/             # Dataflow service
│   │   ├── dataproc/             # Dataproc service
│   │   ├── disks/                # Persistent Disks service
│   │   ├── firestore/            # Firestore service
│   │   ├── gce/                  # Google Compute Engine
│   │   ├── gcs/                  # Google Cloud Storage
│   │   ├── gke/                  # Google Kubernetes Engine
│   │   ├── iam/                  # Identity & Access Management
│   │   ├── logging/              # Cloud Logging service
│   │   ├── net/                  # VPCs, Subnets, Firewalls
│   │   ├── overview/             # Project overview/summary
│   │   ├── pubsub/               # Pub/Sub service
│   │   ├── redis/                # Memorystore (Redis)
│   │   ├── secrets/              # Secret Manager
│   │   └── spanner/              # Cloud Spanner
│   ├── styles/                    # Terminal styling & colors
│   │   └── [color definitions]   # Lipgloss style definitions
│   ├── ui/                        # UI layer & rendering
│   │   ├── model.go              # Main application model (MainModel)
│   │   ├── home.go               # View rendering & layout
│   │   ├── auth_error.go         # Auth error screen
│   │   ├── banner.go             # TGCP banner rendering
│   │   ├── help.go               # Help overlay
│   │   └── components/           # Reusable TUI components
│   │       ├── table.go          # StandardTable with focus management
│   │       ├── filter.go         # Filter input component
│   │       ├── filter_session.go # Generic filtering session
│   │       ├── sidebar.go        # Service list sidebar
│   │       ├── palette.go        # Command palette
│   │       ├── home_menu.go      # Home page menu
│   │       ├── spinner.go        # Loading spinner
│   │       ├── statusbar.go      # Status bar at bottom
│   │       ├── toast.go          # Toast notifications
│   │       ├── confirmation.go   # Confirmation dialogs
│   │       ├── breadcrumb.go     # Navigation breadcrumb
│   │       ├── detail.go         # Detail view component
│   │       └── error.go          # Error display component
│   └── utils/                     # Utility functions
│       ├── logger.go             # Debug logging to ~/.tgcp/debug.log
│       └── tmux.go               # Tmux/terminal utilities
├── docs/                          # User documentation
│   ├── FEATURES.md               # Feature descriptions
│   ├── DEVELOPER_GUIDE.md        # Development guide
│   └── ui_patterns.md            # UI pattern documentation
├── assets/                        # Static assets
├── go.mod                        # Go module definition
├── go.sum                        # Dependency checksums
├── .goreleaser.yaml              # Release configuration
├── README.md                     # Main documentation
├── LICENSE                       # MIT license
└── website/                       # Landing page (Astro + Tailwind v4)
    ├── src/
    │   ├── components/           # Astro components (Header, Hero, Services, etc.)
    │   ├── data/
    │   │   └── services.ts       # Service catalog (KEEP IN SYNC with TUI)
    │   ├── layouts/
    │   │   └── Base.astro        # Base HTML layout
    │   ├── pages/
    │   │   └── index.astro       # Single-page landing site
    │   └── styles/
    │       └── global.css        # Tailwind v4 theme tokens + animations
    ├── public/                   # Static assets (fonts, favicon)
    ├── astro.config.mjs          # Astro config (GitHub Pages base: '/tgcp')
    └── package.json              # Astro + Tailwind dependencies
```

## Directory Purposes

**cmd/tgcp:**
- Purpose: Binary entry point
- Contains: `main.go` with flag parsing, authentication, and Bubbletea startup
- Key files: `main.go`

**internal/config:**
- Purpose: Configuration file handling
- Contains: Config struct, loading from `~/.tgcprc` (YAML format)
- Key files: `config.go`

**internal/core:**
- Purpose: Shared infrastructure (auth, caching, service registry, navigation, events)
- Contains: Core abstractions used by all layers
- Key files:
  - `auth.go` - Google Cloud ADC authentication
  - `cache.go` - Thread-safe TTL cache
  - `registry.go` - Service factory pattern with lazy initialization
  - `navigation.go` - Route model and command palette
  - `events.go` - Event/message types for Bubbletea communication
  - `client.go` - GCP API client management
  - `projects.go` - Project switching logic
  - `version.go` - Build version information

**internal/services:**
- Purpose: GCP service implementations
- Contains: 16+ service packages, each a pluggable module implementing the Service interface
- Key files:
  - `interface.go` - Service interface contract (Name, InitService, Update, View, etc.)
  - `{service}/{service}.go` - Main service struct and interface implementation
  - `{service}/api.go` - GCP API calls and data fetching
  - `{service}/models.go` - Data structures for resources
  - `{service}/views.go` - Rendering logic (optional, some use inline)

**internal/styles:**
- Purpose: Terminal styling and color definitions
- Contains: Centralized color palette, style definitions using Lipgloss
- Key files: Style exports used throughout UI layer

**internal/ui:**
- Purpose: User interface rendering and component management
- Contains: Main model, view rendering, and reusable components
- Key files:
  - `model.go` - MainModel struct, InitialModel initialization
  - `home.go` - Main View() function and landing page rendering
  - `auth_error.go` - Authentication error display
  - `banner.go` - ASCII art banner
  - `help.go` - Help overlay rendering

**internal/ui/components:**
- Purpose: Reusable TUI components
- Contains: Components used across services
- Key files:
  - `table.go` - StandardTable with focus, blur, and window sizing
  - `filter.go` - Filter input UI
  - `sidebar.go` - Service navigation sidebar
  - `palette.go` - Command palette overlay
  - `spinner.go` - Loading indicator
  - `statusbar.go` - Bottom status bar
  - `toast.go` - Notification popups
  - `confirmation.go` - Confirmation dialogs
  - `detail.go` - Detail view wrapper
  - `error.go` - Error display wrapper

**internal/utils:**
- Purpose: Utility functions
- Contains: Logging, terminal utilities
- Key files:
  - `logger.go` - File-based debug logging
  - `tmux.go` - Terminal/Tmux integration

## Key File Locations

**Entry Points:**
- `cmd/tgcp/main.go`: Binary entry point; handles flags, auth, initialization
- `internal/ui/model.go`: InitialModel() function creates the initial UI state
- `internal/ui/home.go`: MainModel.View() and MainModel.Update() methods (main event loop)

**Configuration:**
- `internal/config/config.go`: Configuration loading from `~/.tgcprc`
- User config path: `~/.tgcprc` (YAML format)
- Debug log path: `~/.tgcp/debug.log` (created when `--debug` flag used)

**Core Logic:**
- `internal/core/auth.go`: ADC authentication and project detection
- `internal/core/registry.go`: Service registry with lazy initialization
- `internal/core/cache.go`: In-memory cache with TTL
- `internal/core/navigation.go`: Navigation model and command palette logic
- `internal/core/events.go`: Event message types for inter-component communication

**Service Implementation:**
- `internal/services/interface.go`: Service interface definition
- `internal/services/{service}/{service}.go`: Each service's main implementation
- `internal/services/{service}/api.go`: API calls to Google Cloud
- `internal/services/{service}/models.go`: Data structures

**UI Components:**
- `internal/ui/components/table.go`: StandardTable (used by most services)
- `internal/ui/components/filter.go`: Filter input UI
- `internal/ui/components/palette.go`: Command palette
- `internal/ui/components/sidebar.go`: Service navigation

## Naming Conventions

**Files:**
- Lowercase, underscore-separated: `auth_error.go`, `home_menu.go`, `filter_session.go`
- Service directories match service names: `gce/`, `gcs/`, `bigquery/` (lowercase, no underscores)
- Service files within directory: `{service}.go`, `api.go`, `models.go`, `views.go`

**Directories:**
- Package directories match Go package names (lowercase, no underscores)
- Service directories: single word or abbreviation (`gce`, `gcs`, `bigquery`)
- Functional groupings: `services/`, `components/`, `internal/`

**Go Packages:**
- Match directory name: `package gce`, `package components`, `package core`
- Services use simple names: `package gce`, `package gcs`, `package sql` (for Cloud SQL)

**Types:**
- PascalCase: `Service`, `MainModel`, `FilterModel`, `StandardTable`, `AuthState`
- Service types: `Service` interface implemented by each service
- Model suffix for UI models: `MainModel`, `FilterModel`, `SpinnerModel`

**Functions:**
- Exported: PascalCase: `InitService()`, `Authenticate()`, `RefreshData()`
- Unexported: camelCase: `registerAllServices()`, `renderLandingPage()`
- Factory functions: `New{Type}()`: `NewService()`, `NewCache()`, `NewStandardTable()`

**Variables:**
- Exported: PascalCase: `ViewHome`, `CacheTTL`, `ColorTextPrimary`
- Unexported: camelCase: `selectedInstance`, `filterSession`, `viewState`
- Constants: SCREAMING_SNAKE_CASE or PascalCase: `CacheTTL`, `ViewList`, `TableSelectedFocused`

**Message Types (Bubbletea):**
- Suffix with `Msg`: `instancesMsg`, `bucketsMsg`, `errMsg`, `tickMsg`
- Bubbletea message types from core: `StatusMsg`, `ToastMsg`, `LoadingMsg`, `SwitchToServiceMsg`

## Where to Add New Code

**New GCP Service:**
1. Create directory: `internal/services/{service}/`
2. Implement files:
   - `{service}.go`: Main struct implementing `services.Service` interface
   - `api.go`: GCP API calls using Google Cloud client libraries
   - `models.go`: Data structures for resources
   - `views.go`: (optional) Rendering logic if complex
3. Register in `internal/ui/model.go`: `registerAllServices()` function
4. Add command to palette in `internal/core/navigation.go`: `defaultCommands()` function
5. **Update website**: Add service to `website/src/data/services.ts` (and update `website/src/components/Hero.astro` TUI preview if category structure changes)

**New UI Component:**
1. Create file: `internal/ui/components/{component}.go`
2. Implement component model with `Update(tea.Msg)` and `View()` methods
3. Follow existing patterns (focus/blur, window sizing)
4. Use centralized styles from `internal/styles/`
5. Document public API and usage patterns in comments

**New Feature (Cross-Service):**
1. Define event type in `internal/core/events.go` if inter-component communication needed
2. Implement feature in service(s): `internal/services/{service}/{service}.go`
3. Update UI model if new view state needed: `internal/ui/model.go`
4. Add command to palette if user-facing: `internal/core/navigation.go`

**Utility Function:**
1. Add to existing file in `internal/utils/` if related to existing utilities
2. Create new file if new category: `internal/utils/{category}.go`
3. Ensure functions are exported (PascalCase) if used across packages

**Configuration Addition:**
1. Add field to `Config` struct in `internal/config/config.go`
2. Add defaults in `DefaultConfig()` function
3. YAML tags must match field names (snake_case in YAML)

## Special Directories

**internal/services/{service}:**
- Purpose: Each GCP service is a self-contained module
- Generated: No (all hand-written)
- Committed: Yes (all committed to git)
- Structure: Each service is independent; can be added/removed without affecting others

**website/:**
- Purpose: Landing page at https://yogirk.github.io/tgcp
- Stack: Astro 5 + Tailwind CSS v4 (dark-mode only, terminal aesthetic)
- Key data file: `src/data/services.ts` — must stay in sync with TUI service catalog
- Key components: Hero (TUI preview), Services (tree view), Install (3 methods)
- GitHub stars fetched at build time in `src/components/Header.astro`
- Deployed to GitHub Pages with base path `/tgcp`
- Generated: No (hand-maintained)
- Committed: Yes

**assets/:**
- Purpose: Static files (currently just image.png for README)
- Generated: No (hand-maintained)
- Committed: Yes

**docs/:**
- Purpose: User-facing documentation
- Generated: No (hand-written)
- Committed: Yes
- Contents: FEATURES.md, DEVELOPER_GUIDE.md, UI patterns

**.planning/codebase/:**
- Purpose: Generated codebase analysis documents (created by GSD mapper)
- Generated: Yes (auto-generated)
- Committed: Yes (part of project planning)

**~/.tgcp/:**
- Purpose: User's home directory cache
- Generated: Yes (at runtime)
- Committed: No (local user directory)
- Contents: `debug.log` when `--debug` flag used

## Import Organization

Within Go files, follow standard Go import organization:

```go
// 1. Standard library
import (
    "context"
    "fmt"
    "time"
)

// 2. External packages (third-party)
import (
    tea "github.com/charmbracelet/bubbletea"
    "cloud.google.com/go/compute/apiv1"
    "google.golang.org/api/option"
)

// 3. Internal packages
import (
    "github.com/yogirk/tgcp/internal/core"
    "github.com/yogirk/tgcp/internal/ui/components"
)
```

No path aliases used in codebase; all imports are direct package paths.

---

*Structure analysis: 2026-02-09*
