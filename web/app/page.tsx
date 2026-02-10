import Link from 'next/link';
import { getGoldenDogs } from '@/lib/api-server';
import { resolveLang, t, withLang } from '@/lib/i18n';
import { headers } from 'next/headers';

export const dynamic = 'force-dynamic';

type HomePageProps = {
  searchParams?: { [key: string]: string | string[] | undefined };
};

export default async function HomePage({ searchParams }: HomePageProps) {
  const lang = resolveLang(searchParams?.lang, headers().get('accept-language'));
  const goldenDogs = await getGoldenDogs(8);
  return (
    <div className="min-h-screen">
      <header className="px-6 py-6">
        <div className="max-w-6xl mx-auto flex items-center justify-between">
          <div className="flex items-center gap-3">
            <img
              src="/logo.png"
              alt="EasyMeme"
              className="h-10 w-10 rounded-xl object-cover border border-white/10 bg-white/5"
            />
            <div>
              <p className="text-lg font-semibold">EasyMeme</p>
              <p className="text-xs text-white/60">Personal AI meme hunter</p>
            </div>
          </div>
          <nav className="flex items-center gap-4 text-sm">
            <Link
              className="text-white/70 hover:text-white"
              href={withLang('/golden-dogs', lang)}
            >
              {t(lang, 'nav_golden')}
            </Link>
            <Link
              className="text-white/70 hover:text-white"
              href={withLang('/ai-trades', lang)}
            >
              {t(lang, 'nav_trades')}
            </Link>
            <a
              className="text-white/70 hover:text-white"
              href="https://github.com/easyweb3tools/easymeme"
              target="_blank"
              rel="noreferrer"
            >
              {t(lang, 'nav_github')}
            </a>
            <div className="flex items-center gap-2 text-xs text-white/60">
              <Link
                className={lang === 'zh' ? 'text-white' : 'hover:text-white'}
                href={withLang('/', 'zh')}
              >
                中文
              </Link>
              <span>/</span>
              <Link
                className={lang === 'en' ? 'text-white' : 'hover:text-white'}
                href={withLang('/', 'en')}
              >
                EN
              </Link>
            </div>
          </nav>
        </div>
      </header>

      <main className="px-6 pb-20">
        <section className="max-w-6xl mx-auto grid gap-10 lg:grid-cols-[1.1fr_0.9fr] items-center">
          <div className="space-y-6">
            <div className="inline-flex items-center gap-2 rounded-full border border-white/20 px-4 py-1 text-xs uppercase tracking-[0.2em] text-white/70">
              {t(lang, 'home_badge')}
            </div>
            <h1 className="text-5xl md:text-6xl font-semibold leading-tight">
              {t(lang, 'home_title')}
            </h1>
            <p className="text-lg text-white/70">
              {t(lang, 'home_desc')}
            </p>
            <div className="flex flex-wrap gap-4">
              <Link
                href={withLang('/golden-dogs', lang)}
                className="px-6 py-3 rounded-xl bg-[#ffbf5c] text-black font-semibold"
              >
                {t(lang, 'home_view_golden')}
              </Link>
              <Link
                href={withLang('/ai-trades', lang)}
                className="px-6 py-3 rounded-xl border border-white/30 text-white font-semibold"
              >
                {t(lang, 'home_view_trades')}
              </Link>
              <a
                href="https://github.com/easyweb3tools/easymeme"
                target="_blank"
                rel="noreferrer"
                className="px-6 py-3 rounded-xl border border-white/30 text-white/80 font-semibold"
              >
                {t(lang, 'home_view_github')}
              </a>
            </div>
          </div>
          <div className="rounded-3xl border border-white/15 bg-white/5 p-6 backdrop-blur">
            <h2 className="text-xl font-semibold mb-4">{t(lang, 'home_deploy_title')}</h2>
            <ol className="space-y-3 text-sm text-white/70">
              <li>{t(lang, 'home_deploy_1')}</li>
              <li>{t(lang, 'home_deploy_2')}</li>
              <li>{t(lang, 'home_deploy_3')}</li>
              <li>{t(lang, 'home_deploy_4')}</li>
            </ol>
            <div className="mt-6 rounded-2xl bg-black/40 p-4 text-xs text-white/70">
              <p style={{ fontFamily: 'var(--font-mono)' }}>
                GitHub: easyweb3tools/easymeme
              </p>
            </div>
          </div>
        </section>

        <section className="max-w-6xl mx-auto mt-16">
          <div className="flex items-center justify-between mb-6">
            <div>
              <h2 className="text-2xl font-semibold">{t(lang, 'home_cards_title')}</h2>
              <p className="text-sm text-white/60">
                {t(lang, 'home_cards_sub')}
              </p>
            </div>
            <Link
              href={withLang('/golden-dogs', lang)}
              className="text-sm text-white/70 hover:text-white"
            >
              {t(lang, 'home_cards_view_all')}
            </Link>
          </div>
          <div className="grid gap-4 md:grid-cols-2">
            {goldenDogs.map((token) => (
              <div
                key={token.address}
                className="rounded-2xl border border-white/10 bg-white/5 p-5 transition hover:border-white/30"
              >
                <Link href={withLang(`/tokens/${token.address}`, lang)} className="block">
                  <div className="flex items-center justify-between gap-3">
                    <div>
                      <h3 className="text-lg font-semibold">
                        {token.symbol || 'UNKNOWN'}{' '}
                        <span className="text-white/50 text-sm">{token.name}</span>
                      </h3>
                      <p className="text-xs text-white/60 mt-1">{token.address}</p>
                    </div>
                    <div className="text-right">
                      <p className="text-xs text-white/50">Effective</p>
                      <p className="text-xl font-semibold text-white">
                        {token.effectiveScore}
                      </p>
                    </div>
                  </div>
                  <div className="mt-4 flex flex-wrap gap-2 text-xs font-semibold">
                    <span
                      className={`px-3 py-1 rounded-full ${
                        token.riskLevel === 'safe'
                          ? 'bg-[#7cf2a4] text-black'
                          : token.riskLevel === 'warning'
                            ? 'bg-[#ffbf5c] text-black'
                            : token.riskLevel === 'danger'
                              ? 'bg-[#f07d7d] text-black'
                              : 'bg-white/10 text-white'
                      }`}
                    >
                      {token.riskLevel.toUpperCase()}
                    </span>
                    <span className="px-3 py-1 rounded-full border border-white/20 text-white/70">
                      GD {token.goldenDogScore}
                    </span>
                    <span className="px-3 py-1 rounded-full border border-white/20 text-white/70">
                      Phase {token.phase}
                    </span>
                  </div>
                </Link>
                <div className="mt-4">
                  <a
                    href={`https://gmgn.ai/bsc/token/${token.address}`}
                    target="_blank"
                    rel="noreferrer"
                    className="inline-flex items-center rounded-lg border border-white/20 px-3 py-1.5 text-xs text-white/80 hover:text-white hover:border-white/40"
                  >
                    GMGN
                  </a>
                </div>
              </div>
            ))}
            {goldenDogs.length === 0 && (
              <div className="rounded-2xl border border-white/10 bg-white/5 p-6 text-sm text-white/70">
                {t(lang, 'home_cards_empty')}
              </div>
            )}
          </div>
        </section>

        <section className="max-w-6xl mx-auto mt-16 grid gap-6 md:grid-cols-3">
          {[
            {
              title: t(lang, 'home_feature_1_title'),
              desc: t(lang, 'home_feature_1_desc'),
            },
            {
              title: t(lang, 'home_feature_2_title'),
              desc: t(lang, 'home_feature_2_desc'),
            },
            {
              title: t(lang, 'home_feature_3_title'),
              desc: t(lang, 'home_feature_3_desc'),
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
