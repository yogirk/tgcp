# Shared Error Component Implementation Plan

## Overview

Create a reusable error display component to standardize error presentation across all services, improving UX consistency and reducing code duplication.

## Current State Analysis

### Current Error Patterns

**Pattern 1: Simple string formatting (most common)**
```go
// internal/services/gce/gce.go:358-360
if s.err != nil {
    return fmt.Sprintf("Error: %v", s.err)
}
```

**Pattern 2: Service-specific messages**
```go
// internal/services/cloudrun/run.go:353-355
if s.err != nil {
    return fmt.Sprintf("Error fetching Cloud Run services: %v", s.err)
}
```

**Pattern 3: Inline error checks**
```go
// internal/services/dataflow/views.go:15-17
if s.err != nil {
    return fmt.Sprintf("Error: %v", s.err)
}
```

### Problems with Current Approach

1. **Inconsistent styling** - Plain text, no visual hierarchy
2. **No actionable information** - Just error message, no context
3. **Poor visual design** - No icons, borders, or color coding
4. **Code duplication** - Same pattern repeated in ~17 services
5. **No error recovery hints** - Doesn't suggest what to do next

---

## Proposed Solution

### Component Design

Create `internal/ui/components/error.go` with:
- Consistent visual styling (error icon, colored border, formatted message)
- Support for different error types (API errors, network errors, permission errors)
- Actionable suggestions (retry, check permissions, etc.)
- Keyboard shortcuts for recovery actions

### Visual Design

```
┌─────────────────────────────────────────────────────────┐
│  ⚠  Error Loading Instances                            │
│                                                         │
│  Failed to fetch GCE instances:                        │
│  googleapi: Error 403: Insufficient permissions         │
│                                                         │
│  💡 Suggestions:                                        │
│     • Check IAM permissions for compute.instances.list  │
│     • Verify project ID is correct                      │
│     • Try refreshing with 'r'                           │
│                                                         │
│  [r] Retry  [q] Back                                    │
└─────────────────────────────────────────────────────────┘
```

---

## Implementation Steps

### Step 1: Create Error Component (30 min)

**File:** `internal/ui/components/error.go`

```go
package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/styles"
)

// ErrorModel represents an error display component
type ErrorModel struct {
	Error       error
	Title       string      // e.g., "Error Loading Instances"
	ServiceName string      // e.g., "GCE"
	Suggestions []string    // Helpful suggestions
	Width       int
	Height      int
}

// NewErrorModel creates a new error component
func NewErrorModel(err error, title, serviceName string) ErrorModel {
	return ErrorModel{
		Error:       err,
		Title:       title,
		ServiceName: serviceName,
		Suggestions: generateSuggestions(err, serviceName),
	}
}

// Update handles messages (for future: retry button, etc.)
func (m ErrorModel) Update(msg tea.Msg) (ErrorModel, tea.Cmd) {
	// For now, error component is static
	// Future: Add retry button, dismiss button, etc.
	return m, nil
}

// View renders the error component
func (m ErrorModel) View() string {
	if m.Error == nil {
		return ""
	}

	// Error icon and title
	header := lipgloss.JoinHorizontal(
		lipgloss.Left,
		styles.ErrorStyle.Render("⚠ "),
		styles.ErrorStyle.Bold(true).Render(m.Title),
	)

	// Error message
	errorMsg := m.Error.Error()
	// Wrap long error messages
	if m.Width > 0 {
		maxWidth := m.Width - 10
		if len(errorMsg) > maxWidth {
			errorMsg = errorMsg[:maxWidth-3] + "..."
		}
	}

	// Suggestions section
	var suggestions string
	if len(m.Suggestions) > 0 {
		suggestionLines := make([]string, len(m.Suggestions))
		for i, suggestion := range m.Suggestions {
			suggestionLines[i] = fmt.Sprintf("  • %s", suggestion)
		}
		suggestions = lipgloss.JoinVertical(
			lipgloss.Left,
			styles.SubtleStyle.Render("💡 Suggestions:"),
			strings.Join(suggestionLines, "\n"),
		)
	}

	// Help text
	helpText := styles.HelpStyle.Render("[r] Retry  [q] Back")

	// Combine all parts
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		styles.ValueStyle.Render("Failed: "+errorMsg),
		"",
		suggestions,
		"",
		helpText,
	)

	// Wrap in styled box
	box := styles.BoxStyle.Copy().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorError).
		Padding(1, 2).
		Width(80).
		Render(content)

	return box
}

// generateSuggestions creates helpful suggestions based on error type
func generateSuggestions(err error, serviceName string) []string {
	errStr := err.Error()
	suggestions := []string{}

	// Permission errors
	if strings.Contains(errStr, "403") || 
	   strings.Contains(errStr, "permission") || 
	   strings.Contains(errStr, "Insufficient") {
		suggestions = append(suggestions,
			fmt.Sprintf("Check IAM permissions for %s", serviceName),
			"Verify your account has the required roles",
		)
	}

	// Network errors
	if strings.Contains(errStr, "network") || 
	   strings.Contains(errStr, "timeout") ||
	   strings.Contains(errStr, "connection") {
		suggestions = append(suggestions,
			"Check your internet connection",
			"Verify GCP API is accessible",
		)
	}

	// Not found errors
	if strings.Contains(errStr, "404") || 
	   strings.Contains(errStr, "not found") {
		suggestions = append(suggestions,
			"Verify the resource exists",
			"Check project ID is correct",
		)
	}

	// Rate limit errors
	if strings.Contains(errStr, "429") || 
	   strings.Contains(errStr, "rate limit") {
		suggestions = append(suggestions,
			"API rate limit reached",
			"Wait a moment and try again",
		)
	}

	// Generic fallback
	if len(suggestions) == 0 {
		suggestions = append(suggestions,
			"Try refreshing with 'r'",
			"Check project configuration",
		)
	}

	return suggestions
}

// SetSize updates the component size
func (m *ErrorModel) SetSize(width, height int) {
	m.Width = width
	m.Height = height
}
```

### Step 2: Create Helper Function (15 min)

**File:** `internal/ui/components/error.go` (add to same file)

```go
// RenderError is a convenience function for services to render errors
func RenderError(err error, serviceName, resourceType string) string {
	title := fmt.Sprintf("Error Loading %s", resourceType)
	model := NewErrorModel(err, title, serviceName)
	return model.View()
}
```

### Step 3: Update Service Template (10 min)

**File:** `internal/services/service_template.go.txt`

Replace:
```go
if s.err != nil {
    return fmt.Sprintf("Error: %v", s.err)
}
```

With:
```go
if s.err != nil {
    return components.RenderError(s.err, s.Name(), "Items")
}
```

### Step 4: Update All Services (2-3 hours)

Update each service's `View()` method:

**Example: GCE Service**
```go
// internal/services/gce/gce.go
import "github.com/yogirk/tgcp/internal/ui/components"

func (s *Service) View() string {
    if s.loading {
        return "Loading instances..."
    }
    if s.err != nil {
        return components.RenderError(s.err, s.Name(), "Instances")
    }
    // ... rest of view
}
```

**Services to update:**
1. ✅ GCE (`internal/services/gce/gce.go`)
2. ✅ Cloud Run (`internal/services/cloudrun/run.go`)
3. ✅ Dataflow (`internal/services/dataflow/views.go`)
4. ✅ Pub/Sub (`internal/services/pubsub/views.go`)
5. ✅ GKE (`internal/services/gke/gke.go`)
6. ✅ Cloud SQL (`internal/services/cloudsql/cloudsql.go`)
7. ✅ GCS (`internal/services/gcs/gcs.go`)
8. ✅ BigQuery (`internal/services/bigquery/bigquery.go`)
9. ✅ IAM (`internal/services/iam/iam.go`)
10. ✅ Disks (`internal/services/disks/disks.go`)
11. ✅ Redis (`internal/services/redis/redis.go`)
12. ✅ Spanner (`internal/services/spanner/spanner.go`)
13. ✅ Bigtable (`internal/services/bigtable/bigtable.go`)
14. ✅ Firestore (`internal/services/firestore/firestore.go`)
15. ✅ Dataproc (`internal/services/dataproc/dataproc.go`)
16. ✅ Networking (`internal/services/net/net.go`)
17. ✅ Overview (`internal/services/overview/overview.go`)

### Step 5: Enhanced Error Types (Optional, 30 min)

Add support for categorized errors:

```go
// ErrorCategory helps provide better suggestions
type ErrorCategory int

const (
	ErrorCategoryUnknown ErrorCategory = iota
	ErrorCategoryPermission
	ErrorCategoryNetwork
	ErrorCategoryNotFound
	ErrorCategoryRateLimit
	ErrorCategoryValidation
)

func CategorizeError(err error) ErrorCategory {
	errStr := err.Error()
	if strings.Contains(errStr, "403") || strings.Contains(errStr, "permission") {
		return ErrorCategoryPermission
	}
	// ... more categories
	return ErrorCategoryUnknown
}
```

### Step 6: Add Error Recovery Actions (Optional, 1 hour)

Make error component interactive:

```go
type ErrorModel struct {
	// ... existing fields
	ShowRetry bool
	RetryFunc func() tea.Cmd
}

func (m ErrorModel) Update(msg tea.Msg) (ErrorModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "r" && m.ShowRetry && m.RetryFunc != nil {
			return m, m.RetryFunc()
		}
	}
	return m, nil
}
```

---

## Migration Strategy

### Phase 1: Create Component (Day 1, Morning)
- ✅ Create `error.go` component
- ✅ Add to `components` package
- ✅ Test with one service (GCE)

### Phase 2: Update Services (Day 1, Afternoon)
- ✅ Update all 17 services
- ✅ Update service template
- ✅ Test each service

### Phase 3: Polish (Day 1, Evening)
- ✅ Improve error categorization
- ✅ Add more suggestions
- ✅ Test edge cases

---

## Testing Checklist

- [ ] Error displays correctly with short messages
- [ ] Error displays correctly with long messages (wraps)
- [ ] Permission errors show IAM suggestions
- [ ] Network errors show connection suggestions
- [ ] 404 errors show "not found" suggestions
- [ ] Rate limit errors show wait suggestions
- [ ] Component works in narrow terminals
- [ ] Component works in wide terminals
- [ ] All 17 services use the component
- [ ] Service template updated

---

## Benefits

### Before
```go
// 17 services × 3 lines = 51 lines of duplicate code
if s.err != nil {
    return fmt.Sprintf("Error: %v", s.err)
}
```

### After
```go
// 17 services × 1 line = 17 lines, all consistent
if s.err != nil {
    return components.RenderError(s.err, s.Name(), "Resources")
}
```

**Benefits:**
- ✅ **67% less code** (51 lines → 17 lines)
- ✅ **Consistent UX** across all services
- ✅ **Better error messages** with suggestions
- ✅ **Easier maintenance** - fix once, works everywhere
- ✅ **Professional appearance** with icons and styling

---

## Future Enhancements

1. **Interactive Error Recovery**
   - Retry button that calls service.Refresh()
   - Dismiss button
   - Copy error to clipboard

2. **Error Logging**
   - Log errors to debug file
   - Track error frequency
   - Error analytics

3. **Contextual Help**
   - Link to relevant GCP documentation
   - Show IAM role requirements
   - Display troubleshooting steps

4. **Error Grouping**
   - Group similar errors
   - Show error count
   - Batch recovery

---

## Estimated Time

- **Step 1:** 30 minutes (Create component)
- **Step 2:** 15 minutes (Helper function)
- **Step 3:** 10 minutes (Update template)
- **Step 4:** 2-3 hours (Update all services)
- **Step 5:** 30 minutes (Optional enhancements)
- **Step 6:** 1 hour (Optional interactive features)

**Total:** ~4-5 hours for basic implementation
**Total with enhancements:** ~6-7 hours

---

## Files to Create/Modify

### New Files
- `internal/ui/components/error.go` (new)

### Modified Files
- `internal/services/service_template.go.txt`
- `internal/services/gce/gce.go`
- `internal/services/cloudrun/run.go`
- `internal/services/dataflow/views.go`
- `internal/services/pubsub/views.go`
- `internal/services/gke/gke.go`
- `internal/services/cloudsql/cloudsql.go`
- `internal/services/gcs/gcs.go`
- `internal/services/bigquery/bigquery.go`
- `internal/services/iam/iam.go`
- `internal/services/disks/disks.go`
- `internal/services/redis/redis.go`
- `internal/services/spanner/spanner.go`
- `internal/services/bigtable/bigtable.go`
- `internal/services/firestore/firestore.go`
- `internal/services/dataproc/dataproc.go`
- `internal/services/net/net.go`
- `internal/services/overview/overview.go`

**Total:** 1 new file, 18 modified files

---

## Example Usage

### Before
```go
func (s *Service) View() string {
    if s.err != nil {
        return fmt.Sprintf("Error: %v", s.err)
    }
    return s.table.View()
}
```

### After
```go
import "github.com/yogirk/tgcp/internal/ui/components"

func (s *Service) View() string {
    if s.err != nil {
        return components.RenderError(s.err, s.Name(), "Instances")
    }
    return s.table.View()
}
```

---

## Success Criteria

✅ All services display errors consistently  
✅ Error messages are actionable with suggestions  
✅ Visual design is professional and clear  
✅ Code duplication reduced by 67%  
✅ Component is reusable and maintainable  
✅ No breaking changes to existing functionality  

---

*Ready to implement? This is a low-risk, high-value improvement that will significantly improve the user experience.*
