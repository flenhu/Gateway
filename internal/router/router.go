package router

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/flenhu/Gateway/internal/models"
	"github.com/flenhu/Gateway/internal/provider"
	"github.com/go-chi/chi/v5"
)

type Router struct {
	registry *provider.Registry
}

type serviceStatus struct {
	Service   string `json:"service"`
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

type healthResponse struct {
	Service   string                   `json:"service"`
	Status    string                   `json:"status"`
	Timestamp string                   `json:"timestamp"`
	Providers []providerHealthResponse `json:"providers"`
}

type providerHealthResponse struct {
	Name    models.Provider `json:"name"`
	Healthy bool            `json:"healthy"`
}

type modelsResponse struct {
	Models []models.ModelMapping `json:"models"`
}

type errorEnvelope struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

const maxCompletionRequestBodyBytes = 1 << 20

func New(registry *provider.Registry) http.Handler {
	if registry == nil {
		registry, _ = provider.NewRegistry()
	}

	app := &Router{registry: registry}

	r := chi.NewRouter()

	r.Get("/", app.handleRoot)
	r.Get("/health", app.handleHealth)

	r.Route("/v1", func(r chi.Router) {
		r.Get("/models", app.handleModels)

		r.Group(func(r chi.Router) {
			// Auth and rate limiting middleware will attach here.
			r.Post("/chat/completions", app.handleChatCompletions)
			r.Get("/usage", app.handleUsage)
		})
	})

	return r
}

func (rt *Router) handleRoot(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, serviceStatus{
		Service:   "llm-gateway",
		Status:    "running",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

func (rt *Router) handleHealth(w http.ResponseWriter, _ *http.Request) {
	providers := rt.registry.Providers()
	health := make([]providerHealthResponse, 0, len(providers))
	for _, p := range providers {
		health = append(health, providerHealthResponse{
			Name:    p.Name(),
			Healthy: rt.registry.IsHealthy(p.Name()),
		})
	}

	writeJSON(w, http.StatusOK, healthResponse{
		Service:   "llm-gateway",
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Providers: health,
	})
}

func (rt *Router) handleModels(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, modelsResponse{
		Models: rt.registry.Models(),
	})
}

func (rt *Router) handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxCompletionRequestBodyBytes)

	var req models.CompletionRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must contain a single JSON object")
		return
	}

	if err := validateCompletionRequest(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	if len(rt.registry.Providers()) == 0 {
		writeError(w, http.StatusServiceUnavailable, "no_providers_configured", "no providers are configured")
		return
	}

	candidates, err := rt.registry.Candidates(req.Model, req.PreferredProvider)
	if err != nil {
		status := http.StatusBadRequest
		code := "model_not_supported"

		if errors.Is(err, provider.ErrModelRequired) {
			code = "invalid_request"
		}

		writeError(w, status, code, err.Error())
		return
	}

	allowFallback := fallbackEnabled(req.FallbackEnabled)
	var lastErr error

	for index, current := range candidates {
		resp, err := current.Complete(r.Context(), &req)
		if err == nil {
			if resp.Provider == "" {
				resp.Provider = current.Name()
			}

			resp.GatewayMeta.Provider = current.Name()
			resp.GatewayMeta.FallbackUsed = index > 0
			writeJSON(w, http.StatusOK, resp)
			return
		}

		lastErr = err
		if !allowFallback || index == len(candidates)-1 {
			break
		}
	}

	if lastErr == nil {
		lastErr = errors.New("completion failed")
	}

	writeError(w, http.StatusBadGateway, "provider_request_failed", lastErr.Error())
}

func (rt *Router) handleUsage(w http.ResponseWriter, _ *http.Request) {
	writeError(w, http.StatusNotImplemented, "not_implemented", "usage route is defined but not implemented yet")
}

func validateCompletionRequest(req *models.CompletionRequest) error {
	if req == nil {
		return errors.New("request body is required")
	}

	if strings.TrimSpace(req.Model) == "" {
		return errors.New("model is required")
	}

	if len(req.Messages) == 0 {
		return errors.New("messages must not be empty")
	}

	for _, message := range req.Messages {
		if strings.TrimSpace(message.Content) == "" {
			return errors.New("message content must not be empty")
		}

		switch message.Role {
		case models.RoleSystem, models.RoleUser, models.RoleAssistant:
		default:
			return errors.New("message role must be one of: system, user, assistant")
		}
	}

	return nil
}

func fallbackEnabled(flag *bool) bool {
	return flag == nil || *flag
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorEnvelope{
		Error: apiError{
			Code:    code,
			Message: message,
		},
	})
}
