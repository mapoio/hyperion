package hyperion

import "sync"

// InterceptorRegistry manages the collection of interceptors.
// It provides dynamic registration and thread-safe access to interceptors.
type InterceptorRegistry interface {
	// Register adds an interceptor to the registry
	Register(interceptor Interceptor)

	// GetAll returns all registered interceptors, sorted by order
	GetAll() []Interceptor
}

// interceptorRegistry is the default implementation of InterceptorRegistry
type interceptorRegistry struct {
	mu           sync.RWMutex
	interceptors []Interceptor
}

// NewInterceptorRegistry creates a new interceptor registry.
func NewInterceptorRegistry() InterceptorRegistry {
	return &interceptorRegistry{
		interceptors: make([]Interceptor, 0),
	}
}

// Register adds an interceptor to the registry
func (r *interceptorRegistry) Register(interceptor Interceptor) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.interceptors = append(r.interceptors, interceptor)
}

// GetAll returns all registered interceptors, sorted by order
func (r *interceptorRegistry) GetAll() []Interceptor {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return sorted copy
	return sortInterceptors(r.interceptors)
}
