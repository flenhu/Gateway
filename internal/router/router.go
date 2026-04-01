package router

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type serviceStatus struct {
	Service   string `json:"service"`
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

type modelsResponse struct {
	Models  []string `json:"models"`
	Message string   `json:"message,omitempty"`
}

type apiMessage struct {
	Message string `json:"message"`
}

func New() http.Handler {
	r := chi.NewRouter()

	r.Get("/", handleRoot)
	r.Get("/health", handleHealth)

	r.Route("/v1", func(r chi.Router) {
		r.Get("/models", handleModels)

		r.Group(func(r chi.Router) {
			// Auth and rate limiting middleware will attach here.
			r.Post("/chat/completions", handleChatCompletions)
			r.Get("/usage", handleUsage)
		})
	})

	return r
}

func handleRoot(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, serviceStatus{
		Service:   "llm-gateway",
		Status:    "running",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, serviceStatus{
		Service:   "llm-gateway",
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

func handleModels(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, modelsResponse{
		Models:  []string{},
		Message: "model registry not wired yet",
	})
}

func handleChatCompletions(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusNotImplemented, apiMessage{
		Message: "chat completions route is defined but not implemented yet",
	})
}

func handleUsage(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusNotImplemented, apiMessage{
		Message: "usage route is defined but not implemented yet",
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
