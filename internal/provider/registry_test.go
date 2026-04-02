package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/flenhu/Gateway/internal/models"
)

type mockProvider struct {
	name   models.Provider
	models []models.ModelMapping
}

func (m *mockProvider) Name() models.Provider {
	return m.name
}

func (m *mockProvider) Complete(context.Context, *models.CompletionRequest) (*models.CompletionResponse, error) {
	return nil, errors.New("not implemented")
}

func (m *mockProvider) HealthCheck(context.Context) error {
	return nil
}

func (m *mockProvider) Models() []models.ModelMapping {
	out := make([]models.ModelMapping, len(m.models))
	copy(out, m.models)
	return out
}

func (m *mockProvider) SupportsModel(model string) bool {
	for _, mapping := range m.models {
		if mapping.Alias == model {
			return true
		}
	}

	return false
}

func TestRegistryModelsAggregatesRegisteredProviders(t *testing.T) {
	registry, err := NewRegistry(
		&mockProvider{
			name: models.ProviderGroq,
			models: []models.ModelMapping{
				{Alias: "llama-3.3-70b", Provider: models.ProviderGroq, ProviderModel: "llama-3.3-70b-versatile"},
			},
		},
		&mockProvider{
			name: models.ProviderGoogle,
			models: []models.ModelMapping{
				{Alias: "gemini-2.5-flash", Provider: models.ProviderGoogle, ProviderModel: "gemini-2.5-flash"},
			},
		},
	)
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}

	got := registry.Models()
	if len(got) != 2 {
		t.Fatalf("expected 2 models, got %d", len(got))
	}

	if got[0].Alias != "llama-3.3-70b" || got[1].Alias != "gemini-2.5-flash" {
		t.Fatalf("unexpected model order: %#v", got)
	}
}

func TestRegistryCandidatesPrefersHealthyPreferredProvider(t *testing.T) {
	registry, err := NewRegistry(
		&mockProvider{
			name: models.ProviderGroq,
			models: []models.ModelMapping{
				{Alias: "shared-model", Provider: models.ProviderGroq, ProviderModel: "groq-shared"},
			},
		},
		&mockProvider{
			name: models.ProviderGoogle,
			models: []models.ModelMapping{
				{Alias: "shared-model", Provider: models.ProviderGoogle, ProviderModel: "google-shared"},
			},
		},
	)
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}

	candidates, err := registry.Candidates("shared-model", models.ProviderGoogle)
	if err != nil {
		t.Fatalf("Candidates returned error: %v", err)
	}

	if len(candidates) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(candidates))
	}

	if candidates[0].Name() != models.ProviderGoogle {
		t.Fatalf("expected preferred provider first, got %s", candidates[0].Name())
	}
}

func TestRegistryCandidatesDeprioritizesUnhealthyProvider(t *testing.T) {
	registry, err := NewRegistry(
		&mockProvider{
			name: models.ProviderGroq,
			models: []models.ModelMapping{
				{Alias: "shared-model", Provider: models.ProviderGroq, ProviderModel: "groq-shared"},
			},
		},
		&mockProvider{
			name: models.ProviderGoogle,
			models: []models.ModelMapping{
				{Alias: "shared-model", Provider: models.ProviderGoogle, ProviderModel: "google-shared"},
			},
		},
	)
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}

	if err := registry.SetHealthy(models.ProviderGroq, false); err != nil {
		t.Fatalf("SetHealthy returned error: %v", err)
	}

	candidates, err := registry.Candidates("shared-model", "")
	if err != nil {
		t.Fatalf("Candidates returned error: %v", err)
	}

	if len(candidates) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(candidates))
	}

	if candidates[0].Name() != models.ProviderGoogle {
		t.Fatalf("expected healthy provider first, got %s", candidates[0].Name())
	}

	if candidates[1].Name() != models.ProviderGroq {
		t.Fatalf("expected unhealthy provider last, got %s", candidates[1].Name())
	}
}

func TestRegistryResolveReturnsErrorForUnsupportedModel(t *testing.T) {
	registry, err := NewRegistry(
		&mockProvider{
			name: models.ProviderGroq,
			models: []models.ModelMapping{
				{Alias: "llama-3.3-70b", Provider: models.ProviderGroq, ProviderModel: "llama-3.3-70b-versatile"},
			},
		},
	)
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}

	_, err = registry.Resolve("unknown-model", "")
	if !errors.Is(err, ErrNoProvidersForModel) {
		t.Fatalf("expected ErrNoProvidersForModel, got %v", err)
	}
}

func TestRegistryCandidatesReturnsErrorForBlankModel(t *testing.T) {
	registry, err := NewRegistry(
		&mockProvider{
			name: models.ProviderGroq,
			models: []models.ModelMapping{
				{Alias: "llama-3.3-70b", Provider: models.ProviderGroq, ProviderModel: "llama-3.3-70b-versatile"},
			},
		},
	)
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}

	_, err = registry.Candidates("   ", "")
	if !errors.Is(err, ErrModelRequired) {
		t.Fatalf("expected ErrModelRequired, got %v", err)
	}
}

func TestRegistryCandidatesIncludesUnhealthyPreferredProviderAsFallback(t *testing.T) {
	registry, err := NewRegistry(
		&mockProvider{
			name: models.ProviderGroq,
			models: []models.ModelMapping{
				{Alias: "shared-model", Provider: models.ProviderGroq, ProviderModel: "groq-shared"},
			},
		},
		&mockProvider{
			name: models.ProviderGoogle,
			models: []models.ModelMapping{
				{Alias: "shared-model", Provider: models.ProviderGoogle, ProviderModel: "google-shared"},
			},
		},
	)
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}

	if err := registry.SetHealthy(models.ProviderGoogle, false); err != nil {
		t.Fatalf("SetHealthy returned error: %v", err)
	}

	candidates, err := registry.Candidates("shared-model", models.ProviderGoogle)
	if err != nil {
		t.Fatalf("Candidates returned error: %v", err)
	}

	if len(candidates) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(candidates))
	}

	if candidates[0].Name() != models.ProviderGroq {
		t.Fatalf("expected healthy provider first, got %s", candidates[0].Name())
	}

	if candidates[1].Name() != models.ProviderGoogle {
		t.Fatalf("expected unhealthy preferred provider last, got %s", candidates[1].Name())
	}
}

func TestRegistryRegisterRejectsNilProvider(t *testing.T) {
	registry, err := NewRegistry()
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}

	if err := registry.Register(nil); !errors.Is(err, ErrRegistryProviderNil) {
		t.Fatalf("expected ErrRegistryProviderNil, got %v", err)
	}
}

func TestRegistryRegisterRejectsDuplicateProvider(t *testing.T) {
	registry, err := NewRegistry(
		&mockProvider{
			name: models.ProviderGroq,
			models: []models.ModelMapping{
				{Alias: "llama-3.3-70b", Provider: models.ProviderGroq, ProviderModel: "llama-3.3-70b-versatile"},
			},
		},
	)
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}

	err = registry.Register(&mockProvider{
		name: models.ProviderGroq,
		models: []models.ModelMapping{
			{Alias: "llama-3.1-8b", Provider: models.ProviderGroq, ProviderModel: "llama-3.1-8b-instant"},
		},
	})
	if !errors.Is(err, ErrProviderAlreadyRegistered) {
		t.Fatalf("expected ErrProviderAlreadyRegistered, got %v", err)
	}
}

func TestRegistrySetHealthyRejectsUnknownProvider(t *testing.T) {
	registry, err := NewRegistry(
		&mockProvider{
			name: models.ProviderGroq,
			models: []models.ModelMapping{
				{Alias: "llama-3.3-70b", Provider: models.ProviderGroq, ProviderModel: "llama-3.3-70b-versatile"},
			},
		},
	)
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}

	err = registry.SetHealthy(models.ProviderGoogle, false)
	if !errors.Is(err, ErrProviderNotRegistered) {
		t.Fatalf("expected ErrProviderNotRegistered, got %v", err)
	}
}
