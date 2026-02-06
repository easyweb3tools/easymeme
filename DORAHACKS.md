# EasyMeme — AI-Powered Meme Token Hunter for BNB Chain

> **Track: Agent (AI Agent × Onchain Actions)**
> 
> 🔗 **GitHub:** https://github.com/easyweb3tools/easymeme  
> 🌐 **Demo:** (Coming soon)  
> 📹 **Video:** (Coming soon)

---

## 🎯 TL;DR

**EasyMeme** is an AI-powered token discovery and trading tool that helps BNB Chain users catch early meme coins while avoiding rugs and honeypots.

- ⚡ **Real-time scanning** — Detects new PancakeSwap pools in <500ms
- 🛡️ **AI risk analysis** — Honeypot detection, tax analysis, owner permission checks
- 🚀 **One-click trading** — Buy directly through the interface with wallet integration
- 📡 **WebSocket updates** — Live token feed without page refresh

---

## 🤔 Problem

Meme coin trading on BNB Chain is a high-risk, high-reward game. Traders face critical challenges:

| Pain Point | Impact |
|------------|--------|
| **Information asymmetry** | By the time you find a token, insiders already 10x'd |
| **Honeypot scams** | ~30% of new tokens are honeypots that trap your funds |
| **Manual process** | Copy-paste contract addresses, check multiple sites, miss opportunities |
| **No integrated solution** | Separate tools for scanning, analysis, and trading |

**The result?** Retail traders consistently lose to bots and insiders.

---

## 💡 Solution

EasyMeme combines **real-time chain monitoring**, **AI-powered risk analysis**, and **one-click trading** into a single interface.

### Core Features

#### 1. Real-Time Pool Scanner
```
📡 Listening to PancakeSwap Factory...
🆕 New Token: $PEPE2 (0x1234...abcd)
   └─ Initial LP: 5.2 BNB
   └─ Risk Score: 78/100 (Safe)
   └─ [BUY 0.1 BNB] [BUY 0.5 BNB]
```

- Monitors `PairCreated` events via WebSocket
- Filters WBNB pairs automatically
- Pushes new tokens to frontend in real-time

#### 2. AI Risk Engine
Our analysis engine evaluates each token across multiple dimensions:

| Check | Description |
|-------|-------------|
| 🍯 **Honeypot Detection** | Simulates sell transactions to verify tradability |
| 💰 **Tax Analysis** | Detects buy/sell tax rates |
| 🔐 **Permission Risks** | Checks for mint, pause, blacklist capabilities |
| 🔒 **LP Lock Status** | Verifies liquidity lock on PinkLock/Unicrypt |
| 📊 **Holder Distribution** | Flags concentrated token holdings |

**Output:**
```
┌────────────────────────────────────┐
│ Token: $EXAMPLE                    │
│ Risk Score: 72/100 (Medium Risk)   │
├────────────────────────────────────┤
│ ✅ LP locked 180 days (PinkLock)   │
│ ✅ No mint function                │
│ ⚠️ Sell tax 5% (above average)     │
│ ❌ Owner can modify tax (backdoor) │
└────────────────────────────────────┘
```

#### 3. One-Click Trading
- Pre-set BNB amounts: 0.1 / 0.5 / 1 / 5 BNB
- Auto-slippage optimization
- Direct PancakeSwap integration
- Transaction tracking with BSCScan links

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Frontend (Next.js)                   │
│         RainbowKit + wagmi + Real-time WebSocket        │
└─────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────┐
│                   Backend (Go/Gin)                       │
│    REST API │ WebSocket Hub │ Token Scanner │ Analyzer  │
└─────────────────────────────────────────────────────────┘
                           │
            ┌──────────────┼──────────────┐
            ▼              ▼              ▼
      ┌──────────┐   ┌──────────┐   ┌──────────┐
      │PostgreSQL│   │  Redis   │   │ BSC RPC  │
      │ (Storage)│   │ (Cache)  │   │(WebSocket)│
      └──────────┘   └──────────┘   └──────────┘
```

### Tech Stack

| Layer | Technology |
|-------|------------|
| **Frontend** | Next.js 14, TypeScript, Tailwind CSS, RainbowKit |
| **Backend** | Go 1.22, Gin, GORM, go-ethereum |
| **Database** | PostgreSQL 16, Redis 7 |
| **Blockchain** | BSC RPC (HTTP + WebSocket), PancakeSwap V2 |
| **Deployment** | Docker Compose |

---

## 🔗 Onchain Proof

**Contract Interactions:**
- PancakeSwap Factory V2: `0xcA143Ce32Fe78f1f7019d7d551a6402fC5350c73`
- PancakeSwap Router V2: `0x10ED43C718714eb63d5aA57B78B54704E256024E`

**Sample Transactions:** (Will add after live testing)
- Token discovery tx: `0x...`
- Trade execution tx: `0x...`

---

## 🤖 AI Build Log

This project was built with AI assistance using **Claude (Anthropic)** and **Cursor IDE**.

### AI Usage Highlights:
1. **Architecture Design** — Claude helped design the microservice architecture and data models
2. **Go Backend** — AI generated the scanner, analyzer, and WebSocket services
3. **React Components** — AI built the token cards, risk badges, and trading panels
4. **Bug Detection** — AI code review identified risk scoring logic error (scores were inverted!)

📝 **Full AI conversation logs available in:** `docs/dev-logs/`

---

## 🚀 Quick Start

```bash
# Clone the repo
git clone https://github.com/easyweb3tools/easymeme.git
cd easymeme

# Copy config
cp server/config.toml.example server/config.toml
# Edit config.toml with your BSC RPC and BSCScan API key

# Start with Docker
docker compose -f docker-compose.local.yml up -d

# Access
# Frontend: http://localhost:3000
# Backend: http://localhost:8080
# Health: http://localhost:8080/health
```

---

## 📊 Differentiation

| Feature | EasyMeme | GMGN | DEXTools | Maestro Bot |
|---------|----------|------|----------|-------------|
| Real-time scanning | ✅ | ✅ | ❌ | ✅ |
| AI risk analysis | ✅ | ⚠️ | ❌ | ❌ |
| One-click trading | ✅ | ✅ | ❌ | ✅ |
| Web interface | ✅ | ✅ | ✅ | ❌ |
| Non-custodial | ✅ | ⚠️ | N/A | ❌ |
| Open source | ✅ | ❌ | ❌ | ❌ |

**Key differentiator:** EasyMeme is **fully open source** and **non-custodial** — your keys stay in your wallet.

---

## 🗺️ Roadmap

- [x] MVP: Real-time scanner + Risk analysis + One-click buy
- [ ] Telegram Bot integration
- [ ] Wallet tracking / Copy trading
- [ ] Multi-DEX support (Four.meme, BiSwap)
- [ ] opBNB support

---

## 👨‍💻 Team

**easyweb3.tools** — A one-person studio focused on building practical Web3 tools.

- 🐦 Twitter: [@easyweb3tools](https://twitter.com/easyweb3tools)
- 💬 Telegram: [@easyweb3tools](https://t.me/easyweb3tools)

---

## 📜 License

MIT License — Use it, fork it, build on it.

---

## ⚠️ Disclaimer

This tool is for educational and research purposes only. Cryptocurrency trading involves significant risk. Always DYOR (Do Your Own Research) and never invest more than you can afford to lose.
