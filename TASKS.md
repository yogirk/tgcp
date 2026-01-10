# TGCP Codebase Tasks

**Generated:** 2026-01-06  
**Based on:** Comprehensive codebase review and comparison with CODEBASE_REVIEW.md

---

## ✅ Already Implemented (Verified)

The following improvements from CODEBASE_REVIEW.md have been **successfully implemented**:

1. ✅ **Service Registry Pattern** - Automated service registration in `internal/core/registry.go`
2. ✅ **Shared Error Component** - `components.RenderError()` used across all 17 services
3. ✅ **Shared Confirmation Component** - `components.RenderConfirmation()` standardized across services
4. ✅ **Loading Spinner Component** - `components.RenderSpinner()` provides consistent loading UX
5. ✅ **Project Switching Refactor** - `Reinit()` method implemented in all 17 services
6. ✅ **Table Component Standardization** - `StandardTable` component with Focus/Blur and window size handling
7. ✅ **Filter Component Enhancement** - Enhanced FilterModel with full lifecycle support

**Impact:** ~400+ lines of duplicate code eliminated, significantly improved maintainability.

---

## 🔴 High Priority Tasks

### 1. Lazy Service Initialization
**Status:** ✅ **COMPLETED** (2026-01-06)  
**Priority:** High  
**Impact:** Performance - Startup time, unnecessary API clients

**Implementation:**
- ✅ Modified `ServiceRegistry` to support lazy initialization
- ✅ Added `GetOrInitializeService()` method that initializes services on first access
- ✅ Updated `MainModel` with `getOrInitializeService()` helper method
- ✅ Updated all service access points to use lazy initialization
- ✅ Added initialization tracking to handle project switching correctly
- ✅ Services are now created at startup but not initialized until first access

**Files Modified:**
- `internal/core/registry.go` - Added lazy initialization with tracking
- `internal/ui/model.go` - Added helper method and updated all service access points

**Result:**
- Services are no longer initialized at startup
- Only services that are actually accessed get initialized
- Faster startup time (no 17 API clients created upfront)
- Proper handling of project switching with initialization tracking

---

### 2. Cache Size Limit and Eviction Policy
**Status:** ⚠️ **Not Implemented**  
**Priority:** High  
**Impact:** Memory - Unbounded growth with many projects/resources

**Current Issue:**
- `Cache` in `internal/core/cache.go` has no size limit
- Memory could grow unbounded with many projects/resources
- No eviction policy (only TTL-based expiration)

**Solution:**
- Implement LRU (Least Recently Used) cache or size-based eviction
- Add `MaxSize` configuration option
- Evict oldest items when limit reached

**Files to Modify:**
- `internal/core/cache.go` - Add LRU eviction logic

**Estimated Effort:** Medium (2 days)

---

### 3. Project List Caching with TTL
**Status:** ⚠️ **Not Implemented**  
**Priority:** High  
**Impact:** Data freshness - Projects cached forever

**Current Issue:**
- `ProjectManager.ListProjects()` caches projects in memory with no expiration
- Projects list never refreshes after first fetch
- Should use shared cache with TTL

**Solution:**
- Use `core.Cache` with TTL (e.g., 5 minutes) instead of in-memory slice
- Update `ProjectManager` to use cache

**Files to Modify:**
- `internal/core/projects.go` - Use cache with TTL

**Estimated Effort:** Low (1 day)

---

### 4. Request Deduplication
**Status:** ⚠️ **Not Implemented**  
**Priority:** High  
**Impact:** Efficiency - Duplicate API calls on rapid refresh

**Current Issue:**
- Multiple rapid refreshes could trigger duplicate API calls
- No mechanism to cancel or deduplicate in-flight requests

**Solution:**
- Implement request deduplication/cancellation
- Track in-flight requests by cache key
- Cancel previous request if new one starts for same resource

**Files to Modify:**
- `internal/core/cache.go` - Add request tracking
- Service fetch commands - Add cancellation support

**Estimated Effort:** Medium (2-3 days)

---

## 🟡 Medium Priority Tasks

### 5. Error Message Sanitization
**Status:** ⚠️ **Not Implemented**  
**Priority:** Medium  
**Impact:** Security - Error messages might leak internal details

**Current Issue:**
- Error messages display full error text via `m.Error.Error()`
- Could leak internal details, stack traces, or sensitive information

**Solution:**
- Sanitize error messages before display
- Strip stack traces, internal paths, sensitive data
- Keep user-friendly messages only

**Files to Modify:**
- `internal/ui/components/error.go` - Add sanitization function

**Estimated Effort:** Low (1 day)

---

### 6. Table Truncation Improvements
**Status:** ⚠️ **Partially Addressed**  
**Priority:** Medium  
**Impact:** UX - Better table display on small terminals

**Current Issue:**
- Window size sync happens, but could be more robust
- No minimum column widths
- No horizontal scrolling support

**Solution:**
- Add minimum column widths to `StandardTable`
- Implement horizontal scrolling for wide tables
- Better handling of small terminal sizes

**Files to Modify:**
- `internal/ui/components/table.go` - Enhance StandardTable

**Estimated Effort:** Low (1 day)

---

### 7. Help Overlay Discoverability
**Status:** ✅ **COMPLETED** (2026-01-06)  
**Priority:** Medium  
**Impact:** UX - Help screen not discoverable

**Implementation:**
- ✅ Added `?` key hint in status bar when help is available but not shown
- ✅ Help hint appears in home view when help overlay is not displayed
- ✅ Help hint appears in service view appended to service help text when help overlay is not displayed
- ✅ Hint is conditionally shown only when `ShowHelp` is false, keeping status bar clean when help is active

**Files Modified:**
- `internal/ui/model.go` - Updated dynamic help text logic to include `?:Help` hint conditionally

**Result:**
- Help overlay is now more discoverable
- Users can easily see that `?` key is available to toggle help
- Hint appears contextually in all views (home and service modes)
- Status bar remains clean when help overlay is already displayed

---

### 8. Unit Tests for Core Components
**Status:** ⚠️ **Minimal** (only 1 test file exists: `pricing_test.go`)  
**Priority:** Medium  
**Impact:** Maintainability - Important as codebase grows

**Current Issue:**
- No visible test files except `internal/services/gce/pricing_test.go`
- No test infrastructure for core components

**Solution:**
- Add unit tests for:
  - Cache logic (`internal/core/cache.go`)
  - Rate limiting (`internal/core/client.go`)
  - Retry logic (`internal/core/client.go`)
  - Service registry (`internal/core/registry.go`)

**Files to Create:**
- `internal/core/cache_test.go`
- `internal/core/client_test.go`
- `internal/core/registry_test.go`

**Estimated Effort:** Medium (3 days)

---

### 9. Integration Tests
**Status:** ⚠️ **Not Implemented**  
**Priority:** Medium  
**Impact:** Quality - Test service interactions

**Solution:**
- Create integration tests for:
  - API client interactions (with mocks)
  - Service initialization
  - Project switching
  - Cache invalidation

**Files to Create:**
- `internal/core/integration_test.go`
- `internal/services/integration_test.go`

**Estimated Effort:** Medium (3 days)

---

## 🟢 Low Priority Tasks

### 10. Package-Level Documentation
**Status:** ✅ **COMPLETED** (2026-01-06)  
**Priority:** Low  
**Impact:** Developer Experience

**Implementation:**
- ✅ Added comprehensive package-level documentation to `internal/core/client.go`
- ✅ Documented rate limiting algorithm (Token Bucket) with examples
- ✅ Documented retry logic (Exponential Backoff) with examples
- ✅ Enhanced service template with:
  - Modern filter pattern using `HandleFilterUpdate`
  - Detail view rendering example
  - Action command examples
  - Multiple tables pattern example
  - Better error handling examples

**Files Modified:**
- `internal/core/client.go` - Added package docs and function-level documentation
- `internal/services/service_template.go.txt` - Added examples and patterns

**Result:**
- Developers can now understand rate limiting and retry logic
- Service template includes practical examples for common patterns
- Better onboarding for new contributors

---

### 11. Architecture Diagrams
**Status:** ⚠️ **Not Implemented**  
**Priority:** Low  
**Impact:** Documentation

**Solution:**
- Create diagrams for:
  - Service registration flow
  - Message passing flow
  - Cache invalidation flow

**Files to Create:**
- `docs/architecture.md` (with diagrams)

**Estimated Effort:** Low (1 day)

---

### 12. Developer Guide
**Status:** ✅ **COMPLETED** (2026-01-06)  
**Priority:** Low  
**Impact:** Developer Experience

**Implementation:**
- ✅ Created comprehensive `docs/DEVELOPER_GUIDE.md` with:
  - Step-by-step guide for adding new services (8 steps)
  - Common patterns (7 patterns with code examples)
  - Pitfalls and best practices (5 common pitfalls, 5 best practices)
  - Debugging guide (common issues and solutions)
  - Architecture overview (lifecycle, message flow, component hierarchy)
- ✅ Enhanced `CONTRIBUTING.md` with:
  - Links to Developer Guide
  - Quick reference for adding services
  - Common patterns summary

**Files Created/Modified:**
- `docs/DEVELOPER_GUIDE.md` - Comprehensive 400+ line guide
- `CONTRIBUTING.md` - Enhanced with links and quick reference

**Result:**
- New developers can quickly understand how to add services
- Common patterns are documented with examples
- Debugging guide helps troubleshoot issues
- Better developer onboarding experience

---

### 13. Template Validation Script
**Status:** ⚠️ **Not Implemented**  
**Priority:** Low  
**Impact:** Code Quality

**Solution:**
- Create script to validate new services follow template pattern
- Check for required components (StandardTable, FilterModel, etc.)
- Verify Service interface implementation

**Files to Create:**
- `scripts/validate_service.sh` or `scripts/validate_service.go`

**Estimated Effort:** Low (1 day)

---

### 14. String Concatenation Optimization
**Status:** ⚠️ **Not Implemented**  
**Priority:** Low  
**Impact:** Performance (minor)

**Current Issue:**
- Some views build strings with `+` operator
- Could be optimized for large strings

**Solution:**
- Replace string concatenation with `strings.Builder` for large strings
- Profile to identify hotspots

**Files to Review:**
- Service view rendering code

**Estimated Effort:** Low (1 day)

---

### 15. Table Memoization
**Status:** ⚠️ **Not Implemented**  
**Priority:** Low  
**Impact:** Performance (minor)

**Current Issue:**
- Tables re-render on every update, even if data unchanged

**Solution:**
- Implement diff-based updates or memoization
- Only re-render when data actually changes

**Files to Modify:**
- `internal/ui/components/table.go` - Add memoization

**Estimated Effort:** Medium (2 days)

---

## Summary

### Task Count by Priority
- **High Priority:** 4 tasks
- **Medium Priority:** 5 tasks
- **Low Priority:** 6 tasks
- **Total:** 15 tasks

### Estimated Total Effort
- **High Priority:** ~7-9 days
- **Medium Priority:** ~9-11 days
- **Low Priority:** ~8-10 days
- **Total:** ~24-30 days

### Quick Wins (Low effort, high impact)
1. Project list caching with TTL (1 day)
2. Help overlay discoverability (0.5 days)
3. Error message sanitization (1 day)
4. Table truncation improvements (1 day)

---

## Notes

- All tasks are based on issues identified in CODEBASE_REVIEW.md or discovered during comprehensive review
- Priority is based on impact (performance, security, maintainability) and effort
- Some tasks may be combined or batched for efficiency
- Testing tasks (#8, #9) are important for long-term maintainability as the codebase grows
