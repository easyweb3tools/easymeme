# EasyMeme - BNB Chain 自治 Agent 产品规格书

> **本文档供 Codex 开发使用** | 最后更新: 2026-02-07

---

## 1. 产品定位

**EasyMeme 是什么：**
一个**长期运行**在 BNB Chain 上的**自治 Agent**，能够持续发现、判断、跟踪 Meme 币机会，并**自动执行交易**。

**核心理念：**
- **金狗有时效性** - 代币的"金狗"属性会随时间衰减，识别规则必须动态演进
- **OpenClaw 是学习型 Agent** - 通过 Memory 积累实战经验，从用户互动中学习
- **去中心化个人部署** - EasyMeme 本质上服务个人，建议每个人搭建自己的 AI 交易系统

> 📌 **开源地址**: https://github.com/easyweb3tools/easymeme

**为什么必须用 OpenClaw 构建：**
- Agent 自主决策：不是规则触发，而是 AI 判断
- Memory 积累经验：记住风险模式，从成功/失败中学习
- Cron 持续运行：自动唤醒，无需外部调度
- 用户互动学习：通过 OpenClaw Dialog / Telegram 与用户对话

---

## 2. 系统架构

```
┌──────────────────────────────────────────────────────────────────────┐
│                        EasyMeme MVP V2                                │
├──────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  ┌─────────────┐    ┌──────────────────┐    ┌─────────────────────┐  │
│  │   Web UI    │◄──►│     Server       │◄──►│   OpenClaw Agent    │  │
│  │  (Next.js)  │    │     (Go)         │    │                     │  │
│  └─────────────┘    └──────────────────┘    └─────────────────────┘  │
│        │                    │                        │                │
│        │                    │                        │                │
│  ┌─────▼─────┐       ┌──────▼──────┐         ┌──────▼──────┐         │
│  │ 外部交易  │       │  PostgreSQL │         │  Agent 钱包  │         │
│  │ (GMGN等)  │       │  + Memory   │         │ (托管交易)   │         │
│  └───────────┘       └─────────────┘         └─────────────┘         │
│                                                                       │
└──────────────────────────────────────────────────────────────────────┘
```

### 2.1 职责划分

| 组件 | 职责 | 技术栈 |
|------|------|--------|
| **Server** | 链上数据抓取、数据库存储、提供 API | Go + PostgreSQL |
| **OpenClaw** | AI 分析、金狗识别、自动交易、用户互动学习 | TypeScript + OpenClaw SDK |
| **Web** | 首页自部署指南、金狗列表、AI 交易历史 | Next.js |

### 2.2 数据流

```
1. Server 定时抓取 BSC 链上新代币数据 → 存入 DB
                    ↓
2. OpenClaw Agent 定时从 Server API 获取待分析代币
                    ↓
3. OpenClaw 用 AI 分析风险，判断是否"金狗"，计算时效因子
                    ↓
4. OpenClaw 将分析结果回写到 Server API
                    ↓
5. 如符合自动交易条件，OpenClaw 使用托管钱包执行交易
                    ↓
6. Web 从 Server 获取金狗列表和 AI 交易历史展示
                    ↓
7. 用户如需手动交易，跳转 GMGN/DEXTools
```

---

## 3. Server 规格 (Go)

### 3.1 职责
- **数据抓取**: 监听 PancakeSwap PairCreated 事件
- **数据存储**: PostgreSQL 存储代币信息、分析结果、AI 交易历史
- **API 服务**: 提供 REST API 供 OpenClaw 和 Web 调用

### 3.2 API 端点

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/tokens/pending` | 获取待分析的代币列表 |
| GET | `/api/tokens/analyzed` | 获取已分析代币列表 |
| GET | `/api/tokens/golden-dogs` | 获取金狗列表（含时效信息） |
| GET | `/api/tokens/:address` | 获取单个代币详情 |
| POST | `/api/tokens/:address/analysis` | 回写分析结果 |
| POST | `/api/feedback` | 接收用户反馈（via OpenClaw/Telegram） |
| GET | `/api/ai-trades` | 获取 AI 交易历史 |
| GET | `/api/ai-trades/stats` | AI 交易统计（胜率、盈亏） |
| POST | `/api/wallet/create` | 创建托管钱包 |
| GET | `/api/wallet/balance` | 查询托管钱包余额 |
| POST | `/api/wallet/withdraw` | 从托管钱包提取 |
| POST | `/api/wallet/config` | 配置自动交易参数 |

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
    GoldenDogScore   int       // 金狗基础分数 0-100
    AnalysisResult   JSON      // 详细分析报告
    AnalyzedAt       time.Time
}

// 时效性计算方法
func (t *Token) GoldenDogPhase() string  // EARLY | PEAK | DECLINING | EXPIRED
func (t *Token) TimeDecayFactor() float64 // 0.0-1.0
func (t *Token) EffectiveScore() int      // GoldenDogScore × TimeDecayFactor
```

**AITrade (AI 交易记录)**
```go
type AITrade struct {
    ID              string
    TokenAddress    string
    TokenSymbol     string
    Type            string    // BUY | SELL
    AmountIn        string    // BNB
    AmountOut       string    // Token
    TxHash          string
    Timestamp       time.Time
    
    // AI 决策记录
    GoldenDogScore  int
    DecisionReason  string
    StrategyUsed    string
    
    // 结果追踪
    CurrentValue    string    // 当前价值
    ProfitLoss      float64   // 盈亏 %
}
```

**ManagedWallet (托管钱包)**
```go
type ManagedWallet struct {
    ID              string
    UserID          string
    Address         string
    EncryptedKey    []byte    // AES-256-GCM 加密
    MaxBalance      float64   // 最大余额限制，默认 5 BNB
    CreatedAt       time.Time
}
```

### 3.4 动态金狗时效规则

| 阶段 | 时间范围 | 时效因子 | 说明 |
|------|---------|---------|------|
| EARLY | 0-30分钟 | 1.0 | 刚发现，最佳时机 |
| PEAK | 30分钟-2小时 | 0.8-1.0 | 热度上升期 |
| DECLINING | 2-6小时 | 0.5-0.8 | 热度下降 |
| EXPIRED | >6小时 | <0.5 | 不再推荐 |

### 3.5 安全与校验

- **CORS**: 只允许配置的来源，不能使用 `*` + `AllowCredentials`
- **API Key**: `POST /api/tokens/:address/analysis` 必须校验 `X-API-Key`
- **输入校验**: `riskScore` 必须 0-100，`riskLevel` 只能是 SAFE/WARNING/DANGER
- **托管钱包**: 私钥 AES-256-GCM 加密存储，单钱包最大余额 5 BNB

---

## 4. OpenClaw Agent 规格

### 4.1 核心定位

> **OpenClaw 做 AI 分析 + 自动交易 + 用户互动学习**

OpenClaw 的职责：
- 分析代币是否"金狗"
- 识别风险模式并记忆
- 使用托管钱包自动交易
- 通过 OpenClaw Dialog / Telegram 与用户互动学习

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
│      - 评估风险因素                                  │
│      - 判断是否金狗                                  │
│      - 计算金狗分数                                  │
│            ↓                                         │
│   4. 将分析结果回写 Server API                       │
│            ↓                                         │
│   5. 如 effectiveScore >= 阈值，执行自动买入         │
│            ↓                                         │
│   6. 更新 Memory (记住风险模式和交易结果)            │
│                                                      │
└─────────────────────────────────────────────────────┘
```

### 4.3 Tool 定义

**已有 Tools:**
- `fetchPendingTokens` - 获取待分析代币
- `analyzeTokenRisk` - AI 分析代币风险
- `submitAnalysis` - 回写分析结果

**新增 Tools:**

```typescript
// 执行自动交易
interface ExecuteTradeInput {
  tokenAddress: string;
  type: 'BUY' | 'SELL';
  amountBNB: number;
  reason: string;
}

// 记录代币实际表现
interface RecordOutcomeInput {
  tokenAddress: string;
  outcome: 'MOON' | 'RUG' | 'FLAT';
  maxGain: number;
  maxLoss: number;
  lessonsLearned: string;
}

// 更新识别规则
interface UpdateRuleInput {
  ruleId: string;
  condition: string;
  weight: number;
  source: 'LEARNED' | 'USER_FEEDBACK';
}

// 获取用户反馈
interface GetUserFeedbackInput {
  limit?: number;
}
```

### 4.4 金狗识别（动态规则）

AI 需要综合判断以下因素，权重可通过学习动态调整：

| 因素 | 金狗特征 | 初始权重 | 可学习 |
|------|---------|---------|--------|
| 安全性 | 非貔貅、税率合理、无危险权限 | 必要条件 | ❌ |
| 流动性 | LP 充足且锁定 | 高 | ✅ |
| 持仓分布 | 不集中在少数地址 | 中 | ✅ |
| 创建者历史 | 无 rug 历史 | 高 | ✅ |
| 社区热度 | 有社交媒体关注 | 加分项 | ✅ |

**金狗 ≠ 安全**

金狗是指"值得关注、可能有机会"的代币，需要 AI 做综合判断。

### 4.5 Memory 结构

```typescript
interface LongTermMemory {
  // 风险模式库
  riskPatterns: RiskPattern[];
  
  // 成功/失败案例库
  successCases: TokenOutcome[];
  failureCases: TokenOutcome[];
  
  // 动态识别规则
  goldenDogRules: DynamicRule[];
  
  // 用户信誉
  userReputations: UserReputation[];
}

interface DynamicRule {
  id: string;
  condition: string;
  weight: number;           // -1 到 1
  source: 'INITIAL' | 'LEARNED' | 'USER_FEEDBACK';
  performance: number;      // 规则历史准确率
}
```

### 4.6 用户互动学习

> **互动渠道：OpenClaw Dialog / Telegram，不在 Web 前端**

**反馈机制：**
```typescript
interface UserFeedback {
  tokenAddress: string;
  feedbackType: 'CONFIRM_GOLDEN' | 'DENY_GOLDEN' | 'REPORT_RUG';
  userId: string;           // OpenClaw 用户ID 或 Telegram ID
  channel: 'OPENCLAW_DIALOG' | 'TELEGRAM';
  userReputation: number;   // 用户信誉分 0-100
  feedbackWeight: number;   // 实际权重 = reputation / 100
}
```

**防投毒机制：**
- 新用户（<10次反馈）：权重 = 0.3
- 普通用户：权重 = reputationScore
- 高信誉用户（accuracy > 80%）：权重 = 1.0 + bonus
- 连续错误 ≥3 次：静默，权重 = 0

### 4.7 自动交易策略（可配置）

```typescript
interface AutoTradeConfig {
  // 买入策略
  enabled: boolean;
  maxAmountPerTrade: number;    // BNB，默认 0.1
  minGoldenDogScore: number;    // 触发阈值，默认 75
  dailyBudget: number;          // 每日预算，默认 1 BNB
  confirmThreshold: number;     // 超过此金额需确认，默认 0.5 BNB
  
  // 卖出策略
  takeProfitLevels: number[];   // [0.5, 1.0, 2.0] 即 50%, 100%, 200%
  takeProfitAmounts: number[];  // [0.3, 0.3, 0.4] 每级卖出比例
  stopLoss: number;             // -0.3 即 -30%
  
  // 风控
  maxDailyLoss: number;         // 单日最大亏损，默认 2 BNB
  blacklistedCreators: string[];
}
```

### 4.8 学习回路示例（简化）

**1) 估分输入（analysis 前）**
```json
{
  "riskScore": 78,
  "isGoldenDog": true,
  "riskFactors": {
    "honeypotRisk": "LOW",
    "taxRisk": "MEDIUM",
    "ownerRisk": "LOW",
    "concentrationRisk": "MEDIUM"
  }
}
```

**2) 结果回写（trade outcome 后）**
```json
{
  "tokenAddress": "0xabc...",
  "outcome": "MOON",
  "maxGain": 1.6,
  "maxLoss": -0.1,
  "analysis": {
    "isGoldenDog": true
  }
}
```

**3) 权重变化方向**
- `goldenDogBias` 趋于上调
- `highPenalty` / `mediumPenalty` 趋于下调

---

## 5. Web 规格 (Next.js)

### 5.1 职责

> **前端两大核心：自部署指南 + AI 交易展示**

- 首页突出自部署指南和 GitHub 链接
- 展示金狗列表（含时效状态）
- 展示 AI 交易历史和统计
- 人类交易跳转 GMGN/DEXTools（不在本站交易）

### 5.2 核心页面

| 页面 | 功能 | 优先级 |
|------|------|--------|
| **首页** | Hero 自部署指南 + GitHub 链接 + 快速开始 | P0 |
| **金狗列表** | AI 识别的金狗列表，时效性标记（EARLY/PEAK/DECLINING/EXPIRED） | P0 |
| **AI 交易历史** | Agent 自动交易记录，盈亏统计 | P0 |
| **代币详情** | 单个代币分析报告 + 跳转 GMGN/DEXTools | P1 |

### 5.3 首页 Hero Section

```
┌──────────────────────────────────────────────────────────┐
│                                                          │
│   🐕 EasyMeme - 你的专属 AI Meme 币猎手                    │
│                                                          │
│   自动发现、分析、交易 BNB Chain 上的金狗                   │
│                                                          │
│   ┌────────────────────┐  ┌─────────────────────┐       │
│   │  📦 一键部署       │  │  📚 查看文档        │       │
│   └────────────────────┘  └─────────────────────┘       │
│                                                          │
│   ⭐ GitHub: github.com/easyweb3tools/easymeme            │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

### 5.4 与 Server 交互

- 只读取数据展示
- 不提供交易功能（跳转外部）
- 数据库只存储 AI 交易历史

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
    # 需提供 OpenClaw 配置与 API Key
    
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
WALLET_MASTER_KEY=xxx  # 托管钱包加密密钥
```

**OpenClaw:**
```bash
EASYMEME_SERVER_URL=http://server:8080
EASYMEME_API_KEY=xxx
GEMINI_API_KEY=xxx
# openclaw.json: agents.defaults.model.primary = "google/gemini-3-flash"
```

**Web:**
```bash
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8080/ws
```

---

## 7. 交付验收标准

### Phase 1 - 核心价值（P0）

**Server:**
- [ ] 抓取 PancakeSwap 新池数据
- [ ] 存储代币信息到 PostgreSQL
- [ ] `/api/tokens/pending` 接口
- [ ] `/api/tokens/analyzed` 接口
- [ ] `/api/tokens/golden-dogs` 接口（含时效信息）
- [ ] `/api/tokens/:address/analysis` 接口（校验 API Key）

**OpenClaw:**
- [ ] `fetchPendingTokens` Tool
- [ ] `analyzeTokenRisk` Tool（包含金狗分数）
- [ ] `submitAnalysis` Tool
- [ ] Cron 每 5 分钟自动运行

**Web:**
- [ ] 首页 Hero + 自部署指南 + GitHub 链接
- [ ] 金狗列表页（含时效状态）
- [ ] 代币详情页（跳转 GMGN/DEXTools）

### Phase 2 - AI 自动化（P1）

**Server:**
- [ ] 托管钱包创建 `/api/wallet/create`
- [ ] 托管钱包余额 `/api/wallet/balance`
- [ ] AI 交易历史 `/api/ai-trades`
- [ ] AI 交易统计 `/api/ai-trades/stats`

**OpenClaw:**
- [ ] `executeTrade` Tool（使用托管钱包）
- [ ] `recordOutcome` Tool（记录代币表现）
- [ ] Memory 持久化（风险模式、成功/失败案例）
- [ ] 自动止盈止损

**Web:**
- [ ] AI 交易历史页面
- [ ] AI 交易统计面板

### Phase 3 - 学习增强（P2）

**OpenClaw:**
- [ ] `getUserFeedback` Tool
- [ ] `updateRule` Tool
- [ ] 用户信誉系统（防投毒）
- [ ] 动态规则权重调整

---

*文档结束 - Codex 请按此规格开发*
