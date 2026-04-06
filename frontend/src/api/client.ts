import type {
  ApiError,
  CompletionInput,
  CompletionResponse,
  ModelMapping,
} from '../types/api'
import bestPickResponseJson from '../mocks/bestpick-response.json'
import completionGroqLlama70bJson from '../mocks/completion-groq-llama-70b.json'
import completionGroqLlama8bJson from '../mocks/completion-groq-llama-8b.json'
import modelsJson from '../mocks/models.json'

const USE_MOCKS = import.meta.env.VITE_USE_MOCKS === 'true'
const API_BASE = import.meta.env.VITE_API_BASE_URL

const mockModels = modelsJson.models as ModelMapping[]
const mockBestPickResponse = bestPickResponseJson as CompletionResponse
const mockCompletionByModel: Record<string, CompletionResponse> = {
  'llama-3.3-70b': completionGroqLlama70bJson as CompletionResponse,
  'llama-3.1-8b': completionGroqLlama8bJson as CompletionResponse,
}

export async function getModels(): Promise<ModelMapping[]> {
  if (USE_MOCKS) {
    return mockModels
  }

  const response = await fetchOrThrow(`${API_BASE}/v1/models`)
  const payload = await parseJsonOrThrow(response)

  if (!response.ok) {
    throw normalizeApiError(payload)
  }

  return (payload as { models: ModelMapping[] }).models
}

export async function postCompletion(
  input: CompletionInput,
): Promise<CompletionResponse> {
  if (USE_MOCKS) {
    if (input.model in mockCompletionByModel) {
      return mockCompletionByModel[input.model]
    }

    return mockBestPickResponse
  }

  const response = await fetchOrThrow(`${API_BASE}/v1/chat/completions`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      model: input.model,
      messages: input.messages,
      preferred_provider: input.preferredProvider,
      fallback_enabled: input.fallbackEnabled,
    }),
  })
  const payload = await parseJsonOrThrow(response)

  if (!response.ok) {
    throw normalizeApiError(payload)
  }

  return payload as CompletionResponse
}

async function fetchOrThrow(
  input: RequestInfo | URL,
  init?: RequestInit,
): Promise<Response> {
  try {
    return await fetch(input, init)
  } catch {
    throw {
      error: {
        code: 'network_error',
        message: 'Could not reach the gateway.',
      },
    } satisfies ApiError
  }
}

async function parseJsonOrThrow(response: Response): Promise<unknown> {
  try {
    return await response.json()
  } catch {
    throw {
      error: {
        code: 'invalid_json',
        message: 'The gateway returned invalid JSON.',
      },
    } satisfies ApiError
  }
}

function normalizeApiError(payload: unknown): ApiError {
  if (
    typeof payload === 'object' &&
    payload !== null &&
    'error' in payload &&
    typeof payload.error === 'object' &&
    payload.error !== null &&
    'code' in payload.error &&
    'message' in payload.error &&
    typeof payload.error.code === 'string' &&
    typeof payload.error.message === 'string'
  ) {
    return payload as ApiError
  }

  return {
    error: {
      code: 'unknown_error',
      message: 'The gateway returned an unexpected error response.',
    },
  }
}
