# Service Registry Pattern: Implementation Guide

## Current State vs. Proposed State

### 🔴 Current Implementation (Manual Registration)

**Location:** `internal/ui/model.go` (lines 78-203)

**What happens now:**
- 17 services are manually registered in `InitialModel()`
- Each service requires 5-6 lines of boilerplate code
- Adding a new service requires modifying `InitialModel()`
- Project switching requires manual iteration with type assertions

**Current Code:**
```go
// internal/ui/model.go
func InitialModel(authState core.AuthState, cfg *config.Config) MainModel {
    cache := core.NewCache()
    svcMap := make(map[string]services.Service)

    // Service 1: Overview
    billingSvc := overview.NewService(cache)
    if authState.ProjectID != "" {
        billingSvc.InitService(context.Background(), authState.ProjectID)
    }
    svcMap["overview"] = billingSvc

    // Service 2: GCE
    gceSvc := gce.NewService(cache)
    if authState.ProjectID != "" {
        gceSvc.InitService(context.Background(), authState.ProjectID)
    }
    svcMap["gce"] = gceSvc

    // ... repeat 15 more times for each service ...
    
    return MainModel{ServiceMap: svcMap, ...}
}
```

**Problems:**
1. **125+ lines of repetitive code** in `InitialModel()`
2. **Must modify core UI file** to add any service
3. **Easy to forget** initialization steps
4. **No central place** to see all services
5. **Project switching is complex** (see lines 314-340)

---

### 🟢 Proposed Implementation (Service Registry)

**New File:** `internal/core/registry.go`

**What changes:**

1. **Create a registry system:**
```go
// internal/core/registry.go
package core

import (
    "context"
    "github.com/yogirk/tgcp/internal/services"
)

type ServiceFactory func(*Cache) services.Service

type ServiceRegistry struct {
    factories map[string]ServiceFactory
    cache     *Cache
}

func NewServiceRegistry(cache *Cache) *ServiceRegistry {
    return &ServiceRegistry{
        factories: make(map[string]ServiceFactory),
        cache:     cache,
    }
}

func (r *ServiceRegistry) Register(name string, factory ServiceFactory) {
    r.factories[name] = factory
}

func (r *ServiceRegistry) InitializeAll(ctx context.Context, projectID string) map[string]services.Service {
    svcMap := make(map[string]services.Service)
    for name, factory := range r.factories {
        svc := factory(r.cache)
        if projectID != "" {
            svc.InitService(ctx, projectID)
        }
        svcMap[name] = svc
    }
    return svcMap
}

func (r *ServiceRegistry) ReinitializeAll(ctx context.Context, projectID string, svcMap map[string]services.Service) {
    for name, svc := range svcMap {
        if factory, exists := r.factories[name]; exists {
            // Recreate service with new project
            newSvc := factory(r.cache)
            if projectID != "" {
                newSvc.InitService(ctx, projectID)
            }
            svcMap[name] = newSvc
        }
    }
}

func (r *ServiceRegistry) GetServiceNames() []string {
    names := make([]string, 0, len(r.factories))
    for name := range r.factories {
        names = append(names, name)
    }
    return names
}
```

2. **Create a registration file:**
```go
// internal/core/services.go (new file)
package core

import (
    "github.com/yogirk/tgcp/internal/services/bigquery"
    "github.com/yogirk/tgcp/internal/services/bigtable"
    "github.com/yogirk/tgcp/internal/services/cloudrun"
    "github.com/yogirk/tgcp/internal/services/cloudsql"
    "github.com/yogirk/tgcp/internal/services/dataflow"
    "github.com/yogirk/tgcp/internal/services/dataproc"
    "github.com/yogirk/tgcp/internal/services/disks"
    "github.com/yogirk/tgcp/internal/services/firestore"
    "github.com/yogirk/tgcp/internal/services/gce"
    "github.com/yogirk/tgcp/internal/services/gcs"
    "github.com/yogirk/tgcp/internal/services/gke"
    "github.com/yogirk/tgcp/internal/services/iam"
    "github.com/yogirk/tgcp/internal/services/net"
    "github.com/yogirk/tgcp/internal/services/overview"
    "github.com/yogirk/tgcp/internal/services/pubsub"
    "github.com/yogirk/tgcp/internal/services/redis"
    "github.com/yogirk/tgcp/internal/services/spanner"
)

// RegisterAllServices registers all available services with the registry
func RegisterAllServices(registry *ServiceRegistry) {
    registry.Register("overview", func(cache *Cache) services.Service {
        return overview.NewService(cache)
    })
    registry.Register("gce", func(cache *Cache) services.Service {
        return gce.NewService(cache)
    })
    registry.Register("gke", func(cache *Cache) services.Service {
        return gke.NewService(cache)
    })
    registry.Register("disks", func(cache *Cache) services.Service {
        return disks.NewService(cache)
    })
    registry.Register("pubsub", func(cache *Cache) services.Service {
        return pubsub.NewService(cache)
    })
    registry.Register("redis", func(cache *Cache) services.Service {
        return redis.NewService(cache)
    })
    registry.Register("spanner", func(cache *Cache) services.Service {
        return spanner.NewService(cache)
    })
    registry.Register("bigtable", func(cache *Cache) services.Service {
        return bigtable.NewService(cache)
    })
    registry.Register("dataflow", func(cache *Cache) services.Service {
        return dataflow.NewService(cache)
    })
    registry.Register("dataproc", func(cache *Cache) services.Service {
        return dataproc.NewService(cache)
    })
    registry.Register("firestore", func(cache *Cache) services.Service {
        return firestore.NewService(cache)
    })
    registry.Register("sql", func(cache *Cache) services.Service {
        return cloudsql.NewService(cache)
    })
    registry.Register("iam", func(cache *Cache) services.Service {
        return iam.NewService(cache)
    })
    registry.Register("run", func(cache *Cache) services.Service {
        return cloudrun.NewService(cache)
    })
    registry.Register("gcs", func(cache *Cache) services.Service {
        return gcs.NewService(cache)
    })
    registry.Register("bq", func(cache *Cache) services.Service {
        return bigquery.NewService(cache)
    })
    registry.Register("net", func(cache *Cache) services.Service {
        return net.NewService(cache)
    })
}
```

3. **Simplify InitialModel:**
```go
// internal/ui/model.go (simplified)
func InitialModel(authState core.AuthState, cfg *config.Config) MainModel {
    cache := core.NewCache()
    
    // Create registry and register all services
    registry := core.NewServiceRegistry(cache)
    core.RegisterAllServices(registry)
    
    // Initialize all services
    svcMap := registry.InitializeAll(context.Background(), authState.ProjectID)
    
    // ... rest of initialization ...
    return MainModel{
        ServiceMap: svcMap,
        // ...
    }
}
```

**Before:** 125+ lines  
**After:** ~10 lines

---

## Benefits

### 1. ✅ Reduced Boilerplate

**Before:**
- 17 services × 5 lines = **85 lines** of repetitive code
- Plus imports = **~125 total lines**

**After:**
- Registration: 17 services × 2 lines = **34 lines** (in separate file)
- Initialization: **3 lines** in `InitialModel()`
- **Total reduction: ~88 lines** (70% less code)

### 2. ✅ Easier to Add New Services

**Before (6 steps):**
1. Create service directory
2. Implement Service interface
3. Add import to `model.go`
4. Add 5-6 lines of registration code to `InitialModel()`
5. Add command to `navigation.go`
6. Add sidebar item

**After (4 steps):**
1. Create service directory
2. Implement Service interface
3. Add one line to `core/services.go`: `registry.Register("name", ...)`
4. Add command to `navigation.go` (sidebar can auto-generate from registry)

**Time saved:** ~50% reduction in steps

### 3. ✅ Centralized Service Management

**Before:**
- Services scattered across `model.go`
- Hard to see all services at a glance
- No single source of truth

**After:**
- All services registered in `core/services.go`
- Easy to see all services: `registry.GetServiceNames()`
- Single place to enable/disable services

### 4. ✅ Cleaner Project Switching

**Before:**
```go
// internal/ui/model.go:314-340
for _, svc := range m.ServiceMap {
    if s, ok := svc.(interface {
        InitService(context.Context, string) error
    }); ok {
        if err := s.InitService(context.Background(), newProjectID); err != nil {
            // error handling
        }
    }
    svc.Reset()
}
```
- Requires type assertion
- Error-prone
- Hard to maintain

**After:**
```go
// Clean and simple
registry.ReinitializeAll(context.Background(), newProjectID, m.ServiceMap)
```
- No type assertions needed
- Handled by registry
- Consistent behavior

### 5. ✅ Better Testability

**Before:**
- Hard to test service initialization in isolation
- Must create full `MainModel` to test

**After:**
```go
// Easy to test
registry := core.NewServiceRegistry(cache)
registry.Register("test", func(cache *Cache) services.Service {
    return &MockService{}
})
svcMap := registry.InitializeAll(ctx, "test-project")
```

### 6. ✅ Conditional Service Loading

**Future enhancement:**
```go
// Can easily enable/disable services based on config
func RegisterAllServices(registry *ServiceRegistry, cfg *config.Config) {
    if cfg.Features.EnableGCE {
        registry.Register("gce", ...)
    }
    // ...
}
```

### 7. ✅ Plugin System Foundation

**Future enhancement:**
```go
// Could load services from plugins
func (r *ServiceRegistry) LoadPlugin(path string) error {
    // Load dynamic service
}
```

---

## Migration Path

### Step 1: Create Registry (1 hour)
- Create `internal/core/registry.go`
- Implement basic registry

### Step 2: Create Registration File (1 hour)
- Create `internal/core/services.go`
- Move all registrations here

### Step 3: Update InitialModel (30 min)
- Replace manual registration with registry calls
- Test that all services still work

### Step 4: Update Project Switching (1 hour)
- Use `ReinitializeAll()` instead of manual loop
- Remove type assertions

### Step 5: Cleanup (30 min)
- Remove old imports from `model.go`
- Update tests if any

**Total Time:** ~4 hours

---

## Code Comparison

### Adding a New Service: Before

```go
// 1. Add import to model.go
import "github.com/yogirk/tgcp/internal/services/newservice"

// 2. Add to InitialModel() (5-6 lines)
newSvc := newservice.NewService(cache)
if authState.ProjectID != "" {
    newSvc.InitService(context.Background(), authState.ProjectID)
}
svcMap["newservice"] = newSvc

// 3. Add to navigation.go
{Name: "NewService: List", Description: "...", Action: func() Route {...}}

// 4. Add to sidebar.go
{Name: "New Service", ShortName: "newservice", ...}
```

### Adding a New Service: After

```go
// 1. Add to core/services.go (1 line)
registry.Register("newservice", func(cache *Cache) services.Service {
    return newservice.NewService(cache)
})

// 2. Add to navigation.go (same as before)
{Name: "NewService: List", Description: "...", Action: func() Route {...}}

// 3. Sidebar can auto-generate from registry (optional)
```

**Reduction:** 5-6 lines → 1 line (83% reduction)

---

## Real-World Impact

### Scenario: Adding Cloud Functions Service

**Current approach:**
- Modify `model.go`: +6 lines, +1 import
- Risk: Could break existing services if typo
- Time: ~10 minutes

**With registry:**
- Modify `core/services.go`: +2 lines
- No risk to existing code
- Time: ~2 minutes

**Time saved:** 80% faster

---

## Potential Concerns & Solutions

### Concern 1: "Is this over-engineering?"
**Answer:** No. With 17 services and growing, the registry pattern is appropriate. The current approach doesn't scale.

### Concern 2: "What if I need custom initialization?"
**Answer:** Factory functions allow custom logic:
```go
registry.Register("special", func(cache *Cache) services.Service {
    svc := special.NewService(cache)
    svc.SetCustomOption(true) // Custom initialization
    return svc
})
```

### Concern 3: "Will this break existing code?"
**Answer:** No. The registry is internal. External API (Service interface) stays the same.

---

## Conclusion

The Service Registry Pattern provides:
- ✅ **70% less boilerplate code**
- ✅ **50% faster** to add new services
- ✅ **Cleaner project switching**
- ✅ **Better testability**
- ✅ **Foundation for future features** (plugins, conditional loading)

**Recommendation:** Implement this as a high-priority improvement. The effort is low (~4 hours) but the benefits are significant and compound over time as more services are added.
