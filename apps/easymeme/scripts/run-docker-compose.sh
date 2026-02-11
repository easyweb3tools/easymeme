#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="${ROOT_DIR}/docker-compose.yml"
ENV_FILE="${ROOT_DIR}/.env"
ENV_EXAMPLE_FILE="${ROOT_DIR}/.env.example"

if [[ ! -f "${ENV_FILE}" ]]; then
  if [[ -f "${ENV_EXAMPLE_FILE}" ]]; then
    cp "${ENV_EXAMPLE_FILE}" "${ENV_FILE}"
    echo "Created ${ENV_FILE} from .env.example. Edit it only if you need custom values."
  fi
fi

if [[ -f "${ENV_FILE}" ]]; then
  set -a
  source "${ENV_FILE}"
  set +a
fi

export BSC_RPC_HTTP="${BSC_RPC_HTTP:-https://bsc-dataseed.bnbchain.org}"
export BSC_RPC_WS="${BSC_RPC_WS:-}"
export BSCSCAN_API_KEY="${BSCSCAN_API_KEY:-}"
export EASYMEME_API_KEY="${EASYMEME_API_KEY:-}"
export EASYMEME_USER_ID="${EASYMEME_USER_ID:-default}"
export EASYMEME_API_HMAC_SECRET="${EASYMEME_API_HMAC_SECRET:-}"
export WALLET_MASTER_KEY="${WALLET_MASTER_KEY:-}"
export OPENCLAW_GATEWAY_TOKEN="${OPENCLAW_GATEWAY_TOKEN:-}"
export OPENCLAW_GATEWAY_PORT="${OPENCLAW_GATEWAY_PORT:-18789}"
export OPENCLAW_GATEWAY_BIND="${OPENCLAW_GATEWAY_BIND:-lan}"
export EASYMEME_NOTIFY_CHANNEL="${EASYMEME_NOTIFY_CHANNEL:-}"
export EASYMEME_NOTIFY_TO="${EASYMEME_NOTIFY_TO:-}"
export SERVICE_TOKEN_EASYMEME="${SERVICE_TOKEN_EASYMEME:-dev-token}"
export CORS_ALLOWED_ORIGINS="${CORS_ALLOWED_ORIGINS:-http://localhost:3000}"
export POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-easyweb3}"

cd "$ROOT_DIR"

docker compose -f "$COMPOSE_FILE" up -d --build
