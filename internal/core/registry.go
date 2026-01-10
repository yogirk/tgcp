package core

import (
	"context"

	"github.com/yogirk/tgcp/internal/services"
)

// ServiceFactory is a function that creates a new service instance
type ServiceFactory func(*Cache) services.Service

// ServiceRegistry manages service registration and initialization
type ServiceRegistry struct {
	factories map[string]ServiceFactory
	cache     *Cache
}

// NewServiceRegistry creates a new service registry
func NewServiceRegistry(cache *Cache) *ServiceRegistry {
	return &ServiceRegistry{
		factories: make(map[string]ServiceFactory),
		cache:     cache,
	}
}

// Register registers a service factory with the given name
func (r *ServiceRegistry) Register(name string, factory ServiceFactory) {
	r.factories[name] = factory
}

// InitializeAll creates and initializes all registered services
func (r *ServiceRegistry) InitializeAll(ctx context.Context, projectID string) map[string]services.Service {
	svcMap := make(map[string]services.Service)
	for name, factory := range r.factories {
		svc := factory(r.cache)
		if projectID != "" {
			if err := svc.InitService(ctx, projectID); err != nil {
				// Log error but continue with other services
				// The service will be in the map but may not be fully initialized
				// Individual services should handle this gracefully
			}
		}
		svcMap[name] = svc
	}
	return svcMap
}

// ReinitializeAll reinitializes all services in the given map with a new project ID
func (r *ServiceRegistry) ReinitializeAll(ctx context.Context, projectID string, svcMap map[string]services.Service) {
	for _, svc := range svcMap {
		// Use the new Reinit() method for cleaner project switching
		if err := svc.Reinit(ctx, projectID); err != nil {
			// Log error but continue with other services
			// The service will remain in its previous state if reinit fails
			continue
		}
	}
}

// GetServiceNames returns a list of all registered service names
func (r *ServiceRegistry) GetServiceNames() []string {
	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	return names
}

// IsRegistered checks if a service is registered
func (r *ServiceRegistry) IsRegistered(name string) bool {
	_, exists := r.factories[name]
	return exists
}
