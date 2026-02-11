# EasyMeme - Your Personal Meme Coin AI Trading Agent

> Automatically discover, analyze, and trade high-potential meme tokens on BNB Chain.

English | [中文文档](./README_CN.md)

**Website**: https://meme.easyweb3.tools/

---

## Core Ideas

- **Golden dogs are time-sensitive**: token opportunities decay over time.
- **OpenClaw is a learning agent**: it updates strategy using memory from wins/losses.
- **Personal deployment first**: each user can run an independent autonomous trading stack.

---

## Demo

[![Demo Video](https://img.youtube.com/vi/pRXXaUhgaRE/hqdefault.jpg)](https://youtube.com/shorts/pRXXaUhgaRE?feature=share)

**Recent live auto-trade screenshots**

<p>
  <img src="./demo/auto-trade-1.png" alt="Auto Trade 1" width="300" />
  <img src="./demo/auto-trade-2.png" alt="Auto Trade 2" width="300" />
  <br />
  <img src="./demo/auto-trade-3.png" alt="Auto Trade 3" width="300" />
  <img src="./demo/auto-trade-4.png" alt="Auto Trade 4" width="300" />
</p>

**On-chain tx hashes (BSCScan)**

- Buy: [0x5b4ea9543d106146d45e0e77e2c940dff36d1872103334ede761899e4c841d8f](https://bscscan.com/tx/0x5b4ea9543d106146d45e0e77e2c940dff36d1872103334ede761899e4c841d8f)
- Sell: [0x4fa43a80799ed20b778e9a2264f9a88eab517f9c92318769dcfb0cdcefdeeb4a](https://bscscan.com/tx/0x4fa43a80799ed20b778e9a2264f9a88eab517f9c92318769dcfb0cdcefdeeb4a)

---

## Why OpenClaw

| Capability | OpenClaw Component | EasyMeme Usage |
|------|------|------|
| Autonomous decisions | Agent | Decide whether a token is a golden dog |
| Historical memory | Memory | Learn risk patterns from outcomes |
| Continuous operation | Cron | Wake up every 5 minutes for scanning/analysis |
| User interaction | Dialog / Telegram | Collect feedback and update behavior |

OpenClaw turns EasyMeme from a static tool into a continuously learning agent.

---

## One-Command Start

```bash
git clone https://github.com/easyweb3tools/easymeme
cd easymeme
cp .env.example .env
# Edit .env only if needed (defaults already work for local run)
./scripts/run-docker-compose.sh
```

Then open `http://localhost`.

---

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Server    │────▶│  OpenClaw   │     │    Web      │
│   (Go)      │◀────│   Agent     │     │  (Next.js)  │
└─────────────┘     └─────────────┘     └─────────────┘
      │                   │                   │
 Chain + enrichment    AI analysis        Golden dog list
 Managed wallet        Auto trading       AI trade history
```

| Component | Responsibility |
|------|------|
| `server/` | BSC data ingestion, GoPlus/DEXScreener/BSCScan enrichment, PostgreSQL APIs, managed wallet |
| `openclaw-skill/` | AI risk analysis, golden dog scoring, auto-trade, feedback learning |
| `web/` | Landing page, golden dog list, AI trade history |

---

## Local Development

### Option A: Script startup (recommended)

Keep config in one file:

```bash
cp .env.example .env
```

Minimal useful edits in `.env`:
- `WALLET_MASTER_KEY` (required only for managed-wallet trading)
- `OPENCLAW_GATEWAY_TOKEN` (recommended if gateway exposed)

Then run:

```bash
./scripts/run-docker-compose.sh
```

OpenClaw config is generated at `/home/node/.openclaw/openclaw.json` (inside Docker volume).
For other providers, configure provider API keys and update OpenClaw config.
Reference: `https://docs.openclaw.ai/concepts/model-providers`

### Option B: Start components manually

1. Start DB:
```bash
docker compose up db -d
```

2. Start Server:
```bash
cd server
cp config.toml.example config.toml
export AUTO_MIGRATE=true
export BSC_RPC_HTTP=https://your-bsc-http
export BSC_RPC_WS=wss://your-bsc-ws
export BSCSCAN_API_KEY=your_bscscan_key
export EASYMEME_API_KEY=your_api_key
export CORS_ALLOWED_ORIGINS=http://localhost:3000
export WALLET_MASTER_KEY=your_wallet_master_key
go run ./cmd/server
```

3. Start Web:
```bash
cd web
npm install
npm run dev
```

4. Start OpenClaw plugin locally:
```bash
cd openclaw-skill
npm install && npm run build
export EASYMEME_SERVER_URL=http://localhost:8080
export EASYMEME_API_KEY=your_api_key
export EASYMEME_USER_ID=default
export EASYMEME_API_HMAC_SECRET=your_hmac_secret
openclaw plugins install --link ./
openclaw plugins enable easymeme-openclaw-skill
openclaw agent --local --session-id easymeme --message "analyze token"
```

### Troubleshooting (OpenClaw fetch failed)

- Check server health: `curl http://localhost:8080/health`
- Check `EASYMEME_SERVER_URL` reachability (especially in Docker networking)
- If `EASYMEME_API_KEY` is set, server must use the same key
- If `EASYMEME_API_HMAC_SECRET` is set, OpenClaw must use the same secret

---

## Memory Learning

OpenClaw memory is used for:
- Deduplicating analyzed tokens
- Storing successful/failed risk patterns
- Dynamically updating scoring weights
- Reputation-aware feedback weighting (anti-poisoning)

---

## Auto-Trading Flow

Two trigger modes:

1. User-triggered (Dialog / Telegram)
- User asks to analyze a token
- OpenClaw generates risk analysis and submits it to EasyMeme
- If strategy/risk rules pass, OpenClaw executes managed-wallet trade

2. Cron-triggered
- Every 5 minutes: fetch enriched pending tokens -> analyze -> submit -> optional auto-trade

Execution outline:
1. Get managed wallet info (address/balance)
2. Create wallet if missing
3. Validate risk and budget constraints
4. Execute on-chain trade and record AI trade
5. Write back results and update memory

---

## Hackathon

**Good Vibes Only: OpenClaw Edition**

This project is submitted to the [BNB Chain Hackathon](https://www.bnbchain.org/en/blog/win-a-share-of-100k-with-good-vibes-only-openclaw-edition) Agent Track.

---

## License

MIT
