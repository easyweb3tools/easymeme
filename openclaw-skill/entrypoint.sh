#!/usr/bin/env sh
set -e

STATE_DIR="${OPENCLAW_STATE_DIR:-/home/node/.openclaw}"
CONFIG_PATH="${STATE_DIR}/openclaw.json"
CRON_DIR="${STATE_DIR}/cron"

mkdir -p "${STATE_DIR}" "${CRON_DIR}"

if [ -f /app/openclaw/openclaw.json ]; then
  cp /app/openclaw/openclaw.json "${CONFIG_PATH}"
fi

if [ -d /app/openclaw/cron ]; then
  cp -R /app/openclaw/cron/* "${CRON_DIR}/"
fi

OPENCLAW_CMD="openclaw plugins install --link /app && openclaw plugins enable easymeme-openclaw-skill && openclaw gateway --bind ${OPENCLAW_GATEWAY_BIND:-lan} --port ${OPENCLAW_GATEWAY_PORT:-18789} --verbose --allow-unconfigured --token ${OPENCLAW_GATEWAY_TOKEN}"

if [ "$(id -u)" = "0" ]; then
  chown -R node:node "${STATE_DIR}"
  exec su -s /bin/sh node -c "${OPENCLAW_CMD}"
fi

exec sh -c "${OPENCLAW_CMD}"
