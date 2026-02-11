#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="${ROOT_DIR}/docker-compose.yml"

export BSC_RPC_HTTP="https://bsc-dataseed.bnbchain.org"
export BSC_RPC_WS="wss://bsc-mainnet.nodereal.io/ws/v1/33460afdc9b5404f9c9cfbe2800b9968"
export BSCSCAN_API_KEY="E6C84MGIR6NQ776BTQK6GA9JQII9QASV8D"
export EASYMEME_API_KEY="XRw3V420LP1IpVv2C7p3f"
export EASYMEME_USER_ID="10000"
export EASYMEME_API_HMAC_SECRET="zaSyBfbae36YTd68"
export WALLET_MASTER_KEY="WphGeeMi4safgvbER"
export OPENCLAW_GATEWAY_TOKEN="41a412f04ec0e7d4caa921e79f81a53431ebfa86161a4438"
export EASYMEME_NOTIFY_CHANNEL="telegram"
export EASYMEME_NOTIFY_TO="-5160081771"

cd "$ROOT_DIR"

docker compose -f "$COMPOSE_FILE" up -d --build