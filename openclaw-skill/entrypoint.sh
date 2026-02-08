#!/usr/bin/env sh
set -e

STATE_DIR="${OPENCLAW_STATE_DIR:-/tmp/openclaw-state}"
CONFIG_PATH="${OPENCLAW_CONFIG_PATH:-${STATE_DIR}/openclaw.json}"
CRON_DIR="${STATE_DIR}/cron"

mkdir -p "${STATE_DIR}" "${CRON_DIR}"

if [ -f /app/openclaw/openclaw.json ]; then
  cp /app/openclaw/openclaw.json "${CONFIG_PATH}"
fi

if [ -d /app/openclaw/cron ]; then
  cp -R /app/openclaw/cron/* "${CRON_DIR}/"
fi

./node_modules/.bin/openclaw plugins install --link ./
./node_modules/.bin/openclaw plugins enable easymeme-openclaw-skill

./node_modules/.bin/openclaw gateway \
  --port "${OPENCLAW_GATEWAY_PORT:-18789}" \
  --verbose \
  --allow-unconfigured \
  --token "${OPENCLAW_GATEWAY_TOKEN}"
