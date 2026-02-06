# EasyMeme MVP 开发计划

> 版本：v1.1
> 更新日期：2026-02-06
> 目标周期：4-6 周
> 后端技术栈：Golang

---

## 一、MVP 目标定义

### 1.1 MVP 核心目标

在最短时间内验证 **"AI 驱动的 Meme 币发现与一键交易"** 这一核心假设。

### 1.2 成功标准

| 指标 | 目标值 |
|------|--------|
| 新池发现延迟 | < 500ms |
| 风险检测准确率 | > 85% |
| 交易成功率 | > 95% |
| 首批测试用户 | 100 人 |
| 用户留存（7日） | > 25% |

### 1.3 MVP 功能范围

**包含：**
- ✅ PancakeSwap V2/V3 新池实时扫描
- ✅ 基础合约风险检测（貔貅/权限/税率）
- ✅ 一键快捷买入（预设金额）
- ✅ 钱包集成（MetaMask）
- ✅ 基础 UI 界面

**不包含（Phase 2+）：**
- ❌ 自动跟单策略
- ❌ 智能止盈止损
- ❌ Telegram Bot
- ❌ 多 DEX 聚合
- ❌ 会员系统

---

## 二、技术架构（MVP 版）

### 2.1 简化架构图

```
┌─────────────────────────────────────────────────────────┐
│                    Frontend (Next.js)                    │
│         Landing Page  │  Dashboard  │  Trade Panel       │
└─────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────┐
│                    Backend (Golang)                      │
│    REST API  │  WebSocket Server  │  Scanner Service     │
└─────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
        ┌──────────┐   ┌──────────┐   ┌──────────┐
        │  Redis   │   │ PostgreSQL│   │ BSC RPC  │
        │ (缓存)   │   │  (数据)   │   │ (链上)   │
        └──────────┘   └──────────┘   └──────────┘
```

### 2.2 技术栈确认

| 层级 | 技术选型 | 版本 |
|------|----------|------|
| **前端框架** | Next.js | 14.x |
| **UI 组件** | Tailwind CSS + shadcn/ui | latest |
| **Web3 集成** | wagmi + viem | 2.x |
| **钱包连接** | RainbowKit | 2.x |
| **后端语言** | Golang | 1.22+ |
| **Web 框架** | Gin | 1.9.x |
| **ORM** | GORM | 1.25.x |
| **数据库** | PostgreSQL | 16.x |
| **缓存** | Redis (go-redis) | 9.x |
| **链上交互** | go-ethereum | 1.13.x |
| **WebSocket** | gorilla/websocket | 1.5.x |
| **部署** | Vercel (前端) + Docker/Fly.io (后端) | - |

### 2.3 项目目录结构

```
easymeme/
├── web/                          # Next.js 前端
│   ├── app/
│   │   ├── page.tsx              # Landing Page
│   │   ├── dashboard/            # 仪表盘
│   │   └── trade/                # 交易页
│   ├── components/
│   │   ├── ui/                   # 基础 UI 组件
│   │   ├── token-card/           # 代币卡片
│   │   ├── risk-badge/           # 风险标签
│   │   └── trade-panel/          # 交易面板
│   ├── lib/
│   │   ├── wagmi.ts              # Web3 配置
│   │   └── api.ts                # API 客户端
│   ├── package.json
│   └── next.config.js
│
├── server/                       # Golang 后端
│   ├── cmd/
│   │   └── server/
│   │       └── main.go           # 入口文件
│   ├── internal/
│   │   ├── api/
│   │   │   ├── router.go         # 路由配置
│   │   │   └── handlers/         # HTTP 处理器
│   │   │       ├── token.go
│   │   │       └── trade.go
│   │   ├── service/
│   │   │   ├── scanner.go        # 扫描服务
│   │   │   ├── analyzer.go       # 合约分析
│   │   │   └── trader.go         # 交易执行
│   │   ├── model/
│   │   │   ├── token.go          # Token 模型
│   │   │   └── trade.go          # Trade 模型
│   │   ├── repository/
│   │   │   └── postgres.go       # 数据库操作
│   │   └── websocket/
│   │       └── hub.go            # WebSocket 管理
│   ├── pkg/
│   │   ├── config/               # 配置管理
│   │   ├── ethereum/             # 链上交互封装
│   │   └── utils/                # 工具函数
│   ├── migrations/               # 数据库迁移
│   ├── go.mod
│   ├── go.sum
│   └── Dockerfile
│
├── contracts/                    # 合约 ABI
│   ├── pancake_factory.json
│   └── pancake_router.json
│
├── docs/
│   └── MVP_DEVELOPMENT_PLAN.md
│
├── README.md
├── docker-compose.yml
└── Makefile
```

---

## 三、Sprint 规划

### 3.1 总体时间线

```
Week 1-2: Sprint 1 - 基础设施 + 扫描服务
Week 3:   Sprint 2 - 风险检测引擎
Week 4:   Sprint 3 - 交易功能 + 前端
Week 5:   Sprint 4 - 集成测试 + 优化
Week 6:   Buffer + 发布准备
```

---

### Sprint 1: 基础设施 + 扫描服务（Week 1-2）

#### 目标
搭建项目基础架构，实现新池实时扫描功能。

#### 任务分解

| ID | 任务 | 负责人 | 预估工时 | 优先级 |
|----|------|--------|----------|--------|
| S1-01 | 项目初始化 (Go mod + 目录结构) | 全栈 | 4h | P0 |
| S1-02 | 数据库设计 + GORM 模型 | 后端 | 4h | P0 |
| S1-03 | BSC RPC 连接 + WebSocket 监听 | 后端 | 8h | P0 |
| S1-04 | PancakeSwap Factory 事件解析 | 后端 | 8h | P0 |
| S1-05 | 新池数据入库 + Redis 缓存 | 后端 | 6h | P0 |
| S1-06 | Gorilla WebSocket 推送服务 | 后端 | 6h | P1 |
| S1-07 | 前端项目搭建 + 基础布局 | 前端 | 8h | P1 |
| S1-08 | 钱包连接组件 | 前端 | 4h | P1 |

#### 数据库模型（GORM）

```go
// server/internal/model/token.go
package model

import (
    "time"
    "github.com/shopspring/decimal"
    "gorm.io/datatypes"
)

type Token struct {
    ID               string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    Address          string          `gorm:"uniqueIndex;not null"`
    Name             *string
    Symbol           *string
    Decimals         int             `gorm:"default:18"`

    // 池信息
    PairAddress      *string
    Dex              string          `gorm:"default:pancakeswap"`
    InitialLiquidity decimal.Decimal `gorm:"type:decimal(36,18)"`

    // 风险信息
    RiskScore        *int
    RiskDetails      datatypes.JSON
    IsHoneypot       bool            `gorm:"default:false"`

    // 元数据
    CreatorAddress   *string
    CreatedAt        time.Time       `gorm:"autoCreateTime;index"`
    UpdatedAt        time.Time       `gorm:"autoUpdateTime"`
}

type Trade struct {
    ID           string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    UserAddress  string          `gorm:"index;not null"`
    TokenAddress string          `gorm:"index;not null"`
    Type         string          // buy | sell
    AmountIn     decimal.Decimal `gorm:"type:decimal(36,18)"`
    AmountOut    decimal.Decimal `gorm:"type:decimal(36,18)"`
    TxHash       string          `gorm:"uniqueIndex;not null"`
    Status       string          // pending | success | failed
    GasUsed      decimal.Decimal `gorm:"type:decimal(36,18)"`
    CreatedAt    time.Time       `gorm:"autoCreateTime"`
}

func (Token) TableName() string { return "tokens" }
func (Trade) TableName() string { return "trades" }
```

#### 关键代码：新池监听

```go
// server/internal/service/scanner.go
package service

import (
    "context"
    "log"
    "math/big"
    "strings"

    "github.com/ethereum/go-ethereum"
    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/ethclient"
)

const (
    PancakeFactoryV2 = "0xcA143Ce32Fe78f1f7019d7d551a6402fC5350c73"
    WBNB             = "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c"
)

var pairCreatedEventSig = common.HexToHash(
    "0x0d3648bd0f6ba80134a33ba9275ac585d9d315f0ad8355cddefde31afa28d0e9",
)

type PoolScanner struct {
    client    *ethclient.Client
    wsClient  *ethclient.Client
    analyzer  *Analyzer
    hub       *WebSocketHub
}

func NewPoolScanner(rpcURL, wsURL string, analyzer *Analyzer, hub *WebSocketHub) (*PoolScanner, error) {
    client, err := ethclient.Dial(rpcURL)
    if err != nil {
        return nil, err
    }

    wsClient, err := ethclient.Dial(wsURL)
    if err != nil {
        return nil, err
    }

    return &PoolScanner{
        client:   client,
        wsClient: wsClient,
        analyzer: analyzer,
        hub:      hub,
    }, nil
}

func (s *PoolScanner) Start(ctx context.Context) error {
    factoryAddr := common.HexToAddress(PancakeFactoryV2)

    query := ethereum.FilterQuery{
        Addresses: []common.Address{factoryAddr},
        Topics:    [][]common.Hash{{pairCreatedEventSig}},
    }

    logs := make(chan types.Log)
    sub, err := s.wsClient.SubscribeFilterLogs(ctx, query, logs)
    if err != nil {
        return err
    }

    log.Println("Pool scanner started, listening for PairCreated events...")

    go func() {
        for {
            select {
            case err := <-sub.Err():
                log.Printf("Subscription error: %v", err)
                return
            case vLog := <-logs:
                s.handlePairCreated(vLog)
            case <-ctx.Done():
                return
            }
        }
    }()

    return nil
}

func (s *PoolScanner) handlePairCreated(vLog types.Log) {
    // 解析事件参数
    token0 := common.HexToAddress(vLog.Topics[1].Hex())
    token1 := common.HexToAddress(vLog.Topics[2].Hex())
    pairAddr := common.BytesToAddress(vLog.Data[:32])

    // 找出非 WBNB 的代币
    wbnb := common.HexToAddress(WBNB)
    var targetToken common.Address
    if token0 == wbnb {
        targetToken = token1
    } else if token1 == wbnb {
        targetToken = token0
    } else {
        return // 非 WBNB 配对，跳过
    }

    log.Printf("New pair detected: %s, Token: %s", pairAddr.Hex(), targetToken.Hex())

    // 异步分析代币
    go s.analyzeAndBroadcast(targetToken, pairAddr)
}

func (s *PoolScanner) analyzeAndBroadcast(tokenAddr, pairAddr common.Address) {
    result, err := s.analyzer.Analyze(tokenAddr, pairAddr)
    if err != nil {
        log.Printf("Analyze error: %v", err)
        return
    }

    // 通过 WebSocket 推送给前端
    s.hub.Broadcast(result)
}
```

#### 交付物
- [x] Monorepo 项目结构
- [x] 数据库连接 + 迁移脚本
- [x] 新池扫描服务运行
- [x] 前端基础框架 + 钱包连接

#### WebSocket 推送服务

```go
// server/internal/websocket/hub.go
package websocket

import (
    "encoding/json"
    "log"
    "sync"

    "github.com/gorilla/websocket"
)

type Client struct {
    hub  *Hub
    conn *websocket.Conn
    send chan []byte
}

type Hub struct {
    clients    map[*Client]bool
    broadcast  chan []byte
    register   chan *Client
    unregister chan *Client
    mu         sync.RWMutex
}

func NewHub() *Hub {
    return &Hub{
        clients:    make(map[*Client]bool),
        broadcast:  make(chan []byte, 256),
        register:   make(chan *Client),
        unregister: make(chan *Client),
    }
}

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.mu.Lock()
            h.clients[client] = true
            h.mu.Unlock()
            log.Printf("Client connected, total: %d", len(h.clients))

        case client := <-h.unregister:
            h.mu.Lock()
            if _, ok := h.clients[client]; ok {
                delete(h.clients, client)
                close(client.send)
            }
            h.mu.Unlock()
            log.Printf("Client disconnected, total: %d", len(h.clients))

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

func (h *Hub) Broadcast(data interface{}) {
    message, err := json.Marshal(data)
    if err != nil {
        log.Printf("Marshal error: %v", err)
        return
    }
    h.broadcast <- message
}

func (c *Client) WritePump() {
    defer func() {
        c.conn.Close()
    }()

    for message := range c.send {
        if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
            return
        }
    }
}
```

---

### Sprint 2: 风险检测引擎（Week 3）

#### 目标
实现基础的合约风险检测，包括貔貅检测、权限检查、税率分析。

#### 任务分解

| ID | 任务 | 负责人 | 预估工时 | 优先级 |
|----|------|--------|----------|--------|
| S2-01 | 合约源码获取 (BSCScan API) | 后端 | 4h | P0 |
| S2-02 | 貔貅检测（模拟卖出） | 后端 | 8h | P0 |
| S2-03 | Owner 权限分析 | 后端 | 6h | P0 |
| S2-04 | 买卖税率检测 | 后端 | 6h | P0 |
| S2-05 | 风险评分算法 | 后端 | 4h | P1 |
| S2-06 | 持币分布分析 | 后端 | 4h | P1 |
| S2-07 | 风险报告 API | 后端 | 4h | P1 |

#### 风险评分算法

```go
// server/internal/service/analyzer.go
package service

import (
    "github.com/ethereum/go-ethereum/common"
)

type RiskFactors struct {
    IsHoneypot         bool    `json:"is_honeypot"`
    CanMint            bool    `json:"can_mint"`
    CanPause           bool    `json:"can_pause"`
    CanBlacklist       bool    `json:"can_blacklist"`
    BuyTax             float64 `json:"buy_tax"`
    SellTax            float64 `json:"sell_tax"`
    OwnerCanChangeTax  bool    `json:"owner_can_change_tax"`
    Top10HoldingPercent float64 `json:"top10_holding_percent"`
    LPLocked           bool    `json:"lp_locked"`
    ContractVerified   bool    `json:"contract_verified"`
}

type RiskLevel string

const (
    RiskSafe    RiskLevel = "safe"
    RiskWarning RiskLevel = "warning"
    RiskDanger  RiskLevel = "danger"
)

func CalculateRiskScore(factors RiskFactors) int {
    score := 100

    // 貔貅直接归零
    if factors.IsHoneypot {
        return 0
    }

    if factors.CanMint {
        score -= 30
    }
    if factors.CanPause {
        score -= 20
    }
    if factors.CanBlacklist {
        score -= 25
    }
    if factors.BuyTax > 10 {
        score -= 15
    }
    if factors.SellTax > 10 {
        score -= 15
    }
    if factors.OwnerCanChangeTax {
        score -= 20
    }
    if factors.Top10HoldingPercent > 50 {
        score -= 15
    }
    if !factors.LPLocked {
        score -= 20
    }
    if !factors.ContractVerified {
        score -= 10
    }

    if score < 0 {
        return 0
    }
    return score
}

func GetRiskLevel(score int) RiskLevel {
    if score >= 70 {
        return RiskSafe
    }
    if score >= 40 {
        return RiskWarning
    }
    return RiskDanger
}
```

#### 貔貅检测实现

```go
// server/internal/service/honeypot.go
package service

import (
    "context"
    "math/big"

    "github.com/ethereum/go-ethereum"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
)

const (
    PancakeRouterV2 = "0x10ED43C718714eb63d5aA57B78B54704E256024E"
)

type HoneypotChecker struct {
    client *ethclient.Client
}

func NewHoneypotChecker(client *ethclient.Client) *HoneypotChecker {
    return &HoneypotChecker{client: client}
}

func (h *HoneypotChecker) Check(ctx context.Context, tokenAddr common.Address) (bool, error) {
    router := common.HexToAddress(PancakeRouterV2)
    wbnb := common.HexToAddress(WBNB)

    // 构建模拟买入调用数据
    buyData, err := h.buildSwapCallData(
        "swapExactETHForTokens",
        big.NewInt(0),
        []common.Address{wbnb, tokenAddr},
    )
    if err != nil {
        return true, err
    }

    // 模拟买入
    buyAmount := big.NewInt(1e16) // 0.01 BNB
    _, err = h.client.CallContract(ctx, ethereum.CallMsg{
        To:    &router,
        Value: buyAmount,
        Data:  buyData,
    }, nil)
    if err != nil {
        // 买入失败，可能是貔貅
        return true, nil
    }

    // 构建模拟卖出调用数据
    sellData, err := h.buildSwapCallData(
        "swapExactTokensForETH",
        big.NewInt(1e18), // 假设买到 1 token
        []common.Address{tokenAddr, wbnb},
    )
    if err != nil {
        return true, err
    }

    // 模拟卖出
    _, err = h.client.CallContract(ctx, ethereum.CallMsg{
        To:   &router,
        Data: sellData,
    }, nil)
    if err != nil {
        // 卖出失败，确认是貔貅
        return true, nil
    }

    return false, nil
}

func (h *HoneypotChecker) buildSwapCallData(method string, amount *big.Int, path []common.Address) ([]byte, error) {
    // 实际实现需要使用 ABI 编码
    // 这里简化处理，完整实现需要加载 Router ABI
    return nil, nil
}
```

#### 交付物
- [x] 貔貅检测服务
- [x] Owner 权限分析
- [x] 税率检测
- [x] 风险评分 API
- [x] 风险报告 JSON 结构

---

### Sprint 3: 交易功能 + 前端（Week 4）

#### 目标
实现一键买入功能，完成核心 UI 界面。

#### 任务分解

| ID | 任务 | 负责人 | 预估工时 | 优先级 |
|----|------|--------|----------|--------|
| S3-01 | 交易构建服务 (Swap) | 后端 | 8h | P0 |
| S3-02 | Gas 估算 + 优化 | 后端 | 4h | P0 |
| S3-03 | 交易记录入库 | 后端 | 4h | P1 |
| S3-04 | Dashboard 页面 (新池列表) | 前端 | 8h | P0 |
| S3-05 | Token 详情页 + 风险展示 | 前端 | 6h | P0 |
| S3-06 | 交易面板组件 | 前端 | 8h | P0 |
| S3-07 | 实时价格展示 | 前端 | 4h | P1 |
| S3-08 | 交易历史页 | 前端 | 4h | P2 |

#### 核心页面设计

**Dashboard 页面**
```
┌─────────────────────────────────────────────────────────────┐
│  EasyMeme        [Dashboard] [History]    [0x1234...5678]   │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  🔥 新发现的代币                              [实时更新中]   │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ 🟢 $PEPE2     LP: 50 BNB    Score: 85    [查看] [买入] │   │
│  ├─────────────────────────────────────────────────────┤   │
│  │ 🟡 $DOGE3     LP: 20 BNB    Score: 62    [查看] [买入] │   │
│  ├─────────────────────────────────────────────────────┤   │
│  │ 🔴 $SHIB4     LP: 5 BNB     Score: 25    [查看] [---] │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**交易面板组件**
```
┌─────────────────────────────┐
│      快捷买入 $PEPE2        │
├─────────────────────────────┤
│                             │
│  金额:  [0.1] [0.5] [1] [5] │
│                             │
│  滑点:  [自动优化: 12%]     │
│                             │
│  Gas:   [Normal ▼]          │
│                             │
│  预估获得: ~1,234,567 PEPE2 │
│                             │
│  ┌───────────────────────┐  │
│  │      确认买入          │  │
│  └───────────────────────┘  │
│                             │
└─────────────────────────────┘
```

#### 前端核心组件

```tsx
// apps/web/components/trade-panel/index.tsx
'use client';

import { useState } from 'react';
import { useAccount, useWriteContract } from 'wagmi';
import { parseEther } from 'viem';
import { Button } from '@/components/ui/button';

const AMOUNTS = [0.1, 0.5, 1, 5];

export function TradePanel({ token }: { token: TokenInfo }) {
  const [amount, setAmount] = useState(0.1);
  const { address } = useAccount();
  const { writeContract, isPending } = useWriteContract();

  const handleBuy = async () => {
    writeContract({
      address: PANCAKE_ROUTER,
      abi: ROUTER_ABI,
      functionName: 'swapExactETHForTokensSupportingFeeOnTransferTokens',
      args: [
        0n, // amountOutMin
        [WBNB, token.address],
        address,
        BigInt(Math.floor(Date.now() / 1000) + 1200),
      ],
      value: parseEther(amount.toString()),
    });
  };

  return (
    <div className="p-4 border rounded-lg">
      <h3 className="text-lg font-bold mb-4">快捷买入 ${token.symbol}</h3>

      <div className="flex gap-2 mb-4">
        {AMOUNTS.map((a) => (
          <Button
            key={a}
            variant={amount === a ? 'default' : 'outline'}
            onClick={() => setAmount(a)}
          >
            {a} BNB
          </Button>
        ))}
      </div>

      <Button
        className="w-full"
        onClick={handleBuy}
        disabled={isPending || !address}
      >
        {isPending ? '交易中...' : '确认买入'}
      </Button>
    </div>
  );
}
```

#### 交付物
- [x] 交易构建 + 执行服务
- [x] Dashboard 页面
- [x] Token 详情 + 风险展示
- [x] 交易面板组件
- [x] 交易记录页面

---

### Sprint 4: 集成测试 + 优化（Week 5）

#### 目标
端到端测试、性能优化、Bug 修复。

#### 任务分解

| ID | 任务 | 负责人 | 预估工时 | 优先级 |
|----|------|--------|----------|--------|
| S4-01 | E2E 测试用例编写 | 全栈 | 8h | P0 |
| S4-02 | 扫描延迟优化 (< 500ms) | 后端 | 6h | P0 |
| S4-03 | 前端性能优化 | 前端 | 4h | P1 |
| S4-04 | 错误处理 + 用户提示 | 全栈 | 4h | P0 |
| S4-05 | 移动端响应式适配 | 前端 | 4h | P1 |
| S4-06 | 安全审计（XSS/注入） | 全栈 | 4h | P0 |
| S4-07 | 监控 + 日志配置 | 后端 | 4h | P1 |
| S4-08 | 文档完善 | 全栈 | 4h | P2 |

#### 测试清单

```markdown
## E2E 测试用例

### 扫描服务
- [ ] 新池创建后 500ms 内收到推送
- [ ] 正确解析 token0/token1
- [ ] 数据正确入库

### 风险检测
- [ ] 已知貔貅合约检测为高危
- [ ] 正常合约检测为安全
- [ ] 高税率正确识别

### 交易功能
- [ ] 钱包连接成功
- [ ] 交易签名正确
- [ ] 交易上链成功
- [ ] 交易记录保存

### 前端
- [ ] 新池实时更新
- [ ] 风险标签正确显示
- [ ] 交易面板功能正常
- [ ] 移动端布局正常
```

#### 交付物
- [x] 测试覆盖率 > 80%
- [x] 性能达标（延迟 < 500ms）
- [x] 无 P0 级 Bug
- [x] 部署文档

---

## 四、部署计划

### 4.1 基础设施

| 组件 | 服务商 | 配置 | 预估成本 |
|------|--------|------|----------|
| 前端 | Vercel | Pro Plan | $20/月 |
| 后端 | Fly.io | 2x shared-cpu-1x | $15/月 |
| 数据库 | Fly.io (PostgreSQL) | 1GB | $7/月 |
| 缓存 | Upstash (Redis) | Free Tier | $0/月 |
| RPC | QuickNode / Ankr | BSC 专用 | $50/月 |
| 域名 | Cloudflare | easymeme.xyz | $15/年 |

**MVP 阶段预估：~$95/月**

### 4.2 部署流程

```bash
# 1. 后端构建与部署
cd server

# 本地构建测试
go build -o bin/server ./cmd/server

# Docker 构建
docker build -t easymeme-server .

# 部署到 Fly.io
fly launch
fly deploy

# 2. 数据库迁移
fly ssh console -C "./bin/server migrate"

# 3. 前端部署
cd ../web
npm install
npm run build
vercel --prod

# 4. 环境变量配置 (Fly.io)
fly secrets set DATABASE_URL="postgres://..." \
    REDIS_URL="redis://..." \
    BSC_RPC_HTTP="https://..." \
    BSC_RPC_WS="wss://..."
```

#### Dockerfile (后端)

```dockerfile
# server/Dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/server ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /bin/server /bin/server
COPY migrations /migrations

EXPOSE 8080
CMD ["/bin/server"]
```

### 4.3 监控告警

| 指标 | 告警阈值 | 通知方式 |
|------|----------|----------|
| API 响应时间 | > 2s | Telegram |
| 扫描服务中断 | > 1min | Telegram + 邮件 |
| 错误率 | > 5% | Telegram |
| 数据库连接数 | > 80% | 邮件 |

---

## 五、风险 & 缓解

| 风险 | 可能性 | 影响 | 缓解措施 |
|------|--------|------|----------|
| RPC 节点不稳定 | 高 | 高 | 多节点冗余、自动切换 |
| BSCScan API 限流 | 中 | 中 | 缓存合约源码、申请更高限额 |
| 合约分析误报 | 中 | 高 | 人工复核机制、持续优化算法 |
| 交易失败率高 | 低 | 高 | 动态 Gas、滑点自适应 |

---

## 六、后续迭代方向

MVP 发布后，根据用户反馈优先级排序：

1. **Telegram Bot** - 用户强需求
2. **自动跟单** - 核心差异化功能
3. **止盈止损** - 提升用户收益
4. **Four.meme 集成** - 扩展覆盖范围
5. **会员系统** - 商业化变现

---

## 七、团队分工建议

| 角色 | 职责 | 技能要求 |
|------|------|----------|
| **后端工程师 x1** | Go 后端服务 + 链上交互 | Golang, go-ethereum, PostgreSQL, WebSocket |
| **前端工程师 x1** | Web UI + 钱包集成 | React/Next.js, TypeScript, wagmi/viem |
| **产品/设计 x0.5** | 产品规划 + UI 设计 | Figma, Web3 产品经验 |

---

## 附录

### A. 关键 API 列表

| 端点 | 方法 | 描述 |
|------|------|------|
| `/api/tokens` | GET | 获取最新代币列表 |
| `/api/tokens/:address` | GET | 获取代币详情 + 风险报告 |
| `/api/tokens/:address/analyze` | POST | 触发风险分析 |
| `/api/trades` | POST | 提交交易（预构建） |
| `/api/trades/:txHash` | GET | 查询交易状态 |
| `/ws/tokens` | WS | 新池实时推送 |

### B. 环境变量

```bash
# server/.env.example

# Server
PORT=8080
GIN_MODE=release

# Database
DATABASE_URL="postgres://user:pass@localhost:5432/easymeme?sslmode=disable"

# Redis
REDIS_URL="redis://localhost:6379"

# BSC RPC
BSC_RPC_HTTP="https://bsc-dataseed.binance.org"
BSC_RPC_WS="wss://bsc-ws-node.nariox.org"

# BSCScan
BSCSCAN_API_KEY="your-api-key"

# JWT Secret (for future auth)
JWT_SECRET="your-jwt-secret"
```

```bash
# web/.env.local

NEXT_PUBLIC_API_URL="http://localhost:8080"
NEXT_PUBLIC_WS_URL="ws://localhost:8080/ws"
NEXT_PUBLIC_WALLET_CONNECT_ID="your-wallet-connect-id"
```

### C. Makefile

```makefile
# Makefile

.PHONY: dev build test migrate

# 后端开发
dev-server:
	cd server && go run ./cmd/server

# 前端开发
dev-web:
	cd web && npm run dev

# 构建后端
build-server:
	cd server && go build -o bin/server ./cmd/server

# 运行测试
test:
	cd server && go test -v ./...

# 数据库迁移
migrate:
	cd server && go run ./cmd/server migrate

# Docker 构建
docker-build:
	docker build -t easymeme-server ./server

# 启动所有服务 (docker-compose)
up:
	docker-compose up -d

down:
	docker-compose down
```

### D. 参考资源

- [PancakeSwap 文档](https://docs.pancakeswap.finance/)
- [BSCScan API](https://docs.bscscan.com/)
- [go-ethereum 文档](https://geth.ethereum.org/docs)
- [Gin Web Framework](https://gin-gonic.com/docs/)
- [GORM 文档](https://gorm.io/docs/)
- [wagmi 文档](https://wagmi.sh/)
- [viem 文档](https://viem.sh/)

### E. docker-compose.yml

```yaml
# docker-compose.yml
version: '3.8'

services:
  server:
    build:
      context: ./server
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - GIN_MODE=release
      - DATABASE_URL=postgres://postgres:postgres@db:5432/easymeme?sslmode=disable
      - REDIS_URL=redis://redis:6379
      - BSC_RPC_HTTP=${BSC_RPC_HTTP}
      - BSC_RPC_WS=${BSC_RPC_WS}
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

### F. API Handler 示例

```go
// server/internal/api/handlers/token.go
package handlers

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "easymeme/internal/service"
    "easymeme/internal/repository"
)

type TokenHandler struct {
    repo     *repository.TokenRepository
    analyzer *service.Analyzer
}

func NewTokenHandler(repo *repository.TokenRepository, analyzer *service.Analyzer) *TokenHandler {
    return &TokenHandler{repo: repo, analyzer: analyzer}
}

// GetTokens 获取最新代币列表
func (h *TokenHandler) GetTokens(c *gin.Context) {
    limit := 50
    tokens, err := h.repo.GetLatest(c.Request.Context(), limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"data": tokens})
}

// GetToken 获取代币详情
func (h *TokenHandler) GetToken(c *gin.Context) {
    address := c.Param("address")

    token, err := h.repo.GetByAddress(c.Request.Context(), address)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "token not found"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"data": token})
}

// AnalyzeToken 触发风险分析
func (h *TokenHandler) AnalyzeToken(c *gin.Context) {
    address := c.Param("address")

    result, err := h.analyzer.AnalyzeByAddress(c.Request.Context(), address)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"data": result})
}
```

```go
// server/internal/api/router.go
package api

import (
    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
    "easymeme/internal/api/handlers"
    ws "easymeme/internal/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true },
}

func SetupRouter(tokenHandler *handlers.TokenHandler, hub *ws.Hub) *gin.Engine {
    r := gin.Default()

    // CORS
    r.Use(CORSMiddleware())

    // Health check
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    // API routes
    api := r.Group("/api")
    {
        api.GET("/tokens", tokenHandler.GetTokens)
        api.GET("/tokens/:address", tokenHandler.GetToken)
        api.POST("/tokens/:address/analyze", tokenHandler.AnalyzeToken)
    }

    // WebSocket
    r.GET("/ws", func(c *gin.Context) {
        handleWebSocket(c, hub)
    })

    return r
}

func handleWebSocket(c *gin.Context, hub *ws.Hub) {
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        return
    }

    client := &ws.Client{
        Hub:  hub,
        Conn: conn,
        Send: make(chan []byte, 256),
    }
    hub.Register <- client

    go client.WritePump()
}
```
