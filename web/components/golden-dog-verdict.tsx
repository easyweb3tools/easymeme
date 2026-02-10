import { t, type Lang } from '@/lib/i18n';

type RiskFactors = {
  honeypotRisk?: string;
  taxRisk?: string;
  ownerRisk?: string;
  concentrationRisk?: string;
};

type Props = {
  lang: Lang;
  isGoldenDog: boolean;
  goldenDogScore: number;
  recommendation?: string;
  riskFactors?: RiskFactors;
  goplus?: Record<string, unknown>;
  dexscreener?: Record<string, unknown>;
};

type IndicatorLevel = 'LOW' | 'MEDIUM' | 'HIGH' | 'UNKNOWN';

function toNumber(value: unknown): number {
  if (typeof value === 'number') return value;
  if (typeof value === 'string') {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : 0;
  }
  return 0;
}

function toBool(value: unknown): boolean {
  if (typeof value === 'boolean') return value;
  if (typeof value === 'number') return value !== 0;
  if (typeof value === 'string') {
    const v = value.trim().toLowerCase();
    return v === '1' || v === 'true' || v === 'yes';
  }
  return false;
}

function normalizeTax(value: unknown): number {
  const n = toNumber(value);
  return n <= 1 ? n * 100 : n;
}

function indicatorTone(level: IndicatorLevel) {
  if (level === 'LOW') return 'bg-[#7cf2a4]/20 border-[#7cf2a4]/40 text-[#7cf2a4]';
  if (level === 'MEDIUM') return 'bg-[#ffbf5c]/20 border-[#ffbf5c]/40 text-[#ffbf5c]';
  if (level === 'HIGH') return 'bg-[#f07d7d]/20 border-[#f07d7d]/40 text-[#f07d7d]';
  return 'bg-white/10 border-white/20 text-white/70';
}

function indicatorText(level: IndicatorLevel, lang: Lang) {
  if (level === 'LOW') return t(lang, 'token_indicator_pass');
  if (level === 'MEDIUM') return t(lang, 'token_indicator_medium');
  if (level === 'HIGH') return t(lang, 'token_indicator_fail');
  return 'N/A';
}

export function GoldenDogVerdict({
  lang,
  isGoldenDog,
  goldenDogScore,
  recommendation,
  riskFactors,
  goplus,
  dexscreener,
}: Props) {
  const buysH1 = toNumber((dexscreener?.txns as any)?.h1?.buys);
  const sellsH1 = toNumber((dexscreener?.txns as any)?.h1?.sells);
  const momentumLevel: IndicatorLevel = buysH1 > sellsH1 ? 'LOW' : buysH1 === 0 && sellsH1 === 0 ? 'UNKNOWN' : 'MEDIUM';
  const buyTax = normalizeTax(goplus?.buy_tax);
  const sellTax = normalizeTax(goplus?.sell_tax);

  const safetyLevel = (riskFactors?.honeypotRisk?.toUpperCase() as IndicatorLevel) || 'UNKNOWN';
  const taxLevel = (riskFactors?.taxRisk?.toUpperCase() as IndicatorLevel) || 'UNKNOWN';
  const ownerLevel = (riskFactors?.ownerRisk?.toUpperCase() as IndicatorLevel) || 'UNKNOWN';

  const fallbackRecommendation = isGoldenDog
    ? `${t(lang, 'token_verdict_positive')} ${t(lang, 'token_verdict_momentum', { buys: buysH1, sells: sellsH1 })}`
    : `${t(lang, 'token_verdict_negative')} ${t(lang, 'token_verdict_wait')}`;

  const indicators = [
    {
      label: t(lang, 'token_indicator_safety'),
      level: safetyLevel,
      detail: toBool(goplus?.is_honeypot) ? t(lang, 'token_detail_honeypot_yes') : t(lang, 'token_detail_honeypot_no'),
    },
    {
      label: t(lang, 'token_indicator_tax'),
      level: taxLevel,
      detail: `${t(lang, 'token_detail_tax', { buy: buyTax.toFixed(1), sell: sellTax.toFixed(1) })}`,
    },
    {
      label: t(lang, 'token_indicator_ownership'),
      level: ownerLevel,
      detail: toBool(goplus?.is_mintable) ? t(lang, 'token_detail_mintable_yes') : t(lang, 'token_detail_mintable_no'),
    },
    {
      label: t(lang, 'token_indicator_momentum'),
      level: momentumLevel,
      detail: t(lang, 'token_verdict_momentum', { buys: buysH1, sells: sellsH1 }),
    },
  ];

  return (
    <section className={`rounded-2xl border p-6 ${isGoldenDog ? 'border-[#7cf2a4]/40 shadow-[0_0_24px_rgba(124,242,164,0.18)]' : 'border-white/15 bg-white/5'}`}>
      <div className="flex items-center justify-between gap-4">
        <h2 className="text-xl font-semibold text-white">{t(lang, 'token_verdict_title')}</h2>
        <div className="text-sm text-white/80">{t(lang, 'token_verdict_score', { score: goldenDogScore })}</div>
      </div>
      <p className="mt-4 text-sm text-white/80">“{recommendation || fallbackRecommendation}”</p>
      <div className="mt-5 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
        {indicators.map((item) => (
          <div key={item.label} className={`rounded-xl border px-4 py-3 ${indicatorTone(item.level)}`}>
            <div className="text-xs uppercase tracking-wide">{item.label}</div>
            <div className="mt-1 text-sm font-semibold">{indicatorText(item.level, lang)}</div>
            <div className="mt-1 text-xs text-white/75">{item.detail}</div>
          </div>
        ))}
      </div>
      <div className="mt-4 text-xs text-white/50">isGoldenDog = {String(isGoldenDog)}</div>
    </section>
  );
}
