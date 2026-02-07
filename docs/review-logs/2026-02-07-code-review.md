# EasyMeme Code Review - 2026-02-07

## Summary

Full code review against `docs/PRODUCT_SPEC.md`. Overall completion: ~80%.

---

## Component Status

### Server (Go) - 85% Complete

**Implemented:**
- [x] PancakeSwap PairCreated event listener (`internal/service/scanner.go`)
- [x] PostgreSQL token storage (`internal/repository/repository.go`)
- [x] `GET /api/tokens/pending` endpoint
- [x] `GET /api/tokens/:address` endpoint  
- [x] `POST /api/tokens/:address/analysis` endpoint
- [x] `POST /api/trades` endpoint
- [x] WebSocket real-time updates

**Missing:**
- [ ] `GET /api/tokens/analyzed` endpoint - add filter for `analysis_status = 'analyzed'`
- [ ] CreatorAddress extraction in scanner - currently not fetched from chain

### OpenClaw Skill (TypeScript) - 90% Complete

**Implemented:**
- [x] `fetchPendingTokens` tool (`src/tools.ts`)
- [x] `analyzeTokenRisk` tool (`src/tools.ts`)
- [x] `submitAnalysis` tool (`src/tools.ts`)
- [x] SKILL.md workflow definition

**Missing:**
- [ ] Cron configuration (5-minute interval) - requires OpenClaw platform setup
- [ ] Memory persistence implementation - referenced in SKILL.md but not coded

### Web (Next.js) - 70% Complete

**Implemented:**
- [x] Dashboard page with token list
- [x] Wallet connection (RainbowKit + wagmi)
- [x] Buy transaction via PancakeSwap Router
- [x] WebSocket real-time token updates

**Missing:**
- [ ] Token Detail page - single token detailed analysis view
- [ ] History page - user transaction history
- [ ] Stop-loss/take-profit settings in trade panel

---

## Security Issues

### P0 - Critical (Fix Before Demo)

#### 1. Missing Slippage Protection

**File:** `web/components/trade-panel.tsx:47`

**Problem:**
```typescript
args: [0n, [WBNB, token.address], address, deadline],
//     ↑ amountOutMin = 0 allows 100% slippage
```

Users are vulnerable to sandwich attacks with zero slippage protection.

**Fix:**
```typescript
// Add slippage state
const [slippage, setSlippage] = useState(0.5); // 0.5%

// Calculate minimum output
const expectedOutput = await getAmountsOut(amount, [WBNB, token.address]);
const minOutput = expectedOutput * BigInt(Math.floor((100 - slippage) * 100)) / 10000n;

// Use in swap
args: [minOutput, [WBNB, token.address], address, deadline],
```

#### 2. Insecure CORS Configuration

**File:** `server/internal/router/router.go:17-22`

**Problem:**
```go
r.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"*"},
    AllowCredentials: true,
}))
```

`AllowOrigins: *` with `AllowCredentials: true` violates CORS spec.

**Fix:**
```go
r.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"http://localhost:3000", "https://easymeme.example.com"},
    AllowCredentials: true,
}))
```

### P1 - High Priority

#### 3. No API Authentication

**Files:** `server/internal/router/router.go`, `server/internal/handler/token.go`

**Problem:** All API endpoints are publicly accessible. Anyone can submit fake analysis results.

**Fix:** Add API key middleware for OpenClaw endpoints:
```go
func ApiKeyMiddleware(expectedKey string) gin.HandlerFunc {
    return func(c *gin.Context) {
        key := c.GetHeader("X-API-Key")
        if key != expectedKey {
            c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
            return
        }
        c.Next()
    }
}

// Apply to sensitive routes
api.POST("/tokens/:address/analysis", ApiKeyMiddleware(cfg.ApiKey), tokenHandler.PostTokenAnalysis)
```

#### 4. Missing Input Validation

**File:** `server/internal/handler/token.go:103-211`

**Problem:** `riskScore` range (0-100) and `riskLevel` enum values are not validated.

**Fix:**
```go
if payload.RiskScore < 0 || payload.RiskScore > 100 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "riskScore must be 0-100"})
    return
}

validLevels := map[string]bool{"safe": true, "warning": true, "danger": true}
if !validLevels[strings.ToLower(payload.RiskLevel)] {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid riskLevel"})
    return
}
```

### P2 - Medium Priority

#### 5. Error Message Leakage

**Files:** Multiple handlers

**Problem:** Internal errors exposed to clients via `err.Error()`.

**Fix:** Log detailed errors, return generic messages:
```go
log.Printf("Error: %v", err)
c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
```

#### 6. Hardcoded Contract Addresses

**File:** `web/components/trade-panel.tsx:8-9`

**Fix:** Move to environment variables:
```typescript
const PANCAKE_ROUTER = process.env.NEXT_PUBLIC_PANCAKE_ROUTER!;
const WBNB = process.env.NEXT_PUBLIC_WBNB!;
```

---

## Recommended Task Order

1. Add slippage protection to trade panel
2. Fix CORS configuration
3. Add `/api/tokens/analyzed` endpoint
4. Add API key authentication
5. Create Token Detail page
6. Add input validation
7. Create History page

---

## Files Referenced

- `server/internal/router/router.go`
- `server/internal/handler/token.go`
- `server/internal/handler/trade.go`
- `server/internal/service/scanner.go`
- `server/internal/repository/repository.go`
- `server/internal/model/token.go`
- `openclaw-skill/src/tools.ts`
- `openclaw-skill/src/server-api.ts`
- `openclaw-skill/skills/easymeme/SKILL.md`
- `web/app/dashboard/page.tsx`
- `web/components/trade-panel.tsx`
- `web/components/token-list.tsx`
- `web/lib/api.ts`
