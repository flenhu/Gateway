package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/flenhu/Gateway/internal/models"
)

const (
	groqBaseURL = "https://api.groq.com/openai/v1"
)

var (
	ErrGroqAPIKeyMissing  = errors.New("groq api key is missing")
	ErrCompletionNil      = errors.New("completion request is nil")
	ErrModelNotSupported  = errors.New("model is not supported by provider")
	ErrGroqUpstreamFailed = errors.New("groq upstream request failed")
)

type GroqProvider struct {
	apiKey        string
	baseURL       string
	httpClient    *http.Client
	modelMappings []models.ModelMapping
}

type groqChatCompletionRequest struct {
	Model               string        `json:"model"`
	Messages            []groqMessage `json:"messages"`
	MaxCompletionTokens *int          `json:"max_completion_tokens,omitempty"`
	Temperature         *float64      `json:"temperature,omitempty"`
	Stream              *bool         `json:"stream,omitempty"`
}

type groqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type groqChatCompletionResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []groqChoice           `json:"choices"`
	Usage   groqUsage              `json:"usage"`
	XGroq   map[string]interface{} `json:"x_groq,omitempty"`
}

type groqChoice struct {
	Index        int         `json:"index"`
	Message      groqMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

type groqUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type groqErrorResponse struct {
	Error groqError `json:"error"`
}

type groqError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

func NewGroqProvider(apiKey string, timeout time.Duration) *GroqProvider {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	return &GroqProvider{
		apiKey:     strings.TrimSpace(apiKey),
		baseURL:    groqBaseURL,
		httpClient: &http.Client{Timeout: timeout},
		modelMappings: []models.ModelMapping{
			{
				Alias:         "llama-3.3-70b",
				Provider:      models.ProviderGroq,
				ProviderModel: "llama-3.3-70b-versatile",
				Tier:          models.TierFree,
			},
			{
				Alias:         "llama-3.1-8b",
				Provider:      models.ProviderGroq,
				ProviderModel: "llama-3.1-8b-instant",
				Tier:          models.TierFree,
			},
		},
	}
}

func (p *GroqProvider) Name() models.Provider {
	return models.ProviderGroq
}

func (p *GroqProvider) Models() []models.ModelMapping {
	modelsCopy := make([]models.ModelMapping, len(p.modelMappings))
	copy(modelsCopy, p.modelMappings)
	return modelsCopy
}

func (p *GroqProvider) SupportsModel(model string) bool {
	_, ok := p.lookupModel(model)
	return ok
}

func (p *GroqProvider) Complete(ctx context.Context, req *models.CompletionRequest) (*models.CompletionResponse, error) {
	if req == nil {
		return nil, ErrCompletionNil
	}

	if strings.TrimSpace(p.apiKey) == "" {
		return nil, ErrGroqAPIKeyMissing
	}

	providerModel, ok := p.lookupModel(req.Model)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrModelNotSupported, req.Model)
	}

	payload, err := json.Marshal(groqChatCompletionRequest{
		Model:               providerModel,
		Messages:            toGroqMessages(req.Messages),
		MaxCompletionTokens: req.MaxTokens,
		Temperature:         req.Temperature,
		Stream:              req.Stream,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal groq request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create groq request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	started := time.Now()
	httpResp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send groq request: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		var upstreamErr groqErrorResponse
		if err := json.NewDecoder(httpResp.Body).Decode(&upstreamErr); err == nil && strings.TrimSpace(upstreamErr.Error.Message) != "" {
			return nil, fmt.Errorf("%w: %s", ErrGroqUpstreamFailed, upstreamErr.Error.Message)
		}

		return nil, fmt.Errorf("%w: status %d", ErrGroqUpstreamFailed, httpResp.StatusCode)
	}

	var groqResp groqChatCompletionResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&groqResp); err != nil {
		return nil, fmt.Errorf("decode groq response: %w", err)
	}

	return &models.CompletionResponse{
		ID:       groqResp.ID,
		Object:   groqResp.Object,
		Created:  groqResp.Created,
		Model:    groqResp.Model,
		Provider: models.ProviderGroq,
		Choices:  toGatewayChoices(groqResp.Choices),
		Usage: models.Usage{
			PromptTokens:     groqResp.Usage.PromptTokens,
			CompletionTokens: groqResp.Usage.CompletionTokens,
			TotalTokens:      groqResp.Usage.TotalTokens,
		},
		GatewayMeta: models.GatewayMeta{
			LatencyMs:       time.Since(started).Milliseconds(),
			Provider:        models.ProviderGroq,
			FallbackUsed:    false,
			CostEstimateUSD: estimateGroqCostUSD(providerModel, groqResp.Usage),
		},
	}, nil
}

func (p *GroqProvider) HealthCheck(ctx context.Context) error {
	if strings.TrimSpace(p.apiKey) == "" {
		return ErrGroqAPIKeyMissing
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"/models", nil)
	if err != nil {
		return fmt.Errorf("create groq health request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	httpResp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("send groq health request: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return fmt.Errorf("%w: health status %d", ErrGroqUpstreamFailed, httpResp.StatusCode)
	}

	return nil
}

func (p *GroqProvider) lookupModel(alias string) (string, bool) {
	for _, mapping := range p.modelMappings {
		if mapping.Alias == alias {
			return mapping.ProviderModel, true
		}
	}

	return "", false
}

func toGroqMessages(messages []models.Message) []groqMessage {
	out := make([]groqMessage, 0, len(messages))
	for _, message := range messages {
		out = append(out, groqMessage{
			Role:    string(message.Role),
			Content: message.Content,
		})
	}
	return out
}

func toGatewayChoices(choices []groqChoice) []models.Choice {
	out := make([]models.Choice, 0, len(choices))
	for _, choice := range choices {
		out = append(out, models.Choice{
			Index: choice.Index,
			Message: models.Message{
				Role:    models.Role(choice.Message.Role),
				Content: choice.Message.Content,
			},
			FinishReason: choice.FinishReason,
		})
	}
	return out
}

func estimateGroqCostUSD(model string, usage groqUsage) float64 {
	var promptPerMillion float64
	var completionPerMillion float64

	switch model {
	case "llama-3.3-70b-versatile":
		promptPerMillion = 0.59
		completionPerMillion = 0.79
	case "llama-3.1-8b-instant":
		promptPerMillion = 0.05
		completionPerMillion = 0.08
	default:
		return 0
	}

	promptCost := (float64(usage.PromptTokens) / 1_000_000) * promptPerMillion
	completionCost := (float64(usage.CompletionTokens) / 1_000_000) * completionPerMillion
	return promptCost + completionCost
}
