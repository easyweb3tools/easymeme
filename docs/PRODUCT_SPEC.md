# EasyMeme - BNB Chain 自治 Agent 产品规格书

> **本文档供 Codex 开发使用** | 最后更新: 2026-02-07

---

## 1. 产品定位

**EasyMeme 是什么：**
一个**长期运行**在 BNB Chain 上的**自治 Agent**，能够持续发现、判断、跟踪 Meme 币机会。

**为什么必须用 OpenClaw 构建：**
- Agent 自主决策：不是规则触发，而是 AI 判断
- Memory 积累经验：记住见过的代币和风险模式
- Cron 持续运行：自动唤醒，无需外部调度

---

## 2. 系统架构

```
┌─────────────────────────────────────────────────────────────────┐
│                      EasyMeme Architecture                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   ┌──────────────┐     ┌──────────────┐     ┌──────────────┐   │
│   │   Server     │     │  OpenClaw    │     │    Web       │   │
│   │   (Go)       │────▶│  Agent       │     │  (Next.js)   │   │
│   │              │◀────│              │     │              │   │
│   └──────────────┘     └──────────────┘     └──────────────┘   │
│         │                    │                     │            │
│         │                    │                     │            │
│   ┌─────▼─────┐        ┌─────▼─────┐        ┌─────▼─────┐      │
│   │ 链上数据   │        │ AI 分析   │        │ 用户交易  │      │
│   │ 抓取存储   │        │ 金狗识别  │        │ 钱包签名  │      │
│   └───────────┘        └───────────┘        └───────────┘      │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 2.1 职责划分

| 组件 | 职责 | 技术栈 |
|------|------|--------|
| **Server** | 链上数据抓取、数据库存储、提供 API | Go + PostgreSQL |
| **OpenClaw** | AI 分析、风险判断、金狗识别 | TypeScript + OpenClaw SDK |
| **Web** | UI 展示、钱包连接、交易执行 | Next.js + wagmi |

### 2.2 数据流

```
1. Server 定时抓取 BSC 链上新代币数据 → 存入 DB
                    ↓
2. OpenClaw Agent 定时从 Server API 获取待分析代币
                    ↓
3. OpenClaw 用 AI 分析风险，判断是否"金狗"
                    ↓
4. OpenClaw 将分析结果回写到 Server API
                    ↓
5. Web 从 Server 获取分析结果展示给用户
                    ↓
6. 用户在 Web 端用钱包签名执行交易
                    ↓
7. Server 记录交易历史
```

---

## 3. Server 规格 (Go)

### 3.1 职责
- **数据抓取**: 监听 PancakeSwap PairCreated 事件
- **数据存储**: PostgreSQL 存储代币信息、分析结果、交易历史
- **API 服务**: 提供 REST API 供 OpenClaw 和 Web 调用

### 3.2 API 端点

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/tokens/pending` | 获取待分析的代币列表 |
| GET | `/api/tokens/analyzed` | 获取已分析代币列表 (供 Web) |
| GET | `/api/tokens/:address` | 获取单个代币详情 |
| POST | `/api/tokens/:address/analysis` | 回写分析结果 |
| POST | `/api/trades` | 记录交易历史 |

### 3.3 数据模型

**Token (代币)**
```go
type Token struct {
    Address          string    // 合约地址
    Name             string    
    Symbol           string    
    PairAddress      string    // 交易对地址
    Liquidity        float64   // 流动性 (BNB)
    CreatorAddress   string    
    CreatedAt        time.Time
    
    // 分析结果 (由 OpenClaw 回写)
    AnalysisStatus   string    // pending | analyzed
    RiskScore        int       // 0-100
    RiskLevel        string    // SAFE | WARNING | DANGER
    IsGoldenDog      bool      // 是否金狗
    AnalysisResult   JSON      // 详细分析报告
    AnalyzedAt       time.Time
}
```

### 3.4 安全与校验

- **CORS**: 只能允许配置的来源，不能使用 `*` + `AllowCredentials`
- **API Key**: `POST /api/tokens/:address/analysis` 必须校验 `X-API-Key`
- **输入校验**: `riskScore` 必须 0-100，`riskLevel` 只能是 SAFE/WARNING/DANGER

---

## 4. OpenClaw Agent 规格

### 4.1 核心定位

> **OpenClaw 只做 AI 分析，不做传统数据抓取和存储**

OpenClaw 的优势是 AI 判断和自主决策，让它专注于：
- 分析代币是否"金狗"
- 识别风险模式
- 用 AI 做出判断

### 4.2 工作流程

```
┌─────────────────────────────────────────────────────┐
│                 OpenClaw Agent                       │
├─────────────────────────────────────────────────────┤
│                                                      │
│   1. Cron 触发 (每 5 分钟)                           │
│            ↓                                         │
│   2. 调用 Server API 获取待分析代币                  │
│            ↓                                         │
│   3. AI 分析每个代币 (使用 LLM)                      │
│      - 解读链上数据                                  │
│      - 评估风险因素                                  │
│      - 判断是否金狗                                  │
│            ↓                                         │
│   4. 将分析结果回写 Server API                       │
│            ↓                                         │
│   5. 更新 Memory (记住风险模式)                      │
│                                                      │
└─────────────────────────────────────────────────────┘
```

### 4.3 Tool 定义

**Tool: `fetchPendingTokens`**
```typescript
// 从 Server API 获取待分析代币
interface FetchPendingTokensInput {
  limit?: number;  // 默认 10
}

interface FetchPendingTokensOutput {
  tokens: PendingToken[];
}
```

**Tool: `analyzeTokenRisk`**
```typescript
// AI 分析代币风险 (核心能力)
interface AnalyzeTokenRiskInput {
  token: {
    address: string;
    name: string;
    symbol: string;
    liquidity: number;
    creatorAddress: string;
    // Server 提供的原始链上数据
    contractCode?: string;
    holderDistribution?: HolderInfo[];
    creatorHistory?: CreatorTx[];
  };
  analysis: AnalyzeTokenRiskOutput;
}

interface AnalyzeTokenRiskOutput {
  riskScore: number;        // 0-100
  riskLevel: 'SAFE' | 'WARNING' | 'DANGER';
  isGoldenDog: boolean;     // 核心判断：是否金狗
  
  riskFactors: {
    honeypotRisk: 'LOW' | 'MEDIUM' | 'HIGH';
    taxRisk: 'LOW' | 'MEDIUM' | 'HIGH';
    ownerRisk: 'LOW' | 'MEDIUM' | 'HIGH';
    concentrationRisk: 'LOW' | 'MEDIUM' | 'HIGH';
  };
  
  reasoning: string;        // AI 的判断理由
  recommendation: string;   // 给用户的建议
}
```

**Tool: `submitAnalysis`**
```typescript
// 将分析结果回写 Server
interface SubmitAnalysisInput {
  tokenAddress: string;
  analysis: AnalyzeTokenRiskOutput;
}
```

### 4.4 什么是"金狗"

AI 需要综合判断以下因素：

| 因素 | 金狗特征 | 权重 |
|------|---------|------|
| 安全性 | 非貔貅、税率合理、无危险权限 | 必要条件 |
| 流动性 | LP 充足且锁定 | 高 |
| 持仓分布 | 不集中在少数地址 | 中 |
| 创建者历史 | 无 rug 历史 | 高 |
| 社区热度 | 有社交媒体关注 | 加分项 |

**金狗 ≠ 安全**

金狗是指"值得关注、可能有机会"的代币，需要 AI 做综合判断，而不仅仅是安全检测。

### 4.5 Memory 使用

| Memory Key | 用途 |
|------------|------|
| `analyzedTokens` | 已分析代币记录 (去重) |
| `riskPatterns` | 识别到的风险模式 |
| `goldenDogHistory` | 历史金狗表现追踪 |

---

## 5. Web 规格 (Next.js)

### 5.1 职责
- UI 展示分析结果
- 钱包连接 (MetaMask, Trust Wallet)
- 用户签名执行交易
- 展示交易历史

### 5.2 核心页面

| 页面 | 功能 |
|------|------|
| Dashboard | 展示最新分析的代币列表 |
| Token Detail | 单个代币的详细分析报告 |
| Trade | 一键买入/止盈止损设置 |
| History | 交易历史记录 |

### 5.3 交易安全

- **滑点保护**: 交易必须计算 `amountOutMin`，禁止 0 滑点
- **合约地址配置**: `PANCAKE_ROUTER` 与 `WBNB` 从环境变量注入

### 5.4 与 Server 交互

- 只读取数据，不直接操作链
- 交易由用户钱包签名后发送
- 交易结果通过 Server 记录

---

## 6. 部署与运行

### 6.1 Docker Compose

```yaml
services:
  postgres:
    image: postgres:16
    
  server:
    build: ./server
    depends_on: [postgres]
    
  openclaw:
    build: ./openclaw-skill
    depends_on: [server]
    # 需提供 OpenClaw 配置文件与 API Key
    
  web:
    build: ./web
    depends_on: [server]
```

### 6.2 环境变量

**Server:**
```bash
DATABASE_URL=postgres://...
BSCSCAN_API_KEY=xxx
BSC_RPC_HTTP=https://bsc-dataseed.bnbchain.org
BSC_RPC_WS=wss://...
EASYMEME_API_KEY=xxx
CORS_ALLOWED_ORIGINS=http://localhost:3000
```

**OpenClaw:**
```bash
EASYMEME_SERVER_URL=http://server:8080
EASYMEME_API_KEY=xxx
# OpenClaw 自身的模型配置 (openclaw.json)
agents.defaults.model.primary = "google/gemini-3-flash"
GEMINI_API_KEY=xxx
```

**Web:**
```bash
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8080/ws
NEXT_PUBLIC_PANCAKE_ROUTER=0x10ED43C718714eb63d5aA57B78B54704E256024E
NEXT_PUBLIC_WBNB=0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c
```

---

## 7. 交付验收标准

### Server 必须实现:
- [ ] 抓取 PancakeSwap 新池数据
- [ ] 存储代币信息到 PostgreSQL
- [ ] 提供 `/api/tokens/pending` 接口
- [ ] 提供 `/api/tokens/analyzed` 接口
- [ ] 提供 `/api/tokens/:address/analysis` 接口接收分析结果
  - [ ] 校验 `X-API-Key`

### OpenClaw 必须实现:
- [ ] `fetchPendingTokens` Tool
- [ ] `analyzeTokenRisk` Tool (AI 分析核心)
- [ ] `submitAnalysis` Tool
- [ ] Cron 每 5 分钟自动运行
- [ ] Memory 持久化

### 演示必须展示:
- [ ] Agent 自动从 Server 获取代币
- [ ] AI 分析并判断金狗
- [ ] 分析结果回写 Server
- [ ] Web 展示分析结果

---

*文档结束 - Codex 请按此规格开发*
