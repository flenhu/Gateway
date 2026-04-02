package provider

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/flenhu/Gateway/internal/models"
)

var (
	ErrRegistryProviderNil       = errors.New("provider is nil")
	ErrProviderAlreadyRegistered = errors.New("provider is already registered")
	ErrProviderNotRegistered     = errors.New("provider is not registered")
	ErrModelRequired             = errors.New("model is required")
	ErrNoProvidersForModel       = errors.New("no providers support requested model")
)

// Registry keeps provider lookup and routing concerns out of the HTTP handlers.
type Registry struct {
	mu        sync.RWMutex
	providers map[models.Provider]Provider
	health    map[models.Provider]bool
	order     []models.Provider
}

func NewRegistry(providers ...Provider) (*Registry, error) {
	registry := &Registry{
		providers: make(map[models.Provider]Provider, len(providers)),
		health:    make(map[models.Provider]bool, len(providers)),
		order:     make([]models.Provider, 0, len(providers)),
	}

	for _, p := range providers {
		if err := registry.Register(p); err != nil {
			return nil, err
		}
	}

	return registry, nil
}

func (r *Registry) Register(p Provider) error {
	if p == nil {
		return ErrRegistryProviderNil
	}

	name := p.Name()

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.providers[name]; exists {
		return fmt.Errorf("%w: %s", ErrProviderAlreadyRegistered, name)
	}

	r.providers[name] = p
	r.health[name] = true
	r.order = append(r.order, name)

	return nil
}

func (r *Registry) Get(name models.Provider) (Provider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.providers[name]
	return p, ok
}

func (r *Registry) Providers() []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]Provider, 0, len(r.order))
	for _, name := range r.order {
		out = append(out, r.providers[name])
	}

	return out
}

func (r *Registry) Models() []models.ModelMapping {
	r.mu.RLock()
	defer r.mu.RUnlock()

	allModels := make([]models.ModelMapping, 0)
	for _, name := range r.order {
		allModels = append(allModels, r.providers[name].Models()...)
	}

	return allModels
}

func (r *Registry) SupportsModel(model string) bool {
	model = strings.TrimSpace(model)
	if model == "" {
		return false
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, name := range r.order {
		if r.providers[name].SupportsModel(model) {
			return true
		}
	}

	return false
}

func (r *Registry) Resolve(model string, preferred models.Provider) (Provider, error) {
	candidates, err := r.Candidates(model, preferred)
	if err != nil {
		return nil, err
	}

	return candidates[0], nil
}

func (r *Registry) Candidates(model string, preferred models.Provider) ([]Provider, error) {
	model = strings.TrimSpace(model)
	if model == "" {
		return nil, ErrModelRequired
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	var healthy []Provider
	var unhealthy []Provider
	seen := make(map[models.Provider]struct{}, len(r.order))

	if preferred != "" {
		if preferredProvider, ok := r.providers[preferred]; ok && r.health[preferred] && preferredProvider.SupportsModel(model) {
			healthy = append(healthy, preferredProvider)
			seen[preferred] = struct{}{}
		}
	}

	for _, name := range r.order {
		if _, alreadyIncluded := seen[name]; alreadyIncluded {
			continue
		}

		current := r.providers[name]
		if !current.SupportsModel(model) {
			continue
		}

		if r.health[name] {
			healthy = append(healthy, current)
			continue
		}

		unhealthy = append(unhealthy, current)
	}

	candidates := append(healthy, unhealthy...)
	if len(candidates) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrNoProvidersForModel, model)
	}

	return candidates, nil
}

func (r *Registry) SetHealthy(name models.Provider, healthy bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.providers[name]; !exists {
		return fmt.Errorf("%w: %s", ErrProviderNotRegistered, name)
	}

	r.health[name] = healthy
	return nil
}

func (r *Registry) IsHealthy(name models.Provider) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	healthy, exists := r.health[name]
	return exists && healthy
}
