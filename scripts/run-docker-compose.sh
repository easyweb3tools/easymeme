#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="${ROOT_DIR}/docker-compose.yml"

export BSC_RPC_HTTP="https://bsc-dataseed.bnbchain.org"
export BSC_RPC_WS="wss://bsc-mainnet.nodereal.io/ws/v1/xxx"
export BSCSCAN_API_KEY="xxx"
export EASYMEME_API_KEY="xxx"
export EASYMEME_USER_ID="123"
export EASYMEME_API_HMAC_SECRET="xxx"
export WALLET_MASTER_KEY="xxx"
export OPENCLAW_GATEWAY_TOKEN="xxx"
export EASYMEME_NOTIFY_CHANNEL="telegram"
export EASYMEME_NOTIFY_TO="123456789"

cd "$ROOT_DIR"

docker compose -f "$COMPOSE_FILE" up -d --build