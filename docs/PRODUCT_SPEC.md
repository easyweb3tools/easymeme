# EasyMeme — Product Specification for Codex

> **This document is the single source of truth for Codex development.**
> Last updated: 2026-02-10
> Related: [Data Quality Review](review-logs/2026-02-10-data-quality-review.md)

---

## 1. What Is EasyMeme

An autonomous AI agent on BNB Chain that continuously discovers new meme tokens, analyzes their risk using **real on-chain data**, identifies promising tokens ("golden dogs"), and executes trades automatically via a managed wallet.

**Core principles:**
- Golden dogs are time-sensitive — scores decay over time
- OpenClaw is a learning agent — improves via Memory and user feedback
- Designed for personal deployment — each user runs their own instance

---

## 2. Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Server    │────▶│  OpenClaw   │     │    Web      │
│   (Go)      │◀────│   Agent     │     │  (Next.js)  │
└─────────────┘     └─────────────┘     └─────────────┘
      │                   │                   │
 Chain data +          AI analysis         Golden dog list
 GoPlus/DEXScreener    Auto trading        AI trade history
 PostgreSQL            Memory learning     Deploy guide
```

| Component | Responsibility | Stack |
|-----------|---------------|-------|
| `server/` | Chain data ingestion, **GoPlus/DEXScreener enrichment**, database, REST API, managed wallet | Go + PostgreSQL |
| `openclaw-skill/` | AI risk analysis, golden dog scoring, auto trading, user feedback learning | TypeScript + OpenClaw SDK |
| `web/` | Homepage deploy guide, golden dog list, AI trade history | Next.js |

---

## 3. Server (Go) — CRITICAL CHANGES NEEDED

### 3.1 Data Enrichment Pipeline (NEW — P0)

**Problem:** Scanner currently only captures `address`, `name`, `symbol`, `initialLiquidity`. Fields like `creatorAddress`, `buyTax`, `sellTax`, `isHoneypot` are **never populated**. The AI has no real data to analyze.

**Solution:** After a new token is detected via PairCreated, the Server must enrich it with external API data before marking it as ready for analysis.

```
PairCreated event detected
        ↓
scanner.go saves basic token info (existing)
        ↓
NEW: enrichment goroutine kicks in
        ↓
Step 1: Call GoPlus Security API (FREE, no key needed)
  GET https://api.gopluslabs.io/api/v1/token_security/56?contract_addresses={ADDRESS}
  → Write results to Token fields:
    - IsHoneypot    ← response.is_honeypot
    - BuyTax        ← response.buy_tax (string→float64, e.g. "0.05" = 5%)
    - SellTax       ← response.sell_tax
    - CreatorAddress ← response.creator_address
    - RiskDetails   ← full GoPlus JSON response (store as-is)
        ↓
Step 2: Call DEXScreener API (FREE, no key needed)
  GET https://api.dexscreener.com/latest/dex/pairs/bsc/{PAIR_ADDRESS}
  → Write to a new JSON field or separate table:
    - price, priceChange (5m/1h/6h/24h)
    - volume (5m/1h/6h/24h)
    - txns (buys + sells count)
    - liquidity.usd
        ↓
Step 3: Set AnalysisStatus = "enriched" (new status between "pending" and "analyzed")
        ↓
OpenClaw fetches "enriched" tokens instead of "pending" tokens
```

#### GoPlus API Response (key fields to extract)

```go
// File: server/internal/service/goplus.go (NEW FILE)

type GoPlusResponse struct {
    IsHoneypot          string `json:"is_honeypot"`           // "0" or "1"
    BuyTax              string `json:"buy_tax"`               // e.g. "0.05"
    SellTax             string `json:"sell_tax"`              // e.g. "0.10"
    IsMintable          string `json:"is_mintable"`           // "0" or "1"
    CanTakeBackOwnership string `json:"can_take_back_ownership"` // "0" or "1"
    IsProxy             string `json:"is_proxy"`              // "0" or "1"
    IsOpenSource        string `json:"is_open_source"`        // "0" or "1"
    HolderCount         string `json:"holder_count"`          // e.g. "150"
    LpHolderCount       string `json:"lp_holder_count"`       // e.g. "5"
    CreatorAddress      string `json:"creator_address"`
    OwnerAddress        string `json:"owner_address"`
    TotalSupply         string `json:"total_supply"`
    // Store entire response in RiskDetails JSON field
}
```

**API call example:**
```go
func (s *Scanner) enrichWithGoPlus(ctx context.Context, tokenAddress string) (*GoPlusResponse, error) {
    url := fmt.Sprintf("https://api.gopluslabs.io/api/v1/token_security/56?contract_addresses=%s", tokenAddress)
    resp, err := http.Get(url)
    // Parse response.result[tokenAddress]
    // No API key needed, free tier is sufficient
    // Rate limit: ~30 requests/minute, add 2-second delay between calls
}
```

#### DEXScreener API Response (key fields)

```go
// File: server/internal/service/dexscreener.go (NEW FILE)

type DEXScreenerPair struct {
    PriceUsd      string            `json:"priceUsd"`
    PriceChange   map[string]float64 `json:"priceChange"`   // m5, h1, h6, h24
    Volume        map[string]float64 `json:"volume"`         // m5, h1, h6, h24
    Txns          struct {
        M5  TxnCount `json:"m5"`
        H1  TxnCount `json:"h1"`
        H24 TxnCount `json:"h24"`
    } `json:"txns"`
    Liquidity     struct {
        Usd float64 `json:"usd"`
    } `json:"liquidity"`
}

type TxnCount struct {
    Buys  int `json:"buys"`
    Sells int `json:"sells"`
}
```

**API call:**
```go
func (s *Scanner) enrichWithDEXScreener(ctx context.Context, pairAddress string) (*DEXScreenerPair, error) {
    url := fmt.Sprintf("https://api.dexscreener.com/latest/dex/pairs/bsc/%s", pairAddress)
    // No API key needed
    // Rate limit: 300 requests/minute
}
```

### 3.2 Token Model Updates

```go
type Token struct {
    // Existing fields (keep as-is)
    ID               string
    Address          string
    Name             string
    Symbol           string
    Decimals         int
    PairAddress      string
    Dex              string
    InitialLiquidity decimal.Decimal

    // Analysis status: "pending" → "enriched" → "analyzed"
    AnalysisStatus   string

    // GoPlus enrichment (P0 — MUST populate these)
    IsHoneypot       bool            // ← GoPlus is_honeypot
    BuyTax           float64         // ← GoPlus buy_tax
    SellTax          float64         // ← GoPlus sell_tax
    CreatorAddress   string          // ← GoPlus creator_address
    RiskDetails      datatypes.JSON  // ← Full GoPlus JSON response

    // DEXScreener enrichment (P1)
    MarketData       datatypes.JSON  // NEW: DEXScreener price/volume/txns JSON

    // AI analysis result (written by OpenClaw)
    RiskScore        int
    RiskLevel        string
    AnalysisResult   datatypes.JSON
    IsGoldenDog      bool
    GoldenDogScore   int
    AnalyzedAt       *time.Time

    CreatedAt        time.Time
    UpdatedAt        time.Time
}
```

### 3.3 API Endpoints

| Method | Path | Description | Status |
|--------|------|-------------|--------|
| GET | `/api/tokens/pending` | Tokens pending enrichment + analysis | EXISTS |
| GET | `/api/tokens/analyzed` | Analyzed tokens list | EXISTS |
| GET | `/api/tokens/golden-dogs` | Golden dogs with time decay info | EXISTS |
| GET | `/api/tokens/:address` | Single token detail | EXISTS |
| POST | `/api/tokens/:address/analysis` | Write back AI analysis (requires `X-API-Key`) | EXISTS |
| GET | `/api/ai-trades` | AI trade history | EXISTS |
| GET | `/api/ai-trades/stats` | AI trade statistics | EXISTS |
| POST | `/api/wallet/create` | Create managed wallet | EXISTS |
| GET | `/api/wallet/balance` | Query managed wallet balance | EXISTS |
| POST | `/api/wallet/config` | Configure auto-trade parameters | EXISTS |
| POST | `/api/feedback` | User feedback (via OpenClaw/Telegram) | EXISTS |

### 3.4 Updated `/api/tokens/pending` Response

The pending tokens endpoint should return **enriched** data so the OpenClaw agent has real data to analyze:

```json
{
  "data": [
    {
      "address": "0x1234...",
      "name": "MoonDog",
      "symbol": "MDOG",
      "liquidity": 5.2,
      "pairAddress": "0xabcd...",
      "creatorAddress": "0x5678...",
      "createdAt": "2026-02-10T12:00:00Z",

      "goplus": {
        "is_honeypot": false,
        "buy_tax": 0.03,
        "sell_tax": 0.05,
        "is_mintable": false,
        "is_open_source": true,
        "holder_count": 150,
        "lp_holder_count": 3,
        "owner_address": "0x0000..."
      },

      "dexscreener": {
        "priceUsd": "0.00001234",
        "priceChange": { "m5": 12.5, "h1": 45.0, "h6": -5.0 },
        "volume": { "h1": 1500, "h24": 8000 },
        "txns": { "h1": { "buys": 25, "sells": 8 } },
        "liquidity": { "usd": 15000 }
      }
    }
  ]
}
```

### 3.5 Security

- **CORS**: Only configured origins, never `*` with credentials
- **API Key**: POST analysis endpoint requires `X-API-Key` header
- **Input validation**: `riskScore` must be 0-100, `riskLevel` must be SAFE/WARNING/DANGER
- **Managed wallet**: Private key AES-256-GCM encrypted, max balance 5 BNB per wallet

---

## 4. OpenClaw Agent

### 4.1 Role

OpenClaw performs AI analysis, auto trading, and learning. It does NOT do data fetching — that's the Server's job.

### 4.2 Workflow (Updated)

```
1. Cron triggers every 5 minutes
           ↓
2. Fetch ENRICHED tokens from Server API (tokens with GoPlus+DEXScreener data)
           ↓
3. AI analyzes each token using REAL data:
   - GoPlus security data → determine safety
   - DEXScreener market data → assess momentum
   - Combine into golden dog score
           ↓
4. Submit analysis back to Server
           ↓
5. If effectiveScore >= threshold, execute auto-buy
           ↓
6. Update Memory (risk patterns, trade outcomes)
```

### 4.3 Existing Tools (Already Implemented)

- `fetchPendingTokens` — Fetch tokens pending analysis
- `analyzeTokenRisk` — Record AI risk analysis
- `submitAnalysis` — Submit analysis to server
- `estimateGoldenDogScore` — Score using learned weights
- `executeTrade` — Execute trade via managed wallet
- `recordOutcome` — Record trade outcome, update weights
- `getWalletInfo` — Get managed wallet info
- `getPositions` — Get current AI positions
- `upsertWalletConfig` — Update auto-trade config
- `recordUserFeedback` — Record user feedback with reputation

### 4.4 Golden Dog Scoring

The `estimateScore()` function uses learned weights:

```
score = riskScore × baseMultiplier + goldenDogBias - highPenalty × HIGH_count - mediumPenalty × MEDIUM_count
```

With real GoPlus data, the `riskFactors` (honeypotRisk, taxRisk, ownerRisk, concentrationRisk) will be based on facts instead of LLM guesses.

### 4.5 Time Decay (Already Implemented in `token.go`)

| Phase | Time Range | Decay Factor |
|-------|-----------|--------------|
| EARLY | 0-30min | 1.0 |
| PEAK | 30min-2h | 0.8-1.0 |
| DECLINING | 2-6h | 0.5-0.8 |
| EXPIRED | >6h | 0.4 |

### 4.6 Memory & Learning (Already Implemented in `memory.ts`)

- Weights: `baseMultiplier`, `goldenDogBias`, `highPenalty`, `mediumPenalty`
- Outcomes tracking: MOON / RUG / FLAT
- User feedback with reputation-based weighting
- Anti-poisoning: new users weight=0.3, muted users weight=0

### 4.7 Auto-Trade Config (Already Implemented)

Configurable via `upsertWalletConfig` tool and `wallet_configs` table.

---

## 5. Web (Next.js)

### 5.1 Pages

| Page | Function | Priority |
|------|----------|----------|
| **Homepage** | Hero deploy guide + GitHub link + quick start | P0 |
| **Golden Dog List** | AI-identified golden dogs with time phase badges (EARLY/PEAK/DECLINING/EXPIRED) | P0 |
| **AI Trade History** | Agent auto-trade records with P&L stats | P0 |
| **Token Detail** | Enriched token report with human-readable golden dog explanation — see §5.3 | **P1** |

### 5.2 Key Design Decisions

- No trading on the website — link to GMGN/DEXTools for manual trading
- Database stores only AI trades, no human trade tracking
- User interaction happens via OpenClaw Dialog / Telegram, not web UI

### 5.3 Token Detail Page — Enriched View (P1)

> **Goal: Show enriched indicator data AND explain in plain language WHY a token is (or isn't) a golden dog, so regular users can understand the AI's reasoning at a glance.**

#### 5.3.1 API Change Required

The current `GET /api/tokens/:address/detail` response (`TokenDetailResponse`) must be extended to include the enriched JSON fields that Codex already stores in the Token model. Add these fields to the response struct in `server/internal/handler/token.go`:

```go
// Add to TokenDetailResponse struct (handler/token.go)
GoPlus             any `json:"goplus"`
DEXScreener        any `json:"dexscreener"`
HolderDistribution any `json:"holderDistribution"`
CreatorHistory     any `json:"creatorHistory"`
MarketAlerts       any `json:"marketAlerts"`
```

Unmarshal from the Token model fields (`RiskDetails` → goplus normalized, `MarketData` → dexscreener, `HolderData`, `CreatorHistory`, `MarketAlerts`) the same way `GetPendingTokens` already does. Reuse the same unmarshalling pattern.

Update `web/lib/api-types.ts` `TokenDetail` type accordingly:

```typescript
// Add to TokenDetail type (api-types.ts)
goplus?: Record<string, unknown>;
dexscreener?: Record<string, unknown>;
holderDistribution?: Record<string, unknown>;
creatorHistory?: Record<string, unknown>;
marketAlerts?: Array<Record<string, unknown>>;
```

#### 5.3.2 Page Layout — 6 Sections

The token detail page (`web/app/tokens/[address]/page.tsx`) should render these 6 sections from top to bottom. Below is the exact layout, rendering logic, and data source for each.

**Section 1 — Header (exists, keep as-is)**

No change. Keep symbol, name, address, riskLevel badge, CopyButton, lang toggle.

---

**Section 2 — "Why Is This a Golden Dog?" Explanation Card** ⭐ KEY SECTION

This is the most important new section. It must translate the raw data into a human-readable verdict that any user can understand.

```
┌────────────────────────────────────────────────────────────┐
│  🐕 Golden Dog Verdict                  Score: 72 / 100   │
│                                                            │
│  "This token passes all safety checks, has strong buying   │
│   momentum (3x more buys than sells), and reasonable       │
│   holder distribution. Low risk for a small position."     │
│                                                            │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────┐  │
│  │ Safety   │ │ Tax      │ │ Ownership│ │ Momentum     │  │
│  │ ✅ PASS  │ │ ✅ LOW   │ │ ⚠️ MEDIUM│ │ ✅ STRONG    │  │
│  │ No honey │ │ B:3%/S:5%│ │ Mintable │ │ Buys>Sells   │  │
│  └──────────┘ └──────────┘ └──────────┘ └──────────────┘  │
│                                                            │
│  isGoldenDog = true                                        │
└────────────────────────────────────────────────────────────┘
```

Rendering logic (implement as a client component or inline in the page):

```typescript
// Derive from token.analysisResult or token.goplus + token.dexscreener
const riskFactors = token.analysisResult?.riskFactors as {
  honeypotRisk?: string;
  taxRisk?: string;
  ownerRisk?: string;
  concentrationRisk?: string;
} | undefined;

// Build 4 indicator cards:
const indicators = [
  {
    label: "Safety",       // i18n: token_indicator_safety
    level: riskFactors?.honeypotRisk ?? "UNKNOWN",
    detail: riskFactors?.honeypotRisk === "HIGH"
      ? "Honeypot detected" : "No honeypot"
  },
  {
    label: "Tax",           // i18n: token_indicator_tax
    level: riskFactors?.taxRisk ?? "UNKNOWN",
    detail: `Buy ${formatPct(goplus?.buy_tax)}% / Sell ${formatPct(goplus?.sell_tax)}%`
  },
  {
    label: "Ownership",     // i18n: token_indicator_ownership
    level: riskFactors?.ownerRisk ?? "UNKNOWN",
    detail: goplus?.is_mintable ? "Mintable ⚠️" : "Not mintable"
  },
  {
    label: "Momentum",      // i18n: token_indicator_momentum
    level: getMomentumLevel(dexscreener),
    detail: `${buysH1} buys / ${sellsH1} sells (1h)`
  }
];

// Level → color mapping:
// LOW  → green bg, "PASS" or "LOW"
// MEDIUM → yellow bg, "MEDIUM"
// HIGH → red bg, "HIGH" or "FAIL"
// UNKNOWN → gray bg, "N/A"
```

The verdict text (the large quote) should come from `token.analysisResult?.recommendation`. If not available, generate a simple template sentence from the 4 indicators.

---

**Section 3 — Top Scores Row (exists, keep as-is)**

Keep the existing 3-column grid: effectiveScore, goldenDogScore, riskScore with phase and timeDecay.

---

**Section 4 — Contract Safety (GoPlus Data)**

Show a grid of safety check results from `token.goplus`. This replaces the current raw JSON dump.

```
┌─────────────────────────────────────────────────┐
│  🔒 Contract Safety (GoPlus)                    │
│                                                  │
│  ✅ Honeypot: No      ✅ Open Source: Yes        │
│  ✅ Mintable: No      ✅ Proxy: No               │
│  ✅ Owner Renounced   ❌ Can Blacklist: Yes       │
│                                                  │
│  Holders: 150    LP Holders: 3                   │
│  Creator: 0xabc...def                            │
└─────────────────────────────────────────────────┘
```

Data source: `token.goplus` (or `token.riskDetails.normalized`). Fields to display:

| GoPlus field | Display label | Icon logic |
|---|---|---|
| `is_honeypot` | Honeypot | `false` → ✅, `true` → ❌ |
| `is_open_source` | Open Source | `true` → ✅, `false` → ⚠️ |
| `is_mintable` | Mintable | `false` → ✅, `true` → ❌ |
| `is_proxy` | Proxy Contract | `false` → ✅, `true` → ⚠️ |
| `can_take_back_ownership` | Owner Can Reclaim | `false` → ✅, `true` → ❌ |
| `holder_count` | Holders | plain number |
| `lp_holder_count` | LP Holders | plain number |
| `creator_address` | Creator | truncated address + CopyButton |

---

**Section 5 — Market Data (DEXScreener)**

Show real-time market data from `token.dexscreener`.

```
┌─────────────────────────────────────────────────┐
│  📊 Market Data (DEXScreener)                   │
│                                                  │
│  Price: $0.00001234                              │
│                                                  │
│  Price Change:  1h: +45% ↑  6h: -5% ↓  24h: +120% ↑  │
│                                                  │
│  Volume (1h): $1,500      Liquidity: $15,000     │
│                                                  │
│  Transactions (1h):                              │
│  ██████████████████ 25 buys                      │
│  ██████           8 sells                        │
└─────────────────────────────────────────────────┘
```

Fields to display:

| DEXScreener field | Display | Format |
|---|---|---|
| `priceUsd` | Price | `$` + number |
| `priceChange.h1` / `h6` / `h24` | Price Change 1h/6h/24h | `+XX%` green or `-XX%` red |
| `volume.h1` | Volume (1h) | `$X,XXX` |
| `liquidity.usd` | Liquidity | `$X,XXX` |
| `txns.h1.buys` / `txns.h1.sells` | Buy/Sell bar | horizontal stacked bar, green=buys, red=sells |

If `token.dexscreener` is null/empty, show "Market data not yet available" placeholder.

---

**Section 6 — Holder Distribution**

Show from `token.holderDistribution` if available.

```
┌─────────────────────────────────────────────────┐
│  👥 Holder Distribution                         │
│                                                  │
│  Top 10 holders: 35%                             │
│  ████████████░░░░░░░░░░░░░░░░░░░░  35% / 65%   │
│                                                  │
│  Total tracked holders: 50                       │
└─────────────────────────────────────────────────┘
```

| Field | Display |
|---|---|
| `top10Share` | percentage + horizontal bar (green if <60%, yellow if 60-80%, red if >80%) |
| `total` | total holders count |

If `token.holderDistribution` is null/empty, show "Holder data not yet available".

---

**Section 7 — Market Alerts (optional, only show if alerts exist)**

If `token.marketAlerts` array is non-empty, show alert cards:

```
┌─────────────────────────────────────────────────┐
│  ⚠️ Market Alerts                               │
│                                                  │
│  🔴 LIQUIDITY_DROP — Liquidity dropped 45%       │
│     2026-02-10 22:15 UTC                         │
└─────────────────────────────────────────────────┘
```

Each alert has `type`, `severity`, `change`, `timestamp`. Use red bg for HIGH severity, yellow for MEDIUM.

---

**Keep existing sections:** Basic Info + External Tools links + Raw Analysis JSON (move to bottom, collapse by default).

#### 5.3.3 Styling Requirements

- Dark theme consistent with existing pages (dark bg `#0a0a0f`, white/opacity text)
- Glass-morphism cards (`bg-white/5 border-white/10 rounded-2xl`)
- Green (#7cf2a4) for positive / LOW risk, yellow (#ffbf5c) for MEDIUM, red (#f07d7d) for HIGH / negative
- The "Golden Dog Verdict" card should have a subtle green glow border when `isGoldenDog === true`, or a neutral gray border otherwise
- Responsive: 1 column on mobile, 2 columns on md+ for grid sections
- i18n: Add all new display strings to the i18n system (both `zh` and `en`)

#### 5.3.4 New Components to Create

| Component | File | Purpose |
|---|---|---|
| `GoldenDogVerdict` | `web/components/golden-dog-verdict.tsx` | Section 2 — verdict card with 4 indicators |
| `ContractSafety` | `web/components/contract-safety.tsx` | Section 4 — GoPlus safety grid |
| `MarketDataPanel` | `web/components/market-data-panel.tsx` | Section 5 — DEXScreener data |
| `HolderDistribution` | `web/components/holder-distribution.tsx` | Section 6 — holder bar |
| `MarketAlerts` | `web/components/market-alerts.tsx` | Section 7 — alert cards |

All components should accept their respective data as props (not fetch from API themselves). The page component fetches once and passes data down.

---

## 6. Deployment

Docker Compose with: `db` (postgres:16), `server`, `web`, `openclaw-gateway`, `nginx`.

See `docker-compose.yml` and `scripts/run-docker-compose.sh` for configuration.

---

## 7. Development Priorities

### Iteration 1 — Data Foundation (P0) ✅ DONE

> **Goal: Give the AI real data to work with instead of guessing**

**Server changes:**
- [x] Create `server/internal/service/goplus.go` — GoPlus Security API client
- [x] Create `server/internal/service/dexscreener.go` — DEXScreener API client
- [x] Update `scanner.go` — After saving a new token, call GoPlus + DEXScreener to enrich it
- [x] Add `MarketData` JSON field to Token model for DEXScreener data
- [x] Populate `CreatorAddress`, `BuyTax`, `SellTax`, `IsHoneypot` from GoPlus response
- [x] Store full GoPlus response in `RiskDetails` JSON field
- [x] Add new `AnalysisStatus` value: `"enriched"` (between `"pending"` and `"analyzed"`)
- [x] Update `/api/tokens/pending` to return GoPlus + DEXScreener data in response
- [x] Add rate limiting: 2s delay between GoPlus calls, respect DEXScreener limits

**OpenClaw changes:**
- [x] Update SKILL.md prompt to instruct AI to use GoPlus/DEXScreener data from the pending tokens response
- [x] Update `fetchPendingTokens` tool description to document new enriched fields
- [x] AI should map GoPlus `is_honeypot`→`honeypotRisk`, `buy_tax`/`sell_tax`→`taxRisk`, etc.

### Iteration 2 — Market Intelligence (P1) ✅ DONE

- [x] Periodic DEXScreener refresh (every 5 min for tokens < 6h old)
- [x] Track liquidity changes over time (detect rug pulls)
- [x] Add holder distribution data via BSCScan `tokenholderlist` API
- [x] Creator history lookup via BSCScan

### Iteration 3 — Learning Enhancement (P2) ✅ DONE (basic)

- [x] More granular rule performance tracking (`updateFactorPerformanceOnOutcome`)
- [x] Performance windows (7d/30d/all) via `buildPerformanceWindows`
- [ ] Social signal integration (future — fields pre-created)
- [ ] Smart money wallet tracking (future — fields pre-created)

### Iteration 4 — Token Detail Page Enhancement (P1, CURRENT)

> **Goal: Make the frontend token detail page show enriched data AND explain golden dog reasoning in plain language for regular users.**

**Server changes:**
- [ ] Extend `TokenDetailResponse` in `handler/token.go` to include `goplus`, `dexscreener`, `holderDistribution`, `creatorHistory`, `marketAlerts` fields — same unmarshalling as `GetPendingTokens`
- [ ] Add i18n strings for all new labels (both `zh` and `en`)

**Web changes:**
- [ ] Update `TokenDetail` type in `web/lib/api-types.ts` with new fields
- [ ] Create `web/components/golden-dog-verdict.tsx` — Verdict card with 4 risk factor indicators + plain-language explanation
- [ ] Create `web/components/contract-safety.tsx` — GoPlus safety check grid (✅/❌ icons)
- [ ] Create `web/components/market-data-panel.tsx` — DEXScreener price/volume/txns with colored percentages + buy/sell bar
- [ ] Create `web/components/holder-distribution.tsx` — Top-10 share bar + holder count
- [ ] Create `web/components/market-alerts.tsx` — Alert cards for liquidity drops etc.
- [ ] Update `web/app/tokens/[address]/page.tsx` to render all new sections in correct order
- [ ] Collapse raw JSON section by default (add expand/collapse toggle)
- [ ] Responsive layout: 1 column mobile, 2-column grid on md+
- [ ] All new text strings support i18n (zh + en)

---

*End of spec — Codex: work on Iteration 4 items above*
