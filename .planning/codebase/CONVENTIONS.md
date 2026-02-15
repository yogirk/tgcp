# Coding Conventions

**Analysis Date:** 2026-02-09

## Naming Patterns

**Files:**
- Service files: `[service_name].go` (e.g., `gce.go`, `pricing.go`)
- Test files: `[name]_test.go` (e.g., `pricing_test.go`)
- Utility files: `[function_area].go` (e.g., `logger.go`, `cache.go`, `client.go`)
- Model/struct files: `models.go`, `events.go`
- UI component files: `[component_name].go` (e.g., `filter.go`, `table.go`, `sidebar.go`)
- View files: `views.go` or `[feature]_views.go`

**Functions:**
- CamelCase, starting with lowercase for unexported functions: `getFilteredInstances()`, `updateTable()`
- CamelCase, starting with uppercase for exported functions: `NewCache()`, `EstimateCost()`, `Authenticate()`
- Receiver methods: camelCase following Go convention: `(c *Cache) Set()`, `(s *Service) Name()`
- Verb-first naming for action functions: `RefreshData()`, `ValidateInput()`, `InitializeService()`
- Getter/setter pattern not used; simple noun names for accessors: `Name()`, `ShortName()` (not `GetName()`)
- Factory functions: `New[Type]` pattern: `NewCache()`, `NewFilter()`, `NewStandardTable()`
- Handler functions: `Handle[Event]` or `On[Event]`: (not evident in primary code, but pattern is used for Bubbletea messages)

**Variables:**
- Context variables: `ctx` (always short)
- Model instances: `m` for local receiver (standard Bubbletea pattern)
- Error variables: `err` (always)
- Slice/array variables: plural form: `instances`, `projects`, `commands`
- Configuration variables: `cfg`
- Cache variables: lowercase: `cache *core.Cache`
- State variables: lowercase: `viewState`, `selectedInstance`, `pendingAction`
- Constants: SCREAMING_SNAKE_CASE for constants: `CacheTTL`, `ViewHome`, `FocusSidebar`
- Short-lived local vars in loops: single letters acceptable: `i`, `v`, `tt` (for table tests)
- Long-lived vars: descriptive names

**Types:**
- Struct names: PascalCase: `AuthState`, `Cache`, `MainModel`, `Instance`
- Interface names: PascalCase: `Service` (required), `Writer` patterns
- Type aliases for enums: PascalCase: `ViewMode`, `FocusArea`, `InstanceState`
- Enum constants: PascalCase with context prefix: `ViewHome`, `ViewService`, `StateRunning`, `StateStopped`

## Code Style

**Formatting:**
- Tool: `gofmt` (standard Go formatter, built-in)
- Line length: No explicit limit specified in project, follows Go standard (typically 80-120 chars)
- Tab indentation: Go standard (tabs, 1 tab = 4 spaces in editors)

**Linting:**
- Tool: `go vet` (standard Go linter)
- Config: None specified (uses defaults)
- Run command: `go vet ./...`

**Package Documentation:**
- Packages documented with PDoc comment blocks at package level
- Example: `/Users/rk/Projects/tgcp/internal/core/client.go` includes comprehensive package-level documentation explaining rate limiting, retry logic, and examples
- Multi-line documentation for complex functionality included

**Function Documentation:**
- Exported functions documented with comment starting with function name: `// EstimateCost returns a formatted string estimate...`
- Parameters and return values documented inline in function comment
- Complex algorithms documented with detailed explanation (see `client.go` for token bucket and retry logic documentation)

## Import Organization

**Order:**
1. Standard library imports: `context`, `fmt`, `os`, `time`, `sync`, `http`, etc.
2. External packages: `github.com/charmbracelet/*`, `golang.org/x/*`, `google.golang.org/*`, `gopkg.in/*`
3. Internal packages: `github.com/yogirk/tgcp/internal/*`

**Pattern:** Observed in all files. Example from `internal/ui/model.go`:
```go
import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yogirk/tgcp/internal/config"
	"github.com/yogirk/tgcp/internal/core"
	"github.com/yogirk/tgcp/internal/services"
	// ... more internal imports
)
```

**Path Aliases:**
- `tea` for `github.com/charmbracelet/bubbletea`: `tea.Cmd`, `tea.Model`, `tea.Msg`
- No other aliases observed; full paths used for other packages

**Local Imports:**
- All internal imports use full module path: `github.com/yogirk/tgcp/internal/...`
- No relative imports

## Error Handling

**Pattern - Error Returns:**
- Functions that can fail return `(T, error)` tuple
- Error always last: `InitService(ctx context.Context, projectID string) error`
- Single error type used: `error` interface
- Wrapping with context: `fmt.Errorf("message: %w", err)` for wrapped errors
- Example from `auth.go`: `return state, fmt.Errorf("could not find default credentials: %w", err)`

**Pattern - Error Checking:**
- Explicit nil check: `if err != nil { ... }`
- No error swallowing; errors propagated up call stack or logged
- Errors stored in struct fields for later display: `Service.err` field in gce service

**Pattern - Error Display:**
- UI errors rendered via `components.RenderError()` function
- Status bar shows error messages: `StatusMsg` with `IsError` field
- Toast notifications for transient errors: `ToastMsg` with `ToastError` type
- Startup auth errors trigger `renderAuthError()` view

**No Panic Pattern:**
- No panic usage observed in service/core code
- All runtime errors handled via return values or error fields

## Logging

**Framework:** `github.com/yogirk/tgcp/internal/utils` custom logger

**Implementation:**
- Custom logger in `/Users/rk/Projects/tgcp/internal/utils/logger.go`
- Writes to `~/.tgcp/debug.log` file
- Enabled via `-debug` CLI flag
- Thread-unsafe (single file handle): `var logFile *os.File`

**Logging Patterns:**
- Initialize: `utils.InitLogger()` returns error if home dir or log file creation fails
- Log calls: `utils.Log(format string, args ...interface{})` - printf-style formatting
- Cleanup: `defer utils.CloseLogger()` after initialization
- Only logs when debug mode enabled: `if *debug { utils.InitLogger(); defer utils.CloseLogger() }`
- Log statements in auth flow: `utils.Log("Starting authentication...")`, `utils.Log("Error loading config: %v", err)`

**No Structured Logging:**
- Logs are unstructured text with timestamp prefix: `[15:04:05.000] message`
- Session header: `--- Log session started at [RFC3339 timestamp] ---`

## Comments

**When to Comment:**
- Package-level documentation: Comprehensive multi-line doc comments for complex packages (required for public packages)
- Function-level documentation: All exported functions documented; unexported functions commented for complex logic
- Inline comments: Used to explain non-obvious logic, algorithm details, TODOs for future work
- Comments within loops/conditionals: Used for clarity when logic is complex

**Example Style:**
- Package doc: `// Package core provides core infrastructure for TGCP...`
- Function doc: `// EstimateCost returns a formatted string estimate of the hourly cost`
- Inline: `// 1. Find Credentials`, `// 2. Determine Project ID`, numbered steps for procedural logic
- Field comments: Documented inline: `UserEmail string  // Email of authenticated user`

**JSDoc/TSDoc:**
- Not applicable (Go project, not TypeScript/JavaScript)
- Go uses standard comment syntax only

## Function Design

**Size Guidelines:**
- Functions keep focused scope, typically 20-60 lines
- Large functions break down responsibilities
- Example: `Authenticate()` in `/Users/rk/Projects/tgcp/internal/core/auth.go` ~80 lines with numbered comment sections breaking logic
- Service `Update()` methods often 50-100 lines handling message type switching

**Parameters:**
- Keep to 4-5 parameters maximum
- Use receiver pattern for state: `(s *Service) Method()` over passing service pointer
- Context always first parameter in async-capable functions: `func(ctx context.Context, ...)`
- Group related parameters: `machine string, zone string` grouped together
- Structs for option-style configuration: `StandardTable(columns []table.Column, opts ...TableOption)` uses functional options

**Return Values:**
- Single return type (or one main type + error)
- Error always last: `(T, error)`
- Multiple related values returned as structs or separate fields
- Use named return values for clarity on complex returns (not consistently used)

**Receiver Methods:**
- Receiver by value for immutable operations: `(c Cache) Get()` for read-only
- Receiver by pointer for mutation: `(c *Cache) Set()` for state changes
- Receivers on public types always documented
- Short receiver names standard: `c`, `m`, `s`, `t`, `p`

## Module Design

**Exports:**
- Exported identifiers: UPPERCASE first letter (Go standard)
- Unexported identifiers: lowercase first letter (Go standard)
- Public interfaces exported: `type Service interface { ... }`
- Public structs exported: `type AuthState struct { ... }`
- Factory functions exported: `func NewCache() *Cache`

**Barrel Files (Aggregation):**
- No barrel exports observed (no `__init__.go` pattern)
- Each package serves single responsibility: `services/gce/`, `services/iam/`, etc.
- Interface definitions in separate file: `services/interface.go` defines `Service` interface

**Package Structure:**
- Clear responsibility boundaries: `internal/core`, `internal/ui`, `internal/services`, `internal/config`, `internal/utils`, `internal/styles`
- Services modular: Each GCP service is own package with consistent interface
- No circular imports observed
- Public API (interfaces) separated from implementation

## Concurrency & Goroutines

**Pattern - Synchronous by Default:**
- Most code is synchronous
- Bubbletea `tea.Cmd` used for async operations: return commands from `Update()` methods
- HTTP client configured with rate limiting and retry logic (transports handle concurrency)

**Thread Safety:**
- Mutex protection where needed: `Cache` uses `sync.RWMutex` for concurrent access
- Single logger file handle (not fully thread-safe but acceptable for debug-only logging)
- Token bucket rate limiter: `TokenBucket.mu sync.Mutex` protects token state

**Example - Mutex Usage in Cache:**
```go
func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// ... modify c.items
}
```

## Constants & Configuration

**Configuration Management:**
- YAML config file: `~/.tgcprc` loaded via `config.LoadConfig()`
- Default config struct: `/Users/rk/Projects/tgcp/internal/config/config.go` defines `Config`, `UIConfig`, `FeaturesConfig`
- Environment variable overrides: CLI flags take precedence (project flag, debug flag, version flag)
- Graceful defaults: `DefaultConfig()` returns working defaults if file missing

**Constant Patterns:**
- Cache TTL: `const CacheTTL = 30 * time.Second`
- View states: `const (ViewHome ViewMode = iota; ViewService)`
- Status constants: `const (StateRunning InstanceState = "RUNNING"; ...)` string constants for API mapping
- Hardcoded pricing: `var prices = map[string]float64 { ... }` (not a constant, but stable data)

## Testing Conventions

(See TESTING.md for detailed test patterns)

---

*Convention analysis: 2026-02-09*
