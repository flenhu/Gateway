package provider

import (
	"context"

	"github.com/flenhu/Gateway/internal/models"
)

// Provider is the unified contract every model backend must satisfy.
type Provider interface {
	Name() models.Provider
	Complete(ctx context.Context, req *models.CompletionRequest) (*models.CompletionResponse, error)
	HealthCheck(ctx context.Context) error
	Models() []models.ModelMapping
	SupportsModel(model string) bool
}
