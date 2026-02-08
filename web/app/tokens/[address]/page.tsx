import Link from 'next/link';
import { getTokenDetail } from '@/lib/api-server';
import { CopyButton } from '@/components/copy-button';

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

export default async function TokenDetailPage({
  params,
}: {
  params: { address: string };
}) {
  const token = await getTokenDetail(params.address);
  const dextoolsUrl = token.pairAddress
    ? `https://www.dextools.io/app/en/bnb/pair-explorer/${token.pairAddress}`
    : `https://www.dextools.io/app/en/bnb/token/${token.address}`;
  const gmgnUrl = `https://gmgn.ai/bsc/token/${token.address}`;
  const bscScanUrl = `https://bscscan.com/address/${token.address}`;
  const riskFactors =
    (token.riskDetails?.risk_factors as Record<string, string> | undefined) ??
    (token.analysisResult?.riskFactors as Record<string, string> | undefined);
  const reasoning =
    (token.riskDetails?.reasoning as string | undefined) ??
    (token.analysisResult?.reasoning as string | undefined);
  const recommendation =
    (token.riskDetails?.recommendation as string | undefined) ??
    (token.analysisResult?.recommendation as string | undefined);

  return (
    <div className="min-h-screen px-6 pb-16">
      <header className="max-w-5xl mx-auto py-8">
        <Link href="/golden-dogs" className="text-sm text-white/60 hover:text-white">
          ← 返回金狗列表
        </Link>
        <h1 className="text-3xl font-semibold mt-3">
          {safeText(token.symbol)}{' '}
          <span className="text-white/50 text-lg">{safeText(token.name)}</span>
        </h1>
        <div className="mt-2 flex flex-wrap items-center gap-3">
          <p className="text-sm text-white/60">{token.address}</p>
          <CopyButton value={token.address} />
          <span className={`px-3 py-1 rounded-full text-xs font-semibold ${riskTone(token.riskLevel)}`}>
            {token.riskLevel.toUpperCase()}
          </span>
        </div>
      </header>

      <main className="max-w-5xl mx-auto grid gap-6">
        <section className="rounded-2xl border border-white/10 bg-white/5 p-6">
          <div className="grid gap-4 md:grid-cols-3 text-sm text-white/70">
            <div>
              <p className="text-white/50 text-xs">有效分数</p>
              <p className="text-2xl font-semibold text-white">
                {token.effectiveScore}
              </p>
              <p className="text-xs text-white/50 mt-1">Phase: {token.phase}</p>
            </div>
            <div>
              <p className="text-white/50 text-xs">金狗分数</p>
              <p className="text-2xl font-semibold text-white">
                {token.goldenDogScore}
              </p>
              <p className="text-xs text-white/50 mt-1">
                Time Decay: {(token.timeDecayFactor * 100).toFixed(0)}%
              </p>
            </div>
            <div>
              <p className="text-white/50 text-xs">风险评分</p>
              <p className="text-2xl font-semibold text-white">{token.riskScore}</p>
              <p className="text-xs text-white/50 mt-1">
                Risk Level: {token.riskLevel}
              </p>
            </div>
          </div>
        </section>

        {(reasoning || recommendation) && (
          <section className="rounded-2xl border border-white/10 bg-white/5 p-6 text-sm text-white/70">
            <h2 className="text-lg font-semibold text-white mb-3">决策摘要</h2>
            <div className="grid gap-4 md:grid-cols-2">
              <div className="rounded-xl border border-white/10 bg-black/30 p-4">
                <p className="text-xs uppercase tracking-widest text-white/50 mb-2">
                  Reasoning
                </p>
                <p className="text-sm text-white/80">
                  {reasoning ?? 'N/A'}
                </p>
              </div>
              <div className="rounded-xl border border-white/10 bg-black/30 p-4">
                <p className="text-xs uppercase tracking-widest text-white/50 mb-2">
                  Recommendation
                </p>
                <p className="text-sm text-white/80">
                  {recommendation ?? 'N/A'}
                </p>
              </div>
            </div>
          </section>
        )}

        <section className="grid gap-4 md:grid-cols-2">
          <div className="rounded-2xl border border-white/10 bg-white/5 p-6 text-sm text-white/70 space-y-2">
            <h2 className="text-lg font-semibold text-white mb-2">基础信息</h2>
            <div>DEX: {safeText(token.dex)}</div>
            <div>Liquidity: {token.liquidity.toFixed(4)}</div>
            <div>Creator: {safeText(token.creatorAddress)}</div>
            <div>Created At: {formatDate(token.createdAt)}</div>
            <div>Analyzed At: {formatDate(token.analyzedAt ?? null)}</div>
          </div>
          <div className="rounded-2xl border border-white/10 bg-white/5 p-6 text-sm text-white/70 space-y-3">
            <h2 className="text-lg font-semibold text-white">外部工具</h2>
            <a
              className="block text-white/70 hover:text-white"
              href={gmgnUrl}
              target="_blank"
              rel="noreferrer"
            >
              GMGN
            </a>
            <a
              className="block text-white/70 hover:text-white"
              href={dextoolsUrl}
              target="_blank"
              rel="noreferrer"
            >
              DEXTools
            </a>
            <a
              className="block text-white/70 hover:text-white"
              href={bscScanUrl}
              target="_blank"
              rel="noreferrer"
            >
              BscScan
            </a>
          </div>
        </section>

        {riskFactors && (
          <section className="rounded-2xl border border-white/10 bg-white/5 p-6 text-sm text-white/70">
            <h2 className="text-lg font-semibold text-white mb-3">风险因子</h2>
            <div className="flex flex-wrap gap-2 text-xs">
              {Object.entries(riskFactors).map(([key, value]) => (
                <span
                  key={key}
                  className="px-3 py-1 rounded-full border border-white/20 text-white/70"
                >
                  {key}: {value}
                </span>
              ))}
            </div>
          </section>
        )}

        {(token.riskDetails || token.analysisResult) && (
          <section className="rounded-2xl border border-white/10 bg-white/5 p-6 text-sm text-white/70">
            <h2 className="text-lg font-semibold text-white mb-3">AI 分析详情</h2>
            <pre className="whitespace-pre-wrap text-xs text-white/70">
              {JSON.stringify(
                { riskDetails: token.riskDetails, analysisResult: token.analysisResult },
                null,
                2,
              )}
            </pre>
          </section>
        )}
      </main>
    </div>
  );
}
