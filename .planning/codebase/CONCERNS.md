# Codebase Concerns

**Analysis Date:** 2026-03-03

## Tech Debt

**Insufficient Test Coverage:**
- Issue: Only 1 test file (`pricing_test.go`) exists for entire codebase. No unit tests for core services, API interactions, or UI logic.
- Files: `internal/services/`, `internal/core/`, `internal/ui/`
- Impact: Changes to critical paths (authentication, service initialization, data parsing) could introduce regressions silently. New contributors risk breaking existing functionality.
- Fix approach: Implement comprehensive unit test suite targeting: authentication flow, service initialization/reinitialization, cache operations, error handling paths, and common data parsing scenarios. Aim for 60%+ coverage on critical paths.

**Context.Background() Used for Long-Running Operations:**
- Issue: Multiple locations use `context.Background()` for potentially long-running API calls and service operations that may need cancellation or timeout.
- Files:
  - `internal/ui/model.go` (lines 98, 407, 428, 446, 612, 718, 718)
  - `internal/services/dataflow/api.go`
  - `internal/services/gce/api.go` (line 38)
- Impact: No graceful way to interrupt operations if UI becomes unresponsive. User experience degradation if API calls hang.
- Fix approach: Pass a properly scoped context from the UI event loop. Implement timeout handling for each service API call. Add cancellation support when users navigate away or exit the application.

**Lack of Error Recovery Logging:**
- Issue: Errors are returned but not logged when `debug` flag is not enabled. Service failures silently update UI state without diagnostic context.
- Files: `internal/services/*/*.go` (all service implementations)
- Impact: Difficult to diagnose production issues. Users may not know why a service failed to load data.
- Fix approach: Implement structured error logging at service boundary that logs errors even in non-debug mode. Add error context (service name, operation, project ID) to all returned errors.

**Main Model State Complexity:**
- Issue: `MainModel` in `internal/ui/model.go` (922 lines) contains 25+ fields managing layout, services, components, and state. Complex state transitions during project switching and service navigation.
- Files: `internal/ui/model.go`
- Impact: Difficult to reason about state consistency. High risk of introducing state inconsistencies during modifications. Service reinitialization across all 21 services during project switch is potentially error-prone.
- Fix approach: Consider splitting into smaller sub-models (ProjectModel, ServiceModel, ComponentsModel). Add explicit state machine or transaction-like semantics for project switching. Add invariant checks.

**Synchronization Gaps in Service Reinitialization:**
- Issue: `ServiceRegistry.ReinitializeAll()` (lines 112-140) iterates and reinitializes services serially, with continue-on-error pattern. No validation that all services reinitialized successfully.
- Files: `internal/core/registry.go`
- Impact: If project switch partially fails, some services may operate with old project context while UI shows new project selected.
- Fix approach: Add atomic semantics - either all services reinitialize successfully or entire operation rolls back. Track and report which services failed.

## Known Bugs

**Potential Array Index Out of Bounds in Data Parsing:**
- Symptoms: Crash if API returns unexpected data structure (missing fields or empty arrays).
- Files:
  - `internal/services/gce/api.go` (lines 60-73, 92-97): `parts[len(parts)-1]` assumes non-empty split result
  - `internal/services/disks/disks.go`: Similar array access without length check
  - `internal/services/firestore/api.go`: Path element access without validation
- Trigger: Any GCP API response with unusual structure or missing expected fields
- Workaround: Currently protected by some length checks but inconsistently applied
- Fix approach: Add explicit length validation before array access. Use safe extraction helpers with fallback defaults.

**Version Comparison May Fail on Non-Semantic Versions:**
- Symptoms: Version display or update notification may behave unexpectedly for unusual version formats.
- Files: `internal/core/version.go` (lines 146-159)
- Trigger: Any GitHub release with non-standard tag format (e.g., "release-v1.2.3", "1.2.3.4")
- Workaround: Gracefully defaults to no-update, but doesn't log why
- Fix approach: Add validation and logging for non-standard version formats. Document expected version format.

**Unchecked Empty Email in Authentication:**
- Symptoms: User sees "Unknown" for user email if credentials file is missing and gcloud command fails silently.
- Files: `internal/core/auth.go` (lines 113-119): exec.Command may fail without reporting why
- Trigger: gcloud CLI not installed or not in PATH
- Workaround: Falls back to "Unknown", which is poor UX
- Fix approach: Add explicit error handling and user-facing message. Suggest installing gcloud CLI.

## Security Considerations

**Credentials File Access Without Validation:**
- Risk: Reading ADC credentials file directly without verifying file ownership or permissions could expose credentials if file has weak permissions.
- Files: `internal/core/auth.go` (lines 66-109)
- Current mitigation: Relies on OS file permission enforcement; uses standard ADC location
- Recommendations:
  - Add explicit checks for file permissions (should be readable only by current user)
  - Log file read operations when debug enabled
  - Document requirement for proper ADC setup

**No Input Validation on Project ID:**
- Risk: User-provided project ID (via flag or config) not validated. Could be passed directly to API calls.
- Files: `internal/core/auth.go` (lines 45-46), `cmd/tgcp/main.go` (lines 25, 52-54)
- Current mitigation: None explicitly visible
- Recommendations:
  - Add project ID format validation (alphanumeric, hyphens only)
  - Validate project exists before initializing services
  - Sanitize for any display contexts

**Exec Command Execution for gcloud:**
- Risk: `exec.Command("gcloud", ...)` could be vulnerable if PATH is manipulated or gcloud binary is compromised.
- Files: `internal/core/auth.go` (line 113)
- Current mitigation: Hardcoded command name, no shell interpretation
- Recommendations:
  - Consider using full path to gcloud if determinable
  - Add timeout (currently missing)
  - Log command execution when debug enabled

**HTTP Client for Version Checks:**
- Risk: GitHub API requests don't verify certificate or validate response format thoroughly.
- Files: `internal/core/version.go` (lines 70-84)
- Current mitigation: Basic error handling on HTTP status; JSON unmarshaling error handling present
- Recommendations:
  - Add certificate pinning or explicit CA verification for production
  - Validate release URL is legitimate before opening in browser
  - Add rate limiting to prevent abuse

**Console Output Without Escaping:**
- Risk: Service data displayed in TUI without escaping could allow terminal injection if source is untrusted.
- Files: `internal/services/*/*.go` (all service View methods use fmt.Sprintf)
- Current mitigation: Data comes from GCP APIs only, but TUI rendering is indirect
- Recommendations:
  - Consider escaping special characters in lipgloss rendering
  - Document assumption that GCP API responses are safe

## Performance Bottlenecks

**Cache Without Cleanup or Memory Bounds:**
- Problem: `Cache.items` map grows unbounded. No background cleanup of expired items.
- Files: `internal/core/cache.go`
- Cause: Only removes items on access; memory leak for expired items not accessed again
- Improvement path:
  - Add periodic cleanup goroutine (e.g., every 1 minute)
  - Add max size enforcement with LRU eviction
  - Track cache hit/miss metrics for optimization

**Serial Service Reinitialization on Project Switch:**
- Problem: Switching projects reinitializes 21 services sequentially (loop in `ReinitializeAll`). Each may make API calls.
- Files: `internal/core/registry.go` (lines 120-138)
- Cause: Sequential processing instead of concurrent
- Impact: UI blocks during project switch, visible lag for users
- Improvement path:
  - Use goroutines with WaitGroup to parallelize reinitialization
  - Implement timeout per service
  - Add progress feedback to user

**Synchronous Project Listing on Startup:**
- Problem: `Authenticate()` runs synchronously on main thread during startup. If ADC resolution is slow, startup is slow.
- Files: `cmd/tgcp/main.go` (line 58)
- Impact: Users see long startup delay if ADC lookup or gcloud command is slow
- Improvement path: Move authentication to async after TUI starts. Show loading state.

**All Service Clients Created on InitialModel:**
- Problem: `registry.InitializeAll()` creates instances of all 21 service clients immediately (though lazy initializes their state).
- Files: `internal/ui/model.go` (lines 88-98)
- Impact: Startup memory cost is proportional to number of services; all client libraries loaded
- Improvement path: Defer client creation until first access of each service

**Table Data Not Paginated:**
- Problem: Large result sets (e.g., 10,000 GCS buckets) loaded entirely into memory and rendered.
- Files: `internal/services/gcs/gcs.go`, `internal/services/cloudrun/run.go`, etc.
- Impact: High memory usage and slow rendering for large projects
- Improvement path: Implement pagination or virtual scrolling in table components

## Fragile Areas

**Service Registry State Machine During Project Switch:**
- Files: `internal/ui/model.go` (lines 407-430), `internal/core/registry.go` (lines 112-140)
- Why fragile: Complex multi-step process: validate project → reinitialize all services → update state. If any step partially fails, inconsistent state.
- Safe modification:
  - Add explicit state for "switching projects"
  - Make project switch atomic (all-or-nothing)
  - Add pre-switch validation
  - Test edge cases: invalid project, API timeout during switch
- Test coverage: Minimal - no tests for project switching logic

**Data Parsing in Service APIs:**
- Files: `internal/services/gce/api.go`, `internal/services/firestore/api.go`, `internal/services/net/api.go`
- Why fragile: Direct indexing and field access without defensive checks. Assumes consistent API response structure.
- Safe modification:
  - Use helper functions for safe extraction (e.g., `getOrDefault(parts, index, "Unknown")`)
  - Add nil checks for all optional fields
  - Log unexpected response structures when debug enabled
  - Add golden test data for different GCP API versions
- Test coverage: No tests; API changes could break parsing

**Filter and Search Implementation:**
- Files: `internal/ui/components/filter.go`, `internal/services/*/*.go`
- Why fragile: Custom filter session state across services. FilterModel used by multiple services with different data types.
- Safe modification:
  - Document FilterSession contract clearly
  - Add invariant checks in Update()
  - Test with rapid filter changes and data updates
- Test coverage: No filter tests

**Error Handling in Bubbletea Message Loop:**
- Files: `internal/ui/model.go` (Update method, ~300+ lines)
- Why fragile: Large switch statement with many message type handlers. Missing cases not obvious.
- Safe modification:
  - Add explicit panic/log for unhandled message types
  - Consider separating message handlers into helper functions
  - Add type-safe message variants (use custom types, not just error)
- Test coverage: No message handler tests

**Mouse Support Integration:**
- Files: `internal/ui/model.go`, `internal/ui/components/*`
- Why fragile: Mouse coordinates not validated. Could receive clicks outside rendered area.
- Safe modification:
  - Add bounds checking in all mouse handlers
  - Test with unusual terminal sizes
  - Handle terminal resize during mouse interaction
- Test coverage: No mouse interaction tests

## Scaling Limits

**Cache Memory Growth:**
- Current capacity: Unbounded (see Cache Without Cleanup)
- Limit: Memory exhaustion after extended use with many services
- Scaling path: Implement bounded cache with LRU eviction (suggested 100MB limit)

**Service Count Scalability:**
- Current capacity: 21 services registered and instantiated
- Limit: Each new service adds to startup cost and MainModel complexity
- Scaling path: Plugin-based service registration or separate service loader. Lazy loading of service definitions.

**Project Count in Project Manager:**
- Current capacity: Listed in memory (no pagination observed)
- Limit: Users with 1000+ projects may see memory/rendering issues
- Scaling path: Implement pagination in project selector. Cache only active project details.

**Table Rendering for Large Results:**
- Current capacity: All rows rendered in memory
- Limit: >1000 rows may cause lag
- Scaling path: Virtual scrolling or pagination. Implement --max-results flag per service.

## Dependencies at Risk

**Google Cloud Client Libraries (cloud.google.com/go/*):**
- Risk: Multiple versions of Cloud client libraries. Potential version conflicts with google.golang.org/api.
- Impact: API incompatibilities, authentication failures, missing features
- Migration plan: Pin versions explicitly. Add integration tests against actual GCP APIs.

**Charmbracelet Libraries (bubbletea, lipgloss, bubbles):**
- Risk: Active development; potential breaking changes in major versions
- Impact: UI rendering bugs, input handling issues
- Migration plan: Monitor releases. Test before upgrading. Pin to minor version.

**golang.org/x/oauth2:**
- Risk: ADC implementation may change in future oauth2 versions
- Impact: Authentication failures
- Migration plan: Use latest stable version. Add tests for auth flow. Follow golang.org security advisories.

## Missing Critical Features

**Project Switching Validation:**
- Problem: No validation that user has permissions to project before switching. UI may show project as selected but services fail to initialize.
- Blocks: Smooth multi-project workflows. Users see confusing errors after switching.

**Graceful Shutdown:**
- Problem: No mechanism to cancel in-flight requests on application exit
- Blocks: Clean shutdown. Potential data loss if writes were in progress.

**Audit Logging:**
- Problem: No record of user actions (which resources viewed, operations performed)
- Blocks: Security auditing, usage analytics, debugging user reports

**Concurrent Service Data Refresh:**
- Problem: Services refresh independently. No way to refresh all visible services at once.
- Blocks: Consistent multi-service view of current state

**Offline Mode:**
- Problem: No offline caching for basic resource browsing if connectivity lost
- Blocks: Resilience in flaky network conditions

## Test Coverage Gaps

**Service Initialization and Reinitialization:**
- What's not tested: InitService() method for all services. Reinit() behavior. Handling of initialization failures.
- Files: `internal/services/*/*.go` (all service implementations)
- Risk: Critical startup path untested. Service reinitialization during project switch untested.
- Priority: HIGH - These are core paths that every user exercises

**Error Handling and Recovery:**
- What's not tested: API call failures. Timeout scenarios. Partial response handling. How services behave when API is rate-limited.
- Files: `internal/services/*/*.go`
- Risk: Bugs in error handling silent until production incidents
- Priority: HIGH - Error paths are critical for reliability

**Authentication Flow:**
- What's not tested: ADC resolution with various credential types (user, service account). Behavior with missing/invalid credentials. Project ID resolution logic.
- Files: `internal/core/auth.go`
- Risk: Authentication issues hard to diagnose. Users may see cryptic errors.
- Priority: HIGH - Authentication is gating every user

**UI State Transitions:**
- What's not tested: Service switching. Project switching. Dialog/confirmation flows. Keyboard navigation edge cases.
- Files: `internal/ui/model.go`, `internal/ui/components/*`
- Risk: UI bugs hard to reproduce. State inconsistencies from rapid input
- Priority: MEDIUM - User-facing but less critical than APIs

**Data Parsing from GCP APIs:**
- What's not tested: Array access safety. Field presence validation. Handling of minimal/empty responses.
- Files: `internal/services/*/api.go`
- Risk: Crashes on unexpected API responses. Hard to debug with users' data.
- Priority: MEDIUM - Risk mitigated somewhat by API contract stability but still fragile

**Cache Operations:**
- What's not tested: Concurrent cache access under load. TTL expiration. Cache invalidation logic.
- Files: `internal/core/cache.go`
- Risk: Stale data served. Data corruption. Race conditions.
- Priority: MEDIUM - Concurrency issues manifest unpredictably

**Version Checking and Update Logic:**
- What's not tested: Version comparison across different formats. Update availability detection. Release note parsing.
- Files: `internal/core/version.go`
- Risk: Users miss updates or see false positive update notifications
- Priority: LOW - User-facing but not critical functionality

---

*Concerns audit: 2026-03-03*
