import { useState } from 'react'
import { getModels, postCompletion } from '../api/client'
import { PromptInput } from '../components/PromptInput'
import { ResponseCard } from '../components/ResponseCard'
import type { CompletionResponse } from '../types/api'

export function CompareView() {
  const [prompt, setPrompt] = useState('')
  const [responses, setResponses] = useState<CompletionResponse[]>([])
  const [isLoading, setIsLoading] = useState(false)

  async function handleSubmit() {
    if (!prompt.trim() || isLoading) {
      return
    }

    setIsLoading(true)

    try {
      const models = await getModels()
      const results = await Promise.all(
        models.map((model) =>
          postCompletion({
            model: model.alias,
            messages: [{ role: 'user', content: prompt }],
          }),
        ),
      )

      setResponses(results)
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <section className="mode-panel">
      <div className="mode-panel__copy">
        <p className="mode-panel__label">Comparison Route</p>
        <h2>Compare</h2>
        <p>
          Submit one prompt to every available model and compare the results
          side by side.
        </p>
      </div>

      <PromptInput
        value={prompt}
        onChange={setPrompt}
        onSubmit={handleSubmit}
        disabled={isLoading}
      />

      {responses.length > 0 && (
        <div className="mode-panel__details">
          {responses.map((response) => (
            <ResponseCard
              key={`${response.provider ?? response.gateway_meta.provider}-${response.model}-${response.id}`}
              response={response}
            />
          ))}
        </div>
      )}
    </section>
  )
}
