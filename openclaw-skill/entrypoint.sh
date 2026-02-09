#!/usr/bin/env sh
set -e

STATE_DIR="${OPENCLAW_STATE_DIR:-/home/node/.openclaw}"
CONFIG_PATH="${STATE_DIR}/openclaw.json"
CRON_DIR="${STATE_DIR}/cron"

mkdir -p "${STATE_DIR}" "${CRON_DIR}"

# ✅ 强制 OpenClaw CLI/Gateway 使用同一个 state/config 目录
export OPENCLAW_STATE_DIR="${STATE_DIR}"
export OPENCLAW_CONFIG_PATH="${CONFIG_PATH}"
export HOME="${HOME:-/home/node}"
export XDG_STATE_HOME="${STATE_DIR}"
export XDG_DATA_HOME="${STATE_DIR}"

if [ ! -f "${CONFIG_PATH}" ]; then
  cat > "${CONFIG_PATH}" <<JSON
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
            "input": ["text","image"],
            "cost": { "input": 0, "output": 0, "cacheRead": 0, "cacheWrite": 0 },
            "contextWindow": 200000,
            "maxTokens": 8192
          }
        ]
      }
    }
  },
  "agents": {
    "defaults": {
      "model": { "primary": "anyrouter/claude-opus-4-5-20251101" }
    }
  },
  "gateway": { "mode": "local" },
  "cron": {
    "enabled": true,
    "store": "${STATE_DIR}/cron/jobs.json",
    "maxConcurrentRuns": 1
  }
}
JSON
fi

if [ ! -f "${CRON_DIR}/jobs.json" ]; then
  NOTIFY_CHANNEL="${EASYMEME_NOTIFY_CHANNEL:-}"
  NOTIFY_TO="${EASYMEME_NOTIFY_TO:-}"
  if [ -n "${NOTIFY_CHANNEL}" ] && [ -n "${NOTIFY_TO}" ]; then
    NOTIFY_NOTE="如发现高质量金狗，请使用 message 工具发送通知：channel=${NOTIFY_CHANNEL}, to=${NOTIFY_TO}。消息需包含代币名称、地址、goldenDogScore、riskScore、建议。"
  else
    NOTIFY_NOTE="若已配置 EASYMEME_NOTIFY_CHANNEL 与 EASYMEME_NOTIFY_TO，请在发现金狗时发送通知。"
  fi

  cat > "${CRON_DIR}/jobs.json" <<JSON
{
  "version": 1,
  "jobs": [
    {
      "jobId": "easymeme-golden-dogs",
      "name": "EasyMeme Golden Dogs",
      "enabled": true,
      "schedule": { "kind": "cron", "expr": "*/5 * * * *", "tz": "UTC" },
      "sessionTarget": "isolated",
      "wakeMode": "next-heartbeat",
      "payload": { "kind": "agentTurn", "message": "获取待分析代币 -> AI 分析 -> 回写结果 -> 如符合条件执行自动交易。${NOTIFY_NOTE}" },
      "delivery": { "mode": "none" }
    }
  ]
}
JSON
fi

OPENCLAW_CMD="openclaw plugins install --link /app \
  && openclaw plugins enable easymeme-openclaw-skill \
  && openclaw gateway run --bind ${OPENCLAW_GATEWAY_BIND:-lan} --port ${OPENCLAW_GATEWAY_PORT:-18789} \
     --verbose --allow-unconfigured --token ${OPENCLAW_GATEWAY_TOKEN}"

if [ "$(id -u)" = "0" ]; then
  chown -R node:node "${STATE_DIR}"
  exec su -s /bin/sh node -c "${OPENCLAW_CMD}"
fi

exec sh -c "${OPENCLAW_CMD}"
