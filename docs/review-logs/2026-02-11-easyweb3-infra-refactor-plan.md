# EasyWeb3 Infrastructure Refactoring Plan — 2026-02-11

> Goal: Extract shared third-party integrations from easymeme into a reusable **base-service**, restructure the repository into a monorepo that supports multiple applications under the easyweb3 umbrella.

---

## Table of Contents

1. [Background & Problem Statement](#1-background--problem-statement)
2. [Current Architecture Analysis](#2-current-architecture-analysis)
3. [Target Architecture](#3-target-architecture)
4. [Monorepo Directory Structure](#4-monorepo-directory-structure)
5. [Phase 1: Repository Restructuring](#5-phase-1-repository-restructuring)
6. [Phase 2: Build base-service](#6-phase-2-build-base-service)
7. [Phase 3: Build Go SDK](#7-phase-3-build-go-sdk)
8. [Phase 4: Refactor easymeme to Use base-service](#8-phase-4-refactor-easymeme-to-use-base-service)
9. [Phase 5: Docker Compose Layered Orchestration](#9-phase-5-docker-compose-layered-orchestration)
10. [Database Boundary Design](#10-database-boundary-design)
11. [API Specification for base-service](#11-api-specification-for-base-service)
12. [Authentication Between Services](#12-authentication-between-services)
13. [Migration Checklist](#13-migration-checklist)

---

## 1. Background & Problem Statement

**EasyWeb3** is a platform that currently has only one application: **EasyMeme** (a meme coin AI trading agent for BNB Chain). The current codebase has these problems:

1. **docker-compose.yml lives inside easymeme** — if we add a second application, infrastructure (PostgreSQL, Nginx, Redis) must be duplicated or awkwardly shared.
2. **Third-party API integrations are hardcoded in easymeme's server** — GoPlus, BSCScan, DEXScreener, Ethereum RPC clients all live in `server/internal/service/`. If a new project needs the same data, it must re-implement these integrations and configure the same API keys.
3. **Wallet management is tightly coupled** — encrypted wallet creation, key management, and on-chain transaction execution are inside easymeme's handler layer.
4. **No caching layer** — each API call to GoPlus/BSCScan/DEXScreener hits the upstream directly. Multiple applications would multiply rate limit pressure.

### What We Want

- **One place to configure API keys** — base-service owns all third-party credentials.
- **One place to cache data** — Redis-backed caching in base-service eliminates redundant upstream calls.
- **Each app has its own database** — application-specific data stays isolated.
- **Shared data via API** — token security, market data, wallet management accessed through base-service REST API.
- **Docker Compose layered orchestration** — infra layer (PostgreSQL, Redis, Nginx, base-service) + app layers (easymeme, future apps) composed together.

---

## 2. Current Architecture Analysis

### 2.1 Files to Migrate (Source → Destination)

| Current File | What It Does | Destination |
|---|---|---|
| `server/internal/service/goplus.go` | GoPlus token security API client | `services/base/internal/service/goplus.go` |
| `server/internal/service/bscscan.go` | BSCScan holder/creator data client | `services/base/internal/service/bscscan.go` |
| `server/internal/service/dexscreener.go` | DEXScreener market data client | `services/base/internal/service/dexscreener.go` |
| `server/pkg/ethereum/client.go` | BSC RPC client (go-ethereum) | `services/base/pkg/ethereum/client.go` |
| `server/internal/handler/wallet.go` (wallet CRUD + encrypt/decrypt) | Wallet management | `services/base/internal/handler/wallet.go` |
| `server/internal/model/managed_wallet.go` | Wallet model | `services/base/internal/model/managed_wallet.go` |

### 2.2 Files to Keep in easymeme (Do NOT Migrate)

| File | Reason |
|---|---|
| `server/internal/service/scanner.go` | Application-specific PairCreated listener + enrichment orchestration |
| `server/internal/handler/token.go` | Application-specific token CRUD & golden dog logic |
| `server/internal/handler/trade.go` | Application-specific trade recording |
| `server/internal/handler/ai_trade.go` | Application-specific AI trade recording |
| `server/internal/model/token.go` | Application-specific token model (golden_dog_score, analysis_result, etc.) |
| `server/internal/model/ai_trade.go` | Application-specific AI trade model |
| `server/internal/model/trade.go` | Application-specific trade model |
| `openclaw-skill/` | Application-specific AI agent |
| `web/` | Application-specific frontend |

### 2.3 Current API Keys & Environment Variables

From `server/internal/config/config.go` and `docker-compose.yml`:

```
BSC_RPC_HTTP          → base-service (RPC proxy)
BSC_RPC_WS            → base-service (RPC proxy)
BSCSCAN_API_KEY       → base-service (chain data)
WALLET_MASTER_KEY     → base-service (wallet management)
EASYMEME_API_KEY      → stays in easymeme (app auth)
EASYMEME_USER_ID      → stays in easymeme (app auth)
EASYMEME_API_HMAC_SECRET → stays in easymeme (app auth)
CORS_ALLOWED_ORIGINS  → both (each service configures its own)
```

### 2.4 Current Third-Party API Details

**GoPlus** (`server/internal/service/goplus.go:15`):
```
Endpoint: https://api.gopluslabs.io/api/v1/token_security/56
Auth: None (public API)
Rate limit: 2-second delay between requests (hardcoded in waitRateLimit())
Chain ID: 56 (BSC hardcoded in URL)
```

**BSCScan** (`server/internal/service/bscscan.go:17`):
```
Endpoint: https://api.bscscan.com/api
Auth: Optional API key (BSCSCAN_API_KEY)
Rate limit: 220ms with API key, 1s without (hardcoded in waitRateLimit())
Methods: tokenholderlist, getcontractcreation, txlist
```

**DEXScreener** (`server/internal/service/dexscreener.go:13`):
```
Endpoint: https://api.dexscreener.com/latest/dex/pairs/bsc/
Auth: None (public API)
Rate limit: 10-second HTTP timeout only
Chain: BSC hardcoded in URL path
```

**Ethereum/BSC RPC** (`server/pkg/ethereum/client.go`):
```
HTTP: BSC_RPC_HTTP (default: https://bsc-dataseed.binance.org)
WS: BSC_RPC_WS (optional, for event subscriptions)
Constants: PancakeFactoryV2, PancakeRouterV2, WBNB addresses
```

---

## 3. Target Architecture

```
                     ┌──────────────────────────────────────────────┐
                     │               Nginx Gateway                  │
                     │  meme.easyweb3.tools  → easymeme-web         │
                     │  meme.easyweb3.tools/api → easymeme-server   │
                     │  api.easyweb3.tools   → base-service         │
                     │  xxx.easyweb3.tools   → future-app           │
                     └───────────┬──────────────────────────────────┘
                                 │
           ┌─────────────────────┼─────────────────────┐
           │                     │                     │
  ┌────────▼────────┐  ┌────────▼────────┐  ┌────────▼────────┐
  │   EasyMeme      │  │   Future App    │  │   Future App    │
  │   (app layer)   │  │   (app layer)   │  │   (app layer)   │
  │                 │  │                 │  │                 │
  │ Golden Dog AI   │  │ Own business    │  │ Own business    │
  │ Trading         │  │ Own database    │  │ Own database    │
  │ OpenClaw Agent  │  │                 │  │                 │
  │ easymeme_db     │  │ app2_db         │  │ app3_db         │
  └────────┬────────┘  └────────┬────────┘  └────────┬────────┘
           │                    │                     │
           └────────────────────┼─────────────────────┘
                                │ (HTTP via go-sdk)
                       ┌────────▼────────┐
                       │  base-service   │
                       │                 │
                       │ Token Security  │ ← GoPlus
                       │ Chain Data      │ ← BSCScan
                       │ Market Data     │ ← DEXScreener
                       │ RPC Proxy       │ ← BSC Node
                       │ Wallet Mgmt     │ ← AES-256
                       │ Notifications   │ ← Telegram
                       │                 │
                       │ base_db + Redis │
                       └─────────────────┘
```

---

## 4. Monorepo Directory Structure

Create this structure from the current `easymeme/` repository. The root should be renamed/reorganized to `easyweb3/`.

```
easyweb3/                                 # Root of the monorepo
├── infra/                                # Shared infrastructure layer
│   ├── docker-compose.yml                # Base infra: PostgreSQL, Redis, base-service, Nginx
│   ├── docker-compose.dev.yml            # Dev overrides (ports, volumes, debug)
│   ├── docker-compose.prod.yml           # Prod overrides (images, TLS, watchtower)
│   ├── nginx/
│   │   └── nginx.conf                    # Unified gateway routing
│   ├── scripts/
│   │   └── init-databases.sh             # Creates multiple databases in single PostgreSQL
│   └── .env.example                      # Template for ALL API keys (single source)
│
├── services/
│   └── base/                             # Base service (Go, Gin framework)
│       ├── cmd/server/main.go            # Entry point
│       ├── internal/
│       │   ├── config/config.go          # Configuration (all API keys)
│       │   ├── handler/
│       │   │   ├── token_security.go     # GoPlus security API endpoints
│       │   │   ├── market_data.go        # DEXScreener market API endpoints
│       │   │   ├── chain_data.go         # BSCScan chain data API endpoints
│       │   │   ├── wallet.go             # Wallet management API endpoints
│       │   │   ├── rpc_proxy.go          # RPC proxy API endpoints
│       │   │   └── notification.go       # Notification API endpoints
│       │   ├── model/
│       │   │   ├── token_security_cache.go
│       │   │   ├── market_data_cache.go
│       │   │   ├── managed_wallet.go     # ← moved from easymeme
│       │   │   ├── notification_log.go
│       │   │   └── service_credential.go
│       │   ├── service/
│       │   │   ├── goplus.go             # ← moved from easymeme
│       │   │   ├── bscscan.go            # ← moved from easymeme
│       │   │   ├── dexscreener.go        # ← moved from easymeme
│       │   │   ├── wallet.go             # Wallet encrypt/decrypt logic
│       │   │   ├── telegram.go           # Telegram bot notifications
│       │   │   └── cache.go              # Redis caching layer
│       │   ├── repository/
│       │   │   └── repository.go
│       │   ├── router/
│       │   │   └── router.go             # Route definitions + auth middleware
│       │   └── middleware/
│       │       └── service_auth.go       # Service-to-service auth (Bearer token)
│       ├── pkg/
│       │   └── ethereum/
│       │       └── client.go             # ← moved from easymeme
│       ├── Dockerfile
│       ├── go.mod                        # module easyweb3/base
│       └── go.sum
│
├── apps/
│   └── easymeme/                         # EasyMeme application (slimmed down)
│       ├── server/
│       │   ├── cmd/server/main.go        # Modified entry point (uses base-client)
│       │   ├── internal/
│       │   │   ├── config/config.go      # App-specific config only
│       │   │   ├── handler/
│       │   │   │   ├── token.go          # Kept as-is
│       │   │   │   ├── trade.go          # Kept as-is
│       │   │   │   ├── ai_trade.go       # Kept as-is
│       │   │   │   ├── wallet.go         # Simplified: proxies to base-service
│       │   │   │   └── websocket.go      # Kept as-is
│       │   │   ├── model/                # App-specific models only
│       │   │   │   ├── token.go          # Kept as-is
│       │   │   │   ├── trade.go          # Kept as-is
│       │   │   │   ├── ai_trade.go       # Kept as-is
│       │   │   │   └── ai_position.go    # Kept as-is
│       │   │   ├── service/
│       │   │   │   └── scanner.go        # Modified: calls base-service instead of direct API
│       │   │   ├── repository/
│       │   │   │   └── repository.go     # Slimmed: no wallet/managed_wallet tables
│       │   │   └── router/
│       │   │       └── router.go         # Kept as-is
│       │   ├── Dockerfile
│       │   ├── go.mod                    # module easyweb3/apps/easymeme
│       │   └── go.sum
│       ├── web/                          # Unchanged
│       ├── openclaw-skill/               # Unchanged
│       └── docker-compose.yml            # App-specific services only
│
├── packages/
│   └── go-sdk/                           # Base Service Go client SDK
│       ├── client.go                     # HTTP client with auth
│       ├── token_security.go             # GetTokenSecurity(chain, address)
│       ├── market_data.go                # GetPairData(chain, pairAddress)
│       ├── chain_data.go                 # GetHolderDistribution, GetCreatorHistory
│       ├── wallet.go                     # CreateWallet, GetBalance, ExecuteTrade
│       ├── rpc.go                        # ProxyRPC, SubscribeEvents
│       ├── notification.go               # SendNotification
│       ├── types.go                      # Shared response types
│       ├── go.mod                        # module easyweb3/go-sdk
│       └── go.sum
│
├── Makefile                              # Top-level build/run commands
├── go.work                               # Go workspace file linking all modules
└── README.md
```

---

## 5. Phase 1: Repository Restructuring

### Step 1.1: Create the monorepo skeleton

Create the directory structure. Do NOT move code yet — just create empty directories.

```bash
mkdir -p infra/nginx infra/scripts
mkdir -p services/base/cmd/server
mkdir -p services/base/internal/{config,handler,model,service,repository,router,middleware}
mkdir -p services/base/pkg/ethereum
mkdir -p apps/easymeme
mkdir -p packages/go-sdk
```

### Step 1.2: Move easymeme into apps/easymeme

Move the entire current easymeme project (server, web, openclaw-skill, nginx) into `apps/easymeme/`. Preserve git history.

```bash
# Move directories
git mv server apps/easymeme/server
git mv web apps/easymeme/web
git mv openclaw-skill apps/easymeme/openclaw-skill
git mv nginx/nginx.conf apps/easymeme/nginx-original.conf  # keep for reference
```

### Step 1.3: Update go.mod in easymeme

Change `apps/easymeme/server/go.mod`:
```go
module easyweb3/apps/easymeme
```

Update all internal imports from `easymeme/internal/...` to `easyweb3/apps/easymeme/internal/...` and `easymeme/pkg/...` to `easyweb3/apps/easymeme/pkg/...`.

Files that need import path updates:
- `apps/easymeme/server/cmd/server/main.go`
- `apps/easymeme/server/internal/router/router.go`
- `apps/easymeme/server/internal/handler/*.go`
- `apps/easymeme/server/internal/service/scanner.go`
- `apps/easymeme/server/internal/repository/repository.go`

### Step 1.4: Create go.work

Create `/go.work` at the repo root:

```go
go 1.23

use (
    ./services/base
    ./apps/easymeme/server
    ./packages/go-sdk
)
```

### Step 1.5: Verify easymeme still builds

```bash
cd apps/easymeme/server && go build ./cmd/server/
```

---

## 6. Phase 2: Build base-service

### Step 2.1: Create go.mod

File: `services/base/go.mod`

```go
module easyweb3/base

go 1.23

require (
    github.com/ethereum/go-ethereum v1.13.14
    github.com/gin-contrib/cors v1.7.6
    github.com/gin-gonic/gin v1.10.1
    github.com/go-redis/redis/v9 v9.4.0
    github.com/shopspring/decimal v1.3.1
    github.com/spf13/viper v1.18.2
    gorm.io/driver/postgres v1.5.6
    gorm.io/gorm v1.30.0
)
```

### Step 2.2: Create config

File: `services/base/internal/config/config.go`

```go
package config

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "github.com/spf13/viper"
)

type Config struct {
    Port               string
    DatabaseURL        string
    RedisURL           string

    // Blockchain RPC
    BscRpcHTTP         string
    BscRpcWS           string

    // Third-party API keys
    BscScanAPIKey      string
    TelegramBotToken   string

    // Wallet encryption
    WalletMasterKey    string

    // Service auth
    ServiceTokens      map[string]string // service_id → token

    // CORS
    CorsAllowedOrigins []string
}

func Load() (*Config, error) {
    v := viper.New()
    v.SetConfigType("toml")

    if path := os.Getenv("CONFIG_PATH"); path != "" {
        v.SetConfigFile(path)
    } else {
        v.SetConfigName("config")
        v.AddConfigPath(".")
        v.AddConfigPath(filepath.Join(".", "config"))
    }

    v.SetDefault("port", "8081")
    v.SetDefault("database_url", "")
    v.SetDefault("redis_url", "redis://localhost:6379")
    v.SetDefault("bsc_rpc_http", "https://bsc-dataseed.binance.org")
    v.SetDefault("bsc_rpc_ws", "")
    v.SetDefault("bscscan_api_key", "")
    v.SetDefault("telegram_bot_token", "")
    v.SetDefault("wallet_master_key", "")
    v.SetDefault("cors_allowed_origins", "")

    v.AutomaticEnv()
    _ = v.BindEnv("port", "PORT")
    _ = v.BindEnv("database_url", "DATABASE_URL")
    _ = v.BindEnv("redis_url", "REDIS_URL")
    _ = v.BindEnv("bsc_rpc_http", "BSC_RPC_HTTP")
    _ = v.BindEnv("bsc_rpc_ws", "BSC_RPC_WS")
    _ = v.BindEnv("bscscan_api_key", "BSCSCAN_API_KEY")
    _ = v.BindEnv("telegram_bot_token", "TELEGRAM_BOT_TOKEN")
    _ = v.BindEnv("wallet_master_key", "WALLET_MASTER_KEY")
    _ = v.BindEnv("cors_allowed_origins", "CORS_ALLOWED_ORIGINS")

    if err := v.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            return nil, fmt.Errorf("load config: %w", err)
        }
    }

    // Parse service tokens from env: SERVICE_TOKEN_EASYMEME=xxx, SERVICE_TOKEN_APP2=yyy
    serviceTokens := make(map[string]string)
    for _, env := range os.Environ() {
        if strings.HasPrefix(env, "SERVICE_TOKEN_") {
            parts := strings.SplitN(env, "=", 2)
            if len(parts) == 2 {
                serviceID := strings.ToLower(strings.TrimPrefix(parts[0], "SERVICE_TOKEN_"))
                serviceTokens[serviceID] = parts[1]
            }
        }
    }

    return &Config{
        Port:               v.GetString("port"),
        DatabaseURL:        v.GetString("database_url"),
        RedisURL:           v.GetString("redis_url"),
        BscRpcHTTP:         v.GetString("bsc_rpc_http"),
        BscRpcWS:           v.GetString("bsc_rpc_ws"),
        BscScanAPIKey:      v.GetString("bscscan_api_key"),
        TelegramBotToken:   v.GetString("telegram_bot_token"),
        WalletMasterKey:    v.GetString("wallet_master_key"),
        ServiceTokens:      serviceTokens,
        CorsAllowedOrigins: splitOrigins(v.GetString("cors_allowed_origins")),
    }, nil
}

func splitOrigins(raw string) []string {
    parts := strings.Split(raw, ",")
    origins := make([]string, 0, len(parts))
    for _, part := range parts {
        trimmed := strings.TrimSpace(part)
        if trimmed != "" {
            origins = append(origins, trimmed)
        }
    }
    return origins
}
```

### Step 2.3: Copy and adapt service clients

Copy from easymeme and make these changes:

**`services/base/internal/service/goplus.go`**
- Copy from `apps/easymeme/server/internal/service/goplus.go` (lines 1-127 as-is)
- Change: make chain ID configurable instead of hardcoded `56`
- Add: `GetTokenSecurityMultiChain(ctx, chain, tokenAddress)` that routes to the correct endpoint

```go
// Key change: make chain configurable
var goPlusEndpoints = map[string]string{
    "bsc":      "https://api.gopluslabs.io/api/v1/token_security/56",
    "ethereum": "https://api.gopluslabs.io/api/v1/token_security/1",
    "arbitrum": "https://api.gopluslabs.io/api/v1/token_security/42161",
}

func (c *GoPlusClient) GetTokenSecurity(ctx context.Context, chain, tokenAddress string) (*GoPlusSecurityData, error) {
    endpoint, ok := goPlusEndpoints[chain]
    if !ok {
        return nil, fmt.Errorf("unsupported chain: %s", chain)
    }
    // ... rest is same as current goplus.go but use configurable endpoint
}
```

**`services/base/internal/service/bscscan.go`**
- Copy from `apps/easymeme/server/internal/service/bscscan.go` (lines 1-291 as-is)
- Change: make base URL configurable for multi-chain support

```go
var scanEndpoints = map[string]string{
    "bsc":      "https://api.bscscan.com/api",
    "ethereum": "https://api.etherscan.io/api",
    "arbitrum": "https://api.arbiscan.io/api",
}
```

**`services/base/internal/service/dexscreener.go`**
- Copy from `apps/easymeme/server/internal/service/dexscreener.go` (lines 1-58 as-is)
- Change: make chain configurable in URL path

```go
func (c *DEXScreenerClient) GetPairData(ctx context.Context, chain, pairAddress string) (map[string]interface{}, error) {
    endpoint := fmt.Sprintf("https://api.dexscreener.com/latest/dex/pairs/%s/%s", chain, strings.ToLower(pairAddress))
    // ... rest is same
}
```

**`services/base/pkg/ethereum/client.go`**
- Copy from `apps/easymeme/server/pkg/ethereum/client.go` (lines 1-321 as-is)
- Keep all methods unchanged — base-service uses this for wallet operations and RPC proxy

### Step 2.4: Create Redis caching service

File: `services/base/internal/service/cache.go`

```go
package service

import (
    "context"
    "encoding/json"
    "time"

    "github.com/redis/go-redis/v9"
)

type CacheService struct {
    client *redis.Client
}

func NewCacheService(redisURL string) (*CacheService, error) {
    opts, err := redis.ParseURL(redisURL)
    if err != nil {
        return nil, err
    }
    client := redis.NewClient(opts)
    if err := client.Ping(context.Background()).Err(); err != nil {
        return nil, err
    }
    return &CacheService{client: client}, nil
}

func (c *CacheService) Get(ctx context.Context, key string, dest interface{}) error {
    val, err := c.client.Get(ctx, key).Result()
    if err != nil {
        return err
    }
    return json.Unmarshal([]byte(val), dest)
}

func (c *CacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    return c.client.Set(ctx, key, data, ttl).Err()
}

func (c *CacheService) Delete(ctx context.Context, key string) error {
    return c.client.Del(ctx, key).Err()
}
```

### Step 2.5: Create handlers that wrap services with caching

File: `services/base/internal/handler/token_security.go`

```go
package handler

import (
    "fmt"
    "net/http"
    "time"

    "easyweb3/base/internal/service"

    "github.com/gin-gonic/gin"
)

type TokenSecurityHandler struct {
    goPlus *service.GoPlusClient
    cache  *service.CacheService
}

func NewTokenSecurityHandler(goPlus *service.GoPlusClient, cache *service.CacheService) *TokenSecurityHandler {
    return &TokenSecurityHandler{goPlus: goPlus, cache: cache}
}

// GetTokenSecurity godoc
// @Summary Get token security analysis
// @Description Returns GoPlus security data for a token, cached for 1 hour
// @Tags token-security
// @Param chain path string true "Chain name (bsc, ethereum, arbitrum)"
// @Param address path string true "Token contract address"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/tokens/{chain}/{address}/security [get]
func (h *TokenSecurityHandler) GetTokenSecurity(c *gin.Context) {
    chain := c.Param("chain")
    address := c.Param("address")

    cacheKey := fmt.Sprintf("token_security:%s:%s", chain, address)

    // Try cache first
    var cached service.GoPlusSecurityData
    if err := h.cache.Get(c.Request.Context(), cacheKey, &cached); err == nil {
        c.JSON(http.StatusOK, gin.H{"data": cached, "cached": true})
        return
    }

    // Cache miss — call GoPlus
    data, err := h.goPlus.GetTokenSecurity(c.Request.Context(), chain, address)
    if err != nil {
        c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("goplus: %s", err.Error())})
        return
    }

    // Cache for 1 hour
    _ = h.cache.Set(c.Request.Context(), cacheKey, data, 1*time.Hour)

    c.JSON(http.StatusOK, gin.H{"data": data, "cached": false})
}
```

File: `services/base/internal/handler/market_data.go`

```go
package handler

import (
    "fmt"
    "net/http"
    "time"

    "easyweb3/base/internal/service"

    "github.com/gin-gonic/gin"
)

type MarketDataHandler struct {
    dexScreener *service.DEXScreenerClient
    cache       *service.CacheService
}

func NewMarketDataHandler(dex *service.DEXScreenerClient, cache *service.CacheService) *MarketDataHandler {
    return &MarketDataHandler{dexScreener: dex, cache: cache}
}

// GetPairData godoc
// @Summary Get DEX pair market data
// @Description Returns DEXScreener pair data, cached for 1 minute
// @Tags market-data
// @Param chain path string true "Chain name (bsc, ethereum, arbitrum)"
// @Param pairAddress path string true "DEX pair address"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/market/{chain}/pairs/{pairAddress} [get]
func (h *MarketDataHandler) GetPairData(c *gin.Context) {
    chain := c.Param("chain")
    pairAddress := c.Param("pairAddress")

    cacheKey := fmt.Sprintf("pair_data:%s:%s", chain, pairAddress)

    var cached map[string]interface{}
    if err := h.cache.Get(c.Request.Context(), cacheKey, &cached); err == nil {
        c.JSON(http.StatusOK, gin.H{"data": cached, "cached": true})
        return
    }

    data, err := h.dexScreener.GetPairData(c.Request.Context(), chain, pairAddress)
    if err != nil {
        c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("dexscreener: %s", err.Error())})
        return
    }

    _ = h.cache.Set(c.Request.Context(), cacheKey, data, 1*time.Minute)

    c.JSON(http.StatusOK, gin.H{"data": data, "cached": false})
}
```

File: `services/base/internal/handler/chain_data.go`

```go
package handler

import (
    "fmt"
    "net/http"
    "time"

    "easyweb3/base/internal/service"

    "github.com/gin-gonic/gin"
)

type ChainDataHandler struct {
    bscScan *service.BscScanClient
    cache   *service.CacheService
}

func NewChainDataHandler(bsc *service.BscScanClient, cache *service.CacheService) *ChainDataHandler {
    return &ChainDataHandler{bscScan: bsc, cache: cache}
}

// GetHolderDistribution godoc
// @Summary Get token holder distribution
// @Description Returns top holders, cached for 30 minutes
// @Tags chain-data
// @Param chain path string true "Chain name"
// @Param address path string true "Token contract address"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/tokens/{chain}/{address}/holders [get]
func (h *ChainDataHandler) GetHolderDistribution(c *gin.Context) {
    chain := c.Param("chain")
    address := c.Param("address")

    cacheKey := fmt.Sprintf("holders:%s:%s", chain, address)

    var cached service.HolderDistribution
    if err := h.cache.Get(c.Request.Context(), cacheKey, &cached); err == nil {
        c.JSON(http.StatusOK, gin.H{"data": cached, "cached": true})
        return
    }

    data, err := h.bscScan.FetchHolderDistribution(c.Request.Context(), address)
    if err != nil {
        c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("bscscan: %s", err.Error())})
        return
    }

    _ = h.cache.Set(c.Request.Context(), cacheKey, data, 30*time.Minute)

    c.JSON(http.StatusOK, gin.H{"data": data, "cached": false})
}

// GetCreatorHistory godoc
// @Summary Get token creator history
// @Description Returns creator address and tx history, cached for 1 hour
// @Tags chain-data
// @Param chain path string true "Chain name"
// @Param address path string true "Token contract address"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/tokens/{chain}/{address}/creator [get]
func (h *ChainDataHandler) GetCreatorHistory(c *gin.Context) {
    chain := c.Param("chain")
    address := c.Param("address")

    cacheKey := fmt.Sprintf("creator:%s:%s", chain, address)

    var cached service.CreatorHistory
    if err := h.cache.Get(c.Request.Context(), cacheKey, &cached); err == nil {
        c.JSON(http.StatusOK, gin.H{"data": cached, "cached": true})
        return
    }

    data, err := h.bscScan.FetchCreatorHistory(c.Request.Context(), address)
    if err != nil {
        c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("bscscan: %s", err.Error())})
        return
    }

    _ = h.cache.Set(c.Request.Context(), cacheKey, data, 1*time.Hour)

    c.JSON(http.StatusOK, gin.H{"data": data, "cached": false})
}
```

File: `services/base/internal/handler/notification.go`

```go
package handler

import (
    "fmt"
    "net/http"

    "easyweb3/base/internal/service"

    "github.com/gin-gonic/gin"
)

type NotificationHandler struct {
    telegram *service.TelegramClient
}

func NewNotificationHandler(telegram *service.TelegramClient) *NotificationHandler {
    return &NotificationHandler{telegram: telegram}
}

type SendNotificationRequest struct {
    Channel   string `json:"channel" binding:"required"`   // "telegram"
    To        string `json:"to" binding:"required"`         // chat_id
    Message   string `json:"message" binding:"required"`
    ParseMode string `json:"parse_mode"`                    // "HTML" or "Markdown"
}

// SendNotification godoc
// @Summary Send notification
// @Description Send a notification via the specified channel
// @Tags notifications
// @Param payload body SendNotificationRequest true "Notification payload"
// @Success 200 {object} map[string]string
// @Router /api/v1/notifications/send [post]
func (h *NotificationHandler) SendNotification(c *gin.Context) {
    var req SendNotificationRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    switch req.Channel {
    case "telegram":
        if err := h.telegram.SendMessage(c.Request.Context(), req.To, req.Message, req.ParseMode); err != nil {
            c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("telegram: %s", err.Error())})
            return
        }
    default:
        c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unsupported channel: %s", req.Channel)})
        return
    }

    c.JSON(http.StatusOK, gin.H{"status": "sent"})
}
```

### Step 2.6: Create Telegram service

File: `services/base/internal/service/telegram.go`

```go
package service

import (
    "context"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "time"
)

type TelegramClient struct {
    botToken   string
    httpClient *http.Client
}

func NewTelegramClient(botToken string) *TelegramClient {
    return &TelegramClient{
        botToken:   botToken,
        httpClient: &http.Client{Timeout: 10 * time.Second},
    }
}

func (c *TelegramClient) SendMessage(ctx context.Context, chatID, text, parseMode string) error {
    if c.botToken == "" {
        return fmt.Errorf("telegram bot token not configured")
    }
    if parseMode == "" {
        parseMode = "HTML"
    }

    endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", c.botToken)
    params := url.Values{}
    params.Set("chat_id", chatID)
    params.Set("text", text)
    params.Set("parse_mode", parseMode)

    req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"?"+params.Encode(), nil)
    if err != nil {
        return err
    }

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
        return fmt.Errorf("telegram status %d: %s", resp.StatusCode, string(body))
    }

    return nil
}
```

### Step 2.7: Create service auth middleware

File: `services/base/internal/middleware/service_auth.go`

```go
package middleware

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
)

// ServiceAuthMiddleware validates Bearer tokens from app services.
// serviceTokens is a map of service_id → expected_token.
// The token format is: "Bearer <service_id>:<token>"
func ServiceAuthMiddleware(serviceTokens map[string]string) gin.HandlerFunc {
    if len(serviceTokens) == 0 {
        // No tokens configured — allow all (dev mode)
        return func(c *gin.Context) {
            c.Set("service_id", "anonymous")
            c.Next()
        }
    }
    return func(c *gin.Context) {
        auth := c.GetHeader("Authorization")
        if auth == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
            return
        }

        token := strings.TrimPrefix(auth, "Bearer ")
        if token == auth {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
            return
        }

        // Find matching service
        for serviceID, expectedToken := range serviceTokens {
            if token == expectedToken {
                c.Set("service_id", serviceID)
                c.Next()
                return
            }
        }

        c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid service token"})
    }
}
```

### Step 2.8: Create router

File: `services/base/internal/router/router.go`

```go
package router

import (
    "easyweb3/base/internal/config"
    "easyweb3/base/internal/handler"
    "easyweb3/base/internal/middleware"

    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
)

func Setup(
    cfg *config.Config,
    tokenSecurityHandler *handler.TokenSecurityHandler,
    marketDataHandler *handler.MarketDataHandler,
    chainDataHandler *handler.ChainDataHandler,
    walletHandler *handler.WalletHandler,
    notificationHandler *handler.NotificationHandler,
) *gin.Engine {
    r := gin.Default()

    r.Use(cors.New(cors.Config{
        AllowOrigins:     cfg.CorsAllowedOrigins,
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
        AllowCredentials: true,
    }))

    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok", "service": "base"})
    })

    v1 := r.Group("/api/v1")
    v1.Use(middleware.ServiceAuthMiddleware(cfg.ServiceTokens))
    {
        // Token security (GoPlus)
        v1.GET("/tokens/:chain/:address/security", tokenSecurityHandler.GetTokenSecurity)

        // Chain data (BSCScan)
        v1.GET("/tokens/:chain/:address/holders", chainDataHandler.GetHolderDistribution)
        v1.GET("/tokens/:chain/:address/creator", chainDataHandler.GetCreatorHistory)

        // Market data (DEXScreener)
        v1.GET("/market/:chain/pairs/:pairAddress", marketDataHandler.GetPairData)

        // Wallet management
        v1.POST("/wallets", walletHandler.CreateWallet)
        v1.GET("/wallets/:walletId/balance", walletHandler.GetWalletBalance)
        v1.POST("/wallets/:walletId/execute", walletHandler.ExecuteTrade)

        // Notifications
        v1.POST("/notifications/send", notificationHandler.SendNotification)
    }

    return r
}
```

### Step 2.9: Create entry point

File: `services/base/cmd/server/main.go`

```go
package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"

    "easyweb3/base/internal/config"
    "easyweb3/base/internal/handler"
    "easyweb3/base/internal/repository"
    "easyweb3/base/internal/router"
    "easyweb3/base/internal/service"
    "easyweb3/base/pkg/ethereum"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    repo, err := repository.New(cfg.DatabaseURL)
    if err != nil {
        log.Fatalf("Failed to connect database: %v", err)
    }
    log.Println("Database connected")

    cache, err := service.NewCacheService(cfg.RedisURL)
    if err != nil {
        log.Fatalf("Failed to connect Redis: %v", err)
    }
    log.Println("Redis connected")

    ethClient, err := ethereum.NewClient(cfg.BscRpcHTTP, cfg.BscRpcWS)
    if err != nil {
        log.Fatalf("Failed to connect BSC: %v", err)
    }
    defer ethClient.Close()
    log.Println("BSC RPC connected")

    // Service clients
    goplusClient := service.NewGoPlusClient()
    dexClient := service.NewDEXScreenerClient()
    bscScanClient := service.NewBscScanClient(cfg.BscScanAPIKey)
    telegramClient := service.NewTelegramClient(cfg.TelegramBotToken)

    // Handlers
    tokenSecurityHandler := handler.NewTokenSecurityHandler(goplusClient, cache)
    marketDataHandler := handler.NewMarketDataHandler(dexClient, cache)
    chainDataHandler := handler.NewChainDataHandler(bscScanClient, cache)
    walletHandler := handler.NewWalletHandler(repo, ethClient, cfg.WalletMasterKey)
    notificationHandler := handler.NewNotificationHandler(telegramClient)

    r := router.Setup(cfg, tokenSecurityHandler, marketDataHandler, chainDataHandler, walletHandler, notificationHandler)

    go func() {
        log.Printf("Base service starting on port %s", cfg.Port)
        if err := r.Run(":" + cfg.Port); err != nil {
            log.Fatalf("Server failed: %v", err)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    log.Println("Shutting down...")
}
```

### Step 2.10: Create Dockerfile

File: `services/base/Dockerfile`

```dockerfile
FROM golang:1.23-alpine AS builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/base-server ./cmd/server/

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /bin/base-server /bin/base-server
EXPOSE 8081
CMD ["/bin/base-server"]
```

---

## 7. Phase 3: Build Go SDK

The Go SDK is a lightweight HTTP client that application services use to call base-service.

### Step 3.1: Create SDK client

File: `packages/go-sdk/client.go`

```go
package basesdk

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

type Client struct {
    baseURL    string
    token      string
    httpClient *http.Client
}

func NewClient(baseURL, serviceToken string) *Client {
    return &Client{
        baseURL: baseURL,
        token:   serviceToken,
        httpClient: &http.Client{Timeout: 30 * time.Second},
    }
}

func (c *Client) get(ctx context.Context, path string, result interface{}) error {
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
    if err != nil {
        return err
    }
    req.Header.Set("Authorization", "Bearer "+c.token)
    return c.doRequest(req, result)
}

func (c *Client) post(ctx context.Context, path string, body interface{}, result interface{}) error {
    data, err := json.Marshal(body)
    if err != nil {
        return err
    }
    req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(data))
    if err != nil {
        return err
    }
    req.Header.Set("Authorization", "Bearer "+c.token)
    req.Header.Set("Content-Type", "application/json")
    return c.doRequest(req, result)
}

func (c *Client) doRequest(req *http.Request, result interface{}) error {
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB limit
    if err != nil {
        return err
    }

    if resp.StatusCode >= 400 {
        return fmt.Errorf("base-service %s %d: %s", req.URL.Path, resp.StatusCode, string(body))
    }

    if result != nil {
        var envelope struct {
            Data json.RawMessage `json:"data"`
        }
        if err := json.Unmarshal(body, &envelope); err != nil {
            return json.Unmarshal(body, result)
        }
        return json.Unmarshal(envelope.Data, result)
    }

    return nil
}
```

File: `packages/go-sdk/token_security.go`

```go
package basesdk

import (
    "context"
    "fmt"
)

type TokenSecurityData struct {
    IsHoneypot           string `json:"is_honeypot"`
    BuyTax               string `json:"buy_tax"`
    SellTax              string `json:"sell_tax"`
    IsMintable           string `json:"is_mintable"`
    CanTakeBackOwnership string `json:"can_take_back_ownership"`
    IsProxy              string `json:"is_proxy"`
    IsOpenSource         string `json:"is_open_source"`
    HolderCount          string `json:"holder_count"`
    LpHolderCount        string `json:"lp_holder_count"`
    CreatorAddress       string `json:"creator_address"`
    OwnerAddress         string `json:"owner_address"`
    TotalSupply          string `json:"total_supply"`
    Raw                  map[string]interface{} `json:"raw,omitempty"`
}

func (c *Client) GetTokenSecurity(ctx context.Context, chain, address string) (*TokenSecurityData, error) {
    var result TokenSecurityData
    path := fmt.Sprintf("/api/v1/tokens/%s/%s/security", chain, address)
    if err := c.get(ctx, path, &result); err != nil {
        return nil, err
    }
    return &result, nil
}
```

File: `packages/go-sdk/market_data.go`

```go
package basesdk

import (
    "context"
    "fmt"
)

func (c *Client) GetPairData(ctx context.Context, chain, pairAddress string) (map[string]interface{}, error) {
    var result map[string]interface{}
    path := fmt.Sprintf("/api/v1/market/%s/pairs/%s", chain, pairAddress)
    if err := c.get(ctx, path, &result); err != nil {
        return nil, err
    }
    return result, nil
}
```

File: `packages/go-sdk/chain_data.go`

```go
package basesdk

import (
    "context"
    "fmt"
)

type HolderDistribution struct {
    TopHolders []map[string]interface{} `json:"topHolders"`
    Top10Share float64                  `json:"top10Share"`
    Total      int                      `json:"total"`
    Source     string                   `json:"source"`
}

type CreatorHistory struct {
    CreatorAddress   string                   `json:"creatorAddress"`
    ContractAddress  string                   `json:"contractAddress"`
    CreationTxHash   string                   `json:"creationTxHash"`
    CreatedContracts []string                 `json:"createdContracts"`
    RecentTxs        []map[string]interface{} `json:"recentTxs"`
    Source           string                   `json:"source"`
}

func (c *Client) GetHolderDistribution(ctx context.Context, chain, address string) (*HolderDistribution, error) {
    var result HolderDistribution
    path := fmt.Sprintf("/api/v1/tokens/%s/%s/holders", chain, address)
    if err := c.get(ctx, path, &result); err != nil {
        return nil, err
    }
    return &result, nil
}

func (c *Client) GetCreatorHistory(ctx context.Context, chain, address string) (*CreatorHistory, error) {
    var result CreatorHistory
    path := fmt.Sprintf("/api/v1/tokens/%s/%s/creator", chain, address)
    if err := c.get(ctx, path, &result); err != nil {
        return nil, err
    }
    return &result, nil
}
```

File: `packages/go-sdk/notification.go`

```go
package basesdk

import "context"

type SendNotificationRequest struct {
    Channel   string `json:"channel"`
    To        string `json:"to"`
    Message   string `json:"message"`
    ParseMode string `json:"parse_mode,omitempty"`
}

func (c *Client) SendNotification(ctx context.Context, channel, to, message, parseMode string) error {
    req := SendNotificationRequest{
        Channel:   channel,
        To:        to,
        Message:   message,
        ParseMode: parseMode,
    }
    return c.post(ctx, "/api/v1/notifications/send", req, nil)
}
```

File: `packages/go-sdk/go.mod`

```go
module easyweb3/go-sdk

go 1.23
```

---

## 8. Phase 4: Refactor easymeme to Use base-service

### Step 4.1: Add go-sdk dependency

In `apps/easymeme/server/go.mod`, add:
```go
require easyweb3/go-sdk v0.0.0
```

The `go.work` file handles local resolution.

### Step 4.2: Add base-service config to easymeme

Modify `apps/easymeme/server/internal/config/config.go` — add two fields:

```go
type Config struct {
    // ... existing fields ...
    BaseServiceURL   string  // URL of base-service (e.g., http://base-service:8081)
    BaseServiceToken string  // Service auth token for base-service
}
```

Add corresponding env bindings:
```go
_ = v.BindEnv("base_service_url", "BASE_SERVICE_URL")
_ = v.BindEnv("base_service_token", "BASE_SERVICE_TOKEN")
```

### Step 4.3: Modify scanner.go to use go-sdk

This is the key refactoring step. Currently `scanner.go` directly uses `GoPlusClient`, `DEXScreenerClient`, and `BscScanClient`. After refactoring, it uses the go-sdk `basesdk.Client` instead.

**Current** `scanner.go` struct (lines 24-32):
```go
type Scanner struct {
    client      *ethereum.Client         // KEEP for PairCreated subscription
    repo        *repository.Repository
    hub         Broadcaster
    goPlus      *GoPlusClient            // REMOVE
    dexScreener *DEXScreenerClient       // REMOVE
    bscScan     *BscScanClient           // REMOVE
    stats       *enrichmentStats
}
```

**After** refactoring:
```go
type Scanner struct {
    client      *ethereum.Client         // KEEP — still needed for PairCreated event subscription
    repo        *repository.Repository
    hub         Broadcaster
    base        *basesdk.Client          // NEW — replaces goPlus, dexScreener, bscScan
    stats       *enrichmentStats
}
```

**Current** `NewScanner()` (line 126-136):
```go
func NewScanner(client *ethereum.Client, repo *repository.Repository, hub Broadcaster, bscScanAPIKey string) *Scanner {
    return &Scanner{
        client:      client,
        repo:        repo,
        hub:         hub,
        goPlus:      NewGoPlusClient(),
        dexScreener: NewDEXScreenerClient(),
        bscScan:     NewBscScanClient(bscScanAPIKey),
        stats:       newEnrichmentStats(),
    }
}
```

**After** refactoring:
```go
func NewScanner(client *ethereum.Client, repo *repository.Repository, hub Broadcaster, baseClient *basesdk.Client) *Scanner {
    return &Scanner{
        client: client,
        repo:   repo,
        hub:    hub,
        base:   baseClient,
        stats:  newEnrichmentStats(),
    }
}
```

**Current** `enrichTokenOnce()` (lines 342-431) calls:
```go
goplusData, err := s.goPlus.GetTokenSecurity(ctx, tokenAddress)
pairData, derr := s.dexScreener.GetPairData(ctx, pairAddress)
holders, herr := s.bscScan.FetchHolderDistribution(ctx, tokenAddress)
history, cerr := s.bscScan.FetchCreatorHistory(ctx, tokenAddress)
```

**After** refactoring, replace with:
```go
goplusData, err := s.base.GetTokenSecurity(ctx, "bsc", tokenAddress)
pairData, derr := s.base.GetPairData(ctx, "bsc", pairAddress)
holders, herr := s.base.GetHolderDistribution(ctx, "bsc", tokenAddress)
history, cerr := s.base.GetCreatorHistory(ctx, "bsc", tokenAddress)
```

The return types from the SDK match the current types, so the rest of `enrichTokenOnce()` normalization logic (`normalizeGoPlus`, `normalizeDEXScreener`) can stay the same — just adapt to use the SDK types.

**Important**: The `normalizeGoPlus()` and `normalizeDEXScreener()` functions in scanner.go should be kept as-is since they are application-specific transformation logic. The SDK returns raw data, and each app normalizes it for its own needs.

### Step 4.4: Modify main.go

**Current** `apps/easymeme/server/cmd/server/main.go` (lines 40-50):
```go
ethClient, err := ethereum.NewClient(cfg.BscRpcHTTP, cfg.BscRpcWS)
// ...
scanner := service.NewScanner(ethClient, repo, wsHub, cfg.BscScanAPIKey)
```

**After** refactoring:
```go
import basesdk "easyweb3/go-sdk"

// Base service client
baseClient := basesdk.NewClient(cfg.BaseServiceURL, cfg.BaseServiceToken)

// Ethereum client — only needed for PairCreated event subscription
ethClient, err := ethereum.NewClient(cfg.BscRpcHTTP, cfg.BscRpcWS)

scanner := service.NewScanner(ethClient, repo, wsHub, baseClient)
```

**Note**: easymeme still needs its own `ethereum.Client` for the PairCreated event subscription (WebSocket/polling). This is application-specific real-time functionality that doesn't belong in base-service. However, `BSC_RPC_HTTP` and `BSC_RPC_WS` can be optionally proxied through base-service in the future.

### Step 4.5: Remove migrated files from easymeme

After verifying the refactored easymeme builds and works:

```bash
rm apps/easymeme/server/internal/service/goplus.go
rm apps/easymeme/server/internal/service/bscscan.go
rm apps/easymeme/server/internal/service/dexscreener.go
```

Keep `scanner.go` (it now uses the SDK instead of direct API clients).

### Step 4.6: Simplify easymeme wallet handler

The wallet management (create, encrypt, decrypt, execute trade) should be moved to base-service. Easymeme's wallet handler becomes a thin proxy.

Option A (recommended for Phase 1): Keep wallet handler in easymeme as-is temporarily, migrate in a later phase.

Option B (full migration): Replace the wallet handler to call base-service via SDK. This is more complex because the wallet handler has deep integration with AI trade recording.

**Recommendation**: Start with Option A. The wallet handler migration is a separate task that requires careful handling of the trade execution flow.

---

## 9. Phase 5: Docker Compose Layered Orchestration

### Step 9.1: Infrastructure layer

File: `infra/docker-compose.yml`

```yaml
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: easyweb3
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-easyweb3}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init-databases.sh:/docker-entrypoint-initdb.d/init.sh
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U easyweb3"]
      interval: 5s
      timeout: 5s
      retries: 10

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 10

  base-service:
    build:
      context: ../services/base
    environment:
      - PORT=8081
      - DATABASE_URL=postgres://easyweb3:${POSTGRES_PASSWORD:-easyweb3}@postgres:5432/base_db?sslmode=disable
      - REDIS_URL=redis://redis:6379
      - AUTO_MIGRATE=true
      - BSC_RPC_HTTP=${BSC_RPC_HTTP:-https://bsc-dataseed.bnbchain.org}
      - BSC_RPC_WS=${BSC_RPC_WS:-}
      - BSCSCAN_API_KEY=${BSCSCAN_API_KEY:-}
      - TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN:-}
      - WALLET_MASTER_KEY=${WALLET_MASTER_KEY:-}
      - SERVICE_TOKEN_EASYMEME=${SERVICE_TOKEN_EASYMEME:-dev-token}
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    restart: unless-stopped

  nginx:
    image: nginx:1.25-alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/conf.d/default.conf:ro
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:

networks:
  default:
    name: easyweb3
```

### Step 9.2: Database initialization script

File: `infra/scripts/init-databases.sh`

```bash
#!/bin/bash
set -e

# This script runs inside the PostgreSQL container on first init.
# It creates separate databases for each service/app.

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE DATABASE base_db;
    CREATE DATABASE easymeme_db;
    -- Add future app databases here:
    -- CREATE DATABASE future_app_db;
EOSQL
```

### Step 9.3: Nginx gateway

File: `infra/nginx/nginx.conf`

```nginx
upstream base_service {
    server base-service:8081;
}

upstream easymeme_server {
    server easymeme-server:8080;
}

upstream easymeme_web {
    server easymeme-web:3000;
}

# api.easyweb3.tools → base-service
server {
    listen 80;
    server_name api.easyweb3.tools;

    location / {
        proxy_pass http://base_service;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}

# meme.easyweb3.tools → easymeme
server {
    listen 80;
    server_name meme.easyweb3.tools;

    location /api/ {
        proxy_pass http://easymeme_server;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }

    location /ws {
        proxy_pass http://easymeme_server;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    location / {
        proxy_pass http://easymeme_web;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}

# Default fallback (localhost dev)
server {
    listen 80 default_server;

    location /api/v1/ {
        proxy_pass http://base_service;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location /api/ {
        proxy_pass http://easymeme_server;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location /ws {
        proxy_pass http://easymeme_server;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    location / {
        proxy_pass http://easymeme_web;
        proxy_set_header Host $host;
    }
}
```

### Step 9.4: Easymeme app layer

File: `apps/easymeme/docker-compose.yml`

```yaml
# Usage: docker compose -f infra/docker-compose.yml -f apps/easymeme/docker-compose.yml up
services:
  easymeme-server:
    build:
      context: ./server
    environment:
      - PORT=8080
      - DATABASE_URL=postgres://easyweb3:${POSTGRES_PASSWORD:-easyweb3}@postgres:5432/easymeme_db?sslmode=disable
      - AUTO_MIGRATE=true
      - BASE_SERVICE_URL=http://base-service:8081
      - BASE_SERVICE_TOKEN=${SERVICE_TOKEN_EASYMEME:-dev-token}
      - BSC_RPC_HTTP=${BSC_RPC_HTTP:-https://bsc-dataseed.bnbchain.org}
      - BSC_RPC_WS=${BSC_RPC_WS:-}
      - EASYMEME_API_KEY=${EASYMEME_API_KEY:-}
      - EASYMEME_USER_ID=${EASYMEME_USER_ID:-default}
      - EASYMEME_API_HMAC_SECRET=${EASYMEME_API_HMAC_SECRET:-}
      - WALLET_MASTER_KEY=${WALLET_MASTER_KEY:-}
      - CORS_ALLOWED_ORIGINS=${CORS_ALLOWED_ORIGINS:-http://localhost:3000}
    depends_on:
      - base-service
    restart: unless-stopped

  easymeme-web:
    build:
      context: ./web
      args:
        NEXT_PUBLIC_API_URL: http://easymeme-server:8080
        NEXT_PUBLIC_WS_URL: ws://easymeme-server:8080/ws
    environment:
      - NEXT_PUBLIC_API_URL=http://easymeme-server:8080
      - NEXT_PUBLIC_WS_URL=ws://easymeme-server:8080/ws
    depends_on:
      - easymeme-server
    restart: unless-stopped

  easymeme-openclaw:
    build:
      context: ./openclaw-skill
    environment:
      - HOME=/home/node
      - EASYMEME_SERVER_URL=http://easymeme-server:8080
      - EASYMEME_API_KEY=${EASYMEME_API_KEY:-}
      - EASYMEME_USER_ID=${EASYMEME_USER_ID:-default}
      - EASYMEME_NOTIFY_CHANNEL=${EASYMEME_NOTIFY_CHANNEL:-}
      - EASYMEME_NOTIFY_TO=${EASYMEME_NOTIFY_TO:-}
      - OPENCLAW_STATE_DIR=/home/node/.openclaw
      - OPENCLAW_GATEWAY_PORT=${OPENCLAW_GATEWAY_PORT:-18789}
      - OPENCLAW_GATEWAY_TOKEN=${OPENCLAW_GATEWAY_TOKEN:-}
      - OPENCLAW_GATEWAY_BIND=${OPENCLAW_GATEWAY_BIND:-lan}
    volumes:
      - openclaw_state:/home/node/.openclaw
    depends_on:
      - easymeme-server
    ports:
      - "18789:18789"
    user: "0"
    restart: unless-stopped
    command: ["/app/entrypoint.sh"]

volumes:
  openclaw_state:

networks:
  default:
    name: easyweb3
    external: true
```

### Step 9.5: Environment template

File: `infra/.env.example`

```bash
# ============================================
# EasyWeb3 Infrastructure Environment Variables
# ============================================
# Copy to .env and fill in actual values.

# --- PostgreSQL ---
POSTGRES_PASSWORD=easyweb3

# --- Blockchain RPC ---
BSC_RPC_HTTP=https://bsc-dataseed.bnbchain.org
BSC_RPC_WS=

# --- Third-Party API Keys (configured once, used by base-service) ---
BSCSCAN_API_KEY=
TELEGRAM_BOT_TOKEN=

# --- Security ---
WALLET_MASTER_KEY=

# --- Service Auth Tokens (base-service validates these) ---
SERVICE_TOKEN_EASYMEME=change-me-to-a-random-string
# SERVICE_TOKEN_FUTURE_APP=another-random-string

# --- EasyMeme App-Specific ---
EASYMEME_API_KEY=
EASYMEME_USER_ID=default
EASYMEME_API_HMAC_SECRET=
EASYMEME_NOTIFY_CHANNEL=
EASYMEME_NOTIFY_TO=
OPENCLAW_GATEWAY_PORT=18789
OPENCLAW_GATEWAY_TOKEN=
OPENCLAW_GATEWAY_BIND=lan
CORS_ALLOWED_ORIGINS=http://localhost:3000
```

### Step 9.6: Makefile

File: `Makefile`

```makefile
.PHONY: dev infra stop clean

# Start infra + easymeme (development)
dev:
	docker compose -f infra/docker-compose.yml -f apps/easymeme/docker-compose.yml up --build

# Start only infrastructure (base-service, PostgreSQL, Redis, Nginx)
infra:
	docker compose -f infra/docker-compose.yml up --build

# Stop everything
stop:
	docker compose -f infra/docker-compose.yml -f apps/easymeme/docker-compose.yml down

# Stop and remove volumes (DESTRUCTIVE)
clean:
	docker compose -f infra/docker-compose.yml -f apps/easymeme/docker-compose.yml down -v

# Build base-service only
build-base:
	cd services/base && go build ./cmd/server/

# Build easymeme server only
build-easymeme:
	cd apps/easymeme/server && go build ./cmd/server/

# Build all Go projects
build-all: build-base build-easymeme

# Run tests
test-base:
	cd services/base && go test ./...

test-easymeme:
	cd apps/easymeme/server && go test ./...

test-all: test-base test-easymeme
```

---

## 10. Database Boundary Design

### base_db (owned by base-service)

```sql
-- Token security analysis cache
CREATE TABLE token_security_cache (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chain       VARCHAR(20) NOT NULL,        -- bsc, ethereum, arbitrum
    address     VARCHAR(66) NOT NULL,        -- token contract address
    data        JSONB NOT NULL,              -- raw GoPlus response
    fetched_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(chain, address)
);

-- Market data cache (for data that needs persistence beyond Redis TTL)
CREATE TABLE market_data_cache (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chain          VARCHAR(20) NOT NULL,
    pair_address   VARCHAR(66) NOT NULL,
    data           JSONB NOT NULL,
    fetched_at     TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(chain, pair_address)
);

-- Managed wallets (migrated from easymeme)
CREATE TABLE managed_wallets (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_id     VARCHAR(50) NOT NULL,     -- which app owns this wallet
    user_id        VARCHAR(100) NOT NULL,
    address        VARCHAR(66) UNIQUE NOT NULL,
    encrypted_key  BYTEA NOT NULL,
    balance        DOUBLE PRECISION DEFAULT 0,
    max_balance    DOUBLE PRECISION DEFAULT 5,
    created_at     TIMESTAMP DEFAULT NOW(),
    updated_at     TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_wallets_service_user ON managed_wallets(service_id, user_id);

-- Notification logs
CREATE TABLE notification_logs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_id  VARCHAR(50) NOT NULL,
    channel     VARCHAR(20) NOT NULL,
    recipient   VARCHAR(200) NOT NULL,
    message     TEXT NOT NULL,
    status      VARCHAR(20) DEFAULT 'sent',
    created_at  TIMESTAMP DEFAULT NOW()
);
```

### easymeme_db (owned by easymeme)

Same tables as current easymeme, minus `managed_wallets`:
- `tokens` (with golden_dog_score, analysis_result, etc.)
- `trades`
- `ai_trades`
- `ai_positions`
- `token_market_snapshots`
- `token_alerts`
- `token_price_snapshots`
- `wallet_configs` (app-specific auto-trade config)

---

## 11. API Specification for base-service

### Endpoints Summary

| Method | Path | Description | Cache TTL |
|--------|------|-------------|-----------|
| `GET` | `/health` | Health check | - |
| `GET` | `/api/v1/tokens/{chain}/{address}/security` | GoPlus security data | 1 hour |
| `GET` | `/api/v1/tokens/{chain}/{address}/holders` | BSCScan holder distribution | 30 min |
| `GET` | `/api/v1/tokens/{chain}/{address}/creator` | BSCScan creator history | 1 hour |
| `GET` | `/api/v1/market/{chain}/pairs/{pairAddress}` | DEXScreener pair data | 1 min |
| `POST` | `/api/v1/wallets` | Create managed wallet | - |
| `GET` | `/api/v1/wallets/{walletId}/balance` | Get wallet balance | - |
| `POST` | `/api/v1/wallets/{walletId}/execute` | Execute on-chain trade | - |
| `POST` | `/api/v1/notifications/send` | Send notification | - |

### Response Format

All responses follow this envelope:

```json
{
    "data": { ... },
    "cached": true,
    "error": "optional error message"
}
```

### Authentication

All `/api/v1/*` endpoints require:
```
Authorization: Bearer <service_token>
```

Where `service_token` matches one of the `SERVICE_TOKEN_*` environment variables configured on base-service.

---

## 12. Authentication Between Services

### Service-to-Service Auth Flow

```
easymeme-server                    base-service
      │                                  │
      │  GET /api/v1/tokens/bsc/0x.../security
      │  Authorization: Bearer <SERVICE_TOKEN_EASYMEME>
      │ ────────────────────────────────►│
      │                                  │ Validate token against
      │                                  │ SERVICE_TOKEN_EASYMEME env var
      │                                  │
      │  200 OK                          │
      │  {"data": {...}, "cached": true} │
      │ ◄────────────────────────────────│
```

### Key Design Decisions

1. **Simple Bearer token** (not JWT) — sufficient for internal service mesh, no token expiry complexity.
2. **One token per app** — `SERVICE_TOKEN_EASYMEME`, `SERVICE_TOKEN_APP2`, etc.
3. **Token passed via env var** — both the caller and base-service receive the same token from `.env`.
4. **No HMAC needed** — services communicate over Docker internal network, not exposed to internet.

---

## 13. Migration Checklist

### Phase 1: Repository Restructuring
- [ ] Create monorepo directory structure
- [ ] Move easymeme into `apps/easymeme/`
- [ ] Update `go.mod` module path
- [ ] Update all import paths in Go files
- [ ] Create `go.work` file
- [ ] Verify `go build` succeeds for easymeme
- [ ] Verify existing docker-compose still works

### Phase 2: Build base-service
- [ ] Create `services/base/go.mod`
- [ ] Create config.go with all API key bindings
- [ ] Copy and adapt `goplus.go` (add chain parameter)
- [ ] Copy and adapt `bscscan.go` (add chain parameter)
- [ ] Copy and adapt `dexscreener.go` (add chain parameter)
- [ ] Copy `pkg/ethereum/client.go`
- [ ] Create `cache.go` (Redis caching layer)
- [ ] Create `telegram.go` (notification service)
- [ ] Create handlers: token_security, market_data, chain_data, wallet, notification
- [ ] Create service auth middleware
- [ ] Create router.go
- [ ] Create main.go entry point
- [ ] Create Dockerfile
- [ ] Verify `go build` succeeds
- [ ] Write basic tests for each handler

### Phase 3: Build Go SDK
- [ ] Create `packages/go-sdk/go.mod`
- [ ] Implement `client.go` (HTTP client with auth)
- [ ] Implement `token_security.go`
- [ ] Implement `market_data.go`
- [ ] Implement `chain_data.go`
- [ ] Implement `notification.go`
- [ ] Write tests for SDK client

### Phase 4: Refactor easymeme
- [ ] Add `BASE_SERVICE_URL` and `BASE_SERVICE_TOKEN` to easymeme config
- [ ] Add go-sdk dependency to easymeme
- [ ] Modify `Scanner` struct to use `basesdk.Client`
- [ ] Modify `NewScanner()` to accept `basesdk.Client`
- [ ] Modify `enrichTokenOnce()` to call SDK instead of direct API
- [ ] Modify `refreshTokenMarketData()` to call SDK
- [ ] Update `main.go` to create basesdk.Client and pass to Scanner
- [ ] Remove `goplus.go`, `bscscan.go`, `dexscreener.go` from easymeme
- [ ] Verify `go build` succeeds
- [ ] Run full integration test

### Phase 5: Docker Compose Layered Orchestration
- [ ] Create `infra/docker-compose.yml` (PostgreSQL, Redis, base-service, Nginx)
- [ ] Create `infra/scripts/init-databases.sh`
- [ ] Create `infra/nginx/nginx.conf`
- [ ] Create `infra/.env.example`
- [ ] Create `apps/easymeme/docker-compose.yml`
- [ ] Create `Makefile`
- [ ] Verify `make dev` brings up all services
- [ ] Verify easymeme can reach base-service over Docker network
- [ ] Verify Nginx routes requests correctly
- [ ] Remove old `docker-compose.yml` from project root

### Validation
- [ ] New token detected by scanner → enriched via base-service → data correct
- [ ] DEXScreener market data refreshes work via base-service
- [ ] Redis caching reduces upstream API calls (check logs)
- [ ] Multiple concurrent requests for same token → only one upstream call
- [ ] Wallet creation/trade execution still works
- [ ] OpenClaw agent still functions correctly
- [ ] Web frontend loads and displays data correctly
