# 🐕 EasyMeme - 你的专属 AI Meme 币猎手

> 自动发现、分析、交易 BNB Chain 上的金狗

**开源地址**: https://github.com/easyweb3tools/easymeme

---

## 💡 核心理念

**金狗有时效性** - 代币的"金狗"属性会随时间衰减，识别规则必须动态演进

**OpenClaw 是学习型 Agent** - 通过 Memory 积累实战经验，从成功/失败中学习，越用越聪明

**去中心化个人部署** - EasyMeme 本质上服务个人，建议每个人搭建自己的 AI 自动化交易系统

---

## 🎬 Demo

![Demo](./demo/recording.gif)

Agent 自动：发现新代币 → AI 分析风险 → 识别金狗 → 自动交易

---

## 🔗 为什么必须用 OpenClaw

| 能力 | OpenClaw 组件 | 在 EasyMeme 中的作用 |
|------|--------------|---------------------|
| **自主决策** | Agent | AI 判断代币是否金狗，不靠规则 |
| **历史记忆** | Memory | 记住风险模式，越用越聪明 |
| **持续运行** | Cron | 每 5 分钟自动唤醒分析 |
| **用户互动** | Dialog/Telegram | 与用户对话学习，动态更新规则 |

**核心价值**：OpenClaw 让 EasyMeme 从"工具"变成"会思考、会学习的 Agent"。

---

## 🚀 一键启动

```bash
git clone https://github.com/easyweb3tools/easymeme
cd easymeme
export GEMINI_API_KEY=your_key
docker compose up --build
```

启动后访问 http://localhost:3000

---

## 🧭 架构

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Server    │────▶│  OpenClaw   │     │    Web      │
│   (Go)      │◀────│   Agent     │     │  (Next.js)  │
└─────────────┘     └─────────────┘     └─────────────┘
      │                   │                   │
 链上数据抓取         AI 金狗识别          金狗展示
 托管钱包            自动交易             AI交易历史
```

| 组件 | 职责 |
|------|------|
| `server/` | 抓取 BSC 链上数据，存储到 Postgres，托管钱包管理 |
| `openclaw-skill/` | **AI 分析**，判断金狗，自动交易，用户互动学习 |
| `web/` | 首页自部署指南，金狗列表，AI 交易历史 |

---

## 📦 本地开发

**方式 A：一键启动（推荐）**
```bash
export GEMINI_API_KEY=your_key
# 可选：自有 BSC RPC / BSCScan Key
export BSC_RPC_HTTP=https://your-bsc-http
export BSC_RPC_WS=wss://your-bsc-ws
export BSCSCAN_API_KEY=your_bscscan_key
export EASYMEME_API_KEY=your_api_key

docker compose up -d --build
```

> OpenClaw 官方默认 provider 是 `anthropic`。如果要切换到 Gemini，需要在配置中设置模型，如：
> `agents.defaults.model.primary = "google/gemini-3-flash"`，并提供 `GEMINI_API_KEY`。

本仓库提供了默认的 `openclaw.json`（已设置为 Gemini）。
如需使用其他 provider，请根据官方文档设置对应的 `API_KEY` 环境变量。
Docker Compose 会把该文件挂载到 OpenClaw 配置路径，你可以直接编辑它切换模型。
参考文档：
```
https://docs.openclaw.ai/concepts/model-providers
```

**方式 B：分组件启动（便于调试）**

**1. 启动数据库**
```bash
# 1. 数据库
docker compose up db -d
```

**2. 启动 Server**
```bash
cd server
cp config.toml.example config.toml
# 编辑 config.toml，填入 BSC RPC 和 BscScan Key
export AUTO_MIGRATE=true
export BSC_RPC_HTTP=https://your-bsc-http
export BSC_RPC_WS=wss://your-bsc-ws
export BSCSCAN_API_KEY=your_bscscan_key
export EASYMEME_API_KEY=your_api_key
export CORS_ALLOWED_ORIGINS=http://localhost:3000
export WALLET_MASTER_KEY=your_wallet_master_key
go run ./cmd/server
```

**3. 启动 Web**
```bash
cd web
npm install
export NEXT_PUBLIC_API_URL=http://localhost:8080
export NEXT_PUBLIC_WS_URL=ws://localhost:8080/ws
npm run dev
```

**4. 启动 OpenClaw（本地 embedded）**
```bash
cd openclaw-skill
npm install && npm run build
export EASYMEME_SERVER_URL=http://localhost:8080
export EASYMEME_API_KEY=your_api_key
export EASYMEME_USER_ID=default
# ~/.openclaw/openclaw.json 里配置默认provider为Gemini时，设置GEMINI_API_KEY环境变量
# 其他provider参考官方文档 https://docs.openclaw.ai/concepts/model-providers
export GEMINI_API_KEY=your_key 
openclaw plugins install --link ./
openclaw plugins enable easymeme-openclaw-skill
openclaw agent --local --session-id easymeme --message "分析代币"
```

**常见问题（OpenClaw fetch failed）**
- 确认 Server 已启动：`curl http://localhost:8080/health`
- 确认 `EASYMEME_SERVER_URL` 可访问（Docker 场景注意端口映射）
- 如设置了 `EASYMEME_API_KEY`，Server 也必须配置一致的 `EASYMEME_API_KEY`

---

## 🧠 Memory 学习

OpenClaw Memory 用于：
- 记录已分析代币，避免重复
- 累积风险模式（成功/失败案例）
- 动态调整金狗识别规则权重
- 用户信誉系统（防投毒）

---

## 📊 链上证明

- **Network**: BNB Smart Chain (BSC)
- **Data Source**: BSCScan API + RPC
- **DEX**: PancakeSwap V2

---

## 🏆 Hackathon

**Good Vibes Only: OpenClaw Edition**

本项目参与 BNB Chain 黑客松 Agent Track。

---

## 📜 License

MIT
