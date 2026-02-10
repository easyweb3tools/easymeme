import Link from 'next/link';
import { headers } from 'next/headers';
import { CopyButton } from '@/components/copy-button';
import { ContractSafety } from '@/components/contract-safety';
import { GoldenDogVerdict } from '@/components/golden-dog-verdict';
import { HolderDistribution } from '@/components/holder-distribution';
import { MarketAlerts } from '@/components/market-alerts';
import { MarketDataPanel } from '@/components/market-data-panel';
import { getTokenDetail } from '@/lib/api-server';
import { resolveLang, t, withLang } from '@/lib/i18n';

export const dynamic = 'force-dynamic';

function safeText(value?: string | null) {
  return value && value.length > 0 ? value : 'N/A';
}

function formatDate(value?: string | null) {
  if (!value) return 'N/A';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString();
}

function riskTone(level: string) {
  switch (level) {
    case 'safe':
      return 'bg-[#7cf2a4] text-black';
    case 'warning':
      return 'bg-[#ffbf5c] text-black';
    case 'danger':
      return 'bg-[#f07d7d] text-black';
    default:
      return 'bg-white/10 text-white';
  }
}

type TokenDetailPageProps = {
  params: { address: string };
  searchParams?: { [key: string]: string | string[] | undefined };
};

export default async function TokenDetailPage({ params, searchParams }: TokenDetailPageProps) {
  const lang = resolveLang(searchParams?.lang, headers().get('accept-language'));
  const token = await getTokenDetail(params.address);
  const dextoolsUrl = token.pairAddress
    ? `https://www.dextools.io/app/en/bnb/pair-explorer/${token.pairAddress}`
    : `https://www.dextools.io/app/en/bnb/token/${token.address}`;
  const gmgnUrl = `https://gmgn.ai/bsc/token/${token.address}`;
  const bscScanUrl = `https://bscscan.com/address/${token.address}`;

  const riskFactors =
    (token.analysisResult?.riskFactors as Record<string, string> | undefined) ??
    (token.riskDetails?.risk_factors as Record<string, string> | undefined);
  const recommendation =
    (token.analysisResult?.recommendation as string | undefined) ??
    (token.riskDetails?.reasoning as string | undefined);

  return (
    <div className="min-h-screen px-6 pb-16">
      <header className="max-w-5xl mx-auto py-8">
        <Link href={withLang('/golden-dogs', lang)} className="text-sm text-white/60 hover:text-white">
          {t(lang, 'token_back')}
        </Link>
        <div className="mt-2 flex items-center gap-2 text-xs text-white/60">
          <Link className={lang === 'zh' ? 'text-white' : 'hover:text-white'} href={withLang(`/tokens/${params.address}`, 'zh')}>
            中文
          </Link>
          <span>/</span>
          <Link className={lang === 'en' ? 'text-white' : 'hover:text-white'} href={withLang(`/tokens/${params.address}`, 'en')}>
            EN
          </Link>
        </div>
        <h1 className="text-3xl font-semibold mt-3">
          {safeText(token.symbol)} <span className="text-white/50 text-lg">{safeText(token.name)}</span>
        </h1>
        <div className="mt-2 flex flex-wrap items-center gap-3">
          <p className="text-sm text-white/60">{token.address}</p>
          <CopyButton value={token.address} label={t(lang, 'copy')} copiedLabel={t(lang, 'copied')} />
          <span className={`px-3 py-1 rounded-full text-xs font-semibold ${riskTone(token.riskLevel)}`}>
            {token.riskLevel.toUpperCase()}
          </span>
        </div>
      </header>

      <main className="max-w-5xl mx-auto grid gap-6">
        <GoldenDogVerdict
          lang={lang}
          isGoldenDog={token.isGoldenDog}
          goldenDogScore={token.goldenDogScore}
          recommendation={recommendation}
          riskFactors={riskFactors}
          goplus={token.goplus}
          dexscreener={token.dexscreener}
        />

        <section className="rounded-2xl border border-white/10 bg-white/5 p-6">
          <div className="grid gap-4 md:grid-cols-3 text-sm text-white/70">
            <div>
              <p className="text-white/50 text-xs">{t(lang, 'token_effective')}</p>
              <p className="text-2xl font-semibold text-white">{token.effectiveScore}</p>
              <p className="text-xs text-white/50 mt-1">{t(lang, 'token_phase')}: {token.phase}</p>
            </div>
            <div>
              <p className="text-white/50 text-xs">{t(lang, 'token_golden')}</p>
              <p className="text-2xl font-semibold text-white">{token.goldenDogScore}</p>
              <p className="text-xs text-white/50 mt-1">{t(lang, 'token_time_decay')}: {(token.timeDecayFactor * 100).toFixed(0)}%</p>
            </div>
            <div>
              <p className="text-white/50 text-xs">{t(lang, 'token_risk_score')}</p>
              <p className="text-2xl font-semibold text-white">{token.riskScore}</p>
              <p className="text-xs text-white/50 mt-1">{t(lang, 'gd_filters_risk')}: {token.riskLevel}</p>
            </div>
          </div>
        </section>

        <div className="grid gap-6 md:grid-cols-2">
          <ContractSafety lang={lang} goplus={token.goplus} />
          <MarketDataPanel lang={lang} dexscreener={token.dexscreener} />
        </div>

        <HolderDistribution lang={lang} holderDistribution={token.holderDistribution} />

        <MarketAlerts lang={lang} alerts={token.marketAlerts} />

        <section className="grid gap-4 md:grid-cols-2">
          <div className="rounded-2xl border border-white/10 bg-white/5 p-6 text-sm text-white/70 space-y-2">
            <h2 className="text-lg font-semibold text-white mb-2">{t(lang, 'token_basic')}</h2>
            <div>{t(lang, 'token_dex')}: {safeText(token.dex)}</div>
            <div>{t(lang, 'token_liquidity')}: {token.liquidity.toFixed(4)}</div>
            <div>{t(lang, 'token_creator')}: {safeText(token.creatorAddress)}</div>
            <div>{t(lang, 'token_created')}: {formatDate(token.createdAt)}</div>
            <div>{t(lang, 'token_analyzed')}: {formatDate(token.analyzedAt ?? null)}</div>
          </div>
          <div className="rounded-2xl border border-white/10 bg-white/5 p-6 text-sm text-white/70 space-y-3">
            <h2 className="text-lg font-semibold text-white">{t(lang, 'token_tools')}</h2>
            <a className="block text-white/70 hover:text-white" href={gmgnUrl} target="_blank" rel="noreferrer">GMGN</a>
            <a className="block text-white/70 hover:text-white" href={dextoolsUrl} target="_blank" rel="noreferrer">DEXTools</a>
            <a className="block text-white/70 hover:text-white" href={bscScanUrl} target="_blank" rel="noreferrer">BscScan</a>
          </div>
        </section>

        {(token.riskDetails || token.analysisResult) && (
          <details className="rounded-2xl border border-white/10 bg-white/5 p-6 text-sm text-white/70">
            <summary className="cursor-pointer text-lg font-semibold text-white">{t(lang, 'token_analysis')}</summary>
            <pre className="mt-3 whitespace-pre-wrap text-xs text-white/70">
              {JSON.stringify({ riskDetails: token.riskDetails, analysisResult: token.analysisResult }, null, 2)}
            </pre>
          </details>
        )}
      </main>
    </div>
  );
}
