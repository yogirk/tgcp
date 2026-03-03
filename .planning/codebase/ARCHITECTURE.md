# Architecture

**Analysis Date:** 2026-03-03

## Pattern Overview

**Overall:** Plugin-based service architecture with a centralized TUI (Terminal User Interface) model powered by Bubbletea.

**Key Characteristics:**
- **Service Registry Pattern**: All GCP services (GCE, GCS, BigQuery, etc.) are pluggable modules registered in a centralized registry
- **Lazy Service Initialization**: Services are created but only initialized when first accessed (on-demand)
- **Separation of Concerns**: Clear boundaries between UI components, services, and core infrastructure
- **Message-Driven Architecture**: Bubbletea's message-passing system drives all state updates and inter-component communication
- **Stateful Services**: Each service maintains its own state (data, UI components, view mode) and manages its full lifecycle

## Layers

**Presentation Layer:**
- Purpose: Render TUI and handle user input
- Location: `internal/ui/`
- Contains: Main model (`model.go`), view rendering (`home.go`), components (`components/`)
- Depends on: Services, Core (Navigation, Events), Styles
- Used by: Main application loop

**Service Layer:**
- Purpose: Implement GCP resource browsing and management for specific services
- Location: `internal/services/`
- Contains: Individual service packages (gce/, gcs/, bigquery/, cloudsql/, etc.)
- Depends on: Core (Cache, Events), UI Components, Google Cloud APIs
- Used by: Main UI Model, dispatches events back to UI

**Core Infrastructure Layer:**
- Purpose: Provide shared utilities and orchestration (auth, caching, navigation, event types)
- Location: `internal/core/`
- Contains: Authentication, Cache, Service Registry, Navigation, Event definitions
- Depends on: Google Cloud auth libraries
- Used by: All layers

**Configuration Layer:**
- Purpose: Load and manage application settings
- Location: `internal/config/`
- Contains: Config structure and loading from `~/.tgcprc`
- Depends on: Nothing (leaf dependency)
- Used by: Main entry point

**Command Layer:**
- Purpose: Application entry point and initialization
- Location: `cmd/tgcp/main.go`
- Contains: Flag parsing, logger initialization, authentication, UI startup
- Depends on: All other layers
- Used by: Operating system/binary invocation

## Data Flow

**Application Startup:**

1. `cmd/tgcp/main.go` parses flags and initializes logger
2. `config.LoadConfig()` loads `~/.tgcprc` or defaults
3. `core.Authenticate()` performs ADC check and detects project ID
4. `core.NewServiceRegistry()` creates registry with shared cache
5. `registerAllServices()` registers all 21 GCP service factories
6. `ui.InitialModel()` creates UI model, which calls `registry.InitializeAll()` to create service instances (not initialized yet)
7. `tea.NewProgram()` starts Bubbletea main loop

**User Navigation Flow:**

1. User presses `:` → Command Palette becomes active
2. `NavigationModel.PaletteActive = true` triggers palette rendering
3. User types and selects service → `Command.Action()` returns a `Route`
4. `Route` updates `NavigationModel.CurrentRoute`
5. `MainModel.Update()` detects route change, calls `GetOrInitializeService()` if service needs initialization
6. Service becomes active, rendering replaces content, user input focused to service

**Service Data Refresh Flow:**

1. Service is in focus, user presses `r` (refresh)
2. Service emits `Refresh()` command (returns `tea.Cmd`)
3. Cmd spawns goroutine that fetches data from Google Cloud API
4. Result packaged as service-specific message (e.g., `instancesMsg`, `bucketsMsg`)
5. `Service.Update()` receives message, updates internal state
6. Next `View()` call renders updated data
7. Cache TTL controls background refresh cadence per service

**Inter-Service Communication:**

1. Service A wants to jump to Service B with context (e.g., view logs for a resource)
2. Service A emits `core.SwitchToServiceMsg` or `core.SwitchToLogsMsg`
3. `MainModel.Update()` catches message, switches `CurrentSvc` to target service
4. Passes optional context (filter string, resource ID) to service
5. Target service may filter/populate results based on context

**State Management:**

- **UI State**: Managed by `MainModel` (focus, view mode, palette state, navigation history)
- **Service State**: Each service manages its own internal state (list data, selected item, view state, filters)
- **Global State**: Cache (shared across services), Authentication state, Navigation history
- **Cross-Service State**: Passed via messages (e.g., `SwitchToServiceMsg` includes context)

## Key Abstractions

**Service Interface:**
- Purpose: Defines contract all GCP services must implement
- Location: `internal/services/interface.go`
- Key methods: `Name()`, `InitService()`, `Update()`, `View()`, `Refresh()`, `Focus()`, `Blur()`, `IsRootView()`
- Pattern: All services implement this interface, allowing polymorphic handling by UI layer

**ServiceRegistry:**
- Purpose: Factory pattern for lazy service initialization
- Location: `internal/core/registry.go`
- Pattern: `Register()` stores factory functions, `GetOrInitializeService()` creates and initializes services on-demand
- Thread-safe with RWMutex to prevent race conditions during initialization

**Cache:**
- Purpose: In-memory TTL-based caching for API responses
- Location: `internal/core/cache.go`
- Pattern: Thread-safe map with expiration tracking; services check cache before making API calls

**NavigationModel:**
- Purpose: Command palette, route history, and fuzzy search over available commands
- Location: `internal/core/navigation.go`
- Pattern: Maintains route history, fuzzy-matches commands for palette, routes to services or views

**StandardTable Component:**
- Purpose: Reusable table widget with consistent styling, focus management, and window sizing
- Location: `internal/ui/components/table.go`
- Pattern: Wraps Bubbletea's table.Model, adds focus/blur state, automatic height adjustment, selection handling
- Used by: All services that display list views

**FilterSession Component:**
- Purpose: Generic filtering UI for list items
- Location: `internal/ui/components/filter_session.go`
- Pattern: Generic type allows services to define filter logic for their data types; manages filter input, matching, and rendering

## Entry Points

**Main Executable:**
- Location: `cmd/tgcp/main.go`
- Triggers: User runs `tgcp` command
- Responsibilities: Parse flags, authenticate, initialize services, start TUI event loop

**InitialModel:**
- Location: `internal/ui/model.go` function `InitialModel()`
- Triggers: Called from main after setup
- Responsibilities: Create MainModel, initialize all service instances, set up components, establish initial navigation state

**Service.InitService:**
- Location: `internal/services/{service}/{service}.go`
- Triggers: Called when service is first accessed via `GetOrInitializeService()`
- Responsibilities: Create GCP API clients, fetch initial data, set up internal state

## Error Handling

**Strategy:** Errors are caught at each layer and propagated as event messages or displayed in UI

**Patterns:**
- **Authentication Errors**: Caught in `main.go`, AuthState.Error set; UI renders auth error screen
- **API Errors**: Services catch errors from API calls, store in `.err` field, render error message in View
- **Configuration Errors**: Logged but don't block app startup (defaults used)
- **User-Facing Errors**: Displayed in status bar or as toast notifications
- **Fatal Errors**: Logged to `~/.tgcp/debug.log`, user sees error screen with troubleshooting hints

## Cross-Cutting Concerns

**Logging:**
- Framework: Custom file-based logger (`internal/utils/logger.go`)
- Writes to: `~/.tgcp/debug.log` when `--debug` flag used
- Usage: Core infrastructure (auth, registry) and services log initialization and errors

**Validation:**
- Where: Input validation in UI components (filter text, command palette queries)
- Pattern: Each component validates its input before passing to services
- Fuzzy search engine (`sahilm/fuzzy`) validates command palette queries

**Authentication:**
- Provider: Google Cloud ADC (Application Default Credentials)
- Flow: `core.Authenticate()` uses `google.FindDefaultCredentials()`, checks multiple sources (env var, standard path, gcloud config)
- State: `core.AuthState` holds authentication result and project ID; checked in main update loop before service operations

**Styling:**
- Framework: `charmbracelet/lipgloss` for terminal styling
- Location: Centralized in `internal/styles/` and component-specific styles
- Pattern: Services use common color palette (`ColorTextPrimary`, `ColorTextSecondary`, etc.) for visual consistency

---

*Architecture analysis: 2026-03-03*
