# TGCP Developer Guide

**Last Updated:** 2026-01-06  
**Version:** 1.0

This guide provides comprehensive instructions for developers working on TGCP, including how to add new services, common patterns, pitfalls to avoid, and debugging techniques.

---

## Table of Contents

1. [Getting Started](#getting-started)
2. [Adding a New Service](#adding-a-new-service)
3. [Common Patterns](#common-patterns)
4. [Pitfalls and Best Practices](#pitfalls-and-best-practices)
5. [Debugging Guide](#debugging-guide)
6. [Architecture Overview](#architecture-overview)

---

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Google Cloud SDK (`gcloud`) installed
- Basic understanding of:
  - Go programming
  - Bubbletea (TUI framework)
  - Google Cloud Platform APIs

### Development Setup

1. **Clone the repository:**
   ```bash
   git clone https://github.com/yourusername/tgcp.git
   cd tgcp
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Set up authentication:**
   ```bash
   gcloud auth application-default login
   ```

4. **Run the application:**
   ```bash
   go run ./cmd/tgcp
   ```

5. **Run in debug mode:**
   ```bash
   go run ./cmd/tgcp --debug
   # Debug logs written to: ~/.tgcp/debug.log
   ```

---

## Adding a New Service

### Step-by-Step Guide

#### Step 1: Create Service Directory

```bash
mkdir -p internal/services/newservice
cd internal/services/newservice
```

#### Step 2: Copy and Customize Service Template

```bash
cp ../service_template.go.txt newservice.go
```

The template provides a complete service implementation with:
- Standard table setup
- Filter functionality
- Error handling
- Loading states
- Confirmation dialogs
- Caching

#### Step 3: Implement Service Interface

All services must implement the `services.Service` interface:

```go
type Service interface {
    Name() string
    ShortName() string
    InitService(ctx context.Context, projectID string) error
    Reinit(ctx context.Context, projectID string) error
    Update(msg tea.Msg) (tea.Model, tea.Cmd)
    View() string
    HelpText() string
    Refresh() tea.Cmd
    Focus()
    Blur()
    Reset()
    IsRootView() bool
}
```

**Key Methods:**

- **`Name()`**: Full human-readable name (e.g., "Google Compute Engine")
- **`ShortName()`**: Command palette identifier (e.g., "gce")
- **`InitService()`**: Initialize API client with project ID
- **`Reinit()`**: Reinitialize when switching projects
- **`Update()`**: Handle Bubbletea messages (key presses, data updates, etc.)
- **`View()`**: Render the UI
- **`HelpText()`**: Context-aware help text for status bar
- **`Refresh()`**: Force data reload
- **`Focus()/Blur()`**: Handle input focus (visual highlighting)
- **`Reset()`**: Clear state when navigating away
- **`IsRootView()`**: Return true if at top-level list view

#### Step 4: Create API Client

Create `api.go` in your service directory:

```go
package newservice

import (
    "context"
    "google.golang.org/api/newservice/v1"
    "github.com/yogirk/tgcp/internal/core"
)

type Client struct {
    service *newservice.Service
}

func NewClient(ctx context.Context) (*Client, error) {
    client, err := core.NewHTTPClient(ctx, "https://www.googleapis.com/auth/cloud-platform")
    if err != nil {
        return nil, err
    }
    
    service, err := newservice.New(client)
    if err != nil {
        return nil, err
    }
    
    return &Client{service: service}, nil
}

func (c *Client) ListResources(projectID string) ([]Resource, error) {
    // Implement API call
    // ...
}
```

#### Step 5: Create Models

Create `models.go`:

```go
package newservice

type Resource struct {
    Name   string
    Status string
    ID     string
    // Add your resource fields
}
```

#### Step 6: Register Service

Add your service to `internal/ui/model.go` in the `registerAllServices()` function:

```go
registry.Register("newservice", func(cache *core.Cache) services.Service {
    return newservice.NewService(cache)
})
```

#### Step 7: Add Navigation Command

Add to `internal/core/navigation.go`:

```go
{
    Name:        "NewService: List Resources",
    Description: "View all resources in NewService",
    Action: func() Route {
        return Route{
            View:    ViewServiceList,
            Service: "newservice",
        }
    },
},
```

#### Step 8: Add Sidebar Item

Add to `internal/ui/components/sidebar.go`:

```go
{
    Name:     "NewService",
    ShortName: "newservice",
    Icon:     "📦",
},
```

### Estimated Time

- **Basic service (list only)**: 2-3 hours
- **Service with detail view**: 3-4 hours
- **Service with actions**: 4-6 hours

---

## Common Patterns

### Pattern 1: Filter Handling (Recommended)

Use `HandleFilterUpdate` for clean, maintainable filter logic:

```go
case tea.KeyMsg:
    if s.viewState == ViewList {
        result := components.HandleFilterUpdate(
            &s.filter,
            msg,
            s.items,
            func(items []Item, query string) []Item {
                return s.getFilteredItems(items, query)
            },
            s.updateTable,
        )

        if result.Handled {
            if result.Cmd != nil {
                return s, result.Cmd
            }
            if !result.ShouldContinue {
                return s, nil
            }
        }
    }
```

### Pattern 2: Data Fetching with Caching

Always use caching to reduce API calls:

```go
func (s *Service) fetchDataCmd(force bool) tea.Cmd {
    return func() tea.Msg {
        key := "newservice_data"
        
        // 1. Check cache first
        if !force && s.cache != nil {
            if val, found := s.cache.Get(key); found {
                if items, ok := val.([]Resource); ok {
                    return dataMsg(items)
                }
            }
        }
        
        // 2. API call
        if s.client == nil {
            return errMsg(fmt.Errorf("client not initialized"))
        }
        
        data, err := s.client.ListResources(s.projectID)
        if err != nil {
            return errMsg(err)
        }
        
        // 3. Update cache
        if s.cache != nil {
            s.cache.Set(key, data, CacheTTL)
        }
        
        return dataMsg(data)
    }
}
```

### Pattern 3: Error Display

Always use the shared error component:

```go
func (s *Service) View() string {
    if s.loading && len(s.items) == 0 {
        return components.RenderSpinner("Loading resources...")
    }
    if s.err != nil {
        return components.RenderError(s.err, s.Name(), "Resources")
    }
    // ... rest of view
}
```

### Pattern 4: Confirmation Dialogs

Use the shared confirmation component:

```go
if s.viewState == ViewConfirmation {
    return components.RenderConfirmation(
        s.pendingAction,      // "start", "stop", "delete"
        s.selectedItem.Name,  // Resource name
        "resource",           // Resource type
    )
}
```

### Pattern 5: Action Commands

Implement actions as tea.Cmd:

```go
func (s *Service) StartResourceCmd(resource Resource) tea.Cmd {
    return func() tea.Msg {
        if s.client == nil {
            return actionResultMsg{err: fmt.Errorf("client not initialized")}
        }
        
        err := s.client.StartResource(resource.ID)
        return actionResultMsg{err: err}
    }
}
```

### Pattern 6: Window Size Handling

Use StandardTable's built-in window size handling:

```go
case tea.WindowSizeMsg:
    s.table.HandleWindowSizeDefault(msg)
```

### Pattern 7: Background Refresh

Use ticker for automatic cache invalidation:

```go
func (s *Service) tick() tea.Cmd {
    return tea.Tick(CacheTTL, func(t time.Time) tea.Msg {
        return tickMsg(t)
    })
}

func (s *Service) Init() tea.Cmd {
    return s.tick()
}
```

---

## Pitfalls and Best Practices

### ❌ Common Pitfalls

1. **Forgetting to clear errors on Reset()**
   ```go
   // ❌ BAD
   func (s *Service) Reset() {
       s.viewState = ViewList
   }
   
   // ✅ GOOD
   func (s *Service) Reset() {
       s.viewState = ViewList
       s.err = nil  // Always clear errors!
   }
   ```

2. **Not checking client initialization**
   ```go
   // ❌ BAD
   data, err := s.client.ListResources()
   
   // ✅ GOOD
   if s.client == nil {
       return errMsg(fmt.Errorf("client not initialized"))
   }
   data, err := s.client.ListResources()
   ```

3. **Blocking the UI thread**
   ```go
   // ❌ BAD - blocks UI
   data, err := s.client.ListResources()
   s.items = data
   
   // ✅ GOOD - non-blocking
   return s, s.fetchDataCmd(true)
   ```

4. **Not using filtered items for selection**
   ```go
   // ❌ BAD - uses unfiltered items
   if idx := s.table.Cursor(); idx < len(s.items) {
       s.selectedItem = &s.items[idx]
   }
   
   // ✅ GOOD - uses filtered items
   items := s.getCurrentItems()
   if idx := s.table.Cursor(); idx < len(items) {
       s.selectedItem = &items[idx]
   }
   ```

5. **Forgetting to handle window resize**
   ```go
   // ❌ BAD - table won't resize
   case tea.WindowSizeMsg:
       // Nothing
   
   // ✅ GOOD
   case tea.WindowSizeMsg:
       s.table.HandleWindowSizeDefault(msg)
   ```

### ✅ Best Practices

1. **Always use shared components**
   - `components.RenderError()` for errors
   - `components.RenderSpinner()` for loading
   - `components.RenderConfirmation()` for confirmations
   - `components.NewStandardTable()` for tables

2. **Use appropriate cache TTL**
   - Fast-changing resources: 30 seconds
   - Medium-changing: 60 seconds
   - Slow-changing: 5 minutes

3. **Implement proper error handling**
   - Always check for nil clients
   - Return error messages, don't panic
   - Use `errMsg` type for error messages

4. **Keep views responsive**
   - Use `tea.Cmd` for async operations
   - Never block in `Update()` method
   - Show loading spinners during async operations

5. **Follow naming conventions**
   - Service name: `newservice` (lowercase, no spaces)
   - Short name: matches directory name
   - Message types: `dataMsg`, `errMsg`, `actionResultMsg`

---

## Debugging Guide

### Debug Mode

Run with `--debug` flag to enable debug logging:

```bash
go run ./cmd/tgcp --debug
```

Debug logs are written to `~/.tgcp/debug.log`.

### Common Issues

#### Issue 1: Service Not Appearing

**Symptoms:** Service doesn't show in sidebar or command palette

**Checklist:**
- [ ] Service registered in `registerAllServices()`
- [ ] Navigation command added
- [ ] Sidebar item added
- [ ] Service implements all interface methods

#### Issue 2: Service Not Initializing

**Symptoms:** Service shows error or doesn't load data

**Checklist:**
- [ ] `InitService()` implemented correctly
- [ ] API client created properly
- [ ] Project ID set correctly
- [ ] Check debug log for errors

#### Issue 3: Filter Not Working

**Symptoms:** Filter doesn't respond or doesn't filter items

**Checklist:**
- [ ] Using `HandleFilterUpdate` or manual filter handling
- [ ] `getFilteredItems()` implemented correctly
- [ ] Filter value checked in selection logic
- [ ] Table updated with filtered items

#### Issue 4: Table Not Rendering

**Symptoms:** Empty table or table doesn't show data

**Checklist:**
- [ ] Table rows set correctly: `s.table.SetRows(rows)`
- [ ] Window size handled: `s.table.HandleWindowSizeDefault(msg)`
- [ ] Table initialized with columns
- [ ] Data loaded successfully (check debug log)

#### Issue 5: Actions Not Working

**Symptoms:** Actions don't execute or show errors

**Checklist:**
- [ ] Action command returns `actionResultMsg`
- [ ] Error handling in action command
- [ ] Confirmation dialog shown before action
- [ ] Service refreshed after action

### Debugging Tips

1. **Add logging:**
   ```go
   import "github.com/yogirk/tgcp/internal/utils"
   
   utils.Log("Debug message: %v", value)
   ```

2. **Check state in Update():**
   ```go
   case tea.KeyMsg:
       if msg.String() == "d" { // Debug key
           utils.Log("State: %+v", s)
       }
   ```

3. **Verify message types:**
   - Check that message types match in `Update()` switch
   - Ensure commands return correct message types

4. **Test in isolation:**
   - Create a minimal test service
   - Test one feature at a time
   - Use debug mode to see what's happening

---

## Architecture Overview

### Service Lifecycle

1. **Registration**: Service registered in `registerAllServices()`
2. **Creation**: Service instance created (lazy initialization)
3. **Initialization**: `InitService()` called when first accessed
4. **Update Loop**: `Update()` handles all messages
5. **Rendering**: `View()` renders current state
6. **Project Switch**: `Reinit()` called when project changes

### Message Flow

```
User Input (Key Press)
    ↓
MainModel.Update()
    ↓
Service.Update()
    ↓
Command (tea.Cmd)
    ↓
Background Operation
    ↓
Message (tea.Msg)
    ↓
Service.Update() (handles result)
    ↓
View() (renders updated state)
```

### Component Hierarchy

```
MainModel
├── Sidebar
├── HomeMenu
├── StatusBar
├── Palette
└── CurrentService
    ├── StandardTable
    ├── FilterModel
    └── (Service-specific components)
```

### Caching Strategy

- **Cache Key Format**: `{service}_{resource}` (e.g., `gce_instances`)
- **TTL**: Service-specific (30s-5min)
- **Invalidation**: Automatic on TTL expiration or manual refresh
- **Storage**: In-memory map with thread-safe access

---

## Additional Resources

- **Service Template**: `internal/services/service_template.go.txt`
- **Codebase Review**: `CODEBASE_REVIEW.md`
- **Features Guide**: `FEATURES.md`
- **Contributing Guide**: `CONTRIBUTING.md`

---

## Getting Help

- Check existing services for examples
- Review the service template
- Read the codebase review for architecture details
- Open an issue on GitHub for questions

---

**Happy Coding! 🚀**
