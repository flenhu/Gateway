export type Role = 'system' | 'user' | 'assistant'

export type Provider = 'groq' | 'google' | 'openai' | 'anthropic'

export type Tier = 'free' | 'starter' | 'pro' | 'enterprise'

export type Message = {
  role: Role
  content: string
}

export type Choice = {
  index: number
  message: Message
  finish_reason?: string
}

export type Usage = {
  prompt_tokens: number
  completion_tokens: number
  total_tokens: number
}

export type GatewayMeta = {
  latency_ms: number
  provider: Provider
  fallback_used: boolean
  rate_limit_remaining?: number
  cost_estimate_usd: number
}

export type CompletionResponse = {
  id: string
  object?: string
  created?: number
  model: string
  provider?: Provider
  choices: Choice[]
  usage: Usage
  gateway_meta: GatewayMeta
}

export type ModelMapping = {
  alias: string
  provider: Provider
  provider_model: string
  tier: Tier
}

export type CompletionInput = {
  model: string
  messages: Message[]
  preferredProvider?: Provider
  fallbackEnabled?: boolean
}

export type ApiError = {
  error: {
    code: string
    message: string
  }
}
