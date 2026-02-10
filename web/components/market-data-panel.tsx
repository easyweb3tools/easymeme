import { t, type Lang } from '@/lib/i18n';

type Props = {
  lang: Lang;
  dexscreener?: Record<string, unknown>;
};

function toNumber(v: unknown): number {
  if (typeof v === 'number') return v;
  if (typeof v === 'string') {
    const parsed = Number(v);
    return Number.isFinite(parsed) ? parsed : 0;
  }
  return 0;
}

function money(v: number): string {
  return new Intl.NumberFormat('en-US', { maximumFractionDigits: 2 }).format(v);
}

function pct(v: number): string {
  const sign = v > 0 ? '+' : '';
  return `${sign}${v.toFixed(1)}%`;
}

export function MarketDataPanel({ lang, dexscreener }: Props) {
  if (!dexscreener || Object.keys(dexscreener).length === 0) {
    return <section className="rounded-2xl border border-white/10 bg-white/5 p-6 text-sm text-white/60">{t(lang, 'token_market_empty')}</section>;
  }

  const priceUsd = String(dexscreener.priceUsd ?? '0');
  const priceChange = (dexscreener.priceChange as Record<string, unknown>) || {};
  const volume = (dexscreener.volume as Record<string, unknown>) || {};
  const liquidity = (dexscreener.liquidity as Record<string, unknown>) || {};
  const txH1 = ((dexscreener.txns as Record<string, unknown>)?.h1 as Record<string, unknown>) || {};
  const buys = toNumber(txH1.buys);
  const sells = toNumber(txH1.sells);
  const total = Math.max(1, buys + sells);
  const buyWidth = `${(buys / total) * 100}%`;
  const sellWidth = `${(sells / total) * 100}%`;

  return (
    <section className="rounded-2xl border border-white/10 bg-white/5 p-6">
      <h3 className="text-lg font-semibold text-white">{t(lang, 'token_market_title')}</h3>
      <div className="mt-4 text-sm text-white/80">
        <div>{t(lang, 'token_market_price')}: <span className="text-white">${priceUsd}</span></div>
        <div className="mt-3 flex flex-wrap gap-3">
          {['h1', 'h6', 'h24'].map((k) => {
            const val = toNumber(priceChange[k]);
            const tone = val >= 0 ? 'text-[#7cf2a4]' : 'text-[#f07d7d]';
            return <span key={k} className={tone}>{k.toUpperCase()}: {pct(val)}</span>;
          })}
        </div>
        <div className="mt-3">{t(lang, 'token_market_volume_h1')}: ${money(toNumber(volume.h1))}</div>
        <div>{t(lang, 'token_market_liquidity')}: ${money(toNumber(liquidity.usd))}</div>
      </div>
      <div className="mt-4">
        <div className="mb-2 text-sm text-white/80">{t(lang, 'token_market_txns_h1')}</div>
        <div className="h-3 w-full overflow-hidden rounded-full bg-white/10">
          <div className="h-3 bg-[#7cf2a4]" style={{ width: buyWidth, float: 'left' }} />
          <div className="h-3 bg-[#f07d7d]" style={{ width: sellWidth, float: 'left' }} />
        </div>
        <div className="mt-2 flex justify-between text-xs text-white/70">
          <span>{buys} {t(lang, 'token_market_buys')}</span>
          <span>{sells} {t(lang, 'token_market_sells')}</span>
        </div>
      </div>
    </section>
  );
}
