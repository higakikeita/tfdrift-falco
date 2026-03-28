package provider

import (
	"context"
	"fmt"
	"sync"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	log "github.com/sirupsen/logrus"
)

// Registry manages registered cloud providers.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider
}

// NewRegistry creates a new empty provider registry.
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

// Register adds a provider to the registry.
// Returns an error if a provider with the same name is already registered.
func (r *Registry) Register(p Provider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := p.Name()
	if _, exists := r.providers[name]; exists {
		return fmt.Errorf("provider %q is already registered", name)
	}

	r.providers[name] = p
	log.Infof("Registered provider: %s (events: %d, resources: %d)",
		name, p.SupportedEventCount(), len(p.SupportedResourceTypes()))
	return nil
}

// Get returns the provider with the given name.
func (r *Registry) Get(name string) (Provider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.providers[name]
	return p, ok
}

// All returns all registered providers.
func (r *Registry) All() []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Provider, 0, len(r.providers))
	for _, p := range r.providers {
		result = append(result, p)
	}
	return result
}

// Names returns the names of all registered providers.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// Count returns the number of registered providers.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.providers)
}

// RouteEvent routes a Falco event to the appropriate provider based on source.
// Returns the parsed event and the provider name, or nil if no provider handles it.
func (r *Registry) RouteEvent(source string, fields map[string]string, rawEvent interface{}) (*types.Event, string) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.providers {
		if event := p.ParseEvent(source, fields, rawEvent); event != nil {
			return event, p.Name()
		}
	}

	return nil, ""
}

// GetDiscoverer returns the ResourceDiscoverer for the named provider, if supported.
func (r *Registry) GetDiscoverer(name string) (ResourceDiscoverer, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.providers[name]
	if !ok {
		return nil, false
	}
	d, ok := p.(ResourceDiscoverer)
	return d, ok
}

// GetComparator returns the StateComparator for the named provider, if supported.
func (r *Registry) GetComparator(name string) (StateComparator, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.providers[name]
	if !ok {
		return nil, false
	}
	c, ok := p.(StateComparator)
	return c, ok
}

// DiscoverAll runs resource discovery across all providers that support it.
func (r *Registry) DiscoverAll(ctx context.Context, opts DiscoveryOptions) (map[string][]*types.DiscoveredResource, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make(map[string][]*types.DiscoveredResource)
	for name, p := range r.providers {
		discoverer, ok := p.(ResourceDiscoverer)
		if !ok {
			log.Debugf("Provider %s does not support resource discovery, skipping", name)
			continue
		}

		resources, err := discoverer.DiscoverResources(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("discovery failed for provider %s: %w", name, err)
		}
		results[name] = resources
		log.Infof("Discovered %d resources from provider %s", len(resources), name)
	}

	return results, nil
}

// GetAllCapabilities returns capabilities for all registered providers.
func (r *Registry) GetAllCapabilities() map[string]Capabilities {
	r.mu.RLock()
	defer r.mu.RUnlock()

	caps := make(map[string]Capabilities)
	for name, p := range r.providers {
		caps[name] = GetCapabilities(p)
	}
	return caps
}
