# Testing Patterns

**Analysis Date:** 2026-03-03

## Test Framework

**Runner:**
- Framework: Standard Go `testing` package
- Version: Go 1.25.5 (from `go.mod`)
- Config: None (uses Go defaults via `go test`)

**Assertion Library:**
- Built-in `testing.T` - no external assertion framework
- Manual assertion via error messages: `if got != want { t.Errorf(...) }`

**Run Commands:**
```bash
go test ./...              # Run all tests in project
go test ./internal/...     # Run all tests in internal package
go test ./internal/services/gce/...  # Run tests for specific service
go test -v ./...           # Verbose output (test names + results)
go test -cover ./...       # Show coverage percentage
go test -coverprofile=coverage.out ./...  # Generate coverage report
```

**Test Coverage:**
- Current coverage: Minimal (only 1 test file found: `pricing_test.go`)
- No coverage target enforced or documented
- No coverage reporting in CI/CD observed

## Test File Organization

**Location Pattern:**
- Co-located in same directory as code: `internal/services/gce/pricing_test.go` alongside `pricing.go`
- One test file per feature or domain

**Naming Convention:**
- `[name]_test.go` pattern strictly followed
- Test functions: `Test[FunctionName]` pattern
- Example: `TestEstimateCost` tests `EstimateCost()` function

**Test File Structure:**
```
[service_dir]/
├── [feature].go
├── [feature]_test.go     # Tests for feature
├── [other_feature].go
└── [other_feature]_test.go
```

## Test Structure

**Suite Organization Pattern - Table-Driven Tests:**

From `internal/services/gce/pricing_test.go`:
```go
func TestEstimateCost(t *testing.T) {
	tests := []struct {
		name        string
		machineType string
		zone        string
		disks       []Disk
		want        string
	}{
		{
			"VM only",
			"e2-medium",
			"us-central1-a",
			nil,
			"$0.033/hr",
		},
		// ... more test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EstimateCost(tt.machineType, tt.zone, tt.disks); got != tt.want {
				t.Errorf("EstimateCost() = %v, want %v", got, tt.want)
			}
		})
	}
}
```

**Key Characteristics:**
- Subtests via `t.Run(tt.name, func(t *testing.T) { ... })`
- Each test case is a struct with fields: `name`, `machineType`, `zone`, `disks`, `want`
- Single assertion per subtest: `if got != want { t.Errorf(...) }`
- Test cases stored as `tests` slice of structs
- Loop through cases and run subtests

**Test Naming:**
- Test names descriptive of what's being tested: `"VM only"`, `"VM with Standard Disk"`, `"Unknown VM"`
- Function names match pattern: `Test[PublicFunctionName]`

## Mocking

**Framework:**
- No external mocking library detected
- No mocks observed in codebase (only 1 test file exists)

**Patterns NOT YET ESTABLISHED:**
Given minimal test coverage, mocking patterns not yet established. Based on code structure, expected patterns would be:
- Interface-based mocking: Services use interface types, could mock implementations
- Dependency injection for testability: `NewService(cache *core.Cache)` already accepts dependencies
- Potential mock types for: `Service` interface, HTTP clients, cache implementations

**Current Testing Approach:**
- Unit tests focus on pure functions: `EstimateCost()` is stateless
- No async/concurrent testing observed
- No network mocking in observed tests

## Fixtures and Test Data

**Test Data Pattern:**

Inline struct definition in test file:
```go
tests := []struct {
	name        string
	machineType string
	zone        string
	disks       []Disk
	want        string
}{
	{
		"VM only",
		"e2-medium",
		"us-central1-a",
		nil,
		"$0.033/hr",
	},
	// ...
}
```

**Fixture Location:**
- Inline in test functions (no separate fixture files)
- Test data structs defined within test function
- No `testdata/` directory or fixture files observed

**No Factory Functions:**
- No dedicated factory pattern for test data
- Structs created inline in test cases
- For services, constructors like `NewService()` used directly in tests (not tested in observed file)

## Coverage

**Requirements:**
- None enforced or documented
- No CI/CD checks for coverage observed

**View Coverage:**
```bash
# Generate coverage report
go test -cover ./...

# Generate detailed coverage report file
go test -coverprofile=coverage.out ./...

# View in HTML
go tool cover -html=coverage.out
```

**Current State:**
- Very low coverage: Only `pricing_test.go` exists (1 test file)
- Many packages untested: `auth.go`, `cache.go`, `client.go`, `navigation.go`, `registry.go`, `version.go`, all UI components, all other services
- Critical areas with no tests: Rate limiting logic, retry logic, cache TTL, authentication flow, UI model updates

## Test Types

**Unit Tests:**
- Scope: Single function in isolation
- Approach: Pure function testing with table-driven tests
- State: Stateless functions preferred (like `EstimateCost()`)
- Example: `TestEstimateCost` tests pricing calculation with various machine types and disk configurations

**Integration Tests:**
- Status: Not observed in codebase
- Would test: Service initialization, API interaction, cache behavior
- Pattern not yet established

**E2E Tests:**
- Framework: Not used
- Would require: Terminal UI interaction testing (complex with Bubbletea)
- Not implemented in current codebase

**UI Testing:**
- Status: Not tested
- Reason: Bubbletea TUI components difficult to test without full terminal context
- Alternative: Visual testing (manual verification)

## Common Patterns (Not Yet Established)

**Async Testing:**
- Status: Pattern not established (no async tests in current suite)
- Expected approach for `tea.Cmd` testing: Would need to execute commands and check resulting messages
- No examples in codebase

**Error Testing:**
- Status: Partially tested in `pricing_test.go`
- Example: Tests `"Unknown VM"` case which returns `"N/A"` (error condition)
- Pattern: Include error cases in table-driven test cases, check for expected error strings
- Future improvement needed for: Authentication errors, API errors, configuration errors

**Context Testing:**
- Status: No context timeout or cancellation tests observed
- Would be important for: `Authenticate()`, API calls, rate limiting waits
- Not implemented in current test suite

## Test Dependencies & Test Double Strategies

**What's Tested:**
- Pure functions: `EstimateCost()` - no external dependencies
- Simple lookups: Machine type pricing from static map

**What's NOT Tested (Opportunities):**
- Authentication flow: `Authenticate()` function requires Google credentials
- API interactions: All service methods that call Google Cloud APIs
- Caching behavior: `Cache.Set()`, `Cache.Get()`, TTL expiration
- Rate limiting: `TokenBucket` rate limiter
- Retry logic: `RetryTransport` exponential backoff
- UI model updates: All Bubbletea message handling
- Configuration loading: YAML parsing, default values

**Recommended Test Double Strategy (Not Yet Implemented):**
1. Interface-based mocking for `Service` implementations
2. Mock HTTP client for API testing: Could wrap `http.RoundTripper` interface
3. Fakable cache for testing eviction and TTL
4. Stub credentials for auth flow testing
5. Table-driven tests with nil/empty inputs for edge cases

## Test Execution Workflow

**Run All Tests:**
```bash
go test ./...
```

**Run Specific Service Tests:**
```bash
go test ./internal/services/gce/...
go test ./internal/services/iam/...
```

**Run with Verbose Output:**
```bash
go test -v ./...
```

**Run Single Test:**
```bash
go test ./internal/services/gce/... -run TestEstimateCost
```

**Expected Output Format:**
```
ok  	github.com/yogirk/tgcp/internal/services/gce	0.234s
```

## Test Maintenance

**Test Patterns to Maintain:**
1. Table-driven test structure for multiple scenarios
2. Subtest pattern via `t.Run()` for named test cases
3. Clear test names describing the scenario
4. Single assertion per test case
5. Arrange-Act-Assert structure

**Test File Location:**
- Always co-locate with implementation: `file.go` with `file_test.go`
- Maintain same package: Test code in same package as implementation (not `_test` package)

**When to Add Tests:**
- New public functions should have unit tests
- Bug fixes should add regression tests
- Complex calculation logic should have edge case tests
- Error handling paths need error case tests

## Test Configuration

**Go Test Configuration:**
- Uses Go standard `testing` package
- No external config files needed
- Flags passed to `go test` command directly

**Example Workflow (From CONTRIBUTING.md):**
```bash
go run ./cmd/tgcp                    # Run locally
go run ./cmd/tgcp --debug            # Debug mode
go fmt ./...                         # Format code
go vet ./internal/...                # Run linter
go test ./...                        # Run tests
```

---

*Testing analysis: 2026-03-03*
