# 🦞 EasyMeme - BNB Chain 自治 Agent

> 一个用 OpenClaw 构建的、长期运行的链上 Meme 币猎手

## 🎬 Demo

![Demo](./demo/recording.gif)

Agent 自动：发现新代币 → AI 分析风险 → 识别金狗 → 汇报结果

---

## 🚀 一键启动

```bash
git clone https://github.com/easyweb3tools/easymeme
cd easymeme
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
 链上数据抓取         AI 金狗识别          钱包交易
 数据库存储          风险分析             UI 展示
```

| 组件 | 职责 |
|------|------|
| `server/` | 抓取 BSC 链上数据，存储到 Postgres |
| `openclaw-skill/` | **AI 分析**，判断是否"金狗" |
| `web/` | 钱包连接，交易执行，结果展示 |

---

## 📦 本地开发

**方式 A：一键启动（推荐）**
```bash
# 需要提前设置 Gemini Key（不要提交到仓库）
export GEMINI_API_KEY=your_key
# 可选：建议使用自有 BSC RPC / BSCScan Key，避免公共节点限流
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
go run ./cmd/server
```

**3. 启动 Web**
```bash
cd web
npm install
export NEXT_PUBLIC_API_URL=http://localhost:8080
export NEXT_PUBLIC_WS_URL=ws://localhost:8080/ws
export NEXT_PUBLIC_PANCAKE_ROUTER=0x10ED43C718714eb63d5aA57B78B54704E256024E
export NEXT_PUBLIC_WBNB=0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c
npm run dev
```

**4. 启动 OpenClaw（本地 embedded）**
```bash
cd openclaw-skill
npm install && npm run build
export EASYMEME_SERVER_URL=http://localhost:8080
export EASYMEME_API_KEY=your_api_key
# ~/.openclaw/openclaw.json 里配置默认provider为Gemini时，设置GEMINI_API_KEY环境变量
# 其他provider参考官方文档 https://docs.openclaw.ai/concepts/model-providers
export GEMINI_API_KEY=your_key 
openclaw plugins install --link ./
openclaw plugins enable easymeme-openclaw-skill
openclaw agent --local --session-id easymeme --message "获取待分析代币 -> AI 分析 -> 回写结果"
```

---

## 🔗 为什么必须用 OpenClaw

| 能力 | OpenClaw 组件 | 在 EasyMeme 中的作用 |
|------|--------------|---------------------|
| **自主决策** | Agent | AI 判断代币是否金狗，不靠规则 |
| **历史记忆** | Memory | 记住风险模式，越用越聪明 |
| **持续运行** | Cron | 每 5 分钟自动唤醒分析 |
| **多端响应** | Channels | Telegram/Discord 推送发现 |

**核心价值**：OpenClaw 让 EasyMeme 从"工具"变成"会思考的 Agent"。

---

## 🧠 Memory 未实现/后续

当前版本尚未落地 OpenClaw Memory 持久化（风险模式、已分析代币去重、金狗历史表现等）。
后续计划：

- 接入 OpenClaw Memory（长期记忆）
- 记录已分析代币，避免重复分析
- 累积风险模式与金狗表现

---

## 📊 链上证明

- **Network**: BNB Smart Chain (BSC)
- **Data Source**: BSCScan API + RPC
- **Example**: [View on BSCScan](https://bscscan.com/tx/0x...)

---

## 🏆 Hackathon

**Good Vibes Only: OpenClaw Edition**

本项目参与 BNB Chain 黑客松 Agent Track。

---

## 📜 License

MIT
