# EasyMeme MVP 开发计划

> 本文档为 AI 代码助手（Codex）设计，包含完整的开发指令和代码模板。
> 请按照 TASK 顺序依次完成开发。

---

## 项目概述

**EasyMeme** 是一个 BNB Chain Meme 币扫描和交易工具，包含：
- 实时监控 PancakeSwap 新池创建
- AI 风险检测（貔貅、权限、税率）
- 一键快捷买入
- WebSocket 实时推送

**技术栈：**
- 后端：Golang 1.22+ / Gin / GORM / go-ethereum
- 前端：Next.js 14 / TypeScript / wagmi / viem
- 数据库：PostgreSQL 16
- 缓存：Redis 7

---

## 目录结构

请按以下结构创建项目：

```
easymeme/
├── server/                       # Golang 后端
│   ├── cmd/
│   │   └── server/
│   │       └── main.go
│   ├── internal/
│   │   ├── config/
│   │   │   └── config.go
│   │   ├── model/
│   │   │   ├── token.go
│   │   │   └── trade.go
│   │   ├── repository/
│   │   │   └── repository.go
│   │   ├── service/
│   │   │   ├── scanner.go
│   │   │   ├── analyzer.go
│   │   │   └── trader.go
│   │   ├── handler/
│   │   │   ├── token.go
│   │   │   ├── trade.go
│   │   │   └── websocket.go
│   │   └── router/
│   │       └── router.go
│   ├── pkg/
│   │   └── ethereum/
│   │       └── client.go
│   ├── migrations/
│   │   └── 001_init.sql
│   ├── go.mod
│   ├── go.sum
│   ├── Dockerfile
│   └── .env.example
│
├── web/                          # Next.js 前端
│   ├── app/
│   │   ├── layout.tsx
│   │   ├── page.tsx
│   │   └── dashboard/
│   │       └── page.tsx
│   ├── components/
│   │   ├── providers.tsx
│   │   ├── token-list.tsx
│   │   ├── token-card.tsx
│   │   ├── risk-badge.tsx
│   │   └── trade-panel.tsx
│   ├── lib/
│   │   ├── wagmi.ts
│   │   └── api.ts
│   ├── package.json
│   ├── next.config.js
│   ├── tailwind.config.js
│   └── .env.local
│
├── contracts/
│   ├── pancake_factory_v2.json
│   └── pancake_router_v2.json
│
├── docker-compose.yml
├── Makefile
└── README.md
```

---

## TASK 1: 初始化后端项目

### 1.1 创建 go.mod

**文件：** `server/go.mod`

```go
module easymeme

go 1.22

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/gorilla/websocket v1.5.1
    github.com/ethereum/go-ethereum v1.13.14
    github.com/redis/go-redis/v9 v9.4.0
    github.com/joho/godotenv v1.5.1
    github.com/shopspring/decimal v1.3.1
    gorm.io/gorm v1.25.7
    gorm.io/driver/postgres v1.5.6
)
```

### 1.2 创建配置管理

**文件：** `server/internal/config/config.go`

```go
package config

import (
    "os"

    "github.com/joho/godotenv"
)

type Config struct {
    Port         string
    DatabaseURL  string
    RedisURL     string
    BscRpcHTTP   string
    BscRpcWS     string
    BscscanAPIKey string
}

func Load() (*Config, error) {
    godotenv.Load()

    return &Config{
        Port:          getEnv("PORT", "8080"),
        DatabaseURL:   getEnv("DATABASE_URL", ""),
        RedisURL:      getEnv("REDIS_URL", ""),
        BscRpcHTTP:    getEnv("BSC_RPC_HTTP", "https://bsc-dataseed.binance.org"),
        BscRpcWS:      getEnv("BSC_RPC_WS", "wss://bsc-ws-node.nariox.org"),
        BscscanAPIKey: getEnv("BSCSCAN_API_KEY", ""),
    }, nil
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

### 1.3 创建环境变量模板

**文件：** `server/.env.example`

```bash
PORT=8080
DATABASE_URL=postgres://postgres:postgres@localhost:5432/easymeme?sslmode=disable
REDIS_URL=redis://localhost:6379
BSC_RPC_HTTP=https://bsc-dataseed.binance.org
BSC_RPC_WS=wss://bsc-ws-node.nariox.org
BSCSCAN_API_KEY=your_api_key_here
```

---

## TASK 2: 创建数据模型

### 2.1 Token 模型

**文件：** `server/internal/model/token.go`

```go
package model

import (
    "time"

    "github.com/shopspring/decimal"
    "gorm.io/datatypes"
)

type Token struct {
    ID               string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
    Address          string          `gorm:"uniqueIndex;not null" json:"address"`
    Name             string          `json:"name"`
    Symbol           string          `json:"symbol"`
    Decimals         int             `gorm:"default:18" json:"decimals"`
    PairAddress      string          `json:"pair_address"`
    Dex              string          `gorm:"default:pancakeswap" json:"dex"`
    InitialLiquidity decimal.Decimal `gorm:"type:decimal(36,18)" json:"initial_liquidity"`
    RiskScore        int             `json:"risk_score"`
    RiskLevel        string          `json:"risk_level"` // safe, warning, danger
    RiskDetails      datatypes.JSON  `json:"risk_details"`
    IsHoneypot       bool            `gorm:"default:false" json:"is_honeypot"`
    BuyTax           float64         `json:"buy_tax"`
    SellTax          float64         `json:"sell_tax"`
    CreatorAddress   string          `json:"creator_address"`
    CreatedAt        time.Time       `gorm:"autoCreateTime;index" json:"created_at"`
    UpdatedAt        time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Token) TableName() string {
    return "tokens"
}

// RiskDetails JSON 结构
type RiskDetailsJSON struct {
    CanMint           bool    `json:"can_mint"`
    CanPause          bool    `json:"can_pause"`
    CanBlacklist      bool    `json:"can_blacklist"`
    OwnerCanChangeTax bool    `json:"owner_can_change_tax"`
    LPLocked          bool    `json:"lp_locked"`
    LPLockDays        int     `json:"lp_lock_days"`
    ContractVerified  bool    `json:"contract_verified"`
    Top10Holding      float64 `json:"top10_holding"`
}
```

### 2.2 Trade 模型

**文件：** `server/internal/model/trade.go`

```go
package model

import (
    "time"

    "github.com/shopspring/decimal"
)

type Trade struct {
    ID           string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
    UserAddress  string          `gorm:"index;not null" json:"user_address"`
    TokenAddress string          `gorm:"index;not null" json:"token_address"`
    TokenSymbol  string          `json:"token_symbol"`
    Type         string          `json:"type"` // buy, sell
    AmountIn     decimal.Decimal `gorm:"type:decimal(36,18)" json:"amount_in"`
    AmountOut    decimal.Decimal `gorm:"type:decimal(36,18)" json:"amount_out"`
    TxHash       string          `gorm:"uniqueIndex" json:"tx_hash"`
    Status       string          `json:"status"` // pending, success, failed
    GasUsed      decimal.Decimal `gorm:"type:decimal(36,18)" json:"gas_used"`
    CreatedAt    time.Time       `gorm:"autoCreateTime" json:"created_at"`
}

func (Trade) TableName() string {
    return "trades"
}
```

### 2.3 数据库迁移 SQL

**文件：** `server/migrations/001_init.sql`

```sql
-- 启用 uuid 扩展
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 创建 tokens 表
CREATE TABLE IF NOT EXISTS tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    address VARCHAR(42) UNIQUE NOT NULL,
    name VARCHAR(255),
    symbol VARCHAR(50),
    decimals INTEGER DEFAULT 18,
    pair_address VARCHAR(42),
    dex VARCHAR(50) DEFAULT 'pancakeswap',
    initial_liquidity DECIMAL(36, 18),
    risk_score INTEGER,
    risk_level VARCHAR(20),
    risk_details JSONB,
    is_honeypot BOOLEAN DEFAULT FALSE,
    buy_tax DECIMAL(5, 2),
    sell_tax DECIMAL(5, 2),
    creator_address VARCHAR(42),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 创建 trades 表
CREATE TABLE IF NOT EXISTS trades (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_address VARCHAR(42) NOT NULL,
    token_address VARCHAR(42) NOT NULL,
    token_symbol VARCHAR(50),
    type VARCHAR(10) NOT NULL,
    amount_in DECIMAL(36, 18),
    amount_out DECIMAL(36, 18),
    tx_hash VARCHAR(66) UNIQUE,
    status VARCHAR(20),
    gas_used DECIMAL(36, 18),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_tokens_created_at ON tokens(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_tokens_risk_score ON tokens(risk_score);
CREATE INDEX IF NOT EXISTS idx_trades_user_address ON trades(user_address);
CREATE INDEX IF NOT EXISTS idx_trades_token_address ON trades(token_address);
```

---

## TASK 3: 创建数据库操作层

**文件：** `server/internal/repository/repository.go`

```go
package repository

import (
    "context"

    "easymeme/internal/model"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

type Repository struct {
    db *gorm.DB
}

func New(databaseURL string) (*Repository, error) {
    db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    if err != nil {
        return nil, err
    }

    // 自动迁移
    db.AutoMigrate(&model.Token{}, &model.Trade{})

    return &Repository{db: db}, nil
}

// Token 操作

func (r *Repository) CreateToken(ctx context.Context, token *model.Token) error {
    return r.db.WithContext(ctx).Create(token).Error
}

func (r *Repository) GetTokenByAddress(ctx context.Context, address string) (*model.Token, error) {
    var token model.Token
    err := r.db.WithContext(ctx).Where("address = ?", address).First(&token).Error
    if err != nil {
        return nil, err
    }
    return &token, nil
}

func (r *Repository) GetLatestTokens(ctx context.Context, limit int) ([]model.Token, error) {
    var tokens []model.Token
    err := r.db.WithContext(ctx).
        Order("created_at DESC").
        Limit(limit).
        Find(&tokens).Error
    return tokens, err
}

func (r *Repository) UpdateToken(ctx context.Context, token *model.Token) error {
    return r.db.WithContext(ctx).Save(token).Error
}

func (r *Repository) TokenExists(ctx context.Context, address string) bool {
    var count int64
    r.db.WithContext(ctx).Model(&model.Token{}).Where("address = ?", address).Count(&count)
    return count > 0
}

// Trade 操作

func (r *Repository) CreateTrade(ctx context.Context, trade *model.Trade) error {
    return r.db.WithContext(ctx).Create(trade).Error
}

func (r *Repository) GetTradesByUser(ctx context.Context, userAddress string, limit int) ([]model.Trade, error) {
    var trades []model.Trade
    err := r.db.WithContext(ctx).
        Where("user_address = ?", userAddress).
        Order("created_at DESC").
        Limit(limit).
        Find(&trades).Error
    return trades, err
}

func (r *Repository) UpdateTradeStatus(ctx context.Context, txHash, status string) error {
    return r.db.WithContext(ctx).
        Model(&model.Trade{}).
        Where("tx_hash = ?", txHash).
        Update("status", status).Error
}
```

---

## TASK 4: 创建以太坊客户端封装

**文件：** `server/pkg/ethereum/client.go`

```go
package ethereum

import (
    "context"
    "math/big"

    "github.com/ethereum/go-ethereum"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/ethclient"
)

// 常量定义
const (
    PancakeFactoryV2 = "0xcA143Ce32Fe78f1f7019d7d551a6402fC5350c73"
    PancakeRouterV2  = "0x10ED43C718714eb63d5aA57B78B54704E256024E"
    WBNB             = "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c"
)

// PairCreated 事件签名
var PairCreatedTopic = common.HexToHash("0x0d3648bd0f6ba80134a33ba9275ac585d9d315f0ad8355cddefde31afa28d0e9")

type Client struct {
    http *ethclient.Client
    ws   *ethclient.Client
}

func NewClient(httpURL, wsURL string) (*Client, error) {
    httpClient, err := ethclient.Dial(httpURL)
    if err != nil {
        return nil, err
    }

    wsClient, err := ethclient.Dial(wsURL)
    if err != nil {
        return nil, err
    }

    return &Client{
        http: httpClient,
        ws:   wsClient,
    }, nil
}

func (c *Client) SubscribePairCreated(ctx context.Context) (chan types.Log, ethereum.Subscription, error) {
    query := ethereum.FilterQuery{
        Addresses: []common.Address{common.HexToAddress(PancakeFactoryV2)},
        Topics:    [][]common.Hash{{PairCreatedTopic}},
    }

    logs := make(chan types.Log)
    sub, err := c.ws.SubscribeFilterLogs(ctx, query, logs)
    if err != nil {
        return nil, nil, err
    }

    return logs, sub, nil
}

func (c *Client) GetTokenInfo(ctx context.Context, tokenAddr common.Address) (name, symbol string, decimals uint8, err error) {
    // ERC20 name()
    nameData, err := c.http.CallContract(ctx, ethereum.CallMsg{
        To:   &tokenAddr,
        Data: common.Hex2Bytes("06fdde03"), // name() selector
    }, nil)
    if err == nil && len(nameData) > 0 {
        name = parseString(nameData)
    }

    // ERC20 symbol()
    symbolData, err := c.http.CallContract(ctx, ethereum.CallMsg{
        To:   &tokenAddr,
        Data: common.Hex2Bytes("95d89b41"), // symbol() selector
    }, nil)
    if err == nil && len(symbolData) > 0 {
        symbol = parseString(symbolData)
    }

    // ERC20 decimals()
    decimalsData, err := c.http.CallContract(ctx, ethereum.CallMsg{
        To:   &tokenAddr,
        Data: common.Hex2Bytes("313ce567"), // decimals() selector
    }, nil)
    if err == nil && len(decimalsData) > 0 {
        decimals = uint8(new(big.Int).SetBytes(decimalsData).Uint64())
    } else {
        decimals = 18
    }

    return name, symbol, decimals, nil
}

func (c *Client) GetPairReserves(ctx context.Context, pairAddr common.Address) (reserve0, reserve1 *big.Int, err error) {
    data, err := c.http.CallContract(ctx, ethereum.CallMsg{
        To:   &pairAddr,
        Data: common.Hex2Bytes("0902f1ac"), // getReserves() selector
    }, nil)
    if err != nil {
        return nil, nil, err
    }

    if len(data) >= 64 {
        reserve0 = new(big.Int).SetBytes(data[0:32])
        reserve1 = new(big.Int).SetBytes(data[32:64])
    }
    return reserve0, reserve1, nil
}

func (c *Client) SimulateSell(ctx context.Context, tokenAddr common.Address, amount *big.Int) error {
    router := common.HexToAddress(PancakeRouterV2)
    wbnb := common.HexToAddress(WBNB)

    // 构建 swapExactTokensForETH 调用
    // selector: 0x18cbafe5
    data := common.Hex2Bytes("18cbafe5")
    // 这里简化处理，实际需要完整的 ABI 编码

    _, err := c.http.CallContract(ctx, ethereum.CallMsg{
        To:   &router,
        Data: data,
    }, nil)

    return err
}

func (c *Client) Close() {
    c.http.Close()
    c.ws.Close()
}

// 辅助函数：解析 ABI 编码的字符串
func parseString(data []byte) string {
    if len(data) < 64 {
        return ""
    }
    offset := new(big.Int).SetBytes(data[0:32]).Uint64()
    if offset+32 > uint64(len(data)) {
        return ""
    }
    length := new(big.Int).SetBytes(data[offset : offset+32]).Uint64()
    if offset+32+length > uint64(len(data)) {
        return ""
    }
    return string(data[offset+32 : offset+32+length])
}
```

---

## TASK 5: 创建扫描服务

**文件：** `server/internal/service/scanner.go`

```go
package service

import (
    "context"
    "encoding/json"
    "log"
    "math/big"
    "strings"

    "easymeme/internal/model"
    "easymeme/internal/repository"
    "easymeme/pkg/ethereum"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/shopspring/decimal"
)

type Scanner struct {
    client   *ethereum.Client
    repo     *repository.Repository
    analyzer *Analyzer
    hub      *WebSocketHub
}

func NewScanner(client *ethereum.Client, repo *repository.Repository, analyzer *Analyzer, hub *WebSocketHub) *Scanner {
    return &Scanner{
        client:   client,
        repo:     repo,
        analyzer: analyzer,
        hub:      hub,
    }
}

func (s *Scanner) Start(ctx context.Context) error {
    logs, sub, err := s.client.SubscribePairCreated(ctx)
    if err != nil {
        return err
    }

    log.Println("[Scanner] Started listening for PairCreated events...")

    go func() {
        for {
            select {
            case err := <-sub.Err():
                log.Printf("[Scanner] Subscription error: %v", err)
                return
            case vLog := <-logs:
                go s.handlePairCreated(ctx, vLog)
            case <-ctx.Done():
                log.Println("[Scanner] Stopping...")
                return
            }
        }
    }()

    return nil
}

func (s *Scanner) handlePairCreated(ctx context.Context, vLog types.Log) {
    // 解析事件
    token0 := common.HexToAddress(vLog.Topics[1].Hex())
    token1 := common.HexToAddress(vLog.Topics[2].Hex())
    pairAddr := common.BytesToAddress(vLog.Data[:32])

    // 确定目标代币（非 WBNB 的那个）
    wbnb := common.HexToAddress(ethereum.WBNB)
    var targetToken common.Address
    if strings.EqualFold(token0.Hex(), wbnb.Hex()) {
        targetToken = token1
    } else if strings.EqualFold(token1.Hex(), wbnb.Hex()) {
        targetToken = token0
    } else {
        return // 非 WBNB 配对，跳过
    }

    log.Printf("[Scanner] New pair: %s, Token: %s", pairAddr.Hex(), targetToken.Hex())

    // 检查是否已存在
    if s.repo.TokenExists(ctx, targetToken.Hex()) {
        return
    }

    // 获取代币信息
    name, symbol, decimals, _ := s.client.GetTokenInfo(ctx, targetToken)

    // 获取初始流动性
    reserve0, reserve1, _ := s.client.GetPairReserves(ctx, pairAddr)
    var liquidity *big.Int
    if strings.EqualFold(token0.Hex(), wbnb.Hex()) {
        liquidity = reserve0
    } else {
        liquidity = reserve1
    }

    // 创建 Token 记录
    token := &model.Token{
        Address:          targetToken.Hex(),
        Name:             name,
        Symbol:           symbol,
        Decimals:         int(decimals),
        PairAddress:      pairAddr.Hex(),
        Dex:              "pancakeswap",
        InitialLiquidity: decimal.NewFromBigInt(liquidity, -18),
    }

    // 分析风险
    riskResult := s.analyzer.Analyze(ctx, targetToken)
    token.RiskScore = riskResult.Score
    token.RiskLevel = string(riskResult.Level)
    token.IsHoneypot = riskResult.IsHoneypot
    token.BuyTax = riskResult.BuyTax
    token.SellTax = riskResult.SellTax

    detailsJSON, _ := json.Marshal(riskResult.Details)
    token.RiskDetails = detailsJSON

    // 保存到数据库
    if err := s.repo.CreateToken(ctx, token); err != nil {
        log.Printf("[Scanner] Save token error: %v", err)
        return
    }

    // 推送给前端
    s.hub.Broadcast(map[string]interface{}{
        "type":  "new_token",
        "token": token,
    })

    log.Printf("[Scanner] Token saved: %s (%s), Risk: %d", symbol, targetToken.Hex(), token.RiskScore)
}
```

---

## TASK 6: 创建风险分析服务

**文件：** `server/internal/service/analyzer.go`

```go
package service

import (
    "context"

    "easymeme/pkg/ethereum"

    "github.com/ethereum/go-ethereum/common"
)

type RiskLevel string

const (
    RiskSafe    RiskLevel = "safe"
    RiskWarning RiskLevel = "warning"
    RiskDanger  RiskLevel = "danger"
)

type RiskDetails struct {
    CanMint           bool    `json:"can_mint"`
    CanPause          bool    `json:"can_pause"`
    CanBlacklist      bool    `json:"can_blacklist"`
    OwnerCanChangeTax bool    `json:"owner_can_change_tax"`
    LPLocked          bool    `json:"lp_locked"`
    ContractVerified  bool    `json:"contract_verified"`
    Top10Holding      float64 `json:"top10_holding"`
}

type RiskResult struct {
    Score      int         `json:"score"`
    Level      RiskLevel   `json:"level"`
    IsHoneypot bool        `json:"is_honeypot"`
    BuyTax     float64     `json:"buy_tax"`
    SellTax    float64     `json:"sell_tax"`
    Details    RiskDetails `json:"details"`
}

type Analyzer struct {
    client *ethereum.Client
}

func NewAnalyzer(client *ethereum.Client) *Analyzer {
    return &Analyzer{client: client}
}

func (a *Analyzer) Analyze(ctx context.Context, tokenAddr common.Address) *RiskResult {
    result := &RiskResult{
        Score: 100,
        Level: RiskSafe,
    }

    // 1. 检测貔貅
    isHoneypot := a.checkHoneypot(ctx, tokenAddr)
    if isHoneypot {
        result.Score = 0
        result.Level = RiskDanger
        result.IsHoneypot = true
        return result
    }

    // 2. 获取买卖税率
    buyTax, sellTax := a.getTaxRates(ctx, tokenAddr)
    result.BuyTax = buyTax
    result.SellTax = sellTax

    // 3. 检测合约权限
    details := a.analyzeContract(ctx, tokenAddr)
    result.Details = details

    // 4. 计算风险分数
    result.Score = a.calculateScore(details, buyTax, sellTax)
    result.Level = a.getLevel(result.Score)

    return result
}

func (a *Analyzer) checkHoneypot(ctx context.Context, tokenAddr common.Address) bool {
    // 模拟卖出交易
    // 如果卖出失败，则为貔貅
    err := a.client.SimulateSell(ctx, tokenAddr, nil)
    return err != nil
}

func (a *Analyzer) getTaxRates(ctx context.Context, tokenAddr common.Address) (buyTax, sellTax float64) {
    // TODO: 通过模拟交易计算实际税率
    // 对比预期输出和实际输出的差异
    return 0, 0
}

func (a *Analyzer) analyzeContract(ctx context.Context, tokenAddr common.Address) RiskDetails {
    // TODO: 分析合约字节码或调用 BSCScan API 获取源码分析
    return RiskDetails{
        ContractVerified: true, // 默认值，需要调用 BSCScan API 验证
    }
}

func (a *Analyzer) calculateScore(details RiskDetails, buyTax, sellTax float64) int {
    score := 100

    if details.CanMint {
        score -= 30
    }
    if details.CanPause {
        score -= 20
    }
    if details.CanBlacklist {
        score -= 25
    }
    if details.OwnerCanChangeTax {
        score -= 20
    }
    if !details.LPLocked {
        score -= 20
    }
    if !details.ContractVerified {
        score -= 10
    }
    if buyTax > 10 {
        score -= 15
    }
    if sellTax > 10 {
        score -= 15
    }
    if details.Top10Holding > 50 {
        score -= 15
    }

    if score < 0 {
        return 0
    }
    return score
}

func (a *Analyzer) getLevel(score int) RiskLevel {
    if score >= 70 {
        return RiskSafe
    }
    if score >= 40 {
        return RiskWarning
    }
    return RiskDanger
}
```

---

## TASK 7: 创建 WebSocket Hub

**文件：** `server/internal/handler/websocket.go`

```go
package handler

import (
    "encoding/json"
    "log"
    "net/http"
    "sync"

    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // 允许所有来源，生产环境应该限制
    },
}

type Client struct {
    conn *websocket.Conn
    send chan []byte
}

type WebSocketHub struct {
    clients    map[*Client]bool
    broadcast  chan []byte
    register   chan *Client
    unregister chan *Client
    mu         sync.RWMutex
}

func NewWebSocketHub() *WebSocketHub {
    return &WebSocketHub{
        clients:    make(map[*Client]bool),
        broadcast:  make(chan []byte, 256),
        register:   make(chan *Client),
        unregister: make(chan *Client),
    }
}

func (h *WebSocketHub) Run() {
    for {
        select {
        case client := <-h.register:
            h.mu.Lock()
            h.clients[client] = true
            h.mu.Unlock()
            log.Printf("[WebSocket] Client connected, total: %d", len(h.clients))

        case client := <-h.unregister:
            h.mu.Lock()
            if _, ok := h.clients[client]; ok {
                delete(h.clients, client)
                close(client.send)
            }
            h.mu.Unlock()
            log.Printf("[WebSocket] Client disconnected, total: %d", len(h.clients))

        case message := <-h.broadcast:
            h.mu.RLock()
            for client := range h.clients {
                select {
                case client.send <- message:
                default:
                    close(client.send)
                    delete(h.clients, client)
                }
            }
            h.mu.RUnlock()
        }
    }
}

func (h *WebSocketHub) Broadcast(data interface{}) {
    message, err := json.Marshal(data)
    if err != nil {
        log.Printf("[WebSocket] Marshal error: %v", err)
        return
    }
    h.broadcast <- message
}

func (h *WebSocketHub) HandleWebSocket(c *gin.Context) {
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        log.Printf("[WebSocket] Upgrade error: %v", err)
        return
    }

    client := &Client{
        conn: conn,
        send: make(chan []byte, 256),
    }

    h.register <- client

    go client.writePump()
    go client.readPump(h)
}

func (c *Client) writePump() {
    defer c.conn.Close()

    for message := range c.send {
        if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
            return
        }
    }
}

func (c *Client) readPump(h *WebSocketHub) {
    defer func() {
        h.unregister <- c
        c.conn.Close()
    }()

    for {
        _, _, err := c.conn.ReadMessage()
        if err != nil {
            break
        }
    }
}
```

---

## TASK 8: 创建 HTTP Handlers

**文件：** `server/internal/handler/token.go`

```go
package handler

import (
    "net/http"

    "easymeme/internal/repository"
    "easymeme/internal/service"

    "github.com/ethereum/go-ethereum/common"
    "github.com/gin-gonic/gin"
)

type TokenHandler struct {
    repo     *repository.Repository
    analyzer *service.Analyzer
}

func NewTokenHandler(repo *repository.Repository, analyzer *service.Analyzer) *TokenHandler {
    return &TokenHandler{
        repo:     repo,
        analyzer: analyzer,
    }
}

// GET /api/tokens
func (h *TokenHandler) GetTokens(c *gin.Context) {
    tokens, err := h.repo.GetLatestTokens(c.Request.Context(), 50)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"data": tokens})
}

// GET /api/tokens/:address
func (h *TokenHandler) GetToken(c *gin.Context) {
    address := c.Param("address")

    token, err := h.repo.GetTokenByAddress(c.Request.Context(), address)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Token not found"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"data": token})
}

// POST /api/tokens/:address/analyze
func (h *TokenHandler) AnalyzeToken(c *gin.Context) {
    address := c.Param("address")

    if !common.IsHexAddress(address) {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address"})
        return
    }

    tokenAddr := common.HexToAddress(address)
    result := h.analyzer.Analyze(c.Request.Context(), tokenAddr)

    c.JSON(http.StatusOK, gin.H{"data": result})
}
```

**文件：** `server/internal/handler/trade.go`

```go
package handler

import (
    "net/http"

    "easymeme/internal/model"
    "easymeme/internal/repository"

    "github.com/gin-gonic/gin"
)

type TradeHandler struct {
    repo *repository.Repository
}

func NewTradeHandler(repo *repository.Repository) *TradeHandler {
    return &TradeHandler{repo: repo}
}

// POST /api/trades
func (h *TradeHandler) CreateTrade(c *gin.Context) {
    var trade model.Trade
    if err := c.ShouldBindJSON(&trade); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    trade.Status = "pending"
    if err := h.repo.CreateTrade(c.Request.Context(), &trade); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"data": trade})
}

// GET /api/trades?user=0x...
func (h *TradeHandler) GetTrades(c *gin.Context) {
    userAddress := c.Query("user")
    if userAddress == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "user address required"})
        return
    }

    trades, err := h.repo.GetTradesByUser(c.Request.Context(), userAddress, 50)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": trades})
}

// PATCH /api/trades/:txHash
func (h *TradeHandler) UpdateTradeStatus(c *gin.Context) {
    txHash := c.Param("txHash")

    var body struct {
        Status string `json:"status"`
    }
    if err := c.ShouldBindJSON(&body); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if err := h.repo.UpdateTradeStatus(c.Request.Context(), txHash, body.Status); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "updated"})
}
```

---

## TASK 9: 创建路由配置

**文件：** `server/internal/router/router.go`

```go
package router

import (
    "easymeme/internal/handler"

    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
)

func Setup(
    tokenHandler *handler.TokenHandler,
    tradeHandler *handler.TradeHandler,
    wsHub *handler.WebSocketHub,
) *gin.Engine {
    r := gin.Default()

    // CORS 配置
    r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"*"},
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
        AllowCredentials: true,
    }))

    // Health check
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    // API routes
    api := r.Group("/api")
    {
        // Tokens
        api.GET("/tokens", tokenHandler.GetTokens)
        api.GET("/tokens/:address", tokenHandler.GetToken)
        api.POST("/tokens/:address/analyze", tokenHandler.AnalyzeToken)

        // Trades
        api.POST("/trades", tradeHandler.CreateTrade)
        api.GET("/trades", tradeHandler.GetTrades)
        api.PATCH("/trades/:txHash", tradeHandler.UpdateTradeStatus)
    }

    // WebSocket
    r.GET("/ws", wsHub.HandleWebSocket)

    return r
}
```

---

## TASK 10: 创建主入口文件

**文件：** `server/cmd/server/main.go`

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"

    "easymeme/internal/config"
    "easymeme/internal/handler"
    "easymeme/internal/repository"
    "easymeme/internal/router"
    "easymeme/internal/service"
    "easymeme/pkg/ethereum"
)

func main() {
    // 加载配置
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // 初始化数据库
    repo, err := repository.New(cfg.DatabaseURL)
    if err != nil {
        log.Fatalf("Failed to connect database: %v", err)
    }
    log.Println("Database connected")

    // 初始化以太坊客户端
    ethClient, err := ethereum.NewClient(cfg.BscRpcHTTP, cfg.BscRpcWS)
    if err != nil {
        log.Fatalf("Failed to connect BSC: %v", err)
    }
    defer ethClient.Close()
    log.Println("BSC RPC connected")

    // 初始化 WebSocket Hub
    wsHub := handler.NewWebSocketHub()
    go wsHub.Run()

    // 初始化服务
    analyzer := service.NewAnalyzer(ethClient)
    scanner := service.NewScanner(ethClient, repo, analyzer, wsHub)

    // 启动扫描服务
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    if err := scanner.Start(ctx); err != nil {
        log.Fatalf("Failed to start scanner: %v", err)
    }

    // 初始化 Handlers
    tokenHandler := handler.NewTokenHandler(repo, analyzer)
    tradeHandler := handler.NewTradeHandler(repo)

    // 设置路由
    r := router.Setup(tokenHandler, tradeHandler, wsHub)

    // 启动 HTTP 服务
    go func() {
        log.Printf("Server starting on port %s", cfg.Port)
        if err := r.Run(":" + cfg.Port); err != nil {
            log.Fatalf("Server failed: %v", err)
        }
    }()

    // 优雅退出
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down...")
    cancel()
}
```

---

## TASK 11: 创建 Docker 和部署配置

**文件：** `server/Dockerfile`

```dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app

# 安装依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源码并构建
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/server ./cmd/server

# 运行镜像
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=UTC

COPY --from=builder /bin/server /bin/server
COPY --from=builder /app/migrations /migrations

EXPOSE 8080

CMD ["/bin/server"]
```

**文件：** `docker-compose.yml`

```yaml
version: '3.8'

services:
  server:
    build:
      context: ./server
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - DATABASE_URL=postgres://postgres:postgres@db:5432/easymeme?sslmode=disable
      - REDIS_URL=redis://redis:6379
      - BSC_RPC_HTTP=${BSC_RPC_HTTP:-https://bsc-dataseed.binance.org}
      - BSC_RPC_WS=${BSC_RPC_WS:-wss://bsc-ws-node.nariox.org}
      - BSCSCAN_API_KEY=${BSCSCAN_API_KEY}
    depends_on:
      - db
      - redis
    restart: unless-stopped

  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: easymeme
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./server/migrations:/docker-entrypoint-initdb.d
    ports:
      - "5432:5432"

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

**文件：** `Makefile`

```makefile
.PHONY: dev build test docker-up docker-down

# 开发模式
dev:
	cd server && go run ./cmd/server

# 构建
build:
	cd server && go build -o bin/server ./cmd/server

# 测试
test:
	cd server && go test -v ./...

# Docker 启动
docker-up:
	docker-compose up -d

# Docker 停止
docker-down:
	docker-compose down

# 查看日志
logs:
	docker-compose logs -f server
```

---

## TASK 12: 创建前端项目

### 12.1 package.json

**文件：** `web/package.json`

```json
{
  "name": "easymeme-web",
  "version": "0.1.0",
  "private": true,
  "scripts": {
    "dev": "next dev",
    "build": "next build",
    "start": "next start",
    "lint": "next lint"
  },
  "dependencies": {
    "next": "14.1.0",
    "react": "18.2.0",
    "react-dom": "18.2.0",
    "@rainbow-me/rainbowkit": "^2.0.0",
    "@tanstack/react-query": "^5.17.0",
    "wagmi": "^2.5.0",
    "viem": "^2.7.0",
    "class-variance-authority": "^0.7.0",
    "clsx": "^2.1.0",
    "tailwind-merge": "^2.2.0",
    "lucide-react": "^0.309.0"
  },
  "devDependencies": {
    "typescript": "^5.3.0",
    "@types/node": "^20.11.0",
    "@types/react": "^18.2.0",
    "@types/react-dom": "^18.2.0",
    "tailwindcss": "^3.4.0",
    "postcss": "^8.4.0",
    "autoprefixer": "^10.4.0"
  }
}
```

### 12.2 Wagmi 配置

**文件：** `web/lib/wagmi.ts`

```typescript
import { getDefaultConfig } from '@rainbow-me/rainbowkit';
import { bsc } from 'wagmi/chains';

export const config = getDefaultConfig({
  appName: 'EasyMeme',
  projectId: process.env.NEXT_PUBLIC_WALLET_CONNECT_ID || '',
  chains: [bsc],
  ssr: true,
});
```

### 12.3 API 客户端

**文件：** `web/lib/api.ts`

```typescript
const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
const WS_URL = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/ws';

export interface Token {
  id: string;
  address: string;
  name: string;
  symbol: string;
  decimals: number;
  pair_address: string;
  initial_liquidity: string;
  risk_score: number;
  risk_level: 'safe' | 'warning' | 'danger';
  is_honeypot: boolean;
  buy_tax: number;
  sell_tax: number;
  created_at: string;
}

export async function getTokens(): Promise<Token[]> {
  const res = await fetch(`${API_URL}/api/tokens`);
  const data = await res.json();
  return data.data;
}

export async function getToken(address: string): Promise<Token> {
  const res = await fetch(`${API_URL}/api/tokens/${address}`);
  const data = await res.json();
  return data.data;
}

export function createWebSocket(onMessage: (data: any) => void): WebSocket {
  const ws = new WebSocket(WS_URL);

  ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    onMessage(data);
  };

  ws.onerror = (error) => {
    console.error('WebSocket error:', error);
  };

  return ws;
}
```

### 12.4 Providers

**文件：** `web/components/providers.tsx`

```typescript
'use client';

import { RainbowKitProvider } from '@rainbow-me/rainbowkit';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { WagmiProvider } from 'wagmi';
import { config } from '@/lib/wagmi';

import '@rainbow-me/rainbowkit/styles.css';

const queryClient = new QueryClient();

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <WagmiProvider config={config}>
      <QueryClientProvider client={queryClient}>
        <RainbowKitProvider>
          {children}
        </RainbowKitProvider>
      </QueryClientProvider>
    </WagmiProvider>
  );
}
```

### 12.5 Token List 组件

**文件：** `web/components/token-list.tsx`

```typescript
'use client';

import { useEffect, useState } from 'react';
import { Token, getTokens, createWebSocket } from '@/lib/api';
import { TokenCard } from './token-card';

export function TokenList() {
  const [tokens, setTokens] = useState<Token[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // 加载初始数据
    getTokens().then((data) => {
      setTokens(data);
      setLoading(false);
    });

    // WebSocket 实时更新
    const ws = createWebSocket((data) => {
      if (data.type === 'new_token') {
        setTokens((prev) => [data.token, ...prev].slice(0, 50));
      }
    });

    return () => ws.close();
  }, []);

  if (loading) {
    return <div className="text-center py-8">Loading...</div>;
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-bold">New Tokens</h2>
        <span className="text-sm text-green-500 animate-pulse">● Live</span>
      </div>
      <div className="grid gap-4">
        {tokens.map((token) => (
          <TokenCard key={token.id} token={token} />
        ))}
      </div>
    </div>
  );
}
```

### 12.6 Token Card 组件

**文件：** `web/components/token-card.tsx`

```typescript
'use client';

import { Token } from '@/lib/api';
import { RiskBadge } from './risk-badge';
import { TradePanel } from './trade-panel';
import { useState } from 'react';

export function TokenCard({ token }: { token: Token }) {
  const [showTrade, setShowTrade] = useState(false);

  return (
    <div className="border rounded-lg p-4 bg-card">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <RiskBadge level={token.risk_level} score={token.risk_score} />
          <div>
            <h3 className="font-bold">${token.symbol || 'Unknown'}</h3>
            <p className="text-sm text-muted-foreground truncate w-40">
              {token.address}
            </p>
          </div>
        </div>

        <div className="text-right">
          <p className="text-sm">LP: {parseFloat(token.initial_liquidity).toFixed(2)} BNB</p>
          <p className="text-sm text-muted-foreground">
            Tax: {token.buy_tax}% / {token.sell_tax}%
          </p>
        </div>

        <div className="flex gap-2">
          <a
            href={`https://bscscan.com/token/${token.address}`}
            target="_blank"
            rel="noopener noreferrer"
            className="px-3 py-1 border rounded text-sm hover:bg-accent"
          >
            View
          </a>
          {!token.is_honeypot && token.risk_level !== 'danger' && (
            <button
              onClick={() => setShowTrade(!showTrade)}
              className="px-3 py-1 bg-primary text-primary-foreground rounded text-sm"
            >
              Buy
            </button>
          )}
        </div>
      </div>

      {showTrade && (
        <div className="mt-4 pt-4 border-t">
          <TradePanel token={token} />
        </div>
      )}
    </div>
  );
}
```

### 12.7 Risk Badge 组件

**文件：** `web/components/risk-badge.tsx`

```typescript
export function RiskBadge({ level, score }: { level: string; score: number }) {
  const colors = {
    safe: 'bg-green-500',
    warning: 'bg-yellow-500',
    danger: 'bg-red-500',
  };

  const emoji = {
    safe: '🟢',
    warning: '🟡',
    danger: '🔴',
  };

  return (
    <div className="flex items-center gap-2">
      <span>{emoji[level as keyof typeof emoji] || '⚪'}</span>
      <span
        className={`px-2 py-0.5 rounded text-xs text-white ${
          colors[level as keyof typeof colors] || 'bg-gray-500'
        }`}
      >
        {score}
      </span>
    </div>
  );
}
```

### 12.8 Trade Panel 组件

**文件：** `web/components/trade-panel.tsx`

```typescript
'use client';

import { useState } from 'react';
import { useAccount, useWriteContract, useWaitForTransactionReceipt } from 'wagmi';
import { parseEther } from 'viem';
import { Token } from '@/lib/api';

const PANCAKE_ROUTER = '0x10ED43C718714eb63d5aA57B78B54704E256024E';
const WBNB = '0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c';

const ROUTER_ABI = [
  {
    name: 'swapExactETHForTokensSupportingFeeOnTransferTokens',
    type: 'function',
    inputs: [
      { name: 'amountOutMin', type: 'uint256' },
      { name: 'path', type: 'address[]' },
      { name: 'to', type: 'address' },
      { name: 'deadline', type: 'uint256' },
    ],
    outputs: [],
    stateMutability: 'payable',
  },
] as const;

const AMOUNTS = [0.1, 0.5, 1, 5];

export function TradePanel({ token }: { token: Token }) {
  const [amount, setAmount] = useState(0.1);
  const { address, isConnected } = useAccount();

  const { writeContract, data: hash, isPending } = useWriteContract();

  const { isLoading: isConfirming, isSuccess } = useWaitForTransactionReceipt({
    hash,
  });

  const handleBuy = () => {
    if (!address) return;

    const deadline = BigInt(Math.floor(Date.now() / 1000) + 1200);

    writeContract({
      address: PANCAKE_ROUTER,
      abi: ROUTER_ABI,
      functionName: 'swapExactETHForTokensSupportingFeeOnTransferTokens',
      args: [
        0n,
        [WBNB, token.address as `0x${string}`],
        address,
        deadline,
      ],
      value: parseEther(amount.toString()),
    });
  };

  if (!isConnected) {
    return <p className="text-sm text-muted-foreground">Connect wallet to trade</p>;
  }

  return (
    <div className="space-y-4">
      <div>
        <p className="text-sm mb-2">Amount (BNB)</p>
        <div className="flex gap-2">
          {AMOUNTS.map((a) => (
            <button
              key={a}
              onClick={() => setAmount(a)}
              className={`px-4 py-2 rounded border ${
                amount === a ? 'bg-primary text-primary-foreground' : ''
              }`}
            >
              {a}
            </button>
          ))}
        </div>
      </div>

      <button
        onClick={handleBuy}
        disabled={isPending || isConfirming}
        className="w-full py-3 bg-green-500 text-white rounded font-bold disabled:opacity-50"
      >
        {isPending
          ? 'Confirming...'
          : isConfirming
          ? 'Processing...'
          : `Buy with ${amount} BNB`}
      </button>

      {isSuccess && (
        <p className="text-green-500 text-sm">
          Transaction successful!{' '}
          <a
            href={`https://bscscan.com/tx/${hash}`}
            target="_blank"
            rel="noopener noreferrer"
            className="underline"
          >
            View on BSCScan
          </a>
        </p>
      )}
    </div>
  );
}
```

### 12.9 Layout

**文件：** `web/app/layout.tsx`

```typescript
import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import { Providers } from '@/components/providers';
import './globals.css';

const inter = Inter({ subsets: ['latin'] });

export const metadata: Metadata = {
  title: 'EasyMeme - BNB Chain Meme Token Scanner',
  description: 'AI-powered meme coin discovery and trading tool',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className={inter.className}>
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
```

### 12.10 Dashboard Page

**文件：** `web/app/dashboard/page.tsx`

```typescript
import { ConnectButton } from '@rainbow-me/rainbowkit';
import { TokenList } from '@/components/token-list';

export default function DashboardPage() {
  return (
    <div className="min-h-screen bg-background">
      <header className="border-b">
        <div className="container mx-auto px-4 py-4 flex items-center justify-between">
          <h1 className="text-2xl font-bold">EasyMeme</h1>
          <ConnectButton />
        </div>
      </header>

      <main className="container mx-auto px-4 py-8">
        <TokenList />
      </main>
    </div>
  );
}
```

### 12.11 Home Page (Landing)

**文件：** `web/app/page.tsx`

```typescript
import Link from 'next/link';

export default function HomePage() {
  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-gradient-to-b from-background to-muted">
      <h1 className="text-5xl font-bold mb-4">EasyMeme</h1>
      <p className="text-xl text-muted-foreground mb-8">
        AI-powered BNB Chain Meme Token Scanner
      </p>
      <Link
        href="/dashboard"
        className="px-8 py-4 bg-primary text-primary-foreground rounded-lg text-lg font-bold"
      >
        Launch App
      </Link>
    </div>
  );
}
```

### 12.12 Tailwind 配置

**文件：** `web/tailwind.config.js`

```javascript
/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './app/**/*.{js,ts,jsx,tsx}',
    './components/**/*.{js,ts,jsx,tsx}',
  ],
  theme: {
    extend: {
      colors: {
        background: 'hsl(var(--background))',
        foreground: 'hsl(var(--foreground))',
        card: 'hsl(var(--card))',
        primary: {
          DEFAULT: 'hsl(var(--primary))',
          foreground: 'hsl(var(--primary-foreground))',
        },
        muted: {
          DEFAULT: 'hsl(var(--muted))',
          foreground: 'hsl(var(--muted-foreground))',
        },
        accent: 'hsl(var(--accent))',
      },
    },
  },
  plugins: [],
};
```

**文件：** `web/app/globals.css`

```css
@tailwind base;
@tailwind components;
@tailwind utilities;

:root {
  --background: 0 0% 100%;
  --foreground: 222.2 84% 4.9%;
  --card: 0 0% 98%;
  --primary: 222.2 47.4% 11.2%;
  --primary-foreground: 210 40% 98%;
  --muted: 210 40% 96%;
  --muted-foreground: 215.4 16.3% 46.9%;
  --accent: 210 40% 94%;
}

.dark {
  --background: 222.2 84% 4.9%;
  --foreground: 210 40% 98%;
  --card: 222.2 84% 6%;
  --primary: 210 40% 98%;
  --primary-foreground: 222.2 47.4% 11.2%;
  --muted: 217.2 32.6% 17.5%;
  --muted-foreground: 215 20.2% 65.1%;
  --accent: 217.2 32.6% 12%;
}
```

**文件：** `web/.env.local`

```bash
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8080/ws
NEXT_PUBLIC_WALLET_CONNECT_ID=your_wallet_connect_project_id
```

---

## 执行顺序

```
1. TASK 1: 初始化后端项目 (go.mod, config, .env)
2. TASK 2: 创建数据模型 (model/token.go, model/trade.go)
3. TASK 3: 创建数据库操作层 (repository/repository.go)
4. TASK 4: 创建以太坊客户端封装 (pkg/ethereum/client.go)
5. TASK 5: 创建扫描服务 (service/scanner.go)
6. TASK 6: 创建风险分析服务 (service/analyzer.go)
7. TASK 7: 创建 WebSocket Hub (handler/websocket.go)
8. TASK 8: 创建 HTTP Handlers (handler/token.go, handler/trade.go)
9. TASK 9: 创建路由配置 (router/router.go)
10. TASK 10: 创建主入口文件 (cmd/server/main.go)
11. TASK 11: 创建 Docker 配置 (Dockerfile, docker-compose.yml)
12. TASK 12: 创建前端项目 (所有 web/ 文件)
```

---

## 运行指令

```bash
# 1. 启动数据库和 Redis
docker-compose up -d db redis

# 2. 运行后端
cd server
cp .env.example .env
go mod tidy
go run ./cmd/server

# 3. 运行前端 (新终端)
cd web
npm install
npm run dev

# 4. 访问
# 前端: http://localhost:3000
# 后端: http://localhost:8080
# 健康检查: http://localhost:8080/health
```

---

## API 端点

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | 健康检查 |
| GET | `/api/tokens` | 获取最新代币列表 |
| GET | `/api/tokens/:address` | 获取代币详情 |
| POST | `/api/tokens/:address/analyze` | 触发风险分析 |
| POST | `/api/trades` | 创建交易记录 |
| GET | `/api/trades?user=0x...` | 获取用户交易历史 |
| PATCH | `/api/trades/:txHash` | 更新交易状态 |
| WS | `/ws` | WebSocket 实时推送 |
