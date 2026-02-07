import Link from 'next/link';

export const dynamic = 'force-dynamic';

export default function HomePage() {
  return (
    <div className="min-h-screen">
      <header className="px-6 py-6">
        <div className="max-w-6xl mx-auto flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="h-10 w-10 rounded-xl bg-[#ffbf5c] text-black flex items-center justify-center font-semibold">
              EM
            </div>
            <div>
              <p className="text-lg font-semibold">EasyMeme</p>
              <p className="text-xs text-white/60">Personal AI meme hunter</p>
            </div>
          </div>
          <nav className="flex items-center gap-4 text-sm">
            <Link className="text-white/70 hover:text-white" href="/golden-dogs">
              Golden Dogs
            </Link>
            <Link className="text-white/70 hover:text-white" href="/ai-trades">
              AI Trades
            </Link>
            <a
              className="text-white/70 hover:text-white"
              href="https://github.com/easyweb3tools/easymeme"
              target="_blank"
              rel="noreferrer"
            >
              GitHub
            </a>
          </nav>
        </div>
      </header>

      <main className="px-6 pb-20">
        <section className="max-w-6xl mx-auto grid gap-10 lg:grid-cols-[1.1fr_0.9fr] items-center">
          <div className="space-y-6">
            <div className="inline-flex items-center gap-2 rounded-full border border-white/20 px-4 py-1 text-xs uppercase tracking-[0.2em] text-white/70">
              BNB Chain • Autonomous Agent
            </div>
            <h1 className="text-5xl md:text-6xl font-semibold leading-tight">
              你的专属 AI Meme 币猎手
            </h1>
            <p className="text-lg text-white/70">
              EasyMeme 持续发现、分析并追踪金狗机会。基于 OpenClaw
              的学习型 Agent，支持个人自部署与长期运行。
            </p>
            <div className="flex flex-wrap gap-4">
              <Link
                href="/golden-dogs"
                className="px-6 py-3 rounded-xl bg-[#ffbf5c] text-black font-semibold"
              >
                查看金狗列表
              </Link>
              <Link
                href="/ai-trades"
                className="px-6 py-3 rounded-xl border border-white/30 text-white font-semibold"
              >
                AI 交易历史
              </Link>
              <a
                href="https://github.com/easyweb3tools/easymeme"
                target="_blank"
                rel="noreferrer"
                className="px-6 py-3 rounded-xl border border-white/30 text-white/80 font-semibold"
              >
                查看 GitHub
              </a>
            </div>
          </div>
          <div className="rounded-3xl border border-white/15 bg-white/5 p-6 backdrop-blur">
            <h2 className="text-xl font-semibold mb-4">一键自部署</h2>
            <ol className="space-y-3 text-sm text-white/70">
              <li>1. 拉取仓库并配置 `.env`</li>
              <li>2. `docker compose up --build` 启动服务</li>
              <li>3. OpenClaw 连接 Server 自动分析</li>
              <li>4. Web 查看金狗与 AI 决策</li>
            </ol>
            <div className="mt-6 rounded-2xl bg-black/40 p-4 text-xs text-white/70">
              <p style={{ fontFamily: 'var(--font-mono)' }}>
                GitHub: easyweb3tools/easymeme
              </p>
            </div>
          </div>
        </section>

        <section className="max-w-6xl mx-auto mt-16 grid gap-6 md:grid-cols-3">
          {[
            {
              title: '动态金狗时效',
              desc: '通过时间衰减模型把握黄金窗口期。',
            },
            {
              title: '可学习策略',
              desc: 'OpenClaw Memory 让规则随反馈进化。',
            },
            {
              title: '个人部署',
              desc: '每个人都能拥有自己的 AI 交易系统。',
            },
          ].map((item) => (
            <div
              key={item.title}
              className="rounded-2xl border border-white/10 bg-white/5 p-6"
            >
              <h3 className="text-lg font-semibold mb-2">{item.title}</h3>
              <p className="text-sm text-white/70">{item.desc}</p>
            </div>
          ))}
        </section>
      </main>
    </div>
  );
}
