package router

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flenhu/Gateway/internal/models"
	"github.com/flenhu/Gateway/internal/provider"
)

type testProvider struct {
	name        models.Provider
	models      []models.ModelMapping
	response    *models.CompletionResponse
	err         error
	callCount   int
	lastRequest *models.CompletionRequest
}

func (p *testProvider) Name() models.Provider {
	return p.name
}

func (p *testProvider) Complete(_ context.Context, req *models.CompletionRequest) (*models.CompletionResponse, error) {
	p.callCount++
	p.lastRequest = req
	if p.err != nil {
		return nil, p.err
	}

	return p.response, nil
}

func (p *testProvider) HealthCheck(context.Context) error {
	return nil
}

func (p *testProvider) Models() []models.ModelMapping {
	out := make([]models.ModelMapping, len(p.models))
	copy(out, p.models)
	return out
}

func (p *testProvider) SupportsModel(model string) bool {
	for _, mapping := range p.models {
		if mapping.Alias == model {
			return true
		}
	}

	return false
}

func TestHandleModelsAggregatesRegistryModels(t *testing.T) {
	registry, err := provider.NewRegistry(
		&testProvider{
			name: models.ProviderGroq,
			models: []models.ModelMapping{
				{Alias: "llama-3.3-70b", Provider: models.ProviderGroq, ProviderModel: "llama-3.3-70b-versatile"},
			},
		},
	)
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	rec := httptest.NewRecorder()

	New(registry).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp modelsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(resp.Models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(resp.Models))
	}

	if resp.Models[0].Alias != "llama-3.3-70b" {
		t.Fatalf("unexpected model alias: %s", resp.Models[0].Alias)
	}
}

func TestHandleChatCompletionsRoutesToMatchingProvider(t *testing.T) {
	groq := &testProvider{
		name: models.ProviderGroq,
		models: []models.ModelMapping{
			{Alias: "llama-3.3-70b", Provider: models.ProviderGroq, ProviderModel: "llama-3.3-70b-versatile"},
		},
		response: &models.CompletionResponse{
			ID:       "chatcmpl-1",
			Model:    "llama-3.3-70b-versatile",
			Provider: models.ProviderGroq,
			Choices: []models.Choice{
				{
					Index: 0,
					Message: models.Message{
						Role:    models.RoleAssistant,
						Content: "TCP is reliable.",
					},
				},
			},
		},
	}

	registry, err := provider.NewRegistry(groq)
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}

	body := `{"model":"llama-3.3-70b","messages":[{"role":"user","content":"Explain TCP"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	New(registry).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp models.CompletionResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if groq.callCount != 1 {
		t.Fatalf("expected provider to be called once, got %d", groq.callCount)
	}

	if resp.Provider != models.ProviderGroq {
		t.Fatalf("expected provider groq, got %s", resp.Provider)
	}

	if resp.GatewayMeta.Provider != models.ProviderGroq {
		t.Fatalf("expected gateway meta provider groq, got %s", resp.GatewayMeta.Provider)
	}

	if resp.GatewayMeta.FallbackUsed {
		t.Fatalf("expected fallback_used to be false")
	}
}

func TestHandleChatCompletionsFallsBackWhenEnabled(t *testing.T) {
	primary := &testProvider{
		name: models.ProviderGroq,
		models: []models.ModelMapping{
			{Alias: "shared-model", Provider: models.ProviderGroq, ProviderModel: "groq-shared"},
		},
		err: errors.New("upstream timeout"),
	}

	secondary := &testProvider{
		name: models.ProviderGoogle,
		models: []models.ModelMapping{
			{Alias: "shared-model", Provider: models.ProviderGoogle, ProviderModel: "google-shared"},
		},
		response: &models.CompletionResponse{
			ID:    "chatcmpl-2",
			Model: "google-shared",
			Choices: []models.Choice{
				{
					Index: 0,
					Message: models.Message{
						Role:    models.RoleAssistant,
						Content: "Fallback response.",
					},
				},
			},
		},
	}

	registry, err := provider.NewRegistry(primary, secondary)
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}

	body := `{"model":"shared-model","messages":[{"role":"user","content":"hi"}],"fallback_enabled":true}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	New(registry).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp models.CompletionResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if primary.callCount != 1 || secondary.callCount != 1 {
		t.Fatalf("expected both providers to be tried, got primary=%d secondary=%d", primary.callCount, secondary.callCount)
	}

	if !resp.GatewayMeta.FallbackUsed {
		t.Fatalf("expected fallback_used to be true")
	}

	if resp.GatewayMeta.Provider != models.ProviderGoogle {
		t.Fatalf("expected gateway meta provider google, got %s", resp.GatewayMeta.Provider)
	}
}

func TestHandleChatCompletionsStopsWhenFallbackDisabled(t *testing.T) {
	primary := &testProvider{
		name: models.ProviderGroq,
		models: []models.ModelMapping{
			{Alias: "shared-model", Provider: models.ProviderGroq, ProviderModel: "groq-shared"},
		},
		err: errors.New("upstream timeout"),
	}

	secondary := &testProvider{
		name: models.ProviderGoogle,
		models: []models.ModelMapping{
			{Alias: "shared-model", Provider: models.ProviderGoogle, ProviderModel: "google-shared"},
		},
		response: &models.CompletionResponse{
			ID:    "chatcmpl-3",
			Model: "google-shared",
		},
	}

	registry, err := provider.NewRegistry(primary, secondary)
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}

	body := `{"model":"shared-model","messages":[{"role":"user","content":"hi"}],"fallback_enabled":false}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	New(registry).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected status 502, got %d body=%s", rec.Code, rec.Body.String())
	}

	if primary.callCount != 1 {
		t.Fatalf("expected primary to be called once, got %d", primary.callCount)
	}

	if secondary.callCount != 0 {
		t.Fatalf("expected secondary not to be called, got %d", secondary.callCount)
	}
}

func TestHandleChatCompletionsRejectsInvalidRequest(t *testing.T) {
	registry, err := provider.NewRegistry()
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(`{"model":"","messages":[]}`))
	rec := httptest.NewRecorder()

	New(registry).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandleChatCompletionsReturnsServiceUnavailableWhenNoProvidersConfigured(t *testing.T) {
	registry, err := provider.NewRegistry()
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(`{"model":"llama-3.3-70b","messages":[{"role":"user","content":"hi"}]}`))
	rec := httptest.NewRecorder()

	New(registry).ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp errorEnvelope
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.Error.Code != "no_providers_configured" {
		t.Fatalf("expected no_providers_configured, got %s", resp.Error.Code)
	}
}

func TestHandleChatCompletionsRejectsUnknownModel(t *testing.T) {
	registry, err := provider.NewRegistry(
		&testProvider{
			name: models.ProviderGroq,
			models: []models.ModelMapping{
				{Alias: "llama-3.3-70b", Provider: models.ProviderGroq, ProviderModel: "llama-3.3-70b-versatile"},
			},
		},
	)
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(`{"model":"unknown","messages":[{"role":"user","content":"hi"}]}`))
	rec := httptest.NewRecorder()

	New(registry).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp errorEnvelope
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.Error.Code != "model_not_supported" {
		t.Fatalf("expected model_not_supported, got %s", resp.Error.Code)
	}
}
