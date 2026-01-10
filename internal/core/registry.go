package core

import (
	"context"
	"sync"

	"github.com/yogirk/tgcp/internal/services"
)

// ServiceFactory is a function that creates a new service instance
type ServiceFactory func(*Cache) services.Service

// ServiceRegistry manages service registration and lazy initialization
type ServiceRegistry struct {
	factories      map[string]ServiceFactory
	cache          *Cache
	services       map[string]services.Service // Lazily initialized services
	initialized    map[string]string            // Maps service name to projectID it was initialized with
	mu             sync.RWMutex                 // Protects services map and initialized map
	projectID      string                       // Current project ID for lazy initialization
}

// NewServiceRegistry creates a new service registry
func NewServiceRegistry(cache *Cache) *ServiceRegistry {
	return &ServiceRegistry{
		factories:   make(map[string]ServiceFactory),
		cache:       cache,
		services:    make(map[string]services.Service),
		initialized: make(map[string]string),
	}
}

// Register registers a service factory with the given name
func (r *ServiceRegistry) Register(name string, factory ServiceFactory) {
	r.factories[name] = factory
}

// InitializeAll creates all registered services but does NOT initialize them
// Services are initialized lazily on first access via GetOrInitializeService
func (r *ServiceRegistry) InitializeAll(ctx context.Context, projectID string) map[string]services.Service {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.projectID = projectID
	svcMap := make(map[string]services.Service)
	
	// Create service instances but don't initialize them yet
	for name, factory := range r.factories {
		svc := factory(r.cache)
		svcMap[name] = svc
		r.services[name] = svc
	}
	
	return svcMap
}

// GetOrInitializeService gets a service from the map, initializing it lazily if needed
// This is the key method for lazy initialization - services are only initialized when first accessed
func (r *ServiceRegistry) GetOrInitializeService(ctx context.Context, name string) (services.Service, error) {
	r.mu.RLock()
	svc, exists := r.services[name]
	projectID := r.projectID
	initProjectID, isInitialized := r.initialized[name]
	r.mu.RUnlock()
	
	if !exists {
		// Service not found - check if it's registered
		r.mu.RLock()
		factory, registered := r.factories[name]
		r.mu.RUnlock()
		
		if !registered {
			return nil, nil // Service not registered
		}
		
		// Create the service
		r.mu.Lock()
		svc = factory(r.cache)
		r.services[name] = svc
		r.mu.Unlock()
	}
	
	// Check if service needs initialization or reinitialization
	if projectID != "" {
		// If service was initialized with a different project ID, we need to reinit
		if isInitialized && initProjectID != projectID {
			// Project changed - use Reinit to properly reset and reinitialize
			if err := svc.Reinit(ctx, projectID); err != nil {
				return svc, err
			}
			// Update initialized tracking
			r.mu.Lock()
			r.initialized[name] = projectID
			r.mu.Unlock()
		} else if !isInitialized {
			// Service not yet initialized - initialize it
			if err := svc.InitService(ctx, projectID); err != nil {
				return svc, err // Return service even if init fails, let caller handle error
			}
			// Mark as initialized
			r.mu.Lock()
			r.initialized[name] = projectID
			r.mu.Unlock()
		}
		// If already initialized with same projectID, no action needed
	}
	
	return svc, nil
}

// ReinitializeAll reinitializes all initialized services with a new project ID
func (r *ServiceRegistry) ReinitializeAll(ctx context.Context, projectID string, svcMap map[string]services.Service) {
	r.mu.Lock()
	r.projectID = projectID
	// Clear initialization tracking - services will be reinitialized on next access
	// or we can reinit them now if they're already in the map
	r.mu.Unlock()
	
	// Reinitialize all services that have been created (lazy or not)
	for name, svc := range svcMap {
		// Check if this service was initialized
		r.mu.RLock()
		_, wasInitialized := r.initialized[name]
		r.mu.RUnlock()
		
		if wasInitialized {
			// Use the new Reinit() method for cleaner project switching
			if err := svc.Reinit(ctx, projectID); err != nil {
				// Log error but continue with other services
				// The service will remain in its previous state if reinit fails
				continue
			}
			// Update initialization tracking
			r.mu.Lock()
			r.initialized[name] = projectID
			r.mu.Unlock()
		}
		// If service wasn't initialized, it will be initialized lazily on next access
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
