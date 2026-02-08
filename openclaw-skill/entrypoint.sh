#!/usr/bin/env sh
set -e

STATE_DIR="${OPENCLAW_STATE_DIR:-/home/node/.openclaw}"
CONFIG_PATH="${STATE_DIR}/openclaw.json"
CRON_DIR="${STATE_DIR}/cron"

mkdir -p "${STATE_DIR}" "${CRON_DIR}"

if [ ! -f "${CONFIG_PATH}" ]; then
  cat > "${CONFIG_PATH}" <<'JSON'
{
  "models": {
    "mode": "merge",
    "providers": {
      "anyrouter": {
        "baseUrl": "https://anyrouter.top",
        "apiKey": "sk-free",
        "api": "anthropic-messages",
        "models": [
          {
            "id": "claude-opus-4-5-20251101",
            "name": "Claude Opus 4.5",
            "reasoning": true,
            "input": [
              "text",
              "image"
            ],
            "cost": {
              "input": 0,
              "output": 0,
              "cacheRead": 0,
              "cacheWrite": 0
            },
            "contextWindow": 200000,
            "maxTokens": 8192
          }
        ]
      }
    }
  },
  "agents": {
    "defaults": {
      "model": {
        "primary": "anyrouter/claude-opus-4-5-20251101"
      }
    }
  },
  "gateway": {
    "mode": "local"
  },
  "cron": {
    "enabled": true,
    "store": "/home/node/.openclaw/cron/jobs.json",
    "maxConcurrentRuns": 1
  }
}
JSON
fi

if [ ! -f "${CRON_DIR}/jobs.json" ]; then
  cat > "${CRON_DIR}/jobs.json" <<'JSON'
{
  "version": 1,
  "jobs": [
    {
      "jobId": "easymeme-golden-dogs",
      "name": "EasyMeme Golden Dogs",
      "enabled": true,
      "schedule": {
        "kind": "cron",
        "expr": "*/5 * * * *",
        "tz": "UTC"
      },
      "sessionTarget": "isolated",
      "wakeMode": "next-heartbeat",
      "payload": {
        "kind": "agentTurn",
        "message": "获取待分析代币 -> AI 分析 -> 回写结果 -> 如符合条件执行自动交易"
      },
      "delivery": {
        "mode": "none"
      }
    }
  ]
}
JSON
fi

OPENCLAW_CMD="openclaw plugins install --link /app && openclaw plugins enable easymeme-openclaw-skill && openclaw gateway --bind ${OPENCLAW_GATEWAY_BIND:-lan} --port ${OPENCLAW_GATEWAY_PORT:-18789} --verbose --allow-unconfigured --token ${OPENCLAW_GATEWAY_TOKEN}"

if [ "$(id -u)" = "0" ]; then
  chown -R node:node "${STATE_DIR}"
  exec su -s /bin/sh node -c "${OPENCLAW_CMD}"
fi

exec sh -c "${OPENCLAW_CMD}"
