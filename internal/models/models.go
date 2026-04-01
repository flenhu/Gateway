package models

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

type Provider string

const (
	ProviderGroq      Provider = "groq"
	ProviderGoogle    Provider = "google"
	ProviderOpenAI    Provider = "openai"
	ProviderAnthropic Provider = "anthropic"
)

type Tier string

const (
	TierFree       Tier = "free"
	TierStarter    Tier = "starter"
	TierPro        Tier = "pro"
	TierEnterprise Tier = "enterprise"
)

type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

type CompletionRequest struct {
	Model             string    `json:"model"`
	Messages          []Message `json:"messages"`
	MaxTokens         *int      `json:"max_tokens,omitempty"`
	Temperature       *float64  `json:"temperature,omitempty"`
	Stream            *bool     `json:"stream,omitempty"`
	PreferredProvider Provider  `json:"preferred_provider,omitempty"`
	FallbackEnabled   *bool     `json:"fallback_enabled,omitempty"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason,omitempty"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type GatewayMeta struct {
	LatencyMs          int64    `json:"latency_ms"`
	Provider           Provider `json:"provider"`
	FallbackUsed       bool     `json:"fallback_used"`
	RateLimitRemaining *int     `json:"rate_limit_remaining,omitempty"`
	CostEstimateUSD    float64  `json:"cost_estimate_usd"`
}

type CompletionResponse struct {
	ID          string      `json:"id"`
	Object      string      `json:"object,omitempty"`
	Created     int64       `json:"created,omitempty"`
	Model       string      `json:"model"`
	Provider    Provider    `json:"provider,omitempty"`
	Choices     []Choice    `json:"choices"`
	Usage       Usage       `json:"usage"`
	GatewayMeta GatewayMeta `json:"gateway_meta"`
}

type ModelMapping struct {
	Alias         string   `json:"alias"`
	Provider      Provider `json:"provider"`
	ProviderModel string   `json:"provider_model"`
	Tier          Tier     `json:"tier"`
}
