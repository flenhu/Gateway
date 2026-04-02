#!/usr/bin/env bash

set -euo pipefail

gateway_url="${GATEWAY_URL:-http://127.0.0.1:8080}"
model="${GATEWAY_MODEL:-llama-3.1-8b}"
prompt="${GATEWAY_PROMPT:-Reply with exactly: manual run ok}"
curl_args=(--fail --show-error --silent --max-time 30)

if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required for scripts/smoke_chat.sh" >&2
  exit 1
fi

payload="$(jq -n \
  --arg model "${model}" \
  --arg prompt "${prompt}" \
  '{model: $model, messages: [{role: "user", content: $prompt}]}')"

echo "Checking ${gateway_url}/health"
health_response="$(curl "${curl_args[@]}" "${gateway_url}/health")"
echo "${health_response}"

echo
echo "Checking ${gateway_url}/v1/models"
models_response="$(curl "${curl_args[@]}" "${gateway_url}/v1/models")"
echo "${models_response}"

echo
echo "Posting smoke request to ${gateway_url}/v1/chat/completions"
chat_response="$(curl "${curl_args[@]}" -X POST "${gateway_url}/v1/chat/completions" \
  -H 'Content-Type: application/json' \
  --data-raw "${payload}")"
echo "${chat_response}"
