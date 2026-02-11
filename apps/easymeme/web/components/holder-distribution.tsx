import { t, type Lang } from '@/lib/i18n';

type Props = {
  lang: Lang;
  holderDistribution?: Record<string, unknown>;
};

function toNumber(value: unknown): number {
  if (typeof value === 'number') return value;
  if (typeof value === 'string') {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : 0;
  }
  return 0;
}

export function HolderDistribution({ lang, holderDistribution }: Props) {
  if (!holderDistribution || Object.keys(holderDistribution).length === 0) {
    return <section className="rounded-2xl border border-white/10 bg-white/5 p-6 text-sm text-white/60">{t(lang, 'token_holder_empty')}</section>;
  }

  const top10Share = Math.max(0, Math.min(1, toNumber(holderDistribution.top10Share)));
  const total = toNumber(holderDistribution.total);
  const pct = (top10Share * 100).toFixed(1);
  const restPct = (100 - top10Share * 100).toFixed(1);
  const tone = top10Share > 0.8 ? 'bg-[#f07d7d]' : top10Share >= 0.6 ? 'bg-[#ffbf5c]' : 'bg-[#7cf2a4]';

  return (
    <section className="rounded-2xl border border-white/10 bg-white/5 p-6">
      <h3 className="text-lg font-semibold text-white">{t(lang, 'token_holder_title')}</h3>
      <div className="mt-4 text-sm text-white/80">{t(lang, 'token_holder_top10')}: {pct}%</div>
      <div className="mt-3 h-3 w-full overflow-hidden rounded-full bg-white/10">
        <div className={`h-3 ${tone}`} style={{ width: `${top10Share * 100}%` }} />
      </div>
      <div className="mt-2 text-xs text-white/70">{pct}% / {restPct}%</div>
      <div className="mt-3 text-sm text-white/80">{t(lang, 'token_holder_total')}: {total}</div>
    </section>
  );
}
